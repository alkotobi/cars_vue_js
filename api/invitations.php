<?php
// Invitation CRUD.
//
// This file had no authentication on any method: anyone could list the whole
// invitations table, insert rows, rewrite entries and - via a plain
// `GET /api/invitations.php?id=7` - delete them. GET, POST, PUT and DELETE are all
// now behind an admin token.
//
// Error bodies used to concatenate the driver's message, which names the host,
// the schema and the connecting user; the detail now goes to the error log.

require_once __DIR__ . '/lib/cors.php';
api_send_cors_headers();

header('Content-Type: application/json');
header('Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Authorization, X-Api-Token');

if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    exit(0);
}

// Database configuration for the merhab_invitations database.
// Credentials are resolved by the shared loader in config.php: it reads
// api/config.local.php (git-ignored) and falls back to DB_HOST / DB_USER /
// DB_PASS / DB_NAME. The invitations database name is overridden here because
// it is a third database, distinct from the main app database.
require_once __DIR__ . '/config.php';
require_once __DIR__ . '/lib/appdb.php';

$host = $db_host;
$port = '3306';
$dbname = 'merhab_invitations';
$username = $db_user;
$password = $db_pass;

try {
    $pdo = new PDO("mysql:host=$host;port=$port;dbname=$dbname;charset=utf8", $username, $password);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
} catch (PDOException $e) {
    error_log('invitations.php connect: ' . $e->getMessage());
    http_response_code(500);
    echo json_encode(['error' => 'Database connection failed']);
    exit;
}

// Require an admin before any handler runs. The token comes from the header when
// the caller can set one, or from the query string, which is what the Vue client
// sends for a GET.
require_app_admin($_SERVER['HTTP_X_API_TOKEN'] ?? ($_GET['token'] ?? ''));

$method = $_SERVER['REQUEST_METHOD'];

switch ($method) {
    case 'GET':
        try {
            $stmt = $pdo->query("SELECT * FROM invitations ORDER BY dateInv DESC");
            $invitations = $stmt->fetchAll(PDO::FETCH_ASSOC);
            echo json_encode(['success' => true, 'data' => $invitations]);
        } catch (PDOException $e) {
            http_response_code(500);
            error_log('invitations.php: ' . $e->getMessage());
            echo json_encode(['error' => 'Failed to fetch invitations']);
        }
        break;
        
    case 'POST':
        $input = json_decode(file_get_contents('php://input'), true);
        
        try {
            $stmt = $pdo->prepare("
                INSERT INTO invitations (name, pass, dateInv, value, payment, rate, rem, balance) 
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            ");
            
            $stmt->execute([
                $input['name'] ?? null,
                $input['pass'] ?? null,
                $input['dateInv'] ?? date('Y-m-d H:i:s'),
                $input['value'] ?? 0,
                $input['payment'] ?? 0,
                $input['rate'] ?? 0,
                $input['rem'] ?? null,
                $input['balance'] ?? 0
            ]);
            
            $id = $pdo->lastInsertId();
            echo json_encode(['success' => true, 'id' => $id]);
        } catch (PDOException $e) {
            http_response_code(500);
            error_log('invitations.php: ' . $e->getMessage());
            echo json_encode(['error' => 'Failed to create invitation']);
        }
        break;
        
    case 'PUT':
        $input = json_decode(file_get_contents('php://input'), true);
        $id = $input['id'] ?? null;
        
        if (!$id) {
            http_response_code(400);
            echo json_encode(['error' => 'ID is required']);
            break;
        }
        
        try {
            $stmt = $pdo->prepare("
                UPDATE invitations 
                SET name = ?, pass = ?, dateInv = ?, value = ?, payment = ?, rate = ?, rem = ?, balance = ?
                WHERE id = ?
            ");
            
            $stmt->execute([
                $input['name'] ?? null,
                $input['pass'] ?? null,
                $input['dateInv'] ?? null,
                $input['value'] ?? 0,
                $input['payment'] ?? 0,
                $input['rate'] ?? 0,
                $input['rem'] ?? null,
                $input['balance'] ?? 0,
                $id
            ]);
            
            echo json_encode(['success' => true]);
        } catch (PDOException $e) {
            http_response_code(500);
            error_log('invitations.php: ' . $e->getMessage());
            echo json_encode(['error' => 'Failed to update invitation']);
        }
        break;
        
    case 'DELETE':
        $id = $_GET['id'] ?? null;
        
        if (!$id) {
            http_response_code(400);
            echo json_encode(['error' => 'ID is required']);
            break;
        }
        
        try {
            $stmt = $pdo->prepare("DELETE FROM invitations WHERE id = ?");
            $stmt->execute([$id]);
            
            echo json_encode(['success' => true]);
        } catch (PDOException $e) {
            http_response_code(500);
            error_log('invitations.php: ' . $e->getMessage());
            echo json_encode(['error' => 'Failed to delete invitation']);
        }
        break;
        
    default:
        http_response_code(405);
        echo json_encode(['error' => 'Method not allowed']);
        break;
}
?> 