<?php
// Prevent any HTML output by setting proper headers first
header('Content-Type: application/sql');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Authorization');

// Handle preflight requests
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit();
}

// Disable error display to prevent HTML output
ini_set('display_errors', 0);
error_reporting(E_ALL);

// Include database configuration
require_once 'config.php';

// Get filename from request
$filename = $_GET['filename'] ?? 'merhab_cars_backup_' . date('Y-m-d_H-i-s') . '.sql';

// Set headers for file download
header('Content-Disposition: attachment; filename="' . $filename . '"');
header('Cache-Control: no-cache, must-revalidate');
header('Expires: Sat, 26 Jul 1997 05:00:00 GMT');

try {
    // Database connection details
    $host = $db_config['host'];
    $dbname = $db_config['dbname'];
    $username = $db_config['user'];
    $password = $db_config['pass'];
    
    // Create backup using mysqldump
    $command = "mysqldump --host=$host --user=$username --password=$password $dbname";
    
    // Execute the command
    $output = [];
    $return_var = 0;
    exec($command . " 2>&1", $output, $return_var);
    
    if ($return_var !== 0) {
        throw new Exception('mysqldump failed: ' . implode("\n", $output));
    }
    
    // Output the backup content
    echo implode("\n", $output);
    
} catch (Exception $e) {
    // If mysqldump fails, try alternative method using PHP
    try {
        // Create database connection
        $pdo = new PDO(
            "mysql:host={$db_config['host']};dbname={$db_config['dbname']}", 
            $db_config['user'], 
            $db_config['pass']
        );
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
        
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
        $backup .= "-- Backup method: PHP fallback (mysqldump failed)\n\n";
        
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
        
    } catch (Exception $e2) {
        // If everything fails, output a minimal SQL file with error info
        echo "-- Merhab Cars Database Backup\n";
        echo "-- Generated on: " . date('Y-m-d H:i:s') . "\n";
        echo "-- ERROR: Backup failed\n";
        echo "-- Error details: " . $e2->getMessage() . "\n";
        echo "-- Please check database configuration and permissions\n";
        echo "-- Database host: $host\n";
        echo "-- Database name: $dbname\n";
        echo "-- Username: $username\n";
    }
}
?> 