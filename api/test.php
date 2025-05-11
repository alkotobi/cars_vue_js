<?php
header('Access-Control-Allow-Origin: *');
header('Content-Type: text/html');
?>
<!DOCTYPE html>
<html>
<head>
    <title>PHP Test</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        .test-item { margin: 10px 0; padding: 10px; border: 1px solid #ddd; }
        .success { color: green; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>PHP Functionality Test</h1>
    
    <div class="test-item">
        <h3>1. Basic PHP Execution:</h3>
        <?php 
        echo "<p class='success'>✓ PHP is executing correctly</p>";
        ?>
    </div>

    <div class="test-item">
        <h3>2. PHP Version:</h3>
        <?php 
        echo "<p>Running PHP version: " . phpversion() . "</p>";
        ?>
    </div>

    <div class="test-item">
        <h3>3. MySQL PDO Test:</h3>
        <?php
        try {
            $pdo = new PDO("mysql:host=localhost;dbname=merhab_cars", "root", "nooo");
            echo "<p class='success'>✓ MySQL connection successful</p>";
        } catch(PDOException $e) {
            echo "<p class='error'>✗ MySQL connection failed: " . $e->getMessage() . "</p>";
        }
        ?>
    </div>

    <div class="test-item">
        <h3>4. Required Extensions:</h3>
        <?php
        $required_extensions = ['pdo_mysql', 'json', 'mysqli'];
        foreach($required_extensions as $ext) {
            if(extension_loaded($ext)) {
                echo "<p class='success'>✓ {$ext} extension is loaded</p>";
            } else {
                echo "<p class='error'>✗ {$ext} extension is missing</p>";
            }
        }
        ?>
    </div>
</body>
</html>