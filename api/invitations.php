<?php
header('Content-Type: application/json');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Authorization');

if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    exit(0);
}

// Database configuration
$host = '173.214.163.18';
$port = '3306';
$dbname = 'merhab_invitations';
$username = 'merhab_root';
$password = '@Salima61';

try {
    $pdo = new PDO("mysql:host=$host;port=$port;dbname=$dbname;charset=utf8", $username, $password);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
} catch (PDOException $e) {
    http_response_code(500);
    echo json_encode(['error' => 'Database connection failed: ' . $e->getMessage()]);
    exit;
}

$method = $_SERVER['REQUEST_METHOD'];

switch ($method) {
    case 'GET':
        try {
            $stmt = $pdo->query("SELECT * FROM invitations ORDER BY dateInv DESC");
            $invitations = $stmt->fetchAll(PDO::FETCH_ASSOC);
            echo json_encode(['success' => true, 'data' => $invitations]);
        } catch (PDOException $e) {
            http_response_code(500);
            echo json_encode(['error' => 'Failed to fetch invitations: ' . $e->getMessage()]);
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
            echo json_encode(['error' => 'Failed to create invitation: ' . $e->getMessage()]);
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
            echo json_encode(['error' => 'Failed to update invitation: ' . $e->getMessage()]);
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
            echo json_encode(['error' => 'Failed to delete invitation: ' . $e->getMessage()]);
        }
        break;
        
    default:
        http_response_code(405);
        echo json_encode(['error' => 'Method not allowed']);
        break;
}
?> 