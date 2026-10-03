<?php
// User creation and password changes.
//
// These three handlers replace two more caller-supplied-SQL endpoints:
//
//   insert_user   POSTed a whole `INSERT INTO users (...) VALUES (?, ...)` string
//                 and expected the server to hash params[2]. Reachable by anyone.
//   hash_password POSTed a whole `UPDATE users SET password = ? WHERE ...` string.
//                 Also reachable by anyone - and because LoginView's caller passed
//                 `WHERE username = ?` straight through, any caller could rewrite
//                 any account's password, including the seeded admin's.
//
// Both were named actions, which is what made them look narrower than the raw
// passthrough. They were not: the action name said which columns to hash, the
// payload said which rows to touch.
//
// The write target is now a function argument - a user id for set_user_password, or
// the token's own user for change_own_password - so it cannot be widened by editing
// the query.
//
// Errors leave here as apiErrorDie() codes; the client maps `code` to a locale key.
// Depends on getConnection(), getDbConfig() and apiErrorDie() from api.php, and
// require_api_user() / require_api_admin() from lib/.

require_once __DIR__ . '/../lib/auth.php';

/** bcrypt truncates past 72 bytes; reject rather than silently ignore the tail. */
const USERS_MAX_PASSWORD_BYTES = 72;
const USERS_MIN_PASSWORD_LENGTH = 8;

/**
 * Create a user. Admin only.
 *
 * The INSERT used to arrive as a query string; the columns are fixed here.
 */
function handle_create_user(array $postData): void
{
    $conn = users_connection();
    require_api_admin($conn, $postData);

    $username = trim((string) ($postData['username'] ?? ''));
    $email = trim((string) ($postData['email'] ?? ''));

    if ($username === '' || mb_strlen($username) > 50) {
        apiErrorDie('user_username_invalid');
    }
    if ($email === '' || mb_strlen($email) > 100 || !filter_var($email, FILTER_VALIDATE_EMAIL)) {
        apiErrorDie('user_email_invalid');
    }

    $password = users_check_password((string) ($postData['password'] ?? ''));

    $roleId = (int) ($postData['role_id'] ?? 0);
    if ($roleId <= 0 || !users_role_exists($conn, $roleId)) {
        apiErrorDie('user_role_invalid');
    }

    $bankId = $postData['id_bank_account'] ?? null;
    $bankId = ($bankId === null || $bankId === '' || (int) $bankId <= 0) ? null : (int) $bankId;

    try {
        $stmt = $conn->prepare(
            'INSERT INTO users (username, email, password, role_id, max_unpayed_created_bills,
                                is_diffrent_company, id_bank_account)
             VALUES (?, ?, ?, ?, ?, ?, ?)'
        );
        $stmt->execute([
            $username,
            $email,
            users_hash_password($password),
            $roleId,
            max(0, (int) ($postData['max_unpayed_created_bills'] ?? 0)),
            !empty($postData['is_diffrent_company']) ? 1 : 0,
            $bankId,
        ]);
    } catch (PDOException $e) {
        // username and email are both UNIQUE.
        if ($e->getCode() === '23000') {
            apiErrorDie('user_exists');
        }
        error_log('handle_create_user: ' . $e->getMessage());
        apiErrorDie('user_save_failed');
    }

    echo json_encode(['success' => true, 'id' => (int) $conn->lastInsertId()]);
    exit;
}

/**
 * Change the caller's own password, proving they know the current one.
 *
 * The target row is the token's user. There is no username in this request, so
 * there is nothing for a caller to widen the write to.
 */
function handle_change_own_password(array $postData): void
{
    $conn = users_connection();
    $user = require_api_user($conn, $postData);

    $newPassword = users_check_password((string) ($postData['new_password'] ?? ''));
    $currentPassword = (string) ($postData['current_password'] ?? '');

    if ($currentPassword === '') {
        apiErrorDie('user_current_password_required');
    }

    $stmt = $conn->prepare('SELECT password FROM users WHERE id = ?');
    $stmt->execute([(int) $user['id']]);
    $hash = $stmt->fetchColumn();

    if (!is_string($hash) || !password_verify($currentPassword, $hash)) {
        apiErrorDie('user_current_password_wrong');
    }

    users_write_password($conn, (int) $user['id'], $newPassword);

    // The old token is not invalidated here: the client does that itself by logging
    // out and back in, which is why it now has to re-hold a token to get this far.
    echo json_encode(['success' => true]);
    exit;
}

/**
 * Set another user's password. Admin only.
 *
 * This is the one the old hash_password action served for EditUserForm, where an
 * admin edits someone else. It takes a user id, not a username or a query.
 */
