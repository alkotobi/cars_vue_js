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
// Two files got this wrong in different ways before this existed:
// upload.php and backup.php each connected to $db_config['dbname']
// directly, so on a server whose db_code.json points at a different tenant they
// looked for `users` in a database that does not have it and rejected every valid
// token.
//
// The same registry row also names the tenant's upload folder (files_dir), which
// app_db_files_dir() below resolves for the same reason: a stored path is only
// meaningful next to the folder it was written into, and that folder is per
// tenant.
//
// ## One shared api/ directory
//
// Every tenant used to get its own COPY of api/, and that copy is what identified
// the tenant: the folder the code sat in decided which database it talked to. One
// shared /var/www/api/ serving every tenant breaks that - the same file would be
// every tenant's code - so the tenant now comes from the REQUEST instead:
//
//   /var/www/api/api.php                  -> the primary app        (mount '')
//   /var/www/api/../acme_cars/api/api.php  -> SCRIPT_NAME '/acme_cars/api/api.php'
//                                        -> mount 'acme_cars'
//
// With a shared api/ the mount is the tenant's folder name next to it, and the
// layout is:
//
//   /var/www/
//     api/                     one shared copy (this directory)
//     acme_cars/               the tenant's build + db_code.json
//     acme_cars_files/         the tenant's uploads
//
// app_request_mount() reads the mount, app_dir() turns it into the folder that
// holds db_code.json, and everything downstream is unchanged. The per-tenant-copy
// layout still works, because app_dir() also accepts the case where this very
// directory sits inside the tenant's own folder.

require_once __DIR__ . '/../config.php';

// The registry credentials are needed too, and they have to be loaded HERE, at
// file scope. Loading them inside a function puts $db_manager_config in that
// function's scope, where `global` cannot see it - and if some other file has
// already loaded the file, require_once does nothing at all and the variable is
// simply undefined.
if (is_file(__DIR__ . '/../db_manager_config.php')) {
    require_once __DIR__ . '/../db_manager_config.php';
}

/**
 * The tenant folder name this request is for, or '' for the primary app.
 *
 * Derived from SCRIPT_NAME, which nginx sets to the URL the request arrived on:
 * /acme_cars/api/api.php -> 'acme_cars'. That is the only per-tenant signal
 * available when every tenant shares one api/ directory, and it cannot be
 * spoofed: SCRIPT_NAME is set by nginx from the matched location, not from a
 * header or a query parameter.
 *
 * '' when there is no SCRIPT_NAME at all (CLI, cron) or when the request is for
 * /api/*.php directly, which is the primary app's own mount.
 *
 * The charset is the one deploy/render-nginx.sh accepts for a folder name, so a
 * mount that arrives here is always a name that could have been created in the
 * first place. Anything else is treated as no mount rather than passed on to a
 * filesystem path.
 */
function app_request_mount(): string
{
    static $mount = false;

    if ($mount !== false) {
        return $mount;
    }

    $scriptName = (string) ($_SERVER['SCRIPT_NAME'] ?? '');
    $mount = '';

    if (preg_match('#^/([A-Za-z0-9_][A-Za-z0-9._-]*)/api/#', $scriptName, $m) === 1) {
        $mount = $m[1];
    }

    return $mount;
}

/**
 * The app folder that identifies this deployment - the folder holding db_code.json.
 *
 * Three layouts, all of which exist or have existed on real servers:
 *
 *   shared api/        dirname(__DIR__) is the webroot, the tenant folder is a
 *                      sibling of it        -> webroot/<mount>
 *   per-tenant copy    this file already sits inside the tenant's own folder
 *                                                    -> dirname(__DIR__)
 *   primary app        no mount (mount === '')        -> dirname(__DIR__)
 *
 * The per-tenant copy is detected by the folder's own name matching the mount,
 * which is what keeps an existing deployment working without being redeployed.
 */
function app_dir(): string
{
    static $dir = false;

    if ($dir !== false) {
        return $dir;
    }

    // This file lives in api/lib, so the folder that CONTAINS api/ - the app folder
    // for a single app, the webroot for a shared api/ - is two levels up. One level
    // up is api/ itself, which is never the folder holding db_code.json.
    $root = dirname(__DIR__, 2);
    $mount = app_request_mount();

    // '' means the primary app: api/ sits directly in its own folder. So does a
    // per-tenant copy, which is the case where the folder's own name is the mount.
    if ($mount === '' || basename($root) === $mount) {
        return $dir = $root;
    }

    // Shared api/: the tenant's folder is a sibling of api/, named after the mount.
    return $dir = $root . '/' . $mount;
}

