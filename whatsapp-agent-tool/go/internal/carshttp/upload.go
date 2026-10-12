// Logo/file upload for the cars app. The app has no JSON upload action: files go
// to api/upload.php as multipart with a token field (a form body cannot set the
// X-Api-Token header an <img> would need, so useApi.js sends the token as a
// field too). Mirrors src/composables/useApi.js uploadFile(): base_directory is
// always 'files' and the response's file_path carries the stored path in its
// ?path= query parameter.
package carshttp

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"whatsapp-agent-tool/internal/config"
)

// tenantUploadEndpoint is the tenant's upload.php, alongside api.php.
func tenantUploadEndpoint() string {
	return config.C.Cars.BaseURL + "/" + url.PathEscape(config.C.Cars.Tenant) + "/api/upload.php"
}

// uploadResponse is upload.php's JSON envelope (its own shape, not api.php's):
// success/message/file_path, or success:false with code/error on the auth guard.
type uploadResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FilePath string `json:"file_path"`
	Code     string `json:"code"`
	Error    string `json:"error"`
}

// UploadFile sends localPath to the tenant's upload.php under
// destinationFolder (relative to the tenant's 'files' base) and returns the
// stored relative path the app records (e.g. "uploads/ab12.png"). The file is
// read once and re-sent on the single auth retry withToken performs.
func UploadFile(localPath, destinationFolder string) (string, error) {
	return UploadFileNamed(localPath, destinationFolder, filepath.Base(localPath))
}

// UploadFileNamed is UploadFile with an explicit stored file name, for callers
// that must not collide (the app saves swifts as bill_<id>_<ts>.<ext>). The name
// is only a basename sent to upload.php; it is never used as a local path.
func UploadFileNamed(localPath, destinationFolder, fileName string) (string, error) {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", &ApiError{Code: "upload_failed", Message: "cannot read " + localPath + ": " + err.Error()}
	}
	return withToken(func(token string) (string, error) {
		return uploadOnce(token, fileName, data, destinationFolder)
	})
}

// uploadOnce performs one multipart POST with the given token.
func uploadOnce(token, fileName string, data []byte, destinationFolder string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", &ApiError{Code: "upload_failed", Message: err.Error()}
	}
	if _, err := part.Write(data); err != nil {
		return "", &ApiError{Code: "upload_failed", Message: err.Error()}
	}
	_ = writer.WriteField("base_directory", "files")
	_ = writer.WriteField("destination_folder", destinationFolder)
	_ = writer.WriteField("token", token)
	if err := writer.Close(); err != nil {
		return "", &ApiError{Code: "upload_failed", Message: err.Error()}
	}

	endpoint := tenantUploadEndpoint()
	req, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return "", &ApiError{Code: "network_error", Message: "cannot reach " + endpoint + ": " + err.Error()}
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Cards-Tenant", config.C.Cars.Tenant)
	req.Header.Set("CARDS-TENANT", config.C.Cars.Tenant)

	client := &http.Client{Timeout: requestTimeout}
	res, err := client.Do(req)
	if err != nil {
		return "", &ApiError{Code: "network_error", Message: "cannot reach " + endpoint + ": " + err.Error()}
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", &ApiError{Code: "network_error", Message: "cannot reach " + endpoint + ": " + err.Error()}
	}

	var resp uploadResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", &ApiError{Code: "bad_response", Message: "non-JSON response from upload.php"}
	}
	if resp.Code == "not_authenticated" {
		return "", &ApiError{Code: "not_authenticated", Message: "token rejected"}
	}
	if !resp.Success {
		msg := "upload failed"
		if resp.Message != "" {
			msg = resp.Message
		} else if resp.Error != "" {
			msg = resp.Error
		}
		return "", &ApiError{Code: "upload_failed", Message: msg}
	}
	return extractRelativePath(resp.FilePath), nil
}

// extractRelativePath turns upload.php's file_path back into the relative path
// the app stores. The server answers "/api/upload.php?path=uploads/x.png&
// base_directory=files"; we want "uploads/x.png" (useApi.js does the same with
// URLSearchParams). A value without a query string is returned unchanged.
func extractRelativePath(filePath string) string {
	if !strings.Contains(filePath, "?") {
		return filePath
	}
	parsed, err := url.Parse(filePath)
	if err != nil {
		return filePath
	}
	if p := parsed.Query().Get("path"); p != "" {
		return p
	}
	return filePath
}
