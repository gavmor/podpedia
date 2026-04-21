package main

import "testing"

func TestValidateDownloadRequest_Valid(t *testing.T) {
	cases := []struct {
		url  string
		dest string
	}{
		{"http://example.com/ep.mp3", "/tmp/ep.mp3"},
		{"https://cdn.example.com/audio.mp3", "output/audio.mp3"},
	}
	for _, tc := range cases {
		err := validateDownloadRequest(tc.url, tc.dest)
		if err != nil {
			t.Errorf("validateDownloadRequest(%q, %q) returned unexpected error: %v", tc.url, tc.dest, err)
		}
	}
}

func TestValidateDownloadRequest_MissingURL(t *testing.T) {
	err := validateDownloadRequest("", "/tmp/out.mp3")
	if err == nil {
		t.Error("want error for empty url, got nil")
	}
	if err.Error() != "url required" {
		t.Errorf("want error %q, got %q", "url required", err.Error())
	}
}

func TestValidateDownloadRequest_MissingDest(t *testing.T) {
	err := validateDownloadRequest("http://example.com/ep.mp3", "")
	if err == nil {
		t.Error("want error for empty dest, got nil")
	}
	if err.Error() != "dest required" {
		t.Errorf("want error %q, got %q", "dest required", err.Error())
	}
}

func TestValidateDownloadRequest_NonHTTPUrl(t *testing.T) {
	cases := []string{
		"ftp://example.com/ep.mp3",
		"file:///local/path.mp3",
		"just-a-filename.mp3",
		"/absolute/path.mp3",
	}
	for _, url := range cases {
		err := validateDownloadRequest(url, "/tmp/out.mp3")
		if err == nil {
			t.Errorf("want error for non-http url %q, got nil", url)
		}
		if err.Error() != "url must be http(s)" {
			t.Errorf("want error %q for url %q, got %q", "url must be http(s)", url, err.Error())
		}
	}
}

func TestValidateDownloadRequest_URLTakesPriorityOverDest(t *testing.T) {
	// When url is empty, should error on url before checking dest
	err := validateDownloadRequest("", "")
	if err == nil {
		t.Error("want error, got nil")
	}
	if err.Error() != "url required" {
		t.Errorf("want url error first, got %q", err.Error())
	}
}
