package main

import "testing"

func TestExtractIconUrls(t *testing.T) {
	html := `
    <link rel="icon" href="/icon.ico">
    <link rel="icon" href="/self-closing-icon-tag.ico"/>
    <link rel="shortcut icon" href="/shortcut_icon.ico">
    <link rel="apple-touch-icon" href="/apple-touch-icon.png">
  `
	whitelist := []string{
		"icon",
		"shortcut icon",
		"apple-touch-icon",
	}
	expected := []string{
		"/icon.ico",
		"/self-closing-icon-tag.ico",
		"/shortcut_icon.ico",
		"/apple-touch-icon.png",
	}

	iconUrls := extractIconUrls([]byte(html), whitelist)

	if len(iconUrls) != len(expected) {
		t.Fatal("expected", len(expected), "got", len(iconUrls))
	}

	for i, v := range expected {
		if v != iconUrls[i] {
			t.Error("expected", v, "got", iconUrls[i])
		}
	}
}
