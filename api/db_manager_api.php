<?php
// Start output buffering to prevent any output before headers
ob_start();

// Set CORS headers first, before any output.
// Same-origin only - see lib/cors.php for why the wildcard was removed.
require_once __DIR__ . '/lib/cors.php';
api_send_cors_headers();
header('Content-Type: application/json');

// Handle preflight OPTIONS request
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    http_response_code(200);
    ob_end_clean();
    exit;
}

// Response array (initialize early)
$response = [
    'success' => false,
    'message' => '',
    'data' => null
];

// ---------------------------------------------------------------------------
// DB-manager token helpers
// ---------------------------------------------------------------------------

/**
 * A valid bcrypt hash of a value nobody knows, used to keep a failed login slow
 * when the username does not exist. Same construction as AUTH_DUMMY_HASH in
 * api/actions/auth.php.
 */
const DBM_DUMMY_HASH = '$2y$10$usesomesillystringfore7hnbRJHxXVLeakoG8K30oukPsA.ztMG';

/**
 * The unauthenticated surface of this file, and nothing else.
 *
 * get_database_by_code is called by loadConfig() at boot, before anyone has
 * logged in, to turn this server's tenant config into a real database name and
 * files_dir. login and signup are the credential exchange itself.
 *
 * Everything else - run_sql, delete_database, backup_databases,
 * prepare_upload_folder, update_structure - requires a `login`.api_token.
 *
 * A literal rather than a flag on the gated path: a caller-sent "public" marker
 * would be a bypass, exactly as in api/api.php's PUBLIC_ACTIONS.
 */
const DBM_PUBLIC_ACTIONS = ['get_database_by_code', 'login', 'signup'];

//
// These authenticate against the registry's `login` table, deliberately not
// against the tenant app's `users` table that lib/auth.php reads. See the gate
// further down for why the two realms are kept apart.

/**
 * Resolve the caller from their DB-manager token.
 *
 * @return array{id:int,user:string,active:int}|null null when the token is
 *         missing, unknown, or names a deactivated account.
 */
function dbm_token_user(PDO $conn, string $token): ?array
{
    if ($token === '') {
        return null;
    }

    try {
        $stmt = $conn->prepare('SELECT id, user, active FROM login WHERE api_token = ? LIMIT 1');
        $stmt->execute([$token]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);
    } catch (PDOException $e) {
        // A registry predating api/migrations/031_login_api_token.sql has no such
        // column. That must read as "no token" rather than as a 500 that leaks SQL.
        error_log('dbm_token_user: ' . $e->getMessage());
        return null;
    }

    if (!$row || (int) $row['active'] !== 1) {
        return null;
    }

    return $row;
}

/** A fresh 32-byte token, hex encoded to fit login.api_token VARCHAR(64). */
function dbm_new_token(): string
{
    return bin2hex(random_bytes(32));
}

