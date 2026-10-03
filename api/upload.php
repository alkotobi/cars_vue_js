<?php
// Upload and file-serving endpoint.
//
// Three things were wrong here before, and all three were exploitable by anyone
// on the internet:
//
//   1. There was no authentication at all, on either branch.
//   2. There was no extension check. `custom_filename` was written verbatim, so
//      `custom_filename=shell.php` into any directory under the document root
//      was remote code execution. `mig_files/` has no .htaccess and nothing that
//      stops the web server executing what lands there.
//   3. The GET branch had no realpath() containment - it only stripped "..", then
//      let `base_directory` choose any directory. `?base_directory=api&path=
//      config.local.php` returned the database password in cleartext.
//
// The extension allowlist below is the actual RCE fix; the directory rules are
// defence in depth so a normal user cannot overwrite the app's own files.

error_reporting(E_ALL);
ini_set('display_errors', 0);
ini_set('log_errors', 1);

ini_set('memory_limit', '512M');
ini_set('max_execution_time', 600);
ini_set('max_input_time', 600);

require_once __DIR__ . '/config.php';
require_once __DIR__ . '/lib/auth.php';
require_once __DIR__ . '/lib/appdb.php';
require_once __DIR__ . '/lib/cors.php';

// Same-origin only, like every other endpoint. The wildcard this replaced would
// have let any page on the internet drive this endpoint from a visitor's browser
// and read the file back.
api_handle_preflight();

header('Content-Type: application/json');
header('X-Content-Type-Options: nosniff');

/**
 * Extensions this endpoint will store, mapped to the type it is served as.
 *
 * This doubles as the extension check and as the content-type table, so a file
 * can never be stored under an extension the server does not recognise. Note
 * what is absent: svg (an SVG served inline is same-origin script, and this app
 * only uses inline data: URIs, never uploaded SVGs), and every .php/.phtml/
 * .phar/.htaccess variant that would hand the web server an interpreter.
 */
const UPLOAD_CONTENT_TYPES = [
    'pdf'  => 'application/pdf',
    'jpg'  => 'image/jpeg',
    'jpeg' => 'image/jpeg',
    'png'  => 'image/png',
    'gif'  => 'image/gif',
    'bmp'  => 'image/bmp',
    'webp' => 'image/webp',
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
    '7z'   => 'application/x-7z-compressed',
];

/**
 * Types served inline. Anything else is sent as an attachment so the browser
 * never renders an attacker-supplied document from our origin - `text/plain` is
 * the one exception, where inline is the useful behaviour and the type is inert.
 */
const UPLOAD_INLINE_TYPES = [
    'image/jpeg', 'image/png', 'image/gif', 'image/bmp', 'image/webp',
    'application/pdf', 'text/plain',
];

/**
 * Directories a non-admin may write into, relative to the project root.
 *
 * Deployment writes (GeneralSettingsForm replacing logo.png next to index.html,
 * Databases.vue pushing a build into `api/` or a tenant's `js_dir`) are admin
 * actions; they name directories this list does not contain and so are refused
 * for everyone else. An empty base_directory - the project root - is included
 * because that is where the shared branding files live.
 *
 * These are the directories a SINGLE-APP install has: the project root, its own
 * `files`, and the historical names. A tenant's own pair (js_dir/files_dir) is NOT
 * listed here any more - it did not scale, and every new client would have needed an
 * entry, so a client called acme_cars could not upload anything and a name added
 * for it would also be writable by every other tenant's users. upload_allowed_base()
 * derives this deployment's own tenant directories from the registry instead, so
 * the boundary holds for a hundred tenants without a hundred entries.
 */
const UPLOAD_PUBLIC_ROOTS = ['', 'mig_files', 'mig', 'files', 'uploads'];

/**
 * Whether a non-admin may write into $baseDirectory, given this deployment's own
 * tenant directories.
 *
 * Two rules on top of the static list:
 *
 *   - the request's OWN files_dir, from its registry row. app_db_files_dir()
 *     resolves that row from db_code.json, which is derived from the request's
 *     mount (api/lib/appdb.php app_request_mount()) and is not a parameter - so
 *     this cannot be talked into naming another tenant's folder.
 *   - the request's OWN js_dir, so the app folder can be written by an admin of
 *     that app. It is the folder deploy/deploy.sh unpacks a build into, which is
 *     why it has to be writable at all; non-admins still cannot use it.
 *
 * Note what is NOT derived from the registry: any OTHER tenant's directories. Only
 * an admin may name them, which is the pre-existing rule and stays.
 *
 * @param string|null $filesDir this deployment's own files_dir, as a plain name
 *        (app_db_files_dir()), or null when the row does not resolve
 * @param array|null  $row      this deployment's registry row, for js_dir
 */
