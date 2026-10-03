<?php
// Resolve the database this deployment actually serves.
//
// config.php's db_name is only a default. A single build serves many clients, and
// which tenant database a server talks to is decided at deploy time by the app
// folder's db_code.json (see deploy/deploy.sh), which maps through the
// merhab_databases registry's `dbs` table. api.php works this out in
// resolveDbNameFromCode(); every standalone endpoint that needs to read the
// `users` table - and therefore to authenticate anyone - has to agree with it.
//
// Three files got this wrong in different ways before this existed:
// upload.php, backup.php and invitations.php each connected to $db_config['dbname']
// directly, so on a server whose db_code.json points at a different tenant they
// looked for `users` in a database that does not have it and rejected every valid
// token.

require_once __DIR__ . '/../config.php';

/**
 * The app folder's db_code.json value, or null if it is missing or malformed.
 */
function app_db_code(): ?string
{
    static $code = false;

    if ($code !== false) {
        return $code;
    }

    $file = dirname(__DIR__, 2) . '/db_code.json';
    if (!is_file($file)) {
        return $code = null;
    }

    $data = json_decode((string) @file_get_contents($file), true);
    $value = is_array($data) ? trim((string) ($data['db_code'] ?? '')) : '';

    // Same whitelist api.php applies. db_code is used as a bind parameter, but a
    // code that is not shaped like a code is a sign something is wrong upstream.
    return $code = ($value !== '' && preg_match('/^db_[0-9a-f]+$/', $value)) ? $value : null;
}

/**
 * The tenant database name for this deployment.
 *
 * @return string|null null when db_code.json is absent or does not resolve, in
 *         which case callers should fall back to config.php's db_name.
 */
function app_db_name(): ?string
{
    static $resolved = false;

    if ($resolved !== false) {
        return $resolved;
    }

    $code = app_db_code();
    if ($code === null) {
        return $resolved = null;
    }

    // Resolve through the registry rather than trusting the code as a name.
    $registry = __DIR__ . '/../db_manager_config.php';
    if (!is_file($registry)) {
        return $resolved = null;
    }

    require_once $registry;
    if (!isset($db_manager_config)) {
        return $resolved = null;
    }

    try {
        $pdo = new PDO(
            "mysql:host={$db_manager_config['host']};dbname={$db_manager_config['dbname']}",
            $db_manager_config['user'],
            $db_manager_config['pass'],
            [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
        );
        $stmt = $pdo->prepare('SELECT db_name FROM dbs WHERE db_code = ?');
        $stmt->execute([$code]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);
    } catch (Exception $e) {
        error_log('app_db_name: ' . $e->getMessage());
        return $resolved = null;
    }

    return $resolved = ($row && !empty($row['db_name']) ? (string) $row['db_name'] : null);
}

/**
 * A PDO handle to the database that holds `users`.
 *
 * @throws RuntimeException when the database cannot be reached
 */
function app_db_pdo(): PDO
{
    $dbname = app_db_name() ?? $db_config['dbname'];

    $pdo = new PDO(
        "mysql:host={$db_config['host']};dbname={$dbname}",
        $db_config['user'],
        $db_config['pass']
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    return $pdo;
}

/**
 * Authenticate the caller as an admin, using the deployment's own `users` table.
 *
 * Ends the request when the token is missing, unknown, or not an admin.
 *
 * @return array{id:int,username:string,role_id:int}
 */
function require_app_admin(array $token): array
{
    require_once __DIR__ . '/auth.php';
    return require_api_admin(app_db_pdo(), ['token' => is_string($token) ? $token : '']);
}

/**
 * Authenticate the caller as any user, using the deployment's own `users` table.
 *
 * @return array{id:int,username:string,role_id:int}
 */
function require_app_user(array $token): array
{
    require_once __DIR__ . '/auth.php';
    return require_api_user(app_db_pdo(), ['token' => is_string($token) ? $token : '']);
}
