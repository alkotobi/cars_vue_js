#!/usr/bin/env php
<?php
// Provision a second, independent local app on the Vite dev server, talking to its
// own database and uploading into its own <db>_files/ folder.
//
//   php deploy/setup-mig-27.php                    # database + folders + files + app copy
//   php deploy/setup-mig-27.php --sync-only        # only re-copy the build and api/
//   php deploy/setup-mig-27.php --force            # drop the database and rebuild it
//   php deploy/setup-mig-27.php --db acme_cars     # some other local tenant
//   php deploy/setup-mig-27.php --describe-plan    # print the plan, change nothing
//
// This is the local dev tenant's front end for api/lib/tenant-provision.php, which
// is where the work lives - the same library backs the DB manager's Provision button
// on the server, and having two copies of these steps is how they drift apart.
//
// Re-runnable: every step is idempotent, and the seed is skipped for any table that
// already has rows, so this doubles as "rebuild the tenant from scratch".
//
// Why this copies api/ rather than symlinking it
// -------------------------------------------------
// A tenant folder is a complete deployment: built assets, db_code.json and its own
// api/ (that is what deploy/deploy.sh rsyncs to a client server). The tenant's
// database is chosen by api/lib/appdb.php from db_code.json in the folder holding
// api/. PHP resolves symlinks in __DIR__, so a symlinked api/ would report the repo's
// api/ as its location and read the ROOT db_code.json, sending the whole tenant app
// to the root database. Hence a real copy, refreshed by --sync-only.
//
// On the server there is ONE shared /var/www/api/ instead, and the tenant's database
// comes from the request's mount - the two layouts are both handled by
// api/lib/appdb.php, and --shared-api here selects the server one.
//
// Credentials come from api/config.local.php and api/db_manager_config.local.php -
// never hardcoded, and passed to the mysql client via MYSQL_PWD so they do not land
// in the process list.

$repoRoot = dirname(__DIR__);

// Loaded at file scope, NOT inside a function: both files assign their array to a
// plain variable, and a `require` from inside a function would put it in that
// function's scope where `global` cannot see it (the same trap that made
// dbm_target_credentials() in db_manager_api.php come back empty).
require_once $repoRoot . '/api/config.php';
require_once $repoRoot . '/api/db_manager_config.php';
require_once $repoRoot . '/api/lib/tenant-provision.php';

/**
 * The tenant used when --db is omitted: this machine's own app, which is also the
 * database named in the root db_code.json. Overridable with --db, because a machine
 * can hold more than one and this is only a convenience for the common case.
 */
const DEFAULT_TENANT_DB = 'merhab_cars';

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

function step(string $message): void
{
    echo "\n\033[1m==> {$message}\033[0m\n";
}

function info(string $message): void
{
    echo "    {$message}\n";
}

/**
 * The library logs one line per action; this is where those lines go.
 */
function out(string $line): void
{
    echo '    ' . $line . "\n";
}

/**
 * The value after an option, or null when it is missing or is itself an option.
 *
 * @param list<string> $argv
 */
function arg_value(array $argv, int $i): ?string
{
    $value = $argv[$i + 1] ?? null;

    return ($value === null || str_starts_with($value, '--')) ? null : $value;
}

// ---------------------------------------------------------------------------
// Arguments
// ---------------------------------------------------------------------------

$dbName = DEFAULT_TENANT_DB;
$syncOnly = false;
$force = false;
$describePlan = false;
$sharedApi = false;
$seedSource = null;

for ($i = 1; $i < count($argv); $i++) {
    switch ($argv[$i]) {
        case '--sync-only':
            $syncOnly = true;
            break;
        case '--force':
            $force = true;
            break;
        case '--describe-plan':
            $describePlan = true;
            break;
        case '--shared-api':
            $sharedApi = true;
            break;
        case '--db':
            $value = arg_value($argv, $i);
            if ($value === null) {
                fwrite(STDERR, "error: --db needs a database name\n");
                exit(1);
            }
            $dbName = $value;
            $i++;
            break;
        case '--seed-source':
            $value = arg_value($argv, $i);
            if ($value === null) {
                fwrite(STDERR, "error: --seed-source needs a database name\n");
                exit(1);
            }
            $seedSource = $value;
            $i++;
            break;
        default:
            fwrite(STDERR, 'error: unknown argument: ' . $argv[$i] . "\n");
            fwrite(STDERR, "  try --db <name>, --sync-only, --force, --describe-plan, --shared-api, --seed-source <db>\n");
            exit(1);
    }
}

