<?php
// Simple backup web interface
// This should work on any server configuration

// Disable error display
ini_set('display_errors', 0);
error_reporting(0);

// Automatic backup functionality
function createAutomaticBackup() {
    try {
        require_once 'config.php';
        
        // Create backup directory
        $backup_dir = '../backups';
        if (!is_dir($backup_dir)) {
            mkdir($backup_dir, 0755, true);
        }
        
        // Generate filename with timestamp
        $timestamp = date('Y-m-d_H-i-s');
        $filename = "auto_backup_{$timestamp}.sql";
        $filepath = $backup_dir . '/' . $filename;
        
        // Database connection
        $host = $db_config['host'];
        $dbname = $db_config['dbname'];
        $username = $db_config['user'];
        $password = $db_config['pass'];
        
        // Try mysqldump
        $command = "mysqldump --host=$host --user=$username --password=$password $dbname > $filepath 2>&1";
        $output = [];
        $return_var = 0;
        exec($command, $output, $return_var);
        
        if ($return_var !== 0 || !file_exists($filepath) || filesize($filepath) === 0) {
            // Use PHP method
            $pdo = new PDO(
                "mysql:host={$host};dbname={$dbname}", 
                $username, 
                $password
            );
            $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
            
            // Get all tables
            $tables = [];
            $stmt = $pdo->query("SHOW TABLES");
            while ($row = $stmt->fetch(PDO::FETCH_NUM)) {
                $tables[] = $row[0];
            }
            
            // Generate backup content
            $backup_content = "-- Merhab Cars Automatic Database Backup\n";
            $backup_content .= "-- Generated on: " . date('Y-m-d H:i:s') . "\n";
            $backup_content .= "-- Database: $dbname\n";
            $backup_content .= "-- Backup type: Automatic (30 min interval)\n\n";
            
            foreach ($tables as $table) {
                // Get table structure
                $stmt = $pdo->query("SHOW CREATE TABLE `$table`");
                $row = $stmt->fetch(PDO::FETCH_NUM);
                $backup_content .= $row[1] . ";\n\n";
                
                // Get table data
                $stmt = $pdo->query("SELECT * FROM `$table`");
                $rows = $stmt->fetchAll(PDO::FETCH_ASSOC);
                
                if (!empty($rows)) {
                    $backup_content .= "INSERT INTO `$table` VALUES\n";
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
                    $backup_content .= implode(",\n", $values) . ";\n\n";
                }
            }
            
            file_put_contents($filepath, $backup_content);
        }
        
        if (file_exists($filepath) && filesize($filepath) > 0) {
            // Log the automatic backup
            $log_file = $backup_dir . '/auto_backup_log.txt';
            $log_entry = date('Y-m-d H:i:s') . " - Automatic backup created: $filename (" . filesize($filepath) . " bytes)\n";
            file_put_contents($log_file, $log_entry, FILE_APPEND);
            
            return [
                'success' => true,
                'filename' => $filename,
                'filepath' => $filepath,
                'size' => filesize($filepath)
            ];
        } else {
            throw new Exception('Failed to create automatic backup file');
        }
        
    } catch (Exception $e) {
        // Log the error
        $log_file = '../backups/auto_backup_log.txt';
        $log_entry = date('Y-m-d H:i:s') . " - Automatic backup failed: " . $e->getMessage() . "\n";
        file_put_contents($log_file, $log_entry, FILE_APPEND);
        
        return [
            'success' => false,
            'error' => $e->getMessage()
        ];
    }
}

// Check if automatic backup should run (every 30 minutes)
$auto_backup_enabled = true;
$last_auto_backup_file = '../backups/last_auto_backup.txt';
$auto_backup_interval = 30 * 60; // 30 minutes in seconds

