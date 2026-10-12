package carshttp

import "testing"

func TestExtractRelativePath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "extracts the path query parameter",
			filePath: "/api/upload.php?path=uploads%2Fab12.png&base_directory=files",
			want:     "uploads/ab12.png",
		},
		{
			name:     "keeps unescaped slashes in the path parameter",
			filePath: "/api/upload.php?path=uploads/ab12.png&base_directory=files",
			want:     "uploads/ab12.png",
		},
		{
			name:     "returns a bare value unchanged",
			filePath: "uploads/ab12.png",
			want:     "uploads/ab12.png",
		},
		{
			name:     "falls back to the whole value when path is absent",
			filePath: "/api/upload.php?base_directory=files",
			want:     "/api/upload.php?base_directory=files",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractRelativePath(tt.filePath); got != tt.want {
				t.Errorf("extractRelativePath(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}
