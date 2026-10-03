<?php
// App/database version comparison, for the mismatch dialog.
//
// Public by necessity: DatabaseVersionCheck is mounted unconditionally in App.vue,
// so this query ran on /login, before anyone logged in, and had to stay working
// after the caller-supplied-SQL passthrough was closed. What it returns is a single
// version string that the dialog already shows to anyone loading the login page.
//
// Key order is fixed (id DESC) so this returns the same row the old
// "SELECT version FROM versions ORDER BY id DESC LIMIT 1" did.

/**
 * @return array{success:bool,version:?int}
 */
function handle_get_db_version(): void
{
    $conn = getConnection(getDbConfig());
    if (is_array($conn) && isset($conn['error'])) {
        error_log('handle_get_db_version: ' . $conn['error']);
        apiErrorDie('db_unavailable');
    }

    $stmt = $conn->query('SELECT version FROM versions ORDER BY id DESC LIMIT 1');
    $version = $stmt ? $stmt->fetchColumn() : false;

    // versions.version is an INT and useVersionCheck compares it with `!==`
    // against a JS number, so this has to come back as a number - returning the
    // raw value would hand back a string and report a mismatch forever.
    echo json_encode([
        'success' => true,
        'version' => $version === false || $version === null ? null : (int) $version,
    ]);
    exit;
}
