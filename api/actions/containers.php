<?php
// Container GPS tracking: where a container is, and recording where it went.
//
// This module replaces four call sites that used the `execute_sql` action, which
// ran caller-supplied SQL with no authentication at all. `executeQuery()`
// prepares and executes a statement BEFORE reading its first six characters, so
// the SELECT/INSERT/UPDATE/DELETE switch only shaped the JSON response: a DROP,
// or an UPDATE against users, took effect and then answered "Invalid query
// type". One of those four call sites was in ClientDetailsView, which the router
// exempts from authentication for share-token clients, so arbitrary SQL was
// reachable without logging in at all.
//
// The four statements involved are all here now, and nothing else is:
//
//   - get_container_tracking   reads tracking for some or all containers.
//                              Replaces two calls plus an N+1 loop, and is the
//                              reason the read is safe to expose broadly: it
//                              selects only tracking columns, never users.
//   - save_container_tracking   the one write, an upsert on container_ref.
//
// Requires a valid token for both. The tracking table holds no secrets beyond a
// container's last known position, but it is staff-entered data about real
// freight, and the share-token page reads it through this same handler - so it is
// gated on identity rather than on being harmless.
//
// Errors leave here as apiErrorDie() codes; the client maps `code` to a locale
// key. Depends on getConnection(), getDbConfig() and apiErrorDie() from api.php,
// and require_api_user() from lib/.

require_once __DIR__ . '/../lib/auth.php';

/** A container_ref longer than this is not a container reference. */
const CONTAINERS_MAX_REF = 255;

/** An upper bound on how many refs one request may ask about. */
const CONTAINERS_MAX_REFS = 500;

/**
 * Latest known position per container.
 *
 * The container list is DISTINCT container_ref from cars_stock, LEFT JOINed to
 * the most recent tracking row. That is what ContainersRefList.vue wanted and
 * spent an N+1 loop assembling; it is also what a caller passing a subset of refs
 * gets, since the WHERE clause narrows the same query.
 *
 * @param array<string,mixed> $postData optional container_refs: string[]
 */
function handle_get_container_tracking(array $postData): void
{
    $conn = containers_connection();
    require_api_user($conn, $postData);

    $refs = isset($postData['container_refs']) ? containers_clean_refs($postData['container_refs']) : [];
    if (isset($postData['container_refs']) && !$refs) {
        echo json_encode(['success' => true, 'tracking' => []]);
        exit;
    }

    echo json_encode(['success' => true, 'tracking' => containers_tracking_rows($conn, $refs)]);
    exit;
}

/**
 * Latest known position per container, optionally narrowed to a set of refs.
 *
 * The container list is DISTINCT container_ref from cars_stock, LEFT JOINed to
 * the most recent tracking row. That is what ContainersRefList.vue wanted and
 * spent an N+1 loop assembling. With $refs empty it covers every container; with
 * a list it narrows to those.
 *
 * Shared with client_share.php, whose share-token page needs the same lookup for
 * the cars it is already entitled to see - that page cannot call
 * handle_get_container_tracking() itself, which requires a signed-in user.
 *
 * Selects only tracking columns and the operator's display name. Nothing from
 * users other than username crosses this boundary.
 *
 * @param string[] $refs
 * @return array<int,array<string,mixed>>
 */
function containers_tracking_rows($conn, array $refs = []): array
{
    $where = '';
    $params = [];

    if ($refs) {
        $where = 'WHERE cs.container_ref IN (' . implode(',', array_fill(0, count($refs), '?')) . ')';
        $params = $refs;
    }

    // The correlated subquery picks the newest row per container_ref. The join to
    // users is for the operator's name only.
    $stmt = $conn->prepare(
        "SELECT cs.container_ref, t.tracking, t.time, t.id_user, u.username
         FROM (SELECT DISTINCT container_ref FROM cars_stock
               WHERE container_ref IS NOT NULL AND TRIM(container_ref) <> '') cs
         LEFT JOIN tracking t ON t.id = (
             SELECT id FROM tracking WHERE container_ref = cs.container_ref
             ORDER BY time DESC, id DESC LIMIT 1
         )
         LEFT JOIN users u ON u.id = t.id_user
         $where
         ORDER BY cs.container_ref ASC"
    );
    $stmt->execute($params);

    $rows = [];
    foreach ($stmt->fetchAll(PDO::FETCH_ASSOC) as $row) {
        $rows[] = [
            'container_ref' => (string) $row['container_ref'],
            'tracking' => $row['tracking'],
            'time' => $row['time'],
            'id_user' => $row['id_user'] !== null ? (int) $row['id_user'] : null,
            'username' => $row['username'],
        ];
    }

    return $rows;
}