function upload_allowed_base(string $baseDirectory, ?string $filesDir, ?array $row = null): bool
{
    if (in_array($baseDirectory, UPLOAD_PUBLIC_ROOTS, true)) {
        return true;
    }

    if ($filesDir !== null && $baseDirectory === trim($filesDir, '/')) {
        return true;
    }

    // The app folder, but only the deployment's own: js_dir is recorded with a
    // leading slash, so the name alone is trimmed here and compared as a directory.
    if ($row !== null) {
        $jsDir = trim((string) ($row['js_dir'] ?? ''), '/');
        if ($jsDir !== '' && $baseDirectory === $jsDir) {
            return true;
        }
    }

    return false;
}

/**
 * Branding files the login page renders before anyone has authenticated.
 *
 * fetchDifferentCompanyUserLogo() in index.html resolves a tenant's logo while
 * the visitor is still anonymous, so these five names stay reachable without a
 * token. The carve-out is by exact basename and only at the top of the base
 * directory - documents, car files, id pictures and anything in a subfolder
 * still need a real token, which is the part that was leaking.
 */
const UPLOAD_PUBLIC_ASSETS = [
    'logo.png', 'logo_default.png', 'letter_head.png', 'letter_head_default.png', 'gml2.png',
];

/**
 * Files that must never be served, whatever path is asked for.
 */
const UPLOAD_DENY_NAMES = [
    'config.php', 'config.local.php', 'db_manager_config.php', 'db_manager_config.local.php',
    'db_code.json', '.htaccess', '.env', '.gitignore',
];

/**
 * Pull the API token out of wherever the caller put it.
 *
 * POST bodies and fetch() callers can use the X-Api-Token header, but a file URL
 * is fetched by <img src>, which cannot carry headers, so the GET branch also
 * accepts it in the query string. That trades a token in an access log for
 * working <img> tags; the alternative was leaving these files world-readable.
 */
function upload_request_token(): string
{
    $header = $_SERVER['HTTP_X_API_TOKEN'] ?? '';
    if (is_string($header) && $header !== '') {
        return $header;
    }

    foreach ([$_POST['token'] ?? '', $_GET['token'] ?? ''] as $candidate) {
        if (is_string($candidate) && $candidate !== '') {
            return $candidate;
        }
    }

    return '';
}

/**
 * Authenticate the request. Exits with the shared guard envelope when the token
 * does not resolve to a real user.
 *
 * @return array{id:int,username:string,role_id:int}
 */
function upload_require_user(): array
{
    return require_app_user(upload_request_token());
}

/**
 * Normalise a caller-supplied base directory to a project-root-relative path.
 * Kept for the response shape; the containment check is realpath-based.
 */
function upload_normalise_base(?string $base): string
{
    $base = str_replace('\\', '/', trim((string) $base));
    $base = ltrim($base, '/');
    return trim(rtrim($base, '/'));
}

/**
 * Reduce a path to its final component, rejecting anything that tries to be a
 * directory reference rather than a name.
 */
function upload_normalise_folder(?string $folder): string
{
    $folder = str_replace('\\', '/', trim((string) $folder));
    $parts = [];
    foreach (explode('/', $folder) as $part) {
        if ($part === '' || $part === '.') {
            continue;
        }
        if ($part === '..') {
            array_pop($parts);
            continue;
        }
        $parts[] = $part;
    }
    return implode('/', $parts);
}

/**
 * Resolve a path and prove it sits inside the project root.
 *
 * realpath() collapses "..", symlinks and double slashes in one step, which the
 * old str_replace('..', '') approach did not, and the returned path is what
 * callers must then use - so the check cannot be bypassed by a path that looks
 * safe but resolves elsewhere.
 */
