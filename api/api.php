<?php
header('Access-Control-Allow-Origin: *');
header('Content-Type: application/json');
header('Access-Control-Allow-Methods: POST, OPTIONS');
header('Access-Control-Allow-Headers: Access-Control-Allow-Headers,Content-Type,Access-Control-Allow-Methods,Authorization,X-Requested-With');

// Handle preflight OPTIONS request
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit;
}

// Only accept POST requests
if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    echo json_encode(['success' => false, 'error' => 'Only POST method is allowed']);
    exit;
}

// Include database configuration
require_once 'config.php';

// Get POST data early to check for database name
$postData = json_decode(file_get_contents('php://input'), true);

// Function to get database configuration (uses POST dbname if provided, otherwise uses config.php)
function getDbConfig() {
    global $db_config, $postData;
    
    // If POST data contains dbname, use it; otherwise use default from config
    if ($postData && isset($postData['dbname']) && !empty($postData['dbname'])) {
        return [
            'host' => $db_config['host'],
            'user' => $db_config['user'],
            'pass' => $db_config['pass'],
            'dbname' => $postData['dbname']
        ];
    }
    
    return $db_config;
}

// Function to establish database connection
function getConnection($config) {
    try {
        $conn = new PDO(
            "mysql:host={$config['host']};dbname={$config['dbname']}", 
            $config['user'], 
            $config['pass']
        );
        $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
        return $conn;
    } catch(PDOException $e) {
        return ['error' => $e->getMessage()];
    }
}

// Function to execute SQL query and return appropriate result
function executeQuery($sql, $params = []) {
    try {
        $conn = getConnection(getDbConfig());
        
        if (is_array($conn) && isset($conn['error'])) {
            throw new Exception($conn['error']);
        }

        $stmt = $conn->prepare($sql);
        $stmt->execute($params);

        // Determine query type
        $queryType = strtoupper(substr(trim($sql), 0, 6));
        
        switch($queryType) {
            case 'SELECT':
                return ['success' => true, 'data' => $stmt->fetchAll(PDO::FETCH_ASSOC)];
            case 'INSERT':
                return ['success' => true, 'lastInsertId' => $conn->lastInsertId()];
            case 'UPDATE':
            case 'DELETE':
                return ['success' => true, 'affectedRows' => $stmt->rowCount()];
            default:
                return ['success' => false, 'error' => 'Invalid query type'];
        }
    } catch(Exception $e) {
        return ['success' => false, 'error' => $e->getMessage()];
    }
}

// Function to execute multi-statement SQL query
function executeMultiQuery($sql, $params = []) {
    try {
        $conn = getConnection(getDbConfig());
        
        if (is_array($conn) && isset($conn['error'])) {
            throw new Exception($conn['error']);
        }

        // Split the SQL into individual statements
        $statements = array_filter(array_map('trim', explode(';', $sql)));
        $results = [];
        $lastResult = null;
        $totalAffectedRows = 0;
        $totalResults = 0;

        foreach ($statements as $statement) {
            if (empty($statement)) continue;
            
            $stmt = $conn->prepare($statement);
            $stmt->execute($params);
            
            // Determine query type
            $queryType = strtoupper(substr(trim($statement), 0, 6));
            
            switch($queryType) {
                case 'SELECT':
                    $data = $stmt->fetchAll(PDO::FETCH_ASSOC);
                    $lastResult = ['type' => 'SELECT', 'data' => $data];
                    $totalResults += count($data);
                    break;
                case 'INSERT':
                    $lastResult = ['type' => 'INSERT', 'lastInsertId' => $conn->lastInsertId(), 'affectedRows' => $stmt->rowCount()];
                    $totalAffectedRows += $stmt->rowCount();
                    break;
                case 'UPDATE':
                case 'DELETE':
                    $lastResult = ['type' => $queryType, 'affectedRows' => $stmt->rowCount()];
                    $totalAffectedRows += $stmt->rowCount();
                    break;
                case 'SET':
                case 'START':
                case 'COMMIT':
                case 'ROLLBACK':
                    $lastResult = ['type' => $queryType, 'success' => true];
                    break;
                default:
                    $lastResult = ['type' => 'OTHER', 'success' => true];
            }
            
            $results[] = $lastResult;
        }

        // Return the last SELECT result as the main data, or success status
        if ($lastResult && $lastResult['type'] === 'SELECT') {
            return [
                'success' => true, 
                'results' => $lastResult['data'], 
                'allResults' => $results,
                'totalAffectedRows' => $totalAffectedRows,
                'totalResults' => $totalResults
            ];
        } else {
            return [
                'success' => true, 
                'message' => 'Multi-statement executed successfully', 
                'allResults' => $results,
                'totalAffectedRows' => $totalAffectedRows,
                'totalResults' => $totalResults
            ];
        }
        
    } catch(Exception $e) {
        return ['success' => false, 'error' => $e->getMessage()];
    }
}

