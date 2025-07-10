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

        foreach ($statements as $statement) {
            if (empty($statement)) continue;
            
            $stmt = $conn->prepare($statement);
            $stmt->execute($params);
            
            // Determine query type
            $queryType = strtoupper(substr(trim($statement), 0, 6));
            
            switch($queryType) {
                case 'SELECT':
                    $lastResult = ['type' => 'SELECT', 'data' => $stmt->fetchAll(PDO::FETCH_ASSOC)];
                    break;
                case 'INSERT':
                    $lastResult = ['type' => 'INSERT', 'lastInsertId' => $conn->lastInsertId()];
                    break;
                case 'UPDATE':
                case 'DELETE':
                    $lastResult = ['type' => $queryType, 'affectedRows' => $stmt->rowCount()];
                    break;
                case 'SET':
                case 'START':
                case 'COMMIT':
                    $lastResult = ['type' => $queryType, 'success' => true];
                    break;
                default:
                    $lastResult = ['type' => 'OTHER', 'success' => true];
            }
            
            $results[] = $lastResult;
        }

        // Return the last SELECT result as the main data, or success status
        if ($lastResult && $lastResult['type'] === 'SELECT') {
            return ['success' => true, 'data' => $lastResult['data'], 'allResults' => $results];
        } else {
            return ['success' => true, 'message' => 'Multi-statement executed successfully', 'allResults' => $results];
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
            
            // Execute multi-statement query
            $result = executeMultiQuery($query, $params);
            echo json_encode($result);
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
                if (!isset($postData['query']) ||!isset($postData['params']) || count($postData['params'])!== 4) {
                    http_response_code(400);
                    echo json_encode(['success' => false, 'error' => 'Invalid parameters for password update']);
                    exit;
                }

                // Hash the password (first parameter)
                $hashedPassword = password_hash($postData['params'][2], PASSWORD_DEFAULT);
                
                // Replace the plain password with hashed password in params
                $postData['params'][2] = $hashedPassword;
                
                // Execute the update query with hashed password
                $result = executeQuery($postData['query'], $postData['params']);
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