// Wrap everything in try-catch to ensure headers are always sent
try {
    // Include database configuration
    if (!file_exists(__DIR__ . '/db_manager_config.php')) {
        throw new Exception('db_manager_config.php file not found');
    }
    require_once __DIR__ . '/db_manager_config.php';    
    // Check if config is loaded
    if (!isset($db_manager_config)) {
        throw new Exception('Database configuration not loaded');
    }

    // The provisioning library: the folder-name rules every action below enforces, and
    // the steps behind the provision_tenant / deploy_app_to_tenant actions. Loaded
    // here rather than inside the actions so a missing file fails at request time with
    // a clear reason instead of as an undefined function halfway through a switch case.
    require_once __DIR__ . '/lib/tenant-provision.php';
    
    // Use config values
    $db_host = $db_manager_config['host'];
    $db_user = $db_manager_config['user'];
    $db_pass = $db_manager_config['pass'];
    $db_name = $db_manager_config['dbname'];
    // Get request method
    $method = $_SERVER['REQUEST_METHOD'];
    
    // Get POST/GET data
    //
    // The `??` fallback only rescues a *parse failure* on an array-shaped body. A
    // body of `123` or `"x"` decodes to an int/string, which is not null, so it
    // survives the coalesce and every `is_array($inputData)`-style guard further
    // down is silently skipped while `$inputData['action']` reads as offset 123.
    // Reject anything that is not an object-shaped payload instead of guessing.
    $inputData = [];
    if ($method === 'POST') {
        $rawInput = file_get_contents('php://input');
        $decoded = json_decode((string) $rawInput, true);
        $inputData = is_array($decoded) ? $decoded : $_POST;
    } elseif ($method === 'GET') {
        $inputData = $_GET;
    }
    
    // Establish database connection to merhab_databases
    $conn = new PDO(
        "mysql:host={$db_host};dbname={$db_name}", 
        $db_user, 
        $db_pass
    );
    $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    // Resolve the on-disk path of tenant config for a database row.
    //
    // tenant config is per server: it names the database this deployment talks
    // to, so it is written on the server rather than shipped in the build (see
    // deploy/deploy.sh). It lives in the app folder, which is where the app
    // fetches it from (<mount>tenant config). An empty js_dir therefore means
    // "the app root", not "unconfigured" — that is the default the deploy guide
    // seeds, and rejecting it made this file uneditable in production.
    // Returns null when the resolved path would escape the app root.
    function resolveDbCodeJsonPath($jsDir) {
        $base = realpath(__DIR__ . '/..');
        if ($base === false) {
            return null;
        }
        $base = rtrim($base, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;

        $jsDir = trim((string)$jsDir);
        $jsDir = str_replace('\\', '/', $jsDir);
        $jsDir = trim($jsDir, '/');

        $path = $base . ($jsDir !== '' ? $jsDir . '/' : '') . 'tenant config';

        // Defence in depth: never let a crafted js_dir write outside the app.
        $real = realpath($path);
        if ($real !== false && strpos($real, $base) !== 0) {
            return null;
        }
        if (strpos($path, $base) !== 0 || strpos($path, '..') !== false) {
            return null;
        }
        return $path;
    }

    // Helper function to format Unix directory path
    function formatUnixPath($path) {
        if (empty($path)) {
            return null;
        }
        // Remove any backslashes and replace with forward slashes
        $path = str_replace('\\', '/', $path);
        // Remove any double slashes
        $path = preg_replace('#/+#', '/', $path);
        // Ensure it starts with /
        if (substr($path, 0, 1) !== '/') {
            $path = '/' . $path;
        }
        // Remove trailing slash (unless it's root)
        if (strlen($path) > 1 && substr($path, -1) === '/') {
            $path = rtrim($path, '/');
        }
        return $path;
    }
    
    // Whether this installation has ONE shared api/ or a copy per tenant.
    //
    // Both layouts are supported (api/lib/appdb.php resolves a tenant from the request
    // in the first and from its own location in the second), and which one this is
    // decides whether provisioning writes an api/ into each tenant folder.
    //
    // The answer comes from tenant_has_shared_api(), which probes the filesystem and
    // honours an override in /etc/cars-deploy.json. It used to be worked out here by
    // comparing api_dir against this file's parent, which was wrong for the layout
    // this server actually runs: for a shared /var/www/api it compared /var/www/api
    // with /var/www, said "not shared", and gave every client its own 68-file copy of
    // the API - quietly reverting the one-copy design on the one machine where it
    // matters, and with nothing in the logs to say so.
    function dbm_has_shared_api() {
        return function_exists('tenant_has_shared_api') ? tenant_has_shared_api() : false;
    }

    // Whether this client has anything on disk or in the database yet.
    //
    // The line between "a row an operator just added" and "a live client" is
    // exactly where the operations that cannot be undone belong: provisioning,
    // renaming, deleting. An empty row is a draft and can be edited freely.
    function dbm_is_provisioned(PDO $conn, string $dbName) {
        $dirs = dbm_tenant_dirs($dbName);
        $webroot = dbm_deployment_root();

        if (is_dir($webroot . '/' . $dirs['app_folder']) || is_dir($webroot . '/' . $dirs['files_folder'])) {
            return true;
        }

        $stmt = $conn->prepare('SELECT is_created FROM dbs WHERE db_name = ?');
        $stmt->execute([$dbName]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);

        return $row !== false && (int) $row['is_created'] === 1;
    }

    // The database name a provisioning request is for.
    //
    // Accepts the name or the row id, because the screen has both and the id is what
    // a list row carries.
    function dbm_requested_db_name(PDO $conn, array $inputData) {
        $name = trim((string) ($inputData['db_name'] ?? ''));
        if ($name !== '') {
            return $name;
        }

        $id = (int) ($inputData['id'] ?? 0);
        if ($id <= 0) {
            return '';
        }

        $stmt = $conn->prepare('SELECT db_name FROM dbs WHERE id = ?');
        $stmt->execute([$id]);
        $found = $stmt->fetchColumn();

        return $found === false ? '' : (string) $found;
    }

    // Reject database names that cannot be used safely.
    //
    // The name is chosen by whoever fills in the create form and ends up in three
    // places that are not parameterisable: `CREATE DATABASE`, the PDO DSN, and - the
    // reason the accepted set is now so narrow - the tenant's folders and URL.
    //
    //   * The DSN is semicolon-delimited, so a name holding a `;` would silently
    //     truncate into a different dbname, and a backtick would end the quoted
    //     identifier in the DDL.
    //   * dbm_tenant_dirs() derives an nginx location and a folder name from this
    //     string. A space or a `$` is legal in MySQL and impossible to use in an
    //     unquoted nginx path, so accepting them here only moved the failure to
    //     provisioning time, where it looks like a server problem.
    //
    // So the accepted set is the one the provisioning library enforces
    // (api/lib/tenant-provision.php), which is also what api/lib/appdb.php will
    // resolve at runtime. Delegating to it means the screen, the provisioning and the
    // request handler cannot disagree about what a legal client name is - a name the
    // screen accepts and appdb.php refuses provisions a client whose own site cannot
    // connect to its database.
    function dbm_validate_database_name(string $dbName) {
        if (!function_exists('tenant_assert_valid_db_name')) {
            return ['error' => 'Tenant provisioning library is missing (api/lib/tenant-provision.php).'];
        }

        try {
            tenant_assert_valid_db_name($dbName);
        } catch (Throwable $e) {
            return ['error' => $e->getMessage()];
        }

        return [];
    }

    // The folders and URL prefix a client database gets, derived from its name.
    //
    // Derived rather than typed into the form, which is the whole point: a row whose
    // js_dir and db_name disagree is a tenant whose app is served from one place and
    // whose uploads are written to another, and nothing on the screen shows that. The
    // form still sends them, so a hand-edited value that disagrees is reported rather
    // than silently overwritten.
    //
    // @return array{js_dir:string,files_dir:string,app_folder:string,files_folder:string}
    function dbm_tenant_dirs(string $dbName) {
        return [
            'js_dir' => '/' . $dbName,
            'files_dir' => '/' . tenant_files_dir_name($dbName),
            'app_folder' => basename('/' . $dbName),
            'files_folder' => basename('/' . tenant_files_dir_name($dbName)),
        ];
    }

    // Reject a name that collides with another row under a case-insensitive
    // comparison, and reject a folder that already belongs to a different database.
    //
    // Both halves matter on Linux, where `Acme_Cars` and `acme_cars` are two folders
    // and two databases, but MySQL's default collation treats their names as equal in
    // a WHERE clause and the registry is read that way. Two rows that a lookup cannot
    // tell apart resolve to whichever one the server happens to return first, and the
    // symptom is one client seeing another's data.
    //
    // @return string|null an error message, or null when the name is free
    function dbm_assert_name_available(PDO $conn, string $dbName, ?int $exceptId = null) {
        // LOWER() on both sides, not the column's own collation and not BINARY.
        //
        // The column collation already compares case-insensitively, which is the
        // behaviour being guarded against, so relying on it would make this check
        // depend on a server default nobody set deliberately. BINARY would be the
        // opposite mistake: it matches only byte-identical names, so `MIG_27` and
        // `mig_27` would both be accepted - which is exactly the pair that has to be
        // refused.
        $stmt = $conn->prepare('SELECT id, db_name FROM dbs WHERE LOWER(db_name) = LOWER(?)' . ($exceptId !== null ? ' AND id <> ?' : ''));
        $params = $exceptId !== null ? [$dbName, $exceptId] : [$dbName];
        $stmt->execute($params);

        foreach ($stmt->fetchAll(PDO::FETCH_ASSOC) as $row) {
            $other = (string) $row['db_name'];
            return sprintf(
                "A database called '%s' is already registered (row %d). '%s' and '%s' are the same name on a case-insensitive filesystem lookup, so only one of them can exist.",
                $other,
                (int) $row['id'],
                $other,
                $dbName
            );
        }

        // A folder of that name belonging to a different database is the same problem
        // seen from the disk side.
        $webroot = dbm_deployment_root();
        foreach (dbm_tenant_dirs($dbName) as $key => $folder) {
            if ($key !== 'app_folder' && $key !== 'files_folder') {
                continue;
            }
            if (!is_dir($webroot . '/' . $folder)) {
                continue;
            }

            $owner = dbm_folder_owner($webroot . '/' . $folder);
            if ($owner !== null && strcasecmp($owner, $dbName) !== 0) {
                return sprintf(
                    "The folder %s/%s already belongs to the database '%s'. Delete or rename it before registering '%s'.",
                    basename($webroot),
                    $folder,
                    $owner,
                    $dbName
                );
            }
        }

        return null;
    }

    // Which registered database owns a folder, by reading the tenant config in it.
    //
    // The app folder names its database in tenant config, so that file is the record
    // of who a folder belongs to. Returns null when the folder holds no tenant config
    // or names a database that is not registered.
    function dbm_folder_owner(string $folder) {
        $file = rtrim($folder, '/') . '/tenant config';
        if (!is_file($file)) {
            return null;
        }

        $decoded = json_decode((string) @file_get_contents($file), true);
        $code = is_array($decoded) ? trim((string) ($decoded['db_code'] ?? '')) : '';
        if ($code === '') {
            return null;
        }

        global $conn;
        $stmt = $conn->prepare('SELECT db_name FROM dbs WHERE db_code = ?');
        $stmt->execute([$code]);
        $owner = $stmt->fetchColumn();

        return $owner === false ? null : (string) $owner;
    }

    // Where tenant folders live: the parent of the folder holding api/, when api/ is
    // shared by every tenant, and the app folder itself otherwise.
    //
    // With one shared api/ a tenant's app folder and its upload folder are siblings of
    // api/, which is what app_deployment_root() in api/lib/appdb.php assumes too.
    // Without it - a per-tenant copy of api/, the local layout - they sit inside the
    // app folder.
    function dbm_deployment_root() {
        $apiRoot = realpath(__DIR__ . '/..');

        if ($apiRoot === false) {
            return __DIR__ . '/..';
        }

        $serverConfig = function_exists('tenant_server_config') ? tenant_server_config() : [];
        $sharedApiDir = isset($serverConfig['api_dir']) ? (string) $serverConfig['api_dir'] : '';
        $isShared = $sharedApiDir !== '' && realpath($sharedApiDir) === $apiRoot;

        return $isShared ? dirname($apiRoot) : $apiRoot;
    }

    // Open a connection to the database named in the registry, creating it first if it
    // is not there yet.
    //
    // Returns ['conn' => PDO] on success, or ['error' => string] with a message meant
    // for the person clicking the button rather than for a log file.
    //
    // db_name is free text from the create form, so pressing Create on a name that has
    // never existed used to fail with the raw driver string
    // "SQLSTATE[HY000] [1049] Unknown database 'mig_27'" - which named neither what
    // went wrong nor what to do about it. Creating the database is what the button is
    // for, so do that instead, and keep the raw driver text in the log rather than
    // returning it to a browser.
    function dbm_connect_target(string $dbName) {
        $nameCheck = dbm_validate_database_name($dbName);
        if (isset($nameCheck['error'])) {
            return ['error' => $nameCheck['error']];
        }

        try {
            $config = dbm_target_credentials();
        } catch (Throwable $e) {
            error_log('dbm_connect_target: no API credentials: ' . $e->getMessage());
            return ['error' => 'The API database credentials are not configured on this server.'];
        }

        try {
            $serverConn = new PDO("mysql:host={$config['host']}", $config['user'], $config['pass']);
            $serverConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

            // IF NOT EXISTS keeps this idempotent: pressing Create on a database that is
            // already there - which is the normal case for re-running against an existing
            // tenant - must not fail or reset it. The name is backtick-quoted as well as
            // validated, so nothing here can terminate the identifier early.
            $serverConn->exec(sprintf('CREATE DATABASE IF NOT EXISTS `%s`', str_replace('`', '``', $dbName)));
        } catch (PDOException $e) {
            error_log("dbm_connect_target: cannot create '$dbName': " . $e->getMessage());
            return ['error' => sprintf(
                "Could not create database '%s' using the API's credentials. Check that the MySQL user is allowed to create databases.",
                $dbName
            )];
        }

        try {
            $conn = new PDO(
                "mysql:host={$config['host']};dbname={$dbName}",
                $config['user'],
                $config['pass']
            );
            $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
            return ['conn' => $conn];
        } catch (PDOException $e) {
            error_log("dbm_connect_target: cannot open '$dbName': " . $e->getMessage());
            return ['error' => sprintf(
                "Created database '%s' but could not open it. Check the database name and the API's credentials.",
                $dbName
            )];
        }
    }

    // Credentials used to reach tenant databases: the API's own, since the DB manager
    // operates across every tenant rather than as any one of them.
    function dbm_target_credentials() {
        // A plain `require`, not `require_once`, and the global checked first.
        //
        // config.php assigns `$db_config` in the scope that includes it. When another
        // action has already loaded it at top level, that lands in $GLOBALS and
        // require_once is a no-op. But create_tables can be the first thing to need it,
        // and then the include happens *inside this function* - putting $db_config in
        // function scope, where `global $db_config` finds nothing. That produced
        // "Access denied for user ''@'localhost'" and turned every target database into
        // "Could not verify the database list". config.php only builds an array from the
        // environment, so re-running it is harmless.
        if (isset($GLOBALS['db_config']) && is_array($GLOBALS['db_config'])) {
            return $GLOBALS['db_config'];
        }

        require __DIR__ . '/config.php';
        return $db_config;
    }

    // Which statements from setup.sql are worth sending to the server.
    //
    // CREATE TABLE and INSERT are the tables and their seed rows. DROP TRIGGER and
    // CREATE TRIGGER are the two triggers at the end of the file, which the Create button
    // used to skip entirely: they sit inside `DELIMITER $$` blocks, so the statement
    // splitter shredded them and then dropped what was left. DROP has to come first so
    // that re-running Create against a tenant that already has the triggers succeeds
    // instead of failing with "trigger already exists".
    //
    // ALTER matters just as much. setup.sql carries one, and it is the only thing that
    // puts a foreign key on buy_details.id_car_name - the file has to add it after the
    // fact because buy_details is created before the cars_names it points at. With ALTER
    // missing from this list the statement was filtered out silently: the database
    // reported success, and the only symptom was a constraint that was never there.
    function dbm_is_setup_statement($stmt) {
        foreach (['CREATE TABLE', 'CREATE TRIGGER', 'DROP TRIGGER', 'ALTER TABLE', 'INSERT'] as $keyword) {
            if (stripos($stmt, $keyword) !== false) {
                return true;
            }
        }
        return false;
    }

    // Split a .sql file into executable statements, semicolons included.
    //
    // The previous version stripped comments with preg_replace, collapsed all whitespace,
    // and then did explode(';'). That is only safe for SQL without interesting string
    // literals, and setup.sql has several:
    //
    //   COMMENT 'Token sent by the db-manager UI; checked server-side in ...'
    //   COMMENT 'MIME type (image/* or video/*)'
    //
    // The first cuts a statement in half at the semicolon, so the second half arrives as
    // its own "statement" starting mid-quote - which is where
    // "You have an error in your SQL syntax ... near ''Token sent by the d" came from.
    // The second loses text to the block-comment stripper, leaving an unbalanced quote.
    // Both fail silently in the sense that the other 68 tables still get created, so the
    // Create button reports partial success.
    //
    // So this walks the file once, tracking whether it is inside a single-quoted string
    // (honouring \' escapes and '' doubling), a backtick identifier, a -- line comment,
    // or a /* */ block comment, and only treats a semicolon as a separator at depth zero.
    // String literals are passed through verbatim so the column comments stay intact.
    function dbm_extract_sql_statements($sql) {
        $statements = [];
        $current = '';
        $len = strlen($sql);
        $inString = false;
        $inIdentifier = false;
        $inLineComment = false;
        $inBlockComment = false;
        $delimiter = ';';

        for ($i = 0; $i < $len; $i++) {
            $char = $sql[$i];
            $next = $i + 1 < $len ? $sql[$i + 1] : '';

            if ($inLineComment) {
                if ($char === "\n") {
                    $inLineComment = false;
                    $current .= $char;
                }
                continue;
            }

            if ($inBlockComment) {
                if ($char === '*' && $next === '/') {
                    $inBlockComment = false;
                    $i++;
                } elseif ($char === "\n") {
                    // Keep newlines so line numbers in MySQL errors stay meaningful.
                    $current .= $char;
                }
                continue;
            }

            if ($inString) {
                $current .= $char;
                if ($char === '\\' && $next !== '') {
                    // Backslash escape: consume the escaped character verbatim.
                    $current .= $next;
                    $i++;
                } elseif ($char === "'") {
                    if ($next === "'") {
                        // '' is an escaped quote, not the end of the literal.
                        $current .= $next;
                        $i++;
                    } else {
                        $inString = false;
                    }
                }
                continue;
            }

            if ($inIdentifier) {
                $current .= $char;
                if ($char === '`') {
                    $inIdentifier = false;
                }
                continue;
            }

            // Not inside a string, identifier, or comment.
            if ($char === "'") {
                $inString = true;
                $current .= $char;
                continue;
            }

            if ($char === '`') {
                $inIdentifier = true;
                $current .= $char;
                continue;
            }

            if ($char === '-' && $next === '-') {
                $inLineComment = true;
                $i++;
                continue;
            }

            if ($char === '/' && $next === '*') {
                $inBlockComment = true;
                $i++;
                continue;
            }

            // `DELIMITER $$` is a mysql CLI directive, not SQL, and PDO cannot run it.
            // setup.sql uses it to fence the BEGIN...END bodies of the two triggers: the
            // semicolons inside those bodies are statement syntax, not separators. Honour
            // the directive by switching what counts as a separator, and drop the line
            // itself so it is never sent to the server.
            if (($i === 0 || $sql[$i - 1] === "\n") && substr($sql, $i, 9) === 'DELIMITER' && trim($current) === '') {
                $eol = strpos($sql, "\n", $i);
                $line = $eol === false ? substr($sql, $i) : substr($sql, $i, $eol - $i);
                if (preg_match('/^DELIMITER\s+(\S+)\s*$/i', trim($line), $dm)) {
                    $delimiter = $dm[1];
                    $i = $eol === false ? $len : $eol;
                    continue;
                }
            }

            $delimLen = strlen($delimiter);
            if ($delimLen > 0 && substr($sql, $i, $delimLen) === $delimiter) {
                $stmt = trim($current);
                if ($stmt !== '' && dbm_is_setup_statement($stmt)) {
                    $statements[] = $stmt . ';';
                }
                $current = '';
                $i += $delimLen - 1;
                continue;
            }

            $current .= $char;
        }

        // Trailing statement with no terminating semicolon.
        $stmt = trim($current);
        if ($stmt !== '' && (dbm_is_setup_statement($stmt))) {
            $statements[] = $stmt . ';';
        }

        return $statements;
    }

    // Helper function to validate Unix directory path
    function isValidUnixPath($path) {
        if (empty($path)) {
            return false;
        }
        // Must start with /
        if (substr($path, 0, 1) !== '/') {
            return false;
        }
        // Must not contain invalid characters (Windows drive letters, etc.)
        if (preg_match('/^[a-zA-Z]:/', $path)) {
            return false;
        }
        // Must not contain null bytes or other control characters
        if (strpos($path, "\0") !== false) {
            return false;
        }
        // Should only contain valid Unix path characters (letters, numbers, dots, underscores, hyphens, forward slashes, and spaces)
        // Note: spaces are technically valid but not recommended
        // Using explicit space character instead of \s to prevent control characters (newlines, tabs, etc.)
        if (!preg_match('/^[\/a-zA-Z0-9._\- ]+$/', $path)) {
            return false;
        }
        return true;
    }
    
    /**
     * Build SQL dump body for one database (no FOREIGN_KEY_CHECKS wrapper).
     */
    function buildDatabaseBackupBody(PDO $pdo) {
        $lines = [];
        $stmt = $pdo->query('SHOW TABLES');
        $tables = $stmt->fetchAll(PDO::FETCH_COLUMN);
        
        foreach ($tables as $table) {
            $stmt = $pdo->query("SHOW CREATE TABLE `$table`");
            $row = $stmt->fetch(PDO::FETCH_NUM);
            if (!$row || empty($row[1])) {
                continue;
            }
            
            $lines[] = "DROP TABLE IF EXISTS `$table`;";
            $lines[] = $row[1] . ';';
            $lines[] = '';
            
            $dataStmt = $pdo->query("SELECT * FROM `$table`");
            $rows = $dataStmt->fetchAll(PDO::FETCH_ASSOC);
            if (empty($rows)) {
                continue;
            }
            
            $columns = array_keys($rows[0]);
            $colList = '`' . implode('`, `', $columns) . '`';
            
            foreach (array_chunk($rows, 100) as $chunk) {
                $valueSets = [];
                foreach ($chunk as $dataRow) {
                    $vals = [];
                    foreach ($columns as $col) {
                        $val = $dataRow[$col];
                        $vals[] = $val === null ? 'NULL' : $pdo->quote($val);
                    }
                    $valueSets[] = '(' . implode(', ', $vals) . ')';
                }
                $lines[] = "INSERT INTO `$table` ($colList) VALUES\n" . implode(",\n", $valueSets) . ';';
                $lines[] = '';
            }
        }
        
        return implode("\n", $lines);
    }
    
    /**
     * Wrap dump body with header and FOREIGN_KEY_CHECKS off/on.
     */
    function wrapDatabaseBackupSql($dbname, $body) {
        $lines = [
            '-- Merhab Cars Database Backup',
            '-- Generated on: ' . date('Y-m-d H:i:s'),
            '-- Database: ' . $dbname,
            '',
            'SET NAMES utf8mb4;',
            'SET FOREIGN_KEY_CHECKS=0;',
            '',
            $body,
            '',
            'SET FOREIGN_KEY_CHECKS=1;',
            '',
        ];
        return implode("\n", $lines);
    }
    
    /**
     * Send SQL/ZIP backup as file download and exit (skip JSON response).
     */
    function sendBackupFileDownload($content, $filename, $contentType = 'application/sql') {
        if (ob_get_level()) {
            ob_end_clean();
        }
        api_send_cors_headers();
        header('Content-Type: ' . $contentType);
        header('Content-Disposition: attachment; filename="' . $filename . '"');
        header('Cache-Control: no-cache, must-revalidate');
        echo $content;
        exit;
    }
    
    // Handle different actions
    $action = $inputData['action'] ?? '';

    // Authentication gate.
    //
    // This file used to have none. Its login/signup actions never gated anything,
    // so every action below was reachable by an anonymous POST - including run_sql
    // (arbitrary SQL against every tenant database), update_structure (executes
    // stored SQL), backup_databases (full dumps), delete_database and
    // prepare_upload_folder (recursive delete of an attacker-chosen directory).
    // An attacker only had to chain create_database with run_sql to read the users
    // table, and from there every password hash and api_token.
    //
    // The credential is this realm's own `login` row, not the tenant app's
    // `users.api_token`. That is not a detail: this table lives in the registry
    // next to `dbs` and guards host-level operations with host-level credentials,
    // so it is deliberately a separate trust domain from the app.
    //
    // It was briefly gated on the *app* admin token instead, which is a bootstrap
    // deadlock rather than a security posture: `login`/`signup` were inside the
    // gate, and those are the only way to obtain the credential the UI checks for
    // (localStorage `db_manager_user`) before it renders anything at all. Every
    // call site then failed closed with an empty token, so the whole DB manager
    // was unreachable for everyone, admins included. See api/migrations/031 for
    // the `login`.api_token column this reads.
    //
    // Public actions:
    //   get_database_by_code - loadConfig() calls it at boot, before anyone has
    //     logged in, to turn this server's tenant config into a real database name
    //     and files_dir. It reveals one registry row and nothing else.
    //   login / signup      - the credential exchange itself. Public by necessity.
    //     migrations/030_drop_adv_sql.sql removed an identical execute_sql action
    //     for exactly this reason; this file reintroduced it under a new name.
    //
    // signup being open is not an escalation path: it inserts with `active = 0`
    // (see the case below), so a new account cannot do anything until someone
    // activates it in the registry by hand.
    if (!in_array($action, DBM_PUBLIC_ACTIONS, true)) {
        $token = isset($inputData['token']) && is_string($inputData['token'])
            ? trim($inputData['token'])
            : '';

        $identity = dbm_token_user($conn, $token);
        if ($identity === null) {
            // Deliberately the same envelope this file already uses, with `message`
            // populated. api_auth_fail() in lib/auth.php emits {success, code,
            // error} instead, and every caller in src/components/db-manager reads
            // result.message - so borrowing that helper made a dead API look like a
            // generic "Login failed" with no indication of why.
            if (ob_get_level()) {
                ob_end_clean();
            }
            if (!headers_sent()) {
                header('Content-Type: application/json');
            }
            echo json_encode(['success' => false, 'message' => 'Not authenticated. Sign in to the database manager.']);
            exit;
        }
    }

    switch ($action) {
        case 'test':
            // Test connection
            $response['success'] = true;
            $response['message'] = 'Database connection successful';
            $response['data'] = [
                'host' => $db_host,
                'database' => $db_name
            ];
            break;
            
        case 'signup':
            // Sign up new user
            $user = trim($inputData['user'] ?? '');
            $pass = $inputData['pass'] ?? '';
            
            if (empty($user) || empty($pass)) {
                $response['message'] = 'Username and password are required';
                break;
            }
            
            if (strlen($pass) < 6) {
                $response['message'] = 'Password must be at least 6 characters long';
                break;
            }
            
            // Check if user already exists
            $checkStmt = $conn->prepare("SELECT id FROM login WHERE user = ?");
            $checkStmt->execute([$user]);
            if ($checkStmt->fetch()) {
                $response['message'] = 'Username already exists';
                break;
            }
            
            // Hash password
            $hashedPassword = password_hash($pass, PASSWORD_DEFAULT);
            
            // Insert new user (active = 0, will be activated manually later)
            //
            // This is why signup can sit outside the auth gate without being an
            // escalation path: the row it creates holds no api_token, and
            // dbm_token_user() rejects active != 1 before anything else looks at
            // it. Activation is a manual edit in the registry, which is the point -
            // this screen cannot grant itself access to host-level credentials.
            $insertStmt = $conn->prepare("INSERT INTO login (user, pass, active) VALUES (?, ?, 0)");
            if ($insertStmt->execute([$user, $hashedPassword])) {
                $response['success'] = true;
                $response['message'] = 'User created successfully';
                $response['data'] = [
                    'id' => $conn->lastInsertId(),
                    'user' => $user
                ];
            } else {
                $response['message'] = 'Failed to create user';
            }
            break;
            
        case 'login':
            // Log in and mint this realm's session token.
            //
            // The token is the whole reason this case exists in its current shape:
            // it is the credential every other action in this file is now gated on,
            // and it is stored client-side as localStorage `db_manager_user`, which
            // LoginSignup.vue writes wholesale from `data` - so adding `token` here
            // is all the UI needs to start presenting it.
            //
            // A new token replaces any previous one, matching handle_login() in
            // api/actions/auth.php: one live token per user, and a second sign-in
            // quietly invalidates the first.
            $user = trim($inputData['user'] ?? '');
            $pass = $inputData['pass'] ?? '';

            if (empty($user) || empty($pass)) {
                $response['message'] = 'Username and password are required';
                break;
            }

            // Get user from database
            $stmt = $conn->prepare("SELECT id, user, pass, active FROM login WHERE user = ?");
            $stmt->execute([$user]);
            $userData = $stmt->fetch(PDO::FETCH_ASSOC);

            // Verify against a dummy hash when the username is unknown, so a missing
            // user and a wrong password take the same time.
            $hash = $userData['pass'] ?? DBM_DUMMY_HASH;
            $valid = password_verify((string) $pass, $hash);

            // One message for unknown-user, deactivated and wrong-password alike.
            // These were three distinct replies, which made this a free oracle for
            // "does this username exist and is it enabled?" against a table that
            // guards host-level credentials.
            if (!$userData || !$valid || (int) $userData['active'] !== 1) {
                $response['message'] = 'Invalid username or password';
                break;
            }

            $token = dbm_new_token();
            $issueStmt = $conn->prepare('UPDATE login SET api_token = ? WHERE id = ?');
            $issueStmt->execute([$token, $userData['id']]);

            $response['success'] = true;
            $response['message'] = 'Login successful';
            $response['data'] = [
                'id' => $userData['id'],
                'user' => $userData['user'],
                'token' => $token
            ];
            break;

        case 'logout':
            // Drop the caller's token server-side.
            //
            // DbManagerSidebar.vue used to clear localStorage and nothing else, which
            // left the token live server-side until the next sign-in overwrote it -
            // a shared machine kept a working credential nobody could see. A missing
            // or unknown token is still a success: the caller's intent is to hold
            // no token, which is already true.
            $logoutToken = isset($inputData['token']) && is_string($inputData['token'])
                ? trim($inputData['token'])
                : '';
            if ($logoutToken !== '') {
                $logoutStmt = $conn->prepare('UPDATE login SET api_token = NULL WHERE api_token = ?');
                $logoutStmt->execute([$logoutToken]);
            }

            $response['success'] = true;
            $response['message'] = 'Logged out';
            break;
            
        case 'get_databases':
            // Get all databases
            $stmt = $conn->prepare("SELECT * FROM dbs ORDER BY id DESC");
            $stmt->execute();
            $databases = $stmt->fetchAll(PDO::FETCH_ASSOC);
            
            // Fetch version for each created database
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            foreach ($databases as &$db) {
                $db['version'] = null;
                // Only fetch version if database is created
                if ($db['is_created'] == 1 && !empty($db['db_name'])) {
                    try {
                        $targetConn = new PDO(
                            "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                            $targetDbUser,
                            $targetDbPass
                        );
                        $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                        
                        $versionStmt = $targetConn->prepare("SELECT version FROM versions WHERE id = 1");
                        $versionStmt->execute();
                        $versionData = $versionStmt->fetch(PDO::FETCH_ASSOC);
                        if ($versionData && isset($versionData['version'])) {
                            $db['version'] = intval($versionData['version']);
                        }
                    } catch (PDOException $e) {
                        // If version fetch fails, leave it as null
                        error_log('Failed to fetch version for database ' . $db['db_name'] . ': ' . $e->getMessage());
                    }
                }
            }
            unset($db); // Break reference
            
            $response['success'] = true;
            $response['data'] = $databases;
            break;
            
        case 'get_database_by_code':
            // Get database by db_code
            $db_code = trim($inputData['db_code'] ?? '');
            
            if (empty($db_code)) {
                $response['message'] = 'DB Code is required';
                break;
            }
            
            $stmt = $conn->prepare("SELECT db_name, files_dir, js_dir FROM dbs WHERE db_code = ?");
            $stmt->execute([$db_code]);
            $database = $stmt->fetch(PDO::FETCH_ASSOC);
            
            if (!$database) {
                $response['message'] = 'Database not found';
                break;
            }
            
            $response['success'] = true;
            $response['data'] = $database;
            break;
            
        case 'create_database':
            // Create new database
            $db_name = trim($inputData['db_name'] ?? '');
            
            if (empty($db_name)) {
                $response['message'] = 'DB Name is required';
                break;
            }

            // The name becomes a folder, a URL prefix and an nginx location, so it is
            // held to the same rule the deploy scripts use.
            $nameCheck = dbm_validate_database_name($db_name);
            if (isset($nameCheck['error'])) {
                $response['message'] = $nameCheck['error'];
                break;
            }

            // Two rows that differ only in case are two folders and two databases on
            // Linux but one name to every lookup the app makes.
            $nameTaken = dbm_assert_name_available($conn, $db_name);
            if ($nameTaken !== null) {
                $response['message'] = $nameTaken;
                break;
            }
            
            // Auto-generate db_code from db_name + server timestamp
            $timestamp = time();
            $hashInput = $db_name . '_' . $timestamp;
            // Use first 20 characters of MD5 hash for better uniqueness
            $db_code = 'db_' . substr(md5($hashInput), 0, 20);
            
            // Ensure uniqueness - if code exists, regenerate with microtime
            $checkStmt = $conn->prepare("SELECT id FROM dbs WHERE db_code = ?");
            $checkStmt->execute([$db_code]);
            $attempts = 0;
            while ($checkStmt->fetch() && $attempts < 10) {
                // Use microtime for better uniqueness
                $microtime = microtime(true);
                $hashInput = $db_name . '_' . $microtime;
                $db_code = 'db_' . substr(md5($hashInput), 0, 20);
                $checkStmt->execute([$db_code]);
                $attempts++;
            }
            
            // Prepare data
            $db_host_start = !empty($inputData['db_host_start']) ? $inputData['db_host_start'] : null;
            $db_host_end = !empty($inputData['db_host_end']) ? $inputData['db_host_end'] : null;
            $db_host_cost = !empty($inputData['db_host_cost_per_month']) ? floatval($inputData['db_host_cost_per_month']) : null;
            $serv_host_start = !empty($inputData['serv_host_start']) ? $inputData['serv_host_start'] : null;
            $serv_host_end = !empty($inputData['serv_host_end']) ? $inputData['serv_host_end'] : null;
            $serv_host_cost = !empty($inputData['serv_host_cost_per_month']) ? floatval($inputData['serv_host_cost_per_month']) : null;

            // The folders are derived from the database name, not taken from the form.
            // What the form sent is only reported, so a hand-edited value is visible
            // rather than silently discarded.
            $dirs = dbm_tenant_dirs($db_name);
            $files_dir = $dirs['files_dir'];
            $js_dir = $dirs['js_dir'];

            $submittedJs = trim((string) ($inputData['js_dir'] ?? ''), '/');
            $submittedFiles = trim((string) ($inputData['files_dir'] ?? ''), '/');
            $notes = [];
            if ($submittedJs !== '' && $submittedJs !== $dirs['app_folder']) {
                $notes[] = sprintf("js_dir was %s; set to %s", $dirs['js_dir'], $dirs['js_dir']);
            }
            if ($submittedFiles !== '' && $submittedFiles !== $dirs['files_folder']) {
                $notes[] = sprintf("files_dir was %s; set to %s", $dirs['files_dir'], $dirs['files_dir']);
            }
            
            $insertStmt = $conn->prepare("INSERT INTO dbs (db_code, db_name, db_host_start, db_host_end, db_host_cost_per_month, serv_host_start, serv_host_end, serv_host_cost_per_month, files_dir, js_dir) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)");
            
            if ($insertStmt->execute([$db_code, $db_name, $db_host_start, $db_host_end, $db_host_cost, $serv_host_start, $serv_host_end, $serv_host_cost, $files_dir, $js_dir])) {
                $response['success'] = true;
                $response['message'] = 'Database created successfully';
                $response['data'] = [
                    'id' => (int) $conn->lastInsertId(),
                    'db_code' => $db_code,
                    'db_name' => $db_name,
                    'js_dir' => $js_dir,
                    'files_dir' => $files_dir,
                    'notes' => $notes,
                    // The row exists but nothing has been created on disk yet, so the
                    // UI can point at the Provision button straight away.
                    'next_step' => 'provision_tenant'
                ];
            } else {
                $response['message'] = 'Failed to create database';
            }
            break;
            
        case 'update_database':
            // Update database
            $id = intval($inputData['id'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            // Check if database exists
            $checkStmt = $conn->prepare("SELECT id FROM dbs WHERE id = ?");
            $checkStmt->execute([$id]);
            if (!$checkStmt->fetch()) {
                $response['message'] = 'Database not found';
                break;
            }
            
            $db_name = trim($inputData['db_name'] ?? '');
            
            if (empty($db_name)) {
                $response['message'] = 'DB Name is required';
                break;
            }

            $nameCheck = dbm_validate_database_name($db_name);
            if (isset($nameCheck['error'])) {
                $response['message'] = $nameCheck['error'];
                break;
            }

            $nameTaken = dbm_assert_name_available($conn, $db_name, $id);
            if ($nameTaken !== null) {
                $response['message'] = $nameTaken;
                break;
            }
            
            // Get existing db_code (it cannot be changed)
            $getStmt = $conn->prepare("SELECT db_code, db_name FROM dbs WHERE id = ?");
            $getStmt->execute([$id]);
            $existing = $getStmt->fetch(PDO::FETCH_ASSOC);
            
            // Check if database still exists (race condition protection)
            if (!$existing || !isset($existing['db_code'])) {
                $response['message'] = 'Database not found or was deleted';
                break;
            }
            
            $db_code = $existing['db_code'];
            $notes = [];

            // Renaming a deployed client is not a row edit.
            //
            // The folders are named after the database (dbm_tenant_dirs), so a new name
            // is a different folder and a different URL - the old folder keeps serving
            // the old URL with a tenant config in it, and the row now describes neither.
            // Refuse it once anything exists on disk or in the database, and say what to
            // do instead.
            // The OLD name, not the new one: the question is whether this client is
            // already deployed. Testing the proposed name instead would let a
            // deployed client be renamed to any unused name, which is the one thing
            // this guard exists to stop.
            if ((string) $existing['db_name'] !== $db_name && dbm_is_provisioned($conn, (string) $existing['db_name'])) {
                $response['message'] = sprintf(
                    "'%s' has already been provisioned, so it cannot be renamed to '%s': the folders, the "
                    . "URL and tenant config are all named after the old one. Provision a new database instead, "
                    . "or rename the %s/ and %s/ folders and rewrite tenant config by hand.",
                    $existing['db_name'],
                    $db_name,
                    $existing['db_name'],
                    tenant_files_dir_name((string) $existing['db_name'])
                );
                break;
            }
            
            // Prepare data
            $db_host_start = !empty($inputData['db_host_start']) ? $inputData['db_host_start'] : null;
            $db_host_end = !empty($inputData['db_host_end']) ? $inputData['db_host_end'] : null;
            $db_host_cost = !empty($inputData['db_host_cost_per_month']) ? floatval($inputData['db_host_cost_per_month']) : null;
            $serv_host_start = !empty($inputData['serv_host_start']) ? $inputData['serv_host_start'] : null;
            $serv_host_end = !empty($inputData['serv_host_end']) ? $inputData['serv_host_end'] : null;
            $serv_host_cost = !empty($inputData['serv_host_cost_per_month']) ? floatval($inputData['serv_host_cost_per_month']) : null;

            // Derived from the database name, for the same reason as create_database.
            $dirs = dbm_tenant_dirs($db_name);
            $files_dir = $dirs['files_dir'];
            $js_dir = $dirs['js_dir'];

            $submittedJs = trim((string) ($inputData['js_dir'] ?? ''), '/');
            $submittedFiles = trim((string) ($inputData['files_dir'] ?? ''), '/');
            if ($submittedJs !== '' && $submittedJs !== $dirs['app_folder']) {
                $notes[] = sprintf("js_dir was %s; set to %s", $dirs['js_dir'], $dirs['js_dir']);
            }
            if ($submittedFiles !== '' && $submittedFiles !== $dirs['files_folder']) {
                $notes[] = sprintf("files_dir was %s; set to %s", $dirs['files_dir'], $dirs['files_dir']);
            }

            $updateStmt = $conn->prepare("UPDATE dbs SET db_code = ?, db_name = ?, db_host_start = ?, db_host_end = ?, db_host_cost_per_month = ?, serv_host_start = ?, serv_host_end = ?, serv_host_cost_per_month = ?, files_dir = ?, js_dir = ? WHERE id = ?");
            
            if ($updateStmt->execute([$db_code, $db_name, $db_host_start, $db_host_end, $db_host_cost, $serv_host_start, $serv_host_end, $serv_host_cost, $files_dir, $js_dir, $id])) {
                $response['success'] = true;
                $response['message'] = 'Database updated successfully';
                $response['data'] = [
                    'id' => $id,
                    'db_name' => $db_name,
                    'js_dir' => $js_dir,
                    'files_dir' => $files_dir,
                    'notes' => $notes,
                ];
            } else {
                $response['message'] = 'Failed to update database';
            }
            break;
            
        case 'delete_database':
            // Delete database
            $id = intval($inputData['id'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            // Check if database exists
            $checkStmt = $conn->prepare("SELECT id FROM dbs WHERE id = ?");
            $checkStmt->execute([$id]);
            if (!$checkStmt->fetch()) {
                $response['message'] = 'Database not found';
                break;
            }
            
            $deleteStmt = $conn->prepare("DELETE FROM dbs WHERE id = ?");
            
            if ($deleteStmt->execute([$id])) {
                $response['success'] = true;
                $response['message'] = 'Database deleted successfully';
            } else {
                $response['message'] = 'Failed to delete database';
            }
            break;
            
        case 'create_tables':
            // Create tables from setup.sql
            $id = intval($inputData['id'] ?? 0);
            $version = intval($inputData['version'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            if ($version <= 0) {
                $response['message'] = 'Version is required and must be a positive number';
                break;
            }
            
            // Get database info (including db_code, files_dir, and js_dir for directory and tenant config creation)
            $getStmt = $conn->prepare("SELECT id, db_name, db_code, files_dir, js_dir, is_created FROM dbs WHERE id = ?");
            $getStmt->execute([$id]);
            $dbInfo = $getStmt->fetch(PDO::FETCH_ASSOC);
            
            if (!$dbInfo) {
                $response['message'] = 'Database not found';
                break;
            }
            
            if ($dbInfo['is_created'] == 1) {
                $response['message'] = 'Tables have already been created for this database';
                break;
            }
            
            if (empty($dbInfo['db_name'])) {
                $response['message'] = 'Database name is not set';
                break;
            }
            
            // Read setup.sql file
            $setupSqlPath = __DIR__ . '/setup.sql';
            if (!file_exists($setupSqlPath)) {
                $response['message'] = 'setup.sql file not found';
                break;
            }
            
            $setupSql = file_get_contents($setupSqlPath);
            if ($setupSql === false) {
                $response['message'] = 'Failed to read setup.sql file';
                break;
            }
            
            // Parse SQL file to extract CREATE TABLE and INSERT statements
            $statements = dbm_extract_sql_statements($setupSql);
            if (empty($statements)) {
                $response['message'] = 'No CREATE TABLE or INSERT statements found in setup.sql';
                break;
            }
            
            // Get database credentials from config.php
            $target = dbm_connect_target(trim((string) $dbInfo['db_name']));
            if (isset($target['error'])) {
                $response['message'] = $target['error'];
                break;
            }
            $targetConn = $target['conn'];
            
            // Execute CREATE TABLE statements
            $executedCount = 0;
            $errors = [];
            foreach ($statements as $statement) {
                try {
                    $targetConn->exec($statement);
                    $executedCount++;
                } catch (PDOException $e) {
                    // Keep the driver text out of the response. It carries SQLSTATE codes
                    // and the server's own error wording, which is noise in the UI and
                    // tells the reader nothing they can act on - the log has it.
                    error_log(sprintf(
                        "create_tables: statement %d failed on '%s': %s",
                        $executedCount + count($errors) + 1,
                        $dbInfo['db_name'],
                        $e->getMessage()
                    ));
                    $errors[] = sprintf('statement %d was rejected by MySQL', $executedCount + count($errors) + 1);
                }
            }
            
            if (!empty($errors) && $executedCount === 0) {
                error_log('create_tables: every statement failed for ' . $dbInfo['db_name']);
                $response['message'] = 'No statement could be applied: ' . implode('; ', $errors);
                break;
            }
            
            // Create directories (files_dir and js_dir) if they are configured
            $directoriesCreated = [];
            $directoryErrors = [];
            
            // Helper function to create directory safely
            $createDirectory = function($dirPath, $dirName) use (&$directoriesCreated, &$directoryErrors) {
                if (empty($dirPath)) {
                    return; // Skip if directory path is empty
                }
                
                try {
                    $dir = trim($dirPath);
                    // Remove leading slash if present
                    $dir = ltrim($dir, '/');
                    $dir = rtrim($dir, '/');
                    
                    if (empty($dir)) {
                        return; // Skip if directory is empty after trimming
                    }
                    
                    // Construct the full directory path
                    $fullDirPath = __DIR__ . '/../' . $dir;
                    
                    // Get the real path for security check
                    $realBasePath = realpath(__DIR__ . '/../');
                    if ($realBasePath === false) {
                        throw new Exception('Invalid base directory');
                    }
                    $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                    
                    // Normalize the directory path
                    $normalizedDirPath = str_replace(['\\', '/'], DIRECTORY_SEPARATOR, $fullDirPath);
                    $normalizedDirPath = preg_replace('/' . preg_quote(DIRECTORY_SEPARATOR, '/') . '+/', DIRECTORY_SEPARATOR, $normalizedDirPath);
                    
                    // Create directory and all parent directories if they don't exist
                    if (!is_dir($normalizedDirPath)) {
                        // Create directory recursively (creates parent directories if needed)
                        if (!mkdir($normalizedDirPath, 0755, true)) {
                            throw new Exception('Failed to create directory: ' . $normalizedDirPath);
                        }
                    }
                    
                    // Verify the directory exists and is within the base directory
                    $resolvedDirPath = realpath($normalizedDirPath);
                    if ($resolvedDirPath === false) {
                        throw new Exception('Failed to resolve directory path: ' . $normalizedDirPath);
                    }
                    $resolvedDirPath = rtrim($resolvedDirPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                    
                    // Security check: ensure the resolved path is within the base directory
                    if (strpos($resolvedDirPath, $realBasePath) !== 0) {
                        throw new Exception('Directory traversal attempt detected');
                    }
                    
                    $directoriesCreated[] = $dirName . ' (' . $dir . ')';
                } catch (Exception $e) {
                    $directoryErrors[] = $dirName . ': ' . $e->getMessage();
                    error_log('Failed to create ' . $dirName . ' directory: ' . $e->getMessage());
                }
            };
            
            // Create files_dir if configured
            if (!empty($dbInfo['files_dir'])) {
                $createDirectory($dbInfo['files_dir'], 'files_dir');
            }
            
            // Create js_dir if configured
            if (!empty($dbInfo['js_dir'])) {
                $createDirectory($dbInfo['js_dir'], 'js_dir');
            }
            
            // Update version in versions table
            try {
                $versionStmt = $targetConn->prepare("INSERT INTO versions (id, version) VALUES (1, ?) ON DUPLICATE KEY UPDATE version = ?");
                $versionStmt->execute([$version, $version]);
            } catch (PDOException $e) {
                // If versions table doesn't exist yet or update fails, log but don't fail the whole operation
                error_log('Failed to update version: ' . $e->getMessage());
            }
            
            // Update is_created flag
            $updateStmt = $conn->prepare("UPDATE dbs SET is_created = 1 WHERE id = ?");
            if ($updateStmt->execute([$id])) {
                // Create tenant config file if js_dir is configured
                $dbCodeJsonCreated = false;
                $dbCodeJsonError = null;
                
                if (!empty($dbInfo['js_dir']) && !empty($dbInfo['db_code'])) {
                    try {
                        $jsDir = trim($dbInfo['js_dir']);
                        // Remove leading slash if present
                        $jsDir = ltrim($jsDir, '/');
                        $jsDir = rtrim($jsDir, '/');
                        
                        // Construct the full directory and file path
                        $dirPath = __DIR__ . '/../' . $jsDir;
                        $filePath = $dirPath . '/tenant config';
                        
                        // Get the real path for security check
                        $realBasePath = realpath(__DIR__ . '/../');
                        if ($realBasePath === false) {
                            throw new Exception('Invalid base directory');
                        }
                        $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                        
                        // Normalize the directory path
                        $normalizedDirPath = str_replace(['\\', '/'], DIRECTORY_SEPARATOR, $dirPath);
                        $normalizedDirPath = preg_replace('/' . preg_quote(DIRECTORY_SEPARATOR, '/') . '+/', DIRECTORY_SEPARATOR, $normalizedDirPath);
                        
                        // Create directory and all parent directories if they don't exist
                        if (!is_dir($normalizedDirPath)) {
                            // Create directory recursively (creates parent directories if needed)
                            if (!mkdir($normalizedDirPath, 0755, true)) {
                                throw new Exception('Failed to create directory: ' . $normalizedDirPath);
                            }
                        }
                        
                        // Verify the directory exists and is within the base directory
                        $resolvedDirPath = realpath($normalizedDirPath);
                        if ($resolvedDirPath === false) {
                            throw new Exception('Failed to resolve directory path: ' . $normalizedDirPath);
                        }
                        $resolvedDirPath = rtrim($resolvedDirPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                        
                        // Security check: ensure the resolved path is within the base directory
                        if (strpos($resolvedDirPath, $realBasePath) !== 0) {
                            throw new Exception('Directory traversal attempt detected');
                        }
                        
                        // Create JSON content
                        $jsonContent = [
                            'db_code' => $dbInfo['db_code']
                        ];
                        
                        // Encode JSON with pretty printing
                        $jsonString = json_encode($jsonContent, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                        
                        if ($jsonString === false) {
                            throw new Exception('Failed to encode JSON: ' . json_last_error_msg());
                        }
                        
                        // Write file (only if it doesn't exist)
                        if (!file_exists($filePath)) {
                            if (file_put_contents($filePath, $jsonString) === false) {
                                throw new Exception('Failed to write tenant config file');
                            }
                            $dbCodeJsonCreated = true;
                        }
                    } catch (Exception $e) {
                        $dbCodeJsonError = $e->getMessage();
                        error_log('Failed to create tenant config: ' . $dbCodeJsonError);
                    }
                }
                
                $response['success'] = true;
                // Counted in statements, not tables: the run also applies the setup.sql
                // INSERTs and the two triggers, so "table(s)" would have overstated it.
                $response['message'] = "Successfully ran {$executedCount} statement(s) from setup.sql";
                
                // Add directory creation status
                if (!empty($directoriesCreated)) {
                    $response['message'] .= ', created directories: ' . implode(', ', $directoriesCreated);
                }
                if (!empty($directoryErrors)) {
                    $response['message'] .= ' (directory errors: ' . implode('; ', $directoryErrors) . ')';
                }
                
                // Add tenant config creation status
                if ($dbCodeJsonCreated) {
                    $response['message'] .= ', created tenant config';
                }
                if ($dbCodeJsonError) {
                    $response['message'] .= ' (note: tenant config creation failed: ' . $dbCodeJsonError . ')';
                }
                
                // Add table creation errors if any
                if (!empty($errors)) {
                    $response['message'] .= ' (' . count($errors) . ' statement(s) failed: ' . implode('; ', $errors) . ')';
                }
            } else {
                $response['message'] = 'Tables created but failed to update is_created flag';
            }
            break;
            
        case 'tenant_status':
            // What exists for a client and what is missing, without changing anything.
            // The Provision screen renders this as its checklist, so the button is only
            // pressed once the operator can see what it will do.
            $statusName = dbm_requested_db_name($conn, $inputData);
            $statusCheck = dbm_validate_database_name($statusName);
            if (isset($statusCheck['error'])) {
                $response['message'] = $statusCheck['error'];
                break;
            }

            try {
                $statusTarget = dbm_target_credentials();
                $response['data'] = tenant_status([
                    'db_name' => $statusName,
                    'app_creds' => $statusTarget,
                    'registry_creds' => $db_manager_config,
                    'webroot' => dbm_deployment_root(),
                    'shared_api' => dbm_has_shared_api(),
                ]);
                $response['success'] = true;
                $response['message'] = $response['data']['ready']
                    ? 'This client is fully provisioned.'
                    : 'Not provisioned yet: ' . implode(', ', $response['data']['blocking']);
            } catch (Throwable $e) {
                error_log('tenant_status: ' . $e->getMessage());
                $response['message'] = 'Could not read the tenant status: ' . $e->getMessage();
            }
            break;

        case 'provision_tenant':
            // The whole onboarding, from the browser: database, schema, reference
            // data, folders, tenant config and the build. This is the action that lets a
            // new client be set up without shell access to the server.
            $provisionName = dbm_requested_db_name($conn, $inputData);
            $provisionCheck = dbm_validate_database_name($provisionName);
            if (isset($provisionCheck['error'])) {
                $response['message'] = $provisionCheck['error'];
                break;
            }

            $provisionLog = [];
            try {
                $provisionTarget = dbm_target_credentials();

                // Where the reference data comes from, in order: the request, then the
                // server config. Without the config fallback this defaults to the app
                // database, which on this machine is a live tenant - and the template
                // check then refuses the run with a message about client data, which
                // is correct but is a confusing way to learn that no template is
                // configured.
                $seedSource = trim((string) ($inputData['seed_source'] ?? ''));
                if ($seedSource === '') {
                    $seedSource = trim((string) tenant_server_config()['template_database']);
                }
                $seedSource = $seedSource === '' ? null : $seedSource;

                $report = tenant_provision([
                    'db_name' => $provisionName,
                    'app_creds' => $provisionTarget,
                    'registry_creds' => $db_manager_config,
                    'webroot' => dbm_deployment_root(),
                    'seed_source' => $seedSource,
                    'force' => !empty($inputData['force']),
                    'shared_api' => dbm_has_shared_api(),
                    'log' => static function (string $line) use (&$provisionLog): void {
                        $provisionLog[] = $line;
                    },
                ]);

                // Mark it created, which is what the other actions read to decide
                // whether this is a live client or a draft row.
                $markStmt = $conn->prepare('UPDATE dbs SET is_created = 1 WHERE db_name = ?');
                $markStmt->execute([$provisionName]);

                $response['success'] = true;
                $response['message'] = sprintf('Provisioned %s. Serve it at %s', $provisionName, $report['url_path']);
                $response['data'] = ['report' => $report, 'log' => $provisionLog];
            } catch (Throwable $e) {
                // The log goes back with the failure: a provisioning run that stops at
                // "duplicate column" needs the lines above it to be diagnosable, and
                // the server's error_log is not something the operator can see.
                error_log('provision_tenant: ' . $e->getMessage());
                $response['message'] = $e->getMessage();
                $response['data'] = ['log' => $provisionLog];
            }
            break;

        case 'deploy_app_to_tenant':
            // Re-copy the current build into the ONE shared dist/, which is where
            // every client on this server is served from. Separate from provisioning
            // because it is the action taken after every release, and it must not
            // touch a database.
            //
            // Server-wide, not per client: there is no per-client build to copy any
            // more, so there is no client to name. A db_name is still accepted when
            // sent - an older build of the UI sends one from the Provision screen -
            // but it is not required and deliberately not used to choose a target,
            // because the destination is the same dist/ either way.
            $deployLog = [];
            try {
                $serverConfig = tenant_server_config();
                $distFolder = rtrim(str_replace('\\', '/', dbm_deployment_root()), '/') . '/' . TENANT_DIST_DIR_NAME;

                tenant_deploy_app(
                    ['dist_folder' => $distFolder],
                    (string) $serverConfig['canonical_build'],
                    static function (string $line) use (&$deployLog): void {
                        $deployLog[] = $line;
                    }
                );

                $response['message'] = sprintf(
                    'Deployed the current build to the shared %s/. Every client on this server is served from it.',
                    $distFolder
                );
                $response['data'] = ['dist_folder' => $distFolder, 'log' => $deployLog];
            } catch (Throwable $e) {
                error_log('deploy_app_to_tenant: ' . $e->getMessage());
                $response['message'] = $e->getMessage();
                $response['data'] = ['log' => $deployLog];
            }
            break;

        case 'reload_nginx':
            // Re-generate the server block list from the registry and reload nginx.
            //
            // This is the one action the web user cannot do itself, because writing
            // into /etc/nginx and signalling the master process are root operations.
            // It is delegated to a root-owned script through a sudoers rule that names
            // that one path with no arguments, so this action cannot become a way to
            // run anything else as root:
            //
            //   www-data ALL=(root) NOPASSWD: /usr/local/bin/cars-nginx-render
            //
            // The path comes from the root-owned /etc/cars-deploy.json, never from the
            // request, and it is executed as an argument array with no shell, so there
            // is nothing here for a caller to interpolate into.
            $serverConfig = tenant_server_config();
            $renderCommand = (string) $serverConfig['render_command'];

            if ($renderCommand === '' || !is_file($renderCommand)) {
                $response['message'] = sprintf(
                    'The nginx renderer is not installed: %s does not exist. It is the one root-owned '
                    . 'part of this setup - see DEPLOYMENT.md, "One-time server setup".',
                    $renderCommand === '' ? '(no render_command configured)' : $renderCommand
                );
                break;
            }

            $sudoPath = trim((string) (shell_exec('command -v sudo 2>/dev/null') ?? ''));
            if ($sudoPath === '') {
                $response['message'] = 'sudo is not installed, so nginx cannot be reloaded from here. Reload it over SSH.';
                break;
            }

            $descriptors = [1 => ['pipe', 'w'], 2 => ['pipe', 'w']];
            // -n: never prompt. A web request that blocks on a password prompt would
            // hang until the PHP timeout, and the operator would see nothing.
            $renderProcess = @proc_open(
                [$sudoPath, '-n', $renderCommand],
                $descriptors,
                $renderPipes,
                null,
                ['PATH' => getenv('PATH') ?: '/usr/bin:/bin', 'HOME' => getenv('HOME') ?: '/tmp']
            );

            if (!is_resource($renderProcess)) {
                error_log('reload_nginx: could not start ' . $renderCommand);
                $response['message'] = 'Could not run the nginx renderer.';
                break;
            }

            $renderOut = stream_get_contents($renderPipes[1]);
            fclose($renderPipes[1]);
            $renderErr = stream_get_contents($renderPipes[2]);
            fclose($renderPipes[2]);
            $renderCode = proc_close($renderProcess);

            if ($renderCode === 0) {
                $response['success'] = true;
                $response['message'] = 'nginx configuration regenerated and reloaded.';
                $response['data'] = ['output' => trim($renderOut)];
            } else {
                // The renderer runs `nginx -t` before installing anything, so a
                // non-zero exit means the configuration was NOT replaced and the
                // running server is untouched. Saying so matters: an operator who
                // assumed otherwise would go looking for a broken site instead.
                error_log('reload_nginx: ' . $renderErr);
                $renderMessage = 'nginx was not reloaded; the running configuration is unchanged.';
                if (trim($renderErr) !== '') {
                    $renderMessage .= ' The renderer said: ' . trim($renderErr);
                }
                $response['message'] = $renderMessage;
                $response['data'] = ['output' => trim($renderOut), 'error' => trim($renderErr), 'exit_code' => $renderCode];
            }
            break;

        case 'deployment_config':
            // Whether the one-time server setup is in place, and what it says.
            //
            // Read-only and safe to call before anything exists, which is the point: it
            // is how the screen tells an operator "this server has not been set up yet"
            // instead of letting them press Provision and get a folder-permission error.
            $serverConfig = tenant_server_config();
            $renderCommand = (string) $serverConfig['render_command'];
            $webroot = dbm_deployment_root();

            // These are grouped by what they actually block, because lumping them
            // together as one boolean made the wizard unusable on any machine that
            // is not the production server - and the two halves have nothing to do
            // with each other:
            //
            //   provision - provisioning runs as the web user and creates a database,
            //               folders and tenant config. It never needs root, so a
            //               missing nginx setup must not stop it.
            //   nginx     - only this needs root, and it is the one step with effects
            //               beyond the client being set up.
            //   warn      - does not block either. Worth saying out loud because being
            //               wrong is quiet: a wrong shared-api answer gives every
            //               client a private api/ and the next deploy updates one of
            //               them, and a missing canonical build skips the file copy.
            //
            // Before this split, `render script` and `server_name` gated Provision,
            // so on a dev machine the button was permanently disabled and the flow
            // could only be exercised by calling the API by hand - which is how the
            // DB manager managed to be completely unreachable in production while
            // every test passed.
            // The file that actually answered, not a hardcoded /etc path: on a dev
            // machine that is a file in the repository, and telling someone to edit
            // /etc/cars-deploy.json - which they cannot write without sudo - is a
            // dead end dressed as an instruction.
            $configWhere = (string) ($serverConfig['config_path'] ?? '') ?: 'the server config';
            $notSet = static fn (string $key): string => $key === ''
                ? 'not set in ' . $configWhere
                : $key;

            $checks = [
                [
                    'name' => 'render script',
                    'needs' => 'nginx',
                    'ok' => $renderCommand !== '' && is_file($renderCommand),
                    'detail' => $renderCommand,
                ],
                [
                    'name' => 'server_name',
                    'needs' => 'nginx',
                    'ok' => (string) $serverConfig['server_name'] !== '',
                    'detail' => $notSet((string) $serverConfig['server_name']),
                ],
                [
                    'name' => 'canonical build',
                    'needs' => 'warn',
                    'ok' => (string) $serverConfig['canonical_build'] !== '' && is_dir((string) $serverConfig['canonical_build']),
                    'detail' => $notSet((string) $serverConfig['canonical_build']),
                ],
                [
                    // Without one, every provisioning run falls back to the app
                    // database and is refused by the template check.
                    'name' => 'template database',
                    'needs' => 'provision',
                    'ok' => (string) $serverConfig['template_database'] !== '',
                    'detail' => $notSet((string) $serverConfig['template_database']),
                ],
                [
                    // The web user has to be able to create a tenant's folders.
                    'name' => 'webroot writable',
                    'needs' => 'provision',
                    'ok' => is_writable($webroot),
                    'detail' => $webroot,
                ],
                [
                    'name' => 'shared api/',
                    'needs' => 'warn',
                    'ok' => dbm_has_shared_api(),
                    'detail' => dbm_has_shared_api() ? 'one api/ serves every tenant' : 'each tenant has its own api/ copy',
                ],
            ];

            $unmet = static fn (string $needs): array => array_values(array_map(
                static fn (array $c): string => $c['name'],
                array_filter($checks, static fn (array $c): bool => $c['needs'] === $needs && !$c['ok'])
            ));

            $response['success'] = true;
            $response['data'] = [
                'config' => $serverConfig,
                'webroot' => $webroot,
                'shared_api' => dbm_has_shared_api(),
                'checks' => $checks,
                // Only the checks provisioning itself depends on. See above.
                'ready' => $unmet('provision') === [],
                'unmet' => $unmet('provision'),
                'nginx_ready' => $unmet('nginx') === [],
                'nginx_unmet' => $unmet('nginx'),
                'warnings' => $unmet('warn'),
            ];
            break;

        case 'get_db_updates':
            // Get all db_updates
            $stmt = $conn->prepare("SELECT * FROM db_updates ORDER BY from_version ASC, current_version ASC");
            $stmt->execute();
            $updates = $stmt->fetchAll(PDO::FETCH_ASSOC);
            $response['success'] = true;
            $response['data'] = $updates;
            break;
            
        case 'create_db_update':
            // Create new db_update
            $fromVersion = intval($inputData['from_version'] ?? 0);
            $currentVersion = intval($inputData['current_version'] ?? 0);
            $description = trim($inputData['description'] ?? '');
            $sql = trim($inputData['sql'] ?? '');
            
            if ($fromVersion <= 0 || $currentVersion <= 0) {
                $response['message'] = 'From version and current version must be positive numbers';
                break;
            }
            
            if (empty($sql)) {
                $response['message'] = 'SQL is required';
                break;
            }
            
            if ($currentVersion <= $fromVersion) {
                $response['message'] = 'Current version must be greater than from version';
                break;
            }
            
            $insertStmt = $conn->prepare("INSERT INTO db_updates (from_version, current_version, description, `sql`) VALUES (?, ?, ?, ?)");
            if ($insertStmt->execute([$fromVersion, $currentVersion, $description, $sql])) {
                $response['success'] = true;
                $response['message'] = 'Update created successfully';
                $response['data'] = ['id' => $conn->lastInsertId()];
            } else {
                $response['message'] = 'Failed to create update';
            }
            break;
            
        case 'update_db_update':
            // Update existing db_update
            $id = intval($inputData['id'] ?? 0);
            $fromVersion = intval($inputData['from_version'] ?? 0);
            $currentVersion = intval($inputData['current_version'] ?? 0);
            $description = trim($inputData['description'] ?? '');
            $sql = trim($inputData['sql'] ?? '');
            
            if ($id <= 0) {
                $response['message'] = 'Invalid update ID';
                break;
            }
            
            if ($fromVersion <= 0 || $currentVersion <= 0) {
                $response['message'] = 'From version and current version must be positive numbers';
                break;
            }
            
            if (empty($sql)) {
                $response['message'] = 'SQL is required';
                break;
            }
            
            if ($currentVersion <= $fromVersion) {
                $response['message'] = 'Current version must be greater than from version';
                break;
            }
            
            $updateStmt = $conn->prepare("UPDATE db_updates SET from_version = ?, current_version = ?, description = ?, `sql` = ? WHERE id = ?");
            if ($updateStmt->execute([$fromVersion, $currentVersion, $description, $sql, $id])) {
                $response['success'] = true;
                $response['message'] = 'Update updated successfully';
            } else {
                $response['message'] = 'Failed to update';
            }
            break;
            
        case 'delete_db_update':
            // Delete db_update
            $id = intval($inputData['id'] ?? 0);
            
            if ($id <= 0) {
                $response['message'] = 'Invalid update ID';
                break;
            }
            
            $deleteStmt = $conn->prepare("DELETE FROM db_updates WHERE id = ?");
            if ($deleteStmt->execute([$id])) {
                $response['success'] = true;
                $response['message'] = 'Update deleted successfully';
            } else {
                $response['message'] = 'Failed to delete update';
            }
            break;
            
        case 'update_databases_version':
            // Update version for multiple databases
            $databaseIds = $inputData['database_ids'] ?? [];
            $version = intval($inputData['version'] ?? 0);
            
            if (empty($databaseIds) || !is_array($databaseIds)) {
                $response['message'] = 'Database IDs are required';
                break;
            }
            
            if ($version <= 0) {
                $response['message'] = 'Version must be a positive number';
                break;
            }
            
            // Get database credentials from config.php
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            // Get database names for the selected IDs
            $placeholders = implode(',', array_fill(0, count($databaseIds), '?'));
            $getStmt = $conn->prepare("SELECT id, db_name FROM dbs WHERE id IN ($placeholders)");
            $getStmt->execute($databaseIds);
            $selectedDbs = $getStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($selectedDbs)) {
                $response['message'] = 'No databases found for the selected IDs';
                break;
            }
            
            $successCount = 0;
            $errors = [];
            
            foreach ($selectedDbs as $db) {
                if (empty($db['db_name'])) {
                    $errors[] = "Database ID {$db['id']}: Database name is not set";
                    continue;
                }
                
                try {
                    $targetConn = new PDO(
                        "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                        $targetDbUser,
                        $targetDbPass
                    );
                    $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                    
                    // Update version in versions table
                    $versionStmt = $targetConn->prepare("INSERT INTO versions (id, version) VALUES (1, ?) ON DUPLICATE KEY UPDATE version = ?");
                    $versionStmt->execute([$version, $version]);
                    $successCount++;
                } catch (PDOException $e) {
                    $errors[] = "Database {$db['db_name']}: " . $e->getMessage();
                }
            }
            
            if ($successCount > 0) {
                $response['success'] = true;
                $response['message'] = "Version updated successfully for {$successCount} database(s)";
                if (!empty($errors)) {
                    $response['message'] .= ' (some errors: ' . implode('; ', $errors) . ')';
                }
            } else {
                $response['message'] = 'Failed to update version: ' . implode('; ', $errors);
            }
            break;
            
        case 'update_structure':
            // Update structure for selected databases based on db_updates table
            $databaseIds = $inputData['database_ids'] ?? [];
            
            if (empty($databaseIds) || !is_array($databaseIds)) {
                $response['message'] = 'Database IDs are required';
                break;
            }
            
            // Get all db_updates records
            $updatesStmt = $conn->prepare("SELECT * FROM db_updates ORDER BY from_version ASC, current_version ASC");
            $updatesStmt->execute();
            $dbUpdates = $updatesStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($dbUpdates)) {
                $response['message'] = 'No update records found in db_updates table';
                break;
            }
            
            // Get database credentials from config.php
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            // Get selected databases with their versions
            $placeholders = implode(',', array_fill(0, count($databaseIds), '?'));
            $getStmt = $conn->prepare("SELECT id, db_name FROM dbs WHERE id IN ($placeholders)");
            $getStmt->execute($databaseIds);
            $selectedDbs = $getStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($selectedDbs)) {
                $response['message'] = 'No databases found for the selected IDs';
                break;
            }
            
            // Fetch versions for selected databases
            $dbVersions = [];
            foreach ($selectedDbs as $db) {
                if (empty($db['db_name'])) {
                    continue;
                }
                
                try {
                    $targetConn = new PDO(
                        "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                        $targetDbUser,
                        $targetDbPass
                    );
                    $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                    
                    $versionStmt = $targetConn->prepare("SELECT version FROM versions WHERE id = 1");
                    $versionStmt->execute();
                    $versionData = $versionStmt->fetch(PDO::FETCH_ASSOC);
                    $dbVersions[$db['id']] = [
                        'db_name' => $db['db_name'],
                        'version' => $versionData ? intval($versionData['version']) : null,
                        'conn' => $targetConn
                    ];
                } catch (PDOException $e) {
                    $dbVersions[$db['id']] = [
                        'db_name' => $db['db_name'],
                        'version' => null,
                        'conn' => null,
                        'error' => $e->getMessage()
                    ];
                }
            }
            
            $totalUpdates = 0;
            $totalErrors = 0;
            $errors = [];
            // Track which databases have already been marked with connection/version errors
            // to avoid counting the same error multiple times (once per update record)
            $dbErrorsTracked = [];
            
            // Iterate through db_updates records
            foreach ($dbUpdates as $update) {
                $fromVersion = intval($update['from_version']);
                $currentVersion = intval($update['current_version']);
                $sql = $update['sql'];
                
                // For each selected database, check if version >= from_version
                foreach ($dbVersions as $dbId => $dbInfo) {
                    // Check for connection error (only count once per database)
                    if ($dbInfo['conn'] === null) {
                        if (!isset($dbErrorsTracked[$dbId]['conn'])) {
                            $errors[] = "Database {$dbInfo['db_name']}: Cannot connect - " . ($dbInfo['error'] ?? 'Unknown error');
                            $totalErrors++;
                            $dbErrorsTracked[$dbId]['conn'] = true;
                        }
                        continue;
                    }
                    
                    // Check for version not found (only count once per database)
                    if ($dbInfo['version'] === null) {
                        if (!isset($dbErrorsTracked[$dbId]['version'])) {
                            $errors[] = "Database {$dbInfo['db_name']}: Version not found";
                            $totalErrors++;
                            $dbErrorsTracked[$dbId]['version'] = true;
                        }
                        continue;
                    }
                    
                    // Check if database version >= from_version
                    if ($dbInfo['version'] >= $fromVersion) {
                        try {
                            // Execute the SQL statement
                            $dbInfo['conn']->exec($sql);
                            
                            // Update version to current_version after successful SQL execution
                            $versionUpdateStmt = $dbInfo['conn']->prepare("INSERT INTO versions (id, version) VALUES (1, ?) ON DUPLICATE KEY UPDATE version = ?");
                            $versionUpdateStmt->execute([$currentVersion, $currentVersion]);
                            
                            // Update the version in our tracking array for subsequent checks
                            $dbVersions[$dbId]['version'] = $currentVersion;
                            
                            $totalUpdates++;
                        } catch (PDOException $e) {
                            $errors[] = "Database {$dbInfo['db_name']} (update from v{$fromVersion} to v{$currentVersion}): " . $e->getMessage();
                            $totalErrors++;
                        }
                    }
                }
            }
            
            // Close all connections
            foreach ($dbVersions as $dbInfo) {
                if ($dbInfo['conn'] !== null) {
                    $dbInfo['conn'] = null;
                }
            }
            
            if ($totalUpdates > 0 || $totalErrors === 0) {
                $response['success'] = true;
                $response['message'] = "Structure update completed: {$totalUpdates} SQL statement(s) executed";
                if (!empty($errors)) {
                    $response['message'] .= ' (some errors: ' . implode('; ', array_slice($errors, 0, 5)) . ($totalErrors > 5 ? '...' : '') . ')';
                }
            } else {
                $response['message'] = 'Failed to update structure: ' . implode('; ', array_slice($errors, 0, 5)) . ($totalErrors > 5 ? '...' : '');
            }
            break;
            
        case 'run_sql':
            // Run SQL on selected databases
            $databaseIds = $inputData['database_ids'] ?? [];
            $sql = trim($inputData['sql'] ?? '');
            
            if (empty($databaseIds) || !is_array($databaseIds)) {
                $response['message'] = 'Database IDs are required';
                break;
            }
            
            if (empty($sql)) {
                $response['message'] = 'SQL statement is required';
                break;
            }
            
            // Get database credentials from config.php
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            // Get selected databases
            $placeholders = implode(',', array_fill(0, count($databaseIds), '?'));
            $getStmt = $conn->prepare("SELECT id, db_name FROM dbs WHERE id IN ($placeholders)");
            $getStmt->execute($databaseIds);
            $selectedDbs = $getStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($selectedDbs)) {
                $response['message'] = 'No databases found for the selected IDs';
                break;
            }
            
            $results = [];
            
            // Execute SQL on each database
            foreach ($selectedDbs as $db) {
                if (empty($db['db_name'])) {
                    $results[] = [
                        'db_id' => $db['id'],
                        'db_name' => 'Unknown',
                        'error' => 'Database name is not set'
                    ];
                    continue;
                }
                
                try {
                    $targetConn = new PDO(
                        "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                        $targetDbUser,
                        $targetDbPass
                    );
                    $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                    
                    // Determine if this is a SELECT query or other type
                    $sqlUpper = strtoupper(trim($sql));
                    $isSelect = strpos($sqlUpper, 'SELECT') === 0;
                    
                    if ($isSelect) {
                        // For SELECT queries, fetch results
                        $stmt = $targetConn->prepare($sql);
                        $stmt->execute();
                        $rows = $stmt->fetchAll(PDO::FETCH_ASSOC);
                        $results[] = [
                            'db_id' => $db['id'],
                            'db_name' => $db['db_name'],
                            'rows' => $rows,
                            'row_count' => count($rows)
                        ];
                    } else {
                        // For other queries (INSERT, UPDATE, DELETE, etc.), get affected rows
                        $stmt = $targetConn->prepare($sql);
                        $stmt->execute();
                        $affectedRows = $stmt->rowCount();
                        $results[] = [
                            'db_id' => $db['id'],
                            'db_name' => $db['db_name'],
                            'affected_rows' => $affectedRows
                        ];
                    }
                } catch (PDOException $e) {
                    $results[] = [
                        'db_id' => $db['id'],
                        'db_name' => $db['db_name'],
                        'error' => $e->getMessage()
                    ];
                }
            }
            
            $response['success'] = true;
            $response['message'] = 'SQL executed on ' . count($results) . ' database(s)';
            $response['data'] = $results;
            break;
            
        case 'check_assets':
            // Check which asset files exist in files_dir
            $filesDir = trim($inputData['files_dir'] ?? '');
            
            if (empty($filesDir)) {
                $response['message'] = 'files_dir is required';
                break;
            }
            
            try {
                // Remove leading slash
                $filesDir = ltrim($filesDir, '/');
                $filesDir = rtrim($filesDir, '/');
                
                // Construct full path
                $filesDirPath = __DIR__ . '/../' . $filesDir . '/';
                
                // Security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve and verify files_dir path
                $resolvedFilesDir = realpath($filesDirPath);
                if ($resolvedFilesDir === false) {
                    // Directory doesn't exist
                    $response['success'] = true;
                    $response['data'] = ['existing' => []];
                    break;
                }
                $resolvedFilesDir = rtrim($resolvedFilesDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFilesDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                // Check which files exist
                $assetFiles = ['logo.png', 'letter_head.png', 'gml2.png'];
                $existingFiles = [];
                
                foreach ($assetFiles as $fileName) {
                    $filePath = $resolvedFilesDir . $fileName;
                    if (file_exists($filePath) && is_file($filePath)) {
                        $existingFiles[] = $fileName;
                    }
                }
                
                $response['success'] = true;
                $response['data'] = ['existing' => $existingFiles];
            } catch (Exception $e) {
                $response['message'] = $e->getMessage();
            }
            break;
            
        case 'copy_assets':
            // Copy asset files (logo.png, letter_head.png, gml2.png) from js_dir to files_dir
            $jsDir = trim($inputData['js_dir'] ?? '');
            $filesDir = trim($inputData['files_dir'] ?? '');
            
            if (empty($jsDir) || empty($filesDir)) {
                $response['message'] = 'js_dir and files_dir are required';
                break;
            }
            
            try {
                // Remove leading slashes
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                $filesDir = ltrim($filesDir, '/');
                $filesDir = rtrim($filesDir, '/');
                
                // Construct full paths
                $jsDirPath = __DIR__ . '/../' . $jsDir . '/';
                $filesDirPath = __DIR__ . '/../' . $filesDir . '/';
                
                // Security check - get real base path
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve and verify js_dir path
                $resolvedJsDir = realpath($jsDirPath);
                if ($resolvedJsDir === false) {
                    throw new Exception('js_dir does not exist: ' . $jsDir);
                }
                $resolvedJsDir = rtrim($resolvedJsDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedJsDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected in js_dir');
                }
                
                // Ensure files_dir exists
                if (!file_exists($filesDirPath)) {
                    if (!mkdir($filesDirPath, 0755, true)) {
                        throw new Exception('Failed to create files_dir: ' . $filesDirPath);
                    }
                }
                
                // Resolve and verify files_dir path
                $resolvedFilesDir = realpath($filesDirPath);
                if ($resolvedFilesDir === false) {
                    throw new Exception('Failed to resolve files_dir path');
                }
                $resolvedFilesDir = rtrim($resolvedFilesDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFilesDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected in files_dir');
                }
                
                // Files to copy
                $assetFiles = ['logo.png', 'letter_head.png', 'gml2.png'];
                $copiedFiles = [];
                $errors = [];
                
                foreach ($assetFiles as $fileName) {
                    $sourceFile = $resolvedJsDir . $fileName;
                    $destFile = $resolvedFilesDir . $fileName;
                    
                    // Check if source file exists
                    if (!file_exists($sourceFile)) {
                        $errors[] = "Source file not found: {$fileName}";
                        continue;
                    }
                    
                    // Copy file
                    if (!copy($sourceFile, $destFile)) {
                        $errors[] = "Failed to copy: {$fileName}";
                        continue;
                    }
                    
                    // Set permissions
                    chmod($destFile, 0644);
                    
                    $copiedFiles[] = $fileName;
                }
                
                if (count($errors) > 0) {
                    $response['message'] = 'Some files failed to copy: ' . implode(', ', $errors);
                    $response['data'] = [
                        'copied' => $copiedFiles,
                        'errors' => $errors
                    ];
                } else {
                    $response['success'] = true;
                    $response['message'] = 'All files copied successfully';
                    $response['data'] = [
                        'copied' => $copiedFiles,
                        'files_dir' => $filesDir
                    ];
                }
            } catch (Exception $e) {
                $response['message'] = $e->getMessage();
            }
            break;
            
        case 'ensure_folder':
            // Ensure folder exists: create if doesn't exist (without clearing contents)
            $jsDir = trim($inputData['js_dir'] ?? '');
            
            if (empty($jsDir)) {
                $response['message'] = 'js_dir is required';
                break;
            }
            
            try {
                // Remove leading slash if present (Unix path format, but we use it relative to project root)
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Construct the full folder path (relative to project root)
                $folderPath = __DIR__ . '/../' . $jsDir . '/';
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve the folder path
                $resolvedFolderPath = realpath($folderPath);
                if ($resolvedFolderPath === false) {
                    // Folder doesn't exist, create it
                    if (!mkdir($folderPath, 0755, true)) {
                        throw new Exception('Failed to create folder: ' . $folderPath);
                    }
                    $resolvedFolderPath = realpath($folderPath);
                    if ($resolvedFolderPath === false) {
                        throw new Exception('Failed to resolve created folder path');
                    }
                }
                
                // Verify the resolved path is within the base directory
                $resolvedFolderPath = rtrim($resolvedFolderPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFolderPath, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                $response['success'] = true;
                $response['message'] = 'Folder ensured successfully';
                $response['data'] = [
                    'path' => $resolvedFolderPath
                ];
            } catch (Exception $e) {
                $response['message'] = $e->getMessage();
            }
            break;
            
        case 'prepare_upload_folder':
            // Prepare upload folder: create if doesn't exist, clear if exists
            $databaseId = intval($inputData['database_id'] ?? 0);
            $jsDir = trim($inputData['js_dir'] ?? '');
            
            if ($databaseId <= 0 || empty($jsDir)) {
                $response['message'] = 'Invalid database ID or js_dir';
                break;
            }
            
            try {
                // Remove leading slash if present (Unix path format, but we use it relative to project root)
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Construct the full folder path (relative to project root)
                $folderPath = __DIR__ . '/../' . $jsDir . '/';
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve the folder path
                $resolvedFolderPath = realpath($folderPath);
                if ($resolvedFolderPath === false) {
                    // Folder doesn't exist, create it
                    if (!mkdir($folderPath, 0755, true)) {
                        throw new Exception('Failed to create folder: ' . $folderPath);
                    }
                    $resolvedFolderPath = realpath($folderPath);
                    if ($resolvedFolderPath === false) {
                        throw new Exception('Failed to resolve created folder path');
                    }
                }
                
                // Verify the resolved path is within the base directory
                $resolvedFolderPath = rtrim($resolvedFolderPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFolderPath, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                // Protected files that should NOT be deleted
                $protectedFiles = ['logo.png', 'logo_default.png', 'letter_head.png', 'letter_head_default.png', 'gml2.png', 'tenant config'];
                
                // Delete all contents (files and subdirectories) EXCEPT protected files
                $iterator = new RecursiveIteratorIterator(
                    new RecursiveDirectoryIterator($resolvedFolderPath, RecursiveDirectoryIterator::SKIP_DOTS),
                    RecursiveIteratorIterator::CHILD_FIRST
                );
                
                $deletedFiles = [];
                $skippedFiles = [];
                
                foreach ($iterator as $path) {
                    $fileName = $path->getFilename();
                    $relativePath = str_replace($resolvedFolderPath, '', $path->getPathname());
                    
                    // Skip protected files (only check files in root of js_dir, not subdirectories)
                    if (!$path->isDir() && strpos($relativePath, DIRECTORY_SEPARATOR) === false) {
                        if (in_array($fileName, $protectedFiles)) {
                            $skippedFiles[] = $fileName;
                            continue; // Skip deleting protected files
                        }
                    }
                    
                    // Delete the file or directory
                    if ($path->isDir()) {
                        rmdir($path->getPathname());
                    } else {
                        unlink($path->getPathname());
                        $deletedFiles[] = $fileName;
                    }
                }
                
                // Log what was deleted and what was preserved
                error_log('[prepare_upload_folder] Deleted files: ' . implode(', ', $deletedFiles));
                if (count($skippedFiles) > 0) {
                    error_log('[prepare_upload_folder] Preserved protected files: ' . implode(', ', $skippedFiles));
                }
                
                $response['success'] = true;
                $message = 'Folder prepared successfully';
                if (count($skippedFiles) > 0) {
                    $message .= '. Preserved protected files: ' . implode(', ', $skippedFiles);
                }
                $response['message'] = $message;
                $response['data'] = [
                    'deleted_count' => count($deletedFiles),
                    'preserved_files' => $skippedFiles
                ];
            } catch (Exception $e) {
                $response['message'] = 'Error preparing folder: ' . $e->getMessage();
            }
            break;
            
        case 'read_db_code_json':
            // Read tenant config file from js_dir
            $databaseId = intval($inputData['database_id'] ?? 0);
            
            if ($databaseId <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            try {
                // Get database record
                $stmt = $conn->prepare("SELECT js_dir FROM dbs WHERE id = ?");
                $stmt->execute([$databaseId]);
                $db = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$db) {
                    $response['message'] = 'Database not found';
                    break;
                }
                
                $jsDir = trim($db['js_dir'] ?? '');
                // An empty js_dir means the app root (see resolveDbCodeJsonPath).
                $filePath = resolveDbCodeJsonPath($jsDir);
                if ($filePath === null) {
                    $response['message'] = 'Invalid js_dir for this database';
                    break;
                }
                
                // Debug logging
                error_log('[read_db_code_json] Database ID: ' . $databaseId);
                error_log('[read_db_code_json] js_dir from DB: ' . $jsDir);
                error_log('[read_db_code_json] __DIR__: ' . __DIR__);
                error_log('[read_db_code_json] Constructed filePath: ' . $filePath);
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                error_log('[read_db_code_json] realBasePath: ' . $realBasePath);
                
                // Resolve the file path
                $resolvedFilePath = realpath($filePath);
                error_log('[read_db_code_json] resolvedFilePath: ' . ($resolvedFilePath !== false ? $resolvedFilePath : 'FALSE'));
                
                // Also check if file exists using the constructed path directly
                $fileExistsDirect = file_exists($filePath);
                error_log('[read_db_code_json] file_exists($filePath): ' . ($fileExistsDirect ? 'TRUE' : 'FALSE'));
                
                // Verify the resolved path is within the base directory
                if ($resolvedFilePath !== false) {
                    $resolvedDir = dirname($resolvedFilePath);
                    if (strpos($resolvedDir, $realBasePath) !== 0) {
                        throw new Exception('Directory traversal attempt detected');
                    }
                }
                
                // Read file if exists, otherwise return default structure
                // Use both checks: realpath might fail even if file exists (e.g., symlinks)
                if (($resolvedFilePath !== false && file_exists($resolvedFilePath)) || $fileExistsDirect) {
                    // Use resolved path if available, otherwise use constructed path
                    $actualFilePath = $resolvedFilePath !== false ? $resolvedFilePath : $filePath;
                    error_log('[read_db_code_json] Reading file from: ' . $actualFilePath);
                    
                    $fileContent = file_get_contents($actualFilePath);
                    error_log('[read_db_code_json] File content length: ' . strlen($fileContent));
                    error_log('[read_db_code_json] File content: ' . substr($fileContent, 0, 200));
                    
                    $jsonData = json_decode($fileContent, true);
                    
                    if (json_last_error() !== JSON_ERROR_NONE) {
                        error_log('[read_db_code_json] JSON decode error: ' . json_last_error_msg());
                        throw new Exception('Invalid JSON in tenant config: ' . json_last_error_msg());
                    }
                    
                    error_log('[read_db_code_json] Parsed JSON data: ' . json_encode($jsonData));
                    
                    $response['success'] = true;
                    $response['message'] = 'File read successfully';
                    $response['data'] = [
                        'content' => $jsonData,
                        'exists' => true
                    ];
                } else {
                    // File doesn't exist, return default structure
                    error_log('[read_db_code_json] File not found. resolvedFilePath: ' . ($resolvedFilePath !== false ? $resolvedFilePath : 'FALSE') . ', fileExistsDirect: ' . ($fileExistsDirect ? 'TRUE' : 'FALSE'));
                    $response['success'] = true;
                    $response['message'] = 'File does not exist, returning default structure';
                    $response['data'] = [
                        'content' => ['db_code' => ''],
                        'exists' => false,
                        'debug' => [
                            'js_dir' => $jsDir,
                            'constructed_path' => $filePath,
                            'resolved_path' => $resolvedFilePath !== false ? $resolvedFilePath : null,
                            'file_exists_direct' => $fileExistsDirect,
                            'base_path' => $realBasePath
                        ]
                    ];
                }
            } catch (Exception $e) {
                $response['message'] = 'Error reading file: ' . $e->getMessage();
            }
            break;
            
        case 'write_db_code_json':
            // Write tenant config file to js_dir
            $databaseId = intval($inputData['database_id'] ?? 0);
            $jsonContent = $inputData['content'] ?? null;
            
            if ($databaseId <= 0) {
                $response['message'] = 'Invalid database ID';
                break;
            }
            
            if ($jsonContent === null) {
                $response['message'] = 'JSON content is required';
                break;
            }
            
            try {
                // Validate JSON content
                if (!is_array($jsonContent)) {
                    throw new Exception('Content must be a valid JSON object');
                }
                
                // Get database record
                $stmt = $conn->prepare("SELECT js_dir FROM dbs WHERE id = ?");
                $stmt->execute([$databaseId]);
                $db = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$db) {
                    $response['message'] = 'Database not found';
                    break;
                }
                
                $jsDir = trim($db['js_dir'] ?? '');
                // An empty js_dir means the app root (see resolveDbCodeJsonPath).
                $filePath = resolveDbCodeJsonPath($jsDir);
                if ($filePath === null) {
                    $response['message'] = 'Invalid js_dir for this database';
                    break;
                }
                $dirPath = dirname($filePath);
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Create directory if it doesn't exist
                if (!is_dir($dirPath)) {
                    if (!mkdir($dirPath, 0755, true)) {
                        throw new Exception('Failed to create directory: ' . $dirPath);
                    }
                }
                
                // Verify the directory path is within the base directory
                $resolvedDirPath = realpath($dirPath);
                if ($resolvedDirPath === false) {
                    throw new Exception('Failed to resolve directory path');
                }
                $resolvedDirPath = rtrim($resolvedDirPath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedDirPath, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                // Encode JSON with pretty printing
                $jsonString = json_encode($jsonContent, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                
                if ($jsonString === false) {
                    throw new Exception('Failed to encode JSON: ' . json_last_error_msg());
                }
                
                // Write file
                if (file_put_contents($filePath, $jsonString) === false) {
                    throw new Exception('Failed to write file');
                }
                
                $response['success'] = true;
                $response['message'] = 'File written successfully';
                $response['data'] = [
                    'path' => $jsDir . '/tenant config'
                ];
            } catch (Exception $e) {
                $response['message'] = 'Error writing file: ' . $e->getMessage();
            }
            break;
            
        case 'check_api_files_exist':
            // Report which of these names already exist in the api folder, so the
            // build uploader can warn before overwriting config.php and friends.
            //
            // Moved here from api/api.php, where it was gated on the *app* token
            // while being driven by the *db-manager* realm - a combination that
            // left the caller unable to present any credential it held. Its real
            // damage was not the 401 though: Databases.vue treated `success: false`
            // as "nothing exists" rather than as "I could not find out", so the
            // overwrite prompt was skipped entirely and protected files were
            // replaced silently. The caller now also fails safe; see
            // confirmProtectedOverwrite() there.
            if (!isset($inputData['file_names']) || !is_array($inputData['file_names'])) {
                $response['message'] = 'file_names array is required';
                break;
            }

            $apiDir = __DIR__;

            try {
                // Resolve the api directory once and prove it is inside the app root,
                // so the containment check below is comparing real paths rather
                // than a string that merely looks contained.
                $realBasePath = realpath(__DIR__ . '/..');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;

                $resolvedApiDir = realpath($apiDir);
                if ($resolvedApiDir === false) {
                    throw new Exception('Invalid api directory');
                }
                $resolvedApiDir = rtrim($resolvedApiDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedApiDir, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }

                $results = [];
                foreach ($inputData['file_names'] as $fileName) {
                    if (!is_string($fileName)) {
                        continue;
                    }

                    // basename() alone collapses any path component; the containment
                    // re-check afterwards is what actually decides the answer, since
                    // it runs against realpath() rather than the requested string.
                    $sanitized = basename($fileName);
                    if ($sanitized === '' || $sanitized === '.' || $sanitized === '..') {
                        continue;
                    }

                    $resolvedFilePath = realpath($apiDir . '/' . $sanitized);
                    $exists = false;
                    if ($resolvedFilePath !== false) {
                        $resolvedFileDir = dirname($resolvedFilePath);
                        $resolvedFileDir = rtrim($resolvedFileDir, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                        if (strpos($resolvedFileDir, $resolvedApiDir) === 0) {
                            $exists = is_file($resolvedFilePath);
                        }
                    }

                    // Keyed by the name the caller asked about, which is what the
                    // caller filters on.
                    $results[$fileName] = $exists;
                }

                $response['success'] = true;
                $response['message'] = 'Checked ' . count($results) . ' file(s)';
                $response['data'] = $results;
            } catch (Exception $e) {
                $response['message'] = 'Error checking files: ' . $e->getMessage();
            }
            break;

        case 'check_file_exists':
            // Check if a file exists in js_dir
            $databaseId = intval($inputData['database_id'] ?? 0);
            $fileName = trim($inputData['file_name'] ?? '');
            $jsDir = trim($inputData['js_dir'] ?? '');
            
            if ($databaseId <= 0 || empty($fileName)) {
                $response['message'] = 'Invalid database ID or file name';
                break;
            }
            
            if (empty($jsDir)) {
                // Get js_dir from database if not provided
                $stmt = $conn->prepare("SELECT js_dir FROM dbs WHERE id = ?");
                $stmt->execute([$databaseId]);
                $db = $stmt->fetch(PDO::FETCH_ASSOC);
                if (!$db || empty($db['js_dir'])) {
                    $response['message'] = 'js_dir is not configured for this database';
                    break;
                }
                $jsDir = trim($db['js_dir']);
            }
            
            try {
                // Remove leading slash if present
                $jsDir = ltrim($jsDir, '/');
                $jsDir = rtrim($jsDir, '/');
                
                // Sanitize filename
                $fileName = basename($fileName);
                $fileName = str_replace('..', '', $fileName);
                
                // Construct the full file path
                $filePath = __DIR__ . '/../' . $jsDir . '/' . $fileName;
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Resolve the file path
                $resolvedFilePath = realpath($filePath);
                
                // Verify the resolved path is within the base directory
                if ($resolvedFilePath !== false) {
                    $resolvedDir = dirname($resolvedFilePath);
                    if (strpos($resolvedDir, $realBasePath) !== 0) {
                        throw new Exception('Directory traversal attempt detected');
                    }
                }
                
                $fileExists = file_exists($filePath);
                
                $response['success'] = true;
                $response['message'] = $fileExists ? 'File exists' : 'File does not exist';
                $response['data'] = [
                    'exists' => $fileExists,
                    'file_path' => $jsDir . '/' . $fileName
                ];
            } catch (Exception $e) {
                $response['message'] = 'Error checking file: ' . $e->getMessage();
            }
            break;
            
        case 'backup_databases':
            // Export selected tenant databases as SQL (FK checks off at start, on at end)
            $databaseIds = $inputData['database_ids'] ?? [];
            
            if (empty($databaseIds) || !is_array($databaseIds)) {
                $response['message'] = 'Database IDs are required';
                break;
            }
            
            require_once __DIR__ . '/config.php';
            $targetDbHost = $db_config['host'];
            $targetDbUser = $db_config['user'];
            $targetDbPass = $db_config['pass'];
            
            $placeholders = implode(',', array_fill(0, count($databaseIds), '?'));
            $getStmt = $conn->prepare("SELECT id, db_name, is_created FROM dbs WHERE id IN ($placeholders)");
            $getStmt->execute($databaseIds);
            $selectedDbs = $getStmt->fetchAll(PDO::FETCH_ASSOC);
            
            if (empty($selectedDbs)) {
                $response['message'] = 'No databases found for the selected IDs';
                break;
            }
            
            $backups = [];
            $errors = [];
            
            foreach ($selectedDbs as $db) {
                if (empty($db['db_name'])) {
                    $errors[] = "Database ID {$db['id']}: Database name is not set";
                    continue;
                }
                if (intval($db['is_created']) !== 1) {
                    $errors[] = "Database {$db['db_name']}: Tables have not been created yet";
                    continue;
                }
                
                try {
                    $targetConn = new PDO(
                        "mysql:host={$targetDbHost};dbname={$db['db_name']}",
                        $targetDbUser,
                        $targetDbPass
                    );
                    $targetConn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
                    
                    $body = buildDatabaseBackupBody($targetConn);
                    $dbSafe = preg_replace('/[^a-zA-Z0-9_-]/', '_', $db['db_name']);
                    $fileTimestamp = date('Y-m-d_H-i-s');
                    $backups[] = [
                        'db_name' => $db['db_name'],
                        'filename' => "backup_{$dbSafe}_{$fileTimestamp}.sql",
                        'body' => $body,
                        'sql' => wrapDatabaseBackupSql($db['db_name'], $body),
                    ];
                } catch (PDOException $e) {
                    $errors[] = "Database {$db['db_name']}: " . $e->getMessage();
                }
            }
            
            if (empty($backups)) {
                $response['message'] = 'Backup failed: ' . implode('; ', $errors);
                break;
            }
            
            $timestamp = date('Y-m-d_H-i-s');
            
            if (count($backups) === 1) {
                sendBackupFileDownload($backups[0]['sql'], $backups[0]['filename']);
            }
            
            if (class_exists('ZipArchive')) {
                $zip = new ZipArchive();
                $tmpFile = tempnam(sys_get_temp_dir(), 'dbbackup_');
                if ($zip->open($tmpFile, ZipArchive::CREATE | ZipArchive::OVERWRITE) === true) {
                    foreach ($backups as $backup) {
                        $zip->addFromString($backup['filename'], $backup['sql']);
                    }
                    $zip->close();
                    $zipContent = file_get_contents($tmpFile);
                    unlink($tmpFile);
                    if ($zipContent !== false) {
                        sendBackupFileDownload(
                            $zipContent,
                            "databases_backup_{$timestamp}.zip",
                            'application/zip'
                        );
                    }
                }
                if (isset($tmpFile) && file_exists($tmpFile)) {
                    unlink($tmpFile);
                }
            }
            
            // Fallback: one combined SQL file
            $combined = [
                '-- Merhab Cars Database Backup',
                '-- Generated on: ' . date('Y-m-d H:i:s'),
                '-- Databases: ' . implode(', ', array_column($backups, 'db_name')),
                '',
                'SET NAMES utf8mb4;',
                'SET FOREIGN_KEY_CHECKS=0;',
                '',
            ];
            foreach ($backups as $backup) {
                $combined[] = '-- ========== Database: ' . $backup['db_name'] . ' ==========';
                $combined[] = '';
                $combined[] = $backup['body'];
                $combined[] = '';
            }
            $combined[] = 'SET FOREIGN_KEY_CHECKS=1;';
            $combined[] = '';
            
            sendBackupFileDownload(
                implode("\n", $combined),
                "databases_backup_{$timestamp}.sql"
            );
            break;
            
        default:
            $response['message'] = 'No action specified';
            break;
    }
    
} catch (Exception $e) {
    error_log('DB Manager API Error: ' . $e->getMessage());
    $response['message'] = $e->getMessage();
    $response['success'] = false;
    http_response_code(500);
} catch (Error $e) {
    error_log('DB Manager API Fatal Error: ' . $e->getMessage());
    $response['message'] = 'Fatal error: ' . $e->getMessage();
    $response['success'] = false;
    http_response_code(500);
}

// Ensure headers are sent (re-send CORS headers in case of error)
if (!headers_sent()) {
    api_send_cors_headers();
    header('Content-Type: application/json');
}

// Clean output buffer and send JSON response
if (ob_get_level()) {
    ob_end_clean();
}
echo json_encode($response);
exit;

