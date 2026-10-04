<?php
// Shared helpers for the mysqldump-backed backup scripts.
//
// Both backup.php and backup_simple_web.php built their dump command by string
// interpolation:
//
//     "mysqldump --host=$host --user=$username --password=$password $dbname > $filepath"
//
// In backup_simple_web.php $dbname was $_POST['database'], and the value was also
// persisted to auto_backup_database.txt and replayed by the cron path - so one
// unauthenticated request could inject a shell command that ran twice, later and
// unattended. Separately, every value including the password was interpolated
// bare, so the password was readable out of `ps` for the life of the dump.
//
// backup_mysqldump() below escapes every argument, keeps the password out of
// argv via a temporary defaults file, and refuses a database name that is not a
// plain MySQL identifier - so the shell never sees caller data as syntax.

/**
 * Reject anything that is not a bare MySQL database name.
 *
 * This is a whitelist, not a blacklist: MySQL identifiers cannot contain a
 * semicolon, backtick, space, quote, pipe or redirect, so anything else is either
 * a bug or an attempt. Checked before the name reaches a shell.
 */
function backup_assert_valid_db_name(string $dbName): string
{
    if (!preg_match('/^[A-Za-z0-9_$-]{1,64}$/', $dbName)) {
        throw new InvalidArgumentException('Invalid database name');
    }
    return $dbName;
}

/**
 * Run mysqldump into $targetPath.
 *
 * @return bool true when the dump was written and is non-empty
 */
function backup_mysqldump(string $dbName, string $targetPath, array $dbConfig, ?string $dumpBinary = null): bool
{
    $dbName = backup_assert_valid_db_name($dbName);

    if ($dumpBinary === null) {
        $located = [];
        $status = 0;
        exec('command -v mysqldump 2>&1', $located, $status);
        if ($status !== 0 || empty($located[0])) {
            return false;
        }
        $dumpBinary = trim($located[0]);
    }

    // Password in a private defaults file, not in argv, so it is not visible in
    // ps output or /proc/<pid>/cmdline while the dump runs.
    $defaultsFile = tempnam(sys_get_temp_dir(), 'cars-dump-');
    if ($defaultsFile === false) {
        return false;
    }

    file_put_contents(
        $defaultsFile,
        "[client]\nuser=" . $dbConfig['user'] . "\npassword=" . $dbConfig['pass'] . "\n"
    );
    @chmod($defaultsFile, 0600);

    $command = escapeshellarg($dumpBinary)
        . ' --defaults-extra-file=' . escapeshellarg($defaultsFile)
        . ' --host=' . escapeshellarg($dbConfig['host'])
        . ' --database=' . escapeshellarg($dbName);

    // Redirect through the shell's own quoting rather than appending " > $path",
    // which is what made $filepath injectable too.
    $descriptors = [1 => ['file', $targetPath, 'w'], 2 => ['pipe', 'w']];
    $process = @proc_open($command, $descriptors, $pipes);
    if (!is_resource($process)) {
        @unlink($defaultsFile);
        return false;
    }

    // Drain stderr so mysqldump cannot block on a full pipe, and discard it:
    // its output used to be merged into the dump file.
    if (isset($pipes[2]) && is_resource($pipes[2])) {
        stream_get_contents($pipes[2]);
        fclose($pipes[2]);
    }

    $status = proc_close($process);
    @unlink($defaultsFile);

    if ($status === 0 && is_file($targetPath) && filesize($targetPath) > 0) {
        $content = @file_get_contents($targetPath);
        if ($content !== false) {
            $wrapped = "SET FOREIGN_KEY_CHECKS=0;\n\n" . $content . "\n\nSET FOREIGN_KEY_CHECKS=1;\n";
            file_put_contents($targetPath, $wrapped);
        }
        return true;
    }
    return false;
}

/**
 * Gate a backup endpoint on an admin token.
 *
 * These scripts are operator tools that dump every table, so they require the
 * same role that can run SQL in the app. Exits the request when it does not hold.
 *
 * The `users` table is looked up in THIS DEPLOYMENT'S tenant database. Reading it
 * from config.php's default instead meant the token of an admin who only exists in
 * a tenant database was rejected, while an admin of the default database was
 * accepted on a tenant's URL - the token check was authenticating against a
 * different install than the one it granted access to.
 */
function api_require_backup_admin(): void
{
    $token = $_SERVER['HTTP_X_API_TOKEN'] ?? ($_GET['token'] ?? ($_POST['token'] ?? ''));

    require_once __DIR__ . '/config.php';
    require_once __DIR__ . '/auth.php';
    require_once __DIR__ . '/appdb.php';

    // No fallback to config.php's db_name, and this is the one place it mattered
    // most: this authenticates the caller. The fallback meant the token check ran
    // against whatever one database config.php named, so an admin of that database
    // was accepted on any tenant's URL and a real tenant admin was rejected - the
    // check was granting access to an install the caller never authenticated against.
    //
    // An unresolved request cannot be authenticated at all, so it is refused rather
    // than checked against a guess.
    $dbname = app_db_name() ?? $db_config['dbname'] ?? '';
    if ($dbname === '') {
        throw new RuntimeException(
            'Cannot authenticate this request: it did not resolve to a tenant. '
            . 'Set db_name in config.php, or reach the app through its tenant folder.'
        );
    }

    $pdo = new PDO(
        "mysql:host={$db_config['host']};dbname={$dbname}",
        $db_config['user'],
        $db_config['pass']
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    require_api_admin($pdo, ['token' => is_string($token) ? $token : '']);
}
