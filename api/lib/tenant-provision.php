<?php
// Provision one client of the multi-tenant fleet: its database and its folder,
// which holds its uploads. The database is named after the tenant and the folder is
// named after the database, so the tenant's name is the only thing that identifies
// it - there is no binding file to keep in step with it, and no per-tenant copy of
// the code to drift out of date.
//
// ## Why this is a library and not a script
//
// The first version of this was deploy/setup-mig-27.php, which provisioned the one
// local tenant mig_27 and could only be run from a shell on the machine holding the
// repository. Onboarding a client on the server needs the same steps with no SSH,
// so they are callable from the DB manager's Provision button. The steps are
// identical - that is the point - so they live here and both callers use them:
//
//   deploy/setup-mig-27.php    the local dev tenant (mig_27), unchanged behaviour
//   api/db_manager_api.php     action=provision_tenant, from the browser
//
// ## What it does NOT do
//
// It never copies business data. A new tenant is created empty apart from the
// reference and auth rows it needs to log in (roles, permissions, an admin
// account, lookup lists). The template database it seeds from is verified to be
// reference-only first - see TENANT_BUSINESS_TABLES - so a template that
// accidentally holds one client's cars cannot be handed to the next one.
//
// ## Everything is idempotent
//
// Every step can be re-run: the database is CREATE IF NOT EXISTS, migrations are
// replayed (already-applied errors are recognised, see TENANT_BENIGN_ERRNOS), the
// seed is skipped for any table that already has rows, and folders are mkdir -p.
// Re-running is how a tenant is repaired, and --force rebuilds one from scratch.
//
// ## Failures throw, they do not exit
//
// This file runs inside a web request as often as it runs from a CLI, so every
// error is a TenantProvisionError for the caller to turn into an exit code or a
// JSON response. Nothing here calls exit(), echoes to stdout, or touches $_SERVER.

declare(strict_types=1);

/**
 * A step failed. The message is safe to show an operator; it never contains a
 * password (credentials reach the mysql client through MYSQL_PWD).
 */
class TenantProvisionError extends RuntimeException
{
}

// ---------------------------------------------------------------------------
// Conventions
// ---------------------------------------------------------------------------

/**
 * What a client's database - and therefore its folders - may be called.
 *
 * This is deliberately stricter than MySQL's rules for a database name, because
 * the name ends up in three other places that are not SQL: an nginx location, a
 * URL prefix, and a filesystem path that PHP builds by concatenation. Every one of
 * those needs the name to be a single token, so the allowlist below is short.
 *
 * The name is used unchanged everywhere: the database, the folder under the
 * webroot, and the URL prefix. Dashes and dots were once accepted and normalised
 * to underscores, which produced a folder named `acme-motors` holding a database
 * named `acme_motors` - two names for one client, and every rename, backup and
 * support question then has to know the rule to connect them. Lowercase only,
 * because `Acme` and `acme` are two databases and one folder on a
 * case-insensitive filesystem. api_valid_tenant() enforces exactly this, and the
 * two have to agree: a name this accepts and that one rejects provisions a client
 * whose site cannot resolve its own database.
 */
const TENANT_DB_NAME_PATTERN = '/^[a-z][a-z0-9_]{0,63}$/';

/**
 * Webroot folder names that are not tenants, and so cannot be one.
 *
 * api/ is the shared code and dist/ the shared build. A client called either would
 * have its folder shadowed by the thing it is served from, and its uploads would sit
 * next to shared code. api_valid_tenant() refuses the same two names for the same
 * reason; they are listed here rather than imported because this library runs as root
 * on the server, where requiring a file inside the web root is exactly what must not
 * happen.
 */
const TENANT_RESERVED_NAMES = ['api', 'dist'];

/**
 * The one build every tenant is served from, at the webroot beside the tenant
 * folders. The counterpart to the shared api/: neither is per-tenant, and a tenant
 * folder that held its own copy was the arrangement this replaced.
 */
const TENANT_DIST_DIR_NAME = 'dist';

/**
 * Every tenant's upload folder, relative to the tenant's own folder.
 *
 * One name for all of them, which is the point: api/lib/appdb.php resolves a
 * request's upload folder as app_dir() . '/' . 'files' with no per-tenant
 * configuration to read, and there is no longer a rule that could put one tenant's
 * uploads in another tenant's folder.
 */
const TENANT_FILES_DIR_NAME = 'files';

/**
 * Folders the app can upload into, relative to a tenant's files_dir.
 *
 * upload.php refuses a destination that is not an existing directory ("Invalid
 * upload directory"), so every one of these has to exist before the first upload.
 */
const TENANT_UPLOAD_SUBDIRS = [
    'banks_logos',
    'buy_payment_swifts',
    'buy_pi',
    'cars',
    'chat_files',
    'documents',
    'ids',
    'letter_head',
    'logo',
    'payments_swift',
    'purchase_pi',
    'sell_pi',
    'stamp',
    'uploads',
];

/**
 * Reference and auth rows a tenant needs before anyone can log in.
 *
 * The migration set normally supplies these itself - migration_v19_v25.sql,
 * 025_car_names_brand.sql and 026_brand_renames.sql are where the roles,
 * permissions, admin account and lookup rows come from - so on a clean run every
 * table below already has rows and this step is a no-op. It is kept as a floor: a
 * database rebuilt with force, or a future setup.sql that has lost the seeds, still
 * ends up with an account to log in with, and the reference lists a client curated
 * in the template (its own brand list, its own shipping lines) reach every tenant
 * built from it.
 *
 * This is every table the migrations leave non-empty on a fresh database, less
 * `versions`, which tenant_set_app_version() writes. Deriving it that way is what
 * keeps the two lists below from disagreeing with each other.
 */
const TENANT_SEED_TABLES = [
    'brands',
    'car_file_categories',
    'cars_names',
    'colors',
    'containers',
    'defaults',
    'discharge_ports',
    'jobs',
    'loading_ports',
    'permissions',
    'priorities',
    'role_permissions',
    'roles',
    'shipping_lines',
    'users',
];

/**
 * Tables that hold a client's OWN data and must therefore be empty in the template
 * and in every new tenant.
 *
 * This list is the guard on the seed. It is checked against the template BEFORE
 * anything is copied and against the new tenant AFTER, and a row in any of them
 * stops the run. Without it, "seed from a template database" is one mis-picked
 * template away from handing a new client another client's cars, clients, invoices
 * and chat history - which is exactly the mistake that is hardest to notice and most
 * expensive to undo.
 *
 * It is the set of tables that `api/setup.sql` plus the whole migration set leave
 * EMPTY on a fresh database, so it is derived rather than guessed. Re-derive it after
 * adding a migration that seeds reference data:
 *
 *   php -r 'require "api/config.php"; require "api/lib/tenant-provision.php";
 *     $c = tenant_app_credentials(); $p = tenant_pdo($c, null);
 *     $p->exec("CREATE DATABASE cars_probe_empty DEFAULT CHARACTER SET utf8mb4");
 *     tenant_apply_schema($c, $c, "cars_probe_empty", function ($l) {});
 *     foreach (tenant_pdo($c, "cars_probe_empty")->query(
 *       "SELECT table_name FROM information_schema.tables
 *        WHERE table_schema = DATABASE() AND table_type = \"BASE TABLE\"")->fetchAll(PDO::FETCH_COLUMN) as $t) {
 *       if ((int) tenant_pdo($c, "cars_probe_empty")->query("SELECT COUNT(*) FROM `$t`")->fetchColumn() === 0) echo "$t\n";
 *     }
 *     $p->exec("DROP DATABASE cars_probe_empty");'
 *
 * `dbs` is deliberately absent: every tenant records its own row in its own copy of
 * that table (tenant_write_self_row), so one row is correct.
 */
const TENANT_BUSINESS_TABLES = [
    'banks',
    'buy_bill',
    'buy_details',
    'buy_payments',
    'car_apgrades',
    'car_file_physical_tracking',
    'car_file_transfers',
    'car_files',
    'car_name_media',
    'car_selections',
    'cars_stock',
    'chat_groups',
    'chat_last_read_message',
    'chat_messages',
    'chat_read_by',
    'chat_users',
    'clients',
    'custom_clearance_agents',
    'db_updates',
    'loaded_containers',
    'loading',
    'login',
    'rates',
    'selection_comments',
    'selection_ownership_history',
    'sell_bill',
    'sell_payments',
    'supplier_credibility_checks',
    'suppliers',
    'tasks',
    'team_members',
    'teams',
    'tracking',
    'transfer_details',
    'transfers',
    'transfers_inter',
    'upgrades',
    'warehouses',
];

/**
 * Migrations that must not be replayed against the tenant database.
 */
const TENANT_MIGRATION_SKIPS = [
    // A rollback, paired with 001. Applying it forward would undo the feature.
    '001_car_files_system_rollback.sql' => 'rollback of 001, never applied forward',
    // The only migration in the set that targets the registry database. It is
    // applied to merhab_databases instead (it is guarded by information_schema
    // checks, so re-running is a no-op).
    '031_login_api_token.sql' => 'targets merhab_databases, not the tenant',
];