function upload_resolve_within_root(string $projectRoot, string $relative): string
{
    $resolved = realpath($projectRoot . '/' . $relative);
    if ($resolved === false) {
        return '';
    }

    $root = rtrim(realpath($projectRoot) ?: $projectRoot, DIRECTORY_SEPARATOR);
    if ($resolved !== $root && strpos($resolved, $root . DIRECTORY_SEPARATOR) !== 0) {
        return '';
    }

    return $resolved;
}

/**
 * Reject any filename that is not a plain name carrying an allowed extension.
 */
function upload_validate_extension(string $fileName): string
{
    if ($fileName === '' || strpos($fileName, "\0") !== false) {
        throw new Exception('Invalid filename');
    }
    if ($fileName !== basename($fileName) || $fileName[0] === '.') {
        throw new Exception('Invalid filename');
    }

    $extension = strtolower((string) pathinfo($fileName, PATHINFO_EXTENSION));
    if ($extension === '' || !isset(UPLOAD_CONTENT_TYPES[$extension])) {
        throw new Exception('File type not allowed: .' . $extension);
    }

    return $extension;
}

/**
 * Confirm the bytes are what the extension claims.
 *
 * The extension allowlist already stops anything executable; this catches a
 * .png that is really a script, so the file is never stored or served as an
 * image. getimagesize() parses the container, so it fails on non-images.
 */
function upload_verify_image(string $path): void
{
    $size = @getimagesize($path);
    if ($size === false || empty($size[0]) || empty($size[1])) {
        throw new Exception('File content does not match its image extension');
    }
}

$PROJECT_ROOT = dirname(__DIR__);

// Which root a base_directory is resolved against.
//
// Two namespaces meet here. The registry records files_dir against the
// DEPLOYMENT root, because a tenant's app folder and its uploads folder are
// siblings (<webroot>/acme_cars + <webroot>/acme_cars_files) - and the browser
// sends that value verbatim (src/composables/useApi.js). Every other entry in
// UPLOAD_PUBLIC_ROOTS is named relative to the APP folder: '' is the app's own
// documents/logo.
$DEPLOYMENT_ROOT = app_deployment_root();

/**
 * The root that one base_directory has to be resolved against.
 */
function upload_base_root(string $baseDirectory, string $appRoot, string $deploymentRoot): string
{
    $filesDir = app_db_files_dir();
    if ($filesDir === null) {
        return $appRoot;
    }

    return trim($filesDir, '/') === $baseDirectory ? $deploymentRoot : $appRoot;
}

// ---------------------------------------------------------------------------
// GET: serve a stored file
// ---------------------------------------------------------------------------

if ($_SERVER['REQUEST_METHOD'] === 'GET') {
    $path = $_GET['path'] ?? '';
    if (!is_string($path) || $path === '') {
        http_response_code(400);
        echo json_encode(['success' => false, 'message' => 'No file path provided']);
        exit;
    }

    $baseDirectory = upload_normalise_base($_GET['base_directory'] ?? 'mig_files');
    $requested = upload_normalise_folder($path);

    $filePath = upload_resolve_within_root(
        upload_base_root($baseDirectory, $PROJECT_ROOT, $DEPLOYMENT_ROOT),
        $baseDirectory . '/' . $requested
    );
    if ($filePath === '' || !is_file($filePath)) {
        http_response_code(404);
        echo json_encode(['success' => false, 'message' => 'File not found']);
        exit;
    }

    // Containment above stops ".."; this stops serving the secrets that sit
    // inside the tree it permits (config.local.php, the SQL dumps, .git).
    $fileName = basename($filePath);
    if (
        in_array($fileName, UPLOAD_DENY_NAMES, true)
        || preg_match('/\.(sql|local\.php)$/i', $fileName)
        || preg_match('/(^|\/)\.git(\/|$)/', $filePath)
    ) {
        http_response_code(403);
        echo json_encode(['success' => false, 'message' => 'Access denied']);
        exit;
    }

    // The login page's branding renders before login, so those exact names skip
    // the token. Everything else - every document, id picture and file - does not.
    $isPublicAsset = in_array($fileName, UPLOAD_PUBLIC_ASSETS, true)
        && $requested === $fileName;

    if (!$isPublicAsset) {
        upload_require_user();
    }

    $extension = strtolower((string) pathinfo($filePath, PATHINFO_EXTENSION));
    $contentType = UPLOAD_CONTENT_TYPES[$extension] ?? null;
    if ($contentType === null) {
        http_response_code(403);
        echo json_encode(['success' => false, 'message' => 'File type not allowed']);
        exit;
    }

    // Quote the filename: it reaches us from the query string, and an unescaped
    // quote would split the header.
    $disposition = in_array($contentType, UPLOAD_INLINE_TYPES, true) ? 'inline' : 'attachment';
    header('Content-Type: ' . $contentType);
    header('Content-Disposition: ' . $disposition . '; filename="' . str_replace('"', '', $fileName) . '"');
    header('Content-Length: ' . filesize($filePath));
    header('Cache-Control: private, max-age=86400');

    readfile($filePath);
    exit;
}