// Handle automatic backup trigger
if (isset($_POST['trigger_auto_backup'])) {
    $auto_backup_result = createAutomaticBackup();
    if ($auto_backup_result['success']) {
        // Update last backup time
        file_put_contents($last_auto_backup_file, date('Y-m-d H:i:s'));
        
        // Set flag for JavaScript to trigger download
        $auto_backup_ready = true;
        $auto_backup_filename = $auto_backup_result['filename'];
        $auto_backup_size = $auto_backup_result['size'];
        
        // Return JSON response for AJAX
        if (isset($_POST['ajax'])) {
            header('Content-Type: application/json');
            echo json_encode([
                'success' => true,
                'filename' => $auto_backup_filename,
                'size' => $auto_backup_size,
                'message' => 'Automatic backup created successfully'
            ]);
            exit();
        }
    }
}

// Handle interval setting
if (isset($_POST['set_interval'])) {
    $new_interval = intval($_POST['interval_seconds']);
    if ($new_interval > 0) {
        $interval_file = '../backups/backup_interval.txt';
        file_put_contents($interval_file, $new_interval);
        $interval_success = "Interval updated to $new_interval seconds";
    } else {
        $interval_error = "Invalid interval. Please enter a positive number.";
    }
}

// Handle auto-download setting
if (isset($_POST['set_auto_download'])) {
    $auto_download = isset($_POST['auto_download']) ? '1' : '0';
    $auto_download_file = '../backups/auto_download.txt';
    file_put_contents($auto_download_file, $auto_download);
    $interval_success = "Auto-download setting updated";
}

// Load current interval
$interval_file = '../backups/backup_interval.txt';
if (file_exists($interval_file)) {
    $auto_backup_interval = intval(file_get_contents($interval_file));
    if ($auto_backup_interval <= 0) {
        $auto_backup_interval = 30 * 60; // Default to 30 minutes
    }
} else {
    $auto_backup_interval = 30 * 60; // Default to 30 minutes
}

// Load current auto-download setting
$auto_download_file = '../backups/auto_download.txt';
$auto_download_enabled = file_exists($auto_download_file) ? file_get_contents($auto_download_file) === '1' : true;

if ($auto_backup_enabled) {
    $should_run_auto_backup = false;
    
    if (!file_exists($last_auto_backup_file)) {
        // First time running, create backup
        $should_run_auto_backup = true;
    } else {
        $last_backup_time = file_get_contents($last_auto_backup_file);
        $time_since_last_backup = time() - strtotime($last_backup_time);
        
        if ($time_since_last_backup >= $auto_backup_interval) {
            $should_run_auto_backup = true;
        }
    }
    
    if ($should_run_auto_backup) {
        $auto_backup_result = createAutomaticBackup();
        if ($auto_backup_result['success']) {
            // Update last backup time
            file_put_contents($last_auto_backup_file, date('Y-m-d H:i:s'));
            
            // Set flag for JavaScript to trigger download
            $auto_backup_ready = true;
            $auto_backup_filename = $auto_backup_result['filename'];
            $auto_backup_size = $auto_backup_result['size'];
        }
    }
}

