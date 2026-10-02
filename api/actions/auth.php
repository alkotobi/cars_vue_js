<?php
// Login and logout: the one place the server verifies a password.
//
// LoginView used to fetch the user row through the generic query endpoint,
// including the bcrypt hash, and call the `verify_password` action from the
// browser - so anyone could read every user's hash through api.php and compare
// offline. The `login` action replaces that: the password is checked here,
// never leaves the server, and the answer is a token the client can present on
// calls that need a real credential.
//
// apiTokenUser() in lib/auth.php checks that token. It is not a session: the
// token stays valid until logout, a new login, or a password change.
//
// Depends on getConnection(), getDbConfig() and apiErrorDie() from api.php.

/** A valid bcrypt hash of a value nobody knows, used to keep failed logins slow. */
const AUTH_DUMMY_HASH = '$2y$10$usesomesillystringfore7hnbRJHxXVLeakoG8K30oukPsA.ztMG';

/**
 * Verify credentials and mint a token.
 *
 * @param array<string,mixed> $postData
 */
function handle_login(array $postData): void
{
    $username = trim((string) ($postData['username'] ?? ''));
    $password = (string) ($postData['password'] ?? '');

    if ($username === '' || $password === '') {
        apiErrorDie('invalid_credentials');
    }

    $conn = auth_connection();

    $stmt = $conn->prepare(
        'SELECT u.id, u.username, u.role_id, u.password, u.is_diffrent_company, u.path_logo,
                u.path_letter_head, u.path_stamp, u.path_contract_terms, u.id_bank_account,
                u.max_unpayed_created_bills, r.role_name
         FROM users u
         LEFT JOIN roles r ON r.id = u.role_id
         WHERE u.username = ?
         LIMIT 1'
    );
    $stmt->execute([$username]);
    $user = $stmt->fetch(PDO::FETCH_ASSOC);

    // Hash against a dummy when the username is unknown: a missing user and a
    // wrong password then take the same time, so the response cannot be used to
    // enumerate accounts.
    $hash = $user['password'] ?? AUTH_DUMMY_HASH;
    $valid = password_verify($password, $hash);
    if (!$user || !$valid) {
        apiErrorDie('invalid_credentials');
    }

    // A new token replaces any previous one: one live token per user, and a
    // second login quietly invalidates the first.
    $token = bin2hex(random_bytes(32));
    $update = $conn->prepare('UPDATE users SET api_token = ? WHERE id = ?');
    $update->execute([$token, $user['id']]);

    $permissions = auth_permissions($conn, (int) $user['role_id']);

    // The hash is deliberately absent from the response: the client has no use
    // for it now that verification happens here.
    unset($user['password']);

    echo json_encode([
        'success' => true,
        'token' => $token,
        'user' => $user + ['permissions' => $permissions],
    ]);
    exit;
}

/**
 * Drop the caller's token. Missing or unknown tokens are still a success: the
 * caller's intent - hold no token - is satisfied either way.
 *
 * @param array<string,mixed> $postData
 */
function handle_logout(array $postData): void
{
    $token = trim((string) ($postData['token'] ?? ''));
    if ($token !== '') {
        $conn = auth_connection();
        $stmt = $conn->prepare('UPDATE users SET api_token = NULL WHERE api_token = ?');
        $stmt->execute([$token]);
    }

    echo json_encode(['success' => true]);
    exit;
}

/**
 * The rows the login response needs, in the shape the app already stores:
 * a list of { permission_name, description }.
 *
 * @return array<int,array<string,string>>
 */
function auth_permissions($conn, int $roleId): array
{
    $stmt = $conn->prepare(
        'SELECT p.permission_name, p.description
         FROM permissions p
         JOIN role_permissions rp ON p.id = rp.permission_id
         WHERE rp.role_id = ?
         ORDER BY p.id'
    );
    $stmt->execute([$roleId]);

    return $stmt->fetchAll(PDO::FETCH_ASSOC);
}

function auth_connection(): PDO
{
    $conn = getConnection(getDbConfig());
    if (is_array($conn) && isset($conn['error'])) {
        error_log('auth_connection: ' . $conn['error']);
        apiErrorDie('db_unavailable');
    }

    return $conn;
}