// Check if this is a special action request
if (isset($postData['action'])) {
    switch($postData['action']) {
        case 'ping':
            // Simple ping action for cookie verification
            echo json_encode(['success' => true, 'message' => 'pong']);
            exit;
            
        case 'execute_sql':
            // Check if user is admin (you may need to implement proper session/auth check)
            // For now, we'll add basic security measures
            if (!isset($postData['query']) || empty($postData['query'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'message' => 'SQL query is required']);
                exit;
            }
            
            $query = $postData['query'];
            $params = isset($postData['params']) ? $postData['params'] : [];
            
            // Basic security: prevent multiple statements
            if (strpos($query, ';') !== false && strpos($query, ';') !== strlen($query) - 1) {
                http_response_code(400);
                echo json_encode(['success' => false, 'message' => 'Multiple statements not allowed']);
                exit;
            }
            
            // Execute the query with parameters
            $result = executeQuery($query, $params);
            
            if ($result['success']) {
                if (isset($result['data'])) {
                    echo json_encode(['success' => true, 'results' => $result['data']]);
                } else if (isset($result['affectedRows'])) {
                    echo json_encode(['success' => true, 'affectedRows' => $result['affectedRows']]);
                } else if (isset($result['lastInsertId'])) {
                    echo json_encode(['success' => true, 'lastInsertId' => $result['lastInsertId']]);
                } else {
                    echo json_encode(['success' => true, 'message' => 'Query executed successfully']);
                }
            } else {
                echo json_encode(['success' => false, 'message' => $result['error']]);
            }
            exit;

        case 'execute_multi_sql':
            // Handle multi-statement SQL queries
            if (!isset($postData['query']) || empty($postData['query'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'message' => 'SQL query is required']);
                exit;
            }
            
            $query = $postData['query'];
            $params = isset($postData['params']) ? $postData['params'] : [];
            
            // Execute the multi-statement query
            $result = executeMultiQuery($query, $params);
            
            if ($result['success']) {
            echo json_encode($result);
            } else {
                echo json_encode(['success' => false, 'message' => $result['error']]);
            }
            exit;

        case 'get_unique_containers_ref':
            // Get all unique, non-null containers_ref from cars_stock table
            $query = "SELECT DISTINCT container_ref FROM cars_stock WHERE container_ref IS NOT NULL AND container_ref != '' ORDER BY container_ref ASC";
            $result = executeQuery($query);
            
            if ($result['success']) {
                echo json_encode(['success' => true, 'data' => $result['data']]);
            } else {
                echo json_encode(['success' => false, 'error' => $result['error']]);
            }
            exit;

        case 'verify_password':
            if (!isset($postData['password']) || !isset($postData['hash'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Missing password or hash']);
                exit;
            }
            
            $isValid = password_verify($postData['password'], $postData['hash']);
            echo json_encode(['success' => $isValid]);
            exit;

        case 'hash_password':
            if (!isset($postData['query']) || !isset($postData['params']) || count($postData['params']) !== 2) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Invalid parameters for password update']);
                exit;
            }

            // Hash the password (first parameter)
            $hashedPassword = password_hash($postData['params'][0], PASSWORD_DEFAULT);
            
            // Replace the plain password with hashed password in params
            $postData['params'][0] = $hashedPassword;
            
            // Execute the update query with hashed password
            $result = executeQuery($postData['query'], $postData['params']);
            echo json_encode($result);
            exit;

            case 'insert_user':
                if (!isset($postData['query']) ||!isset($postData['params']) || count($postData['params'])!== 5) {
                    http_response_code(400);
                    echo json_encode(['success' => false, 'error' => 'Invalid parameters for user creation']);
                    exit;
                }

                // Hash the password (third parameter - index 2)
                $hashedPassword = password_hash($postData['params'][2], PASSWORD_DEFAULT);
                
                // Replace the plain password with hashed password in params
                $postData['params'][2] = $hashedPassword;
                
                // Execute the insert query with hashed password
                $result = executeQuery($postData['query'], $postData['params']);
                echo json_encode($result);
                exit;

        case 'assign_multiple_vins':
            if (!isset($postData['assignments']) || !is_array($postData['assignments'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Assignments array is required']);
                exit;
            }
            
            $assignments = $postData['assignments'];
            $successCount = 0;
            $errors = [];
            
            try {
                $conn = getConnection(getDbConfig());
                
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Start transaction
                $conn->beginTransaction();
                
                foreach ($assignments as $assignment) {
                    if (!isset($assignment['carId']) || !array_key_exists('vin', $assignment)) {
                        $errors[] = 'Invalid assignment data';
                        continue;
                    }
                    
                    $carId = $assignment['carId'];
                    $vin = $assignment['vin'];
                    
                    // Update the car with the new VIN
                    $stmt = $conn->prepare('UPDATE cars_stock SET vin = ? WHERE id = ?');
                    $result = $stmt->execute([$vin, $carId]);
                    
                    if ($result) {
                        $successCount++;
                    } else {
                        $errors[] = "Failed to update car ID: $carId";
                    }
                }
                
                // Commit transaction if all updates were successful
                if (empty($errors)) {
                    $conn->commit();
                    echo json_encode([
                        'success' => true, 
                        'message' => "Successfully assigned VINs to $successCount cars",
                        'assignedCount' => $successCount
                    ]);
                } else {
                    // Rollback on any error
                    $conn->rollBack();
                    echo json_encode([
                        'success' => false, 
                        'error' => 'Some assignments failed',
                        'errors' => $errors,
                        'successCount' => $successCount
                    ]);
                }
                
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
                exit;

        // ============================================
        // Car Files Management Actions
        // ============================================
        
        case 'get_car_file_categories':
            // Get all file categories ordered by display_order
            $query = "SELECT * FROM car_file_categories ORDER BY display_order ASC, importance_level ASC";
            $result = executeQuery($query);
            echo json_encode($result);
            exit;

        case 'get_car_files':
            // Get files for a specific car with permission checking
            if (!isset($postData['car_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'car_id is required']);
                exit;
            }
            
            $carId = intval($postData['car_id']);
            $userId = isset($postData['user_id']) ? intval($postData['user_id']) : null;
            $isAdmin = isset($postData['is_admin']) ? (bool)$postData['is_admin'] : false;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check if client_id column exists
                $hasClientIdColumn = false;
                try {
                    $checkColumnQuery = "SELECT COUNT(*) as cnt FROM INFORMATION_SCHEMA.COLUMNS 
                        WHERE TABLE_SCHEMA = DATABASE() 
                        AND TABLE_NAME = 'car_file_physical_tracking' 
                        AND COLUMN_NAME = 'client_id'";
                    $stmt = $conn->prepare($checkColumnQuery);
                    $stmt->execute();
                    $result = $stmt->fetch(PDO::FETCH_ASSOC);
                    $hasClientIdColumn = ($result && $result['cnt'] > 0);
                } catch (Exception $e) {
                    // If check fails, assume column doesn't exist
                    $hasClientIdColumn = false;
                }
                
                // Build query with permission filtering
                $clientFields = $hasClientIdColumn 
                    ? "cpt.client_id,
                        cl.name as client_name,
                        cl.email as client_email,
                        cl.mobiles as client_contact,"
                    : "NULL as client_id,
                        NULL as client_name,
                        NULL as client_email,
                        NULL as client_contact,";
                
                $clientJoin = $hasClientIdColumn 
                    ? "LEFT JOIN clients cl ON cpt.client_id = cl.id"
                    : "";
                
                $query = "
                    SELECT 
                        cf.*,
                        cfc.category_name,
                        cfc.importance_level,
                        cfc.is_required,
                        cfc.display_order,
                        cfc.visibility_scope as category_visibility,
                        u.username as uploaded_by_username,
                        cpt.id as tracking_id,
                        cpt.current_holder_id,
                        u_holder.username as current_holder_username,
                        cpt.custom_clearance_agent_id,
                        cca.name as agent_name,
                        cpt.checkout_type,
                        {$clientFields}
                        cpt.status as physical_status,
                        cpt.checked_out_at,
                        cpt.checked_in_at,
                        cpt.expected_return_date
                    FROM car_files cf
                    INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
                    LEFT JOIN users u ON cf.uploaded_by = u.id
                    LEFT JOIN (
                        SELECT cpt1.*
                        FROM car_file_physical_tracking cpt1
                        INNER JOIN (
                            SELECT car_file_id, MAX(id) as max_id
                            FROM car_file_physical_tracking
                            WHERE status IN ('available', 'checked_out')
                            GROUP BY car_file_id
                        ) cpt2 ON cpt1.car_file_id = cpt2.car_file_id AND cpt1.id = cpt2.max_id
                        WHERE cpt1.status IN ('available', 'checked_out')
                    ) cpt ON cf.id = cpt.car_file_id
                    LEFT JOIN users u_holder ON cpt.current_holder_id = u_holder.id
                    LEFT JOIN custom_clearance_agents cca ON cpt.custom_clearance_agent_id = cca.id
                    {$clientJoin}
                    WHERE cf.car_id = ? AND cf.is_active = 1
                ";
                
                $params = [$carId];
                
                // Add permission filtering if not admin
                if (!$isAdmin && $userId) {
                    $query .= " AND (
                        cf.uploaded_by = ? OR
                        cpt.current_holder_id = ? OR
                        cf.visibility_scope = 'public' OR
                        (cf.visibility_scope IS NULL AND cfc.visibility_scope = 'public')
                    )";
                    $params[] = $userId;
                    $params[] = $userId;
                }
                
                $query .= " ORDER BY cfc.display_order ASC, cfc.importance_level ASC, cf.uploaded_at DESC";
                
                $stmt = $conn->prepare($query);
                $stmt->execute($params);
                $files = $stmt->fetchAll(PDO::FETCH_ASSOC);
                
                echo json_encode(['success' => true, 'data' => $files]);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'create_car_file':
            // Create a new file record (after upload)
            // Automatically assigns the uploader as the first owner
            if (!isset($postData['car_id']) || !isset($postData['category_id']) || 
                !isset($postData['file_path']) || !isset($postData['file_name']) ||
                !isset($postData['uploaded_by'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Missing required fields']);
                exit;
            }
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                $uploadedBy = intval($postData['uploaded_by']);
                
                // Insert file record
                $query = "INSERT INTO car_files 
                    (car_id, category_id, file_path, file_name, file_size, file_type, uploaded_by, notes, visibility_scope)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)";
                
                $params = [
                    intval($postData['car_id']),
                    intval($postData['category_id']),
                    $postData['file_path'],
                    $postData['file_name'],
                    isset($postData['file_size']) ? intval($postData['file_size']) : null,
                    isset($postData['file_type']) ? $postData['file_type'] : null,
                    $uploadedBy,
                    isset($postData['notes']) ? $postData['notes'] : null,
                    isset($postData['visibility_scope']) ? $postData['visibility_scope'] : null
                ];
                
                $stmt = $conn->prepare($query);
                $stmt->execute($params);
                $fileId = $conn->lastInsertId();
                
                // Create tracking record - file starts as available (not checked out)
                // The file is available until someone checks it out
                $trackingQuery = "INSERT INTO car_file_physical_tracking 
                    (car_file_id, current_holder_id, status, checkout_type, transfer_notes)
                    VALUES (?, NULL, 'available', 'user', 'File uploaded - available for checkout')";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'lastInsertId' => $fileId]);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'delete_car_file':
            // Soft delete a file (only uploader or admin can delete)
            if (!isset($postData['file_id']) || !isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and user_id are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            $userId = intval($postData['user_id']);
            $isAdmin = isset($postData['is_admin']) ? (bool)$postData['is_admin'] : false;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check if user can delete (uploader or admin)
                $checkQuery = "SELECT uploaded_by FROM car_files WHERE id = ? AND is_active = 1";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $file = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$file) {
                    throw new Exception('File not found');
                }
                
                if (!$isAdmin && $file['uploaded_by'] != $userId) {
                    throw new Exception('Permission denied: Only the uploader or admin can delete this file');
                }
                
                // Soft delete
                $updateQuery = "UPDATE car_files SET is_active = 0 WHERE id = ?";
                $result = executeQuery($updateQuery, [$fileId]);
                echo json_encode($result);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'checkout_physical_copy':
            // Check out a physical copy of a file (to user, client, or custom clearance agent)
            if (!isset($postData['file_id']) || !isset($postData['checkout_type'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and checkout_type are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            $checkoutType = $postData['checkout_type']; // 'user', 'client', or 'custom_clearance_agent'
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            $userId = isset($postData['user_id']) ? intval($postData['user_id']) : null;
            $agentId = isset($postData['agent_id']) ? intval($postData['agent_id']) : null;
            $clientId = isset($postData['client_id']) ? intval($postData['client_id']) : null;
            $clientName = isset($postData['client_name']) ? trim($postData['client_name']) : null;
            $clientContact = isset($postData['client_contact']) ? trim($postData['client_contact']) : null;
            
            // Validate based on checkout type
            if ($checkoutType === 'user' && !$userId) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'user_id is required for user checkout']);
                exit;
            }
            if ($checkoutType === 'custom_clearance_agent' && !$agentId) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'agent_id is required for agent checkout']);
                exit;
            }
            if ($checkoutType === 'client' && !$clientId) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'client_id is required for client checkout']);
                exit;
            }
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Check if file exists
                $checkQuery = "SELECT id, uploaded_by FROM car_files WHERE id = ? AND is_active = 1";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $file = $stmt->fetch(PDO::FETCH_ASSOC);
                if (!$file) {
                    throw new Exception('File not found or inactive');
                }
                
                // Get current tracking
                $trackingQuery = "SELECT id, current_holder_id, custom_clearance_agent_id, checkout_type, client_id 
                    FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out'";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId]);
                $existing = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if ($existing) {
                    throw new Exception('File is already checked out. Please check it in first.');
                }
                
                // Get previous holder info for transfer history (from previous checkout if any)
                $fromUserId = null;
                $fromAgentId = null;
                $fromClientName = null;
                // Note: $existing will be null here since we throw if it exists, but keeping for future use
                
                // Determine who is performing the checkout (current user)
                $performedBy = isset($postData['performed_by']) ? intval($postData['performed_by']) : $file['uploaded_by'];
                
                // Create or update tracking record
                $insertQuery = "INSERT INTO car_file_physical_tracking 
                    (car_file_id, current_holder_id, custom_clearance_agent_id, checkout_type, 
                     client_id, checked_out_at, status, transfer_notes)
                    VALUES (?, ?, ?, ?, ?, NOW(), 'checked_out', ?)";
                $stmt = $conn->prepare($insertQuery);
                $stmt->execute([
                    $fileId, 
                    $checkoutType === 'user' ? $userId : null,
                    $checkoutType === 'custom_clearance_agent' ? $agentId : null,
                    $checkoutType,
                    $checkoutType === 'client' ? $clientId : null,
                    $notes
                ]);
                
                // Determine transfer type and create transfer history record
                $transferType = 'user_to_user';
                $toUserId = null;
                $toAgentId = null;
                $toClientName = null;
                
                if ($checkoutType === 'user') {
                    $toUserId = $userId;
                    $transferType = 'user_to_user';
                } elseif ($checkoutType === 'custom_clearance_agent') {
                    $toAgentId = $agentId;
                    $transferType = 'user_to_agent';
                } elseif ($checkoutType === 'client') {
                    // Get client name from database using client_id
                    if ($clientId) {
                        $clientNameQuery = "SELECT name FROM clients WHERE id = ?";
                        $stmt = $conn->prepare($clientNameQuery);
                        $stmt->execute([$clientId]);
                        $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                        $toClientName = $clientData ? $clientData['name'] : null;
                    }
                    $transferType = 'user_to_client';
                }
                
                // Insert transfer record - to_user_id can be NULL when checking out to agent or client
                $transferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, from_agent_id, to_user_id, to_agent_id, 
                     from_client_name, to_client_name, transferred_by, notes, transfer_type)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([
                    $fileId, 
                    $fromUserId, 
                    $fromAgentId, 
                    $toUserId,  // NULL for agent/client checkout
                    $toAgentId,
                    $fromClientName, 
                    $toClientName, 
                    $performedBy,
                    $notes ?: 'Checked out', 
                    $transferType
                ]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'File checked out successfully']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'rollback_checkout':
            // Rollback a checkout operation (admin only)
            if (!isset($postData['file_id']) || !isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and user_id are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            $userId = intval($postData['user_id']); // Admin user ID
            $notes = isset($postData['notes']) ? $postData['notes'] : 'Rollback checkout (admin)';
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Verify user is admin
                $adminCheck = "SELECT role_id FROM users WHERE id = ?";
                $stmt = $conn->prepare($adminCheck);
                $stmt->execute([$userId]);
                $user = $stmt->fetch(PDO::FETCH_ASSOC);
                if (!$user || $user['role_id'] != 1) {
                    throw new Exception('Only administrators can rollback checkouts');
                }
                
                $conn->beginTransaction();
                
                // Get current tracking record
                $trackingQuery = "SELECT id, status FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out'";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File is not currently checked out');
                }
                
                // Get the most recent transfer record ID before deleting
                $getTransferQuery = "SELECT id FROM car_file_transfers 
                    WHERE car_file_id = ? 
                    ORDER BY transferred_at DESC 
                    LIMIT 1";
                $stmt = $conn->prepare($getTransferQuery);
                $stmt->execute([$fileId]);
                $transferRecord = $stmt->fetch(PDO::FETCH_ASSOC);
                
                // Delete the checkout tracking record
                $deleteTrackingQuery = "DELETE FROM car_file_physical_tracking WHERE id = ?";
                $stmt = $conn->prepare($deleteTrackingQuery);
                $stmt->execute([$tracking['id']]);
                
                // Delete the most recent transfer record (the checkout record) if it exists
                if ($transferRecord) {
                    $deleteTransferQuery = "DELETE FROM car_file_transfers WHERE id = ?";
                    $stmt = $conn->prepare($deleteTransferQuery);
                    $stmt->execute([$transferRecord['id']]);
                }
                
                // Create a new available tracking record
                $insertTrackingQuery = "INSERT INTO car_file_physical_tracking 
                    (car_file_id, current_holder_id, status, checkout_type, transfer_notes)
                    VALUES (?, NULL, 'available', 'user', ?)";
                $stmt = $conn->prepare($insertTrackingQuery);
                $stmt->execute([$fileId, $notes]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'Checkout rolled back successfully']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'checkin_physical_copy':
            // Check in a physical copy
            if (!isset($postData['file_id']) || !isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and user_id are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            $userId = intval($postData['user_id']);
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check if user is the current holder
                $checkQuery = "SELECT id, current_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out' AND current_holder_id = ?";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId, $userId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File is not checked out to you');
                }
                
                // Get current holder before checkin for transfer history
                $currentHolderId = $tracking['current_holder_id'];
                
                // Update tracking
                $updateQuery = "UPDATE car_file_physical_tracking 
                    SET previous_holder_id = current_holder_id,
                        current_holder_id = NULL, checked_in_at = NOW(), 
                        status = 'available', transfer_notes = ?
                    WHERE id = ?";
                $stmt = $conn->prepare($updateQuery);
                $stmt->execute([$notes, $tracking['id']]);
                
                // Update the last transfer record to mark it as returned (if exists)
                // This marks the transfer that gave the file to the current holder as returned
                // We also update transferred_by to show who checked it in
                // We need to update the most recent transfer where this user received the file
                $updateLastTransferQuery = "UPDATE car_file_transfers 
                    SET returned_at = NOW(), return_notes = ?, transferred_by = ?
                    WHERE car_file_id = ? AND to_user_id = ? AND returned_at IS NULL
                    ORDER BY transferred_at DESC LIMIT 1";
                $stmt = $conn->prepare($updateLastTransferQuery);
                $updateResult = $stmt->execute([$notes, $userId, $fileId, $currentHolderId]);
                
                // If no record was updated (shouldn't happen, but just in case), log it
                if ($stmt->rowCount() === 0) {
                    error_log("Warning: No transfer record found to update for check-in. File ID: $fileId, User ID: $currentHolderId");
                }
                
                echo json_encode(['success' => true, 'message' => 'File checked in successfully']);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'transfer_physical_copy':
            // Transfer physical copy from one user to another
            if (!isset($postData['file_id']) || !isset($postData['from_user_id']) || 
                !isset($postData['to_user_id']) || !isset($postData['transferred_by'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Missing required fields']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            $fromUserId = intval($postData['from_user_id']);
            $toUserId = intval($postData['to_user_id']);
            $transferredBy = intval($postData['transferred_by']);
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            $expectedReturnDate = isset($postData['expected_return_date']) ? $postData['expected_return_date'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Verify current holder
                $checkQuery = "SELECT id, current_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out'";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File is not checked out');
                }
                
                if ($tracking['current_holder_id'] != $fromUserId) {
                    throw new Exception('File is not currently held by the specified user');
                }
                
                // Update tracking
                $updateQuery = "UPDATE car_file_physical_tracking 
                    SET previous_holder_id = ?, current_holder_id = ?, 
                        transferred_by = ?, transferred_at = NOW(),
                        expected_return_date = ?, transfer_notes = ?
                    WHERE id = ?";
                $stmt = $conn->prepare($updateQuery);
                $stmt->execute([
                    $fromUserId, $toUserId, $transferredBy, 
                    $expectedReturnDate, $notes, $tracking['id']
                ]);
                
                // Create transfer history record
                $transferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, to_user_id, transferred_by, notes, return_expected_date)
                    VALUES (?, ?, ?, ?, ?, ?)";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([
                    $fileId, $fromUserId, $toUserId, $transferredBy, 
                    $notes, $expectedReturnDate
                ]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'File transferred successfully']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'get_file_transfer_history':
            // Get transfer history for a file
            if (!isset($postData['file_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id is required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            
            $query = "SELECT 
                cft.*,
                u_from.username as from_username,
                u_to.username as to_username,
                u_transfer.username as transferred_by_username,
                cft.returned_at,
                cft.return_notes
                FROM car_file_transfers cft
                LEFT JOIN users u_from ON cft.from_user_id = u_from.id
                INNER JOIN users u_to ON cft.to_user_id = u_to.id
                INNER JOIN users u_transfer ON cft.transferred_by = u_transfer.id
                WHERE cft.car_file_id = ?
                ORDER BY cft.transferred_at DESC";
            
            $result = executeQuery($query, [$fileId]);
            echo json_encode($result);
            exit;

        case 'create_file_category':
            // Create a new file category (admin only)
            if (!isset($postData['category_name']) || !isset($postData['is_admin']) || !$postData['is_admin']) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Admin access required']);
                exit;
            }
            
            $query = "INSERT INTO car_file_categories 
                (category_name, importance_level, is_required, display_order, description)
                VALUES (?, ?, ?, ?, ?)";
            
            $params = [
                $postData['category_name'],
                isset($postData['importance_level']) ? intval($postData['importance_level']) : 3,
                isset($postData['is_required']) ? (int)$postData['is_required'] : 0,
                isset($postData['display_order']) ? intval($postData['display_order']) : 0,
                isset($postData['description']) ? $postData['description'] : null
            ];
            
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'update_file_category':
            // Update a file category (admin only)
            if (!isset($postData['category_id']) || !isset($postData['is_admin']) || !$postData['is_admin']) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Admin access required']);
                exit;
            }
            
            $categoryId = intval($postData['category_id']);
            $updates = [];
            $params = [];
            
            if (isset($postData['category_name'])) {
                $updates[] = "category_name = ?";
                $params[] = $postData['category_name'];
            }
            if (isset($postData['importance_level'])) {
                $updates[] = "importance_level = ?";
                $params[] = intval($postData['importance_level']);
            }
            if (isset($postData['is_required'])) {
                $updates[] = "is_required = ?";
                $params[] = (int)$postData['is_required'];
            }
            if (isset($postData['display_order'])) {
                $updates[] = "display_order = ?";
                $params[] = intval($postData['display_order']);
            }
            if (isset($postData['description'])) {
                $updates[] = "description = ?";
                $params[] = $postData['description'];
            }
            
            if (empty($updates)) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'No fields to update']);
                exit;
            }
            
            $params[] = $categoryId;
            $query = "UPDATE car_file_categories SET " . implode(', ', $updates) . " WHERE id = ?";
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'get_my_physical_copies':
            // Get all physical copies currently held by a user
            if (!isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'user_id is required']);
                exit;
            }
            
            $userId = intval($postData['user_id']);
            
            $query = "SELECT 
                cf.*,
                cfc.category_name,
                cfc.importance_level,
                cs.vin,
                cpt.id as tracking_id,
                cpt.checked_out_at,
                cpt.expected_return_date,
                cpt.transfer_notes
                FROM car_file_physical_tracking cpt
                INNER JOIN car_files cf ON cpt.car_file_id = cf.id
                INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
                INNER JOIN cars_stock cs ON cf.car_id = cs.id
                WHERE cpt.current_holder_id = ? AND cpt.status = 'checked_out'
                ORDER BY cpt.checked_out_at DESC";
            
            $result = executeQuery($query, [$userId]);
            echo json_encode($result);
            exit;

        case 'get_users_for_transfer':
            // Get list of users for transfer dropdown (exclude current holder)
            if (!isset($postData['file_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id is required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Get current holder
                $holderQuery = "SELECT current_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out'";
                $stmt = $conn->prepare($holderQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                $currentHolderId = $tracking ? $tracking['current_holder_id'] : null;
                
                // Get all users except current holder
                $query = "SELECT id, username, email FROM users WHERE id != ? ORDER BY username ASC";
                $params = $currentHolderId ? [$currentHolderId] : [0];
                
                $result = executeQuery($query, $params);
                echo json_encode($result);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'get_custom_clearance_agents':
            // Get all active custom clearance agents
            $query = "SELECT * FROM custom_clearance_agents WHERE is_active = 1 ORDER BY name ASC";
            $result = executeQuery($query);
            echo json_encode($result);
            exit;

        case 'create_custom_clearance_agent':
            // Create a new custom clearance agent (admin only)
            if (!isset($postData['is_admin']) || !$postData['is_admin']) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Admin access required']);
                exit;
            }

            $name = isset($postData['name']) ? trim($postData['name']) : '';
            if (empty($name)) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Name is required']);
                exit;
            }

            $query = "INSERT INTO custom_clearance_agents 
                (name, contact_person, phone, email, address, license_number, notes, is_active)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)";
            
            $params = [
                $name,
                isset($postData['contact_person']) ? $postData['contact_person'] : null,
                isset($postData['phone']) ? $postData['phone'] : null,
                isset($postData['email']) ? $postData['email'] : null,
                isset($postData['address']) ? $postData['address'] : null,
                isset($postData['license_number']) ? $postData['license_number'] : null,
                isset($postData['notes']) ? $postData['notes'] : null,
                isset($postData['is_active']) ? (int)$postData['is_active'] : 1
            ];
            
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'update_custom_clearance_agent':
            // Update a custom clearance agent (admin only)
            if (!isset($postData['is_admin']) || !$postData['is_admin']) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Admin access required']);
                exit;
            }

            if (!isset($postData['id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Agent ID is required']);
                exit;
            }

            $agentId = intval($postData['id']);
            $updates = [];
            $params = [];

            if (isset($postData['name'])) {
                $updates[] = "name = ?";
                $params[] = trim($postData['name']);
            }
            if (isset($postData['contact_person'])) {
                $updates[] = "contact_person = ?";
                $params[] = $postData['contact_person'];
            }
            if (isset($postData['phone'])) {
                $updates[] = "phone = ?";
                $params[] = $postData['phone'];
            }
            if (isset($postData['email'])) {
                $updates[] = "email = ?";
                $params[] = $postData['email'];
            }
            if (isset($postData['address'])) {
                $updates[] = "address = ?";
                $params[] = $postData['address'];
            }
            if (isset($postData['license_number'])) {
                $updates[] = "license_number = ?";
                $params[] = $postData['license_number'];
            }
            if (isset($postData['notes'])) {
                $updates[] = "notes = ?";
                $params[] = $postData['notes'];
            }
            if (isset($postData['is_active'])) {
                $updates[] = "is_active = ?";
                $params[] = (int)$postData['is_active'];
            }

            if (empty($updates)) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'No fields to update']);
                exit;
            }

            $params[] = $agentId;
            $query = "UPDATE custom_clearance_agents SET " . implode(', ', $updates) . " WHERE id = ?";
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'delete_custom_clearance_agent':
            // Soft delete a custom clearance agent (admin only)
            if (!isset($postData['is_admin']) || !$postData['is_admin']) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Admin access required']);
                exit;
            }

            if (!isset($postData['id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Agent ID is required']);
                exit;
            }

            $agentId = intval($postData['id']);
            $query = "UPDATE custom_clearance_agents SET is_active = 0 WHERE id = ?";
            $result = executeQuery($query, [$agentId]);
            echo json_encode($result);
            exit;
    
        default:
            http_response_code(400);
            echo json_encode(['success' => false, 'error' => 'Invalid action']);
            exit;
    }
}

// Regular query handling
if (!isset($postData['query']) || empty($postData['query'])) {
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Query is required']);
    exit;
}

// Get query and parameters
$query = $postData['query'];
$params = isset($postData['params']) ? $postData['params'] : [];

// Execute query and return result
http_response_code(200);
echo json_encode(executeQuery($query, $params));
exit;
?>