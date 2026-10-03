<?php
// CORS headers, restricted to same-origin.
//
// Every API file used to send `Access-Control-Allow-Origin: *`. That is not
// exploitable on its own - the app authenticates with a token in the request
// body rather than a cookie, so there is nothing ambient for a cross-origin page
// to ride. It does mean, though, that any page on the internet could drive these
// endpoints from a visitor's browser and read the replies, which turned every
// finding above into a drive-by: a backup dump or an arbitrary-SQL call needed no
// token from the attacker, only a victim who happened to visit a hostile page.
//
// The app is same-origin by construction - resolveApiBaseUrl() builds URLs from
// window.location, and the Vite dev server proxies /api - so the wildcard was
// never needed. Reflecting only an origin that matches this host's own keeps
// preflights working without granting anything to another site.

/**
 * Send the CORS headers for this request, if any.
 *
 * Reflects the request's Origin only when it matches the host being served, so
 * the response stays readable to same-origin callers and opaque to everyone else.
 * Sends nothing at all when there is no Origin header, which is the normal case
 * for a normal request.
 */
function api_send_cors_headers(): void
{
    if (headers_sent()) {
        return;
    }

    header('Vary: Origin');

    $origin = $_SERVER['HTTP_ORIGIN'] ?? '';
    if (!is_string($origin) || $origin === '') {
        return;
    }

    $parts = parse_url($origin);
    if ($parts === false || empty($parts['host'])) {
        return;
    }

    $selfHost = $_SERVER['HTTP_HOST'] ?? '';
    $selfHost = strtolower(explode(':', $selfHost)[0]);
    $originHost = strtolower($parts['host']);

    // Same host, and the same scheme when we can tell - a matching host over a
    // different scheme (http page reading an https API) is not something to grant.
    $originScheme = strtolower($parts['scheme'] ?? '');
    $selfScheme = (!empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off') ? 'https' : 'http';

    if ($originHost !== $selfHost || ($originScheme !== '' && $originScheme !== $selfScheme)) {
        return;
    }

    header('Access-Control-Allow-Origin: ' . $origin);
    header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
    header('Access-Control-Allow-Headers: Content-Type, Accept, X-Api-Token, Authorization, X-Requested-With');
    header('Access-Control-Expose-Headers: Content-Disposition');
    header('Access-Control-Max-Age: 86400');
}

/**
 * Answer a CORS preflight and end the request.
 */
function api_handle_preflight(): void
{
    if (($_SERVER['REQUEST_METHOD'] ?? '') !== 'OPTIONS') {
        return;
    }

    api_send_cors_headers();
    http_response_code(200);
    exit;
}