/**
 * Record where a container is. One row per container_ref (UNIQUE), so this is an
 * upsert rather than an append.
 *
 * id_user comes from the token, never from the request. The caller used to send a
 * literal 1 with a comment saying it should come from the session, which recorded
 * every save against whichever user happened to hold id 1.
 */
function handle_save_container_tracking(array $postData): void
{
    $conn = containers_connection();
    $user = require_api_user($conn, $postData);

    $ref = containers_clean_ref((string) ($postData['container_ref'] ?? ''));
    if ($ref === '') {
        apiErrorDie('container_ref_required');
    }

    // "lat,lng" as written by the map. Validated here because this column is fed
    // straight into parseFloat() by three different screens.
    $coords = trim((string) ($postData['tracking'] ?? ''));
    if ($coords !== '' && !preg_match('/^-?\d{1,3}(?:\.\d+)?\s*,\s*-?\d{1,3}(?:\.\d+)?$/', $coords)) {
        apiErrorDie('container_tracking_invalid');
    }

    $latLong = array_map('floatval', array_map('trim', explode(',', $coords)));
    if (count($latLong) === 2) {
        if (abs($latLong[0]) > 90 || abs($latLong[1]) > 180) {
            apiErrorDie('container_tracking_invalid');
        }
    } elseif ($coords !== '') {
        apiErrorDie('container_tracking_invalid');
    }

    try {
        $stmt = $conn->prepare(
            'INSERT INTO tracking (container_ref, tracking, time, id_user)
             VALUES (?, ?, NOW(), ?)
             ON DUPLICATE KEY UPDATE
                tracking = VALUES(tracking), time = VALUES(time), id_user = VALUES(id_user)'
        );
        $stmt->execute([$ref, $coords === '' ? null : $coords, (int) $user['id']]);
    } catch (PDOException $e) {
        error_log('handle_save_container_tracking: ' . $e->getMessage());
        apiErrorDie('container_tracking_save_failed');
    }

    echo json_encode(['success' => true, 'container_ref' => $ref]);
    exit;
}

/**
 * @param mixed $raw the container_refs field: a list, or a single string
 * @return string[]
 */
function containers_clean_refs($raw): array
{
    if (is_string($raw)) {
        $raw = [$raw];
    }

    if (!is_array($raw)) {
        return [];
    }

    $refs = [];
    foreach ($raw as $value) {
        if (!is_scalar($value)) {
            continue;
        }
        $ref = containers_clean_ref((string) $value);
        if ($ref !== '' && !in_array($ref, $refs, true)) {
            $refs[] = $ref;
        }
        if (count($refs) >= CONTAINERS_MAX_REFS) {
            break;
        }
    }

    return $refs;
}

/**
 * Bound and de-duplicate a whitespace value.
 *
 * Bound because these become bound parameters, but also because container_ref is
 * a lookup key the URL bar can be made to carry.
 */
function containers_clean_ref(string $ref): string
{
    $ref = trim(preg_replace('/\s+/u', ' ', $ref) ?? '');

    if (mb_strlen($ref) > CONTAINERS_MAX_REF) {
        return '';
    }

    return $ref;
}

function containers_connection(): PDO
{
    $conn = getConnection(getDbConfig());
    if (is_array($conn) && isset($conn['error'])) {
        error_log('containers_connection: ' . $conn['error']);
        apiErrorDie('db_unavailable');
    }

    return $conn;
}