$appCreds = tenant_app_credentials();
$registryCreds = tenant_registry_credentials();

// Locally the deployment root is the repository: the tenant folder and its uploads
// are siblings of api/, which is what app_deployment_root() works out from the
// registry's leading slash on files_dir.
$webroot = $sharedApi ? dirname((string) tenant_server_config()['api_dir']) : $repoRoot;
// Where a new tenant's reference data comes from. --seed-source wins, then the
// template_database this machine is configured with, and only then the app database.
//
// The fallback matters less than it looks, because it used to be the only option and
// is wrong the moment this machine is itself a live tenant: the app database is then
// the tenant being provisioned, and tenant_plan() refuses it outright ("the seed
// source cannot be the tenant database itself"). Reading template_database first is
// what lets a plain invocation work on a machine whose own database is merhab_cars.
$seedSource ??= (string) (tenant_server_config()['template_database'] ?: '');
$seedSource = $seedSource !== '' ? $seedSource : $appCreds['dbname'];

try {
    $dbName = tenant_assert_valid_db_name($dbName);

    if ($describePlan) {
        // Read the row so the plan carries the real db_code, then print it. No
        // database, folder or file is touched.
        $row = tenant_find_row(tenant_pdo($registryCreds, $registryCreds['dbname']), $dbName);
        $plan = tenant_plan([
            'db_name' => $dbName,
            'db_code' => $row['db_code'],
            'webroot' => $webroot,
            'seed_source' => $seedSource,
            'shared_api' => $sharedApi,
        ]);

        echo json_encode($plan, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES), "\n";
        exit(0);
    }

    if ($syncOnly) {
        // A sync needs the row and the plan, but not the database: this is what
        // `npm run mig27` runs after every build.
        $status = tenant_status([
            'db_name' => $dbName,
            'app_creds' => $appCreds,
            'registry_creds' => $registryCreds,
            'webroot' => $webroot,
            'shared_api' => $sharedApi,
        ]);

        echo "Tenant {$dbName}: re-copying the build only\n";
        info('--sync-only: leaving the database and the folders alone');

        step('App copy');
        $buildSource = is_dir($repoRoot . '/dist') ? $repoRoot . '/dist' : '';
        tenant_deploy_app($status['plan'], $buildSource, 'out');

        echo "\n\033[32mDone.\033[0m Build refreshed for {$dbName}.\n\n";
        echo "  The tenant app is a prebuilt snapshot, so it has no HMR: run `npm run mig27`\n";
        echo "  after changing src/. Restart `npm run dev` if vite.config.js changed.\n";
        exit(0);
    }

    echo "Tenant {$dbName} (registry row read from " . $registryCreds['dbname'] . ")\n";
    info('app folder:  ' . $webroot . '/' . $dbName);
    info('uploads:     ' . $webroot . '/' . tenant_files_dir_name($dbName));
    info('api:         ' . $appCreds['host'] . ' as ' . $appCreds['user']);
    info('registry:    ' . $registryCreds['host'] . ' as ' . $registryCreds['user']);
    info('seed source: ' . $seedSource);
    info('layout:      ' . ($sharedApi ? 'one shared api/ (server layout)' : 'per-tenant api/ copy (local layout)'));

    $report = tenant_provision([
        'db_name' => $dbName,
        'app_creds' => $appCreds,
        'registry_creds' => $registryCreds,
        'webroot' => $webroot,
        'seed_source' => $seedSource,
        'force' => $force,
        'shared_api' => $sharedApi,
        // Local development seeds from this machine's own tenant by default, which
        // is not a template: it holds this developer's test data. The check exists
        // to stop a real template being handed to a real client, and on this machine
        // the source is chosen by hand every time, so it is skipped here - with the
        // names it would have to check printed instead.
        'skip_template_check' => true,
        'log' => 'out',
    ]);
} catch (TenantProvisionError $e) {
    fwrite(STDERR, "\nerror: " . $e->getMessage() . "\n");
    exit(1);
}

echo "\n\033[32mDone.\033[0m Independent apps, one dev server:\n\n";
echo '  root app    http://localhost:5173/cars                      -> ' . $appCreds['dbname'] . " (db_code from /db_code.json)\n";
echo "  tenant app  http://localhost:5173/" . ltrim((string) $report['js_dir'], '/') . "/cars   -> {$report['db_name']} (uploads " . basename($report['files_dir']) . "/)\n\n";
echo "  The tenant app is a prebuilt snapshot, so it has no HMR: run `npm run mig27`\n";
echo "  after changing src/. Restart `npm run dev` if vite.config.js changed.\n";
