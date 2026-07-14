//go:build windows

package main

import "testing"

func TestParseSettingsINI(t *testing.T) {
	defaults := defaultSettings()
	content := `[Settings]
Language=en
LeftClickEnabled=false
AutoOffMinutes=45
PowerConflictNoticeShown=true
`
	got := parseSettingsINI(content, defaults)
	if got.Language != languageEnglish {
		t.Fatalf("Language = %q, want %q", got.Language, languageEnglish)
	}
	if got.LeftClickEnabled {
		t.Fatal("LeftClickEnabled = true, want false")
	}
	if got.AutoOffMinutes != 45 {
		t.Fatalf("AutoOffMinutes = %d, want 45", got.AutoOffMinutes)
	}
	if !got.PowerConflictNoticeShown {
		t.Fatal("PowerConflictNoticeShown = false, want true")
	}
}

func TestParseSettingsINIRejectsInvalidValues(t *testing.T) {
	defaults := defaultSettings()
	content := `[Settings]
Language=xx
LeftClickEnabled=maybe
AutoOffMinutes=10
PowerConflictNoticeShown=unknown
`
	got := parseSettingsINI(content, defaults)
	if got != defaults {
		t.Fatalf("invalid settings changed defaults: got %+v, want %+v", got, defaults)
	}
}
