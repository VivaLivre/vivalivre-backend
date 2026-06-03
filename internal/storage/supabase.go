package storage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// UploadToSupabase uploads a file to the Supabase Storage bucket and returns
// the public URL. It uses the Supabase REST API (not the client SDK).
//
// Parameters:
//   - bucket: the storage bucket name (e.g. "bathroom_photos")
//   - fileReader: the file content reader
//   - originalFilename: the original name of the uploaded file (used for extension)
//   - contentType: the MIME type of the file (e.g. "image/jpeg")
func UploadToSupabase(bucket string, fileReader io.Reader, originalFilename string, contentType string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return "", fmt.Errorf("SUPABASE_URL and SUPABASE_KEY environment variables are required")
	}

	// Generate a unique filename preserving the original extension
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".jpg" // fallback
	}
	objectName := uuid.New().String() + ext

	// Build the Supabase Storage REST API URL
	// POST /storage/v1/object/{bucket}/{objectName}
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucket, objectName)

	req, err := http.NewRequest(http.MethodPost, uploadURL, fileReader)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload to Supabase Storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Build public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucket, objectName)

	return publicURL, nil
}