// Handle backup creation
if (isset($_POST['create_backup'])) {
    try {
        require_once 'config.php';
        
        // Create backup directory
        $backup_dir = '../backups';
        if (!is_dir($backup_dir)) {
            mkdir($backup_dir, 0755, true);
        }
        
        // Generate filename
        $timestamp = date('Y-m-d_H-i-s');
        $filename = "merhab_cars_backup_{$timestamp}.sql";
        $filepath = $backup_dir . '/' . $filename;
        
        // Database connection
        $host = $db_config['host'];
        $dbname = $db_config['dbname'];
        $username = $db_config['user'];
        $password = $db_config['pass'];
        
        // Try mysqldump
        $command = "mysqldump --host=$host --user=$username --password=$password $dbname > $filepath 2>&1";
        $output = [];
        $return_var = 0;
        exec($command, $output, $return_var);
        
        if ($return_var !== 0 || !file_exists($filepath) || filesize($filepath) === 0) {
            // Use PHP method
            $pdo = new PDO(
                "mysql:host={$host};dbname={$dbname}", 
                $username, 
                $password
            );
            $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
            
            // Get all tables
            $tables = [];
            $stmt = $pdo->query("SHOW TABLES");
            while ($row = $stmt->fetch(PDO::FETCH_NUM)) {
                $tables[] = $row[0];
            }
            
            // Generate backup content
            $backup_content = "-- Merhab Cars Database Backup\n";
            $backup_content .= "-- Generated on: " . date('Y-m-d H:i:s') . "\n";
            $backup_content .= "-- Database: $dbname\n\n";
            
            foreach ($tables as $table) {
                // Get table structure
                $stmt = $pdo->query("SHOW CREATE TABLE `$table`");
                $row = $stmt->fetch(PDO::FETCH_NUM);
                $backup_content .= $row[1] . ";\n\n";
                
                // Get table data
                $stmt = $pdo->query("SELECT * FROM `$table`");
                $rows = $stmt->fetchAll(PDO::FETCH_ASSOC);
                
                if (!empty($rows)) {
                    $backup_content .= "INSERT INTO `$table` VALUES\n";
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
                    $backup_content .= implode(",\n", $values) . ";\n\n";
                }
            }
            
            file_put_contents($filepath, $backup_content);
        }
        
        if (file_exists($filepath) && filesize($filepath) > 0) {
            $file_size = filesize($filepath);
            $file_size_kb = round($file_size / 1024, 2);
            $download_url = "/backups/$filename";
            
            $success_message = "✓ Backup created successfully!\n\nFile: $filename\nSize: $file_size_kb KB\n\n<a href='$download_url' download='$filename' style='background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; display: inline-block; margin-top: 10px;'>Download Backup</a>";
        } else {
            throw new Exception('Failed to create backup file');
        }
        
    } catch (Exception $e) {
        $error_message = "✗ Backup failed: " . $e->getMessage();
    }
}

// Get list of existing backups
$backup_dir = '../backups';
$existing_backups = [];
if (is_dir($backup_dir)) {
    $files = glob($backup_dir . '/*.sql');
    foreach ($files as $file) {
        $existing_backups[] = [
            'filename' => basename($file),
            'size' => filesize($file),
            'created' => date('Y-m-d H:i:s', filemtime($file)),
            'download_url' => '/backups/' . basename($file),
            'is_auto' => strpos(basename($file), 'auto_backup_') === 0
        ];
    }
}

