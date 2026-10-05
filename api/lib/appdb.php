<?php
// Resolve the database this deployment actually serves.
//
// Which database a request talks to is decided by the TENANT in the request, and
// the tenant is its own name. There is no binding file and no registry lookup: the
// folder is named after the database, so the name in the URL is the database name,
// and api.php and every standalone endpoint that authenticates a caller (upload.php,
// backup.php) resolve it the same way - through app_db_name() below.
//
// Three things got this wrong before, each in its own file: db_code.json was the
// original binding, then the folder the code sat in, then the merhab_databases
// registry row. All three could disagree with each other and with the folder on
// disk, and each disagreement opened a tenant that was not being asked for.
//
// ## The layout
//
//   /var/www/
//     api/                     one shared copy (this directory)
//     dist/                    one shared build
//     acme_cars/               the tenant's uploads, and nothing else
//       files/
//
// ## Two ways to name the tenant
//
//   path        /acme_cars/api/api.php          -> SCRIPT_NAME
//   subdomain   acme_cars.example.com/api/api.php -> Host, against base_domain
//
// Both are the same name, used unchanged as the database name. The webroot itself
// (/api/api.php, example.com/api/api.php) resolves to nothing, and callers must
// refuse rather than substitute a configured database - config.php holds connection
// credentials and no database, and when it did hold one it was a fixed name, so an
// unresolved request opened whichever tenant that name happened to be.

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
 * Folder names at the webroot that are not tenants.
 *
 * Every tenant is a folder at the webroot, so the two namespaces share a level and
 * 'api' and 'dist' would each be a legal tenant name. They are not: they are the
 * shared code and the shared build, and a tenant called api would be unreachable
 * (its location block would be shadowed) and would collide with the folder the
 * whole installation runs from.
 */
const APP_RESERVED_NAMES = ['api', 'dist'];

/**
 * What a tenant name may be, and nothing else.
 *
 * The name is used unchanged as a database name, an nginx location, a URL segment
 * and a filesystem path, so the charset has to be the intersection of all four
 * rather than the most permissive one. Lowercase only, because the folder, the URL
 * and the database would then stop being the same string on a case-insensitive
 * filesystem: two tenants differing only in case would share one folder.
 *
 * No dots and no dashes, which is the part that used to be relaxed. They were
 * normalised to underscores so the name could be a MySQL identifier, and that
 * normalisation is exactly what breaks the arrangement - a folder named
 * acme-motors holding a database named acme_motors is two names, so a rename, a
 * backup or a support question has to know the rule to connect them. Refusing the
 * name is cheaper than encoding it.
 *
 * @return string|null the name unchanged, or null when it is not a tenant name
 */
function app_valid_tenant(string $name): ?string
{
    $name = trim($name);

    // 64 characters is MySQL's limit on a database name; longer than that is a
    // database this app could not connect to even though it passed every other
    // check, so it is refused here rather than at the point of failure.
    if (preg_match('/^[a-z][a-z0-9_]{0,63}$/', $name) !== 1) {
        return null;
    }

    // 'api' and 'dist' are the shared folders, and they match the tenant pattern
    // exactly. Without this, the root's own SCRIPT_NAME (/api/api.php) parses as a
    // request for a tenant called api - so the webroot stops resolving to nothing
    // and starts resolving to a database named api.
    if (in_array($name, APP_RESERVED_NAMES, true)) {
        return null;
    }

    return $name;
}

/**
 * The base domain tenants are served on, or '' when it is not configured.
 *
 * Deliberately not guessed. The request's Host header is the only other signal, and
 * it is chosen by whoever sent the request - so "the first label is the tenant"
 * means a request to example.com resolves to a tenant called example, and a
 * request to anything.example.net with the same shape resolves to that domain's
 * labels. Comparing against a configured value turns "is this host mine?" into a
 * question with a known answer, and every host that does not match is simply not a
 * tenant.
 *
 * Empty is a real answer, not a missing one: it means subdomain routing is off and
 * only the path form works.
 */
