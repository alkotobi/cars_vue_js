<?php
header('Content-Type: application/json');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: POST');
header('Access-Control-Allow-Headers: Content-Type');

// Database connection
$db_config = require_once __DIR__ . '/config/database.php';
$conn = new mysqli($db_config['host'], $db_config['username'], $db_config['password'], $db_config['database']);

if ($conn->connect_error) {
    die(json_encode([
        'success' => false,
        'message' => 'Database connection failed'
    ]));
}

// Response array
$response = [
    'success' => false,
    'message' => '',
    'uploaded_files' => []
];

try {
    // Check if it's a POST request
    if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
        throw new Exception('Only POST method is allowed');
    }

    // Start transaction
    $conn->begin_transaction();

    // First, insert the client record to get the ID
    // Assuming you're sending other client data as POST parameters
    $stmt = $conn->prepare("INSERT INTO clients (name, other_fields) VALUES (?, ?)");
    $name = $_POST['name'] ?? '';
    $other_fields = $_POST['other_fields'] ?? '';
    $stmt->bind_param("ss", $name, $other_fields);
    
    if (!$stmt->execute()) {
        throw new Exception('Failed to create client record');
    }
    
    $clientId = $conn->insert_id;
    
    // Handle multiple file uploads
    if (!isset($_FILES['id_documents'])) {
        throw new Exception('No ID documents were uploaded');
    }

    $files = $_FILES['id_documents'];
    $uploadedPaths = [];
    
    // Create base upload directory if it doesn't exist
    $baseUploadDir = __DIR__ . '/../public/uploads/ids/';
    if (!file_exists($baseUploadDir)) {
        mkdir($baseUploadDir, 0755, true);
    }

    // Process each uploaded file
    for ($i = 0; $i < count($files['name']); $i++) {
        $fileError = $files['error'][$i];
        $fileSize = $files['size'][$i];
        $fileTmpName = $files['tmp_name'][$i];
        $fileExtension = pathinfo($files['name'][$i], PATHINFO_EXTENSION);

        // Validate file upload
        if ($fileError !== UPLOAD_ERR_OK) {
            throw new Exception('File upload failed with error code: ' . $fileError);
        }

        // Maximum file size (5MB)
        $maxFileSize = 5 * 1024 * 1024;
        if ($fileSize > $maxFileSize) {
            throw new Exception('File size exceeds limit of 5MB');
        }

        // Generate filename using client ID and index
        $finalFileName = $clientId . '_' . ($i + 1) . '.' . $fileExtension;
        $finalFilePath = $baseUploadDir . $finalFileName;
        
        // Move uploaded file
        if (!move_uploaded_file($fileTmpName, $finalFilePath)) {
            throw new Exception('Failed to move uploaded file');
        }

        $relativePath = '/uploads/ids/' . $finalFileName;
        $uploadedPaths[] = $relativePath;

        // Update client record with the ID copy path
        // If this is the first/main ID document
        if ($i === 0) {
            $updateStmt = $conn->prepare("UPDATE clients SET id_copy_path = ? WHERE id = ?");
            $updateStmt->bind_param("si", $relativePath, $clientId);
            if (!$updateStmt->execute()) {
                throw new Exception('Failed to update client record with ID copy path');
            }
            $updateStmt->close();
        }
    }

    // Commit transaction
    $conn->commit();

    // Success response
    $response['success'] = true;
    $response['message'] = 'Client created and ID documents uploaded successfully';
    $response['client_id'] = $clientId;
    $response['uploaded_files'] = $uploadedPaths;

} catch (Exception $e) {
    // Rollback transaction on error
    if ($conn->connect_error === false) {
        $conn->rollback();
    }
    $response['message'] = $e->getMessage();
} finally {
    // Close database connection
    $conn->close();
}

// Send JSON response
echo json_encode($response); 