/**
 * The app folder's db_code.json value, or null if it is missing or malformed.
 */
function app_db_code(): ?string
{
    static $code = false;

    if ($code !== false) {
        return $code;
    }

    $file = app_dir() . '/db_code.json';
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
 * A PDO handle to the registry (merhab_databases), shared by the resolvers below.
 *
 * One connection, because app_db_name() and app_db_files_dir() are called from
 * the same request - often the same statement - and each would otherwise open
 * its own.
 */
function app_registry_pdo(): PDO
{
    static $pdo = null;

    if ($pdo !== null) {
        return $pdo;
    }

    // Resolve through the registry rather than trusting the code as a name.
    global $db_manager_config;

    if (!isset($db_manager_config) || !is_array($db_manager_config)) {
        throw new RuntimeException('db_manager_config.php did not define $db_manager_config');
    }

    $pdo = new PDO(
        "mysql:host={$db_manager_config['host']};dbname={$db_manager_config['dbname']}",
        $db_manager_config['user'],
        $db_manager_config['pass'],
        [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
    );

    return $pdo;
}

/**
 * This deployment's registry row, or null when the code does not resolve.
 *
 * The values come back as they were RECORDED, not normalised, because one piece of
 * information is carried by the shape and nothing else: a files_dir with a leading
 * slash means "the folder is a sibling of the app folder" (every tenant), while no
 * slash means "the folder is inside the app folder" (a single-app install). That
 * single bit is what tells app_deployment_root() which of the two layouts it is
 * looking at, and trimming it away here would make the two indistinguishable.
 * Callers that build a path use app_db_files_dir(), which normalises.
 *
 * @return array{db_name:string,files_dir:string,js_dir:string}|null
 */
function app_db_row(): ?array
{
    static $row = false;

    if ($row !== false) {
        return $row;
    }

    $code = app_db_code();
    if ($code === null) {
        return $row = null;
    }

    try {
        $stmt = app_registry_pdo()->prepare('SELECT db_name, files_dir, js_dir FROM dbs WHERE db_code = ?');
        $stmt->execute([$code]);
        $found = $stmt->fetch(PDO::FETCH_ASSOC);
    } catch (Exception $e) {
        error_log('app_db_row: ' . $e->getMessage());
        return $row = null;
    }

    if (!$found || empty($found['db_name'])) {
        return $row = null;
    }

    $slashToDirSep = static function ($value): string {
        return str_replace('\\', '/', trim((string) $value));
    };

    return $row = [
        'db_name' => (string) $found['db_name'],
        'files_dir' => $slashToDirSep($found['files_dir'] ?? ''),
        'js_dir' => $slashToDirSep($found['js_dir'] ?? ''),
    ];
}

/**
 * The tenant database name for this deployment.
 *
 * @return string|null null when db_code.json is absent or does not resolve. Callers
 *         must refuse rather than substitute a configured database name: config.php
 *         holds credentials and no database in a multi-tenant setup, and when it did
 *         hold one that was a fixed name, so an unresolved request opened whichever
 *         tenant that name happened to be.
 */
function app_db_name(): ?string
{
    static $resolved = false;

    if ($resolved !== false) {
        return $resolved;
    }

    $row = app_db_row();

    return $resolved = ($row !== null ? $row['db_name'] : null);
}

/**
 * The upload folder for this deployment, relative to the project root.
 *
 * Needed wherever a stored path is turned back into a file on disk: replacing a
 * car file and deleting one both unlink the old file, and both used to look the
 * folder up with `SELECT js_dir FROM dbs LIMIT 1` against the TENANT connection.
 * That table is created by setup.sql but never written to - it is empty on every
 * install, merhab_cars included - so both branches silently fell back to
 * 'mig_files' and no deployment other than that one ever deleted anything. The
 * upload folder is per tenant, so it has to come from the registry, the same
 * place the browser reads it from (src/composables/useApi.js).
 *
 * @return string|null null when it cannot be resolved; callers fall back to
 *         'mig_files', which is what they did before.
 */
function app_db_files_dir(): ?string
{
    static $resolved = false;

    if ($resolved !== false) {
        return $resolved;
    }

    $row = app_db_row();

    if ($row === null) {
        return $resolved = null;
    }

    // The registry form accepts a leading/trailing slash and a Windows separator,
    // and this value is concatenated onto a filesystem path, so it is reduced to a
    // plain relative name here. An empty value is legitimate - the deployment
    // uploads into the project root - and is returned as-is.
    return $resolved = trim($row['files_dir'], '/');
}

/**
 * The folder that `files_dir` is recorded against.
 *
 * The registry stores a tenant's files_dir and js_dir as SIBLINGS, which is why
 * they carry a leading slash: a tenant is deployed as <webroot>/acme_cars (the app
 * folder, what js_dir names) next to <webroot>/acme_cars_files (its uploads). So
 * for a tenant the deployment root is the app folder's PARENT, while a single-app
 * install - no slash on files_dir, uploads inside the app folder - is its own root.
 *
 * That recorded shape is what distinguishes the layouts, because the two are
 * otherwise indistinguishable on disk:
 *
 *   /var/www/world-automobile.com/cars/api/api.php  + files_dir 'files'  -> app folder
 *   /var/www/world-automobile.com/cars/api/api.php  + files_dir '/c_files'
 *                                                          -> app folder's parent
 *   /var/www/api/api.php                            + files_dir 'files'  -> app folder
 *   /var/www/api/api.php                            + files_dir '/c_files'
 *                                                          -> app folder's parent
 *
 * The first two rows are the same layout with the same mount ('cars' from
 * SCRIPT_NAME), so the filesystem cannot tell them apart - only the registry can,
 * and app_dir() has already resolved WHICH tenant this request is by the time this
 * runs.
 */
function app_deployment_root(): string
{
    static $root = false;

    if ($root !== false) {
        return $root;
    }

    $appDir = app_dir();

    // A leading slash on the RECORDED value is the sibling marker. app_db_files_dir()
    // normalises it away for path building, so the row is read directly.
    $row = app_db_row();
    $recorded = $row === null ? '' : trim($row['files_dir']);

    return $root = ($recorded !== '' && str_starts_with($recorded, '/'))
        ? dirname($appDir)
        : $appDir;
}

/**
 * A PDO handle to the database that holds `users`.
 *
 * @throws RuntimeException when the database cannot be reached
 */
function app_db_pdo(): PDO
{
    // Without this `global`, $db_config is undefined inside the function: the
    // host, user and pass all arrive as null and PDO fails with "Access denied
    // for user ''@'localhost'". Every endpoint that authenticates through here
    // (upload.php, backup.php) was returning that.
    global $db_config;

    // The config.php fallback that used to follow has gone with it, and this was
    // the last caller still reaching for one.
    //
    // It is not a theoretical leak, it is the same leak api.php and backup.php
    // were closed for, reached through the one path they do not cover: auth. Set
    // db_name in config.php and a request to the repository root's upload.php
    // authenticated against that database's users and api_tokens - verified by
    // pointing the root at merhab_cars and getting "not_authenticated" (it
    // connected and looked for a token) where a name that does not exist gives
    // "Unknown database". The root is not a tenant and has no business
    // authenticating anyone.
    //
    // app_db_name()'s own docblock already required this: callers must refuse
    // rather than substitute a configured name.
    $dbname = app_db_name() ?? '';
    if ($dbname === '') {
        throw new RuntimeException(
            'Cannot authenticate: this request did not resolve to a tenant. '
            . 'Set db_name in config.php, or reach the app through its tenant folder.'
        );
    }

    $pdo = new PDO(
        "mysql:host={$db_config['host']};dbname={$dbname}",
        $db_config['user'],
        $db_config['pass']
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    return $pdo;
}

/**
 * Authenticate the caller as any user, using the deployment's own `users` table.
 *
 * The token is a STRING, from the X-Api-Token header or the query string. It was
 * typed `array`, and the caller - upload.php's upload_require_user() - passes a
 * string, so every POST upload died on
 * `TypeError: require_app_user(): Argument #1 ($token) must be of type array,
 * string given`: a 500 with an empty body, because upload.php turns errors into
 * JSON and this one happened before it could.
 *
 * There was an admin-only twin of this, require_app_admin(). It went with
 * invitations.php, its only caller; require_api_admin() itself is untouched and is
 * still reachable for anything that needs an admin check.
 *
 * @return array{id:int,username:string,role_id:int}
 */
function require_app_user($token): array
{
    require_once __DIR__ . '/auth.php';
    return require_api_user(app_db_pdo(), ['token' => is_string($token) ? $token : '']);
}