/**
 * Migrations that belong to the REGISTRY database rather than to any tenant.
 *
 * Listed separately from TENANT_MIGRATION_SKIPS because "skipped for the tenant" and
 * "applied to merhab_databases" are different outcomes, and the plan reports them
 * differently: a plan that said 031 was skipped would not say who it went to.
 */
const TENANT_REGISTRY_MIGRATIONS = ['031_login_api_token.sql'];

/**
 * MySQL errors that mean "this migration is already in setup.sql", which is the
 * normal state: setup.sql is kept in sync with the migration set (DEPLOYMENT.md
 * 4.4), so replaying the migrations on a fresh database repeats work. Anything NOT
 * on this list stops the run.
 */
const TENANT_BENIGN_ERRNOS = [
    1022 => 'duplicate key',
    1050 => 'table already exists',
    1060 => 'duplicate column',
    1061 => 'duplicate key name',
    1062 => 'duplicate entry',
    1091 => 'index or key does not exist (already dropped)',
    1359 => 'trigger already exists',
    1826 => 'duplicate index name',
];

/**
 * Must match src/composables/useVersionCheck.js. A mismatch is what shows the
 * version-mismatch banner, so it is written here rather than left at NULL.
 */
const TENANT_APP_VERSION = 25;

// ---------------------------------------------------------------------------
// Pure helpers
// ---------------------------------------------------------------------------

/**
 * Check a client/database name, returning it unchanged.
 *
 * @throws TenantProvisionError
 */
function tenant_assert_valid_db_name(string $dbName): string
{
    $dbName = trim($dbName);

    if ($dbName === '') {
        throw new TenantProvisionError('the database name is empty');
    }

    if (preg_match(TENANT_DB_NAME_PATTERN, $dbName) !== 1) {
        throw new TenantProvisionError(
            "'{$dbName}' cannot be used as a client name. Use a lowercase letter, "
            . 'then lowercase letters, digits or underscores, starting with the letter '
            . '- it is the database name, the folder under the webroot and the URL '
            . 'prefix all at once, so uppercase, dashes, dots and spaces are refused '
            . 'rather than normalised into a name that no longer matches.'
        );
    }

    if (in_array($dbName, TENANT_RESERVED_NAMES, true)) {
        throw new TenantProvisionError(
            "'{$dbName}' is the shared {$dbName}/ folder at the webroot, so it cannot also be a client. "
            . 'Choose another name.'
        );
    }

    return $dbName;
}

/**
 * Backslashes to slashes, no leading or trailing slash.
 *
 * The registry records files_dir and js_dir with a leading slash because they are
 * recorded against the DEPLOYMENT root (a tenant's app folder and its uploads
 * folder are siblings). This returns the bare name; the leading slash is put back
 * when the row is written.
 */
function tenant_normalise_dir(string $dir): string
{
    return trim(str_replace('\\', '/', trim($dir)), '/');
}

/**
 * The upload folder name derived from a database name.
 *
 * 'files', inside the tenant's own folder. It used to be <db>_files, a SIBLING of
 * the tenant folder, which meant the uploads lived next to the app rather than in
 * it - so the tenant's folder was never the whole tenant, and anything that treated
 * it as such (a backup, an rsync of the tenant, a listing) had to know the naming
 * rule too. One folder per tenant now holds everything that tenant owns.
 *
 * Derived, never typed: the same name on every tenant, so there is nothing to
 * disagree with.
 */
function tenant_files_dir_name(string $dbName): string
{
    // Validated for the side effect of refusing a name that is not a tenant name.
    // The return value does not depend on it any more.
    tenant_assert_valid_db_name($dbName);

    return TENANT_FILES_DIR_NAME;
}

/**
 * The plan for one tenant, computed without touching a database or the filesystem.
 *
 * Pure on purpose: the DB manager renders this as the confirmation step before the
 * Provision button does anything, and `php api/lib/tenant-provision.php
 * --describe-plan <db>` prints the same thing for an operator. If the plan and the
 * run disagreed, the confirmation would be worth nothing.
 *
 * @param array{
 *     db_name:string,
 *     db_code?:string|null,
 *     webroot?:string,
 *     seed_source?:string|null,
 *     shared_api?:bool,
 *     app_version?:int
 * } $opts
 * @return array<string,mixed>
 */
function tenant_plan(array $opts): array
{
    $dbName = tenant_assert_valid_db_name((string) ($opts['db_name'] ?? ''));

    // The deployment root: the folder that holds api/, the tenant folders and the
    // upload folders. Locally that is the repository, on the server /var/www.
    $webroot = rtrim(str_replace('\\', '/', (string) ($opts['webroot'] ?? '')), '/');
    if ($webroot === '') {
        $webroot = dirname(__DIR__, 2);
    }

    // Inside the tenant folder, not beside it: the tenant folder is the whole
    // tenant now, so it holds both its app and its uploads.
    $filesDirName = tenant_files_dir_name($dbName);
    $seedSource = isset($opts['seed_source']) && $opts['seed_source'] !== null
        ? (string) $opts['seed_source']
        : null;

    if ($seedSource !== null) {
        $seedSource = tenant_assert_valid_db_name($seedSource);
        if ($seedSource === $dbName) {
            throw new TenantProvisionError('the seed source cannot be the tenant database itself');
        }
    }

    $migrations = [];
    foreach (tenant_migration_files() as $file) {
        $name = basename($file);

        // Checked before the skip list, because 031 is in both: for a tenant it is
        // skipped, and this says where it went instead.
        if (in_array($name, TENANT_REGISTRY_MIGRATIONS, true)) {
            $migrations[] = [
                'file' => $name,
                'target' => 'registry',
                'reason' => 'the only migration that targets the registry database',
            ];
            continue;
        }

        if (isset(TENANT_MIGRATION_SKIPS[$name])) {
            $migrations[] = ['file' => $name, 'target' => 'skip', 'reason' => TENANT_MIGRATION_SKIPS[$name]];
            continue;
        }

        $migrations[] = ['file' => $name, 'target' => 'tenant', 'reason' => ''];
    }

    return [
        'db_name' => $dbName,
        'db_code' => isset($opts['db_code']) && $opts['db_code'] !== null ? (string) $opts['db_code'] : null,
        'webroot' => $webroot,
        // Bare names for the filesystem, leading-slash forms for the registry.
        'app_folder' => $webroot . '/' . $dbName,
        'files_folder' => $webroot . '/' . $dbName . '/' . $filesDirName,
        'js_dir' => '/' . $dbName,
        'files_dir' => '/' . $dbName . '/' . $filesDirName,
        'url_path' => '/' . $dbName . '/',
        'upload_subdirs' => TENANT_UPLOAD_SUBDIRS,
        'seed_tables' => TENANT_SEED_TABLES,
        'business_tables' => TENANT_BUSINESS_TABLES,
        'seed_source' => $seedSource,
        'app_version' => isset($opts['app_version']) ? (int) $opts['app_version'] : TENANT_APP_VERSION,
        // The one build every tenant is served from, beside the tenant folders and
        // at the same level. nginx aliases it into each tenant's location block, so
        // it is part of the plan rather than something a deploy decides later.
        'dist_folder' => $webroot . '/' . TENANT_DIST_DIR_NAME,
        'shared_api' => !empty($opts['shared_api']),
        'migrations' => $migrations,
    ];
}

/**
 * Every migration file, in order. Reading the directory (rather than a hardcoded
 * list) is what makes a new migration apply to a new tenant without editing this
 * file.
 *
 * @return list<string>
 */
function tenant_migration_files(): array
{
    $files = glob(__DIR__ . '/../migrations/*.sql') ?: [];
    sort($files);

    return $files;
}

// ---------------------------------------------------------------------------
// Server configuration
// ---------------------------------------------------------------------------

/**
 * Root-owned deployment settings, if the one-time server setup wrote them.
 *
 * This is how the web user learns the things it must not be able to change: where
 * the webroot is, which folder holds the canonical build, which PHP-FPM socket the
 * nginx template uses. Written once by root (see deploy/cars-deploy.example.json),
 * read-only to everyone else. Absent locally and on an unconfigured server, in
 * which case the defaults below are the single-app development layout.
 *
 * @return array<string,string>
 */
