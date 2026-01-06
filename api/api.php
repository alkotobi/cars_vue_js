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

// Function to check if a user has a specific permission
function hasPermission($conn, $userId, $permissionName) {
    try {
        // Admins have all permissions
        $adminCheck = "SELECT role_id FROM users WHERE id = ?";
        $stmt = $conn->prepare($adminCheck);
        $stmt->execute([$userId]);
        $user = $stmt->fetch(PDO::FETCH_ASSOC);
        
        if ($user && $user['role_id'] == 1) {
            return true; // Admin has all permissions
        }
        
        // Check if user's role has the permission
        $permissionQuery = "SELECT COUNT(*) as cnt 
            FROM role_permissions rp
            INNER JOIN permissions p ON rp.permission_id = p.id
            INNER JOIN users u ON rp.role_id = u.role_id
            WHERE u.id = ? AND p.permission_name = ?";
        $stmt = $conn->prepare($permissionQuery);
        $stmt->execute([$userId, $permissionName]);
        $result = $stmt->fetch(PDO::FETCH_ASSOC);
        
        return ($result && $result['cnt'] > 0);
    } catch (Exception $e) {
        error_log("Error checking permission: " . $e->getMessage());
        return false;
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
            if (!isset($postData['car_id']) || !isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'car_id and user_id are required']);
                exit;
            }
            
            $carId = intval($postData['car_id']);
            $userId = intval($postData['user_id']);
            $isAdmin = isset($postData['is_admin']) ? (bool)$postData['is_admin'] : false;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check permission: can_upload_car_files or admin
                if (!$isAdmin && !hasPermission($conn, $userId, 'can_upload_car_files')) {
                    http_response_code(403);
                    echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to view car files']);
                    exit;
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
                        cpt.previous_holder_id,
                        u_prev.username as previous_holder_username,
                        cpt.custom_clearance_agent_id,
                        cca.name as agent_name,
                        cpt.checkout_type,
                        {$clientFields}
                        cpt.status as physical_status,
                        cpt.checked_out_at,
                        cpt.checked_in_at,
                        cpt.expected_return_date,
                        CASE WHEN pending_transfer.id IS NOT NULL THEN 1 ELSE 0 END as has_pending_transfer,
                        pending_transfer.id as pending_transfer_id,
                        pending_transfer.to_user_id as pending_transfer_to_user_id
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
                    LEFT JOIN users u_prev ON cpt.previous_holder_id = u_prev.id
                    LEFT JOIN custom_clearance_agents cca ON cpt.custom_clearance_agent_id = cca.id
                    LEFT JOIN (
                        SELECT id, car_file_id, to_user_id
                        FROM car_file_transfers
                        WHERE transfer_status = 'pending'
                    ) pending_transfer ON cf.id = pending_transfer.car_file_id
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
                $carId = intval($postData['car_id']);
                $categoryId = intval($postData['category_id']);
                
                // Check if category already has an active file for this car
                $checkQuery = "SELECT id, file_path FROM car_files 
                    WHERE car_id = ? AND category_id = ? AND is_active = 1
                    LIMIT 1";
                $checkStmt = $conn->prepare($checkQuery);
                $checkStmt->execute([$carId, $categoryId]);
                $existingFile = $checkStmt->fetch(PDO::FETCH_ASSOC);
                
                // If file exists, soft delete it first (replace scenario)
                if ($existingFile) {
                    $existingFileId = $existingFile['id'];
                    $existingFilePath = $existingFile['file_path'];
                    
                    // Get base directory for file deletion
                    $baseDirectory = 'mig_files'; // Default
                    try {
                        $configQuery = "SELECT js_dir FROM dbs LIMIT 1";
                        $configStmt = $conn->prepare($configQuery);
                        $configStmt->execute();
                        $dbConfig = $configStmt->fetch(PDO::FETCH_ASSOC);
                        if ($dbConfig && isset($dbConfig['js_dir'])) {
                            $jsDir = $dbConfig['js_dir'];
                            $dbCodeJsonPath = __DIR__ . '/../' . $jsDir . '/db_code.json';
                            if (file_exists($dbCodeJsonPath)) {
                                $dbCodeContent = file_get_contents($dbCodeJsonPath);
                                $dbCodeData = json_decode($dbCodeContent, true);
                                if ($dbCodeData && isset($dbCodeData['files_dir'])) {
                                    $baseDirectory = $dbCodeData['files_dir'];
                                }
                            }
                        }
                    } catch (Exception $e) {
                        error_log('Could not get base directory: ' . $e->getMessage());
                    }
                    
                    // Delete physical file if it exists
                    $baseDirectory = str_replace('..', '', $baseDirectory);
                    $baseDirectory = ltrim($baseDirectory, '/');
                    $baseDirectory = rtrim($baseDirectory, '/');
                    $existingFilePath = str_replace('..', '', $existingFilePath);
                    $existingFilePath = ltrim($existingFilePath, '/');
                    $fullFilePath = __DIR__ . '/../' . $baseDirectory . '/' . $existingFilePath;
                    
                    if (file_exists($fullFilePath)) {
                        @unlink($fullFilePath);
                    }
                    
                    // Soft delete the existing file record
                    $deleteQuery = "UPDATE car_files SET is_active = 0 WHERE id = ?";
                    $deleteStmt = $conn->prepare($deleteQuery);
                    $deleteStmt->execute([$existingFileId]);
                }
                
                // Insert file record
                $query = "INSERT INTO car_files 
                    (car_id, category_id, file_path, file_name, file_size, file_type, uploaded_by, notes, visibility_scope)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)";
                
                $params = [
                    $carId,
                    $categoryId,
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
                
                // Create tracking record - uploader becomes the current holder
                $trackingQuery = "INSERT INTO car_file_physical_tracking 
                    (car_file_id, current_holder_id, status, checkout_type, checked_out_at, transfer_notes)
                    VALUES (?, ?, 'checked_out', 'user', NOW(), 'File uploaded - uploader is initial holder')";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId, $uploadedBy]);
                
                // Create initial transfer record showing uploader as holder
                $initialTransferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, to_user_id, transferred_by, notes, transfer_status, transfer_type)
                    VALUES (?, NULL, ?, ?, 'File uploaded - uploader is initial holder', 'approved', 'user_to_user')";
                $stmt = $conn->prepare($initialTransferQuery);
                $stmt->execute([$fileId, $uploadedBy, $uploadedBy]);
                
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
            // Delete a file (admin only) - deletes both database record and physical file
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
                
                // Only admin can delete
                if (!$isAdmin) {
                    http_response_code(403);
                    throw new Exception('Permission denied: Only admin can delete files');
                }
                
                // Get file info including file_path
                $checkQuery = "SELECT id, file_path, uploaded_by FROM car_files WHERE id = ? AND is_active = 1";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $file = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$file) {
                    throw new Exception('File not found');
                }
                
                $filePath = $file['file_path'];
                
                // Get base directory from db_code.json or use default
                $baseDirectory = 'mig_files'; // Default
                try {
                    $configQuery = "SELECT js_dir FROM dbs LIMIT 1";
                    $configStmt = $conn->prepare($configQuery);
                    $configStmt->execute();
                    $dbConfig = $configStmt->fetch(PDO::FETCH_ASSOC);
                    if ($dbConfig && isset($dbConfig['js_dir'])) {
                        // Try to read db_code.json to get files_dir
                        $jsDir = $dbConfig['js_dir'];
                        $dbCodeJsonPath = __DIR__ . '/../' . $jsDir . '/db_code.json';
                        if (file_exists($dbCodeJsonPath)) {
                            $dbCodeContent = file_get_contents($dbCodeJsonPath);
                            $dbCodeData = json_decode($dbCodeContent, true);
                            if ($dbCodeData && isset($dbCodeData['files_dir'])) {
                                $baseDirectory = $dbCodeData['files_dir'];
                            }
                        }
                    }
                } catch (Exception $e) {
                    // Use default if we can't get it
                    error_log('Could not get base directory from db_code.json: ' . $e->getMessage());
                }
                
                // Construct full file path
                $baseDirectory = str_replace('..', '', $baseDirectory);
                $baseDirectory = ltrim($baseDirectory, '/');
                $baseDirectory = rtrim($baseDirectory, '/');
                
                $filePath = str_replace('..', '', $filePath);
                $filePath = ltrim($filePath, '/');
                
                $fullFilePath = __DIR__ . '/../' . $baseDirectory . '/' . $filePath;
                
                // Begin transaction
                $conn->beginTransaction();
                
                // Delete physical file if it exists
                $fileDeleted = false;
                if (file_exists($fullFilePath)) {
                    if (unlink($fullFilePath)) {
                        $fileDeleted = true;
                        error_log("Deleted physical file: $fullFilePath");
                    } else {
                        error_log("Failed to delete physical file: $fullFilePath");
                        // Continue with database deletion even if file deletion fails
                    }
                } else {
                    error_log("Physical file not found: $fullFilePath");
                    // Continue with database deletion even if file doesn't exist
                }
                
                // Soft delete from database
                $updateQuery = "UPDATE car_files SET is_active = 0 WHERE id = ?";
                $stmt = $conn->prepare($updateQuery);
                $stmt->execute([$fileId]);
                
                $conn->commit();
                
                echo json_encode([
                    'success' => true,
                    'affectedRows' => $stmt->rowCount(),
                    'fileDeleted' => $fileDeleted,
                    'message' => $fileDeleted ? 'File and database record deleted successfully' : 'Database record deleted (file was not found on server)'
                ]);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
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
                
                // Get current tracking - file must be checked out to someone
                $trackingQuery = "SELECT id, current_holder_id, custom_clearance_agent_id, checkout_type, client_id, status
                    FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status IN ('available', 'checked_out')";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File tracking record not found');
                }
                
                // Verify that the current user is the current holder
                $performedBy = isset($postData['performed_by']) ? intval($postData['performed_by']) : null;
                if (!$performedBy) {
                    throw new Exception('performed_by is required');
                }
                
                // Check if file is available (shouldn't happen after upload, but handle it)
                if ($tracking['status'] === 'available') {
                    // If available, only the uploader can checkout
                    if ($tracking['current_holder_id'] != $performedBy && $file['uploaded_by'] != $performedBy) {
                        throw new Exception('Only the current holder can checkout this file');
                    }
                } else {
                    // File is checked out - verify current holder
                    if ($tracking['current_holder_id'] != $performedBy) {
                        throw new Exception('Only the current holder can checkout this file');
                    }
                }
                
                // Get previous holder info for transfer history
                $fromUserId = $tracking['current_holder_id'];
                $fromAgentId = $tracking['custom_clearance_agent_id'];
                $fromClientName = null;
                if ($tracking['client_id']) {
                    $clientQuery = "SELECT name FROM clients WHERE id = ?";
                    $stmt = $conn->prepare($clientQuery);
                    $stmt->execute([$tracking['client_id']]);
                    $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                    $fromClientName = $clientData ? $clientData['name'] : null;
                }
                
                // Update tracking record
                // When checking out to user: current_holder_id changes to that user
                // When checking out to client/transiteur: current_holder_id becomes NULL
                // The previous_holder_id stores who checked it out (so they can rollback)
                $newHolderId = null;
                if ($checkoutType === 'user') {
                    $newHolderId = $userId;  // User becomes the holder
                } else {
                    // For client/transiteur checkout, set current_holder_id to NULL
                    // previous_holder_id will store who checked it out (for rollback)
                    $newHolderId = null;
                }
                
                $updateQuery = "UPDATE car_file_physical_tracking 
                    SET previous_holder_id = current_holder_id,
                        current_holder_id = ?,
                        custom_clearance_agent_id = ?,
                        checkout_type = ?,
                        client_id = ?,
                        checked_out_at = NOW(),
                        status = 'checked_out',
                        transfer_notes = ?
                    WHERE id = ?";
                $stmt = $conn->prepare($updateQuery);
                $stmt->execute([
                    $newHolderId,
                    $checkoutType === 'custom_clearance_agent' ? $agentId : null,
                    $checkoutType,
                    $checkoutType === 'client' ? $clientId : null,
                    $notes,
                    $tracking['id']
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
                
                // Insert transfer record for checkout - to_user_id can be NULL when checking out to agent or client
                $checkoutNotes = 'Checkout: ' . ($notes ?: 'Checked out');
                if ($checkoutType === 'user') {
                    $checkoutNotes = 'Checkout to user: ' . ($notes ?: 'Checked out to user');
                } elseif ($checkoutType === 'custom_clearance_agent') {
                    $checkoutNotes = 'Checkout to transiteur: ' . ($notes ?: 'Checked out to transiteur');
                } elseif ($checkoutType === 'client') {
                    $checkoutNotes = 'Checkout to client: ' . ($notes ?: 'Checked out to client');
                }
                
                $transferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, from_agent_id, to_user_id, to_agent_id, 
                     from_client_name, to_client_name, transferred_by, notes, transfer_type, transfer_status)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'approved')";
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
                    $checkoutNotes, 
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
            // Rollback a checkout/transfer operation (admin only, and only for transfers/checkouts, not initial upload)
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
                
                $conn->beginTransaction();
                
                // Get current tracking record
                $trackingQuery = "SELECT id, status, current_holder_id, previous_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out'";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File is not currently checked out');
                }
                
                // Verify user can rollback:
                // - If current_holder_id is set: must be the current holder
                // - If current_holder_id is NULL (checked out to client/transiteur): must be the previous holder
                $canRollback = false;
                if ($tracking['current_holder_id'] && $tracking['current_holder_id'] == $userId) {
                    $canRollback = true;  // Current holder
                } elseif (!$tracking['current_holder_id'] && $tracking['previous_holder_id'] == $userId) {
                    $canRollback = true;  // Previous holder (when checked out to client/transiteur)
                }
                
                if (!$canRollback) {
                    throw new Exception('Only the current holder or previous holder (if checked out to client/transiteur) can rollback the checkout');
                }
                
                // Check if this is the initial upload transfer - cannot rollback initial upload
                // The initial transfer has from_user_id = NULL (upload) or is the first transfer
                $transferCountQuery = "SELECT COUNT(*) as cnt FROM car_file_transfers 
                    WHERE car_file_id = ? AND transfer_status = 'approved'";
                $stmt = $conn->prepare($transferCountQuery);
                $stmt->execute([$fileId]);
                $transferCount = $stmt->fetch(PDO::FETCH_ASSOC)['cnt'];
                
                // Get the most recent transfer record to check what to rollback to
                $getTransferQuery = "SELECT id, from_user_id, to_user_id, from_agent_id, to_agent_id, 
                    from_client_name, to_client_name, transfer_type, notes 
                    FROM car_file_transfers 
                    WHERE car_file_id = ? AND transfer_status = 'approved'
                    ORDER BY transferred_at DESC 
                    LIMIT 1";
                $stmt = $conn->prepare($getTransferQuery);
                $stmt->execute([$fileId]);
                $lastTransfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$lastTransfer) {
                    throw new Exception('No transfer record found to rollback');
                }
                
                // If this is the only transfer and it's the initial upload (from_user_id is NULL or notes contain "uploaded")
                if ($transferCount <= 1 && ($lastTransfer['from_user_id'] === null || 
                    stripos($lastTransfer['notes'], 'uploaded') !== false || 
                    stripos($lastTransfer['notes'], 'initial holder') !== false)) {
                    throw new Exception('Cannot rollback initial file upload. Use delete instead if you want to remove the file.');
                }
                
                if (!$lastTransfer) {
                    throw new Exception('No transfer record found to rollback');
                }
                
                // Determine the previous holder from the last transfer
                $previousHolderId = $lastTransfer['from_user_id'];
                $previousAgentId = $lastTransfer['from_agent_id'];
                $previousClientName = $lastTransfer['from_client_name'];
                
                // If from_user_id is NULL, this might be the initial upload - check notes
                if ($previousHolderId === null && (
                    stripos($lastTransfer['notes'], 'uploaded') !== false || 
                    stripos($lastTransfer['notes'], 'initial holder') !== false)) {
                    // This is the initial upload - cannot rollback
                    throw new Exception('Cannot rollback initial file upload. Use delete instead if you want to remove the file.');
                }
                
                // If no previous holder found, get it from the file's uploaded_by
                if ($previousHolderId === null && $previousAgentId === null) {
                    $fileQuery = "SELECT uploaded_by FROM car_files WHERE id = ?";
                    $stmt = $conn->prepare($fileQuery);
                    $stmt->execute([$fileId]);
                    $file = $stmt->fetch(PDO::FETCH_ASSOC);
                    $previousHolderId = $file ? $file['uploaded_by'] : null;
                }
                
                if ($previousHolderId === null && $previousAgentId === null && !$previousClientName) {
                    throw new Exception('Cannot determine previous holder. Cannot rollback.');
                }
                
                // Get current holder info for the rollback transfer record
                $currentHolderId = $tracking['current_holder_id'];
                $currentAgentId = null;
                $currentClientName = null;
                
                // Get current agent/client info if needed
                $currentTrackingQuery = "SELECT custom_clearance_agent_id, client_id, checkout_type 
                    FROM car_file_physical_tracking 
                    WHERE id = ?";
                $stmt = $conn->prepare($currentTrackingQuery);
                $stmt->execute([$tracking['id']]);
                $currentTracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if ($currentTracking) {
                    $currentAgentId = $currentTracking['custom_clearance_agent_id'];
                    if ($currentTracking['client_id']) {
                        $clientQuery = "SELECT name FROM clients WHERE id = ?";
                        $stmt = $conn->prepare($clientQuery);
                        $stmt->execute([$currentTracking['client_id']]);
                        $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                        $currentClientName = $clientData ? $clientData['name'] : null;
                    }
                }
                
                // Determine reverse transfer type
                $reverseTransferType = 'user_to_user';
                if ($lastTransfer['transfer_type'] === 'user_to_agent' || $lastTransfer['transfer_type'] === 'agent_to_user') {
                    $reverseTransferType = $currentAgentId ? 'agent_to_user' : 'user_to_agent';
                } elseif ($lastTransfer['transfer_type'] === 'user_to_client' || $lastTransfer['transfer_type'] === 'client_to_user') {
                    $reverseTransferType = $currentClientName ? 'client_to_user' : 'user_to_client';
                }
                
                // Add a new transfer record for rollback (reversing the last transfer)
                // This creates a history entry instead of deleting the last one
                // Get the username of who is performing the rollback
                $rollbackUserQuery = "SELECT username FROM users WHERE id = ?";
                $stmt = $conn->prepare($rollbackUserQuery);
                $stmt->execute([$userId]);
                $rollbackUser = $stmt->fetch(PDO::FETCH_ASSOC);
                $rollbackUsername = $rollbackUser ? $rollbackUser['username'] : 'Unknown User';
                
                $rollbackNotes = 'Rollback checkout';
                if ($lastTransfer['notes']) {
                    $rollbackNotes = 'Rollback checkout: Rolled back "' . $lastTransfer['notes'] . '"';
                }
                if ($notes) {
                    $rollbackNotes = 'Rollback checkout: ' . $notes;
                }
                // Add who performed it (the transferred_by field will also show this, but it's good to have in notes too)
                $rollbackNotes .= ' (by ' . $rollbackUsername . ')';
                
                $rollbackTransferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, to_user_id, from_agent_id, to_agent_id,
                     from_client_name, to_client_name, transferred_by, notes, transfer_type, transfer_status)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'approved')";
                $stmt = $conn->prepare($rollbackTransferQuery);
                $stmt->execute([
                    $fileId,
                    $currentHolderId,  // from: current holder (or NULL if checked out to client/transiteur)
                    $previousHolderId,  // to: previous holder
                    $currentAgentId,    // from agent
                    $previousAgentId,   // to agent
                    $currentClientName, // from client
                    $previousClientName, // to client
                    $userId,            // performed by current user
                    $rollbackNotes,
                    $reverseTransferType
                ]);
                
                // Update tracking to restore previous holder (file must always have a holder)
                $updateTrackingQuery = "UPDATE car_file_physical_tracking 
                    SET previous_holder_id = current_holder_id,
                        current_holder_id = ?,
                        custom_clearance_agent_id = ?,
                        client_id = ?,
                        checkout_type = ?,
                        checked_out_at = NOW(),
                        status = 'checked_out',
                        transfer_notes = ?
                    WHERE id = ?";
                
                // Get client_id if previous holder is a client
                $previousClientId = null;
                if ($previousClientName && !$previousHolderId && !$previousAgentId) {
                    $clientQuery = "SELECT id FROM clients WHERE name = ? LIMIT 1";
                    $stmt = $conn->prepare($clientQuery);
                    $stmt->execute([$previousClientName]);
                    $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                    $previousClientId = $clientData ? $clientData['id'] : null;
                }
                
                $checkoutType = 'user';
                if ($previousAgentId) {
                    $checkoutType = 'custom_clearance_agent';
                } elseif ($previousClientId || $previousClientName) {
                    $checkoutType = 'client';
                }
                
                $stmt = $conn->prepare($updateTrackingQuery);
                $stmt->execute([
                    $previousHolderId,
                    $previousAgentId,
                    $previousClientId,
                    $checkoutType,
                    $notes ?: 'Rollback: restored previous holder',
                    $tracking['id']
                ]);
                
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
            // Transfer physical copy from one user to another (creates pending transfer)
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
                
                // Check permission: can_upload_car_files or admin
                if (!hasPermission($conn, $transferredBy, 'can_upload_car_files')) {
                    http_response_code(403);
                    echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to transfer car files']);
                    exit;
                }
                
                $conn->beginTransaction();
                
                // Get current tracking status
                $checkQuery = "SELECT id, current_holder_id, status FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status IN ('available', 'checked_out')";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File tracking record not found');
                }
                
                // If file is available, allow transfer if uploader is transferring
                if ($tracking['status'] === 'available') {
                    // Verify the uploader is the one transferring
                    $fileQuery = "SELECT uploaded_by FROM car_files WHERE id = ?";
                    $stmt = $conn->prepare($fileQuery);
                    $stmt->execute([$fileId]);
                    $file = $stmt->fetch(PDO::FETCH_ASSOC);
                    
                    if (!$file || $file['uploaded_by'] != $fromUserId) {
                        throw new Exception('Only the uploader can transfer an available file');
                    }
                } else {
                    // File is checked out - verify current holder
                    if ($tracking['current_holder_id'] != $fromUserId) {
                        throw new Exception('File is not currently held by the specified user');
                    }
                }
                
                // Check if there's already a pending transfer for this file
                $pendingCheckQuery = "SELECT id FROM car_file_transfers 
                    WHERE car_file_id = ? AND transfer_status = 'pending'";
                $stmt = $conn->prepare($pendingCheckQuery);
                $stmt->execute([$fileId]);
                $pendingTransfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if ($pendingTransfer) {
                    throw new Exception('There is already a pending transfer for this file');
                }
                
                // Create pending transfer record (DO NOT update tracking yet - wait for approval)
                $transferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, to_user_id, transferred_by, notes, return_expected_date, transfer_status, transfer_type)
                    VALUES (?, ?, ?, ?, ?, ?, 'pending', 'user_to_user')";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([
                    $fileId, $fromUserId, $toUserId, $transferredBy, 
                    $notes, $expectedReturnDate
                ]);
                
                $transferId = $conn->lastInsertId();
                
                // Debug logging
                error_log("transfer_physical_copy: Created transfer ID=$transferId, file_id=$fileId, from_user=$fromUserId, to_user=$toUserId, status=pending");
                
                // Verify the transfer was created correctly
                $verifyQuery = "SELECT id, transfer_status, to_user_id FROM car_file_transfers WHERE id = ?";
                $verifyStmt = $conn->prepare($verifyQuery);
                $verifyStmt->execute([$transferId]);
                $verifyResult = $verifyStmt->fetch(PDO::FETCH_ASSOC);
                error_log("transfer_physical_copy: Verified transfer: " . json_encode($verifyResult));
                
                $conn->commit();
                echo json_encode([
                    'success' => true, 
                    'message' => 'Transfer request sent. Waiting for recipient approval.',
                    'transfer_id' => $transferId,
                    'debug' => $verifyResult
                ]);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'approve_transfer':
            // Approve a pending transfer
            if (!isset($postData['transfer_id']) || !isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'transfer_id and user_id are required']);
                exit;
            }
            
            $transferId = intval($postData['transfer_id']);
            $userId = intval($postData['user_id']);
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Get the pending transfer
                $transferQuery = "SELECT cft.*, cf.file_name 
                    FROM car_file_transfers cft
                    INNER JOIN car_files cf ON cft.car_file_id = cf.id
                    WHERE cft.id = ? AND cft.transfer_status = 'pending'";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([$transferId]);
                $transfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$transfer) {
                    throw new Exception('Pending transfer not found');
                }
                
                // Verify the user is the recipient
                if ($transfer['to_user_id'] != $userId) {
                    throw new Exception('You are not the recipient of this transfer');
                }
                
                // Update transfer status to approved
                $updateTransferQuery = "UPDATE car_file_transfers 
                    SET transfer_status = 'approved', 
                        notes = CONCAT(COALESCE(notes, ''), IF(notes IS NULL OR notes = '', '', ' | '), ?)
                    WHERE id = ?";
                $stmt = $conn->prepare($updateTransferQuery);
                $stmt->execute([$notes ?: 'Transfer approved', $transferId]);
                
                // Now update the physical tracking
                // Handle both available and checked_out files
                $trackingQuery = "SELECT id, status, current_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status IN ('available', 'checked_out')";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$transfer['car_file_id']]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if ($tracking) {
                    // Update tracking: set status to checked_out and assign to recipient
                    $updateTrackingQuery = "UPDATE car_file_physical_tracking 
                        SET previous_holder_id = ?, 
                            current_holder_id = ?, 
                            status = 'checked_out',
                            checked_out_at = NOW(),
                            transferred_by = ?, 
                            transferred_at = NOW(),
                            expected_return_date = ?, 
                            transfer_notes = ?
                        WHERE id = ?";
                    $stmt = $conn->prepare($updateTrackingQuery);
                    $stmt->execute([
                        $tracking['current_holder_id'], // previous_holder (could be NULL if was available)
                        $transfer['to_user_id'], 
                        $transfer['transferred_by'],
                        $transfer['return_expected_date'],
                        $transfer['notes'],
                        $tracking['id']
                    ]);
                } else {
                    // Create new tracking record if it doesn't exist (shouldn't happen, but safety check)
                    $insertTrackingQuery = "INSERT INTO car_file_physical_tracking 
                        (car_file_id, previous_holder_id, current_holder_id, status, checked_out_at, 
                         transferred_by, transferred_at, expected_return_date, transfer_notes, checkout_type)
                        VALUES (?, ?, ?, 'checked_out', NOW(), ?, NOW(), ?, ?, 'user')";
                    $stmt = $conn->prepare($insertTrackingQuery);
                    $stmt->execute([
                        $transfer['car_file_id'],
                        $transfer['from_user_id'],
                        $transfer['to_user_id'],
                        $transfer['transferred_by'],
                        $transfer['return_expected_date'],
                        $transfer['notes']
                    ]);
                }
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'Transfer approved successfully']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'reject_transfer':
            // Reject a pending transfer
            if (!isset($postData['transfer_id']) || !isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'transfer_id and user_id are required']);
                exit;
            }
            
            $transferId = intval($postData['transfer_id']);
            $userId = intval($postData['user_id']);
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Get the pending transfer
                $transferQuery = "SELECT * FROM car_file_transfers 
                    WHERE id = ? AND transfer_status = 'pending'";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([$transferId]);
                $transfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$transfer) {
                    throw new Exception('Pending transfer not found');
                }
                
                // Verify the user is the recipient
                if ($transfer['to_user_id'] != $userId) {
                    throw new Exception('You are not the recipient of this transfer');
                }
                
                // Update transfer status to rejected
                $updateTransferQuery = "UPDATE car_file_transfers 
                    SET transfer_status = 'rejected', 
                        notes = CONCAT(COALESCE(notes, ''), IF(notes IS NULL OR notes = '', '', ' | '), ?)
                    WHERE id = ?";
                $stmt = $conn->prepare($updateTransferQuery);
                $stmt->execute([$notes ?: 'Transfer rejected', $transferId]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'Transfer rejected']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'get_pending_transfers':
            // Get pending transfers for the current user
            if (!isset($postData['user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'user_id is required']);
                exit;
            }
            
            $userId = intval($postData['user_id']);
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check if transfer_status column exists
                $hasTransferStatus = false;
                try {
                    $checkColumnQuery = "SELECT COUNT(*) as cnt FROM INFORMATION_SCHEMA.COLUMNS 
                        WHERE TABLE_SCHEMA = DATABASE() 
                        AND TABLE_NAME = 'car_file_transfers' 
                        AND COLUMN_NAME = 'transfer_status'";
                    $stmt = $conn->prepare($checkColumnQuery);
                    $stmt->execute();
                    $result = $stmt->fetch(PDO::FETCH_ASSOC);
                    $hasTransferStatus = ($result && $result['cnt'] > 0);
                } catch (Exception $e) {
                    error_log("Error checking transfer_status column: " . $e->getMessage());
                }
                
                // Build query - if transfer_status doesn't exist, get all transfers to this user
                if ($hasTransferStatus) {
                    $query = "
                        SELECT 
                            cft.*,
                            cf.file_name,
                            cf.car_id,
                            cfc.category_name,
                            u_from.username as from_username,
                            u_to.username as to_username,
                            u_transfer.username as transferred_by_username,
                            cs.vin
                        FROM car_file_transfers cft
                        INNER JOIN car_files cf ON cft.car_file_id = cf.id
                        INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
                        LEFT JOIN cars_stock cs ON cf.car_id = cs.id
                        LEFT JOIN users u_from ON cft.from_user_id = u_from.id
                        LEFT JOIN users u_to ON cft.to_user_id = u_to.id
                        LEFT JOIN users u_transfer ON cft.transferred_by = u_transfer.id
                        WHERE cft.to_user_id = ? AND cft.transfer_status = 'pending'
                        ORDER BY cft.transferred_at DESC
                    ";
                } else {
                    // Fallback: if column doesn't exist, return empty array and log warning
                    error_log("WARNING: transfer_status column does not exist. Please run migration 004_pending_transfers.sql");
                    $transfers = [];
                    echo json_encode([
                        'success' => true, 
                        'data' => $transfers,
                        'warning' => 'transfer_status column not found. Please run migration 004_pending_transfers.sql'
                    ]);
                    exit;
                }
                
                $stmt = $conn->prepare($query);
                $stmt->execute([$userId]);
                $transfers = $stmt->fetchAll(PDO::FETCH_ASSOC);
                
                // Debug logging
                error_log("get_pending_transfers: User ID = $userId, Found " . count($transfers) . " transfers");
                
                // Also check all pending transfers regardless of user (for debugging)
                $debugQuery = "SELECT id, car_file_id, from_user_id, to_user_id, transfer_status, transferred_at 
                    FROM car_file_transfers 
                    WHERE transfer_status = 'pending' 
                    ORDER BY transferred_at DESC LIMIT 10";
                $debugStmt = $conn->prepare($debugQuery);
                $debugStmt->execute();
                $allPending = $debugStmt->fetchAll(PDO::FETCH_ASSOC);
                error_log("get_pending_transfers: All pending transfers in DB: " . json_encode($allPending));
                
                if (count($transfers) > 0) {
                    error_log("First transfer for user $userId: " . json_encode($transfers[0]));
                } else {
                    error_log("get_pending_transfers: No transfers found for user $userId. All pending transfers: " . count($allPending));
                }
                
                echo json_encode([
                    'success' => true, 
                    'data' => $transfers,
                    'debug' => [
                        'user_id' => $userId,
                        'found_count' => count($transfers),
                        'all_pending_count' => count($allPending),
                        'all_pending' => $allPending
                    ]
                ]);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'get_file_transfer_history':
            // Get transfer history for a file
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
                
                // Check permission: can_upload_car_files or admin
                if (!$isAdmin && !hasPermission($conn, $userId, 'can_upload_car_files')) {
                    http_response_code(403);
                    echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to view car file history']);
                    exit;
                }
                
                $query = "SELECT 
                    cft.*,
                    u_from.username as from_username,
                    u_to.username as to_username,
                    u_transfer.username as transferred_by_username,
                    cca_to.name as to_agent_name,
                    cca_from.name as from_agent_name,
                    cft.to_client_name,
                    cft.from_client_name,
                    cft.returned_at,
                    cft.return_notes
                    FROM car_file_transfers cft
                    LEFT JOIN users u_from ON cft.from_user_id = u_from.id
                    LEFT JOIN users u_to ON cft.to_user_id = u_to.id
                    LEFT JOIN users u_transfer ON cft.transferred_by = u_transfer.id
                    LEFT JOIN custom_clearance_agents cca_to ON cft.to_agent_id = cca_to.id
                    LEFT JOIN custom_clearance_agents cca_from ON cft.from_agent_id = cca_from.id
                    WHERE cft.car_file_id = ? AND cft.transfer_status = 'approved'
                    ORDER BY cft.transferred_at DESC";
                
                $result = executeQuery($query, [$fileId]);
                echo json_encode($result);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
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

        case 'delete_file_category':
            // Delete a file category (admin only)
            if (!isset($postData['category_id']) || !isset($postData['is_admin']) || !$postData['is_admin']) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Admin access required']);
                exit;
            }
            
            $categoryId = intval($postData['category_id']);
            
            // Check if category has files
            $checkQuery = "SELECT COUNT(*) as file_count FROM car_files WHERE category_id = ?";
            $checkResult = executeQuery($checkQuery, [$categoryId]);
            
            if ($checkResult['success'] && isset($checkResult['data'][0]['file_count']) && $checkResult['data'][0]['file_count'] > 0) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Cannot delete category: It has associated files. Please delete or reassign files first.']);
                exit;
            }
            
            $query = "DELETE FROM car_file_categories WHERE id = ?";
            $result = executeQuery($query, [$categoryId]);
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

        case 'save_contract_terms':
            // Save contract terms to public/contract_terms.json (admin only)
            if (!isset($postData['is_admin']) || !$postData['is_admin']) {
                http_response_code(403);
                echo json_encode(['success' => false, 'error' => 'Admin access required']);
                exit;
            }

            if (!isset($postData['content'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Content is required']);
                exit;
            }

            try {
                $jsonContent = $postData['content'];
                
                // Validate JSON content
                if (!is_array($jsonContent)) {
                    throw new Exception('Content must be a valid JSON object');
                }
                
                // Ensure required structure
                if (!isset($jsonContent['enabledLanguages']) || !isset($jsonContent['terms'])) {
                    throw new Exception('Content must include enabledLanguages and terms');
                }
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Construct the file path to public/contract_terms.json
                $filePath = __DIR__ . '/../public/contract_terms.json';
                
                // Verify the file path is within the base directory
                $resolvedFilePath = realpath(dirname($filePath));
                if ($resolvedFilePath === false) {
                    // Directory doesn't exist, create it
                    if (!mkdir(dirname($filePath), 0755, true)) {
                        throw new Exception('Failed to create directory');
                    }
                    $resolvedFilePath = realpath(dirname($filePath));
                    if ($resolvedFilePath === false) {
                        throw new Exception('Failed to resolve directory path');
                    }
                }
                $resolvedFilePath = rtrim($resolvedFilePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFilePath, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                // Encode JSON with pretty printing
                $jsonString = json_encode($jsonContent, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                
                if ($jsonString === false) {
                    throw new Exception('Failed to encode JSON: ' . json_last_error_msg());
                }
                
                // Write file
                if (file_put_contents($filePath, $jsonString) === false) {
                    throw new Exception('Failed to write file');
                }
                
                echo json_encode([
                    'success' => true,
                    'message' => 'Contract terms saved successfully'
                ]);
            } catch (Exception $e) {
                http_response_code(500);
                echo json_encode(['success' => false, 'error' => 'Error saving file: ' . $e->getMessage()]);
            }
            exit;
    
        case 'check_api_files_exist':
            // Check if files exist in the api folder
            if (!isset($postData['file_names']) || !is_array($postData['file_names'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_names array is required']);
                exit;
            }

            try {
                $fileNames = $postData['file_names'];
                $apiDir = __DIR__; // Current directory is the api folder
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Verify api directory is within base path
                $resolvedApiDir = realpath($apiDir);
                if ($resolvedApiDir === false) {
                    throw new Exception('Invalid api directory');
                }
                $resolvedApiDir = rtrim($resolvedApiDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedApiDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                $results = [];
                foreach ($fileNames as $fileName) {
                    // Sanitize filename
                    $sanitizedFileName = basename($fileName);
                    $sanitizedFileName = str_replace('..', '', $sanitizedFileName);
                    
                    // Construct file path
                    $filePath = $apiDir . '/' . $sanitizedFileName;
                    
                    // Verify file path is within api directory
                    $resolvedFilePath = realpath($filePath);
                    $exists = false;
                    if ($resolvedFilePath !== false) {
                        $resolvedFileDir = dirname($resolvedFilePath);
                        $resolvedFileDir = rtrim($resolvedFileDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                        if (strpos($resolvedFileDir, $resolvedApiDir) === 0) {
                            $exists = file_exists($filePath);
                        }
                    }
                    
                    $results[$fileName] = $exists;
                }
                
                echo json_encode([
                    'success' => true,
                    'data' => $results
                ]);
            } catch (Exception $e) {
                http_response_code(500);
                echo json_encode(['success' => false, 'error' => 'Error checking files: ' . $e->getMessage()]);
            }
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

// Check if query is updating payment_confirmed column - requires permission check
$queryUpper = strtoupper(trim($query));
$isPaymentConfirmedUpdate = (strpos($queryUpper, 'UPDATE') === 0 && (strpos($queryUpper, 'PAYMENT_CONFIRMED') !== false || strpos($queryUpper, '`payment_confirmed`') !== false));

if ($isPaymentConfirmedUpdate) {
    // Require user_id and is_admin for payment_confirmed updates
    if (!isset($postData['user_id']) || !isset($postData['is_admin'])) {
        http_response_code(400);
        echo json_encode(['success' => false, 'error' => 'User authentication required for payment confirmation updates']);
        exit;
    }
    
    $userId = intval($postData['user_id']);
    $isAdmin = isset($postData['is_admin']) ? (bool)$postData['is_admin'] : false;
    
    try {
        $conn = getConnection(getDbConfig());
        if (is_array($conn) && isset($conn['error'])) {
            throw new Exception($conn['error']);
        }
        
        // Check permission: must be admin or have can_confirm_payment permission
        if (!$isAdmin && !hasPermission($conn, $userId, 'can_confirm_payment')) {
            http_response_code(403);
            echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to confirm payments']);
            exit;
        }
        
        // Check if this is updating sell_bill table
        $isSellBillUpdate = (stripos($queryUpper, 'UPDATE') === 0 && stripos($queryUpper, 'SELL_BILL') !== false);
        
        if ($isSellBillUpdate) {
            // Check if this is a simple payment_confirmed-only update (old format)
            // Or a full update that includes payment_confirmed (new format)
            $isSimplePaymentUpdate = (stripos($queryUpper, 'SET') !== false && 
                                     stripos($queryUpper, 'PAYMENT_CONFIRMED') !== false &&
                                     substr_count($queryUpper, '?') <= 3); // Simple: SET payment_confirmed = ? WHERE id = ?
            
            if ($isSimplePaymentUpdate && count($params) >= 2) {
                // Old simple format: UPDATE sell_bill SET payment_confirmed = ? WHERE id = ?
                $newStatus = intval($params[0]);
                $billId = intval($params[1]);
                
                // Execute the sell_bill update using the existing connection
                try {
                    // Update both payment_confirmed and payment_confirmed_by_user_id in a single query
                    // Set payment_confirmed_by_user_id to userId if confirming (1), NULL if unconfirming (0)
                    $confirmedByUserId = $newStatus == 1 ? $userId : null;
                    
                    // Modify the query to also update payment_confirmed_by_user_id
                    $updatedQuery = "UPDATE sell_bill SET payment_confirmed = ?, payment_confirmed_by_user_id = ? WHERE id = ?";
                    $updatedParams = [$newStatus, $confirmedByUserId, $billId];
                    
                    $stmt = $conn->prepare($updatedQuery);
                    $stmt->execute($updatedParams);
                    $affectedRows = $stmt->rowCount();
                    $result = ['success' => true, 'affectedRows' => $affectedRows];
                    
                    // If update was successful, update related cars
                    if ($affectedRows > 0) {
                        
                        // Update all related cars
                        $updateCarsQuery = "UPDATE cars_stock SET payment_confirmed = ? WHERE id_sell = ?";
                        $carsStmt = $conn->prepare($updateCarsQuery);
                        $carsStmt->execute([$newStatus, $billId]);
                        $carsAffected = $carsStmt->rowCount();
                        
                        // Log the update (optional, for debugging)
                        if ($carsAffected > 0) {
                            error_log("Updated payment_confirmed for $carsAffected cars related to sell_bill ID $billId to status $newStatus by user $userId");
                        }
                    }
                    
                    // Return the sell_bill update result
                    http_response_code(200);
                    echo json_encode($result);
                    exit;
                } catch (Exception $e) {
                    http_response_code(500);
                    echo json_encode(['success' => false, 'error' => $e->getMessage()]);
                    exit;
                }
            } else {
                // Full update query that includes payment_confirmed along with other fields
                // We need to inject payment_confirmed_by_user_id into the query
                try {
                    // Find the payment_confirmed parameter index by counting placeholders before it
                    $paymentConfirmedIndex = -1;
                    
                    // Extract the SET clause
                    if (preg_match('/SET\s+(.+?)\s+WHERE/i', $query, $matches)) {
                        $setClause = $matches[1];
                        
                        // Count question marks before payment_confirmed
                        $beforePaymentConfirmed = substr($setClause, 0, stripos($setClause, 'payment_confirmed'));
                        $paymentConfirmedIndex = substr_count($beforePaymentConfirmed, '?');
                    }
                    
                    // Find bill ID (last parameter)
                    $billIdIndex = count($params) - 1;
                    $billId = intval($params[$billIdIndex]);
                    
                    // Get payment_confirmed value
                    if ($paymentConfirmedIndex >= 0 && isset($params[$paymentConfirmedIndex])) {
                        $newStatus = intval($params[$paymentConfirmedIndex]);
                        $confirmedByUserId = $newStatus == 1 ? $userId : null;
                        
                        // Modify query to include payment_confirmed_by_user_id
                        // Insert it right after payment_confirmed
                        $modifiedQuery = preg_replace(
                            '/(payment_confirmed\s*=\s*\?)/i',
                            '$1, payment_confirmed_by_user_id = ?',
                            $query
                        );
                        
                        // Insert the confirmed_by_user_id parameter right after payment_confirmed
                        $modifiedParams = $params;
                        array_splice($modifiedParams, $paymentConfirmedIndex + 1, 0, $confirmedByUserId);
                        
                        // Execute the modified query
                        $stmt = $conn->prepare($modifiedQuery);
                        $stmt->execute($modifiedParams);
                        $affectedRows = $stmt->rowCount();
                        
                        // If update was successful, update related cars
                        if ($affectedRows > 0 && $newStatus !== null) {
                            $updateCarsQuery = "UPDATE cars_stock SET payment_confirmed = ? WHERE id_sell = ?";
                            $carsStmt = $conn->prepare($updateCarsQuery);
                            $carsStmt->execute([$newStatus, $billId]);
                        }
                        
                        $result = ['success' => true, 'affectedRows' => $affectedRows];
                        http_response_code(200);
                        echo json_encode($result);
                        exit;
                    }
                } catch (Exception $e) {
                    // If modification fails, fall through to normal execution
                    error_log("Error modifying payment_confirmed query: " . $e->getMessage());
                }
            }
        }
    } catch (Exception $e) {
        http_response_code(500);
        echo json_encode(['success' => false, 'error' => 'Error checking permissions: ' . $e->getMessage()]);
        exit;
    }
}

// Execute query and return result
http_response_code(200);
echo json_encode(executeQuery($query, $params));
exit;
?>