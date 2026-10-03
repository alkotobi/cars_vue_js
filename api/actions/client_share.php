<?php
// Read-only client share page data, keyed by share_token.
//
// ClientDetailsView is the one authenticated-looking screen the router exempts:
// it loads from /clients/:token so a client can be shown their own cars from a
// link they were sent, with nobody logged in. Its five reads went through the
// caller-supplied-SQL passthrough with requiresAuth: false, which meant anyone
// holding any share token - or no token at all - could run arbitrary SQL against
// the database. Closing the passthrough means these have to live here instead.
//
// Deliberately public: a share_token is the credential. It is high entropy (see
// clients.share_token), it is not guessable, and possession of it is exactly the
// authority to see this one client's cars. Nothing beyond that client's own rows
// is reachable - the client is looked up by token and every subsequent query is
// keyed off the resulting id, never off client input.
//
// This handler is read-only by construction: it contains no INSERT/UPDATE/DELETE,
// so a mistake in a future edit is visible in the diff.
//
// Errors leave here as apiErrorDie() codes; the client maps `code` to a locale
// key. Depends on getConnection(), getDbConfig() and apiErrorDie() from api.php.

require_once __DIR__ . '/containers.php';

/** A share token longer than this is not one we minted. */
const SHARE_MAX_TOKEN = 255;

/**
 * Everything ClientDetailsView needs, in one round trip.
 *
 * Previously five sequential calls: resolve token to client id, read the client,
 * read their cars, read those cars' files, read those cars' tracking. The client
 * id is resolved once here and every query hangs off it.
 */
function handle_get_client_share_data(array $postData): void
{
    $conn = share_data_connection();

    $token = share_clean_token((string) ($postData['share_token'] ?? ''));
    if ($token === '') {
        apiErrorDie('share_token_required');
    }

    // Lookup is by token alone; the token *is* the credential.
    $client = share_find_client($conn, $token);
    if (!$client) {
        apiErrorDie('share_token_invalid');
    }

    $clientId = (int) $client['id'];

    echo json_encode([
        'success' => true,
        'client' => $client,
        'cars' => share_cars($conn, $clientId),
        'files' => share_files($conn, $clientId),
        'tracking' => containers_tracking_rows($conn, share_container_refs($conn, $clientId)),
    ]);
    exit;
}

/**
 * @return array<string,mixed>|null null when the token matches no client
 */
function share_find_client($conn, string $token): ?array
{
    $stmt = $conn->prepare(
        'SELECT c.id, c.name, c.email, c.mobiles, c.id_no, c.address, c.is_broker,
                COUNT(cs.id) as cars_count
         FROM clients c
         LEFT JOIN cars_stock cs ON c.id = cs.id_client
         WHERE c.share_token = ?
         GROUP BY c.id'
    );
    $stmt->execute([$token]);

    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    if (!$row) {
        return null;
    }

    return [
        'id' => (int) $row['id'],
        'name' => $row['name'],
        'email' => $row['email'],
        'mobiles' => $row['mobiles'],
        'id_no' => $row['id_no'],
        'address' => $row['address'],
        'is_broker' => $row['is_broker'],
        'cars_count' => (int) $row['cars_count'],
    ];
}

/**
 * The client's own cars, visible ones only.
 *
 * total_paid is what the client has already paid, from sell_bill/sell_payments.
 * Kept because the page renders it.
 *
 * @return array<int,array<string,mixed>>
 */
function share_cars($conn, int $clientId): array
{
    $stmt = $conn->prepare(
        "SELECT cs.*,
                w.warhouse_name as warehouse_name,
                bd.amount as buy_price,
                bd.year,
                bd.is_used_car,
                bd.is_big_car,
                cn.car_name as model,
                b.brand,
                c.color,
                dp.discharge_port,
                cs.container_ref,
                COALESCE(
                  (SELECT SUM(sp.amount_da)
                   FROM sell_bill sb
                   JOIN sell_payments sp ON sp.id_sell_bill = sb.id
                   WHERE sb.id = cs.id_sell),
                  0
                ) as total_paid,
                CASE
                  WHEN cs.id_client IS NOT NULL THEN 'Reserved'
                  ELSE 'Available'
                END as status
         FROM cars_stock cs
         LEFT JOIN warehouses w ON cs.id_warehouse = w.id
         LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
         LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
         LEFT JOIN brands b ON cn.id_brand = b.id
         LEFT JOIN colors c ON bd.id_color = c.id
         LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
         WHERE cs.id_client = ? AND cs.hidden = 0
         ORDER BY cs.in_wharhouse_date DESC, cs.id DESC"
    );
    $stmt->execute([$clientId]);

    return $stmt->fetchAll(PDO::FETCH_ASSOC);
}

/**
 * Documents belonging to the client's cars, flattened with car_id.
 *
 * The client grouped these by car_id itself; doing it here saves the client a
 * pass and keeps the grouping rule in one place.
 *
 * @return array<int,array<string,mixed>>
 */
function share_files($conn, int $clientId): array
{
    $stmt = $conn->prepare(
        "SELECT cf.id, cf.car_id, cf.category_id, cf.file_path, cf.file_name,
                cf.file_size, cf.file_type, cf.uploaded_at,
                cfc.category_name, cfc.display_order
         FROM car_files cf
         INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
         INNER JOIN cars_stock cs ON cs.id = cf.car_id
         WHERE cs.id_client = ? AND cs.hidden = 0 AND cf.is_active = 1
         ORDER BY cf.car_id, cfc.display_order, cf.uploaded_at DESC"
    );
    $stmt->execute([$clientId]);

    $files = [];
    foreach ($stmt->fetchAll(PDO::FETCH_ASSOC) as $row) {
        $files[] = [
            'id' => (int) $row['id'],
            'car_id' => (int) $row['car_id'],
            'category_id' => (int) $row['category_id'],
            'file_path' => $row['file_path'],
            'file_name' => $row['file_name'],
            'file_size' => $row['file_size'],
            'file_type' => $row['file_type'],
            'uploaded_at' => $row['uploaded_at'],
            'category_name' => $row['category_name'],
            'display_order' => (int) $row['display_order'],
        ];
    }

    return $files;
}

/**
 * @return string[] container refs of this client's cars
 */
function share_container_refs($conn, int $clientId): array
{
    $stmt = $conn->prepare(
        "SELECT DISTINCT container_ref FROM cars_stock
         WHERE id_client = ? AND hidden = 0
           AND container_ref IS NOT NULL AND TRIM(container_ref) <> ''"
    );
    $stmt->execute([$clientId]);

    $refs = [];
    foreach ($stmt->fetchAll(PDO::FETCH_COLUMN) as $ref) {
        $clean = containers_clean_ref((string) $ref);
        if ($clean !== '') {
            $refs[] = $clean;
        }
    }

    return $refs;
}

function share_clean_token(string $token): string
{
    $token = trim(preg_replace('/\s+/u', '', $token) ?? '');

    if (mb_strlen($token) > SHARE_MAX_TOKEN) {
        return '';
    }

    return $token;
}

function share_data_connection(): PDO
{
    $conn = getConnection(getDbConfig());
    if (is_array($conn) && isset($conn['error'])) {
        error_log('share_data_connection: ' . $conn['error']);
        apiErrorDie('db_unavailable');
    }

    return $conn;
}