function tenant_server_config(): array
{
    static $config = null;

    if ($config !== null) {
        return $config;
    }

    $defaults = [
        // Where api/ lives. Its PARENT is the deployment root for a shared api/.
        'api_dir' => dirname(__DIR__),
        // The deployment root the rendered server block serves out of. Required by the
        // renderer, and it was missing from this list while the renderer demanded it -
        // so no config file could ever satisfy it and "Apply nginx" failed on a
        // complete /etc/cars-deploy.json. Defaults to the folder containing api/, which
        // is the same definition tenant_has_shared_api() uses.
        'webroot' => dirname(__DIR__, 2),
        'php_socket' => 'php-fpm.sock',
        'canonical_build' => '',
        // The database every new client is built from. Left empty, provisioning falls
        // back to the app database - which on a machine that is also being used as a
        // live tenant is the wrong answer, and the template check refuses it with a
        // message about client data. Naming it here is what stops every provisioning
        // run from failing that way.
        'template_database' => '',
        'server_name' => '',
        // Lets Encrypt live/ directory for server_name. Empty means "no TLS
        // configured", which the nginx renderer refuses rather than quietly
        // emitting a server block with a certificate path that does not exist.
        'cert_dir' => '',
        'render_command' => '/usr/local/bin/cars-nginx-render',
        'render_output' => '/etc/nginx/sites-available/cars-multitenant.conf',
        'web_group' => '',
        // Credentials for the dbs registry, so the root-run renderer can read the
        // tenant list without inheriting the web server's environment. Not a
        // duplicate of api/config.php in spirit - that file reads from the
        // environment, which sudo scrubs, so a root process gets nothing from it.
        'registry' => ['host' => 'localhost', 'dbname' => 'merhab_databases', 'user' => '', 'pass' => ''],

        // Whether one api/ serves every tenant. null means "work it out", which is
        // the normal case; set it explicitly only to override the probe below.
        'shared_api' => null,

        // Whether this server block is the catch-all for its address and port.
        // Needed when clients arrive by bare IP or by an unlisted name; without it
        // nginx serves its own default site for anything that misses a location.
        'default_server' => false,
        // Whether to add the [::] listen lines. Off by default, which is what the
        // renderer has always emitted - documented here because the key existed in
        // the example config but could not be read from any file, so setting it did
        // nothing.
        'ipv6' => false,
    ];

    // Keys that are booleans, so a JSON true/false is not silently dropped by the
    // string-only loader below - which would leave "shared_api": false reading as
    // "unset, go and probe", which is the opposite of what was asked for.
    $booleanKeys = ['shared_api', 'default_server', 'ipv6'];

    // Where the server's configuration comes from, in order:
    //
    //   1. CARS_DEPLOY_CONFIG - an explicit, per-process override.
    //   2. /etc/cars-deploy.json - the production authority: root-owned, and out of
    //      reach of the web user.
    //   3. deploy/cars-deploy.local.json - a development convenience.
    //
    // (3) exists because /etc needs sudo to create, and without some way to name a
    // template database on a laptop the wizard cannot provision anything - which is
    // how the local path went unexercised long enough to hold a real bug. It is
    // deliberately outside api/: that directory is the web root, and this file holds
    // registry credentials, so a copy under api/ would be readable over HTTP. It is
    // also not deployed - deploy.sh ships dist/ and api/ only - so it cannot appear
    // on a server to be overridden by a stale copy.
    //
    // None of this decides what root does: sudo scrubs the environment, and
    // cars-nginx-render passes --config explicitly.
    $path = '/etc/cars-deploy.json';
    $override = getenv('CARS_DEPLOY_CONFIG');

    if (is_string($override) && $override !== '' && is_file($override) && is_readable($override)) {
        $path = $override;
    } elseif (!is_file($path) || !is_readable($path)) {
        $localConfig = dirname(__DIR__, 2) . '/deploy/cars-deploy.local.json';
        if (is_file($localConfig) && is_readable($localConfig)) {
            $path = $localConfig;
        }
    }

    if (is_file($path) && is_readable($path)) {
        $decoded = json_decode((string) @file_get_contents($path), true);
        if (is_array($decoded)) {
            foreach ($decoded as $key => $value) {
                if ($key === 'registry') {
                    if (is_array($value)) {
                        foreach ($value as $credKey => $cred) {
                            if (is_string($cred) && array_key_exists($credKey, $defaults['registry'])) {
                                $defaults['registry'][$credKey] = $cred;
                            }
                        }
                    }
                    continue;
                }
                // Unknown keys are ignored rather than passed through: this file is
                // root-owned but the app also reads it as www-data, and a typo in a
                // key name should be inert, not a silently unset value.
                if (in_array($key, $booleanKeys, true)) {
                    // Tri-state on purpose: null is "not configured, probe for it",
                    // which is different from false.
                    if ($value === null || is_bool($value)) {
                        $defaults[$key] = $value;
                    }
                    continue;
                }
                if (array_key_exists($key, $defaults) && is_string($value)) {
                    $defaults[$key] = $value;
                }
            }
        }
    }

    // Which file answered, so a message about a missing setting can name the file the
    // operator actually has to edit. On a development machine that is not /etc, and
    // "not set in /etc/cars-deploy.json" would send them to create a file they have no
    // permission to write.
    $defaults['config_path'] = is_file($path) ? $path : '';

    return $config = $defaults;
}

/**
 * Does one api/ serve every tenant on this installation?
 *
 * The distinction decides whether provisioning writes a copy of api/ into each
 * tenant's folder, and getting it wrong is expensive in a quiet way: 68 files per
 * client, every one of which has to be the same version, where a stale copy is not
 * a visible fault but a client quietly running an old authentication bug. So the
 * answer is worked out from the filesystem rather than assumed, and can be
 * overridden in /etc/cars-deploy.json when the probe cannot tell.
 *
 * The probe: in the shared layout the folder containing api/ is the deployment
 * root, and it holds other tenants' folders rather than an application. In the
 * per-tenant layout api/ sits inside one client's folder, next to the index.html of
 * that client's own build. So the question "is there an index.html beside api/" has
 * the answer in both cases:
 *
 *   shared      /var/www/api        -> /var/www/index.html        does not exist  => shared
 *   per-tenant  /var/www/acme/api   -> /var/www/acme/index.html   exists          => not shared
 *
 * One case it cannot decide: a server whose deployment root also holds a stray
 * index.html, which is exactly why the config override exists.
 */
function tenant_has_shared_api(): bool
{
    $configured = tenant_server_config()['shared_api'];
    if (is_bool($configured)) {
        return $configured;
    }

    // dirname(__DIR__, 2), not dirname(__DIR__): this file is in api/lib, so one
    // level up is api/ itself - which never holds the built application - and two is
    // the folder that contains api/.
    return !is_file(dirname(__DIR__, 2) . '/index.html');
}

// ---------------------------------------------------------------------------
// Credentials and connections
// ---------------------------------------------------------------------------

/**
 * @return array{host:string,user:string,pass:string,dbname:string}
 */
function tenant_app_credentials(): array
{
    global $db_config;

    if (!isset($db_config['dbname'])) {
        throw new TenantProvisionError('api/config.php has not been loaded');
    }

    return [
        'host' => $db_config['host'],
        'user' => $db_config['user'],
        'pass' => $db_config['pass'],
        'dbname' => $db_config['dbname'],
    ];
}

/**
 * @return array{host:string,user:string,pass:string,dbname:string}
 */
function tenant_registry_credentials(): array
{
    global $db_manager_config;

    if (!isset($db_manager_config['dbname'])) {
        throw new TenantProvisionError('api/db_manager_config.php has not been loaded');
    }

    return [
        'host' => $db_manager_config['host'],
        'user' => $db_manager_config['user'],
        'pass' => $db_manager_config['pass'],
        'dbname' => $db_manager_config['dbname'],
    ];
}

/**
 * @param array<string,string> $creds
 */
function tenant_pdo(array $creds, ?string $dbname = null): PDO
{
    $dsn = "mysql:host={$creds['host']}" . ($dbname !== null ? ";dbname={$dbname}" : '');

    try {
        $pdo = new PDO($dsn, $creds['user'], $creds['pass'], [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]);
    } catch (PDOException $e) {
        // PDO's message can contain the DSN, which carries the host and database
        // but never the password. It is still an operator tool, so the message is
        // kept - without it a wrong host and a wrong password look identical.
        throw new TenantProvisionError('cannot connect to MySQL as ' . $creds['user'] . ': ' . $e->getMessage());
    }

    $pdo->exec('SET NAMES utf8mb4');

    return $pdo;
}

// ---------------------------------------------------------------------------
// mysql CLI
// ---------------------------------------------------------------------------

/**
 * Run SQL through the mysql client, which understands DELIMITER, triggers and
 * comments - none of which a hand-rolled statement splitter does.
 *
 * @param array<string,string> $creds
 * @param list<string> $args
 * @param callable(string):void|null $log
 * @return array{code:int,output:string}
 */
function tenant_mysql_run(array $creds, array $args, string $sql, ?callable $log = null): array
{
    $command = array_merge(
        [
            'mysql',
            // First option by requirement: keeps a stray ~/.my.cnf from redirecting
            // the connection somewhere else.
            '--no-defaults',
            '--protocol=TCP',
            '--host=' . $creds['host'],
            '--user=' . $creds['user'],
            '--default-character-set=utf8mb4',
        ],
        $args
    );

    $descriptors = [
        0 => ['pipe', 'r'],
        1 => ['pipe', 'w'],
        2 => ['pipe', 'w'],
    ];

    // Only MYSQL_PWD (plus what mysql itself needs) - the password must not be
    // visible in `ps` output.
    $env = [
        'MYSQL_PWD' => $creds['pass'],
        'PATH' => getenv('PATH') ?: '/usr/bin:/bin',
        'HOME' => getenv('HOME') ?: '/tmp',
        'LANG' => 'en_US.UTF-8',
    ];

    $process = @proc_open($command, $descriptors, $pipes, null, $env);
    if (!is_resource($process)) {
        throw new TenantProvisionError('could not start the mysql client');
    }

    fwrite($pipes[0], $sql);
    fclose($pipes[0]);

    // Sequential reads: the error volume here is a handful of lines per file, so
    // neither pipe can fill and block the other.
    $stdout = stream_get_contents($pipes[1]);
    fclose($pipes[1]);
    $stderr = stream_get_contents($pipes[2]);
    fclose($pipes[2]);

    return ['code' => proc_close($process), 'output' => trim($stdout . "\n" . $stderr)];
}

