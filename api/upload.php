<?php
// Enable error reporting for debugging
error_reporting(E_ALL);
ini_set('display_errors', 1);

// Set memory and execution time for large files
ini_set('memory_limit', '512M');
ini_set('max_execution_time', 600);
ini_set('max_input_time', 600);

// Enable CORS
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST');
header('Access-Control-Allow-Headers: Content-Type, Accept');

// Handle GET requests for file serving
if ($_SERVER['REQUEST_METHOD'] === 'GET') {
    try {
        // Check if path parameter is provided
        if (!isset($_GET['path'])) {
            throw new Exception('No file path provided');
        }

        // Get and sanitize the file path
        $requestedPath = $_GET['path'];
        
        // Remove any parent directory references for security
        $requestedPath = str_replace('..', '', $requestedPath);
        $requestedPath = str_replace('//', '/', $requestedPath);
        
        // Construct the full file path
        //$filePath = __DIR__ . '/../files/' . $requestedPath;
        $filePath = __DIR__ . '/../mig_files/' . $requestedPath;

        // Check if file exists
        if (!file_exists($filePath)) {
            throw new Exception('File not found');
        }

        // Get file information
        $fileInfo = pathinfo($filePath);
        $fileName = $fileInfo['basename'];

        // Set content type based on file extension
        $extension = strtolower($fileInfo['extension']);
        $contentType = 'application/octet-stream'; // default

        // Define allowed content types
        $contentTypes = [
            'pdf'  => 'application/pdf',
            'jpg'  => 'image/jpeg',
            'jpeg' => 'image/jpeg',
            'png'  => 'image/png',
            'gif'  => 'image/gif',
            'bmp'  => 'image/bmp',
            'webp' => 'image/webp',
            'svg'  => 'image/svg+xml',
            'mp4'  => 'video/mp4',
            'webm' => 'video/webm',
            'avi'  => 'video/x-msvideo',
            'mov'  => 'video/quicktime',
            'mp3'  => 'audio/mpeg',
            'wav'  => 'audio/wav',
            'ogg'  => 'audio/ogg',
            'doc'  => 'application/msword',
            'docx' => 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
            'xls'  => 'application/vnd.ms-excel',
            'xlsx' => 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
            'ppt'  => 'application/vnd.ms-powerpoint',
            'pptx' => 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
            'txt'  => 'text/plain',
            'zip'  => 'application/zip',
            'rar'  => 'application/x-rar-compressed',
            '7z'   => 'application/x-7z-compressed'
        ];

        // Set the appropriate content type if known
        if (isset($contentTypes[$extension])) {
            $contentType = $contentTypes[$extension];
        }

        // Set headers for file download
        header('Content-Type: ' . $contentType);
        header('Content-Disposition: inline; filename="' . $fileName . '"');
        header('Content-Length: ' . filesize($filePath));
        header('Cache-Control: public, max-age=86400');
        
        // Clear output buffer
        if (ob_get_level()) {
            ob_end_clean();
        }

        // Read and output file
        readfile($filePath);
        exit;

    } catch (Exception $e) {
        // Clear any output that might have been sent
        if (ob_get_level()) {
            ob_end_clean();
        }

        // Set status code and content type
        http_response_code(404);
        header('Content-Type: application/json');

        // Send error response
        echo json_encode([
            'success' => false,
            'message' => $e->getMessage()
        ]);
        exit;
    }
}

// Handle POST requests for file upload
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Set JSON content type for upload responses
    header('Content-Type: application/json');

    // Response array
    $response = [
        'success' => false,
        'message' => '',
        'file_path' => ''
    ];

    try {
        // Log incoming request for debugging
        error_log('Upload request received. POST data: ' . print_r($_POST, true));
        error_log('Files data: ' . print_r($_FILES, true));

        // Check if file was uploaded
        if (!isset($_FILES['file'])) {
            throw new Exception('No file was uploaded');
        }

        // Get the file details
        $file = $_FILES['file'];
        $fileName = $file['name'];
        $fileTmpName = $file['tmp_name'];
        $fileError = $file['error'];
        $fileSize = $file['size'];

        // Get destination folder and filename from POST data
        $destinationFolder = isset($_POST['destination_folder']) ? trim($_POST['destination_folder']) : 'uploads';
        $customFileName = isset($_POST['custom_filename']) ? trim($_POST['custom_filename']) : '';

        // Validate file upload
        if ($fileError !== UPLOAD_ERR_OK) {
            throw new Exception('File upload failed with error code: ' . $fileError);
        }

        // Maximum file size (100MB)
        $maxFileSize = 100 * 1024 * 1024;
        if ($fileSize > $maxFileSize) {
            throw new Exception('File size exceeds limit of 100MB');
        }

        // Create base upload directory if it doesn't exist
        $baseUploadDir = __DIR__ . '/../mig_files/';
        error_log('Base upload directory: ' . $baseUploadDir);
        
        if (!file_exists($baseUploadDir)) {
            error_log('Creating base upload directory...');
            if (!mkdir($baseUploadDir, 0755, true)) {
                $error = error_get_last();
                throw new Exception('Failed to create base upload directory: ' . ($error['message'] ?? 'Unknown error'));
            }
        }

        // Check if base directory is writable
        if (!is_writable($baseUploadDir)) {
            throw new Exception('Base upload directory is not writable: ' . $baseUploadDir);
        }

        // Create and validate destination folder path
        $destinationPath = $baseUploadDir . rtrim($destinationFolder, '/') . '/';
        error_log('Destination path: ' . $destinationPath);
        
        if (!file_exists($destinationPath)) {
            error_log('Creating destination directory...');
            if (!mkdir($destinationPath, 0755, true)) {
                $error = error_get_last();
                throw new Exception('Failed to create destination directory: ' . ($error['message'] ?? 'Unknown error'));
            }
        }

        // Check if destination directory is writable
        if (!is_writable($destinationPath)) {
            throw new Exception('Destination directory is not writable: ' . $destinationPath);
        }

        // Generate final filename
        if (empty($customFileName)) {
            // Generate unique filename if no custom name provided
            $fileExtension = pathinfo($fileName, PATHINFO_EXTENSION);
            $finalFileName = uniqid() . '.' . $fileExtension;
        } else {
            $finalFileName = $customFileName;
        }

        // Full path for the file
        $finalFilePath = $destinationPath . $finalFileName;
        error_log('Final file path: ' . $finalFilePath);

        // Move uploaded file
        if (!move_uploaded_file($fileTmpName, $finalFilePath)) {
            $error = error_get_last();
            throw new Exception('Failed to move uploaded file from ' . $fileTmpName . ' to ' . $finalFilePath . ': ' . ($error['message'] ?? 'Unknown error'));
        }

        // Set proper permissions
        if (!chmod($finalFilePath, 0644)) {
            error_log('Warning: Failed to set file permissions for ' . $finalFilePath);
        }

        // Success response with path that points back to this same file
        $response['success'] = true;
        $response['message'] = 'File uploaded successfully';
        $response['file_path'] = '/api/upload.php?path=' . $destinationFolder . '/' . $finalFileName;

    } catch (Exception $e) {
        error_log('Upload error: ' . $e->getMessage());
        error_log('Upload error trace: ' . $e->getTraceAsString());
        $response['message'] = $e->getMessage();
    }

    // Ensure no output before JSON
    if (ob_get_length()) ob_clean();

    // Send JSON response
    echo json_encode($response);
    exit;
}

// If neither GET nor POST, return method not allowed
header('Content-Type: application/json');
http_response_code(405);
echo json_encode([
    'success' => false,
    'message' => 'Method not allowed'
]);
exit; 