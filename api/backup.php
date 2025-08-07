<?php
// Force output buffering to prevent any HTML leakage
ob_start();

// Set headers immediately to prevent any HTML output
header('Content-Type: application/sql');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Authorization');

// Disable all error reporting and output
ini_set('display_errors', 0);
ini_set('display_startup_errors', 0);
error_reporting(0);

// Handle preflight requests
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit();
}

// Clear any existing output
ob_clean();

try {
    // Include database configuration
    require_once 'config.php';
    
    // Get filename from request
    $filename = $_GET['filename'] ?? 'merhab_cars_backup_' . date('Y-m-d_H-i-s') . '.sql';
    
    // Set headers for file download
    header('Content-Disposition: attachment; filename="' . $filename . '"');
    header('Cache-Control: no-cache, must-revalidate');
    header('Expires: Sat, 26 Jul 1997 05:00:00 GMT');
    
    // Database connection details
    $host = $db_config['host'];
    $dbname = $db_config['dbname'];
    $username = $db_config['user'];
    $password = $db_config['pass'];
    
    // Create database connection first
    $pdo = new PDO(
        "mysql:host={$host};dbname={$dbname}", 
        $username, 
        $password
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    // Try mysqldump first
    $mysqldump_available = false;
    $mysqldump_output = [];
    $return_var = 0;
    
    // Test if mysqldump is available
    exec("which mysqldump 2>&1", $mysqldump_output, $return_var);
    if ($return_var === 0) {
        $mysqldump_path = trim($mysqldump_output[0]);
        $command = "$mysqldump_path --host=$host --user=$username --password=$password $dbname";
        
        // Execute mysqldump
        $output = [];
        $return_var = 0;
        exec($command . " 2>&1", $output, $return_var);
        
        if ($return_var === 0 && !empty($output)) {
            $mysqldump_available = true;
            echo implode("\n", $output);
        }
    }
    
    // If mysqldump failed or not available, use PHP method
    if (!$mysqldump_available) {
        // Get all tables
        $tables = [];
        $stmt = $pdo->query("SHOW TABLES");
        while ($row = $stmt->fetch(PDO::FETCH_NUM)) {
            $tables[] = $row[0];
        }
        
        // Generate SQL backup
        $backup = "-- Merhab Cars Database Backup\n";
        $backup .= "-- Generated on: " . date('Y-m-d H:i:s') . "\n";
        $backup .= "-- Database: $dbname\n";
        $backup .= "-- Backup method: PHP (mysqldump not available)\n\n";
        
        foreach ($tables as $table) {
            // Get table structure
            $stmt = $pdo->query("SHOW CREATE TABLE `$table`");
            $row = $stmt->fetch(PDO::FETCH_NUM);
            $backup .= $row[1] . ";\n\n";
            
            // Get table data
            $stmt = $pdo->query("SELECT * FROM `$table`");
            $rows = $stmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (!empty($rows)) {
                $backup .= "INSERT INTO `$table` VALUES\n";
                $values = [];
                foreach ($rows as $row) {
                    $rowValues = [];
                    foreach ($row as $value) {
                        if ($value === null) {
                            $rowValues[] = 'NULL';
                        } else {
                            $rowValues[] = "'" . addslashes($value) . "'";
                        }
                    }
                    $values[] = "(" . implode(', ', $rowValues) . ")";
                }
                $backup .= implode(",\n", $values) . ";\n\n";
            }
        }
        
        echo $backup;
    }
    
} catch (Exception $e) {
    // Clear any output and create error SQL file
    ob_clean();
    
    echo "-- Merhab Cars Database Backup\n";
    echo "-- Generated on: " . date('Y-m-d H:i:s') . "\n";
    echo "-- ERROR: Backup failed\n";
    echo "-- Error details: " . $e->getMessage() . "\n";
    echo "-- Please check database configuration and permissions\n";
    echo "-- Database host: " . ($host ?? 'unknown') . "\n";
    echo "-- Database name: " . ($dbname ?? 'unknown') . "\n";
    echo "-- Username: " . ($username ?? 'unknown') . "\n";
    echo "-- PHP Version: " . phpversion() . "\n";
    echo "-- Server: " . ($_SERVER['SERVER_SOFTWARE'] ?? 'unknown') . "\n";
}

// Flush and end output
ob_end_flush();
exit();
?> 