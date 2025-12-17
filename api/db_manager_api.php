<?php
// Start output buffering to prevent any output before headers
ob_start();

// Set CORS headers first, before any output
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Accept, Authorization, X-Requested-With');
header('Access-Control-Max-Age: 86400'); // Cache preflight for 24 hours
header('Content-Type: application/json');

// Handle preflight OPTIONS request
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    ob_end_clean();
    exit;
}

// Response array (initialize early)
$response = [
    'success' => false,
    'message' => '',
    'data' => null
];

// Wrap everything in try-catch to ensure headers are always sent
try {
    // Include database configuration
    if (!file_exists(__DIR__ . '/db_manager_config.php')) {
        throw new Exception('db_manager_config.php file not found');
    }
    require_once __DIR__ . '/db_manager_config.php';
    
    // Check if config is loaded
    if (!isset($db_manager_config)) {
        throw new Exception('Database configuration not loaded');
    }
    
    // Use config values
    $db_host = $db_manager_config['host'];
    $db_user = $db_manager_config['user'];
    $db_pass = $db_manager_config['pass'];
    $db_name = $db_manager_config['dbname'];
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
        // Using explicit space character instead of \s to prevent control characters (newlines, tabs, etc.)
        if (!preg_match('/^[\/a-zA-Z0-9._\- ]+$/', $path)) {
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
            
            // Fetch version for each created database
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            foreach ($databases as &$db) {
                $db['version'] = null;
                // Only fetch version if database is created
                if ($db['is_created'] == 1 && !empty($db['db_name'])) {
                    try {
                        $targetConn = new PDO(
                            "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                            $targetDbUser,
                            $targetDbPass
                        );
                        $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                        
                        $versionStmt = $targetConn->prepare("SELECT version FROM versions WHERE id = 1");
                        $versionStmt->execute();
                        $versionData = $versionStmt->fetch(PDO::FETCH_ASSOC);
                        if ($versionData && isset($versionData['version'])) {
                            $db['version'] = intval($versionData['version']);
                        }
                    } catch (PDOException $e) {
                        // If version fetch fails, leave it as null
                        error_log('Failed to fetch version for database ' . $db['db_name'] . ': ' . $e->getMessage());
                    }
                }
            }
            unset($db); // Break reference
            
            $response['success'] = true;
            $response['data'] = $databases;
            break;
            
        case 'get_database_by_code':
            // Get database by db_code
            $db_code = trim($inputData['db_code'] ?? '');
            
            if (empty($db_code)) {
                $response['message'] = 'DB Code is required';
                break;
            }
            
            $stmt = $conn->prepare("SELECT db_name, files_dir, js_dir FROM dbs WHERE db_code = ?");
            $stmt->execute([$db_code]);
            $database = $stmt->fetch(PDO::FETCH_ASSOC);
            
            if (!$database) {
                $response['message'] = 'Database not found';
                break;
            }
            
            $response['success'] = true;
            $response['data'] = $database;
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
            
            // Check if database still exists (race condition protection)
            if (!$existing || !isset($existing['db_code'])) {
                $response['message'] = 'Database not found or was deleted';
                break;
            }
            
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
            
        case 'create_tables':
            // Create tables from setup.sql
            $id = intval($inputData['id'] ?? 0);
            $version = intval($inputData['version'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            if ($version <= 0) {
                $response['message'] = 'Version is required and must be a positive number';
                break;
            }
            
            // Get database info (including db_code, files_dir, and js_dir for directory and db_code.json creation)
            $getStmt = $conn->prepare("SELECT id, db_name, db_code, files_dir, js_dir, is_created FROM dbs WHERE id = ?");
            $getStmt->execute([$id]);
            $dbInfo = $getStmt->fetch(PDO::FETCH_ASSOC);
            
            if (!$dbInfo) {
                $response['message'] = 'Database not found';
                break;
            }
            
            if ($dbInfo['is_created'] == 1) {
                $response['message'] = 'Tables have already been created for this database';
                break;
            }
            
            if (empty($dbInfo['db_name'])) {
                $response['message'] = 'Database name is not set';
                break;
            }
            
            // Read setup.sql file
            $setupSqlPath = __DIR__ . '/setup.sql';
            if (!file_exists($setupSqlPath)) {
                $response['message'] = 'setup.sql file not found';
                break;
            }
            
            $setupSql = file_get_contents($setupSqlPath);
            if ($setupSql === false) {
                $response['message'] = 'Failed to read setup.sql file';
                break;
            }
            
            // Parse SQL file to extract CREATE TABLE statements
            // Remove comments
            $setupSql = preg_replace('/--.*$/m', '', $setupSql); // Remove single-line comments
            $setupSql = preg_replace('/\/\*.*?\*\//s', '', $setupSql); // Remove multi-line comments
            
            // Normalize whitespace (replace multiple whitespace with single space, but preserve newlines within statements)
            $setupSql = preg_replace('/\s+/', ' ', $setupSql);
            
            // Split by semicolons and filter for CREATE TABLE and INSERT statements
            $statements = array_filter(
                array_map('trim', explode(';', $setupSql)),
                function($stmt) {
                    $stmt = trim($stmt);
                    return !empty($stmt) && (stripos($stmt, 'CREATE TABLE') !== false || stripos($stmt, 'INSERT') !== false);
                }
            );
            
            // Add semicolons back to statements (they were removed by explode)
            $statements = array_map(function($stmt) {
                return trim($stmt) . ';';
            }, $statements);
            
            if (empty($statements)) {
                $response['message'] = 'No CREATE TABLE statements found in setup.sql';
                break;
            }
            
            // Get database credentials from config.php
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            $targetDbName = $dbInfo['db_name'];
            
            // Connect to target database
            try {
                $targetConn = new PDO(
                    "mysql:host={$targetDbHost};dbname={$targetDbName}",
                    $targetDbUser,
                    $targetDbPass
                );
                $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
            } catch (PDOException $e) {
                $response['message'] = 'Failed to connect to target database: ' . $e->getMessage();
                break;
            }
            
            // Execute CREATE TABLE statements
            $executedCount = 0;
            $errors = [];
            foreach ($statements as $statement) {
                try {
                    $targetConn->exec($statement);
                    $executedCount++;
                } catch (PDOException $e) {
                    $errors[] = $e->getMessage();
                }
            }
            
            if (!empty($errors) && $executedCount === 0) {
                $response['message'] = 'Failed to create tables: ' . implode('; ', $errors);
                break;
            }
            
            // Create directories (files_dir and js_dir) if they are configured
            $directoriesCreated = [];
            $directoryErrors = [];
            
            // Helper function to create directory safely
            $createDirectory = function($dirPath, $dirName) use (&$directoriesCreated, &$directoryErrors) {
                if (empty($dirPath)) {
                    return; // Skip if directory path is empty
                }
                
                try {
                    $dir = trim($dirPath);
                    // Remove leading slash if present
                    $dir = ltrim($dir, '/');
                    $dir = rtrim($dir, '/');
                    
                    if (empty($dir)) {
                        return; // Skip if directory is empty after trimming
                    }
                    
                    // Construct the full directory path
                    $fullDirPath = __DIR__ . '/../' . $dir;
                    
                    // Get the real path for security check
                    $realBasePath = realpath(__DIR__ . '/../');
                    if ($realBasePath === false) {
                        throw new Exception('Invalid base directory');
                    }
                    $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                    
                    // Normalize the directory path
                    $normalizedDirPath = str_replace(['\\', '/'], DIRECTORY_SEPARATOR, $fullDirPath);
                    $normalizedDirPath = preg_replace('/' . preg_quote(DIRECTORY_SEPARATOR, '/') . '+/', DIRECTORY_SEPARATOR, $normalizedDirPath);
                    
                    // Create directory and all parent directories if they don't exist
                    if (!is_dir($normalizedDirPath)) {
                        // Create directory recursively (creates parent directories if needed)
                        if (!mkdir($normalizedDirPath, 0755, true)) {
                            throw new Exception('Failed to create directory: ' . $normalizedDirPath);
                        }
                    }
                    
                    // Verify the directory exists and is within the base directory
                    $resolvedDirPath = realpath($normalizedDirPath);
                    if ($resolvedDirPath === false) {
                        throw new Exception('Failed to resolve directory path: ' . $normalizedDirPath);
                    }
                    $resolvedDirPath = rtrim($resolvedDirPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                    
                    // Security check: ensure the resolved path is within the base directory
                    if (strpos($resolvedDirPath, $realBasePath) !== 0) {
                        throw new Exception('Directory traversal attempt detected');
                    }
                    
                    $directoriesCreated[] = $dirName . ' (' . $dir . ')';
                } catch (Exception $e) {
                    $directoryErrors[] = $dirName . ': ' . $e->getMessage();
                    error_log('Failed to create ' . $dirName . ' directory: ' . $e->getMessage());
                }
            };
            
            // Create files_dir if configured
            if (!empty($dbInfo['files_dir'])) {
                $createDirectory($dbInfo['files_dir'], 'files_dir');
            }
            
            // Create js_dir if configured
            if (!empty($dbInfo['js_dir'])) {
                $createDirectory($dbInfo['js_dir'], 'js_dir');
            }
            
            // Update version in versions table
            try {
                $versionStmt = $targetConn->prepare("INSERT INTO versions (id, version) VALUES (1, ?) ON DUPLICATE KEY UPDATE version = ?");
                $versionStmt->execute([$version, $version]);
            } catch (PDOException $e) {
                // If versions table doesn't exist yet or update fails, log but don't fail the whole operation
                error_log('Failed to update version: ' . $e->getMessage());
            }
            
            // Update is_created flag
            $updateStmt = $conn->prepare("UPDATE dbs SET is_created = 1 WHERE id = ?");
            if ($updateStmt->execute([$id])) {
                // Create db_code.json file if js_dir is configured
                $dbCodeJsonCreated = false;
                $dbCodeJsonError = null;
                
                if (!empty($dbInfo['js_dir']) && !empty($dbInfo['db_code'])) {
                    try {
                        $jsDir = trim($dbInfo['js_dir']);
                        // Remove leading slash if present
                        $jsDir = ltrim($jsDir, '/');
                        $jsDir = rtrim($jsDir, '/');
                        
                        // Construct the full directory and file path
                        $dirPath = __DIR__ . '/../' . $jsDir;
                        $filePath = $dirPath . '/db_code.json';
                        
                        // Get the real path for security check
                        $realBasePath = realpath(__DIR__ . '/../');
                        if ($realBasePath === false) {
                            throw new Exception('Invalid base directory');
                        }
                        $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                        
                        // Normalize the directory path
                        $normalizedDirPath = str_replace(['\\', '/'], DIRECTORY_SEPARATOR, $dirPath);
                        $normalizedDirPath = preg_replace('/' . preg_quote(DIRECTORY_SEPARATOR, '/') . '+/', DIRECTORY_SEPARATOR, $normalizedDirPath);
                        
                        // Create directory and all parent directories if they don't exist
                        if (!is_dir($normalizedDirPath)) {
                            // Create directory recursively (creates parent directories if needed)
                            if (!mkdir($normalizedDirPath, 0755, true)) {
                                throw new Exception('Failed to create directory: ' . $normalizedDirPath);
                            }
                        }
                        
                        // Verify the directory exists and is within the base directory
                        $resolvedDirPath = realpath($normalizedDirPath);
                        if ($resolvedDirPath === false) {
                            throw new Exception('Failed to resolve directory path: ' . $normalizedDirPath);
                        }
                        $resolvedDirPath = rtrim($resolvedDirPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                        
                        // Security check: ensure the resolved path is within the base directory
                        if (strpos($resolvedDirPath, $realBasePath) !== 0) {
                            throw new Exception('Directory traversal attempt detected');
                        }
                        
                        // Create JSON content
                        $jsonContent = [
                            'db_code' => $dbInfo['db_code']
                        ];
                        
                        // Encode JSON with pretty printing
                        $jsonString = json_encode($jsonContent, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                        
                        if ($jsonString === false) {
                            throw new Exception('Failed to encode JSON: ' . json_last_error_msg());
                        }
                        
                        // Write file (only if it doesn't exist)
                        if (!file_exists($filePath)) {
                            if (file_put_contents($filePath, $jsonString) === false) {
                                throw new Exception('Failed to write db_code.json file');
                            }
                            $dbCodeJsonCreated = true;
                        }
                    } catch (Exception $e) {
                        $dbCodeJsonError = $e->getMessage();
                        error_log('Failed to create db_code.json: ' . $dbCodeJsonError);
                    }
                }
                
                $response['success'] = true;
                $response['message'] = "Successfully created {$executedCount} table(s)";
                
                // Add directory creation status
                if (!empty($directoriesCreated)) {
                    $response['message'] .= ', created directories: ' . implode(', ', $directoriesCreated);
                }
                if (!empty($directoryErrors)) {
                    $response['message'] .= ' (directory errors: ' . implode('; ', $directoryErrors) . ')';
                }
                
                // Add db_code.json creation status
                if ($dbCodeJsonCreated) {
                    $response['message'] .= ', created db_code.json';
                }
                if ($dbCodeJsonError) {
                    $response['message'] .= ' (note: db_code.json creation failed: ' . $dbCodeJsonError . ')';
                }
                
                // Add table creation errors if any
                if (!empty($errors)) {
                    $response['message'] .= ' (some table errors occurred: ' . implode('; ', $errors) . ')';
                }
            } else {
                $response['message'] = 'Tables created but failed to update is_created flag';
            }
            break;
            
        case 'get_db_updates':
            // Get all db_updates
            $stmt = $conn->prepare("SELECT * FROM db_updates ORDER BY from_version ASC, current_version ASC");
            $stmt->execute();
            $updates = $stmt->fetchAll(PDO::FETCH_ASSOC);
            $response['success'] = true;
            $response['data'] = $updates;
            break;
            
        case 'create_db_update':
            // Create new db_update
            $fromVersion = intval($inputData['from_version'] ?? 0);
            $currentVersion = intval($inputData['current_version'] ?? 0);
            $description = trim($inputData['description'] ?? '');
            $sql = trim($inputData['sql'] ?? '');
            
            if ($fromVersion <= 0 || $currentVersion <= 0) {
                $response['message'] = 'From version and current version must be positive numbers';
                break;
            }
            
            if (empty($sql)) {
                $response['message'] = 'SQL is required';
                break;
            }
            
            if ($currentVersion <= $fromVersion) {
                $response['message'] = 'Current version must be greater than from version';
                break;
            }
            
            $insertStmt = $conn->prepare("INSERT INTO db_updates (from_version, current_version, description, `sql`) VALUES (?, ?, ?, ?)");
            if ($insertStmt->execute([$fromVersion, $currentVersion, $description, $sql])) {
                $response['success'] = true;
                $response['message'] = 'Update created successfully';
                $response['data'] = ['id' => $conn->lastInsertId()];
            } else {
                $response['message'] = 'Failed to create update';
            }
            break;
            
        case 'update_db_update':
            // Update existing db_update
            $id = intval($inputData['id'] ?? 0);
            $fromVersion = intval($inputData['from_version'] ?? 0);
            $currentVersion = intval($inputData['current_version'] ?? 0);
            $description = trim($inputData['description'] ?? '');
            $sql = trim($inputData['sql'] ?? '');
            
            if ($id <= 0) {
                $response['message'] = 'Invalid update ID';
                break;
            }
            
            if ($fromVersion <= 0 || $currentVersion <= 0) {
                $response['message'] = 'From version and current version must be positive numbers';
                break;
            }
            
            if (empty($sql)) {
                $response['message'] = 'SQL is required';
                break;
            }
            
            if ($currentVersion <= $fromVersion) {
                $response['message'] = 'Current version must be greater than from version';
                break;
            }
            
            $updateStmt = $conn->prepare("UPDATE db_updates SET from_version = ?, current_version = ?, description = ?, `sql` = ? WHERE id = ?");
            if ($updateStmt->execute([$fromVersion, $currentVersion, $description, $sql, $id])) {
                $response['success'] = true;
                $response['message'] = 'Update updated successfully';
            } else {
                $response['message'] = 'Failed to update';
            }
            break;
            
        case 'delete_db_update':
            // Delete db_update
            $id = intval($inputData['id'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid update ID';
                break;
            }
            
            $deleteStmt = $conn->prepare("DELETE FROM db_updates WHERE id = ?");
            if ($deleteStmt->execute([$id])) {
                $response['success'] = true;
                $response['message'] = 'Update deleted successfully';
            } else {
                $response['message'] = 'Failed to delete update';
            }
            break;
            
        case 'update_databases_version':
            // Update version for multiple databases
            $databaseIds = $inputData['database_ids'] ?? [];
            $version = intval($inputData['version'] ?? 0);
            
            if (empty($databaseIds) || !is_array($databaseIds)) {
                $response['message'] = 'Database IDs are required';
                break;
            }
            
            if ($version <= 0) {
                $response['message'] = 'Version must be a positive number';
                break;
            }
            
            // Get database credentials from config.php
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            // Get database names for the selected IDs
            $placeholders = implode(',', array_fill(0, count($databaseIds), '?'));
            $getStmt = $conn->prepare("SELECT id, db_name FROM dbs WHERE id IN ($placeholders)");
            $getStmt->execute($databaseIds);
            $selectedDbs = $getStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($selectedDbs)) {
                $response['message'] = 'No databases found for the selected IDs';
                break;
            }
            
            $successCount = 0;
            $errors = [];
            
            foreach ($selectedDbs as $db) {
                if (empty($db['db_name'])) {
                    $errors[] = "Database ID {$db['id']}: Database name is not set";
                    continue;
                }
                
                try {
                    $targetConn = new PDO(
                        "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                        $targetDbUser,
                        $targetDbPass
                    );
                    $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                    
                    // Update version in versions table
                    $versionStmt = $targetConn->prepare("INSERT INTO versions (id, version) VALUES (1, ?) ON DUPLICATE KEY UPDATE version = ?");
                    $versionStmt->execute([$version, $version]);
                    $successCount++;
                } catch (PDOException $e) {
                    $errors[] = "Database {$db['db_name']}: " . $e->getMessage();
                }
            }
            
            if ($successCount > 0) {
                $response['success'] = true;
                $response['message'] = "Version updated successfully for {$successCount} database(s)";
                if (!empty($errors)) {
                    $response['message'] .= ' (some errors: ' . implode('; ', $errors) . ')';
                }
            } else {
                $response['message'] = 'Failed to update version: ' . implode('; ', $errors);
            }
            break;
            
        case 'update_structure':
            // Update structure for selected databases based on db_updates table
            $databaseIds = $inputData['database_ids'] ?? [];
            
            if (empty($databaseIds) || !is_array($databaseIds)) {
                $response['message'] = 'Database IDs are required';
                break;
            }
            
            // Get all db_updates records
            $updatesStmt = $conn->prepare("SELECT * FROM db_updates ORDER BY from_version ASC, current_version ASC");
            $updatesStmt->execute();
            $dbUpdates = $updatesStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($dbUpdates)) {
                $response['message'] = 'No update records found in db_updates table';
                break;
            }
            
            // Get database credentials from config.php
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            // Get selected databases with their versions
            $placeholders = implode(',', array_fill(0, count($databaseIds), '?'));
            $getStmt = $conn->prepare("SELECT id, db_name FROM dbs WHERE id IN ($placeholders)");
            $getStmt->execute($databaseIds);
            $selectedDbs = $getStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($selectedDbs)) {
                $response['message'] = 'No databases found for the selected IDs';
                break;
            }
            
            // Fetch versions for selected databases
            $dbVersions = [];
            foreach ($selectedDbs as $db) {
                if (empty($db['db_name'])) {
                    continue;
                }
                
                try {
                    $targetConn = new PDO(
                        "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                        $targetDbUser,
                        $targetDbPass
                    );
                    $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                    
                    $versionStmt = $targetConn->prepare("SELECT version FROM versions WHERE id = 1");
                    $versionStmt->execute();
                    $versionData = $versionStmt->fetch(PDO::FETCH_ASSOC);
                    $dbVersions[$db['id']] = [
                        'db_name' => $db['db_name'],
                        'version' => $versionData ? intval($versionData['version']) : null,
                        'conn' => $targetConn
                    ];
                } catch (PDOException $e) {
                    $dbVersions[$db['id']] = [
                        'db_name' => $db['db_name'],
                        'version' => null,
                        'conn' => null,
                        'error' => $e->getMessage()
                    ];
                }
            }
            
            $totalUpdates = 0;
            $totalErrors = 0;
            $errors = [];
            // Track which databases have already been marked with connection/version errors
            // to avoid counting the same error multiple times (once per update record)
            $dbErrorsTracked = [];
            
            // Iterate through db_updates records
            foreach ($dbUpdates as $update) {
                $fromVersion = intval($update['from_version']);
                $currentVersion = intval($update['current_version']);
                $sql = $update['sql'];
                
                // For each selected database, check if version >= from_version
                foreach ($dbVersions as $dbId => $dbInfo) {
                    // Check for connection error (only count once per database)
                    if ($dbInfo['conn'] === null) {
                        if (!isset($dbErrorsTracked[$dbId]['conn'])) {
                            $errors[] = "Database {$dbInfo['db_name']}: Cannot connect - " . ($dbInfo['error'] ?? 'Unknown error');
                            $totalErrors++;
                            $dbErrorsTracked[$dbId]['conn'] = true;
                        }
                        continue;
                    }
                    
                    // Check for version not found (only count once per database)
                    if ($dbInfo['version'] === null) {
                        if (!isset($dbErrorsTracked[$dbId]['version'])) {
                            $errors[] = "Database {$dbInfo['db_name']}: Version not found";
                            $totalErrors++;
                            $dbErrorsTracked[$dbId]['version'] = true;
                        }
                        continue;
                    }
                    
                    // Check if database version >= from_version
                    if ($dbInfo['version'] >= $fromVersion) {
                        try {
                            // Execute the SQL statement
                            $dbInfo['conn']->exec($sql);
                            
                            // Update version to current_version after successful SQL execution
                            $versionUpdateStmt = $dbInfo['conn']->prepare("INSERT INTO versions (id, version) VALUES (1, ?) ON DUPLICATE KEY UPDATE version = ?");
                            $versionUpdateStmt->execute([$currentVersion, $currentVersion]);
                            
                            // Update the version in our tracking array for subsequent checks
                            $dbVersions[$dbId]['version'] = $currentVersion;
                            
                            $totalUpdates++;
                        } catch (PDOException $e) {
                            $errors[] = "Database {$dbInfo['db_name']} (update from v{$fromVersion} to v{$currentVersion}): " . $e->getMessage();
                            $totalErrors++;
                        }
                    }
                }
            }
            
            // Close all connections
            foreach ($dbVersions as $dbInfo) {
                if ($dbInfo['conn'] !== null) {
                    $dbInfo['conn'] = null;
                }
            }
            
            if ($totalUpdates > 0 || $totalErrors === 0) {
                $response['success'] = true;
                $response['message'] = "Structure update completed: {$totalUpdates} SQL statement(s) executed";
                if (!empty($errors)) {
                    $response['message'] .= ' (some errors: ' . implode('; ', array_slice($errors, 0, 5)) . ($totalErrors > 5 ? '...' : '') . ')';
                }
            } else {
                $response['message'] = 'Failed to update structure: ' . implode('; ', array_slice($errors, 0, 5)) . ($totalErrors > 5 ? '...' : '');
            }
            break;
            
        case 'run_sql':
            // Run SQL on selected databases
            $databaseIds = $inputData['database_ids'] ?? [];
            $sql = trim($inputData['sql'] ?? '');
            
            if (empty($databaseIds) || !is_array($databaseIds)) {
                $response['message'] = 'Database IDs are required';
                break;
            }
            
            if (empty($sql)) {
                $response['message'] = 'SQL statement is required';
                break;
            }
            
            // Get database credentials from config.php
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            // Get selected databases
            $placeholders = implode(',', array_fill(0, count($databaseIds), '?'));
            $getStmt = $conn->prepare("SELECT id, db_name FROM dbs WHERE id IN ($placeholders)");
            $getStmt->execute($databaseIds);
            $selectedDbs = $getStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($selectedDbs)) {
                $response['message'] = 'No databases found for the selected IDs';
                break;
            }
            
            $results = [];
            
            // Execute SQL on each database
            foreach ($selectedDbs as $db) {
                if (empty($db['db_name'])) {
                    $results[] = [
                        'db_id' => $db['id'],
                        'db_name' => 'Unknown',
                        'error' => 'Database name is not set'
                    ];
                    continue;
                }
                
                try {
                    $targetConn = new PDO(
                        "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                        $targetDbUser,
                        $targetDbPass
                    );
                    $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                    
                    // Determine if this is a SELECT query or other type
                    $sqlUpper = strtoupper(trim($sql));
                    $isSelect = strpos($sqlUpper, 'SELECT') === 0;
                    
                    if ($isSelect) {
                        // For SELECT queries, fetch results
                        $stmt = $targetConn->prepare($sql);
                        $stmt->execute();
                        $rows = $stmt->fetchAll(PDO::FETCH_ASSOC);
                        $results[] = [
                            'db_id' => $db['id'],
                            'db_name' => $db['db_name'],
                            'rows' => $rows,
                            'row_count' => count($rows)
                        ];
                    } else {
                        // For other queries (INSERT, UPDATE, DELETE, etc.), get affected rows
                        $stmt = $targetConn->prepare($sql);
                        $stmt->execute();
                        $affectedRows = $stmt->rowCount();
                        $results[] = [
                            'db_id' => $db['id'],
                            'db_name' => $db['db_name'],
                            'affected_rows' => $affectedRows
                        ];
                    }
                } catch (PDOException $e) {
                    $results[] = [
                        'db_id' => $db['id'],
                        'db_name' => $db['db_name'],
                        'error' => $e->getMessage()
                    ];
                }
            }
            
            $response['success'] = true;
            $response['message'] = 'SQL executed on ' . count($results) . ' database(s)';
            $response['data'] = $results;
            break;
            
        case 'check_assets':
            // Check which asset files exist in files_dir
            $filesDir = trim($inputData['files_dir'] ?? '');
            
            if (empty($filesDir)) {
                $response['message'] = 'files_dir is required';
                break;
            }
            
            try {
                // Remove leading slash
                $filesDir = ltrim($filesDir, '/');
                $filesDir = rtrim($filesDir, '/');
                
                // Construct full path
                $filesDirPath = __DIR__ . '/../' . $filesDir . '/';
                
                // Security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve and verify files_dir path
                $resolvedFilesDir = realpath($filesDirPath);
                if ($resolvedFilesDir === false) {
                    // Directory doesn't exist
                    $response['success'] = true;
                    $response['data'] = ['existing' => []];
                    break;
                }
                $resolvedFilesDir = rtrim($resolvedFilesDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFilesDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                // Check which files exist
                $assetFiles = ['logo.png', 'letter_head.png', 'gml2.png'];
                $existingFiles = [];
                
                foreach ($assetFiles as $fileName) {
                    $filePath = $resolvedFilesDir . $fileName;
                    if (file_exists($filePath) && is_file($filePath)) {
                        $existingFiles[] = $fileName;
                    }
                }
                
                $response['success'] = true;
                $response['data'] = ['existing' => $existingFiles];
            } catch (Exception $e) {
                $response['message'] = $e->getMessage();
            }
            break;
            
        case 'copy_assets':
            // Copy asset files (logo.png, letter_head.png, gml2.png) from js_dir to files_dir
            $jsDir = trim($inputData['js_dir'] ?? '');
            $filesDir = trim($inputData['files_dir'] ?? '');
            
            if (empty($jsDir) || empty($filesDir)) {
                $response['message'] = 'js_dir and files_dir are required';
                break;
            }
            
            try {
                // Remove leading slashes
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                $filesDir = ltrim($filesDir, '/');
                $filesDir = rtrim($filesDir, '/');
                
                // Construct full paths
                $jsDirPath = __DIR__ . '/../' . $jsDir . '/';
                $filesDirPath = __DIR__ . '/../' . $filesDir . '/';
                
                // Security check - get real base path
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve and verify js_dir path
                $resolvedJsDir = realpath($jsDirPath);
                if ($resolvedJsDir === false) {
                    throw new Exception('js_dir does not exist: ' . $jsDir);
                }
                $resolvedJsDir = rtrim($resolvedJsDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedJsDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected in js_dir');
                }
                
                // Ensure files_dir exists
                if (!file_exists($filesDirPath)) {
                    if (!mkdir($filesDirPath, 0755, true)) {
                        throw new Exception('Failed to create files_dir: ' . $filesDirPath);
                    }
                }
                
                // Resolve and verify files_dir path
                $resolvedFilesDir = realpath($filesDirPath);
                if ($resolvedFilesDir === false) {
                    throw new Exception('Failed to resolve files_dir path');
                }
                $resolvedFilesDir = rtrim($resolvedFilesDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFilesDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected in files_dir');
                }
                
                // Files to copy
                $assetFiles = ['logo.png', 'letter_head.png', 'gml2.png'];
                $copiedFiles = [];
                $errors = [];
                
                foreach ($assetFiles as $fileName) {
                    $sourceFile = $resolvedJsDir . $fileName;
                    $destFile = $resolvedFilesDir . $fileName;
                    
                    // Check if source file exists
                    if (!file_exists($sourceFile)) {
                        $errors[] = "Source file not found: {$fileName}";
                        continue;
                    }
                    
                    // Copy file
                    if (!copy($sourceFile, $destFile)) {
                        $errors[] = "Failed to copy: {$fileName}";
                        continue;
                    }
                    
                    // Set permissions
                    chmod($destFile, 0644);
                    
                    $copiedFiles[] = $fileName;
                }
                
                if (count($errors) > 0) {
                    $response['message'] = 'Some files failed to copy: ' . implode(', ', $errors);
                    $response['data'] = [
                        'copied' => $copiedFiles,
                        'errors' => $errors
                    ];
                } else {
                    $response['success'] = true;
                    $response['message'] = 'All files copied successfully';
                    $response['data'] = [
                        'copied' => $copiedFiles,
                        'files_dir' => $filesDir
                    ];
                }
            } catch (Exception $e) {
                $response['message'] = $e->getMessage();
            }
            break;
            
        case 'ensure_folder':
            // Ensure folder exists: create if doesn't exist (without clearing contents)
            $jsDir = trim($inputData['js_dir'] ?? '');
            
            if (empty($jsDir)) {
                $response['message'] = 'js_dir is required';
                break;
            }
            
            try {
                // Remove leading slash if present (Unix path format, but we use it relative to project root)
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Construct the full folder path (relative to project root)
                $folderPath = __DIR__ . '/../' . $jsDir . '/';
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve the folder path
                $resolvedFolderPath = realpath($folderPath);
                if ($resolvedFolderPath === false) {
                    // Folder doesn't exist, create it
                    if (!mkdir($folderPath, 0755, true)) {
                        throw new Exception('Failed to create folder: ' . $folderPath);
                    }
                    $resolvedFolderPath = realpath($folderPath);
                    if ($resolvedFolderPath === false) {
                        throw new Exception('Failed to resolve created folder path');
                    }
                }
                
                // Verify the resolved path is within the base directory
                $resolvedFolderPath = rtrim($resolvedFolderPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFolderPath, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                $response['success'] = true;
                $response['message'] = 'Folder ensured successfully';
                $response['data'] = [
                    'path' => $resolvedFolderPath
                ];
            } catch (Exception $e) {
                $response['message'] = $e->getMessage();
            }
            break;
            
        case 'prepare_upload_folder':
            // Prepare upload folder: create if doesn't exist, clear if exists
            $databaseId = intval($inputData['database_id'] ?? 0);
            $jsDir = trim($inputData['js_dir'] ?? '');
            
            if ($databaseId <= 0 || empty($jsDir)) {
                $response['message'] = 'Invalid database ID or js_dir';
                break;
            }
            
            try {
                // Remove leading slash if present (Unix path format, but we use it relative to project root)
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Construct the full folder path (relative to project root)
                $folderPath = __DIR__ . '/../' . $jsDir . '/';
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve the folder path
                $resolvedFolderPath = realpath($folderPath);
                if ($resolvedFolderPath === false) {
                    // Folder doesn't exist, create it
                    if (!mkdir($folderPath, 0755, true)) {
                        throw new Exception('Failed to create folder: ' . $folderPath);
                    }
                    $resolvedFolderPath = realpath($folderPath);
                    if ($resolvedFolderPath === false) {
                        throw new Exception('Failed to resolve created folder path');
                    }
                }
                
                // Verify the resolved path is within the base directory
                $resolvedFolderPath = rtrim($resolvedFolderPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFolderPath, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                // Protected files that should NOT be deleted
                $protectedFiles = ['logo.png', 'letter_head.png', 'gml2.png', 'db_code.json'];
                
                // Delete all contents (files and subdirectories) EXCEPT protected files
                $iterator = new RecursiveIteratorIterator(
                    new RecursiveDirectoryIterator($resolvedFolderPath, RecursiveDirectoryIterator::SKIP_DOTS),
                    RecursiveIteratorIterator::CHILD_FIRST
                );
                
                $deletedFiles = [];
                $skippedFiles = [];
                
                foreach ($iterator as $path) {
                    $fileName = $path->getFilename();
                    $relativePath = str_replace($resolvedFolderPath, '', $path->getPathname());
                    
                    // Skip protected files (only check files in root of js_dir, not subdirectories)
                    if (!$path->isDir() && strpos($relativePath, DIRECTORY_SEPARATOR) === false) {
                        if (in_array($fileName, $protectedFiles)) {
                            $skippedFiles[] = $fileName;
                            continue; // Skip deleting protected files
                        }
                    }
                    
                    // Delete the file or directory
                    if ($path->isDir()) {
                        rmdir($path->getPathname());
                    } else {
                        unlink($path->getPathname());
                        $deletedFiles[] = $fileName;
                    }
                }
                
                // Log what was deleted and what was preserved
                error_log('[prepare_upload_folder] Deleted files: ' . implode(', ', $deletedFiles));
                if (count($skippedFiles) > 0) {
                    error_log('[prepare_upload_folder] Preserved protected files: ' . implode(', ', $skippedFiles));
                }
                
                $response['success'] = true;
                $message = 'Folder prepared successfully';
                if (count($skippedFiles) > 0) {
                    $message .= '. Preserved protected files: ' . implode(', ', $skippedFiles);
                }
                $response['message'] = $message;
                $response['data'] = [
                    'deleted_count' => count($deletedFiles),
                    'preserved_files' => $skippedFiles
                ];
            } catch (Exception $e) {
                $response['message'] = 'Error preparing folder: ' . $e->getMessage();
            }
            break;
            
        case 'read_db_code_json':
            // Read db_code.json file from js_dir
            $databaseId = intval($inputData['database_id'] ?? 0);
            
            if ($databaseId <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            try {
                // Get database record
                $stmt = $conn->prepare("SELECT js_dir FROM dbs WHERE id = ?");
                $stmt->execute([$databaseId]);
                $db = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$db) {
                    $response['message'] = 'Database not found';
                    break;
                }
                
                $jsDir = trim($db['js_dir'] ?? '');
                if (empty($jsDir)) {
                    $response['message'] = 'js_dir is not configured for this database';
                    break;
                }
                
                // Remove leading slash if present
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Construct the full file path
                $filePath = __DIR__ . '/../' . $jsDir . '/db_code.json';
                
                // Debug logging
                error_log('[read_db_code_json] Database ID: ' . $databaseId);
                error_log('[read_db_code_json] js_dir from DB: ' . $jsDir);
                error_log('[read_db_code_json] __DIR__: ' . __DIR__);
                error_log('[read_db_code_json] Constructed filePath: ' . $filePath);
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                error_log('[read_db_code_json] realBasePath: ' . $realBasePath);
                
                // Resolve the file path
                $resolvedFilePath = realpath($filePath);
                error_log('[read_db_code_json] resolvedFilePath: ' . ($resolvedFilePath !== false ? $resolvedFilePath : 'FALSE'));
                
                // Also check if file exists using the constructed path directly
                $fileExistsDirect = file_exists($filePath);
                error_log('[read_db_code_json] file_exists($filePath): ' . ($fileExistsDirect ? 'TRUE' : 'FALSE'));
                
                // Verify the resolved path is within the base directory
                if ($resolvedFilePath !== false) {
                    $resolvedDir = dirname($resolvedFilePath);
                    if (strpos($resolvedDir, $realBasePath) !== 0) {
                        throw new Exception('Directory traversal attempt detected');
                    }
                }
                
                // Read file if exists, otherwise return default structure
                // Use both checks: realpath might fail even if file exists (e.g., symlinks)
                if (($resolvedFilePath !== false && file_exists($resolvedFilePath)) || $fileExistsDirect) {
                    // Use resolved path if available, otherwise use constructed path
                    $actualFilePath = $resolvedFilePath !== false ? $resolvedFilePath : $filePath;
                    error_log('[read_db_code_json] Reading file from: ' . $actualFilePath);
                    
                    $fileContent = file_get_contents($actualFilePath);
                    error_log('[read_db_code_json] File content length: ' . strlen($fileContent));
                    error_log('[read_db_code_json] File content: ' . substr($fileContent, 0, 200));
                    
                    $jsonData = json_decode($fileContent, true);
                    
                    if (json_last_error() !== JSON_ERROR_NONE) {
                        error_log('[read_db_code_json] JSON decode error: ' . json_last_error_msg());
                        throw new Exception('Invalid JSON in db_code.json: ' . json_last_error_msg());
                    }
                    
                    error_log('[read_db_code_json] Parsed JSON data: ' . json_encode($jsonData));
                    
                    $response['success'] = true;
                    $response['message'] = 'File read successfully';
                    $response['data'] = [
                        'content' => $jsonData,
                        'exists' => true
                    ];
                } else {
                    // File doesn't exist, return default structure
                    error_log('[read_db_code_json] File not found. resolvedFilePath: ' . ($resolvedFilePath !== false ? $resolvedFilePath : 'FALSE') . ', fileExistsDirect: ' . ($fileExistsDirect ? 'TRUE' : 'FALSE'));
                    $response['success'] = true;
                    $response['message'] = 'File does not exist, returning default structure';
                    $response['data'] = [
                        'content' => ['db_code' => ''],
                        'exists' => false,
                        'debug' => [
                            'js_dir' => $jsDir,
                            'constructed_path' => $filePath,
                            'resolved_path' => $resolvedFilePath !== false ? $resolvedFilePath : null,
                            'file_exists_direct' => $fileExistsDirect,
                            'base_path' => $realBasePath
                        ]
                    ];
                }
            } catch (Exception $e) {
                $response['message'] = 'Error reading file: ' . $e->getMessage();
            }
            break;
            
        case 'write_db_code_json':
            // Write db_code.json file to js_dir
            $databaseId = intval($inputData['database_id'] ?? 0);
            $jsonContent = $inputData['content'] ?? null;
            
            if ($databaseId <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            if ($jsonContent === null) {
                $response['message'] = 'JSON content is required';
                break;
            }
            
            try {
                // Validate JSON content
                if (!is_array($jsonContent)) {
                    throw new Exception('Content must be a valid JSON object');
                }
                
                // Get database record
                $stmt = $conn->prepare("SELECT js_dir FROM dbs WHERE id = ?");
                $stmt->execute([$databaseId]);
                $db = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$db) {
                    $response['message'] = 'Database not found';
                    break;
                }
                
                $jsDir = trim($db['js_dir'] ?? '');
                if (empty($jsDir)) {
                    $response['message'] = 'js_dir is not configured for this database';
                    break;
                }
                
                // Remove leading slash if present
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Construct the full directory and file path
                $dirPath = __DIR__ . '/../' . $jsDir;
                $filePath = $dirPath . '/db_code.json';
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Create directory if it doesn't exist
                if (!is_dir($dirPath)) {
                    if (!mkdir($dirPath, 0755, true)) {
                        throw new Exception('Failed to create directory: ' . $dirPath);
                    }
                }
                
                // Verify the directory path is within the base directory
                $resolvedDirPath = realpath($dirPath);
                if ($resolvedDirPath === false) {
                    throw new Exception('Failed to resolve directory path');
                }
                $resolvedDirPath = rtrim($resolvedDirPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedDirPath, $realBasePath) !== 0) {
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
                
                $response['success'] = true;
                $response['message'] = 'File written successfully';
                $response['data'] = [
                    'path' => $jsDir . '/db_code.json'
                ];
            } catch (Exception $e) {
                $response['message'] = 'Error writing file: ' . $e->getMessage();
            }
            break;
            
        case 'check_file_exists':
            // Check if a file exists in js_dir
            $databaseId = intval($inputData['database_id'] ?? 0);
            $fileName = trim($inputData['file_name'] ?? '');
            $jsDir = trim($inputData['js_dir'] ?? '');
            
            if ($databaseId <= 0 || empty($fileName)) {
                $response['message'] = 'Invalid database ID or file name';
                break;
            }
            
            if (empty($jsDir)) {
                // Get js_dir from database if not provided
                $stmt = $conn->prepare("SELECT js_dir FROM dbs WHERE id = ?");
                $stmt->execute([$databaseId]);
                $db = $stmt->fetch(PDO::FETCH_ASSOC);
                if (!$db || empty($db['js_dir'])) {
                    $response['message'] = 'js_dir is not configured for this database';
                    break;
                }
                $jsDir = trim($db['js_dir']);
            }
            
            try {
                // Remove leading slash if present
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Sanitize filename
                $fileName = basename($fileName);
                $fileName = str_replace('..', '', $fileName);
                
                // Construct the full file path
                $filePath = __DIR__ . '/../' . $jsDir . '/' . $fileName;
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve the file path
                $resolvedFilePath = realpath($filePath);
                
                // Verify the resolved path is within the base directory
                if ($resolvedFilePath !== false) {
                    $resolvedDir = dirname($resolvedFilePath);
                    if (strpos($resolvedDir, $realBasePath) !== 0) {
                        throw new Exception('Directory traversal attempt detected');
                    }
                }
                
                $fileExists = file_exists($filePath);
                
                $response['success'] = true;
                $response['message'] = $fileExists ? 'File exists' : 'File does not exist';
                $response['data'] = [
                    'exists' => $fileExists,
                    'file_path' => $jsDir . '/' . $fileName
                ];
            } catch (Exception $e) {
                $response['message'] = 'Error checking file: ' . $e->getMessage();
            }
            break;
            
        default:
            $response['message'] = 'No action specified';
            break;
    }
    
} catch (Exception $e) {
    error_log('DB Manager API Error: ' . $e->getMessage());
    $response['message'] = $e->getMessage();
    $response['success'] = false;
    http_response_code(500);
} catch (Error $e) {
    error_log('DB Manager API Fatal Error: ' . $e->getMessage());
    $response['message'] = 'Fatal error: ' . $e->getMessage();
    $response['success'] = false;
    http_response_code(500);
}

// Ensure headers are sent (re-send CORS headers in case of error)
if (!headers_sent()) {
    header('Access-Control-Allow-Origin: *');
    header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
    header('Access-Control-Allow-Headers: Content-Type, Accept, Authorization, X-Requested-With');
    header('Content-Type: application/json');
}

// Clean output buffer and send JSON response
if (ob_get_level()) {
    ob_end_clean();
}
echo json_encode($response);
exit;

