<?php
// Optional settings: values a server may legitimately not have.
//
// config.php resolves the database credentials and throws when one is missing,
// because the app cannot run without them. An integration key is a different
// animal: a client server that has never been given an AI key is still a
// perfectly working app, so these resolve to a default instead of throwing and
// let the caller report "not configured".
//
// Resolution order, matching config.php:
//   1. api/config.local.php   per machine, git-ignored, excluded from deploy.sh
//   2. environment variable
//   3. the default passed by the caller

/**
 * @return array<string,mixed> contents of api/config.local.php, or [] if absent
 */
function app_config_local(): array
{
    static $local = null;
    if ($local !== null) {
        return $local;
    }

    $local = [];
    $file = __DIR__ . '/../config.local.php';
    if (is_file($file)) {
        $loaded = require $file;
        if (is_array($loaded)) {
            $local = $loaded;
        }
    }

    return $local;
}

/**
 * Read an optional setting as a string. An empty value anywhere in the chain
 * falls through to the next source, so a blank entry in config.local.php cannot
 * shadow a working environment variable.
 */
function app_setting(string $key, string $env, string $default = ''): string
{
    $local = app_config_local();
    if (isset($local[$key]) && $local[$key] !== '') {
        return (string) $local[$key];
    }

    $fromEnv = getenv($env);
    if ($fromEnv !== false && $fromEnv !== '') {
        return $fromEnv;
    }

    return $default;
}

function app_setting_int(string $key, string $env, int $default): int
{
    $value = app_setting($key, $env, '');
    if ($value === '' || !is_numeric($value)) {
        return $default;
    }

    return (int) $value;
}

/**
 * Read an optional on/off setting.
 *
 * Deliberately forgiving about spelling - '1', 'true', 'yes' and 'on' all mean
 * on - because these get typed by hand into config.local.php on every server.
 *
 * An unrecognised value falls back to the default instead of guessing. A typo
 * like 'flase' must not silently flip a feature on or off, and silently is the
 * one thing a switch must never do.
 */
function app_setting_bool(string $key, string $env, bool $default): bool
{
    $value = strtolower(trim(app_setting($key, $env, '')));
    if ($value === '') {
        return $default;
    }

    if (in_array($value, ['1', 'true', 'yes', 'on'], true)) {
        return true;
    }

    if (in_array($value, ['0', 'false', 'no', 'off'], true)) {
        return false;
    }

    error_log("app_setting_bool: ignoring unrecognised value for $key");

    return $default;
}
