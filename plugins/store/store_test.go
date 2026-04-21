package main

import "testing"

func TestSlug_AlphanumericPassThrough(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"Hello123", "Hello123"},
		{"abc-def", "abc-def"},
		{"ABC", "ABC"},
	}
	for _, tc := range cases {
		got := slug(tc.input)
		if got != tc.want {
			t.Errorf("slug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSlug_SpecialCharsReplacedWithUnderscore(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"hello world", "hello_world"},
		{"foo/bar", "foo_bar"},
		{"http://example.com/ep?id=1", "http___example_com_ep_id_1"},
		{"ep 001: intro!", "ep_001__intro_"},
	}
	for _, tc := range cases {
		got := slug(tc.input)
		if got != tc.want {
			t.Errorf("slug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSlug_EmptyString(t *testing.T) {
	if got := slug(""); got != "" {
		t.Errorf("slug(\"\") = %q, want \"\"", got)
	}
}

func TestSlug_HyphenPreserved(t *testing.T) {
	got := slug("my-episode-123")
	want := "my-episode-123"
	if got != want {
		t.Errorf("slug(%q) = %q, want %q", "my-episode-123", got, want)
	}
}

func TestSlug_OnlySpecialChars(t *testing.T) {
	got := slug("!@#$%^&*()")
	want := "__________"
	if got != want {
		t.Errorf("slug(%q) = %q, want %q", "!@#$%^&*()", got, want)
	}
}