function app_base_domain(): string
{
    static $domain = false;

    if ($domain !== false) {
        return $domain;
    }

    $fromEnv = getenv('CARS_BASE_DOMAIN');
    $configured = (is_string($fromEnv) && $fromEnv !== '') ? $fromEnv : '';

    if ($configured === '') {
        global $db_config;
        $configured = is_array($db_config) && isset($db_config['base_domain']) && is_string($db_config['base_domain'])
            ? $db_config['base_domain']
            : '';
    }

    $configured = strtolower(trim($configured, " \t\n\r\0\x0B."));
    // A leading '*.' wildcard is how a certificate is usually issued for this; the
    // name being compared against never carries it.
    if (str_starts_with($configured, '*.')) {
        $configured = substr($configured, 2);
    }

    return $domain = $configured;
}

/**
 * The tenant this request is for, or '' when it is not a tenant request.
 *
 * Three sources, in the order that survives a wrong answer:
 *
 *   1. CARTS_TENANT, set by the server's own location block. It is the only source
 *      that is not chosen by the caller, so it is checked first and it wins: a
 *      generated location block knows the tenant it was generated for, and no Host
 *      header or URL can overrule it.
 *   2. the Host header, as <tenant>.<base_domain>, and only when base_domain is
 *      configured.
 *   3. SCRIPT_NAME, as /<tenant>/api/...
 *
 * SCRIPT_NAME rather than REQUEST_URI, because it is what the web server matched
 * the location block on, so it cannot be rewritten to look like a tenant the server
 * did not route. REQUEST_URI is the raw request line and is not used for identity.
 */
function app_request_tenant(): string
{
    static $tenant = false;

    if ($tenant !== false) {
        return $tenant;
    }

    $tenant = '';

    // Present-and-set is what makes this authoritative; absent or empty means the
    // server did not name a tenant, so the other sources get their turn.
    if (isset($_SERVER['CARDS_TENANT']) && is_string($_SERVER['CARDS_TENANT'])) {
        $fromServer = trim($_SERVER['CARDS_TENANT']);
        if ($fromServer !== '') {
            $valid = app_valid_tenant($fromServer);
            if ($valid !== null) {
                return $tenant = $valid;
            }
            // A set-but-invalid value is not a reason to try something else: the
            // server said which tenant this is and it is not one, so no other source
            // may override that with a guess.
            return $tenant = '';
        }
    }
    // Fallback for dev proxy: allow X-Cards-Tenant header
    if (isset($_SERVER['HTTP_X_CARDS_TENANT']) && is_string($_SERVER['HTTP_X_CARDS_TENANT'])) {
        $fromHeader = trim($_SERVER['HTTP_X_CARDS_TENANT']);
        if ($fromHeader !== '') {
            $valid = app_valid_tenant($fromHeader);
            if ($valid !== null) {
                return $tenant = $valid;
            }
            return $tenant = '';
        }
    }

    $baseDomain = app_base_domain();
    if ($baseDomain !== '') {
        $host = strtolower(trim((string) ($_SERVER['HTTP_HOST'] ?? '')));
        $host = (string) preg_replace('/:\d+$/', '', $host);
        if ($host !== '' && substr($host, -strlen('.' . $baseDomain)) === '.' . $baseDomain) {
            $label = substr($host, 0, -strlen('.' . $baseDomain));
            $valid = app_valid_tenant($label);
            if ($valid !== null) {
                return $tenant = $valid;
            }
            // The host IS ours but the label is not a tenant name, so this is a
            // request to the bare domain with a stray prefix: not a tenant, and
            // certainly not a reason to read a path segment instead.
            return $tenant = '';
        }
        if ($host !== '' && $host !== $baseDomain && substr($host, -strlen('.' . $baseDomain)) !== '.' . $baseDomain) {
            // Some other domain entirely. Not an error to report - it just is not
            // this installation, and it must not reach a tenant.
            return $tenant = '';
        }
    }

    $scriptName = (string) ($_SERVER['SCRIPT_NAME'] ?? '');
    if ($scriptName !== '' && preg_match('#^/([a-z][a-z0-9_]{0,63})/#', $scriptName, $m) === 1) {
        $valid = app_valid_tenant($m[1]);
        if ($valid !== null) {
            $tenant = $valid;
        }
    }

    return $tenant;
}

