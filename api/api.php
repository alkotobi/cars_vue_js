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

// Include database configuration
require_once 'config.php';

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
    global $db_config;
    
    try {
        $conn = getConnection($db_config);
        
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
    global $db_config;
    
    try {
        $conn = getConnection($db_config);
        
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

// Only accept POST requests
if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    echo json_encode(['success' => false, 'error' => 'Only POST method is allowed']);
    exit;
}

// Get POST data
$postData = json_decode(file_get_contents('php://input'), true);

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
                $conn = getConnection($db_config);
                
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