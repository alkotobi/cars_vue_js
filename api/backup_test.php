<?php
// Diagnostic script to test backup functionality
// This will help identify issues on the remote server

header('Content-Type: text/plain');
echo "=== Merhab Cars Backup Diagnostic ===\n\n";

// Test 1: Check PHP version and extensions
echo "1. PHP Environment:\n";
echo "PHP Version: " . phpversion() . "\n";
echo "PDO MySQL: " . (extension_loaded('pdo_mysql') ? 'Available' : 'NOT AVAILABLE') . "\n";
echo "exec function: " . (function_exists('exec') ? 'Available' : 'NOT AVAILABLE') . "\n";
echo "mysqldump command: ";

$mysqldump_output = [];
$return_var = 0;
exec("which mysqldump 2>&1", $mysqldump_output, $return_var);
if ($return_var === 0) {
    echo trim($mysqldump_output[0]) . "\n";
} else {
    echo "NOT FOUND\n";
}

echo "\n2. Database Configuration:\n";
require_once 'config.php';
echo "Host: " . $db_config['host'] . "\n";
echo "Database: " . $db_config['dbname'] . "\n";
echo "User: " . $db_config['user'] . "\n";
echo "Password: " . (empty($db_config['pass']) ? 'Empty' : 'Set') . "\n";

echo "\n3. Database Connection Test:\n";
try {
    $pdo = new PDO(
        "mysql:host={$db_config['host']};dbname={$db_config['dbname']}", 
        $db_config['user'], 
        $db_config['pass']
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    echo "✓ Database connection successful\n";
    
    // Test table count
    $stmt = $pdo->query("SHOW TABLES");
    $tables = $stmt->fetchAll(PDO::FETCH_NUM);
    echo "✓ Found " . count($tables) . " tables\n";
    
} catch (Exception $e) {
    echo "✗ Database connection failed: " . $e->getMessage() . "\n";
}

echo "\n4. mysqldump Test:\n";
$command = "mysqldump --host={$db_config['host']} --user={$db_config['user']} --password={$db_config['pass']} {$db_config['dbname']} --no-data --tables users 2>&1";
$output = [];
$return_var = 0;
exec($command, $output, $return_var);

if ($return_var === 0) {
    echo "✓ mysqldump test successful\n";
    echo "Output preview: " . substr(implode("\n", $output), 0, 100) . "...\n";
} else {
    echo "✗ mysqldump test failed\n";
    echo "Error: " . implode("\n", $output) . "\n";
}

echo "\n5. File Permissions:\n";
echo "Current directory writable: " . (is_writable('.') ? 'Yes' : 'No') . "\n";
echo "api directory writable: " . (is_writable('./api') ? 'Yes' : 'No') . "\n";

echo "\n6. Server Information:\n";
echo "Server Software: " . ($_SERVER['SERVER_SOFTWARE'] ?? 'Unknown') . "\n";
echo "Document Root: " . ($_SERVER['DOCUMENT_ROOT'] ?? 'Unknown') . "\n";
echo "Script Path: " . __FILE__ . "\n";

echo "\n=== Diagnostic Complete ===\n";
?> 