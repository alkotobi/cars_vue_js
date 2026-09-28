<?php
// Database configuration loader.
//
// This file is tracked in git and must NOT contain credentials.
// Resolution order:
//   1. api/config.local.php  (per-machine overrides, git-ignored)
//   2. Environment variables: DB_HOST / DB_USER / DB_PASS / DB_NAME
//
// To set up a machine: copy api/config.example.php to api/config.local.php
// and fill in the values. See DEPLOYMENT.md for the production path.

$configLocalFile = __DIR__ . '/config.local.php';
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

    $fromEnv = getenv($env);
    if ($fromEnv !== false && $fromEnv !== '') {
        return $fromEnv;
    }

    if ($default !== null) {
        return $default;
    }

    throw new RuntimeException(
        "Missing database setting '$key'. Create " . __DIR__ . "/config.local.php "
        . "(see config.example.php) or set the $env environment variable."
    );
};

$db_host = $resolveConfig('db_host', 'DB_HOST', '127.0.0.1');
$db_user = $resolveConfig('db_user', 'DB_USER');
$db_pass = $resolveConfig('db_pass', 'DB_PASS');
$db_name = $resolveConfig('db_name', 'DB_NAME', 'merhab_cars');

// Create a config array that can be used by other files
$db_config = [
    'host' => $db_host,
    'user' => $db_user,
    'pass' => $db_pass,
    'dbname' => $db_name
];
