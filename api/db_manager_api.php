<?php
header('Access-Control-Allow-Origin: *');
header('Content-Type: application/json');
header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
header('Access-Control-Allow-Headers: Access-Control-Allow-Headers,Content-Type,Access-Control-Allow-Methods,Authorization,X-Requested-With');

// Handle preflight OPTIONS request
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit;
}

// Database configuration for merhab_databases
$db_host = 'localhost';
$db_user = 'root';
$db_pass = 'nooo';
$db_name = 'merhab_databases';

// Response array
$response = [
    'success' => false,
    'message' => '',
    'data' => null
];

try {
    // Get request method
    $method = $_SERVER['REQUEST_METHOD'];
    
    // Get POST/GET data
    $inputData = [];
    if ($method === 'POST') {
        $rawInput = file_get_contents('php://input');
        $inputData = json_decode($rawInput, true) ?? $_POST;
    } elseif ($method === 'GET') {
        $inputData = $_GET;
    }
    
    // Establish database connection to merhab_databases
    $conn = new PDO(
        "mysql:host={$db_host};dbname={$db_name}", 
        $db_user, 
        $db_pass
    );
    $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    // Helper function to format Unix directory path
    function formatUnixPath($path) {
        if (empty($path)) {
            return null;
        }
        // Remove any backslashes and replace with forward slashes
        $path = str_replace('\\', '/', $path);
        // Remove any double slashes
        $path = preg_replace('#/+#', '/', $path);
        // Ensure it starts with /
        if (substr($path, 0, 1) !== '/') {
            $path = '/' . $path;
        }
        // Remove trailing slash (unless it's root)
        if (strlen($path) > 1 && substr($path, -1) === '/') {
            $path = rtrim($path, '/');
        }
        return $path;
    }
    
    // Helper function to validate Unix directory path
    function isValidUnixPath($path) {
        if (empty($path)) {
            return false;
        }
        // Must start with /
        if (substr($path, 0, 1) !== '/') {
            return false;
        }
        // Must not contain invalid characters (Windows drive letters, etc.)
        if (preg_match('/^[a-zA-Z]:/', $path)) {
            return false;
        }
        // Must not contain null bytes or other control characters
        if (strpos($path, "\0") !== false) {
            return false;
        }
        // Should only contain valid Unix path characters (letters, numbers, dots, underscores, hyphens, forward slashes, and spaces)
        // Note: spaces are technically valid but not recommended
        if (!preg_match('/^[\/a-zA-Z0-9._\-\s]+$/', $path)) {
            return false;
        }
        return true;
    }
    
    // Handle different actions
    $action = $inputData['action'] ?? '';
    
    switch ($action) {
        case 'test':
            // Test connection
            $response['success'] = true;
            $response['message'] = 'Database connection successful';
            $response['data'] = [
                'host' => $db_host,
                'database' => $db_name
            ];
            break;
            
        case 'signup':
            // Sign up new user
            $user = trim($inputData['user'] ?? '');
            $pass = $inputData['pass'] ?? '';
            
            if (empty($user) || empty($pass)) {
                $response['message'] = 'Username and password are required';
                break;
            }
            
            if (strlen($pass) < 6) {
                $response['message'] = 'Password must be at least 6 characters long';
                break;
            }
            
            // Check if user already exists
            $checkStmt = $conn->prepare("SELECT id FROM login WHERE user = ?");
            $checkStmt->execute([$user]);
            if ($checkStmt->fetch()) {
                $response['message'] = 'Username already exists';
                break;
            }
            
            // Hash password
            $hashedPassword = password_hash($pass, PASSWORD_DEFAULT);
            
            // Insert new user (active = 0, will be activated manually later)
            $insertStmt = $conn->prepare("INSERT INTO login (user, pass, active) VALUES (?, ?, 0)");
            if ($insertStmt->execute([$user, $hashedPassword])) {
                $response['success'] = true;
                $response['message'] = 'User created successfully';
                $response['data'] = [
                    'id' => $conn->lastInsertId(),
                    'user' => $user
                ];
            } else {
                $response['message'] = 'Failed to create user';
            }
            break;
            
        case 'login':
            // Login user
            $user = trim($inputData['user'] ?? '');
            $pass = $inputData['pass'] ?? '';
            
            if (empty($user) || empty($pass)) {
                $response['message'] = 'Username and password are required';
                break;
            }
            
            // Get user from database
            $stmt = $conn->prepare("SELECT id, user, pass, active FROM login WHERE user = ?");
            $stmt->execute([$user]);
            $userData = $stmt->fetch(PDO::FETCH_ASSOC);
            
            if (!$userData) {
                $response['message'] = 'Invalid username or password';
                break;
            }
            
            // Check if user is active
            if ($userData['active'] != 1) {
                $response['message'] = 'Account is not active';
                break;
            }
            
            // Verify password
            if (password_verify($pass, $userData['pass'])) {
                $response['success'] = true;
                $response['message'] = 'Login successful';
                $response['data'] = [
                    'id' => $userData['id'],
                    'user' => $userData['user']
                ];
            } else {
                $response['message'] = 'Invalid username or password';
            }
            break;
            
        case 'get_databases':
            // Get all databases
            $stmt = $conn->prepare("SELECT * FROM dbs ORDER BY id DESC");
            $stmt->execute();
            $databases = $stmt->fetchAll(PDO::FETCH_ASSOC);
            $response['success'] = true;
            $response['data'] = $databases;
            break;
            
        case 'create_database':
            // Create new database
            $db_name = trim($inputData['db_name'] ?? '');
            
            if (empty($db_name)) {
                $response['message'] = 'DB Name is required';
                break;
            }
            
            // Auto-generate db_code from db_name + server timestamp
            $timestamp = time();
            $hashInput = $db_name . '_' . $timestamp;
            // Use first 20 characters of MD5 hash for better uniqueness
            $db_code = 'db_' . substr(md5($hashInput), 0, 20);
            
            // Ensure uniqueness - if code exists, regenerate with microtime
            $checkStmt = $conn->prepare("SELECT id FROM dbs WHERE db_code = ?");
            $checkStmt->execute([$db_code]);
            $attempts = 0;
            while ($checkStmt->fetch() && $attempts < 10) {
                // Use microtime for better uniqueness
                $microtime = microtime(true);
                $hashInput = $db_name . '_' . $microtime;
                $db_code = 'db_' . substr(md5($hashInput), 0, 20);
                $checkStmt->execute([$db_code]);
                $attempts++;
            }
            
            // Prepare data
            $db_host_start = !empty($inputData['db_host_start']) ? $inputData['db_host_start'] : null;
            $db_host_end = !empty($inputData['db_host_end']) ? $inputData['db_host_end'] : null;
            $db_host_cost = !empty($inputData['db_host_cost_per_month']) ? floatval($inputData['db_host_cost_per_month']) : null;
            $serv_host_start = !empty($inputData['serv_host_start']) ? $inputData['serv_host_start'] : null;
            $serv_host_end = !empty($inputData['serv_host_end']) ? $inputData['serv_host_end'] : null;
            $serv_host_cost = !empty($inputData['serv_host_cost_per_month']) ? floatval($inputData['serv_host_cost_per_month']) : null;
            
            // Format and validate Unix directory paths
            $files_dir = !empty($inputData['files_dir']) ? formatUnixPath(trim($inputData['files_dir'])) : null;
            $js_dir = !empty($inputData['js_dir']) ? formatUnixPath(trim($inputData['js_dir'])) : null;
            
            // Validate paths if provided
            if ($files_dir !== null && !isValidUnixPath($files_dir)) {
                $response['message'] = 'Invalid Files Directory path format. Must be a valid Unix path (e.g., /path/to/directory)';
                break;
            }
            if ($js_dir !== null && !isValidUnixPath($js_dir)) {
                $response['message'] = 'Invalid JS Directory path format. Must be a valid Unix path (e.g., /path/to/directory)';
                break;
            }
            
            $insertStmt = $conn->prepare("INSERT INTO dbs (db_code, db_name, db_host_start, db_host_end, db_host_cost_per_month, serv_host_start, serv_host_end, serv_host_cost_per_month, files_dir, js_dir) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)");
            
            if ($insertStmt->execute([$db_code, $db_name, $db_host_start, $db_host_end, $db_host_cost, $serv_host_start, $serv_host_end, $serv_host_cost, $files_dir, $js_dir])) {
                $response['success'] = true;
                $response['message'] = 'Database created successfully';
                $response['data'] = [
                    'id' => $conn->lastInsertId(),
                    'db_code' => $db_code,
                    'db_name' => $db_name
                ];
            } else {
                $response['message'] = 'Failed to create database';
            }
            break;
            
        case 'update_database':
            // Update database
            $id = intval($inputData['id'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            // Check if database exists
            $checkStmt = $conn->prepare("SELECT id FROM dbs WHERE id = ?");
            $checkStmt->execute([$id]);
            if (!$checkStmt->fetch()) {
                $response['message'] = 'Database not found';
                break;
            }
            
            $db_name = trim($inputData['db_name'] ?? '');
            
            if (empty($db_name)) {
                $response['message'] = 'DB Name is required';
                break;
            }
            
            // Get existing db_code (it cannot be changed)
            $getStmt = $conn->prepare("SELECT db_code FROM dbs WHERE id = ?");
            $getStmt->execute([$id]);
            $existing = $getStmt->fetch(PDO::FETCH_ASSOC);
            $db_code = $existing['db_code'];
            
            // Prepare data
            $db_host_start = !empty($inputData['db_host_start']) ? $inputData['db_host_start'] : null;
            $db_host_end = !empty($inputData['db_host_end']) ? $inputData['db_host_end'] : null;
            $db_host_cost = !empty($inputData['db_host_cost_per_month']) ? floatval($inputData['db_host_cost_per_month']) : null;
            $serv_host_start = !empty($inputData['serv_host_start']) ? $inputData['serv_host_start'] : null;
            $serv_host_end = !empty($inputData['serv_host_end']) ? $inputData['serv_host_end'] : null;
            $serv_host_cost = !empty($inputData['serv_host_cost_per_month']) ? floatval($inputData['serv_host_cost_per_month']) : null;
            
            // Format and validate Unix directory paths
            $files_dir = !empty($inputData['files_dir']) ? formatUnixPath(trim($inputData['files_dir'])) : null;
            $js_dir = !empty($inputData['js_dir']) ? formatUnixPath(trim($inputData['js_dir'])) : null;
            
            // Validate paths if provided
            if ($files_dir !== null && !isValidUnixPath($files_dir)) {
                $response['message'] = 'Invalid Files Directory path format. Must be a valid Unix path (e.g., /path/to/directory)';
                break;
            }
            if ($js_dir !== null && !isValidUnixPath($js_dir)) {
                $response['message'] = 'Invalid JS Directory path format. Must be a valid Unix path (e.g., /path/to/directory)';
                break;
            }
            
            $updateStmt = $conn->prepare("UPDATE dbs SET db_code = ?, db_name = ?, db_host_start = ?, db_host_end = ?, db_host_cost_per_month = ?, serv_host_start = ?, serv_host_end = ?, serv_host_cost_per_month = ?, files_dir = ?, js_dir = ? WHERE id = ?");
            
            if ($updateStmt->execute([$db_code, $db_name, $db_host_start, $db_host_end, $db_host_cost, $serv_host_start, $serv_host_end, $serv_host_cost, $files_dir, $js_dir, $id])) {
                $response['success'] = true;
                $response['message'] = 'Database updated successfully';
            } else {
                $response['message'] = 'Failed to update database';
            }
            break;
            
        case 'delete_database':
            // Delete database
            $id = intval($inputData['id'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            // Check if database exists
            $checkStmt = $conn->prepare("SELECT id FROM dbs WHERE id = ?");
            $checkStmt->execute([$id]);
            if (!$checkStmt->fetch()) {
                $response['message'] = 'Database not found';
                break;
            }
            
            $deleteStmt = $conn->prepare("DELETE FROM dbs WHERE id = ?");
            
            if ($deleteStmt->execute([$id])) {
                $response['success'] = true;
                $response['message'] = 'Database deleted successfully';
            } else {
                $response['message'] = 'Failed to delete database';
            }
            break;
            
        default:
            $response['message'] = 'No action specified';
            break;
    }
    
} catch (Exception $e) {
    error_log('DB Manager API Error: ' . $e->getMessage());
    $response['message'] = $e->getMessage();
    http_response_code(500);
}

// Ensure no output before JSON
if (ob_get_length()) ob_clean();

// Send JSON response
echo json_encode($response);
exit;

