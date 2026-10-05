<?php
// Database configuration loader.
//
// This file is tracked in git and must NOT contain credentials.
// Resolution order:
//   1. api/config.local.php  (per-machine overrides, git-ignored)
//   2. Environment variables: DB_HOST / DB_USER / DB_PASS
//
// db_name is deliberately absent from both. This file resolves CONNECTION settings;
// which database to connect to is decided per request from tenant name (see lib/appdb.php). See the note at its definition
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

$db_name = $resolveConfig('db_name', 'DB_NAME', '');

// The bare domain tenants are served on, for <tenant>.<base_domain>. Empty means
// subdomain requests are not recognised at all and only the path form works.
//
// No default, and it is read here rather than in appdb.php because it is a setting
// like the others - and because it has to be a SETTING. Host is chosen by whoever
// sends the request, so inferring the tenant from it means example.com resolves to
// a tenant called "example". With this empty the comparison fails for every host, so
// subdomain routing is off rather than wrong.
$base_domain = $resolveConfig('base_domain', 'CARS_BASE_DOMAIN', '');

// Create a config array that can be used by other files
$db_config = [
    'host' => $db_host,
    'user' => $db_user,
    'pass' => $db_pass,
    'dbname' => $db_name,
    'base_domain' => $base_domain,
];