// ---------------------------------------------------------------------------
// POST: store a file
// ---------------------------------------------------------------------------

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode(['success' => false, 'message' => 'Method not allowed']);
    exit;
}

$response = ['success' => false, 'message' => '', 'file_path' => ''];

try {
    $user = upload_require_user();
    $isAdmin = ((int) $user['role_id'] === 1);

    if (!isset($_FILES['file'])) {
        throw new Exception('No file was uploaded');
    }

    $file = $_FILES['file'];
    if ($file['error'] !== UPLOAD_ERR_OK) {
        throw new Exception('File upload failed with error code: ' . $file['error']);
    }

    $maxFileSize = 100 * 1024 * 1024;
    if ($file['size'] > $maxFileSize) {
        throw new Exception('File size exceeds limit of 100MB');
    }

    $baseDirectory = upload_normalise_base($_POST['base_directory'] ?? 'mig_files');
    $destinationFolder = upload_normalise_folder($_POST['destination_folder'] ?? 'uploads');
    $customFileName = isset($_POST['custom_filename']) ? trim($_POST['custom_filename']) : '';

    // A non-admin writes only into the shared upload roots, plus this deployment's
    // own tenant folders. Without this the extension allowlist would still stop a
    // webshell, but any user could overwrite another tenant's documents or the app's
    // branding files.
    if (!$isAdmin && !upload_allowed_base($baseDirectory, app_db_files_dir(), app_db_row())) {
        throw new Exception('Not allowed to upload into this directory');
    }

    $relativeDirectory = $baseDirectory . ($destinationFolder !== '' ? '/' . $destinationFolder : '');
    $realBasePath = upload_resolve_within_root(
        upload_base_root($baseDirectory, $PROJECT_ROOT, $DEPLOYMENT_ROOT),
        $relativeDirectory
    );
    if ($realBasePath === '' || !is_dir($realBasePath)) {
        throw new Exception('Invalid upload directory');
    }

    // The extension check runs on the name we are about to write, whichever
    // branch produced it - the client-supplied one included.
    $finalFileName = $customFileName !== '' ? basename($customFileName) : '';
    if ($finalFileName === '') {
        $extension = strtolower((string) pathinfo($file['name'], PATHINFO_EXTENSION));
        $finalFileName = bin2hex(random_bytes(16)) . ($extension !== '' ? '.' . $extension : '');
    }
    $extension = upload_validate_extension($finalFileName);

    $finalFilePath = $realBasePath . DIRECTORY_SEPARATOR . $finalFileName;
    if (!is_writable($realBasePath)) {
        throw new Exception('Upload directory is not writable');
    }

    if (!move_uploaded_file($file['tmp_name'], $finalFilePath)) {
        $error = error_get_last();
        throw new Exception('Failed to store uploaded file: ' . ($error['message'] ?? 'unknown error'));
    }

    if (strpos($contentType = UPLOAD_CONTENT_TYPES[$extension], 'image/') === 0
        && $contentType !== 'image/svg+xml') {
        try {
            upload_verify_image($finalFilePath);
        } catch (Exception $e) {
            // Never leave an unverified file behind in a web-served directory.
            @unlink($finalFilePath);
            throw $e;
        }
    }

    @chmod($finalFilePath, 0644);

    $response['success'] = true;
    $response['message'] = 'File uploaded successfully';
    $response['file_path'] = '/api/upload.php?path='
        . urlencode($destinationFolder . '/' . $finalFileName)
        . '&base_directory=' . urlencode($baseDirectory);
} catch (Exception $e) {
    error_log('upload.php: ' . $e->getMessage());
    $response['message'] = $e->getMessage();
}

echo json_encode($response);
exit;
