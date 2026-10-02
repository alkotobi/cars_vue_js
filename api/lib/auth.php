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
// Loaded from api.php only: the guards below call apiErrorDie() from there, and
// say so loudly rather than duplicating the response shape.

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
 * Gate an action on a valid token belonging to an admin. Exits the request with
 * not_authenticated / not_admin when it does not hold.
 *
 * The status stays 200, as it does everywhere else in api.php: useApi retries
 * any 403/429 three times with backoff (it was written to get past bot
 * protection) and then throws an opaque "HTTP 401" string, so a real 4xx here
 * would turn one clear message into a 7-second retry storm. The client
 * distinguishes the outcome from the `code` field either way.
 *
 * @return array{id:int,username:string,role_id:int} the authenticated admin
 */
function require_api_admin($conn, array $postData): array
{
    if (!function_exists('apiErrorDie')) {
        throw new RuntimeException('require_api_admin() must be called from api.php');
    }

    $user = api_token_user($conn, $postData['token'] ?? '');
    if (!$user) {
        apiErrorDie('not_authenticated');
    }

    if ((int) $user['role_id'] !== 1) {
        apiErrorDie('not_admin');
    }

    return $user;
}
