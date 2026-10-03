<?php
// Colors reference data: the list behind every colour picker in the app.
//
// This module exists because the colours table used to be written through the
// generic {query, params} passthrough in api.php. That path executes whatever
// SQL it is handed and its only gate is a special case for payment_confirmed
// writes, so "is this user an admin?" was really answered by a v-if in the
// component, reading role_id out of localStorage that the user controls. Anyone
// could POST a DELETE FROM colors. Writes go through here instead, gated on the
// api_token minted at login - the only credential this codebase can actually
// verify server-side (see lib/auth.php).
//
// Reading the list needs a token but no particular role: colours are reference
// data that every user picks from while entering a car, and CarColorBulkEditForm
// and AddColorDialog both write here as ordinary users. Deleting one is
// admin-only, which is what the old UI already claimed, and is refused outright
// while any buy_details / cars_stock / priorities row still points at it.
//
// Errors leave here as apiErrorDie() codes, never raw English: the client maps
// `code` to a locale key. Depends on getConnection(), getDbConfig() and
// apiErrorDie() from api.php, and require_api_user() / require_api_admin() from
// lib/.

require_once __DIR__ . '/../lib/auth.php';

/** colors.color is varchar(255); anything longer is rejected, not truncated. */
const COLORS_MAX_NAME = 255;

/** #RRGGBB. The leading hash is optional on input and always present on output. */
const COLORS_HEX_PATTERN = '/^#?([0-9A-Fa-f]{6})$/';

/**
 * Every table holding a colour reference. Delete checks all of them.
 *
 * None of these three has a foreign key to colors in api/setup.sql - the
 * constraint only exists in the older mysqldump in sql/ - so a delete would
 * silently orphan rows on a fresh install and be rejected by MySQL on a restored
 * one. Counting them here makes the answer the same either way.
 */
const COLORS_REFERENCE_TABLES = [
    'buy_details' => 'buyDetails',
    'cars_stock' => 'carsStock',
    'priorities' => 'priorities',
];

/** @return array<int,array<string,mixed>> */
function handle_get_colors(array $postData): void
{
    $conn = colors_connection();
    require_api_user($conn, $postData);

    $stmt = $conn->query('SELECT id, color, hexa FROM colors ORDER BY color ASC');
    $colors = $stmt->fetchAll(PDO::FETCH_ASSOC);

    echo json_encode(['success' => true, 'colors' => $colors]);
    exit;
}

/**
 * @return int the new row id, mirrored as lastInsertId because AddColorDialog
 *         already reads that key off the generic path it used to call.
 */
function handle_create_color(array $postData): void
{
    $conn = colors_connection();
    require_api_user($conn, $postData);

    [$name, $hexa] = colors_require_fields($postData);

    if (colors_duplicate($conn, $name, $hexa, 0)) {
        apiErrorDie('color_exists', 200, ['field' => colors_duplicate_field($conn, $name, $hexa, 0)]);
    }

    try {
        $stmt = $conn->prepare('INSERT INTO colors (color, hexa) VALUES (?, ?)');
        $stmt->execute([$name, $hexa]);
    } catch (PDOException $e) {
        // Backstop for the race where two requests pass the check above together
        // and the UNIQUE index is what actually catches the second one. The index
        // is the authority; the SELECT is only there to give a readable message.
        if (colors_is_duplicate_key($e)) {
            apiErrorDie('color_exists', 200, ['field' => 'color']);
        }
        error_log('handle_create_color: ' . $e->getMessage());
        apiErrorDie('color_save_failed');
    }

    $id = (int) $conn->lastInsertId();

    echo json_encode([
        'success' => true,
        'id' => $id,
        'lastInsertId' => $id,
        'color' => $name,
        'hexa' => $hexa,
    ]);
    exit;
}

function handle_update_color(array $postData): void
{
    $conn = colors_connection();
    require_api_user($conn, $postData);

    $id = (int) ($postData['id'] ?? 0);
    if ($id <= 0) {
        apiErrorDie('color_not_found');
    }

    [$name, $hexa] = colors_require_fields($postData);

    if (!colors_exists($conn, $id)) {
        apiErrorDie('color_not_found');
    }

    // Excluded by id so re-saving a row unchanged is not a self-collision.
    if (colors_duplicate($conn, $name, $hexa, $id)) {
        apiErrorDie('color_exists', 200, ['field' => colors_duplicate_field($conn, $name, $hexa, $id)]);
    }

    try {
        $stmt = $conn->prepare('UPDATE colors SET color = ?, hexa = ? WHERE id = ?');
        $stmt->execute([$name, $hexa, $id]);
    } catch (PDOException $e) {
        if (colors_is_duplicate_key($e)) {
            apiErrorDie('color_exists', 200, ['field' => 'color']);
        }
        error_log('handle_update_color: ' . $e->getMessage());
        apiErrorDie('color_save_failed');
    }

    echo json_encode(['success' => true, 'id' => $id, 'color' => $name, 'hexa' => $hexa]);
    exit;
}

/**
 * Admin only, and only when nothing points at the row.
 *
 * The reference counts are the point of this handler: a colour is picked on a
 * buy detail, a car in stock and a priority band, and deleting one silently
 * would leave all three showing a blank colour.
 */
