<?php
// API token verification.
//
// The app has no server-side session. LoginView used to fetch the user row
// (password hash included), verify it in the browser and keep the result in
// localStorage, and api.php trusted whatever the client claimed - it never read
// the `token` field useApi sends. A token minted by the login action is the
// first credential this codebase can actually check server-side, so endpoints
// that spend money or return bulk data gate on these helpers.
//
// The guards below prefer apiErrorDie() from api.php when it is loaded, so the
// response shape stays identical there. Standalone endpoints (upload.php,
// db_manager_api.php, invitations.php, ...) get the same JSON shape from
// api_auth_fail() instead, which is why these helpers are usable outside api.php.

/**
 * Emit a guard failure and end the request.
 *
 * Delegates to apiErrorDie() when api.php is the entry point, so the response
 * shape is byte-identical there. Standalone files have no apiErrorDie(), and
 * silently skipping the exit would leave the endpoint unauthenticated, so they
 * fall back to the same `{success:false, code, error}` envelope.
 */
function api_auth_fail(string $code): void
{
    if (function_exists('apiErrorDie')) {
        apiErrorDie($code);
    }

    if (!headers_sent()) {
        header('Content-Type: application/json');
    }
    echo json_encode(['success' => false, 'code' => $code, 'error' => $code]);
    exit;
}

/**
 * Resolve the caller from their API token.
 *
 * @return array{id:int,username:string,role_id:int}|null null when the token is
 *         missing or unknown. A null here means "unauthenticated", never
 *         "not allowed" - the caller decides what an anonymous request gets.
 */
function api_token_user($conn, $token): ?array
{
    if (!is_string($token)) {
        return null;
    }

    $token = trim($token);
    if ($token === '') {
        return null;
    }

    try {
        $stmt = $conn->prepare('SELECT id, username, role_id FROM users WHERE api_token = ? LIMIT 1');
        $stmt->execute([$token]);
        $row = $stmt->fetch(PDO::FETCH_ASSOC);
    } catch (Exception $e) {
        // A missing api_token column (migration not applied yet) must read as
        // "no token", not as a 500 that leaks the SQL error to the client.
        error_log('api_token_user: ' . $e->getMessage());
        return null;
    }

    return $row ? $row : null;
}

/**
 * Gate an action on a valid token, whatever role it belongs to. Exits the
 * request with not_authenticated when there is not one.
 *
 * The status stays 200, as it does everywhere else in api.php: useApi retries
 * any 403/429 three times with backoff (it was written to get past bot
 * protection) and then throws an opaque "HTTP 401" string, so a real 4xx here
 * would turn one clear message into a 7-second retry storm. The client
 * distinguishes the outcome from the `code` field either way.
 *
 * This is the floor for anything that reads or writes a row: it is strictly more
 * than the caller sending *a* token, because the token has to resolve to a real
 * user. Note it says nothing about permission - an endpoint serving reference
 * data every user needs should ask for this, not require_api_admin().
 *
 * @return array{id:int,username:string,role_id:int} the authenticated user
 */
function require_api_user($conn, array $postData): array
{
    $user = api_token_user($conn, $postData['token'] ?? '');
    if (!$user) {
        api_auth_fail('not_authenticated');
    }

    return $user;
}

/**
 * Gate an action on a valid token belonging to an admin. Exits the request with
 * not_authenticated / not_admin when it does not hold.
 *
 * @return array{id:int,username:string,role_id:int} the authenticated admin
 */
function require_api_admin($conn, array $postData): array
{
    $user = require_api_user($conn, $postData);

    if ((int) $user['role_id'] !== 1) {
        api_auth_fail('not_admin');
    }

    return $user;
}
