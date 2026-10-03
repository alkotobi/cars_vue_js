<?php
// Streams a full SQL dump of the app database.
//
// This endpoint had no authentication whatsoever and answered a plain GET, so
// anyone could pull every table - including users.password (bcrypt) and
// users.api_token, the live bearer credential - with one request. It is now
// admin-only.
//
// The mysqldump invocation also interpolated its arguments into a shell string
// and passed the password on the command line, where any local user can read it
// out of ps. Both are fixed below: escapeshellarg() on every value, and the
// password moved into a temporary defaults file that is removed afterwards.

ob_start();

header('Content-Type: application/sql');

require_once __DIR__ . '/lib/cors.php';
api_send_cors_headers();
header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Authorization, X-Api-Token');

// Disable all error reporting and output
ini_set('display_errors', 0);
ini_set('display_startup_errors', 0);
error_reporting(0);

// Handle preflight requests
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    exit();
}

ob_clean();

try {
    require_once __DIR__ . '/config.php';
    require_once __DIR__ . '/lib/auth.php';

    $pdo = new PDO(
        "mysql:host={$db_config['host']};dbname={$db_config['dbname']}",
        $db_config['user'],
        $db_config['pass']
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    // Gate on an admin token. Accepted as a query parameter because this is
    // fetched by a plain link, and from the header when a caller can set one.
    $token = $_SERVER['HTTP_X_API_TOKEN'] ?? ($_GET['token'] ?? '');
    require_api_admin($pdo, ['token' => $token]);

    // Filename is caller-influenced, so keep it to a safe character set rather
    // than reflecting it into the header verbatim.
    $filename = $_GET['filename'] ?? ('merhab_cars_backup_' . date('Y-m-d_H-i-s') . '.sql');
    $filename = preg_replace('/[^A-Za-z0-9._-]/', '_', (string)$filename);

    header('Content-Disposition: attachment; filename="' . $filename . '"');
    header('Cache-Control: no-cache, must-revalidate');
    header('Expires: Sat, 26 Jul 1997 05:00:00 GMT');

    $host = $db_config['host'];
    $dbname = $db_config['dbname'];
    $username = $db_config['user'];
    $password = $db_config['pass'];

    // Try mysqldump first
    $mysqldump_available = false;
    $mysqldump_output = [];
    $return_var = 0;

    exec('command -v mysqldump 2>&1', $mysqldump_output, $return_var);
    if ($return_var === 0) {
        $mysqldump_path = trim($mysqldump_output[0]);

        // Password goes in a private defaults file rather than argv, so it does not
        // appear in ps output or in this process's command line.
        $defaultsFile = tempnam(sys_get_temp_dir(), 'cars-dump-');
        if ($defaultsFile !== false) {
            file_put_contents($defaultsFile, "[client]\nuser=" . $username . "\npassword=" . $password . "\n");
            @chmod($defaultsFile, 0600);
        }

        $command = escapeshellarg($mysqldump_path)
            . ($defaultsFile !== false ? ' --defaults-extra-file=' . escapeshellarg($defaultsFile) : '')
            . ' --host=' . escapeshellarg($host)
            . ' --database=' . escapeshellarg($dbname);

        // Capture stdout; stderr is dropped rather than merged in, because mysqldump's
        // "Using a password on the command line interface can be insecure" warning
        // was being written into the dump itself.
        $descriptors = [1 => ['pipe', 'w'], 2 => ['pipe', 'w']];
        $process = proc_open($command, $descriptors, $pipes);
        if (is_resource($process)) {
            $dump = stream_get_contents($pipes[1]);
            fclose($pipes[1]);
            fclose($pipes[2]);
            $return_var = proc_close($process);

            if ($return_var === 0 && $dump !== false && trim((string)$dump) !== '') {
                $mysqldump_available = true;
                echo $dump;
            }
        }

        if ($defaultsFile !== false && file_exists($defaultsFile)) {
            @unlink($defaultsFile);
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

    // The failure text used to name the database host, database and username,
    // which turns a misconfigured box into a free reconnaissance report. Log the
    // detail for the operator and tell the caller only that it failed.
    error_log('backup.php failed: ' . $e->getMessage());

    echo "-- Merhab Cars Database Backup\n";
    echo "-- Generated on: " . date('Y-m-d_H:i:s') . "\n";
    echo "-- ERROR: Backup failed\n";
    echo "-- Error details: see the server error log\n";
}

// Flush and end output
ob_end_flush();
exit();
?> 