function handle_delete_color(array $postData): void
{
    $conn = colors_connection();
    require_api_admin($conn, $postData);

    $id = (int) ($postData['id'] ?? 0);
    if ($id <= 0 || !colors_exists($conn, $id)) {
        apiErrorDie('color_not_found');
    }

    $references = colors_reference_counts($conn, $id);
    $total = array_sum($references);
    if ($total > 0) {
        apiErrorDie('color_in_use', 200, [
            'count' => $total,
            'tables' => array_keys(array_filter($references)),
        ]);
    }

    $stmt = $conn->prepare('DELETE FROM colors WHERE id = ?');
    $stmt->execute([$id]);

    echo json_encode(['success' => true, 'id' => $id, 'deleted' => $stmt->rowCount()]);
    exit;
}

/**
 * Read and validate the submitted name and hex.
 *
 * @return array{0:string,1:?string} name, and hexa as null when not supplied
 */
function colors_require_fields(array $postData): array
{
    $name = trim((string) ($postData['color'] ?? ''));

    // Collapse internal runs of whitespace so "PEARLY  WHITE" cannot become a
    // second row that looks identical to "PEARLY WHITE" in every dropdown.
    $name = trim(preg_replace('/\s+/u', ' ', $name) ?? '');

    if ($name === '') {
        apiErrorDie('color_name_required');
    }

    if (mb_strlen($name) > COLORS_MAX_NAME) {
        apiErrorDie('color_name_too_long');
    }

    $rawHexa = $postData['hexa'] ?? null;
    if ($rawHexa === null || trim((string) $rawHexa) === '') {
        return [$name, null];
    }

    $hexa = trim((string) $rawHexa);
    if (!preg_match(COLORS_HEX_PATTERN, $hexa, $m)) {
        apiErrorDie('color_invalid_hexa');
    }

    // Uppercased so the same colour cannot be stored twice in different cases:
    // colors.hexa is UNIQUE under a case-insensitive collation, and the seeded
    // rows are lowercase, so normalising on write keeps new rows comparable.
    return [$name, '#' . strtoupper($m[1])];
}

/** @return bool whether the row exists */
function colors_exists($conn, int $id): bool
{
    $stmt = $conn->prepare('SELECT 1 FROM colors WHERE id = ? LIMIT 1');
    $stmt->execute([$id]);

    return (bool) $stmt->fetchColumn();
}

/**
 * Whether this name or hex is already taken by another row.
 *
 * Compared with SQL rather than in PHP so the answer uses the column's own
 * collation - the same rules the UNIQUE index will apply.
 */
function colors_duplicate($conn, string $name, ?string $hexa, int $excludeId): bool
{
    return colors_duplicate_field($conn, $name, $hexa, $excludeId) !== null;
}

/**
 * Which field collides, or null when neither does.
 *
 * Both are reported separately because colors.color and colors.hexa are two
 * independent unique keys and the message should say which one was refused.
 */
function colors_duplicate_field($conn, string $name, ?string $hexa, int $excludeId): ?string
{
    $stmt = $conn->prepare('SELECT 1 FROM colors WHERE color = ? AND id <> ? LIMIT 1');
    $stmt->execute([$name, $excludeId]);
    if ($stmt->fetchColumn()) {
        return 'color';
    }

    // A NULL hexa never collides: MySQL lets UNIQUE columns repeat NULL, which
    // is why the seeded default of '#000000' is not special-cased anywhere.
    if ($hexa !== null) {
        $stmt = $conn->prepare('SELECT 1 FROM colors WHERE hexa = ? AND id <> ? LIMIT 1');
        $stmt->execute([$hexa, $excludeId]);
        if ($stmt->fetchColumn()) {
            return 'hexa';
        }
    }

    return null;
}

/**
 * How many rows point at this colour, per referencing table.
 *
 * @return array<string,int> table name => count, missing tables omitted
 */
function colors_reference_counts($conn, int $id): array
{
    $counts = [];

    foreach (array_keys(COLORS_REFERENCE_TABLES) as $table) {
        try {
            $stmt = $conn->prepare("SELECT COUNT(*) FROM `{$table}` WHERE id_color = ?");
            $stmt->execute([$id]);
            $count = (int) $stmt->fetchColumn();
        } catch (PDOException $e) {
            // A server predating one of these columns must not be told to delete
            // a colour it cannot fully account for.
            error_log("colors_reference_counts({$table}): " . $e->getMessage());
            apiErrorDie('db_schema_outdated');
        }

        if ($count > 0) {
            $counts[$table] = $count;
        }
    }

    return $counts;
}

/** MySQL's duplicate-entry error, which the UNIQUE keys raise as SQLSTATE 23000. */
function colors_is_duplicate_key(PDOException $e): bool
{
    return $e->getCode() === '23000';
}

function colors_connection(): PDO
{
    $conn = getConnection(getDbConfig());
    if (is_array($conn) && isset($conn['error'])) {
        error_log('colors_connection: ' . $conn['error']);
        apiErrorDie('db_unavailable');
    }

    return $conn;
}