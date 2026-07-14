package main

import "fmt"

const (
	languageJapanese = "ja"
	languageEnglish  = "en"
)

var localizedText = map[string]map[string]string{
	languageJapanese: {
		"startupFailed":           "起動に失敗しました。",
		"settingsLoadFailed":      "設定ファイルの読み込みまたは作成に失敗しました。設定はこの起動中のみ有効になる場合があります。",
		"settingsSaveFailed":      "設定ファイルの保存に失敗しました。",
		"notifyMonitorOffTitle":   "モニターオフ",
		"notifyMonitorOffBody":    "すべてのモニターをオフにします。",
		"notifyInputUnavailable":  "入力状態を取得できないため、再点灯防止監視を開始できませんでした。",
		"menuTurnOff":             "モニターをオフ",
		"menuAutoOff":             "自動モニターオフ",
		"menuNone":                "なし",
		"menu15Minutes":           "15分",
		"menu30Minutes":           "30分",
		"menu45Minutes":           "45分",
		"menu1Hour":               "1時間",
		"menu2Hours":              "2時間",
		"menuLeftClick":           "左クリックでモニターをオフ",
		"menuLanguage":            "言語 / Language",
		"menuJapanese":            "日本語",
		"menuEnglish":             "English",
		"menuPowerSettings":       "Windowsの画面オフ設定を開く",
		"menuStartupAdd":          "スタートアップに登録",
		"menuStartupRemove":       "スタートアップ登録を解除",
		"menuExit":                "終了",
		"statusPending":           "状態: 2秒後にモニターオフ",
		"statusProtection":        "状態: 再点灯防止監視中（30分固定）",
		"statusIdle":              "状態: 待機中 / 自動オフ %s",
		"tooltipPending":          "%s - 2秒後に消灯",
		"tooltipProtection":       "%s - 再点灯防止監視中",
		"tooltipIdle":             "%s - 自動オフ %s",
		"notifyAutoTitle":         "自動モニターオフ",
		"notifyAutoChanged":       "%sに変更しました。",
		"notifyLeftClickTitle":    "左クリック操作",
		"notifyLeftClickEnabled":  "左クリックでモニターをオフにします。",
		"notifyLeftClickDisabled": "左クリックによるモニターオフを無効にしました。",
		"notifyLanguageTitle":     "言語",
		"notifyLanguageChanged":   "表示言語を日本語に変更しました。",
		"notifyStartupTitle":      "スタートアップ",
		"notifyStartupAdded":      "Windowsログオン時に起動するよう登録しました。",
		"notifyStartupRemoved":    "スタートアップ登録を解除しました。",
		"startupChangeFailed":     "スタートアップ設定の変更に失敗しました。",
		"powerConflictTitle":      "Windows省電力との競合防止",
		"powerConflictBody":       "本アプリの自動オフを使う場合、Windows標準の画面オフは「なし」または本アプリより長い時間を推奨します。",
		"openPowerSettingsFailed": "Windowsの電源設定を開けませんでした。",
	},
	languageEnglish: {
		"startupFailed":           "Failed to start.",
		"settingsLoadFailed":      "Failed to read or create the settings file. Changes may apply only for this session.",
		"settingsSaveFailed":      "Failed to save the settings file.",
		"notifyMonitorOffTitle":   "Monitor off",
		"notifyMonitorOffBody":    "Turning off all monitors.",
		"notifyInputUnavailable":  "Could not read the input state, so wake-prevention monitoring could not be started.",
		"menuTurnOff":             "Turn off monitors",
		"menuAutoOff":             "Idle auto-off",
		"menuNone":                "None",
		"menu15Minutes":           "15 minutes",
		"menu30Minutes":           "30 minutes",
		"menu45Minutes":           "45 minutes",
		"menu1Hour":               "1 hour",
		"menu2Hours":              "2 hours",
		"menuLeftClick":           "Turn off monitors with left click",
		"menuLanguage":            "Language / 言語",
		"menuJapanese":            "日本語",
		"menuEnglish":             "English",
		"menuPowerSettings":       "Open Windows display power settings",
		"menuStartupAdd":          "Add to startup",
		"menuStartupRemove":       "Remove from startup",
		"menuExit":                "Exit",
		"statusPending":           "Status: monitors turn off in 2 seconds",
		"statusProtection":        "Status: wake-prevention monitoring (fixed at 30 min)",
		"statusIdle":              "Status: idle / auto-off %s",
		"tooltipPending":          "%s - turning off in 2 seconds",
		"tooltipProtection":       "%s - wake-prevention monitoring",
		"tooltipIdle":             "%s - auto-off %s",
		"notifyAutoTitle":         "Idle auto-off",
		"notifyAutoChanged":       "Changed to %s.",
		"notifyLeftClickTitle":    "Left-click action",
		"notifyLeftClickEnabled":  "Left-clicking the tray icon will turn off the monitors.",
		"notifyLeftClickDisabled": "Monitor-off by left click has been disabled.",
		"notifyLanguageTitle":     "Language",
		"notifyLanguageChanged":   "Display language changed to English.",
		"notifyStartupTitle":      "Startup",
		"notifyStartupAdded":      "The app will start when you sign in to Windows.",
		"notifyStartupRemoved":    "Removed from Windows startup.",
		"startupChangeFailed":     "Failed to change the startup setting.",
		"powerConflictTitle":      "Avoiding Windows power-setting conflicts",
		"powerConflictBody":       "When using this app's idle auto-off, set Windows display-off to Never or to a longer duration than this app.",
		"openPowerSettingsFailed": "Could not open Windows power settings.",
	},
}

func isValidLanguage(language string) bool {
	return language == languageJapanese || language == languageEnglish
}

func translate(language, key string) string {
	if !isValidLanguage(language) {
		language = languageJapanese
	}
	if value, ok := localizedText[language][key]; ok {
		return value
	}
	if value, ok := localizedText[languageJapanese][key]; ok {
		return value
	}
	return key
}

func formatMinutes(language string, minutes int) string {
	if language == languageEnglish {
		switch minutes {
		case 0:
			return "None"
		case 60:
			return "1 hour"
		case 120:
			return "2 hours"
		default:
			return fmt.Sprintf("%d minutes", minutes)
		}
	}

	switch minutes {
	case 0:
		return "なし"
	case 60:
		return "1時間"
	case 120:
		return "2時間"
	default:
		return fmt.Sprintf("%d分", minutes)
	}
}
