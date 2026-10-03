<?php
// Report everything to the error log, but never to the response body.
//
// This file is only ever consumed by JSON.parse(), so a stray warning printed
// ahead of the payload breaks the whole response - the browser sees
// `Unexpected token '<'` and the actual cause is nowhere in the message. That is
// not hypothetical: a warning here hid behind exactly that error while the
// payment_confirmed gate was mis-classifying a read as a write.
//
// display_errors defaults to 1 on a lot of PHP builds, so this has to be set
// rather than assumed. error_reporting stays at E_ALL, so nothing is lost - it
// goes to error_log, which is where a production failure belongs. backup.php and
// upload.php already did this; api.php did not.
error_reporting(E_ALL);
ini_set('display_errors', '0');

require_once __DIR__ . '/lib/cors.php';
api_send_cors_headers();
header('Content-Type: application/json');

// Handle preflight OPTIONS request
api_handle_preflight();

// Only accept POST requests
if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    echo json_encode(['success' => false, 'error' => 'Only POST method is allowed']);
    exit;
}

// Include database configuration
require_once __DIR__ . '/config.php';

// require_api_user() / require_api_admin() guard the caller-supplied SQL endpoint
// at the bottom of this file, so the helpers have to exist from here on.
require_once __DIR__ . '/lib/auth.php';
// app_db_name() backs getDbConfig() and must be loaded before the first query.
require_once __DIR__ . '/lib/appdb.php';

// The decoded request body. Read once, at the top, because the action gate below,
// getDbConfig() and every handler in the switch all work from it.
//
// The is_array() guard is load-bearing, not defensive tidiness. A body that is not
// a JSON object - malformed, or a bare `123` / `"x"` / `null` - decodes to null or
// a scalar rather than throwing. `isset($postData['action'])` is then false, so the
// whole switch is skipped and control reaches require_api_user($conn, $postData)
// below, whose second parameter is a non-nullable `array`. That is an uncaught
// TypeError: a fatal, not an exception this file can shape into JSON. With
// display_errors off the client received an empty body and HTTP 500, which
// readJsonResponse() then reported as "returned an empty body" - naming neither
// the endpoint's contract nor the TypeError underneath it.
$postData = json_decode(file_get_contents('php://input'), true);
if (!is_array($postData)) {
    apiErrorDie('invalid_request', 400);
}

// Resolve the real database name for this deployment, server-side.
//
// @return string|null null when there is no db_code.json or it does not resolve, in
//         which case getDbConfig() falls back to config.php.
function resolveDbNameFromCode(): ?string
{
    // app_db_code() reads db_code.json and validates the shape; app_db_name() walks
    // the merhab_databases registry for it. Both live in lib/appdb.php so that this
    // file and every standalone endpoint resolve the same database - two code paths
    // here would be two chances to disagree about which database is being served.
    return app_db_name();
}

// Function to get database configuration
//
// The database this server serves is decided here, on the server, from the
// per-deployment db_code.json that deploy/deploy.sh writes. It used to come from
// the request instead: the browser read db_code.json (it sits in the webroot),
// asked db_manager_api.php to turn the code into a real name, and posted that name
// as `dbname`, which this function used verbatim. Anyone could post any name, so a
// caller could aim the API's own credentials at a database it was never meant to
// touch.
//
// The POST `dbname` is ignored. See resolveDbNameFromCode() below.
function getDbConfig() {
    static $resolved = null;

    // Cached: this runs on every query, and re-reading the file plus opening a
    // second connection per request would be absurd overhead for a value that
    // cannot change while the process is running.
    if ($resolved !== null) {
        return $resolved;
    }

    global $db_config;

    $dbName = resolveDbNameFromCode();
    if ($dbName !== null) {
        $resolved = [
            'host' => $db_config['host'],
            'user' => $db_config['user'],
            'pass' => $db_config['pass'],
            'dbname' => $dbName
        ];
        return $resolved;
    }

    // No db_code.json, or it does not resolve. Fall back to config.php rather than
    // failing every request: a developer running against a single database should
    // not need the per-server file that only a real deployment has.
    return $resolved = $db_config;
}

// Function to establish database connection
function getConnection($config) {
    try {
        $conn = new PDO(
            "mysql:host={$config['host']};dbname={$config['dbname']}", 
            $config['user'], 
            $config['pass']
        );
        $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
        return $conn;
    } catch(PDOException $e) {
        return ['error' => $e->getMessage()];
    }
}

// Strip the comments and whitespace that can legally precede the first SQL token.
//
// Used to decide whether a statement needs a permission check. This is not a SQL
// parser and does not need to be: the only question is "does this statement start
// with something other than UPDATE", and leading /* */, -- and # comments are the
// only way to move the verb away from byte 0 without changing what it means.
function api_strip_leading_sql($sql) {
    $sql = (string)$sql;
    $guard = 0;

    do {
        $before = $sql;
        $sql = ltrim($sql);
        $sql = preg_replace('/^\/\*.*?\*\//s', '', $sql);   // /* block comment */
        $sql = preg_replace('/^--[^\n]*/', '', $sql);      // -- line comment
        $sql = preg_replace('/^#[^\n]*/', '', $sql);       // # line comment
        $sql = ltrim($sql);
    } while ($sql !== $before && ++$guard < 100);

    return $sql;
}

// Function to execute SQL query and return appropriate result
function executeQuery($sql, $params = []) {
    try {
        $conn = getConnection(getDbConfig());
        
        if (is_array($conn) && isset($conn['error'])) {
            throw new Exception($conn['error']);
        }

        $stmt = $conn->prepare($sql);
        $stmt->execute($params);

        // Determine query type
        $queryType = strtoupper(substr(trim($sql), 0, 6));
        
        switch($queryType) {
            case 'SELECT':
                return ['success' => true, 'data' => $stmt->fetchAll(PDO::FETCH_ASSOC)];
            case 'INSERT':
                return ['success' => true, 'lastInsertId' => $conn->lastInsertId()];
            case 'UPDATE':
            case 'DELETE':
                return ['success' => true, 'affectedRows' => $stmt->rowCount()];
            default:
                return ['success' => false, 'error' => 'Invalid query type'];
        }
    } catch(Exception $e) {
        return ['success' => false, 'error' => $e->getMessage()];
    }
}

// Function to check if a user has a specific permission
function hasPermission($conn, $userId, $permissionName) {
    try {
        // Admins have all permissions
        $adminCheck = "SELECT role_id FROM users WHERE id = ?";
        $stmt = $conn->prepare($adminCheck);
        $stmt->execute([$userId]);
        $user = $stmt->fetch(PDO::FETCH_ASSOC);
        
        if ($user && $user['role_id'] == 1) {
            return true; // Admin has all permissions
        }
        
        // Check if user's role has the permission
        $permissionQuery = "SELECT COUNT(*) as cnt 
            FROM role_permissions rp
            INNER JOIN permissions p ON rp.permission_id = p.id
            INNER JOIN users u ON rp.role_id = u.role_id
            WHERE u.id = ? AND p.permission_name = ?";
        $stmt = $conn->prepare($permissionQuery);
        $stmt->execute([$userId, $permissionName]);
        $result = $stmt->fetch(PDO::FETCH_ASSOC);
        
        return ($result && $result['cnt'] > 0);
    } catch (Exception $e) {
        error_log("Error checking permission: " . $e->getMessage());
        return false;
    }
}

// Admin check derived from the database, not from a client-supplied flag. The
// rest of this file historically trusted $postData['is_admin'], which anyone can
// set to true. Buy detail edits/deletes are gated on this instead.
function isAdminUser($conn, $userId) {
    try {
        $stmt = $conn->prepare('SELECT role_id FROM users WHERE id = ?');
        $stmt->execute([$userId]);
        $user = $stmt->fetch(PDO::FETCH_ASSOC);
        return $user && (int)$user['role_id'] === 1;
    } catch (Exception $e) {
        error_log("Error checking admin: " . $e->getMessage());
        return false;
    }
}

// A car is "committed" the moment downstream workflow has touched it. Any such
// car must never be silently removed by editing or deleting its buy detail.
function isCommittedCarSql() {
    return "
        (cs.vin IS NOT NULL AND TRIM(cs.vin) <> '')
        OR (cs.id_sell IS NOT NULL)
        OR (cs.id_client IS NOT NULL)
        OR (cs.id_port_loading IS NOT NULL)
        OR (cs.id_port_discharge IS NOT NULL)
        OR (cs.date_loding IS NOT NULL)
        OR (cs.container_ref IS NOT NULL AND TRIM(cs.container_ref) <> '')
        OR (cs.id_loaded_container IS NOT NULL)
        OR (cs.date_assigned IS NOT NULL)
        OR (cs.payment_confirmed != 0)
        OR EXISTS (SELECT 1 FROM car_files cf WHERE cf.car_id = cs.id)
    ";
}

