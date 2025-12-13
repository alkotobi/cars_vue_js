<?php
// Database configuration for merhab_databases
// IMPORTANT: This file should be added to .gitignore to prevent committing credentials

// Load from environment variables if available, otherwise use defaults
$db_host = $_ENV['DB_MANAGER_HOST'] ?? 'localhost';
$db_user = $_ENV['DB_MANAGER_USER'] ?? 'root';
$db_pass = $_ENV['DB_MANAGER_PASS'] ?? 'nooo';
$db_name = $_ENV['DB_MANAGER_NAME'] ?? 'merhab_databases';

// Create a config array that can be used by other files
$db_manager_config = [
    'host' => $db_host,
    'user' => $db_user,
    'pass' => $db_pass,
    'dbname' => $db_name
];