// Get automatic backup status
$auto_backup_status = '';
if (file_exists($last_auto_backup_file)) {
    $last_backup_time = file_get_contents($last_auto_backup_file);
    $time_since_last_backup = time() - strtotime($last_backup_time);
    $minutes_since_last = round($time_since_last_backup / 60, 1);
    $next_backup_in = max(0, ($auto_backup_interval / 60) - $minutes_since_last);
    
    $auto_backup_status = "Last automatic backup: $last_backup_time\nNext backup in: $next_backup_in minutes\nCurrent interval: " . round($auto_backup_interval / 60, 1) . " minutes";
} else {
    $auto_backup_status = "Automatic backup not yet run\nCurrent interval: " . round($auto_backup_interval / 60, 1) . " minutes";
}
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Merhab Cars - Database Backup</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .btn {
            background: #007bff;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            margin: 10px 5px;
        }
        .btn:hover {
            background: #0056b3;
        }
        .status {
            margin: 20px 0;
            padding: 15px;
            border-radius: 5px;
        }
        .success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .info {
            background: #d1ecf1;
            color: #0c5460;
            border: 1px solid #bee5eb;
        }
        .backup-item {
            background: #f8f9fa;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
            border: 1px solid #dee2e6;
        }
        .backup-item.auto {
            border-left: 4px solid #28a745;
        }
        .backup-item h3 {
            margin: 0 0 10px 0;
            color: #333;
        }
        .backup-item p {
            margin: 5px 0;
            color: #666;
        }
        .auto-badge {
            background: #28a745;
            color: white;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 12px;
            margin-left: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔄 Database Backup System</h1>
        
        <div style="text-align: center;">
            <form method="POST" style="display: inline;">
                <button type="submit" name="create_backup" class="btn">Create New Backup</button>
            </form>
        </div>
        
        <div style="margin: 20px 0; padding: 20px; background: #f8f9fa; border-radius: 5px;">
            <h3>⚙️ Automatic Backup Settings</h3>
            <form method="POST" style="display: flex; align-items: center; gap: 10px; justify-content: center; flex-wrap: wrap;">
                <label for="interval_seconds">Backup Interval (seconds):</label>
                <input type="number" id="interval_seconds" name="interval_seconds" value="<?php echo $auto_backup_interval; ?>" min="60" max="86400" style="padding: 8px; border: 1px solid #ddd; border-radius: 3px; width: 120px;">
                <button type="submit" name="set_interval" class="btn" style="background: #28a745;">Update Interval</button>
            </form>
            <form method="POST" style="margin-top: 10px; display: flex; align-items: center; gap: 10px; justify-content: center; flex-wrap: wrap;">
                <label style="display: flex; align-items: center; gap: 5px;">
                    <input type="checkbox" name="auto_download" <?php echo $auto_download_enabled ? 'checked' : ''; ?> style="margin: 0;">
                    Auto-download backups
                </label>
                <button type="submit" name="set_auto_download" class="btn" style="background: #17a2b8;">Update Setting</button>
            </form>
            <p style="margin: 10px 0 0 0; font-size: 14px; color: #666;">
                Minimum: 60 seconds (1 minute)<br>
                Maximum: 86400 seconds (24 hours)<br>
                Current: <?php echo $auto_backup_interval; ?> seconds (<?php echo round($auto_backup_interval / 60, 1); ?> minutes)
            </p>
            
            <div style="margin-top: 15px; padding: 10px; background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 3px;">
                <strong>📁 Download Location:</strong><br>
                Files download to your browser's default download folder.<br>
                To change download location:<br>
                • <strong>Chrome:</strong> Settings → Advanced → Downloads → Change<br>
                • <strong>Firefox:</strong> Settings → General → Downloads → Save files to<br>
                • <strong>Safari:</strong> Preferences → General → File download location
            </div>
        </div>
        
        <?php if (isset($success_message)): ?>
            <div class="status success">
                <?php echo nl2br($success_message); ?>
            </div>
        <?php endif; ?>
        
        <?php if (isset($error_message)): ?>
            <div class="status error">
                <?php echo nl2br($error_message); ?>
            </div>
        <?php endif; ?>
        
        <?php if (isset($interval_success)): ?>
            <div class="status success">
                <?php echo nl2br($interval_success); ?>
            </div>
        <?php endif; ?>
        
        <?php if (isset($interval_error)): ?>
            <div class="status error">
                <?php echo nl2br($interval_error); ?>
            </div>
        <?php endif; ?>
        
        <div class="status info">
            <strong>🕒 Automatic Backup Status:</strong><br>
            <?php echo nl2br($auto_backup_status); ?>
            <div id="backup-countdown" style="margin-top: 10px; font-weight: bold; color: #007bff;">
                Calculating countdown...
            </div>
        </div>
        
        <div style="margin-top: 30px;">
            <h2>Available Backups</h2>
            <?php if (empty($existing_backups)): ?>
                <p>No backups found.</p>
            <?php else: ?>
                <?php foreach ($existing_backups as $backup): ?>
                    <div class="backup-item <?php echo $backup['is_auto'] ? 'auto' : ''; ?>">
                        <h3>
                            <?php echo htmlspecialchars($backup['filename']); ?>
                            <?php if ($backup['is_auto']): ?>
                                <span class="auto-badge">AUTO</span>
                            <?php endif; ?>
                        </h3>
                        <p><strong>Size:</strong> <?php echo round($backup['size'] / 1024, 1); ?> KB</p>
                        <p><strong>Created:</strong> <?php echo $backup['created']; ?></p>
                        <a href="<?php echo $backup['download_url']; ?>" download="<?php echo $backup['filename']; ?>" class="btn">Download</a>
                    </div>
                <?php endforeach; ?>
            <?php endif; ?>
        </div>
    </div>

    <?php if (isset($auto_backup_ready) && $auto_backup_ready): ?>
    <script>
        // Auto-download the automatic backup
        setTimeout(function() {
            const link = document.createElement('a');
            link.href = '/backups/<?php echo $auto_backup_filename; ?>';
            link.download = '<?php echo $auto_backup_filename; ?>';
            link.style.display = 'none';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            
            // Show notification
            alert('🔄 Automatic backup created and downloaded!\n\nFile: <?php echo $auto_backup_filename; ?>\nSize: <?php echo round($auto_backup_size / 1024, 1); ?> KB');
        }, 1000);
    </script>
    <?php endif; ?>
    
    <script>
        // Automatic backup timer functionality
        let backupInterval = <?php echo $auto_backup_interval; ?> * 1000; // Convert to milliseconds
        let lastBackupTime = <?php echo file_exists($last_auto_backup_file) ? strtotime(file_get_contents($last_auto_backup_file)) : 0; ?> * 1000;
        let nextBackupTime = lastBackupTime + backupInterval;
        let timerRunning = false;
        let autoDownloadEnabled = <?php echo $auto_download_enabled ? 'true' : 'false'; ?>;
        
        function startBackupTimer() {
            if (timerRunning) return;
            timerRunning = true;
            
            function checkAndTriggerBackup() {
                const now = Date.now();
                
                if (now >= nextBackupTime) {
                    // Trigger automatic backup via AJAX
                    console.log('🔄 Triggering automatic backup...');
                    
                    fetch(window.location.href, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/x-www-form-urlencoded',
                        },
                        body: 'trigger_auto_backup=1&ajax=1'
                    })
                    .then(response => response.json())
                    .then(data => {
                        if (data.success) {
                            // Update next backup time
                            nextBackupTime = Date.now() + backupInterval;
                            
                            // Download file only if auto-download is enabled
                            if (autoDownloadEnabled) {
                                const link = document.createElement('a');
                                link.href = '/backups/' + data.filename;
                                link.download = data.filename;
                                link.style.display = 'none';
                                document.body.appendChild(link);
                                link.click();
                                document.body.removeChild(link);
                            }
                            
                            // Refresh the page to update the backup list
                            setTimeout(() => {
                                window.location.reload();
                            }, 1000);
                        }
                    })
                    .catch(error => {
                        console.error('Backup error:', error);
                    });
                } else {
                    // Update countdown display
                    const timeLeft = Math.max(0, nextBackupTime - now);
                    const minutesLeft = Math.floor(timeLeft / 60000);
                    const secondsLeft = Math.floor((timeLeft % 60000) / 1000);
                    
                    const countdownElement = document.getElementById('backup-countdown');
                    if (countdownElement) {
                        countdownElement.textContent = `Next backup in: ${minutesLeft}m ${secondsLeft}s`;
                    }
                    
                    // Check again in 1 second
                    setTimeout(checkAndTriggerBackup, 1000);
                }
            }
            
            checkAndTriggerBackup();
        }
        
        // Start the timer when page loads
        document.addEventListener('DOMContentLoaded', function() {
            startBackupTimer();
        });
        
        // Update interval when user changes it
        document.getElementById('interval_seconds')?.addEventListener('change', function() {
            backupInterval = this.value * 1000;
            nextBackupTime = Date.now() + backupInterval;
            console.log('⏰ Backup interval updated to:', this.value, 'seconds');
        });
    </script>
</body>
</html> 