/**
 * The name this deployment is reached under, for upload.php's documentation and for
 * anything still calling it a mount. Same value as app_request_tenant(); the term
 * only survived the cutover.
 */
function app_request_mount(): string
{
    return app_request_tenant();
}

/**
 * The webroot: the folder that contains api/, dist/ and every tenant folder.
 *
 * This file is in api/lib, so api/ is one level up and the folder holding it is
 * two. That folder is the webroot in this layout - the tenant folder is a child of
 * it, named after the tenant - and it is also the root of the repository in local
 * development, which is the same arrangement.
 */
function app_webroot(): string
{
    return dirname(__DIR__, 2);
}

/**
 * The folder this request's tenant owns, or the webroot when there is no tenant.
 *
 * There is no per-tenant copy of the code to detect any more, so this is always
 * webroot/<tenant>. The basename check is what keeps a tenant folder that contains
 * api/ from pointing at itself, which is how the old per-tenant layout keeps
 * working during the cutover instead of resolving to a folder that does not exist.
 */
function app_dir(): string
{
    static $dir = false;

    if ($dir !== false) {
        return $dir;
    }

    $root = app_webroot();
    $tenant = app_request_tenant();

    if ($tenant === '' || basename($root) === $tenant) {
        return $dir = $root;
    }

    return $dir = $root . '/' . $tenant;
}

/**
 * The tenant database name for this request.
 *
 * The tenant name, unchanged. That is the whole point of the layout: the folder is
 * named after the database, so there is nothing to look up and nothing that can
 * disagree with the folder on disk.
 *
 * @return string|null null when the request is not a tenant request. Callers must
 *         refuse rather than substitute a configured database name - config.php
 *         holds credentials and no database, and when it did hold one that was a
 *         fixed name, so an unresolved request opened whichever tenant that name
 *         happened to be.
 */
function app_db_name(): ?string
{
    static $resolved = false;

    if ($resolved !== false) {
        return $resolved;
    }

    $tenant = app_request_tenant();

    return $resolved = ($tenant === '' ? null : $tenant);
}

/**
 * The tenant's upload folder, relative to the tenant's own folder.
 *
 * Needed wherever a stored path is turned back into a file on disk: replacing a car
 * file and deleting one both unlink the old file, and both used to look the folder
 * up with `SELECT js_dir FROM dbs LIMIT 1` against the TENANT connection. That table
 * is created by setup.sql but never written to, so both branches silently fell back
 * to a fixed name and no deployment ever deleted anything.
 *
 * Always 'files', so there is nothing per-tenant to resolve and nothing to get out
 * of step with the folder. The browser sends the same value (src/composables/useApi.js).
 *
 * @return string|null null when the request is not a tenant request.
 */
function app_db_files_dir(): ?string
{
    static $resolved = false;

    if ($resolved !== false) {
        return $resolved;
    }

    if (app_request_tenant() === '') {
        return $resolved = null;
    }

    return $resolved = 'files';
}

/**
 * The folder a stored path is resolved against: the tenant's own folder.
 *
 * Used to be two different answers - the app folder for a single-app install, its
 * parent for a tenant whose uploads were a sibling - which meant the correct root
 * had to be inferred from the shape of a recorded value. There is one layout now, so
 * this is the folder and nothing else, and the guess is gone with it.
 */
function app_deployment_root(): string
{
    static $root = false;

    if ($root !== false) {
        return $root;
    }

    return $root = app_dir();
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
            'Cannot authenticate: this request did not resolve to a tenant. Reach the app through its tenant path or tenant subdomain.'
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
