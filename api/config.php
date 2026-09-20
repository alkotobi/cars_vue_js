<?php
// Database configuration

// ============================================================
// ACTIVE: world-automobile.com server (163.245.214.125)
// ============================================================
$db_host = '127.0.0.1';
$db_user = 'merhab_root';
$db_pass = '@Salima61';
$db_name = 'merhab_cars';

// ------------------------------------------------------------
// Old production server (kept for reference/reuse)
// $db_host = '173.214.163.18';
// $db_user = 'merhab_root';
// $db_pass = '@Salima61';
// $db_name = 'merhab_cars';
//
// Local development (Mac/Windows)
// $db_host = 'localhost';
// $db_user = 'root';
// $db_pass = 'nooo';
// $db_name = 'merhab_cars';
// ------------------------------------------------------------

// Create a config array that can be used by other files
$db_config = [
    'host' => $db_host,
    'user' => $db_user,
    'pass' => $db_pass,
    'dbname' => $db_name
];