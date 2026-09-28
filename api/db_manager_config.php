<?php
// Database configuration loader for the merhab_databases (db-manager) database.
//
// This file is tracked in git and must NOT contain credentials.
// Resolution order:
//   1. api/db_manager_config.local.php  (per-machine overrides, git-ignored)
//   2. Environment variables: DB_MANAGER_HOST / DB_MANAGER_USER /
//      DB_MANAGER_PASS / DB_MANAGER_NAME
//
// To set up a machine: copy api/db_manager_config.example.php to
// api/db_manager_config.local.php and fill in the values.

$configLocalFile = __DIR__ . '/db_manager_config.local.php';
$configLocal = [];

if (is_file($configLocalFile)) {
    $loaded = require $configLocalFile;
    if (is_array($loaded)) {
        $configLocal = $loaded;
    }
}

$resolveConfig = static function (string $key, string $env, ?string $default = null) use ($configLocal) {
    if (isset($configLocal[$key]) && $configLocal[$key] !== '') {
        return $configLocal[$key];
    }

    // getenv() rather than $_ENV: $_ENV is only populated when the php.ini
    // variables_order includes "E", which is not the default.
    $fromEnv = getenv($env);
    if ($fromEnv !== false && $fromEnv !== '') {
        return $fromEnv;
    }

    if ($default !== null) {
        return $default;
    }

    throw new RuntimeException(
        "Missing db-manager setting '$key'. Create " . __DIR__ . "/db_manager_config.local.php "
        . "(see db_manager_config.example.php) or set the $env environment variable."
    );
};

$db_host = $resolveConfig('db_host', 'DB_MANAGER_HOST', '127.0.0.1');
$db_user = $resolveConfig('db_user', 'DB_MANAGER_USER');
$db_pass = $resolveConfig('db_pass', 'DB_MANAGER_PASS');
$db_name = $resolveConfig('db_name', 'DB_MANAGER_NAME', 'merhab_databases');

// Create a config array that can be used by other files
$db_manager_config = [
    'host' => $db_host,
    'user' => $db_user,
    'pass' => $db_pass,
    'dbname' => $db_name
];
