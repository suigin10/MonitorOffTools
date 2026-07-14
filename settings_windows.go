//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const configFileName = "MonitorOffTools.ini"

type appSettings struct {
	Language                 string
	LeftClickEnabled         bool
	AutoOffMinutes           int
	PowerConflictNoticeShown bool
}

func defaultSettings() appSettings {
	return appSettings{
		Language:                 languageJapanese,
		LeftClickEnabled:         true,
		AutoOffMinutes:           defaultAutoOffMinutes,
		PowerConflictNoticeShown: false,
	}
}

func settingsFilePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), configFileName), nil
}

func loadAppSettings() (appSettings, string, error) {
	settings := defaultSettings()
	path, err := settingsFilePath()
	if err != nil {
		return settings, "", err
	}

	data, readErr := os.ReadFile(path)
	if readErr == nil {
		settings = parseSettingsINI(string(data), settings)
		return settings, path, nil
	}
	if !os.IsNotExist(readErr) {
		return settings, path, readErr
	}

	// v1.0.0以前のレジストリ設定を初回だけINIへ移行します。
	if value, ok := readLegacySettingDWORD(autoOffValueName); ok && isValidAutoOffMinutes(int(value)) {
		settings.AutoOffMinutes = int(value)
	}
	if value, ok := readLegacySettingDWORD(conflictNoticeValueName); ok {
		settings.PowerConflictNoticeShown = value != 0
	}

	if writeErr := saveAppSettings(path, settings); writeErr != nil {
		return settings, path, writeErr
	}
	return settings, path, nil
}

func parseSettingsINI(content string, defaults appSettings) appSettings {
	settings := defaults
	section := ""
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		if section != "settings" {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		switch key {
		case "language":
			value = strings.ToLower(value)
			if isValidLanguage(value) {
				settings.Language = value
			}
		case "leftclickenabled":
			if parsed, err := strconv.ParseBool(value); err == nil {
				settings.LeftClickEnabled = parsed
			}
		case "autooffminutes":
			if parsed, err := strconv.Atoi(value); err == nil && isValidAutoOffMinutes(parsed) {
				settings.AutoOffMinutes = parsed
			}
		case "powerconflictnoticeshown":
			if parsed, err := strconv.ParseBool(value); err == nil {
				settings.PowerConflictNoticeShown = parsed
			}
		}
	}
	return settings
}

func saveAppSettings(path string, settings appSettings) error {
	if path == "" {
		return fmt.Errorf("settings path is empty")
	}
	content := fmt.Sprintf(
		"; MonitorOffTools settings\r\n[Settings]\r\nLanguage=%s\r\nLeftClickEnabled=%t\r\nAutoOffMinutes=%d\r\nPowerConflictNoticeShown=%t\r\n",
		settings.Language,
		settings.LeftClickEnabled,
		settings.AutoOffMinutes,
		settings.PowerConflictNoticeShown,
	)
	return os.WriteFile(path, []byte(content), 0o644)
}

func readLegacySettingDWORD(valueName string) (uint32, bool) {
	if value, ok := readSettingsDWORDAt(settingsKeyPath, valueName); ok {
		return value, true
	}
	return readSettingsDWORDAt(legacySettingsKeyPath, valueName)
}