/**
 * Run one SQL file, distinguishing "already applied" from a real failure.
 *
 * @param array<string,string> $creds
 * @param callable(string):void|null $log
 * @throws TenantProvisionError on a failure that is not "already applied"
 */
function tenant_mysql_apply_file(array $creds, string $database, string $file, string $label, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    if (!is_file($file)) {
        throw new TenantProvisionError("cannot read {$file}");
    }

    $sql = file_get_contents($file);
    if ($sql === false) {
        throw new TenantProvisionError("cannot read {$file}");
    }

    // migration_v19_v25.sql is a phpMyAdmin export whose section headings are drawn
    // as a row of dashes. That is not a comment - only `--` starts one - so the
    // client stops on it with a syntax error at the first separator. Dropping the
    // rule lines is the whole fix; nothing else in the file is touched.
    $sql = preg_replace('/^[ \t]*-{3,}[ \t]*\r?\n/m', '', $sql) ?? $sql;

    $result = tenant_mysql_run($creds, ['--database=' . $database], $sql);
    if ($result['code'] === 0) {
        $log("{$label}: ok");

        return;
    }

    preg_match_all('/ERROR (\d+)/', $result['output'], $matches);
    $codes = array_values(array_unique(array_map('intval', $matches[1])));

    if ($codes === []) {
        throw new TenantProvisionError("{$label} failed:\n" . $result['output']);
    }

    $unexpected = array_diff($codes, array_keys(TENANT_BENIGN_ERRNOS));
    if ($unexpected !== []) {
        throw new TenantProvisionError(
            "{$label} failed with errors that are not 'already applied':\n"
            . $result['output']
            . "\n    (error codes: " . implode(', ', $unexpected) . ' - nothing after this file was applied)'
        );
    }

    $reasons = array_map(static fn (int $code): string => $code . ' ' . TENANT_BENIGN_ERRNOS[$code], $codes);
    $log("{$label}: already applied (" . implode(', ', $reasons) . ')');
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

/**
 * The registry row is how a client is named before it exists: the DB manager
 * screen creates the row, and this reads it back.
 *
 * db_code is read but no longer used for anything: it wrote the tenant's db_code.json,
 * which is gone. The column stays, so the table is not the thing that has to change
 * during the cutover - but nothing reads it to pick a database.
 *
 * @return array{id:int,db_code:string,db_name:string,files_dir:string,js_dir:string}
 * @throws TenantProvisionError when there is no row for this database
 */
function tenant_find_row(PDO $registry, string $dbName): array
{
    $stmt = $registry->prepare('SELECT id, db_code, db_name, files_dir, js_dir FROM dbs WHERE db_name = ?');
    $stmt->execute([$dbName]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);

    if (!$row) {
        throw new TenantProvisionError(
            "no row for db_name \"{$dbName}\" in the registry."
            . "\n    Add it first: DB manager -> Databases -> Add, with the database name \"{$dbName}\"."
            . "\n    A folder is not a database; the two are named the same, which is what keeps them from drifting."
        );
    }

    $dbCode = trim((string) ($row['db_code'] ?? ''));

    return [
        'id' => (int) $row['id'],
        // api/lib/appdb.php only accepts db_[0-9a-f]+, so a row whose code does not
        // match that would provision a folder that resolves to nothing.
        'db_code' => $dbCode,
        'db_name' => trim((string) $row['db_name']),
        'files_dir' => tenant_normalise_dir((string) ($row['files_dir'] ?? '')),
        'js_dir' => tenant_normalise_dir((string) ($row['js_dir'] ?? '')),
    ];
}

/**
 * Make the row's js_dir/files_dir agree with the derived names.
 *
 * The folders are a function of the database name, so a row that says otherwise is
 * a row that would send this tenant's uploads into a folder belonging to something
 * else. Reconciling it here - rather than trusting whatever the form submitted - is
 * what keeps that from being a decision an operator has to get right.
 *
 * @param array{id:int,db_code:string,db_name:string,files_dir:string,js_dir:string} $row
 * @param array<string,mixed> $plan
 * @param callable(string):void|null $log
 * @return array{row:array<string,mixed>,changed:bool}
 */
function tenant_reconcile_row_dirs(PDO $registry, array $row, array $plan, ?callable $log = null): array
{
    $log ??= static function (string $line): void {
    };

    $wantJs = tenant_normalise_dir($plan['js_dir']);
    $wantFiles = tenant_normalise_dir($plan['files_dir']);

    if ($row['js_dir'] === $wantJs && $row['files_dir'] === $wantFiles) {
        return ['row' => $row, 'changed' => false];
    }

    $stmt = $registry->prepare('UPDATE dbs SET js_dir = ?, files_dir = ? WHERE id = ?');
    $stmt->execute([$plan['js_dir'], $plan['files_dir'], $row['id']]);

    $log(sprintf(
        'registry row %d: js_dir %s -> %s, files_dir %s -> %s (derived from the database name)',
        $row['id'],
        $row['js_dir'] === '' ? '(empty)' : '/' . $row['js_dir'],
        $plan['js_dir'],
        $row['files_dir'] === '' ? '(empty)' : '/' . $row['files_dir'],
        $plan['files_dir']
    ));

    $row['js_dir'] = $wantJs;
    $row['files_dir'] = $wantFiles;

    return ['row' => $row, 'changed' => true];
}

// ---------------------------------------------------------------------------
// Database
// ---------------------------------------------------------------------------

function tenant_database_exists(array $appCreds, string $database): bool
{
    $pdo = tenant_pdo($appCreds);
    $stmt = $pdo->prepare(
        'SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?'
    );
    $stmt->execute([$database]);

    return (int) $stmt->fetchColumn() > 0;
}

function tenant_create_database(array $appCreds, string $database, bool $force, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    $pdo = tenant_pdo($appCreds);
    $quoted = '`' . str_replace('`', '``', $database) . '`';

    if ($force) {
        $pdo->exec('DROP DATABASE IF EXISTS ' . $quoted);
        $log('dropped ' . $database . ' (force)');
    }

    $pdo->exec(sprintf(
        'CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci',
        $quoted
    ));
    $log($database . ($force ? ' created' : ' exists'));
}

/**
 * Apply api/setup.sql, then every migration in order.
 *
 * The migration replay is mostly a no-op by design (see TENANT_BENIGN_ERRNOS):
 * setup.sql is already at the final shape, so each file repeats work that is done.
 * It is run anyway so the tenant is provably at the migration set's state, and so a
 * migration that has NOT been folded into setup.sql yet is not silently missed.
 *
 * @param array<string,string> $appCreds
 * @param array<string,string> $registryCreds
 * @param callable(string):void|null $log
 */
function tenant_apply_schema(array $appCreds, array $registryCreds, string $database, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    tenant_mysql_apply_file($appCreds, $database, __DIR__ . '/../setup.sql', 'api/setup.sql', $log);

    foreach (tenant_migration_files() as $file) {
        $name = basename($file);

        if (isset(TENANT_MIGRATION_SKIPS[$name])) {
            $log("skipped {$name} (" . TENANT_MIGRATION_SKIPS[$name] . ')');
            continue;
        }

        tenant_mysql_apply_file($appCreds, $database, $file, $name, $log);
    }

    // The registry credentials, not the app ones: config.php's db_name is the app
    // database (merhab_cars), and passing the app credentials here sent this
    // migration to the tenant's sibling instead of merhab_databases.
    foreach (TENANT_REGISTRY_MIGRATIONS as $name) {
        $path = __DIR__ . '/../migrations/' . $name;
        if (!is_file($path)) {
            continue;
        }

        tenant_mysql_apply_file(
            $registryCreds,
            $registryCreds['dbname'],
            $path,
            $name . ' -> ' . $registryCreds['dbname'],
            $log
        );
    }
}

// ---------------------------------------------------------------------------
// Reference data
// ---------------------------------------------------------------------------

/**
 * Copy the reference/auth rows from the template database.
 *
 * Column lists are intersected rather than using SELECT *, because the source may
 * predate a migration that setup.sql already contains, and a column count mismatch
 * would abort the whole seed.
 *
 * @param callable(string):void|null $log
 * @return array<string,int> rows copied per table
 */
function tenant_seed_reference(PDO $target, array $appCreds, string $sourceDatabase, ?callable $log = null): array
{
    $log ??= static function (string $line): void {
    };

    $source = tenant_pdo($appCreds, $sourceDatabase);
    $copied = [];

    foreach (TENANT_SEED_TABLES as $table) {
        if (!tenant_table_exists($target, $table)) {
            $log("{$table}: not in this schema, skipped");
            continue;
        }

        $existing = (int) $target->query("SELECT COUNT(*) FROM `{$table}`")->fetchColumn();
        if ($existing > 0) {
            $log("{$table}: already has {$existing} row(s), left alone");
            continue;
        }

        $columns = array_values(array_intersect(
            tenant_table_columns($target, $table),
            tenant_table_columns($source, $table)
        ));

        if ($columns === []) {
            $log("{$table}: no common columns with {$sourceDatabase}, skipped");
            continue;
        }

        $list = implode(', ', array_map(static fn (string $c): string => "`{$c}`", $columns));
        $target->exec("INSERT INTO `{$table}` ({$list}) SELECT {$list} FROM `{$sourceDatabase}`.`{$table}`");
        $target->exec("ALTER TABLE `{$table}` AUTO_INCREMENT = 1");

        // A session token copied from the template would be a live credential in a
        // second place; the tenant mints its own on first login.
        if (in_array('api_token', $columns, true)) {
            $target->exec("UPDATE `{$table}` SET `api_token` = NULL WHERE `api_token` IS NOT NULL");
        }

        $copied[$table] = (int) $target->query("SELECT COUNT(*) FROM `{$table}`")->fetchColumn();
        $log("{$table}: seeded {$copied[$table]} row(s) from {$sourceDatabase}");
    }

    return $copied;
}

/**
 * Refuse to seed from a template that holds client data.
 *
 * @param callable(string):void|null $log
 * @throws TenantProvisionError listing every non-empty table
 */
function tenant_assert_template_clean(array $appCreds, string $sourceDatabase, ?callable $log = null): array
{
    $log ??= static function (string $line): void {
    };

    $source = tenant_pdo($appCreds, $sourceDatabase);
    $dirty = [];

    foreach (TENANT_BUSINESS_TABLES as $table) {
        if (!tenant_table_exists($source, $table)) {
            continue;
        }

        $count = (int) $source->query("SELECT COUNT(*) FROM `{$table}`")->fetchColumn();
        if ($count > 0) {
            $dirty[$table] = $count;
        }
    }

    if ($dirty !== []) {
        $lines = [];
        foreach ($dirty as $table => $count) {
            $lines[] = "    {$table}: {$count} row(s)";
        }

        throw new TenantProvisionError(
            "\"{$sourceDatabase}\" holds client data, so it cannot be used as a template:\n"
            . implode("\n", $lines)
            . "\n    A template carries reference and auth rows only. Empty those tables, or point"
            . "\n    the seed source at a database that is already empty."
        );
    }

    $log("{$sourceDatabase}: reference-only, verified across " . count(TENANT_BUSINESS_TABLES) . ' business tables');

    return $dirty;
}

/**
 * Confirm a freshly provisioned tenant has no business data in it.
 *
 * The other half of tenant_assert_template_clean: even with a clean template, a
 * step that copied too much would be caught here rather than at the first invoice.
 *
 * @param callable(string):void|null $log
 */
function tenant_assert_tenant_empty(PDO $target, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    $dirty = [];
    foreach (TENANT_BUSINESS_TABLES as $table) {
        if (!tenant_table_exists($target, $table)) {
            continue;
        }

        $count = (int) $target->query("SELECT COUNT(*) FROM `{$table}`")->fetchColumn();
        if ($count > 0) {
            $dirty[$table] = $count;
        }
    }

    if ($dirty !== []) {
        $lines = [];
        foreach ($dirty as $table => $count) {
            $lines[] = "    {$table}: {$count} row(s)";
        }

        throw new TenantProvisionError(
            "the new tenant already has client data in it:\n" . implode("\n", $lines)
            . "\n    Nothing else ran after this check, so this is a template or seed problem, not a"
            . "\n    partial provisioning. Re-run with force once the template is fixed."
        );
    }

    $log('no client data in any of ' . count(TENANT_BUSINESS_TABLES) . ' business tables');
}

/**
 * The tenant's own copy of its registry row.
 *
 * Each tenant database carries a `dbs` table (setup.sql creates it). Nothing reads
 * it any more - api.php and api/lib/appdb.php resolve the tenant from the request
 * itself, and the upload folder from the tenant folder's own name - but the row is
 * written so the table is not a lie about what this deployment is.
 *
 * @param array<string,mixed> $tenant
 * @param callable(string):void|null $log
 */
function tenant_write_self_row(PDO $target, array $tenant, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    if (!tenant_table_exists($target, 'dbs')) {
        $log('dbs: not in this schema, skipped');

        return;
    }

    $existing = (int) $target->query('SELECT COUNT(*) FROM `dbs`')->fetchColumn();
    if ($existing > 0) {
        $log('dbs: already has a row, left alone');

        return;
    }

    $stmt = $target->prepare(
        'INSERT INTO `dbs` (db_code, db_name, files_dir, js_dir, is_created) VALUES (?, ?, ?, ?, 1)'
    );
    $stmt->execute([$tenant['db_code'], $tenant['db_name'], $tenant['files_dir'], $tenant['js_dir']]);

    $log('dbs: recorded this deployment (db_code ' . $tenant['db_code'] . ')');
}

function tenant_set_app_version(PDO $target, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    if (!tenant_table_exists($target, 'versions')) {
        $log('versions: not in this schema, skipped');

        return;
    }

    $existing = (int) $target->query('SELECT COUNT(*) FROM `versions`')->fetchColumn();
    if ($existing > 0) {
        $log('versions: already set, left alone');

        return;
    }

    $target->exec('INSERT INTO `versions` (id, version) VALUES (' . TENANT_APP_VERSION . ')');
    $log('versions: set to ' . TENANT_APP_VERSION);
}

function tenant_table_exists(PDO $pdo, string $table): bool
{
    $stmt = $pdo->prepare(
        'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?'
    );
    $stmt->execute([$table]);

    return (int) $stmt->fetchColumn() > 0;
}

/**
 * @return list<string>
 */
function tenant_table_columns(PDO $pdo, string $table): array
{
    $stmt = $pdo->prepare(
        'SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?'
    );
    $stmt->execute([$table]);

    return array_map('strval', $stmt->fetchAll(PDO::FETCH_COLUMN));
}

// ---------------------------------------------------------------------------
// Folders and files
// ---------------------------------------------------------------------------

/**
 * Create the tenant's app folder, its upload folder and every upload subdirectory.
 *
 * @param array<string,mixed> $plan
 * @param callable(string):void|null $log
 * @return array{created:list<string>,group:string}
 */
function tenant_ensure_dirs(array $plan, ?string $group = null, ?callable $log = null): array
{
    $log ??= static function (string $line): void {
    };

    $webroot = (string) $plan['webroot'];
    $created = [];

    foreach ([(string) $plan['app_folder'], (string) $plan['files_folder']] as $dir) {
        if (!is_dir($dir) && !mkdir($dir, 0755, true) && !is_dir($dir)) {
            throw new TenantProvisionError(
                'could not create ' . $dir . '. On the server the webroot has to be writable by the web'
                . ' user: chgrp ' . ($group !== null && $group !== '' ? $group : 'www-data')
                . ' ' . $webroot . ' && chmod g+w ' . $webroot
            );
        }
        $log('folder: ' . $dir);
    }

    // upload.php resolves the destination folder and refuses anything that is not a
    // directory, so these have to exist before the first upload.
    foreach (TENANT_UPLOAD_SUBDIRS as $subdir) {
        $path = (string) $plan['files_folder'] . '/' . $subdir;
        if (!is_dir($path) && !mkdir($path, 0755, true) && !is_dir($path)) {
            throw new TenantProvisionError('could not create ' . $path);
        }
        $created[] = $path;
    }
    $log('folders: ' . count(TENANT_UPLOAD_SUBDIRS) . ' upload subdirectories under '
        . (string) $plan['files_folder'] . '/');

    // A setgid directory keeps the group on everything created inside it, so the
    // web user can write into the uploads without the folder being world-writable.
    if ($group !== null && $group !== '') {
        foreach ([(string) $plan['files_folder']] as $dir) {
            @chgrp($dir, $group);
            @chmod($dir, 02755);
        }
    }

    return ['created' => $created, 'group' => (string) $group];
}

/**
 * A tenant folder is bound to its database by its NAME, so it needs no binding
 * file. db_code.json did that job once: it was fetched by the build to find the
 * upload folder and read by api/lib/appdb.php to pick the database. With one shared
 * api/ that is circular anyway - a per-tenant file read by shared code.
 *
 * So this deletes it instead of writing it. Provisioning is idempotent and is how
 * a tenant is repaired, which makes this the right place to clear a stale copy: a
 * leftover db_code.json in the folder is a second, disagreeing answer to "which
 * database is this", and leaving it would mean the cutover was never finished.
 *
 * @param array<string,mixed> $plan
 * @param callable(string):void|null $log
 */
function tenant_remove_legacy_binding_files(array $plan, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    foreach (['db_code.json'] as $name) {
        $path = (string) $plan['app_folder'] . '/' . $name;
        if (!is_file($path)) {
            continue;
        }
        if (!unlink($path)) {
            throw new TenantProvisionError('could not remove the legacy ' . $path);
        }
        $log('removed ' . $path . ' (a tenant folder is bound to its database by its name)');
    }
}

/**
 * Remove a per-tenant copy of the shared api/.
 *
 * An old deployment kept a full copy of the code in the tenant folder, as the way to
 * identify the tenant: api/lib/appdb.php read it, and it did not exist for the shared
 * api/. So the copy is not merely redundant - it is an older version of the code,
 * sitting exactly where a request to /<tenant>/api/ used to find it, and leaving it
 * means the cutover depends on which copy a request happens to reach.
 *
 * *.local.php is kept, deliberately. Those hold credentials, are not in the
 * repository, and an operator may have copied them in; a migration is not the place
 * to delete them. The folder is then reported as needing a manual delete rather than
 * forced, so the leftover is visible instead of silent.
 *
 * @param array<string,mixed> $plan
 * @param callable(string):void|null $log
 */
function tenant_remove_legacy_app_copy(array $plan, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    $appCopy = (string) $plan['app_folder'] . '/api';
    if (!is_dir($appCopy)) {
        return;
    }

    $kept = [];
    $iterator = new RecursiveIteratorIterator(
        new RecursiveDirectoryIterator($appCopy, FilesystemIterator::SKIP_DOTS),
        RecursiveIteratorIterator::CHILD_FIRST
    );
    foreach ($iterator as $item) {
        if (preg_match('/\.local\.php$/i', $item->getFilename()) === 1) {
            $kept[] = $item->getPathname();
            continue;
        }
        if ($item->isDir()) {
            @rmdir($item->getPathname());
        } else {
            @unlink($item->getPathname());
        }
    }

    if (!@rmdir($appCopy)) {
        throw new TenantProvisionError(
            'could not remove the legacy ' . $appCopy . ' copy of the shared api/. Everything in it has been'
            . ' deleted except ' . (count($kept) > 0 ? implode(', ', $kept) : 'nothing')
            . '. Move the credential file(s) out of the way, then delete the folder by hand.'
        );
    }

    $log('removed ' . $appCopy . ' (one shared api/ serves every tenant)');
}

/**
 * Copy a built app into the one shared dist/ every tenant is served from.
 *
 * There is no per-tenant build any more. nginx aliases <webroot>/dist/ into every
 * tenant's location block, so deploying a release is one rsync that updates every
 * client at once - and there is nothing per-tenant left to exclude, because the
 * tenant folders hold only their own files/ and this destination has none.
 *
 * The function keeps the name and the $plan argument it had when it wrote into a
 * tenant folder, and takes the destination from $plan['dist_folder'], so its two
 * callers (the DB manager's re-deploy action and the CLI's --sync-only) do not each
 * have to reconstruct the webroot. Provisioning deliberately does NOT call it:
 * adding a client must not re-copy a build, and it only reports whether the shared
 * dist/ is present.
 *
 * @param array<string,mixed> $plan
 * @param callable(string):void|null $log
 */
function tenant_deploy_app(array $plan, ?string $buildSource, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    if ($buildSource === null || $buildSource === '') {
        throw new TenantProvisionError(
            'no build to deploy: no canonical build folder is configured. Set canonical_build in'
            . ' /etc/cars-deploy.json, or pass a build folder.'
        );
    }

    $buildSource = rtrim($buildSource, '/');
    if (!is_dir($buildSource)) {
        throw new TenantProvisionError(
            "the build folder {$buildSource} does not exist. Build it first, or point canonical_build at"
            . ' the folder that holds the current build.'
        );
    }

    $target = (string) ($plan['dist_folder'] ?? '');

    // --delete prunes the old hashed assets, which is the only way a feature
    // deleted from src/ stops being reachable on every deployed client.
    //
    // Safe here because of what dist/ does NOT contain. Uploads are in each
    // <tenant>/files/ and credentials in api/*.local.php, neither of which is under
    // this destination - and that is the whole argument for putting the build
    // somewhere of its own instead of inside a tenant folder, where the same
    // --delete would need files/ excluded and could delete a client's documents by
    // accident.
    tenant_rsync(
        [
            '--archive',
            '--delete',
            '--exclude', '.DS_Store',
        ],
        [$buildSource . '/', $target . '/'],
        'build -> ' . basename($target) . '/',
        $log
    );

    // No folder per client is written. One api/ and one dist/ serve them all; a
    // client folder holds files/ and nothing else.
}

/**
 * @param list<string> $options
 * @param list<string> $endpoints
 * @param callable(string):void|null $log
 */
function tenant_rsync(array $options, array $endpoints, string $label, ?callable $log = null): void
{
    $log ??= static function (string $line): void {
    };

    $command = array_merge(['rsync', '--itemize-changes'], $options, $endpoints);

    $process = @proc_open(
        $command,
        [1 => ['pipe', 'w'], 2 => ['pipe', 'w']],
        $pipes,
        null,
        ['PATH' => getenv('PATH') ?: '/usr/bin:/bin', 'HOME' => getenv('HOME') ?: '/tmp']
    );

    if (!is_resource($process)) {
        throw new TenantProvisionError('could not start rsync');
    }

    $stdout = stream_get_contents($pipes[1]);
    fclose($pipes[1]);
    $stderr = stream_get_contents($pipes[2]);
    fclose($pipes[2]);

    $code = proc_close($process);
    if ($code !== 0) {
        throw new TenantProvisionError("rsync {$label} failed:\n" . trim($stdout . "\n" . $stderr));
    }

    $changed = trim($stdout);
    $log($label . ($changed === '' ? ': already up to date' : ': ' . count(explode("\n", $changed)) . ' file(s) transferred'));
}

// ---------------------------------------------------------------------------
// Orchestration
// ---------------------------------------------------------------------------

/**
 * Provision one tenant.
 *
 * @param array{
 *     db_name:string,
 *     app_creds?:array<string,string>,
 *     registry_creds?:array<string,string>,
 *     webroot?:string,
 *     seed_source?:string|null,
 *     force?:bool,
 *     shared_api?:bool,
 *     group?:string|null,
 *     log?:callable(string):void|null,
 *     skip_template_check?:bool
 * } $opts
 * @return array<string,mixed> a report, safe to hand to the UI as JSON
 * @throws TenantProvisionError
 */
function tenant_provision(array $opts): array
{
    $log = isset($opts['log']) && is_callable($opts['log'])
        ? $opts['log']
        : static function (string $line): void {
        };

    $appCreds = $opts['app_creds'] ?? tenant_app_credentials();
    $registryCreds = $opts['registry_creds'] ?? tenant_registry_credentials();
    $dbName = tenant_assert_valid_db_name((string) ($opts['db_name'] ?? ''));

    $sharedApi = !empty($opts['shared_api']);
    if ($sharedApi && !isset($opts['webroot'])) {
        // With one shared api/, the deployment root is api/'s parent.
        $webroot = dirname((string) tenant_server_config()['api_dir']);
    } else {
        $webroot = (string) ($opts['webroot'] ?? '');
    }

    // Where the reference data is copied from. Deliberately NOT defaulted to
    // $appCreds['dbname']: that used to be a live tenant on this machine, so a
    // provisioning run with no seed source would quietly rebuild the new tenant from
    // another client's data. Both callers (db_manager_api.php, deploy/setup-mig-27.php)
    // read template_database from the server config and pass it in; when neither is
    // set there is nothing safe to fall back to, so it is refused.
    $seedSource = tenant_assert_valid_db_name((string) ($opts['seed_source'] ?? ''));

    $plan = tenant_plan([
        'db_name' => $dbName,
        'webroot' => $webroot,
        'seed_source' => $seedSource,
        'shared_api' => $sharedApi,
    ]);

    $registry = tenant_pdo($registryCreds, $registryCreds['dbname']);
    $row = tenant_find_row($registry, $dbName);
    $plan['db_code'] = $row['db_code'];
    $reconciled = tenant_reconcile_row_dirs($registry, $row, $plan, $log);

    $report = [
        'db_name' => $dbName,
        'db_code' => $row['db_code'],
        'registry_row_id' => $row['id'],
        'app_folder' => $plan['app_folder'],
        'files_folder' => $plan['files_folder'],
        'js_dir' => $plan['js_dir'],
        'files_dir' => $plan['files_dir'],
        'url_path' => $plan['url_path'],
        'seed_source' => $plan['seed_source'],
        'shared_api' => $sharedApi,
        'reconciled_row_dirs' => $reconciled['changed'],
        'seeded' => [],
        'steps' => [],
    ];

    $step = static function (string $name) use (&$report, $log): void {
        $report['steps'][] = $name;
        $log("\n==> {$name}");
    };

    // Before anything is created, dropped or written: a template holding client data
    // is the one mistake that would otherwise leave a half-built tenant behind, and
    // it is entirely predictable from the source alone.
    if (empty($opts['skip_template_check'])) {
        $step('Template check: ' . $plan['seed_source']);
        tenant_assert_template_clean($appCreds, (string) $plan['seed_source'], $log);
    }

    $step('Database');
    tenant_create_database($appCreds, $dbName, !empty($opts['force']), $log);

    $step('Schema: api/setup.sql, then api/migrations in order');
    tenant_apply_schema($appCreds, $registryCreds, $dbName, $log);

    $step('Reference data');
    $target = tenant_pdo($appCreds, $dbName);
    $report['seeded'] = tenant_seed_reference($target, $appCreds, (string) $plan['seed_source'], $log);

    // Spelled out, because the log line above is misleading on its own.
    //
    // setup.sql carries the reference INSERTs, so by the time this runs every seed
    // table already has rows and tenant_seed_reference() - which only fills an EMPTY
    // table, deliberately, so a provisioning run can never overwrite a tenant's edited
    // roles or permissions - copies nothing. "Reference data from cars_template" in
    // the log implies a copy happened; none did. An operator who trusted that would
    // assume the template's reference data had been applied when the tenant actually
    // has setup.sql's.
    $filledFromSource = array_keys(array_filter($report['seeded'], static fn (int $rows): bool => $rows > 0));
    $report['reference'] = [
        'source' => (string) $plan['seed_source'],
        // The source of truth for a new tenant's reference data.
        'from_setup_sql' => count(TENANT_SEED_TABLES) - count($filledFromSource),
        'filled_from_source' => $filledFromSource,
    ];
    $log(sprintf(
        'reference data: setup.sql provides it; %d of %d table(s) were empty and filled from %s',
        count($filledFromSource),
        count(TENANT_SEED_TABLES),
        $plan['seed_source']
    ));

    tenant_assert_tenant_empty($target, $log);
    tenant_write_self_row($target, $reconciled['row'], $log);
    tenant_set_app_version($target, $log);

    $step('Folders');
    $group = array_key_exists('group', $opts) ? $opts['group'] : tenant_server_config()['web_group'];
    tenant_ensure_dirs($plan, $group === null ? '' : (string) $group, $log);

    // Before the build lands: an old deployment's build and db_code.json are stale
    // copies, and this step is what makes re-running provisioning a repair.
    $step('Legacy copies');
    tenant_remove_legacy_binding_files($plan, $log);
    tenant_remove_legacy_app_copy($plan, $log);

    // The shared build is not this step's business.
    //
    // It used to be: a build was rsynced into the client folder as the last step of
    // provisioning, so every new client re-copied the whole release - and on a server
    // where dist/ was empty, it also meant the run's success said nothing about
    // whether the client would actually load. The build is now one folder beside the
    // client folders, deployed once per release and shared by all of them, so
    // provisioning a client neither needs it nor writes it. What it can still do is
    // report whether the shared build is there, which is the thing an operator
    // actually needs to know at this point.
    $step('Shared build');
    $distFolder = (string) $plan['dist_folder'];
    if (is_file($distFolder . '/index.html')) {
        $log('shared build: ' . $distFolder . ' - every client on this server is served from it');
    } else {
        $log('shared build: NOT DEPLOYED. ' . $distFolder . '/index.html is missing, so this');
        $log('               client will 404 until a build is deployed there. Set');
        $log('               canonical_build in /etc/cars-deploy.json and use "Re-deploy build".');
    }

    return $report;
}

/**
 * What exists for a tenant and what is missing, without changing anything.
 *
 * This is what the Provision screen's checklist renders, so that the button is only
 * pressed once the operator can see what it will do. Every check is a read.
 *
 * @param array{
 *     db_name:string,
 *     app_creds?:array<string,string>,
 *     registry_creds?:array<string,string>,
 *     webroot?:string,
 *     shared_api?:bool
 * } $opts
 * @return array<string,mixed>
 */
function tenant_status(array $opts): array
{
    $dbName = tenant_assert_valid_db_name((string) ($opts['db_name'] ?? ''));
    $appCreds = $opts['app_creds'] ?? tenant_app_credentials();
    $registryCreds = $opts['registry_creds'] ?? tenant_registry_credentials();
    $sharedApi = !empty($opts['shared_api']);

    $webroot = (string) ($opts['webroot'] ?? '');
    if ($webroot === '') {
        $webroot = $sharedApi
            ? dirname((string) tenant_server_config()['api_dir'])
            : dirname(__DIR__, 2);
    }

    $plan = tenant_plan(['db_name' => $dbName, 'webroot' => $webroot, 'shared_api' => $sharedApi]);

    $checks = [];
    $add = static function (string $name, bool $ok, string $detail) use (&$checks): void {
        $checks[] = ['name' => $name, 'ok' => $ok, 'detail' => $detail];
    };

    // 1. The registry row, which is what a client is called before it exists.
    $row = null;
    $registryError = null;
    try {
        $registry = tenant_pdo($registryCreds, $registryCreds['dbname']);
        $stmt = $registry->prepare('SELECT id, db_code, db_name, files_dir, js_dir, is_created FROM dbs WHERE db_name = ?');
        $stmt->execute([$dbName]);
        $found = $stmt->fetch(PDO::FETCH_ASSOC);
        if ($found) {
            $row = [
                'id' => (int) $found['id'],
                'db_code' => trim((string) ($found['db_code'] ?? '')),
                'files_dir' => tenant_normalise_dir((string) ($found['files_dir'] ?? '')),
                'js_dir' => tenant_normalise_dir((string) ($found['js_dir'] ?? '')),
                'is_created' => (int) ($found['is_created'] ?? 0),
            ];
            $add('registry row', true, 'row ' . $row['id'] . ', db_code ' . ($row['db_code'] === '' ? '(none)' : $row['db_code']));
        } else {
            $add('registry row', false, 'no row with db_name = ' . $dbName . ' - add it in Databases -> Add first');
        }
    } catch (TenantProvisionError $e) {
        $registryError = $e->getMessage();
        $add('registry row', false, 'could not read the registry: ' . $e->getMessage());
    }

    $plan['db_code'] = $row['db_code'] ?? null;

    // The recorded folders have to be the derived ones, or the tenant's uploads
    // would land somewhere the plan does not describe.
    if ($row !== null) {
        $dirsMatch = $row['js_dir'] === tenant_normalise_dir($plan['js_dir'])
            && $row['files_dir'] === tenant_normalise_dir($plan['files_dir']);
        $add(
            'recorded folders',
            $dirsMatch,
            $dirsMatch
                ? $plan['js_dir'] . ' and ' . $plan['files_dir']
                : sprintf('row says js_dir /%s and files_dir /%s; provisioning will set them to %s and %s', $row['js_dir'], $row['files_dir'], $plan['js_dir'], $plan['files_dir'])
        );
    }

    // 2. The database, and whether it looks like this app's schema.
    $databaseExists = false;
    $tableCount = 0;
    $version = null;
    $adminCount = null;
    try {
        $databaseExists = tenant_database_exists($appCreds, $dbName);
        $add('database', $databaseExists, $databaseExists ? $dbName . ' exists' : $dbName . ' does not exist yet');

        if ($databaseExists) {
            $pdo = tenant_pdo($appCreds, $dbName);
            $tableCount = (int) $pdo->query(
                'SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = \'BASE TABLE\''
            )->fetchColumn();
            $add('schema', $tableCount > 0, $tableCount . ' base table(s)');

            if (tenant_table_exists($pdo, 'versions')) {
                $version = $pdo->query('SELECT version FROM versions ORDER BY id DESC LIMIT 1')->fetchColumn();
                $add('app version', $version !== null && (int) $version === TENANT_APP_VERSION, 'versions.version = ' . var_export($version, true) . ' (this build is ' . TENANT_APP_VERSION . ')');
            }

            if (tenant_table_exists($pdo, 'users')) {
                $adminCount = (int) $pdo->query('SELECT COUNT(*) FROM users WHERE role_id = 1')->fetchColumn();
                $add('admin account', $adminCount > 0, $adminCount . ' admin(s)');
            }

            $dbsRows = tenant_table_exists($pdo, 'dbs')
                ? (int) $pdo->query('SELECT COUNT(*) FROM `dbs`')->fetchColumn()
                : 0;
            $add('self row', $dbsRows > 0, $dbsRows . ' row(s) in the tenant\'s own dbs table');
        }
    } catch (TenantProvisionError $e) {
        $add('database', false, $e->getMessage());
    }

    // 3. Folders on disk.
    $add('app folder', is_dir($plan['app_folder']), is_dir($plan['app_folder']) ? $plan['app_folder'] : $plan['app_folder'] . ' does not exist yet');
    // The tenant folder holds uploads and nothing else. Assets come from the one
    // shared dist/, so a build sitting in here is a leftover, not a missing step -
    // reported as a problem to clear, the other way round from how it used to read.
    $add('no build copy', !is_file($plan['app_folder'] . '/index.html'), is_file($plan['app_folder'] . '/index.html')
        ? $plan['app_folder'] . '/index.html is there - the build is shared now, remove this copy'
        : 'served from the shared build');
    $add('no db_code.json', !is_file($plan['app_folder'] . '/db_code.json'), is_file($plan['app_folder'] . '/db_code.json')
        ? $plan['app_folder'] . '/db_code.json is there - the folder name binds the database now, remove this file'
        : 'the tenant name is the binding');
    $add('no api copy', !is_dir($plan['app_folder'] . '/api'), is_dir($plan['app_folder'] . '/api')
        ? $plan['app_folder'] . '/api is there - the api is shared now, remove this copy'
        : 'served from the shared api');

    $filesFolder = is_dir($plan['files_folder']);
    $add('uploads folder', $filesFolder, $filesFolder ? $plan['files_folder'] : $plan['files_folder'] . ' does not exist yet');

    $missingSubdirs = [];
    foreach (TENANT_UPLOAD_SUBDIRS as $subdir) {
        if (!is_dir($plan['files_folder'] . '/' . $subdir)) {
            $missingSubdirs[] = $subdir;
        }
    }
    $add(
        'upload subdirectories',
        $filesFolder && $missingSubdirs === [],
        $missingSubdirs === []
            ? count(TENANT_UPLOAD_SUBDIRS) . ' present'
            : count($missingSubdirs) . ' missing (' . implode(', ', array_slice($missingSubdirs, 0, 5)) . (count($missingSubdirs) > 5 ? ', ...' : '') . ')'
    );

    // 4. How the request will be routed. With a shared api/ the tenant folder does
    // NOT hold an api/ - that is the point - so this check is about the shared one.
    if ($sharedApi) {
        $apiDir = (string) tenant_server_config()['api_dir'];
        $add('shared api/', is_dir($apiDir), $apiDir);
        $add('shared api/ is not inside the tenant folder', dirname($apiDir) !== $plan['app_folder'], $apiDir);
    } else {
        $add('api copy', is_dir($plan['app_folder'] . '/api'), $plan['app_folder'] . '/api');
    }

    // 5. Is a tenant of this name already deployed somewhere else? Two rows that differ
    // only in case are two folders and two databases on Linux, but one name to every
    // lookup the app makes - and the symptom, one client seeing another's data, is
    // impossible to trace back here. LOWER() on both sides rather than the column's
    // own collation, so the check does not depend on a server default.
    if ($row !== null && $registryError === null) {
        try {
            $clash = $registry->prepare(
                'SELECT db_name FROM dbs WHERE LOWER(db_name) = LOWER(?) AND id <> ?'
            );
            $clash->execute([$dbName, $row['id']]);
            $others = array_values(array_map('strval', $clash->fetchAll(PDO::FETCH_COLUMN)));
            $add('no case-only duplicate', $others === [], $others === [] ? 'none' : 'also registered as ' . implode(', ', $others));
        } catch (PDOException $e) {
            $add('no case-only duplicate', true, 'not checked: ' . $e->getMessage());
        }
    }

    $blocking = [];
    foreach ($checks as $check) {
        if (!$check['ok']) {
            $blocking[] = $check['name'];
        }
    }

    return [
        'db_name' => $dbName,
        'plan' => $plan,
        'registry_row' => $row,
        'database_exists' => $databaseExists,
        'table_count' => $tableCount,
        'version' => $version,
        'admin_count' => $adminCount,
        'checks' => $checks,
        'ready' => $blocking === [],
        'blocking' => $blocking,
    ];
}

// ---------------------------------------------------------------------------
// CLI: describe a plan without touching anything, and build the template
// ---------------------------------------------------------------------------

/**
 * Build the reference-only database that every client is created from.
 *
 * Without this the bootstrap is circular: provisioning copies a template, and the
 * template only exists if something already built one. The E2E run that first proved
 * the flow together had to write a throwaway PHP script to do it, which is exactly
 * the kind of step that gets skipped on the next server.
 *
 * --force is refused rather than offered quietly. The usual source of reference rows
 * is the development database, which also holds client data; re-running this against
 * it and calling the result a template is how a client's cars end up as another
 * client's dropdown options. Rebuild deliberately, into a new name.
 */
function tenant_build_template(array $opts): array
{
    $log = isset($opts['log']) && is_callable($opts['log'])
        ? $opts['log']
        : static function (string $line): void {
        };

    $appCreds = $opts['app_creds'] ?? tenant_app_credentials();
    $registryCreds = $opts['registry_creds'] ?? tenant_registry_credentials();
    $dbName = tenant_assert_valid_db_name((string) ($opts['db_name'] ?? ''));
    // Refused rather than defaulted to $appCreds['dbname'], for the same reason as in
    // tenant_provision() above: that name is a live tenant, not a template.
    $source = tenant_assert_valid_db_name((string) ($opts['seed_source'] ?? ''));

    if ($source === $dbName) {
        throw new TenantProvisionError(
            "refusing: the template would be built from itself ({$dbName}). Nothing would be copied."
        );
    }

    $existed = tenant_database_exists($appCreds, $dbName);
    if ($existed && empty($opts['force'])) {
        throw new TenantProvisionError(
            "refusing: {$dbName} already exists. Pass --force to rebuild it, and prefer a new name -"
            . ' --force over an existing template is how client data becomes reference data.'
        );
    }

    if ($existed) {
        $log("dropping the existing {$dbName}");
        tenant_mysql_run($appCreds, ['--drop'], "DROP DATABASE IF EXISTS `{$dbName}`", $log);
    }

    tenant_create_database($appCreds, $dbName, true, $log);

    // The target has to exist as a catalog for tenant_pdo() below, which is why the
    // schema is applied through the file-runner rather than through the target PDO.
    tenant_apply_schema($appCreds, $registryCreds, $dbName, $log);

    $target = tenant_pdo($appCreds, $dbName);
    $copied = tenant_seed_reference($target, $appCreds, $source, $log);
    $clean = tenant_assert_template_clean($appCreds, $dbName, $log);
    tenant_set_app_version($target, $log);

    // setup.sql ships the reference rows as INSERTs, so on this project the schema
    // step usually fills these tables and the seed step copies nothing. Reporting
    // only $copied would therefore say "0 reference tables" about a template that
    // has 39 permissions in it - true, and useless. Counted from the target instead,
    // so the summary describes the database rather than the path taken to build it.
    $populated = [];
    foreach (TENANT_SEED_TABLES as $table) {
        if (!tenant_table_exists($target, $table)) {
            continue;
        }
        $rows = (int) $target->query("SELECT COUNT(*) FROM `{$table}`")->fetchColumn();
        if ($rows > 0) {
            $populated[$table] = ['rows' => $rows, 'from' => isset($copied[$table]) ? $source : 'setup.sql'];
        }
    }

    $fromSource = count(array_filter($populated, static fn (array $p): bool => $p['from'] === $source));
    $fromSetup = count($populated) - $fromSource;

    $log(sprintf(
        'template %s ready: %d reference table(s) with rows (%d copied from %s, %d already in setup.sql)',
        $dbName,
        count($populated),
        $fromSource,
        $source,
        $fromSetup
    ));

    return [
        'db_name' => $dbName,
        'seed_source' => $source,
        'rebuilt' => $existed,
        'reference_tables' => $populated,
        'business_tables_verified_empty' => count(TENANT_BUSINESS_TABLES),
        'had_client_data' => $clean,
        'ok' => true,
    ];
}

if (PHP_SAPI === 'cli' && isset($argv[0]) && realpath($argv[0]) === realpath(__FILE__)) {
    // Loaded here, at file scope, because everything below reads the credentials out
    // of globals that only these two files set. Without this the CLI could never run:
    // the guard fires, the credential helpers throw "config.php has not been loaded",
    // and the only way to use either verb was to write a wrapper script first - which
    // is why the bootstrap had no documented command until now.
    require_once __DIR__ . '/../config.php';
    require_once __DIR__ . '/../db_manager_config.php';

    $args = array_slice($argv, 1);
    $flag = $args[0] ?? '';

    if ($flag === '--describe-plan' || $flag === '--build-template') {
        $dbName = $args[1] ?? '';
        $opts = ['db_name' => $dbName];

        foreach (array_slice($args, 2) as $i => $arg) {
            if ($arg === '--seed-source' && isset($args[$i + 3])) {
                $opts['seed_source'] = $args[$i + 3];
            }
            if ($arg === '--webroot' && isset($args[$i + 3])) {
                $opts['webroot'] = $args[$i + 3];
            }
            if ($arg === '--shared-api') {
                $opts['shared_api'] = true;
            }
            if ($arg === '--force') {
                $opts['force'] = true;
            }
        }

        try {
            if ($flag === '--build-template') {
                echo json_encode(
                    tenant_build_template($opts + ['log' => static function (string $line): void {
                        fwrite(STDOUT, $line . "\n");
                    }]),
                    JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES
                ), "\n";
                exit(0);
            }

            echo json_encode(tenant_plan($opts), JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES), "\n";
        } catch (TenantProvisionError $e) {
            fwrite(STDERR, $e->getMessage() . "\n");
            exit(1);
        }

        exit(0);
    }

    fwrite(STDERR, "usage:\n"
        . "  php api/lib/tenant-provision.php --describe-plan <db_name> [--seed-source <db>] [--webroot <dir>] [--shared-api]\n"
        . "  php api/lib/tenant-provision.php --build-template <db_name> [--seed-source <db>] [--force]\n");
    exit(1);
}
