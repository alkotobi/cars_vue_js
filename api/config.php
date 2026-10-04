<?php
// Database configuration loader.
//
// This file is tracked in git and must NOT contain credentials.
// Resolution order:
//   1. api/config.local.php  (per-machine overrides, git-ignored)
//   2. Environment variables: DB_HOST / DB_USER / DB_PASS
//
// db_name is deliberately absent from both. This file resolves CONNECTION settings;
// which database to connect to is decided per request from db_code.json and the
// merhab_databases registry (see lib/appdb.php). See the note at its definition
// below.
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

// No default, and no tenant name. Which database this process serves is decided per
// request by db_code.json -> the merhab_databases registry (see lib/appdb.php), never
// by a value baked into configuration. A default here was a silent way for a request
// that had resolved nothing to land in one specific tenant, and the name it named
// (merhab_cars) was a live database rather than a placeholder. Empty means "no
// database configured"; callers that must have one ask for it by name.
// Empty is allowed and means "no database configured". It is not rejected here:
// api/install.php and the provisioning CLI both load this file while working out
// which database to use, and a throw at include time would take them down before they
// could ask. Every consumer that actually needs a name checks for it and refuses
// with a message naming the setting - api.php's getDbConfig() is the one on the
// request path.
$db_name = $resolveConfig('db_name', 'DB_NAME', '');


// Create a config array that can be used by other files
$db_config = [
    'host' => $db_host,
    'user' => $db_user,
    'pass' => $db_pass,
    'dbname' => $db_name
];