function handle_set_user_password(array $postData): void
{
    $conn = users_connection();
    require_api_admin($conn, $postData);

    $userId = (int) ($postData['user_id'] ?? 0);
    if ($userId <= 0) {
        apiErrorDie('user_id_required');
    }

    if (!users_exists($conn, $userId)) {
        apiErrorDie('user_not_found');
    }

    $password = users_check_password((string) ($postData['password'] ?? ''));

    users_write_password($conn, $userId, $password);

    echo json_encode(['success' => true]);
    exit;
}

/**
 * Change a password from the login screen, before there is a session.
 *
 * Public by necessity: LoginView toggles this form while logged out, so there is no
 * token to check. Knowing the current password is the authorisation, and it is
 * checked here against the stored hash.
 *
 * This replaces a three-call client-side flow that did the opposite: it selected
 * `u.password` and shipped the bcrypt hash to the browser, POSTed it back to
 * `verify_password` to be compared there, and then POSTed
 * `UPDATE users SET password = ? WHERE username = ?` with `username` taken from the
 * form. All three were reachable anonymously, so anyone could rewrite any account's
 * password - and the hash itself was disclosed to whoever asked. The username is
 * still caller-supplied (it has to be, it is the login screen), but it now decides
 * only which hash to verify, never which row to write.
 *
 * Errors for a wrong username and a wrong password are deliberately the same code,
 * so this cannot be used to enumerate accounts.
 */
function handle_change_password_with_credentials(array $postData): void
{
    $conn = users_connection();

    $username = trim((string) ($postData['username'] ?? ''));
    $currentPassword = (string) ($postData['current_password'] ?? '');
    $newPassword = users_check_password((string) ($postData['new_password'] ?? ''));

    if ($username === '' || $currentPassword === '') {
        apiErrorDie('user_credentials_required');
    }

    $stmt = $conn->prepare('SELECT id, password FROM users WHERE username = ?');
    $stmt->execute([$username]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);

    // password_verify against a dummy hash when there is no such user, so a missing
    // username and a wrong password take the same time.
    $hash = is_array($row) && is_string($row['password'])
        ? $row['password']
        : '$2y$10$usesomesillystringfore7hnbRJHxXVLeakoG8K30oukPsA.ztMG';

    if (!password_verify($currentPassword, $hash) || !is_array($row)) {
        apiErrorDie('user_credentials_wrong');
    }

    users_write_password($conn, (int) $row['id'], $newPassword);

    echo json_encode(['success' => true]);
    exit;
}

/**
 * @throws RuntimeException via apiErrorDie on a bad password
 */
function users_check_password(string $password): string
{
    if (strlen($password) < USERS_MIN_PASSWORD_LENGTH) {
        apiErrorDie('user_password_too_short');
    }
    // password_hash() would silently cut at 72 bytes, leaving the tail ignored.
    if (strlen($password) > USERS_MAX_PASSWORD_BYTES) {
        apiErrorDie('user_password_too_long');
    }

    return $password;
}

function users_hash_password(string $password): string
{
    return password_hash($password, PASSWORD_DEFAULT);
}

/**
 * Rotate the stored hash and mint a fresh api_token.
 *
 * A new token because the old one is about to stop matching: after a password
 * change the client logs out, but any copy of the previous token would otherwise
 * keep working against this account.
 */
function users_write_password(PDO $conn, int $userId, string $password): void
{
    try {
        $stmt = $conn->prepare(
            'UPDATE users SET password = ?, api_token = ? WHERE id = ?'
        );
        $stmt->execute([users_hash_password($password), users_new_token(), $userId]);
    } catch (PDOException $e) {
        error_log('users_write_password: ' . $e->getMessage());
        apiErrorDie('user_save_failed');
    }
}

function users_new_token(): string
{
    return bin2hex(random_bytes(32));
}

function users_exists(PDO $conn, int $userId): bool
{
    $stmt = $conn->prepare('SELECT 1 FROM users WHERE id = ?');
    $stmt->execute([$userId]);

    return (bool) $stmt->fetchColumn();
}

function users_role_exists(PDO $conn, int $roleId): bool
{
    $stmt = $conn->prepare('SELECT 1 FROM roles WHERE id = ?');
    $stmt->execute([$roleId]);

    return (bool) $stmt->fetchColumn();
}

function users_connection(): PDO
{
    $conn = getConnection(getDbConfig());
    if (is_array($conn) && isset($conn['error'])) {
        error_log('users_connection: ' . $conn['error']);
        apiErrorDie('db_unavailable');
    }

    return $conn;
}
