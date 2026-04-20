package pipeline

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAudioMetadata(t *testing.T) {
	tests := []struct {
		name          string
		contentLength string
		acceptRanges  string
		wantSize      int64
		wantRange     bool
		wantErr       bool
	}{
		{
			name:          "standard server with range support",
			contentLength: "1024",
			acceptRanges:  "bytes",
			wantSize:      1024,
			wantRange:     true,
			wantErr:       false,
		},
		{
			name:          "server without range support",
			contentLength: "2048",
			acceptRanges:  "",
			wantSize:      2048,
			wantRange:     false,
			wantErr:       false,
		},
		{
			name:          "invalid content length",
			contentLength: "invalid",
			acceptRanges:  "bytes",
			wantSize:      0,
			wantRange:     false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead {
					t.Errorf("Expected method HEAD, got %s", r.Method)
				}
				w.Header().Set("Content-Length", tt.contentLength)
				if tt.acceptRanges != "" {
					w.Header().Set("Accept-Ranges", tt.acceptRanges)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			size, supportsRange, err := GetAudioMetadata(ts.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAudioMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if size != tt.wantSize {
				t.Errorf("GetAudioMetadata() size = %v, want %v", size, tt.wantSize)
			}
			if supportsRange != tt.wantRange {
				t.Errorf("GetAudioMetadata() supportsRange = %v, want %v", supportsRange, tt.wantRange)
			}
		})
	}
}