function countCommittedCars($conn, $detailId) {
    $stmt = $conn->prepare(
        'SELECT COUNT(*) AS cnt FROM cars_stock cs WHERE cs.id_buy_details = ? AND (' . isCommittedCarSql() . ')'
    );
    $stmt->execute([$detailId]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    return (int)$row['cnt'];
}

function countDetailCars($conn, $detailId) {
    $stmt = $conn->prepare('SELECT COUNT(*) AS cnt FROM cars_stock WHERE id_buy_details = ?');
    $stmt->execute([$detailId]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    return (int)$row['cnt'];
}

// Emit an error payload carrying a stable machine-readable `code` so the client
// can translate it into the active locale. Never include a raw English message
// when a code exists: the client maps `code` -> i18n key.
function apiErrorDie($code, $status = 200, $meta = null) {
    if ($status) {
        http_response_code($status);
    }
    $payload = ['success' => false, 'code' => $code, 'error' => $code];
    if ($meta !== null) {
        $payload['meta'] = $meta;
    }
    echo json_encode($payload);
    exit;
}

// Actions reachable without a session.
//
// Everything else needs a token. The list is a literal on purpose: it is the whole
// unauthenticated surface of this file, so it should be readable at a glance and
// should not be derivable from anything a caller controls. Anything genuinely
// public gets its own entry here rather than a flag on a gated path - a
// client-sent "public" marker would be a bypass for the gate.
const PUBLIC_ACTIONS = [
    // Cookie/bot verification, called before login.
    'ping',
    // The credential exchange itself.
    'login',
    // Self-service password change for a user who cannot log in yet. It requires
    // the current password, verified server-side.
    'change_password_with_credentials',
    // /clients/:token share link - the token in the path is the credential.
    'get_client_share_data',
    // Version dialog, mounted unconditionally in App.vue including on /login.
    'get_db_version',
];

// Check if this is a special action request
if (isset($postData['action'])) {
    // Everything else needs a session. A dozen actions in the switch below used to
    // decide what to do by reading $postData['is_admin'], which any caller can
    // simply set; gating the whole switch turns those checks into a question about
    // a real user rather than about a claim. A token verified server-side is the
    // only thing in this codebase that says who the caller is.
    //
    // $currentUser is the answer, and the only answer. Handlers used to re-derive
    // the caller from the payload - $postData['user_id'], $postData['is_admin'],
    // $postData['performed_by'], $postData['transferred_by'] - so a request could
    // present its own valid token *and* name somebody else, and the authorisation
    // check then passed for the name in the body. Read identity from $currentUser,
    // never from $postData.
    if (!in_array($postData['action'], PUBLIC_ACTIONS, true)) {
        $gateConn = getConnection(getDbConfig());
        if (is_array($gateConn) && isset($gateConn['error'])) {
            error_log('action gate: ' . $gateConn['error']);
            apiErrorDie('db_unavailable');
        }
        $currentUser = require_api_user($gateConn, $postData);
    }

    switch($postData['action']) {
        case 'ping':
            // Simple ping action for cookie verification
            echo json_encode(['success' => true, 'message' => 'pong']);
            exit;
            
        // Two arbitrary-statement endpoints used to sit here. One ran whatever
        // statement the request carried behind a semicolon check that
        // `SELECT 1; DROP TABLE users` slipped past; its sibling split on ';' and ran
        // the batch with no check at all. Neither looked at a token, so both were
        // reachable by anyone who could reach the server - and ClientDetailsView,
        // which the router exempts from authentication for share-token clients, called
        // the first one. Callers now use named actions in api/actions/; see
        // src/views/advancedSqlRemoval.spec.js.

        case 'get_unique_containers_ref':
            // Get all unique, non-null containers_ref from cars_stock table
            $query = "SELECT DISTINCT container_ref FROM cars_stock WHERE container_ref IS NOT NULL AND container_ref != '' ORDER BY container_ref ASC";
            $result = executeQuery($query);
            
            if ($result['success']) {
                echo json_encode(['success' => true, 'data' => $result['data']]);
            } else {
                echo json_encode(['success' => false, 'error' => $result['error']]);
            }
            exit;

        // verify_password, hash_password and insert_user used to sit here. They read
        // the statement off the payload - hash_password took
        // `UPDATE users SET password = ? WHERE username = ?` straight from the login
        // form - so "named action" bought nothing: any caller could rewrite any
        // account's password. verify_password was worse, answering "does this password
        // match this hash" for any pair, which is a free offline-cracking oracle for
        // anyone holding a stolen hash. The replacements are handle_create_user,
        // handle_change_own_password and handle_set_user_password in
        // actions/users.php, each gated on a token and, where relevant, an admin.

        case 'assign_multiple_vins':
            if (!isset($postData['assignments']) || !is_array($postData['assignments'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Assignments array is required']);
                exit;
            }
            
            $assignments = $postData['assignments'];
            $successCount = 0;
            $errors = [];
            
            try {
                $conn = getConnection(getDbConfig());
                
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Start transaction
                $conn->beginTransaction();
                
                foreach ($assignments as $assignment) {
                    if (!isset($assignment['carId']) || !array_key_exists('vin', $assignment)) {
                        $errors[] = 'Invalid assignment data';
                        continue;
                    }
                    
                    $carId = $assignment['carId'];
                    $vin = $assignment['vin'];
                    
                    // Update the car with the new VIN
                    $stmt = $conn->prepare('UPDATE cars_stock SET vin = ? WHERE id = ?');
                    $result = $stmt->execute([$vin, $carId]);
                    
                    if ($result) {
                        $successCount++;
                    } else {
                        $errors[] = "Failed to update car ID: $carId";
                    }
                }
                
                // Commit transaction if all updates were successful
                if (empty($errors)) {
                    $conn->commit();
                    echo json_encode([
                        'success' => true, 
                        'message' => "Successfully assigned VINs to $successCount cars",
                        'assignedCount' => $successCount
                    ]);
                } else {
                    // Rollback on any error
                    $conn->rollBack();
                    echo json_encode([
                        'success' => false, 
                        'error' => 'Some assignments failed',
                        'errors' => $errors,
                        'successCount' => $successCount
                    ]);
                }
                
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
                exit;

        case 'create_stock_from_details':
            // Expand a pending bill's buy_details into individual cars_stock rows.
            //
            // The whole expansion happens in one transaction: the previous client-side
            // loop issued one INSERT per QTY unit, so a failure partway through left a
            // pending bill holding partial stock, and a retry inserted the remainder on
            // top of it. Details are read here by bill id rather than trusted from the
            // client, so the inserted rows are derived from committed data.
            if (!isset($postData['bill_id'])) {
                apiErrorDie('invalid_request', 400);
            }

            $billId = intval($postData['bill_id']);
            if ($billId <= 0) {
                apiErrorDie('invalid_request', 400);
            }

            $createdCars = [];
            $inTransaction = false;

            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }

                $conn->beginTransaction();
                $inTransaction = true;

                // Lock the bill row for the rest of the transaction so a concurrent
                // request cannot also observe it as pending and expand it again.
                $billStmt = $conn->prepare('SELECT id, is_stock_updated FROM buy_bill WHERE id = ? FOR UPDATE');
                $billStmt->execute([$billId]);
                $bill = $billStmt->fetch(PDO::FETCH_ASSOC);

                if (!$bill) {
                    apiErrorDie('bill_not_found');
                }
                if ((int)$bill['is_stock_updated'] === 1) {
                    apiErrorDie('stock_already_updated');
                }

                $detailStmt = $conn->prepare(
                    'SELECT bd.id, bd.QTY, bd.price_sell, bd.notes, bd.is_used_car, bd.is_big_car, bd.id_color,
                            cn.car_name, clr.color, clr.hexa, bb.bill_ref AS buy_bill_ref, bb.date_buy
                     FROM buy_details bd
                     LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
                     LEFT JOIN colors clr ON bd.id_color = clr.id
                     LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
                     WHERE bd.id_buy_bill = ?'
                );
                $detailStmt->execute([$billId]);
                $details = $detailStmt->fetchAll(PDO::FETCH_ASSOC);

                if (empty($details)) {
                    apiErrorDie('no_details');
                }

                // Refuse to expand if this bill already has stock rows. Covers the case
                // where is_stock_updated was reset by hand after a successful expansion.
                $detailIds = array_map(function ($d) { return intval($d['id']); }, $details);
                $placeholders = implode(',', array_fill(0, count($detailIds), '?'));
                $existingStmt = $conn->prepare(
                    "SELECT id_buy_details, COUNT(*) as cnt FROM cars_stock
                     WHERE id_buy_details IN ($placeholders) GROUP BY id_buy_details"
                );
                $existingStmt->execute($detailIds);
                $existing = $existingStmt->fetchAll(PDO::FETCH_ASSOC);

                if (!empty($existing)) {
                    $detailParts = [];
                    foreach ($existing as $row) {
                        $detailParts[] = [
                            'id' => intval($row['id_buy_details']),
                            'count' => intval($row['cnt']),
                        ];
                    }
                    apiErrorDie('stock_rows_exist', 200, ['details' => $detailParts]);
                }

                $insertStmt = $conn->prepare(
                    'INSERT INTO cars_stock
                     (id_buy_details, price_cell, notes, is_used_car, is_big_car, id_color)
                     VALUES (?, ?, ?, ?, ?, ?)'
                );

                foreach ($details as $detail) {
                    $qty = intval($detail['QTY']);
                    if ($qty <= 0) {
                        // This is the only refusal that can happen after rows have
                        // already been inserted, so roll back explicitly instead of
                        // relying on the connection closing to release the transaction.
                        $conn->rollBack();
                        $inTransaction = false;
                        apiErrorDie('invalid_detail_qty', 200, ['detailId' => intval($detail['id'])]);
                    }

                    for ($i = 0; $i < $qty; $i++) {
                        $insertStmt->execute([
                            $detail['id'],
                            $detail['price_sell'],
                            $detail['notes'],
                            $detail['is_used_car'],
                            $detail['is_big_car'],
                            $detail['id_color'],
                        ]);

                        $createdCars[] = [
                            'id' => intval($conn->lastInsertId()),
                            'id_buy_details' => intval($detail['id']),
                            'price_cell' => $detail['price_sell'],
                            'notes' => $detail['notes'],
                            'is_used_car' => $detail['is_used_car'],
                            'is_big_car' => $detail['is_big_car'],
                            'id_color' => $detail['id_color'],
                            'buy_bill_id' => $billId,
                            // Joined fields the grid needs; it reads these from
                            // cars_names/colors, so it cannot fill them itself.
                            'car_name' => $detail['car_name'],
                            'color' => $detail['color'],
                            'hexa' => $detail['hexa'],
                            'buy_bill_ref' => $detail['buy_bill_ref'],
                            'date_buy' => $detail['date_buy'],
                        ];
                    }
                }

                $flagStmt = $conn->prepare('UPDATE buy_bill SET is_stock_updated = 1 WHERE id = ?');
                $flagStmt->execute([$billId]);

                $conn->commit();
                $inTransaction = false;

                echo json_encode([
                    'success' => true,
                    'message' => 'Stock updated successfully',
                    'createdCount' => count($createdCars),
                    'createdCars' => $createdCars,
                ]);
            } catch (Exception $e) {
                if ($inTransaction && $conn instanceof PDO && $conn->inTransaction()) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
                exit;

        case 'update_buy_detail':
            // Admin-only edit of a buy detail, including after stock has been
            // updated. Refuses whenever any car of the detail is committed, so a
            // sold/VIN'd/filed/assigned car can never be altered out from under the
            // workflow. QTY changes create or remove cars: the delta is reconciled
            // against the actual cars_stock count for the detail, not the stored QTY.
            foreach (['bill_id', 'detail_id', 'user_id'] as $requiredKey) {
                if (!isset($postData[$requiredKey])) {
                    apiErrorDie('invalid_request', 400);
                }
            }

            $billId = intval($postData['bill_id']);
            $detailId = intval($postData['detail_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];

            if ($billId <= 0 || $detailId <= 0 || $userId <= 0) {
                apiErrorDie('invalid_request', 400);
            }

            $qtyValue = isset($postData['QTY']) ? $postData['QTY'] : null;
            if ($qtyValue === null || intval($qtyValue) < 1) {
                apiErrorDie('invalid_qty');
            }
            $newQty = intval($qtyValue);

            $amount = isset($postData['amount']) && $postData['amount'] !== '' ? floatval($postData['amount']) : 0;
            $priceSell = isset($postData['price_sell']) && $postData['price_sell'] !== '' ? floatval($postData['price_sell']) : 0;
            $year = isset($postData['year']) ? intval($postData['year']) : null;
            $month = isset($postData['month']) ? intval($postData['month']) : null;
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            $isUsed = isset($postData['is_used_car']) ? (int)(bool)$postData['is_used_car'] : 0;
            $isBig = isset($postData['is_big_car']) ? (int)(bool)$postData['is_big_car'] : 0;

            $createdCars = [];
            $removedCarIds = [];
            $inTransaction = false;

            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }

                if (!isAdminUser($conn, $userId)) {
                    apiErrorDie('not_admin');
                }

                $conn->beginTransaction();
                $inTransaction = true;

                $billStmt = $conn->prepare('SELECT id, is_stock_updated FROM buy_bill WHERE id = ? FOR UPDATE');
                $billStmt->execute([$billId]);
                $bill = $billStmt->fetch(PDO::FETCH_ASSOC);
                if (!$bill) {
                    apiErrorDie('bill_not_found');
                }

                $detailStmt = $conn->prepare('SELECT bd.id, bd.QTY, bd.price_sell, bd.notes, bd.is_used_car, bd.is_big_car, bd.id_color, cn.car_name, clr.color, clr.hexa, bb.bill_ref AS buy_bill_ref, bb.date_buy FROM buy_details bd LEFT JOIN cars_names cn ON bd.id_car_name = cn.id LEFT JOIN colors clr ON bd.id_color = clr.id LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id WHERE bd.id = ? AND bd.id_buy_bill = ?');
                $detailStmt->execute([$detailId, $billId]);
                $detail = $detailStmt->fetch(PDO::FETCH_ASSOC);

                if (!$detail) {
                    apiErrorDie('detail_not_found');
                }

                $updateStmt = $conn->prepare(
                    'UPDATE buy_details SET QTY = ?, amount = ?, year = ?, month = ?, price_sell = ?, notes = ?, is_used_car = ?, is_big_car = ? WHERE id = ?'
                );
                $updateStmt->execute([$newQty, $amount, $year, $month, $priceSell, $notes, $isUsed, $isBig, $detailId]);

                // Cars only exist once the bill has been expanded. On a pending bill
                // this edit must not create them: "Update Stock" is the only action
                // that turns QTY into cars_stock rows, otherwise saving a QTY on a
                // pending bill silently does the expansion and leaves the bill pending
                // with cars already attached (which then trips the stock_rows_exist
                // guard on the real Update Stock click).
                if ((int)$bill['is_stock_updated'] !== 1) {
                    $billAmountStmt = $conn->prepare(
                        'UPDATE buy_bill SET amount = (SELECT COALESCE(SUM(amount * QTY), 0) FROM buy_details WHERE id_buy_bill = ?) WHERE id = ?'
                    );
                    $billAmountStmt->execute([$billId, $billId]);

                    $conn->commit();
                    $inTransaction = false;

                    echo json_encode([
                        'success' => true,
                        'message' => 'Buy detail updated successfully',
                        'createdCount' => 0,
                        'createdCars' => [],
                        'removedCount' => 0,
                        'removedCarIds' => [],
                    ]);
                    exit;
                }

                $currentCars = countDetailCars($conn, $detailId);
                $committedCars = countCommittedCars($conn, $detailId);

                // Propagate the snapshot fields to this detail's cars, but only to the
                // untouched ones. A committed car is mid-workflow downstream and must
                // keep the price/notes it was actually purchased with.
                $propagateStmt = $conn->prepare(
                    'UPDATE cars_stock cs SET cs.price_cell = ?, cs.notes = ?, cs.is_used_car = ?, cs.is_big_car = ? WHERE cs.id_buy_details = ? AND NOT (' . isCommittedCarSql() . ')'
                );
                $propagateStmt->execute([$priceSell, $notes, $isUsed, $isBig, $detailId]);

                $delta = $newQty - $currentCars;

                if ($delta > 0) {
                    $insertStmt = $conn->prepare(
                        'INSERT INTO cars_stock (id_buy_details, price_cell, notes, is_used_car, is_big_car, id_color) VALUES (?, ?, ?, ?, ?, ?)'
                    );
                    for ($i = 0; $i < $delta; $i++) {
                        $insertStmt->execute([$detailId, $priceSell, $notes, $isUsed, $isBig, $detail['id_color']]);
                        $createdCars[] = [
                            'id' => intval($conn->lastInsertId()),
                            'id_buy_details' => $detailId,
                            'price_cell' => $priceSell,
                            'notes' => $notes,
                            'is_used_car' => $isUsed,
                            'is_big_car' => $isBig,
                            'id_color' => $detail['id_color'],
                            'buy_bill_id' => $billId,
                            // The grid reads these from joined tables, so the client
                            // cannot fill them in itself. Without them a freshly
                            // created car renders as a blank row.
                            'car_name' => $detail['car_name'],
                            'color' => $detail['color'],
                            'hexa' => $detail['hexa'],
                            'buy_bill_ref' => $detail['buy_bill_ref'],
                            'date_buy' => $detail['date_buy'],
                        ];
                    }
                } elseif ($delta < 0) {
                    // Reducing QTY removes cars, but only the ones nothing depends on.
                    // Committed cars (VIN, sold, client, ports, loading, files, payment)
                    // stay: they are real, in-progress work and deleting one to satisfy a
                    // number would destroy the workflow. If the requested cut reaches
                    // past the untouched cars, refuse with the real numbers rather than
                    // quietly keeping the old QTY.
                    $idsToRemove = (-$delta);
                    $removableCars = $currentCars - $committedCars;

                    if ($idsToRemove > $removableCars) {
                        $conn->rollBack();
                        $inTransaction = false;
                        apiErrorDie('qty_below_committed', 200, [
                            'requested' => $idsToRemove,
                            'removable' => max(0, $removableCars),
                            'committed' => $committedCars,
                            'minQty' => $committedCars,
                        ]);
                    }

                    // Take the most recently created untouched cars first. The
                    // NOT(...) keeps committed cars out of the candidate set entirely,
                    // so the LIMIT can only ever select cars that are safe to drop.
                    // $idsToRemove is a validated integer; LIMIT must be inlined
                    // because PDO emulated prepares would quote a bound value.
                    $idsStmt = $conn->prepare(
                        "SELECT cs.id FROM cars_stock cs WHERE cs.id_buy_details = ? AND NOT (" . isCommittedCarSql() . ") ORDER BY cs.id DESC LIMIT {$idsToRemove}"
                    );
                    $idsStmt->execute([$detailId]);
                    $idsToDelete = [];
                    while ($row = $idsStmt->fetch(PDO::FETCH_ASSOC)) {
                        $idsToDelete[] = intval($row['id']);
                        $removedCarIds[] = intval($row['id']);
                    }

                    if (!empty($idsToDelete)) {
                        $inPlaceholders = implode(',', array_fill(0, count($idsToDelete), '?'));
                        $deleteStmt = $conn->prepare("DELETE FROM cars_stock WHERE id IN ($inPlaceholders)");
                        $deleteStmt->execute($idsToDelete);
                    }
                }

                $billAmountStmt = $conn->prepare(
                    'UPDATE buy_bill SET amount = (SELECT COALESCE(SUM(amount * QTY), 0) FROM buy_details WHERE id_buy_bill = ?) WHERE id = ?'
                );
                $billAmountStmt->execute([$billId, $billId]);

                $conn->commit();
                $inTransaction = false;

                echo json_encode([
                    'success' => true,
                    'message' => 'Buy detail updated successfully',
                    'createdCount' => count($createdCars),
                    'createdCars' => $createdCars,
                    'removedCount' => count($removedCarIds),
                    'removedCarIds' => $removedCarIds,
                ]);
            } catch (Exception $e) {
                if ($inTransaction && $conn instanceof PDO && $conn->inTransaction()) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
                exit;

        case 'delete_buy_detail':
            // Admin-only delete of a buy detail, including after stock has been
            // updated. Destroys the detail's cars too, so it refuses whenever any of
            // them is committed. car_files and upgrades cascade by FK, but their
            // presence is itself a committed state, so the cascade never fires.
            foreach (['bill_id', 'detail_id', 'user_id'] as $requiredKey) {
                if (!isset($postData[$requiredKey])) {
                    apiErrorDie('invalid_request', 400);
                }
            }

            $billId = intval($postData['bill_id']);
            $detailId = intval($postData['detail_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];

            if ($billId <= 0 || $detailId <= 0 || $userId <= 0) {
                apiErrorDie('invalid_request', 400);
            }

            $removedCarIds = [];
            $inTransaction = false;

            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }

                if (!isAdminUser($conn, $userId)) {
                    apiErrorDie('not_admin');
                }

                $conn->beginTransaction();
                $inTransaction = true;

                $billStmt = $conn->prepare('SELECT id FROM buy_bill WHERE id = ? FOR UPDATE');
                $billStmt->execute([$billId]);
                if (!$billStmt->fetch(PDO::FETCH_ASSOC)) {
                    apiErrorDie('bill_not_found');
                }

                $detailStmt = $conn->prepare('SELECT id FROM buy_details WHERE id = ? AND id_buy_bill = ?');
                $detailStmt->execute([$detailId, $billId]);
                if (!$detailStmt->fetch(PDO::FETCH_ASSOC)) {
                    apiErrorDie('detail_not_found');
                }

                if (countCommittedCars($conn, $detailId) > 0) {
                    apiErrorDie('detail_locked');
                }

                // Collect the car ids being deleted for the client before removing them.
                $idsStmt = $conn->prepare('SELECT id FROM cars_stock WHERE id_buy_details = ? ORDER BY id ASC');
                $idsStmt->execute([$detailId]);
                while ($row = $idsStmt->fetch(PDO::FETCH_ASSOC)) {
                    $removedCarIds[] = intval($row['id']);
                }

                $deleteCarsStmt = $conn->prepare('DELETE FROM cars_stock WHERE id_buy_details = ?');
                $deleteCarsStmt->execute([$detailId]);

                $deleteDetailStmt = $conn->prepare('DELETE FROM buy_details WHERE id = ? AND id_buy_bill = ?');
                $deleteDetailStmt->execute([$detailId, $billId]);

                $billAmountStmt = $conn->prepare(
                    'UPDATE buy_bill SET amount = (SELECT COALESCE(SUM(amount * QTY), 0) FROM buy_details WHERE id_buy_bill = ?) WHERE id = ?'
                );
                $billAmountStmt->execute([$billId, $billId]);

                $conn->commit();
                $inTransaction = false;

                echo json_encode([
                    'success' => true,
                    'message' => 'Buy detail deleted successfully',
                    'removedCount' => count($removedCarIds),
                    'removedCarIds' => $removedCarIds,
                ]);
            } catch (Exception $e) {
                if ($inTransaction && $conn instanceof PDO && $conn->inTransaction()) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
                exit;

        case 'delete_buy_bill':
            // Admin-only delete of a whole bill. Unlike the old client-side version
            // (delete payments, then details, then the bill, each its own request),
            // this runs in one transaction, so a failure cannot leave a bill with its
            // payments and details already gone.
            //
            // An expanded bill owns cars_stock rows, so they are cascade deleted here
            // rather than orphaned. That is only safe while none of them is committed:
            // once a car has a VIN, a buyer, files, an assignment or a loading, the
            // downstream workflow depends on it and the bill must be refused.
            foreach (['bill_id', 'user_id'] as $requiredKey) {
                if (!isset($postData[$requiredKey])) {
                    apiErrorDie('invalid_request', 400);
                }
            }

            $billId = intval($postData['bill_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];

            if ($billId <= 0 || $userId <= 0) {
                apiErrorDie('invalid_request', 400);
            }

            $removedCarIds = [];
            $inTransaction = false;

            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }

                if (!isAdminUser($conn, $userId)) {
                    apiErrorDie('not_admin');
                }

                $conn->beginTransaction();
                $inTransaction = true;

                $billStmt = $conn->prepare('SELECT id FROM buy_bill WHERE id = ? FOR UPDATE');
                $billStmt->execute([$billId]);
                if (!$billStmt->fetch(PDO::FETCH_ASSOC)) {
                    apiErrorDie('bill_not_found');
                }

                $detailStmt = $conn->prepare('SELECT id FROM buy_details WHERE id_buy_bill = ?');
                $detailStmt->execute([$billId]);
                $detailIds = array_map(function ($row) {
                    return intval($row['id']);
                }, $detailStmt->fetchAll(PDO::FETCH_ASSOC));

                $committedCount = 0;
                if (!empty($detailIds)) {
                    $placeholders = implode(',', array_fill(0, count($detailIds), '?'));

                    $committedStmt = $conn->prepare(
                        'SELECT COUNT(*) AS cnt FROM cars_stock cs
                         WHERE cs.id_buy_details IN (' . $placeholders . ')
                           AND (' . isCommittedCarSql() . ')'
                    );
                    $committedStmt->execute($detailIds);
                    $committedCount = (int)$committedStmt->fetch(PDO::FETCH_ASSOC)['cnt'];

                    if ($committedCount > 0) {
                        apiErrorDie('bill_locked', 200, ['count' => $committedCount]);
                    }

                    $carsStmt = $conn->prepare(
                        "SELECT id FROM cars_stock WHERE id_buy_details IN ($placeholders) ORDER BY id ASC"
                    );
                    $carsStmt->execute($detailIds);
                    while ($row = $carsStmt->fetch(PDO::FETCH_ASSOC)) {
                        $removedCarIds[] = intval($row['id']);
                    }

                    // Collect the ids first so the client can drop them from the grid.
                    $deleteCarsStmt = $conn->prepare("DELETE FROM cars_stock WHERE id_buy_details IN ($placeholders)");
                    $deleteCarsStmt->execute($detailIds);
                }

                $deletePaymentsStmt = $conn->prepare('DELETE FROM buy_payments WHERE id_buy_bill = ?');
                $deletePaymentsStmt->execute([$billId]);

                $deleteDetailsStmt = $conn->prepare('DELETE FROM buy_details WHERE id_buy_bill = ?');
                $deleteDetailsStmt->execute([$billId]);

                $deleteBillStmt = $conn->prepare('DELETE FROM buy_bill WHERE id = ?');
                $deleteBillStmt->execute([$billId]);

                $conn->commit();
                $inTransaction = false;

                echo json_encode([
                    'success' => true,
                    'message' => 'Buy bill deleted successfully',
                    'removedCount' => count($removedCarIds),
                    'removedCarIds' => $removedCarIds,
                ]);
            } catch (Exception $e) {
                if ($inTransaction && $conn instanceof PDO && $conn->inTransaction()) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
                exit;

        // ============================================
        // Car Files Management Actions
        // ============================================
        
        case 'get_car_file_categories':
            // Get all file categories ordered by display_order
            $query = "SELECT * FROM car_file_categories ORDER BY display_order ASC, importance_level ASC";
            $result = executeQuery($query);
            echo json_encode($result);
            exit;

        case 'get_car_files':
            // Get files for a specific car with permission checking
            if (!isset($postData['car_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'car_id and user_id are required']);
                exit;
            }
            
            $carId = intval($postData['car_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            // Derived from the verified token, not from the payload. Read as
            // `(bool)$postData['is_admin']` this was a field the caller sets, and the
            // permission checks it guards were skipped entirely by sending true.
            $isAdmin = ((int) $currentUser['role_id'] === 1);
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check permission: can_upload_car_files or admin
                if (!$isAdmin && !hasPermission($conn, $userId, 'can_upload_car_files')) {
                    http_response_code(403);
                    echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to view car files']);
                    exit;
                }
                
                // Check if client_id column exists
                $hasClientIdColumn = false;
                try {
                    $checkColumnQuery = "SELECT COUNT(*) as cnt FROM INFORMATION_SCHEMA.COLUMNS 
                        WHERE TABLE_SCHEMA = DATABASE() 
                        AND TABLE_NAME = 'car_file_physical_tracking' 
                        AND COLUMN_NAME = 'client_id'";
                    $stmt = $conn->prepare($checkColumnQuery);
                    $stmt->execute();
                    $result = $stmt->fetch(PDO::FETCH_ASSOC);
                    $hasClientIdColumn = ($result && $result['cnt'] > 0);
                } catch (Exception $e) {
                    // If check fails, assume column doesn't exist
                    $hasClientIdColumn = false;
                }
                
                // Build query with permission filtering
                $clientFields = $hasClientIdColumn 
                    ? "cpt.client_id,
                        cl.name as client_name,
                        cl.email as client_email,
                        cl.mobiles as client_contact,"
                    : "NULL as client_id,
                        NULL as client_name,
                        NULL as client_email,
                        NULL as client_contact,";
                
                $clientJoin = $hasClientIdColumn 
                    ? "LEFT JOIN clients cl ON cpt.client_id = cl.id"
                    : "";
                
                $query = "
                    SELECT 
                        cf.*,
                        cfc.category_name,
                        cfc.importance_level,
                        cfc.is_required,
                        cfc.display_order,
                        cfc.visibility_scope as category_visibility,
                        u.username as uploaded_by_username,
                        cpt.id as tracking_id,
                        cpt.current_holder_id,
                        u_holder.username as current_holder_username,
                        cpt.previous_holder_id,
                        u_prev.username as previous_holder_username,
                        cpt.custom_clearance_agent_id,
                        cca.name as agent_name,
                        cpt.checkout_type,
                        {$clientFields}
                        cpt.status as physical_status,
                        cpt.checked_out_at,
                        cpt.checked_in_at,
                        cpt.expected_return_date,
                        CASE WHEN pending_transfer.id IS NOT NULL THEN 1 ELSE 0 END as has_pending_transfer,
                        pending_transfer.id as pending_transfer_id,
                        pending_transfer.to_user_id as pending_transfer_to_user_id
                    FROM car_files cf
                    INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
                    LEFT JOIN users u ON cf.uploaded_by = u.id
                    LEFT JOIN (
                        SELECT cpt1.*
                        FROM car_file_physical_tracking cpt1
                        INNER JOIN (
                            SELECT car_file_id, MAX(id) as max_id
                            FROM car_file_physical_tracking
                            WHERE status IN ('available', 'checked_out')
                            GROUP BY car_file_id
                        ) cpt2 ON cpt1.car_file_id = cpt2.car_file_id AND cpt1.id = cpt2.max_id
                        WHERE cpt1.status IN ('available', 'checked_out')
                    ) cpt ON cf.id = cpt.car_file_id
                    LEFT JOIN users u_holder ON cpt.current_holder_id = u_holder.id
                    LEFT JOIN users u_prev ON cpt.previous_holder_id = u_prev.id
                    LEFT JOIN custom_clearance_agents cca ON cpt.custom_clearance_agent_id = cca.id
                    LEFT JOIN (
                        SELECT id, car_file_id, to_user_id
                        FROM car_file_transfers
                        WHERE transfer_status = 'pending'
                    ) pending_transfer ON cf.id = pending_transfer.car_file_id
                    {$clientJoin}
                    WHERE cf.car_id = ? AND cf.is_active = 1
                ";
                
                $params = [$carId];
                
                // Add permission filtering if not admin
                if (!$isAdmin && $userId) {
                    $query .= " AND (
                        cf.uploaded_by = ? OR
                        cpt.current_holder_id = ? OR
                        cf.visibility_scope = 'public' OR
                        (cf.visibility_scope IS NULL AND cfc.visibility_scope = 'public')
                    )";
                    $params[] = $userId;
                    $params[] = $userId;
                }
                
                $query .= " ORDER BY cfc.display_order ASC, cfc.importance_level ASC, cf.uploaded_at DESC";
                
                $stmt = $conn->prepare($query);
                $stmt->execute($params);
                $files = $stmt->fetchAll(PDO::FETCH_ASSOC);
                
                echo json_encode(['success' => true, 'data' => $files]);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'create_car_file':
            // Create a new file record (after upload)
            // Automatically assigns the uploader as the first owner
            if (!isset($postData['car_id']) || !isset($postData['category_id']) || 
                !isset($postData['file_path']) || !isset($postData['file_name'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Missing required fields']);
                exit;
            }
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // The uploader is the caller. Read from the payload this was forgeable:
                // any user could claim the upload was somebody else's.
                $uploadedBy = (int) $currentUser['id'];
                $carId = intval($postData['car_id']);
                $categoryId = intval($postData['category_id']);
                
                // Check if category already has an active file for this car
                $checkQuery = "SELECT id, file_path FROM car_files 
                    WHERE car_id = ? AND category_id = ? AND is_active = 1
                    LIMIT 1";
                $checkStmt = $conn->prepare($checkQuery);
                $checkStmt->execute([$carId, $categoryId]);
                $existingFile = $checkStmt->fetch(PDO::FETCH_ASSOC);
                
                // If file exists, soft delete it first (replace scenario)
                if ($existingFile) {
                    $existingFileId = $existingFile['id'];
                    $existingFilePath = $existingFile['file_path'];
                    
// Get base directory for file deletion.
                    //
                    // Per tenant, from the registry: this is where the browser wrote
                    // the file (src/composables/useApi.js reads the same row). It used
                    // to be read with `SELECT js_dir FROM dbs LIMIT 1` on the TENANT
                    // connection, where that table is created by setup.sql but never
                    // written to - so it always came back empty and every delete went
                    // looking in mig_files, whichever deployment it was.
                    $baseDirectory = app_db_files_dir() ?? 'mig_files';

                // Delete physical file if it exists
                $baseDirectory = str_replace('..', '', $baseDirectory);
                $baseDirectory = ltrim($baseDirectory, '/');
                $baseDirectory = rtrim($baseDirectory, '/');
                $existingFilePath = str_replace('..', '', $existingFilePath);
                $existingFilePath = ltrim($existingFilePath, '/');
                    // app_deployment_root(), not __DIR__ . '/..': the registry records
                    // files_dir next to the app folder (<root>/mig_27_files), so for a
                    // tenant this is the parent, while for a single-app install it is
                    // the app folder itself.
                    $fullFilePath = app_deployment_root() . '/' . $baseDirectory . '/' . $existingFilePath;
                    
                    if (file_exists($fullFilePath)) {
                        @unlink($fullFilePath);
                    }
                    
                    // Soft delete the existing file record
                    $deleteQuery = "UPDATE car_files SET is_active = 0 WHERE id = ?";
                    $deleteStmt = $conn->prepare($deleteQuery);
                    $deleteStmt->execute([$existingFileId]);
                }
                
                // Insert file record
                $query = "INSERT INTO car_files 
                    (car_id, category_id, file_path, file_name, file_size, file_type, uploaded_by, notes, visibility_scope)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)";
                
                $params = [
                    $carId,
                    $categoryId,
                    $postData['file_path'],
                    $postData['file_name'],
                    isset($postData['file_size']) ? intval($postData['file_size']) : null,
                    isset($postData['file_type']) ? $postData['file_type'] : null,
                    $uploadedBy,
                    isset($postData['notes']) ? $postData['notes'] : null,
                    isset($postData['visibility_scope']) ? $postData['visibility_scope'] : null
                ];
                
                $stmt = $conn->prepare($query);
                $stmt->execute($params);
                $fileId = $conn->lastInsertId();
                
                // Create tracking record - uploader becomes the current holder
                $trackingQuery = "INSERT INTO car_file_physical_tracking 
                    (car_file_id, current_holder_id, status, checkout_type, checked_out_at, transfer_notes)
                    VALUES (?, ?, 'checked_out', 'user', NOW(), 'File uploaded - uploader is initial holder')";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId, $uploadedBy]);
                
                // Create initial transfer record showing uploader as holder
                $initialTransferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, to_user_id, transferred_by, notes, transfer_status, transfer_type)
                    VALUES (?, NULL, ?, ?, 'File uploaded - uploader is initial holder', 'approved', 'user_to_user')";
                $stmt = $conn->prepare($initialTransferQuery);
                $stmt->execute([$fileId, $uploadedBy, $uploadedBy]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'lastInsertId' => $fileId]);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'delete_car_file':
            // Delete a file (admin only) - deletes both database record and physical file
            if (!isset($postData['file_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and user_id are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            // Derived from the verified token, not from the payload. Read as
            // `(bool)$postData['is_admin']` this was a field the caller sets, and the
            // permission checks it guards were skipped entirely by sending true.
            $isAdmin = ((int) $currentUser['role_id'] === 1);
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Only admin can delete
                if (!$isAdmin) {
                    http_response_code(403);
                    throw new Exception('Permission denied: Only admin can delete files');
                }
                
                // Get file info including file_path
                $checkQuery = "SELECT id, file_path, uploaded_by FROM car_files WHERE id = ? AND is_active = 1";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $file = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$file) {
                    throw new Exception('File not found');
                }
                
                $filePath = $file['file_path'];
                
                // Get base directory from the registry, or fall back to the historical default.
                // See the note on the same lookup in add_car_file: the tenant's own
                // `dbs` table is empty on every install, so reading it answered
                // nothing and every delete looked in mig_files.
                $baseDirectory = app_db_files_dir() ?? 'mig_files';

                // Construct full file path
                $baseDirectory = str_replace('..', '', $baseDirectory);
                $baseDirectory = ltrim($baseDirectory, '/');
                $baseDirectory = rtrim($baseDirectory, '/');
                
                $filePath = str_replace('..', '', $filePath);
                $filePath = ltrim($filePath, '/');
                
                $fullFilePath = app_deployment_root() . '/' . $baseDirectory . '/' . $filePath;
                
                // Begin transaction
                $conn->beginTransaction();
                
                // Delete physical file if it exists
                $fileDeleted = false;
                if (file_exists($fullFilePath)) {
                    if (unlink($fullFilePath)) {
                        $fileDeleted = true;
                        error_log("Deleted physical file: $fullFilePath");
                    } else {
                        error_log("Failed to delete physical file: $fullFilePath");
                        // Continue with database deletion even if file deletion fails
                    }
                } else {
                    error_log("Physical file not found: $fullFilePath");
                    // Continue with database deletion even if file doesn't exist
                }
                
                // Soft delete from database
                $updateQuery = "UPDATE car_files SET is_active = 0 WHERE id = ?";
                $stmt = $conn->prepare($updateQuery);
                $stmt->execute([$fileId]);
                
                $conn->commit();
                
                echo json_encode([
                    'success' => true,
                    'affectedRows' => $stmt->rowCount(),
                    'fileDeleted' => $fileDeleted,
                    'message' => $fileDeleted ? 'File and database record deleted successfully' : 'Database record deleted (file was not found on server)'
                ]);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'checkout_physical_copy':
            // Check out a physical copy of a file (to user, client, or custom clearance agent)
            if (!isset($postData['file_id']) || !isset($postData['checkout_type'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and checkout_type are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            $checkoutType = $postData['checkout_type']; // 'user', 'client', or 'custom_clearance_agent'
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            $userId = isset($postData['user_id']) ? intval($postData['user_id']) : null;
            $agentId = isset($postData['agent_id']) ? intval($postData['agent_id']) : null;
            $clientId = isset($postData['client_id']) ? intval($postData['client_id']) : null;
            $clientName = isset($postData['client_name']) ? trim($postData['client_name']) : null;
            $clientContact = isset($postData['client_contact']) ? trim($postData['client_contact']) : null;
            
            // Validate based on checkout type
            if ($checkoutType === 'user' && !$userId) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'user_id is required for user checkout']);
                exit;
            }
            if ($checkoutType === 'custom_clearance_agent' && !$agentId) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'agent_id is required for agent checkout']);
                exit;
            }
            if ($checkoutType === 'client' && !$clientId) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'client_id is required for client checkout']);
                exit;
            }
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Check if file exists
                $checkQuery = "SELECT id, uploaded_by FROM car_files WHERE id = ? AND is_active = 1";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $file = $stmt->fetch(PDO::FETCH_ASSOC);
                if (!$file) {
                    throw new Exception('File not found or inactive');
                }
                
                // Get current tracking - file must be checked out to someone
                $trackingQuery = "SELECT id, current_holder_id, custom_clearance_agent_id, checkout_type, client_id, status
                    FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status IN ('available', 'checked_out')";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File tracking record not found');
                }
                
                // Verify that the current user is the current holder.
                // The comment said "the current user" but the code read it off the
                // payload, so the check below ran against whoever the body named.
                $performedBy = (int) $currentUser['id'];
                
                // Check if file is available (shouldn't happen after upload, but handle it)
                if ($tracking['status'] === 'available') {
                    // If available, only the uploader can checkout
                    if ($tracking['current_holder_id'] != $performedBy && $file['uploaded_by'] != $performedBy) {
                        throw new Exception('Only the current holder can checkout this file');
                    }
                } else {
                    // File is checked out - verify current holder
                    if ($tracking['current_holder_id'] != $performedBy) {
                        throw new Exception('Only the current holder can checkout this file');
                    }
                }
                
                // Get previous holder info for transfer history
                $fromUserId = $tracking['current_holder_id'];
                $fromAgentId = $tracking['custom_clearance_agent_id'];
                $fromClientName = null;
                if ($tracking['client_id']) {
                    $clientQuery = "SELECT name FROM clients WHERE id = ?";
                    $stmt = $conn->prepare($clientQuery);
                    $stmt->execute([$tracking['client_id']]);
                    $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                    $fromClientName = $clientData ? $clientData['name'] : null;
                }
                
                // Update tracking record
                // When checking out to user: current_holder_id changes to that user
                // When checking out to client/transiteur: current_holder_id becomes NULL
                // The previous_holder_id stores who checked it out (so they can rollback)
                $newHolderId = null;
                if ($checkoutType === 'user') {
                    $newHolderId = $userId;  // User becomes the holder
                } else {
                    // For client/transiteur checkout, set current_holder_id to NULL
                    // previous_holder_id will store who checked it out (for rollback)
                    $newHolderId = null;
                }
                
                $updateQuery = "UPDATE car_file_physical_tracking 
                    SET previous_holder_id = current_holder_id,
                        current_holder_id = ?,
                        custom_clearance_agent_id = ?,
                        checkout_type = ?,
                        client_id = ?,
                        checked_out_at = NOW(),
                        status = 'checked_out',
                        transfer_notes = ?
                    WHERE id = ?";
                $stmt = $conn->prepare($updateQuery);
                $stmt->execute([
                    $newHolderId,
                    $checkoutType === 'custom_clearance_agent' ? $agentId : null,
                    $checkoutType,
                    $checkoutType === 'client' ? $clientId : null,
                    $notes,
                    $tracking['id']
                ]);
                
                // Determine transfer type and create transfer history record
                $transferType = 'user_to_user';
                $toUserId = null;
                $toAgentId = null;
                $toClientName = null;
                
                if ($checkoutType === 'user') {
                    $toUserId = $userId;
                    $transferType = 'user_to_user';
                } elseif ($checkoutType === 'custom_clearance_agent') {
                    $toAgentId = $agentId;
                    $transferType = 'user_to_agent';
                } elseif ($checkoutType === 'client') {
                    // Get client name from database using client_id
                    if ($clientId) {
                        $clientNameQuery = "SELECT name FROM clients WHERE id = ?";
                        $stmt = $conn->prepare($clientNameQuery);
                        $stmt->execute([$clientId]);
                        $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                        $toClientName = $clientData ? $clientData['name'] : null;
                    }
                    $transferType = 'user_to_client';
                }
                
                // Insert transfer record for checkout - to_user_id can be NULL when checking out to agent or client
                $checkoutNotes = 'Checkout: ' . ($notes ?: 'Checked out');
                if ($checkoutType === 'user') {
                    $checkoutNotes = 'Checkout to user: ' . ($notes ?: 'Checked out to user');
                } elseif ($checkoutType === 'custom_clearance_agent') {
                    $checkoutNotes = 'Checkout to transiteur: ' . ($notes ?: 'Checked out to transiteur');
                } elseif ($checkoutType === 'client') {
                    $checkoutNotes = 'Checkout to client: ' . ($notes ?: 'Checked out to client');
                }
                
                $transferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, from_agent_id, to_user_id, to_agent_id, 
                     from_client_name, to_client_name, transferred_by, notes, transfer_type, transfer_status)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'approved')";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([
                    $fileId, 
                    $fromUserId, 
                    $fromAgentId, 
                    $toUserId,  // NULL for agent/client checkout
                    $toAgentId,
                    $fromClientName, 
                    $toClientName, 
                    $performedBy,
                    $checkoutNotes, 
                    $transferType
                ]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'File checked out successfully']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'rollback_checkout':
            // Rollback a checkout/transfer operation (admin only, and only for transfers/checkouts, not initial upload)
            if (!isset($postData['file_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and user_id are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            $notes = isset($postData['notes']) ? $postData['notes'] : 'Rollback checkout (admin)';
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Get current tracking record
                $trackingQuery = "SELECT id, status, current_holder_id, previous_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out'";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File is not currently checked out');
                }
                
                // Verify user can rollback:
                // - If current_holder_id is set: must be the current holder
                // - If current_holder_id is NULL (checked out to client/transiteur): must be the previous holder
                $canRollback = false;
                if ($tracking['current_holder_id'] && $tracking['current_holder_id'] == $userId) {
                    $canRollback = true;  // Current holder
                } elseif (!$tracking['current_holder_id'] && $tracking['previous_holder_id'] == $userId) {
                    $canRollback = true;  // Previous holder (when checked out to client/transiteur)
                }
                
                if (!$canRollback) {
                    throw new Exception('Only the current holder or previous holder (if checked out to client/transiteur) can rollback the checkout');
                }
                
                // Check if this is the initial upload transfer - cannot rollback initial upload
                // The initial transfer has from_user_id = NULL (upload) or is the first transfer
                $transferCountQuery = "SELECT COUNT(*) as cnt FROM car_file_transfers 
                    WHERE car_file_id = ? AND transfer_status = 'approved'";
                $stmt = $conn->prepare($transferCountQuery);
                $stmt->execute([$fileId]);
                $transferCount = $stmt->fetch(PDO::FETCH_ASSOC)['cnt'];
                
                // Get the most recent transfer record to check what to rollback to
                $getTransferQuery = "SELECT id, from_user_id, to_user_id, from_agent_id, to_agent_id, 
                    from_client_name, to_client_name, transfer_type, notes 
                    FROM car_file_transfers 
                    WHERE car_file_id = ? AND transfer_status = 'approved'
                    ORDER BY transferred_at DESC 
                    LIMIT 1";
                $stmt = $conn->prepare($getTransferQuery);
                $stmt->execute([$fileId]);
                $lastTransfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$lastTransfer) {
                    throw new Exception('No transfer record found to rollback');
                }
                
                // If this is the only transfer and it's the initial upload (from_user_id is NULL or notes contain "uploaded")
                if ($transferCount <= 1 && ($lastTransfer['from_user_id'] === null || 
                    stripos($lastTransfer['notes'], 'uploaded') !== false || 
                    stripos($lastTransfer['notes'], 'initial holder') !== false)) {
                    throw new Exception('Cannot rollback initial file upload. Use delete instead if you want to remove the file.');
                }
                
                if (!$lastTransfer) {
                    throw new Exception('No transfer record found to rollback');
                }
                
                // Determine the previous holder from the last transfer
                $previousHolderId = $lastTransfer['from_user_id'];
                $previousAgentId = $lastTransfer['from_agent_id'];
                $previousClientName = $lastTransfer['from_client_name'];
                
                // If from_user_id is NULL, this might be the initial upload - check notes
                if ($previousHolderId === null && (
                    stripos($lastTransfer['notes'], 'uploaded') !== false || 
                    stripos($lastTransfer['notes'], 'initial holder') !== false)) {
                    // This is the initial upload - cannot rollback
                    throw new Exception('Cannot rollback initial file upload. Use delete instead if you want to remove the file.');
                }
                
                // If no previous holder found, get it from the file's uploaded_by
                if ($previousHolderId === null && $previousAgentId === null) {
                    $fileQuery = "SELECT uploaded_by FROM car_files WHERE id = ?";
                    $stmt = $conn->prepare($fileQuery);
                    $stmt->execute([$fileId]);
                    $file = $stmt->fetch(PDO::FETCH_ASSOC);
                    $previousHolderId = $file ? $file['uploaded_by'] : null;
                }
                
                if ($previousHolderId === null && $previousAgentId === null && !$previousClientName) {
                    throw new Exception('Cannot determine previous holder. Cannot rollback.');
                }
                
                // Get current holder info for the rollback transfer record
                $currentHolderId = $tracking['current_holder_id'];
                $currentAgentId = null;
                $currentClientName = null;
                
                // Get current agent/client info if needed
                $currentTrackingQuery = "SELECT custom_clearance_agent_id, client_id, checkout_type 
                    FROM car_file_physical_tracking 
                    WHERE id = ?";
                $stmt = $conn->prepare($currentTrackingQuery);
                $stmt->execute([$tracking['id']]);
                $currentTracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if ($currentTracking) {
                    $currentAgentId = $currentTracking['custom_clearance_agent_id'];
                    if ($currentTracking['client_id']) {
                        $clientQuery = "SELECT name FROM clients WHERE id = ?";
                        $stmt = $conn->prepare($clientQuery);
                        $stmt->execute([$currentTracking['client_id']]);
                        $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                        $currentClientName = $clientData ? $clientData['name'] : null;
                    }
                }
                
                // Determine reverse transfer type
                $reverseTransferType = 'user_to_user';
                if ($lastTransfer['transfer_type'] === 'user_to_agent' || $lastTransfer['transfer_type'] === 'agent_to_user') {
                    $reverseTransferType = $currentAgentId ? 'agent_to_user' : 'user_to_agent';
                } elseif ($lastTransfer['transfer_type'] === 'user_to_client' || $lastTransfer['transfer_type'] === 'client_to_user') {
                    $reverseTransferType = $currentClientName ? 'client_to_user' : 'user_to_client';
                }
                
                // Add a new transfer record for rollback (reversing the last transfer)
                // This creates a history entry instead of deleting the last one
                // Get the username of who is performing the rollback
                $rollbackUserQuery = "SELECT username FROM users WHERE id = ?";
                $stmt = $conn->prepare($rollbackUserQuery);
                $stmt->execute([$userId]);
                $rollbackUser = $stmt->fetch(PDO::FETCH_ASSOC);
                $rollbackUsername = $rollbackUser ? $rollbackUser['username'] : 'Unknown User';
                
                $rollbackNotes = 'Rollback checkout';
                if ($lastTransfer['notes']) {
                    $rollbackNotes = 'Rollback checkout: Rolled back "' . $lastTransfer['notes'] . '"';
                }
                if ($notes) {
                    $rollbackNotes = 'Rollback checkout: ' . $notes;
                }
                // Add who performed it (the transferred_by field will also show this, but it's good to have in notes too)
                $rollbackNotes .= ' (by ' . $rollbackUsername . ')';
                
                $rollbackTransferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, to_user_id, from_agent_id, to_agent_id,
                     from_client_name, to_client_name, transferred_by, notes, transfer_type, transfer_status)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'approved')";
                $stmt = $conn->prepare($rollbackTransferQuery);
                $stmt->execute([
                    $fileId,
                    $currentHolderId,  // from: current holder (or NULL if checked out to client/transiteur)
                    $previousHolderId,  // to: previous holder
                    $currentAgentId,    // from agent
                    $previousAgentId,   // to agent
                    $currentClientName, // from client
                    $previousClientName, // to client
                    $userId,            // performed by current user
                    $rollbackNotes,
                    $reverseTransferType
                ]);
                
                // Update tracking to restore previous holder (file must always have a holder)
                $updateTrackingQuery = "UPDATE car_file_physical_tracking 
                    SET previous_holder_id = current_holder_id,
                        current_holder_id = ?,
                        custom_clearance_agent_id = ?,
                        client_id = ?,
                        checkout_type = ?,
                        checked_out_at = NOW(),
                        status = 'checked_out',
                        transfer_notes = ?
                    WHERE id = ?";
                
                // Get client_id if previous holder is a client
                $previousClientId = null;
                if ($previousClientName && !$previousHolderId && !$previousAgentId) {
                    $clientQuery = "SELECT id FROM clients WHERE name = ? LIMIT 1";
                    $stmt = $conn->prepare($clientQuery);
                    $stmt->execute([$previousClientName]);
                    $clientData = $stmt->fetch(PDO::FETCH_ASSOC);
                    $previousClientId = $clientData ? $clientData['id'] : null;
                }
                
                $checkoutType = 'user';
                if ($previousAgentId) {
                    $checkoutType = 'custom_clearance_agent';
                } elseif ($previousClientId || $previousClientName) {
                    $checkoutType = 'client';
                }
                
                $stmt = $conn->prepare($updateTrackingQuery);
                $stmt->execute([
                    $previousHolderId,
                    $previousAgentId,
                    $previousClientId,
                    $checkoutType,
                    $notes ?: 'Rollback: restored previous holder',
                    $tracking['id']
                ]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'Checkout rolled back successfully']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'checkin_physical_copy':
            // Check in a physical copy
            if (!isset($postData['file_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and user_id are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check if user is the current holder
                $checkQuery = "SELECT id, current_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out' AND current_holder_id = ?";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId, $userId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File is not checked out to you');
                }
                
                // Get current holder before checkin for transfer history
                $currentHolderId = $tracking['current_holder_id'];
                
                // Update tracking
                $updateQuery = "UPDATE car_file_physical_tracking 
                    SET previous_holder_id = current_holder_id,
                        current_holder_id = NULL, checked_in_at = NOW(), 
                        status = 'available', transfer_notes = ?
                    WHERE id = ?";
                $stmt = $conn->prepare($updateQuery);
                $stmt->execute([$notes, $tracking['id']]);
                
                // Update the last transfer record to mark it as returned (if exists)
                // This marks the transfer that gave the file to the current holder as returned
                // We also update transferred_by to show who checked it in
                // We need to update the most recent transfer where this user received the file
                $updateLastTransferQuery = "UPDATE car_file_transfers 
                    SET returned_at = NOW(), return_notes = ?, transferred_by = ?
                    WHERE car_file_id = ? AND to_user_id = ? AND returned_at IS NULL
                    ORDER BY transferred_at DESC LIMIT 1";
                $stmt = $conn->prepare($updateLastTransferQuery);
                $updateResult = $stmt->execute([$notes, $userId, $fileId, $currentHolderId]);
                
                // If no record was updated (shouldn't happen, but just in case), log it
                if ($stmt->rowCount() === 0) {
                    error_log("Warning: No transfer record found to update for check-in. File ID: $fileId, User ID: $currentHolderId");
                }
                
                echo json_encode(['success' => true, 'message' => 'File checked in successfully']);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'transfer_physical_copy':
            // Transfer physical copy from one user to another (creates pending transfer)
            if (!isset($postData['file_id']) || !isset($postData['from_user_id']) || 
                !isset($postData['to_user_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Missing required fields']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            $fromUserId = intval($postData['from_user_id']);
            $toUserId = intval($postData['to_user_id']);
            // The actor is the caller. This value was passed straight to
            // hasPermission() below, so naming user_id 1 granted admin permissions.
            $transferredBy = (int) $currentUser['id'];
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            $expectedReturnDate = isset($postData['expected_return_date']) ? $postData['expected_return_date'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check permission: can_upload_car_files or admin
                if (!hasPermission($conn, $transferredBy, 'can_upload_car_files')) {
                    http_response_code(403);
                    echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to transfer car files']);
                    exit;
                }
                
                $conn->beginTransaction();
                
                // Get current tracking status
                $checkQuery = "SELECT id, current_holder_id, status FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status IN ('available', 'checked_out')";
                $stmt = $conn->prepare($checkQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$tracking) {
                    throw new Exception('File tracking record not found');
                }
                
                // If file is available, allow transfer if uploader is transferring
                if ($tracking['status'] === 'available') {
                    // Verify the uploader is the one transferring
                    $fileQuery = "SELECT uploaded_by FROM car_files WHERE id = ?";
                    $stmt = $conn->prepare($fileQuery);
                    $stmt->execute([$fileId]);
                    $file = $stmt->fetch(PDO::FETCH_ASSOC);
                    
                    if (!$file || $file['uploaded_by'] != $fromUserId) {
                        throw new Exception('Only the uploader can transfer an available file');
                    }
                } else {
                    // File is checked out - verify current holder
                    if ($tracking['current_holder_id'] != $fromUserId) {
                        throw new Exception('File is not currently held by the specified user');
                    }
                }
                
                // Check if there's already a pending transfer for this file
                $pendingCheckQuery = "SELECT id FROM car_file_transfers 
                    WHERE car_file_id = ? AND transfer_status = 'pending'";
                $stmt = $conn->prepare($pendingCheckQuery);
                $stmt->execute([$fileId]);
                $pendingTransfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if ($pendingTransfer) {
                    throw new Exception('There is already a pending transfer for this file');
                }
                
                // Create pending transfer record (DO NOT update tracking yet - wait for approval)
                $transferQuery = "INSERT INTO car_file_transfers 
                    (car_file_id, from_user_id, to_user_id, transferred_by, notes, return_expected_date, transfer_status, transfer_type)
                    VALUES (?, ?, ?, ?, ?, ?, 'pending', 'user_to_user')";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([
                    $fileId, $fromUserId, $toUserId, $transferredBy, 
                    $notes, $expectedReturnDate
                ]);
                
                $transferId = $conn->lastInsertId();
                
                // Debug logging
                error_log("transfer_physical_copy: Created transfer ID=$transferId, file_id=$fileId, from_user=$fromUserId, to_user=$toUserId, status=pending");
                
                // Verify the transfer was created correctly
                $verifyQuery = "SELECT id, transfer_status, to_user_id FROM car_file_transfers WHERE id = ?";
                $verifyStmt = $conn->prepare($verifyQuery);
                $verifyStmt->execute([$transferId]);
                $verifyResult = $verifyStmt->fetch(PDO::FETCH_ASSOC);
                error_log("transfer_physical_copy: Verified transfer: " . json_encode($verifyResult));
                
                $conn->commit();
                echo json_encode([
                    'success' => true, 
                    'message' => 'Transfer request sent. Waiting for recipient approval.',
                    'transfer_id' => $transferId,
                    'debug' => $verifyResult
                ]);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'approve_transfer':
            // Approve a pending transfer
            if (!isset($postData['transfer_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'transfer_id and user_id are required']);
                exit;
            }
            
            $transferId = intval($postData['transfer_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Get the pending transfer
                $transferQuery = "SELECT cft.*, cf.file_name 
                    FROM car_file_transfers cft
                    INNER JOIN car_files cf ON cft.car_file_id = cf.id
                    WHERE cft.id = ? AND cft.transfer_status = 'pending'";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([$transferId]);
                $transfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$transfer) {
                    throw new Exception('Pending transfer not found');
                }
                
                // Verify the user is the recipient
                if ($transfer['to_user_id'] != $userId) {
                    throw new Exception('You are not the recipient of this transfer');
                }
                
                // Update transfer status to approved
                $updateTransferQuery = "UPDATE car_file_transfers 
                    SET transfer_status = 'approved', 
                        notes = CONCAT(COALESCE(notes, ''), IF(notes IS NULL OR notes = '', '', ' | '), ?)
                    WHERE id = ?";
                $stmt = $conn->prepare($updateTransferQuery);
                $stmt->execute([$notes ?: 'Transfer approved', $transferId]);
                
                // Now update the physical tracking
                // Handle both available and checked_out files
                $trackingQuery = "SELECT id, status, current_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status IN ('available', 'checked_out')";
                $stmt = $conn->prepare($trackingQuery);
                $stmt->execute([$transfer['car_file_id']]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if ($tracking) {
                    // Update tracking: set status to checked_out and assign to recipient
                    $updateTrackingQuery = "UPDATE car_file_physical_tracking 
                        SET previous_holder_id = ?, 
                            current_holder_id = ?, 
                            status = 'checked_out',
                            checked_out_at = NOW(),
                            transferred_by = ?, 
                            transferred_at = NOW(),
                            expected_return_date = ?, 
                            transfer_notes = ?
                        WHERE id = ?";
                    $stmt = $conn->prepare($updateTrackingQuery);
                    $stmt->execute([
                        $tracking['current_holder_id'], // previous_holder (could be NULL if was available)
                        $transfer['to_user_id'], 
                        $transfer['transferred_by'],
                        $transfer['return_expected_date'],
                        $transfer['notes'],
                        $tracking['id']
                    ]);
                } else {
                    // Create new tracking record if it doesn't exist (shouldn't happen, but safety check)
                    $insertTrackingQuery = "INSERT INTO car_file_physical_tracking 
                        (car_file_id, previous_holder_id, current_holder_id, status, checked_out_at, 
                         transferred_by, transferred_at, expected_return_date, transfer_notes, checkout_type)
                        VALUES (?, ?, ?, 'checked_out', NOW(), ?, NOW(), ?, ?, 'user')";
                    $stmt = $conn->prepare($insertTrackingQuery);
                    $stmt->execute([
                        $transfer['car_file_id'],
                        $transfer['from_user_id'],
                        $transfer['to_user_id'],
                        $transfer['transferred_by'],
                        $transfer['return_expected_date'],
                        $transfer['notes']
                    ]);
                }
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'Transfer approved successfully']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'reject_transfer':
            // Reject a pending transfer
            if (!isset($postData['transfer_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'transfer_id and user_id are required']);
                exit;
            }
            
            $transferId = intval($postData['transfer_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            $notes = isset($postData['notes']) ? $postData['notes'] : null;
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                $conn->beginTransaction();
                
                // Get the pending transfer
                $transferQuery = "SELECT * FROM car_file_transfers 
                    WHERE id = ? AND transfer_status = 'pending'";
                $stmt = $conn->prepare($transferQuery);
                $stmt->execute([$transferId]);
                $transfer = $stmt->fetch(PDO::FETCH_ASSOC);
                
                if (!$transfer) {
                    throw new Exception('Pending transfer not found');
                }
                
                // Verify the user is the recipient
                if ($transfer['to_user_id'] != $userId) {
                    throw new Exception('You are not the recipient of this transfer');
                }
                
                // Update transfer status to rejected
                $updateTransferQuery = "UPDATE car_file_transfers 
                    SET transfer_status = 'rejected', 
                        notes = CONCAT(COALESCE(notes, ''), IF(notes IS NULL OR notes = '', '', ' | '), ?)
                    WHERE id = ?";
                $stmt = $conn->prepare($updateTransferQuery);
                $stmt->execute([$notes ?: 'Transfer rejected', $transferId]);
                
                $conn->commit();
                echo json_encode(['success' => true, 'message' => 'Transfer rejected']);
            } catch (Exception $e) {
                if (isset($conn)) {
                    $conn->rollBack();
                }
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'get_pending_transfers':
            // Get pending transfers for the current user
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check if transfer_status column exists
                $hasTransferStatus = false;
                try {
                    $checkColumnQuery = "SELECT COUNT(*) as cnt FROM INFORMATION_SCHEMA.COLUMNS 
                        WHERE TABLE_SCHEMA = DATABASE() 
                        AND TABLE_NAME = 'car_file_transfers' 
                        AND COLUMN_NAME = 'transfer_status'";
                    $stmt = $conn->prepare($checkColumnQuery);
                    $stmt->execute();
                    $result = $stmt->fetch(PDO::FETCH_ASSOC);
                    $hasTransferStatus = ($result && $result['cnt'] > 0);
                } catch (Exception $e) {
                    error_log("Error checking transfer_status column: " . $e->getMessage());
                }
                
                // Build query - if transfer_status doesn't exist, get all transfers to this user
                if ($hasTransferStatus) {
                    $query = "
                        SELECT 
                            cft.*,
                            cf.file_name,
                            cf.car_id,
                            cfc.category_name,
                            u_from.username as from_username,
                            u_to.username as to_username,
                            u_transfer.username as transferred_by_username,
                            cs.vin
                        FROM car_file_transfers cft
                        INNER JOIN car_files cf ON cft.car_file_id = cf.id
                        INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
                        LEFT JOIN cars_stock cs ON cf.car_id = cs.id
                        LEFT JOIN users u_from ON cft.from_user_id = u_from.id
                        LEFT JOIN users u_to ON cft.to_user_id = u_to.id
                        LEFT JOIN users u_transfer ON cft.transferred_by = u_transfer.id
                        WHERE cft.to_user_id = ? AND cft.transfer_status = 'pending'
                        ORDER BY cft.transferred_at DESC
                    ";
                } else {
                    // Fallback: if column doesn't exist, return empty array and log warning
                    error_log("WARNING: transfer_status column does not exist. Please run migration 004_pending_transfers.sql");
                    $transfers = [];
                    echo json_encode([
                        'success' => true, 
                        'data' => $transfers,
                        'warning' => 'transfer_status column not found. Please run migration 004_pending_transfers.sql'
                    ]);
                    exit;
                }
                
                $stmt = $conn->prepare($query);
                $stmt->execute([$userId]);
                $transfers = $stmt->fetchAll(PDO::FETCH_ASSOC);
                
                // Debug logging
                error_log("get_pending_transfers: User ID = $userId, Found " . count($transfers) . " transfers");
                
                // Also check all pending transfers regardless of user (for debugging)
                $debugQuery = "SELECT id, car_file_id, from_user_id, to_user_id, transfer_status, transferred_at 
                    FROM car_file_transfers 
                    WHERE transfer_status = 'pending' 
                    ORDER BY transferred_at DESC LIMIT 10";
                $debugStmt = $conn->prepare($debugQuery);
                $debugStmt->execute();
                $allPending = $debugStmt->fetchAll(PDO::FETCH_ASSOC);
                error_log("get_pending_transfers: All pending transfers in DB: " . json_encode($allPending));
                
                if (count($transfers) > 0) {
                    error_log("First transfer for user $userId: " . json_encode($transfers[0]));
                } else {
                    error_log("get_pending_transfers: No transfers found for user $userId. All pending transfers: " . count($allPending));
                }
                
                echo json_encode([
                    'success' => true, 
                    'data' => $transfers,
                    'debug' => [
                        'user_id' => $userId,
                        'found_count' => count($transfers),
                        'all_pending_count' => count($allPending),
                        'all_pending' => $allPending
                    ]
                ]);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'get_file_transfer_history':
            // Get transfer history for a file
            if (!isset($postData['file_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id and user_id are required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            // Derived from the verified token, not from the payload. Read as
            // `(bool)$postData['is_admin']` this was a field the caller sets, and the
            // permission checks it guards were skipped entirely by sending true.
            $isAdmin = ((int) $currentUser['role_id'] === 1);
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Check permission: can_upload_car_files or admin
                if (!$isAdmin && !hasPermission($conn, $userId, 'can_upload_car_files')) {
                    http_response_code(403);
                    echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to view car file history']);
                    exit;
                }
                
                $query = "SELECT 
                    cft.*,
                    u_from.username as from_username,
                    u_to.username as to_username,
                    u_transfer.username as transferred_by_username,
                    cca_to.name as to_agent_name,
                    cca_from.name as from_agent_name,
                    cft.to_client_name,
                    cft.from_client_name,
                    cft.returned_at,
                    cft.return_notes
                    FROM car_file_transfers cft
                    LEFT JOIN users u_from ON cft.from_user_id = u_from.id
                    LEFT JOIN users u_to ON cft.to_user_id = u_to.id
                    LEFT JOIN users u_transfer ON cft.transferred_by = u_transfer.id
                    LEFT JOIN custom_clearance_agents cca_to ON cft.to_agent_id = cca_to.id
                    LEFT JOIN custom_clearance_agents cca_from ON cft.from_agent_id = cca_from.id
                    WHERE cft.car_file_id = ? AND cft.transfer_status = 'approved'
                    ORDER BY cft.transferred_at DESC";
                
                $result = executeQuery($query, [$fileId]);
                echo json_encode($result);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'create_file_category':
            // Create a new file category (admin only)
            // Admin, decided by the token verified at the action gate. This used to be
            // `!$postData['is_admin']` - a field the caller sets, so sending true was
            // all that stood between the internet and these writes.
            require_api_admin($gateConn, $postData);
            if (!isset($postData['category_name'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Category name is required']);
                exit;
            }
            
            $query = "INSERT INTO car_file_categories 
                (category_name, importance_level, is_required, display_order, description)
                VALUES (?, ?, ?, ?, ?)";
            
            $params = [
                $postData['category_name'],
                isset($postData['importance_level']) ? intval($postData['importance_level']) : 3,
                isset($postData['is_required']) ? (int)$postData['is_required'] : 0,
                isset($postData['display_order']) ? intval($postData['display_order']) : 0,
                isset($postData['description']) ? $postData['description'] : null
            ];
            
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'update_file_category':
            // Update a file category (admin only)
            // Admin, decided by the token verified at the action gate. This used to be
            // `!$postData['is_admin']` - a field the caller sets, so sending true was
            // all that stood between the internet and these writes.
            require_api_admin($gateConn, $postData);
            if (!isset($postData['category_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Category id is required']);
                exit;
            }
            
            $categoryId = intval($postData['category_id']);
            $updates = [];
            $params = [];
            
            if (isset($postData['category_name'])) {
                $updates[] = "category_name = ?";
                $params[] = $postData['category_name'];
            }
            if (isset($postData['importance_level'])) {
                $updates[] = "importance_level = ?";
                $params[] = intval($postData['importance_level']);
            }
            if (isset($postData['is_required'])) {
                $updates[] = "is_required = ?";
                $params[] = (int)$postData['is_required'];
            }
            if (isset($postData['display_order'])) {
                $updates[] = "display_order = ?";
                $params[] = intval($postData['display_order']);
            }
            if (isset($postData['description'])) {
                $updates[] = "description = ?";
                $params[] = $postData['description'];
            }
            
            if (empty($updates)) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'No fields to update']);
                exit;
            }
            
            $params[] = $categoryId;
            $query = "UPDATE car_file_categories SET " . implode(', ', $updates) . " WHERE id = ?";
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'delete_file_category':
            // Delete a file category (admin only)
            // Admin, decided by the token verified at the action gate. This used to be
            // `!$postData['is_admin']` - a field the caller sets, so sending true was
            // all that stood between the internet and these writes.
            require_api_admin($gateConn, $postData);
            if (!isset($postData['category_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Category id is required']);
                exit;
            }
            
            $categoryId = intval($postData['category_id']);
            
            // Check if category has files
            $checkQuery = "SELECT COUNT(*) as file_count FROM car_files WHERE category_id = ?";
            $checkResult = executeQuery($checkQuery, [$categoryId]);
            
            if ($checkResult['success'] && isset($checkResult['data'][0]['file_count']) && $checkResult['data'][0]['file_count'] > 0) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Cannot delete category: It has associated files. Please delete or reassign files first.']);
                exit;
            }
            
            $query = "DELETE FROM car_file_categories WHERE id = ?";
            $result = executeQuery($query, [$categoryId]);
            echo json_encode($result);
            exit;

        case 'get_my_physical_copies':
            // Get all physical copies currently held by a user
            // The caller is whoever the token resolved to at the action gate, not
            // whoever the request body says. This read $postData['user_id'] and then
            // treated it as the subject of the permission check below, so a non-admin
            // could send user_id: 1 (the seeded admin) and pass. isAdminUser() and
            // hasPermission() answer "is user N allowed", which is not the same
            // question as "is the caller allowed".
            $userId = (int) $currentUser['id'];
            
            $query = "SELECT 
                cf.*,
                cfc.category_name,
                cfc.importance_level,
                cs.vin,
                cpt.id as tracking_id,
                cpt.checked_out_at,
                cpt.expected_return_date,
                cpt.transfer_notes
                FROM car_file_physical_tracking cpt
                INNER JOIN car_files cf ON cpt.car_file_id = cf.id
                INNER JOIN car_file_categories cfc ON cf.category_id = cfc.id
                INNER JOIN cars_stock cs ON cf.car_id = cs.id
                WHERE cpt.current_holder_id = ? AND cpt.status = 'checked_out'
                ORDER BY cpt.checked_out_at DESC";
            
            $result = executeQuery($query, [$userId]);
            echo json_encode($result);
            exit;

        case 'get_users_for_transfer':
            // Get list of users for transfer dropdown (exclude current holder)
            if (!isset($postData['file_id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'file_id is required']);
                exit;
            }
            
            $fileId = intval($postData['file_id']);
            
            try {
                $conn = getConnection(getDbConfig());
                if (is_array($conn) && isset($conn['error'])) {
                    throw new Exception($conn['error']);
                }
                
                // Get current holder
                $holderQuery = "SELECT current_holder_id FROM car_file_physical_tracking 
                    WHERE car_file_id = ? AND status = 'checked_out'";
                $stmt = $conn->prepare($holderQuery);
                $stmt->execute([$fileId]);
                $tracking = $stmt->fetch(PDO::FETCH_ASSOC);
                $currentHolderId = $tracking ? $tracking['current_holder_id'] : null;
                
                // Get all users except current holder
                $query = "SELECT id, username, email FROM users WHERE id != ? ORDER BY username ASC";
                $params = $currentHolderId ? [$currentHolderId] : [0];
                
                $result = executeQuery($query, $params);
                echo json_encode($result);
            } catch (Exception $e) {
                echo json_encode(['success' => false, 'error' => $e->getMessage()]);
            }
            exit;

        case 'get_custom_clearance_agents':
            // Get all active custom clearance agents
            $query = "SELECT * FROM custom_clearance_agents WHERE is_active = 1 ORDER BY name ASC";
            $result = executeQuery($query);
            echo json_encode($result);
            exit;

        case 'create_custom_clearance_agent':
            // Create a new custom clearance agent (admin only)
            // Admin, decided by the token verified at the action gate. This used to be
            // `!$postData['is_admin']` - a field the caller sets, so sending true was
            // all that stood between the internet and these writes.
            require_api_admin($gateConn, $postData);

            $name = isset($postData['name']) ? trim($postData['name']) : '';
            if (empty($name)) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Name is required']);
                exit;
            }

            $query = "INSERT INTO custom_clearance_agents 
                (name, contact_person, phone, email, address, license_number, notes, is_active)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)";
            
            $params = [
                $name,
                isset($postData['contact_person']) ? $postData['contact_person'] : null,
                isset($postData['phone']) ? $postData['phone'] : null,
                isset($postData['email']) ? $postData['email'] : null,
                isset($postData['address']) ? $postData['address'] : null,
                isset($postData['license_number']) ? $postData['license_number'] : null,
                isset($postData['notes']) ? $postData['notes'] : null,
                isset($postData['is_active']) ? (int)$postData['is_active'] : 1
            ];
            
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'update_custom_clearance_agent':
            // Update a custom clearance agent (admin only)
            // Admin, decided by the token verified at the action gate. This used to be
            // `!$postData['is_admin']` - a field the caller sets, so sending true was
            // all that stood between the internet and these writes.
            require_api_admin($gateConn, $postData);

            if (!isset($postData['id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Agent ID is required']);
                exit;
            }

            $agentId = intval($postData['id']);
            $updates = [];
            $params = [];

            if (isset($postData['name'])) {
                $updates[] = "name = ?";
                $params[] = trim($postData['name']);
            }
            if (isset($postData['contact_person'])) {
                $updates[] = "contact_person = ?";
                $params[] = $postData['contact_person'];
            }
            if (isset($postData['phone'])) {
                $updates[] = "phone = ?";
                $params[] = $postData['phone'];
            }
            if (isset($postData['email'])) {
                $updates[] = "email = ?";
                $params[] = $postData['email'];
            }
            if (isset($postData['address'])) {
                $updates[] = "address = ?";
                $params[] = $postData['address'];
            }
            if (isset($postData['license_number'])) {
                $updates[] = "license_number = ?";
                $params[] = $postData['license_number'];
            }
            if (isset($postData['notes'])) {
                $updates[] = "notes = ?";
                $params[] = $postData['notes'];
            }
            if (isset($postData['is_active'])) {
                $updates[] = "is_active = ?";
                $params[] = (int)$postData['is_active'];
            }

            if (empty($updates)) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'No fields to update']);
                exit;
            }

            $params[] = $agentId;
            $query = "UPDATE custom_clearance_agents SET " . implode(', ', $updates) . " WHERE id = ?";
            $result = executeQuery($query, $params);
            echo json_encode($result);
            exit;

        case 'delete_custom_clearance_agent':
            // Soft delete a custom clearance agent (admin only)
            // Admin, decided by the token verified at the action gate. This used to be
            // `!$postData['is_admin']` - a field the caller sets, so sending true was
            // all that stood between the internet and these writes.
            require_api_admin($gateConn, $postData);

            if (!isset($postData['id'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Agent ID is required']);
                exit;
            }

            $agentId = intval($postData['id']);
            $query = "UPDATE custom_clearance_agents SET is_active = 0 WHERE id = ?";
            $result = executeQuery($query, [$agentId]);
            echo json_encode($result);
            exit;

        case 'save_contract_terms':
            // Save contract terms to public/contract_terms.json (admin only)
            // Admin, decided by the token verified at the action gate. This used to be
            // `!$postData['is_admin']` - a field the caller sets, so sending true was
            // all that stood between the internet and these writes.
            require_api_admin($gateConn, $postData);

            if (!isset($postData['content'])) {
                http_response_code(400);
                echo json_encode(['success' => false, 'error' => 'Content is required']);
                exit;
            }

            try {
                $jsonContent = $postData['content'];
                
                // Validate JSON content
                if (!is_array($jsonContent)) {
                    throw new Exception('Content must be a valid JSON object');
                }
                
                // Ensure required structure
                if (!isset($jsonContent['enabledLanguages']) || !isset($jsonContent['terms'])) {
                    throw new Exception('Content must include enabledLanguages and terms');
                }
                
                // Get the real path for security check
                $realBasePath = realpath(__DIR__ . '/../');
                if ($realBasePath === false) {
                    throw new Exception('Invalid base directory');
                }
                $realBasePath = rtrim($realBasePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                
                // Construct the file path to public/contract_terms.json
                $filePath = __DIR__ . '/../public/contract_terms.json';
                
                // Verify the file path is within the base directory
                $resolvedFilePath = realpath(dirname($filePath));
                if ($resolvedFilePath === false) {
                    // Directory doesn't exist, create it
                    if (!mkdir(dirname($filePath), 0755, true)) {
                        throw new Exception('Failed to create directory');
                    }
                    $resolvedFilePath = realpath(dirname($filePath));
                    if ($resolvedFilePath === false) {
                        throw new Exception('Failed to resolve directory path');
                    }
                }
                $resolvedFilePath = rtrim($resolvedFilePath, DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR;
                if (strpos($resolvedFilePath, $realBasePath) !== 0) {
                    throw new Exception('Directory traversal attempt detected');
                }
                
                // Encode JSON with pretty printing
                $jsonString = json_encode($jsonContent, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
                
                if ($jsonString === false) {
                    throw new Exception('Failed to encode JSON: ' . json_last_error_msg());
                }
                
                // Write file
                if (file_put_contents($filePath, $jsonString) === false) {
                    throw new Exception('Failed to write file');
                }
                
                echo json_encode([
                    'success' => true,
                    'message' => 'Contract terms saved successfully'
                ]);
            } catch (Exception $e) {
                http_response_code(500);
                echo json_encode(['success' => false, 'error' => 'Error saving file: ' . $e->getMessage()]);
            }
            exit;
    
        // check_api_files_exist used to live here. It answers "does this file exist in
        // the api folder", which is a question about the deployment rather than about
        // the tenant's data, and its only caller is the DB manager's build uploader
        // (Databases.vue). Sitting in this file it sat behind the *app* token while
        // being driven by the *db-manager* realm, so it could not be given a token its
        // caller actually holds - and because its failure was a `success: false`
        // rather than a thrown error, the caller silently skipped the "these files
        // will replace existing files" prompt instead of falling back to it. It now
        // lives in db_manager_api.php alongside the other deployment actions, gated
        // on login.api_token. See the case there.
    
        // Login and supplier credibility checks live in their own modules; this
        // switch only dispatches. Required here rather than at the top of the
        // file so an unused module costs nothing on every other action.
        case 'login':
            require_once __DIR__ . '/actions/auth.php';
            handle_login($postData);
            exit;

        case 'logout':
            require_once __DIR__ . '/actions/auth.php';
            handle_logout($postData);
            exit;

        case 'assess_supplier_credibility':
            require_once __DIR__ . '/actions/supplier_credibility.php';
            handle_assess_supplier_credibility($postData);
            exit;

        case 'get_supplier_credibility_checks':
            require_once __DIR__ . '/actions/supplier_credibility.php';
            handle_get_supplier_credibility_checks($postData);
            exit;

        case 'get_supplier_credibility_latest':
            require_once __DIR__ . '/actions/supplier_credibility.php';
            handle_get_supplier_credibility_latest($postData);
            exit;

        // The share page is public by design: the share_token in the path is the
        // credential, which is why this one is in PUBLIC_ACTIONS above.
        case 'get_client_share_data':
            require_once __DIR__ . '/actions/client_share.php';
            handle_get_client_share_data($postData);
            exit;

        // See the note on PUBLIC_ACTIONS: DatabaseVersionCheck is mounted
        // unconditionally in App.vue, so this runs on /login too.
        case 'get_db_version':
            require_once __DIR__ . '/actions/version.php';
            handle_get_db_version();
            exit;

        case 'get_colors':
            require_once __DIR__ . '/actions/colors.php';
            handle_get_colors($postData);
            exit;

        case 'create_color':
            require_once __DIR__ . '/actions/colors.php';
            handle_create_color($postData);
            exit;

        case 'update_color':
            require_once __DIR__ . '/actions/colors.php';
            handle_update_color($postData);
            exit;

        case 'delete_color':
            require_once __DIR__ . '/actions/colors.php';
            handle_delete_color($postData);
            exit;

        case 'get_container_tracking':
            require_once __DIR__ . '/actions/containers.php';
            handle_get_container_tracking($postData);
            exit;

        case 'save_container_tracking':
            require_once __DIR__ . '/actions/containers.php';
            handle_save_container_tracking($postData);
            exit;

        case 'create_user':
            require_once __DIR__ . '/actions/users.php';
            handle_create_user($postData);
            exit;

        case 'change_own_password':
            require_once __DIR__ . '/actions/users.php';
            handle_change_own_password($postData);
            exit;

        case 'set_user_password':
            require_once __DIR__ . '/actions/users.php';
            handle_set_user_password($postData);
            exit;

        // Public for the same reason as login itself: this is the recovery path for
        // a user who cannot authenticate. It requires the current password, which
        // the handler verifies server-side.
        case 'change_password_with_credentials':
            require_once __DIR__ . '/actions/users.php';
            handle_change_password_with_credentials($postData);
            exit;

        default:
            http_response_code(400);
            echo json_encode(['success' => false, 'error' => 'Invalid action']);
            exit;
    }
}

// Regular query handling
//
// This path sits outside the switch above, so the action gate never sees it: a
// request carrying `query` with no `action` walked straight past authentication.
// Anything genuinely public must be its own action, not a flag on this path: a
// client-sent "public" marker would be a bypass for this gate.
$passthroughConn = getConnection(getDbConfig());
if (is_array($passthroughConn) && isset($passthroughConn['error'])) {
    error_log('query passthrough: ' . $passthroughConn['error']);
    apiErrorDie('db_unavailable');
}
$passthroughUser = require_api_user($passthroughConn, $postData);

if (!isset($postData['query']) || empty($postData['query'])) {
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Query is required']);
    exit;
}

// Get query and parameters
$query = $postData['query'];
$params = isset($postData['params']) ? $postData['params'] : [];

// Check if query is updating payment_confirmed column - requires permission check
//
// This used to be a plain prefix sniff on the raw string:
//
//     strpos($queryUpper, 'UPDATE') === 0 && strpos($queryUpper, 'PAYMENT_CONFIRMED') !== false
//
// which a single leading comment defeats - `/*x*/ UPDATE sell_bill SET
// payment_confirmed = 1 ...` puts UPDATE at byte 5, so the whole permission block
// below was skipped and the query ran anyway.
//
// So classify against the statement with its leading comments and whitespace
// removed. That is the only reason this is not just strpos() again: the verb has to
// be the first real token, and a comment in front of it must not hide it.
//
// $queryUpper must stay defined for the rest of this block - the checks inside it
// classify the statement further (is it a sell_bill update, is it the simple
// payment_confirmed-only form). Deriving it from the stripped SQL also makes those
// inner prefix tests comment-proof for free.
$queryUpper = strtoupper(api_strip_leading_sql(trim($query)));

// UPDATE, INSERT and REPLACE can all set payment_confirmed. The original sniff
// only recognised UPDATE, so `INSERT INTO sell_bill (payment_confirmed) VALUES (1)`
// wrote the column with no permission check at all. Widen the verb set - the
// permission gate below runs before the shape detection, and anything that is not
// a sell_bill UPDATE simply falls through to normal execution once the gate passes.
$isPaymentConfirmedWrite = (
    preg_match('/^(UPDATE|INSERT|REPLACE)\b/', $queryUpper) === 1
    && strpos($queryUpper, 'PAYMENT_CONFIRMED') !== false
);

if ($isPaymentConfirmedWrite) {
    // Identify the caller from the token verified above, not from the request.
    //
    // This used to read `user_id` and `is_admin` off the payload: send is_admin: true
    // and the hasPermission() check below was skipped entirely, so any caller could
    // mark payments confirmed. $passthroughUser comes from require_api_user(), which
    // matched the token against users.api_token.
    $userId = (int) $passthroughUser['id'];
    $isAdmin = ((int) $passthroughUser['role_id'] === 1);

    try {
        $conn = getConnection(getDbConfig());
        if (is_array($conn) && isset($conn['error'])) {
            throw new Exception($conn['error']);
        }
        
        // Check permission: must be admin or have can_confirm_payment permission
        if (!$isAdmin && !hasPermission($conn, $userId, 'can_confirm_payment')) {
            http_response_code(403);
            echo json_encode(['success' => false, 'error' => 'Permission denied: You do not have permission to confirm payments']);
            exit;
        }
        
        // Check if this is updating sell_bill table
        $isSellBillUpdate = (stripos($queryUpper, 'UPDATE') === 0 && stripos($queryUpper, 'SELL_BILL') !== false);
        
        if ($isSellBillUpdate) {
            // Check if this is a simple payment_confirmed-only update (old format)
            // Or a full update that includes payment_confirmed (new format)
            $isSimplePaymentUpdate = (stripos($queryUpper, 'SET') !== false && 
                                     stripos($queryUpper, 'PAYMENT_CONFIRMED') !== false &&
                                     substr_count($queryUpper, '?') <= 3); // Simple: SET payment_confirmed = ? WHERE id = ?
            
            if ($isSimplePaymentUpdate && count($params) >= 2) {
                // Old simple format: UPDATE sell_bill SET payment_confirmed = ? WHERE id = ?
                $newStatus = intval($params[0]);
                $billId = intval($params[1]);
                
                // Execute the sell_bill update using the existing connection
                try {
                    // Update both payment_confirmed and payment_confirmed_by_user_id in a single query
                    // Set payment_confirmed_by_user_id to userId if confirming (1), NULL if unconfirming (0)
                    $confirmedByUserId = $newStatus == 1 ? $userId : null;
                    
                    // Modify the query to also update payment_confirmed_by_user_id
                    $updatedQuery = "UPDATE sell_bill SET payment_confirmed = ?, payment_confirmed_by_user_id = ? WHERE id = ?";
                    $updatedParams = [$newStatus, $confirmedByUserId, $billId];
                    
                    $stmt = $conn->prepare($updatedQuery);
                    $stmt->execute($updatedParams);
                    $affectedRows = $stmt->rowCount();
                    $result = ['success' => true, 'affectedRows' => $affectedRows];
                    
                    // If update was successful, update related cars
                    if ($affectedRows > 0) {
                        
                        // Update all related cars
                        $updateCarsQuery = "UPDATE cars_stock SET payment_confirmed = ? WHERE id_sell = ?";
                        $carsStmt = $conn->prepare($updateCarsQuery);
                        $carsStmt->execute([$newStatus, $billId]);
                        $carsAffected = $carsStmt->rowCount();
                        
                        // Log the update (optional, for debugging)
                        if ($carsAffected > 0) {
                            error_log("Updated payment_confirmed for $carsAffected cars related to sell_bill ID $billId to status $newStatus by user $userId");
                        }
                    }
                    
                    // Return the sell_bill update result
                    http_response_code(200);
                    echo json_encode($result);
                    exit;
                } catch (Exception $e) {
                    http_response_code(500);
                    echo json_encode(['success' => false, 'error' => $e->getMessage()]);
                    exit;
                }
            } else {
                // Full update query that includes payment_confirmed along with other fields
                // We need to inject payment_confirmed_by_user_id into the query
                try {
                    // Find the payment_confirmed parameter index by counting placeholders before it
                    $paymentConfirmedIndex = -1;
                    
                    // Extract the SET clause
                    if (preg_match('/SET\s+(.+?)\s+WHERE/i', $query, $matches)) {
                        $setClause = $matches[1];
                        
                        // Count question marks before payment_confirmed
                        $beforePaymentConfirmed = substr($setClause, 0, stripos($setClause, 'payment_confirmed'));
                        $paymentConfirmedIndex = substr_count($beforePaymentConfirmed, '?');
                    }
                    
                    // Find bill ID (last parameter)
                    $billIdIndex = count($params) - 1;
                    $billId = intval($params[$billIdIndex]);
                    
                    // Get payment_confirmed value
                    if ($paymentConfirmedIndex >= 0 && isset($params[$paymentConfirmedIndex])) {
                        $newStatus = intval($params[$paymentConfirmedIndex]);
                        $confirmedByUserId = $newStatus == 1 ? $userId : null;
                        
                        // Modify query to include payment_confirmed_by_user_id
                        // Insert it right after payment_confirmed
                        $modifiedQuery = preg_replace(
                            '/(payment_confirmed\s*=\s*\?)/i',
                            '$1, payment_confirmed_by_user_id = ?',
                            $query
                        );
                        
                        // Insert the confirmed_by_user_id parameter right after payment_confirmed
                        $modifiedParams = $params;
                        array_splice($modifiedParams, $paymentConfirmedIndex + 1, 0, $confirmedByUserId);
                        
                        // Execute the modified query
                        $stmt = $conn->prepare($modifiedQuery);
                        $stmt->execute($modifiedParams);
                        $affectedRows = $stmt->rowCount();
                        
                        // If update was successful, update related cars
                        if ($affectedRows > 0 && $newStatus !== null) {
                            $updateCarsQuery = "UPDATE cars_stock SET payment_confirmed = ? WHERE id_sell = ?";
                            $carsStmt = $conn->prepare($updateCarsQuery);
                            $carsStmt->execute([$newStatus, $billId]);
                        }
                        
                        $result = ['success' => true, 'affectedRows' => $affectedRows];
                        http_response_code(200);
                        echo json_encode($result);
                        exit;
                    }
                } catch (Exception $e) {
                    // If modification fails, fall through to normal execution
                    error_log("Error modifying payment_confirmed query: " . $e->getMessage());
                }
            }
        }
    } catch (Exception $e) {
        http_response_code(500);
        echo json_encode(['success' => false, 'error' => 'Error checking permissions: ' . $e->getMessage()]);
        exit;
    }
}

// Execute query and return result
http_response_code(200);
echo json_encode(executeQuery($query, $params));
exit;
?>