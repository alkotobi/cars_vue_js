<?php
// Database configuration for merhab_databases
// IMPORTANT: This file should be added to .gitignore to prevent committing credentials

// ============================================================
// ACTIVE: world-automobile.com server (163.245.214.125)
// ============================================================
$db_host = $_ENV['DB_MANAGER_HOST'] ?? '127.0.0.1';
$db_user = $_ENV['DB_MANAGER_USER'] ?? 'merhab_root';
$db_pass = $_ENV['DB_MANAGER_PASS'] ?? '@Salima61';
$db_name = $_ENV['DB_MANAGER_NAME'] ?? 'merhab_databases';

// ------------------------------------------------------------
// Old production server (kept for reference/reuse)
// $db_host = '173.214.163.18';
// $db_user = 'merhab_root';
// $db_pass = '@Salima61';
// $db_name = 'merhab_databases';
//
// Local development (Mac/Windows)
// $db_host = 'localhost';
// $db_user = 'root';
// $db_pass = 'nooo';
// $db_name = 'merhab_databases';
// ------------------------------------------------------------

// Create a config array that can be used by other files
$db_manager_config = [
    'host' => $db_host,
    'user' => $db_user,
    'pass' => $db_pass,
    'dbname' => $db_name
];