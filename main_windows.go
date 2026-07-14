//go:build windows

package main

import (
	"embed"
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	appName     = "MonitorOffTools"
	mutexName   = `Local\MonitorOffTool.SingleInstance`
	windowClass = "MonitorOffTools.HiddenWindow"

	wmDestroy       = 0x0002
	wmTimer         = 0x0113
	wmApp           = 0x8000
	wmTrayCallback  = wmApp + 1
	wmLButtonUp     = 0x0202
	wmRButtonUp     = 0x0205
	wmContextMenu   = 0x007B
	wmSysCommand    = 0x0112
	scMonitorPower  = 0xF170
	monitorPowerOff = 2

	timerMonitorOffDelay  = 1
	timerProtection       = 2
	timerWakeDetection    = 3
	timerAutoOffCheck     = 4
	timerMonitorOffUnlock = 5

	monitorOffDelayMS    = 2_000
	monitorOffCooldownMS = 2_000
	wakeInputGuardMS     = 1_500
	protectionPeriodMS   = 30 * 60 * 1_000
	wakeDetectionMS      = 250
	autoOffCheckMS       = 1_000

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage  = 0x00000001
	nifIcon     = 0x00000002
	nifTip      = 0x00000004
	nifInfo     = 0x00000010
	niifInfo    = 0x00000001
	niifNoSound = 0x00000010

	mfString    = 0x00000000
	mfGrayed    = 0x00000001
	mfChecked   = 0x00000008
	mfPopup     = 0x00000010
	mfSeparator = 0x00000800

	tpmRightButton = 0x0002
	tpmNonotify    = 0x0080
	tpmReturnCmd   = 0x0100

	cmdTurnOff         = 1001
	cmdStatus          = 1002
	cmdStartup         = 1003
	cmdExit            = 1004
	cmdPowerSettings   = 1005
	cmdToggleLeftClick = 1006

	cmdAutoNone = 1100
	cmdAuto15   = 1101
	cmdAuto30   = 1102
	cmdAuto45   = 1103
	cmdAuto60   = 1104
	cmdAuto120  = 1105

	cmdLanguageJapanese = 1200
	cmdLanguageEnglish  = 1201

	smtoAbortIfHung = 0x0002
	smtoErrorOnExit = 0x0020

	errorAlreadyExists = 183

	regOptionNonVolatile = 0
	regSZ                = 1
	regDWORD             = 4
	keyQueryValue        = 0x0001
	keySetValue          = 0x0002

	lrDefaultColor = 0x0000
)

const (
	hkeyCurrentUser         = uintptr(0x80000001)
	runKeyPath              = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValueName            = "MonitorOffTools"
	legacyRunValueName      = "MonitorOffTool"
	settingsKeyPath         = `Software\MonitorOffTools`
	legacySettingsKeyPath   = `Software\MonitorOffTool`
	autoOffValueName        = "AutoOffMinutes"
	conflictNoticeValueName = "PowerConflictNoticeShown"
	defaultAutoOffMinutes   = 0
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procSetTimer                 = user32.NewProc("SetTimer")
	procKillTimer                = user32.NewProc("KillTimer")
	procSendMessageTimeoutW      = user32.NewProc("SendMessageTimeoutW")
	procGetLastInputInfo         = user32.NewProc("GetLastInputInfo")
	procCreatePopupMenu          = user32.NewProc("CreatePopupMenu")
	procAppendMenuW              = user32.NewProc("AppendMenuW")
	procTrackPopupMenu           = user32.NewProc("TrackPopupMenu")
	procDestroyMenu              = user32.NewProc("DestroyMenu")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	procDestroyIcon              = user32.NewProc("DestroyIcon")
	procLoadIconW                = user32.NewProc("LoadIconW")
	procLoadCursorW              = user32.NewProc("LoadCursorW")
	procMessageBoxW              = user32.NewProc("MessageBoxW")
	procRegisterWindowMessageW   = user32.NewProc("RegisterWindowMessageW")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procCreateMutexW     = kernel32.NewProc("CreateMutexW")
	procCloseHandle      = kernel32.NewProc("CloseHandle")
	procGetTickCount     = kernel32.NewProc("GetTickCount")

	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	procShellExecuteW    = shell32.NewProc("ShellExecuteW")

	procRegOpenKeyExW    = advapi32.NewProc("RegOpenKeyExW")
	procRegCreateKeyExW  = advapi32.NewProc("RegCreateKeyExW")
	procRegQueryValueExW = advapi32.NewProc("RegQueryValueExW")
	procRegSetValueExW   = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW  = advapi32.NewProc("RegDeleteValueW")
	procRegCloseKey      = advapi32.NewProc("RegCloseKey")
)

//go:embed app.ico
var embeddedFiles embed.FS

var app application

var taskbarCreatedMessage uint32

type point struct {
	X int32
	Y int32
}

type msg struct {
	HWnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type lastInputInfo struct {
	CbSize uint32
	DwTime uint32
}

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type notifyIconData struct {
	CbSize       uint32
	HWnd         uintptr
	UID          uint32
	UFlags       uint32
	UCallbackMsg uint32
	HIcon        uintptr
	SzTip        [128]uint16
	DwState      uint32
	DwStateMask  uint32
	SzInfo       [256]uint16
	UTimeout     uint32
	SzInfoTitle  [64]uint16
	DwInfoFlags  uint32
	GuidItem     guid
	HBalloonIcon uintptr
}

const (
	wndClassExSize     = unsafe.Sizeof(wndClassEx{})
	msgSize            = unsafe.Sizeof(msg{})
	notifyIconDataSize = unsafe.Sizeof(notifyIconData{})
)

var (
	_ [80 - wndClassExSize]byte
	_ [wndClassExSize - 80]byte
	_ [48 - msgSize]byte
	_ [msgSize - 48]byte
	_ [976 - notifyIconDataSize]byte
	_ [notifyIconDataSize - 976]byte
)

type application struct {
	hwnd                     uintptr
	hInstance                uintptr
	icon                     uintptr
	iconOwned                bool
	monitorOffPending        bool
	monitorOffLocked         bool
	pendingAutomatic         bool
	protectionActive         bool
	inputTickAtOff           uint32
	inputTickAtSchedule      uint32
	manualOffBlockedUntil    uint32
	autoOffMinutes           int
	language                 string
	leftClickEnabled         bool
	powerConflictNoticeShown bool
	settingsPath             string
}

func main() {
	runtime.LockOSThread()

	mutex, alreadyRunning := createSingleInstanceMutex()
	if alreadyRunning {
		return
	}
	if mutex != 0 {
		defer procCloseHandle.Call(mutex)
	}

	if err := app.run(); err != nil {
		messageBox(app.tr("startupFailed") + "\n\n" + err.Error())
	}
}

func (a *application) run() error {
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	if hInstance == 0 {
		return fmt.Errorf("GetModuleHandleW failed")
	}
	a.hInstance = hInstance
	a.icon, a.iconOwned = loadMonitorIcon()
	settings, settingsPath, settingsErr := loadAppSettings()
	a.autoOffMinutes = settings.AutoOffMinutes
	a.language = settings.Language
	a.leftClickEnabled = settings.LeftClickEnabled
	a.powerConflictNoticeShown = settings.PowerConflictNoticeShown
	a.settingsPath = settingsPath
	migrateLegacyStartup()

	className := mustUTF16Ptr(windowClass)
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   syscall.NewCallback(windowProc),
		HInstance:     hInstance,
		HIcon:         a.icon,
		HCursor:       loadArrowCursor(),
		LpszClassName: className,
		HIconSm:       a.icon,
	}

	atom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		return fmt.Errorf("RegisterClassExW failed")
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(mustUTF16Ptr(appName))),
		0,
		0, 0, 0, 0,
		0, 0, hInstance, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed")
	}
	a.hwnd = hwnd

	taskbarCreatedMessage = registerWindowMessage("TaskbarCreated")
	a.addTrayIcon()
	if settingsErr != nil {
		messageBox(a.tr("settingsLoadFailed") + "\n\n" + settingsErr.Error())
	}
	procSetTimer.Call(a.hwnd, timerAutoOffCheck, autoOffCheckMS, 0)
	if a.autoOffMinutes != 0 {
		a.showPowerConflictNoticeOnce()
	}

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			return fmt.Errorf("GetMessageW failed")
		}
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	if a.iconOwned && a.icon != 0 {
		procDestroyIcon.Call(a.icon)
	}
	return nil
}

func windowProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	if taskbarCreatedMessage != 0 && message == taskbarCreatedMessage {
		app.addTrayIcon()
		return 0
	}

	switch message {
	case wmTrayCallback:
		switch uint32(lParam) {
		case wmLButtonUp:
			if app.leftClickEnabled {
				app.scheduleMonitorOff(false)
			}
		case wmRButtonUp, wmContextMenu:
			app.showContextMenu()
		}
		return 0

	case wmTimer:
		app.onTimer(uintptr(wParam))
		return 0

	case wmDestroy:
		app.removeTrayIcon()
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return ret
}

func (a *application) scheduleMonitorOff(automatic bool) {
	// 左クリックと右クリックの消灯操作は同じロックを共有します。
	// 2秒待機中から消灯命令送信後のクールダウン終了まで、追加操作を無視します。
	if a.monitorOffLocked {
		return
	}

	if !automatic {
		// マウスでモニターを復帰させたクリックが、トレイアイコンの
		// 左クリックとして続けて届くことがあります。そのクリックでは
		// 消灯通知や新しい消灯予約を発生させません。
		if a.consumeWakeInputBeforeManualOff() || a.manualOffTemporarilyBlocked() {
			return
		}
	}

	if automatic {
		tick, ok := getLastInputTick()
		if !ok {
			return
		}
		a.inputTickAtSchedule = tick
	}

	a.stopProtection()
	a.monitorOffLocked = true
	a.monitorOffPending = true
	a.pendingAutomatic = automatic
	procSetTimer.Call(a.hwnd, timerMonitorOffDelay, monitorOffDelayMS, 0)
	a.updateTooltip()

	if !automatic {
		// モニター消灯中にWindows側へ通知音が保留され、復帰時に
		// 音だけ再生されることがあるため、この通知だけ無音にします。
		a.showSilentNotification(a.tr("notifyMonitorOffTitle"), a.tr("notifyMonitorOffBody"))
	}
}

func (a *application) onTimer(timerID uintptr) {
	switch timerID {
	case timerMonitorOffDelay:
		procKillTimer.Call(a.hwnd, timerMonitorOffDelay)
		automatic := a.pendingAutomatic
		a.monitorOffPending = false
		a.pendingAutomatic = false

		if automatic {
			current, ok := getLastInputTick()
			if !ok || current != a.inputTickAtSchedule {
				a.releaseMonitorOffLock()
				a.updateTooltip()
				return
			}
		}

		a.startProtectionAndTurnOff()
		a.beginMonitorOffCooldown()

	case timerProtection:
		a.onProtectionInterval()

	case timerWakeDetection:
		a.detectWakeInput()

	case timerAutoOffCheck:
		a.checkAutoOffTimer()

	case timerMonitorOffUnlock:
		a.releaseMonitorOffLock()
	}
}

func (a *application) beginMonitorOffCooldown() {
	procKillTimer.Call(a.hwnd, timerMonitorOffUnlock)
	procSetTimer.Call(a.hwnd, timerMonitorOffUnlock, monitorOffCooldownMS, 0)
}

func (a *application) releaseMonitorOffLock() {
	procKillTimer.Call(a.hwnd, timerMonitorOffUnlock)
	a.monitorOffLocked = false
	a.updateTooltip()
}

func (a *application) checkAutoOffTimer() {
	if a.autoOffMinutes == 0 || a.monitorOffLocked || a.protectionActive {
		return
	}

	lastInput, ok := getLastInputTick()
	if !ok {
		return
	}

	now := getTickCount()
	idleMS := uint32(now - lastInput)
	thresholdMS := uint32(a.autoOffMinutes) * 60 * 1000
	if idleMS >= thresholdMS {
		a.scheduleMonitorOff(true)
	}
}

func (a *application) startProtectionAndTurnOff() {
	tick, ok := getLastInputTick()
	if !ok {
		turnOffAllMonitors()
		a.updateTooltip()
		a.showNotification(appName, a.tr("notifyInputUnavailable"))
		return
	}

	a.inputTickAtOff = tick
	a.protectionActive = true

	procKillTimer.Call(a.hwnd, timerProtection)
	procKillTimer.Call(a.hwnd, timerWakeDetection)
	procSetTimer.Call(a.hwnd, timerProtection, protectionPeriodMS, 0)
	procSetTimer.Call(a.hwnd, timerWakeDetection, wakeDetectionMS, 0)

	a.updateTooltip()
	turnOffAllMonitors()
}

func (a *application) detectWakeInput() {
	if !a.protectionActive {
		return
	}

	current, ok := getLastInputTick()
	if ok && current != a.inputTickAtOff {
		a.handleUserWake()
	}
}

func (a *application) consumeWakeInputBeforeManualOff() bool {
	if !a.protectionActive {
		return false
	}

	current, ok := getLastInputTick()
	if !ok || current == a.inputTickAtOff {
		return false
	}

	// 復帰検出タイマーより先にトレイクリック通知が届いた場合も、
	// ここで復帰入力として処理し、そのクリックを消費します。
	a.handleUserWake()
	return true
}

func (a *application) handleUserWake() {
	// 消灯前に表示したバルーン通知がWindows側に残っている場合があるため、
	// 復帰時に明示的に消去します。
	a.clearNotification()
	a.blockManualOffAfterWake()
	a.stopProtection()
}

func (a *application) blockManualOffAfterWake() {
	a.manualOffBlockedUntil = getTickCount() + wakeInputGuardMS
}

func (a *application) manualOffTemporarilyBlocked() bool {
	until := a.manualOffBlockedUntil
	if until == 0 {
		return false
	}

	now := getTickCount()
	if int32(now-until) < 0 {
		return true
	}

	a.manualOffBlockedUntil = 0
	return false
}

func (a *application) onProtectionInterval() {
	if !a.protectionActive {
		return
	}

	current, ok := getLastInputTick()
	if !ok {
		return
	}

	if current != a.inputTickAtOff {
		a.stopProtection()
		return
	}

	// ユーザー入力がないまま30分経過した場合、Windowsや常駐ソフトなどによる
	// 勝手な再点灯に備えてモニターOFF命令を再送します。
	turnOffAllMonitors()
}

func (a *application) stopProtection() {
	a.protectionActive = false
	procKillTimer.Call(a.hwnd, timerProtection)
	procKillTimer.Call(a.hwnd, timerWakeDetection)
	a.updateTooltip()
}

func (a *application) showContextMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	turnOffFlags := uint32(mfString)
	if a.monitorOffLocked || a.manualOffTemporarilyBlocked() {
		turnOffFlags |= mfGrayed
	}
	appendMenu(menu, turnOffFlags, cmdTurnOff, a.tr("menuTurnOff"))
	appendMenu(menu, mfString|mfGrayed, cmdStatus, a.statusText())
	appendMenu(menu, mfSeparator, 0, "")

	autoMenu, _, _ := procCreatePopupMenu.Call()
	if autoMenu != 0 {
		a.appendAutoOffMenuItem(autoMenu, cmdAutoNone, 0, a.tr("menuNone"))
		a.appendAutoOffMenuItem(autoMenu, cmdAuto15, 15, a.tr("menu15Minutes"))
		a.appendAutoOffMenuItem(autoMenu, cmdAuto30, 30, a.tr("menu30Minutes"))
		a.appendAutoOffMenuItem(autoMenu, cmdAuto45, 45, a.tr("menu45Minutes"))
		a.appendAutoOffMenuItem(autoMenu, cmdAuto60, 60, a.tr("menu1Hour"))
		a.appendAutoOffMenuItem(autoMenu, cmdAuto120, 120, a.tr("menu2Hours"))
		appendMenu(menu, mfPopup, autoMenu, a.tr("menuAutoOff"))
	}

	leftClickFlags := uint32(mfString)
	if a.leftClickEnabled {
		leftClickFlags |= mfChecked
	}
	appendMenu(menu, leftClickFlags, cmdToggleLeftClick, a.tr("menuLeftClick"))

	languageMenu, _, _ := procCreatePopupMenu.Call()
	if languageMenu != 0 {
		japaneseFlags := uint32(mfString)
		englishFlags := uint32(mfString)
		if a.language == languageJapanese {
			japaneseFlags |= mfChecked
		} else {
			englishFlags |= mfChecked
		}
		appendMenu(languageMenu, japaneseFlags, cmdLanguageJapanese, a.tr("menuJapanese"))
		appendMenu(languageMenu, englishFlags, cmdLanguageEnglish, a.tr("menuEnglish"))
		appendMenu(menu, mfPopup, languageMenu, a.tr("menuLanguage"))
	}

	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, cmdPowerSettings, a.tr("menuPowerSettings"))
	appendMenu(menu, mfSeparator, 0, "")

	startupText := a.tr("menuStartupAdd")
	if startupEnabled() {
		startupText = a.tr("menuStartupRemove")
	}
	appendMenu(menu, mfString, cmdStartup, startupText)
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, cmdExit, a.tr("menuExit"))

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(a.hwnd)

	command, _, _ := procTrackPopupMenu.Call(
		menu,
		tpmRightButton|tpmNonotify|tpmReturnCmd,
		uintptr(pt.X), uintptr(pt.Y), 0,
		a.hwnd, 0,
	)

	switch command {
	case cmdTurnOff:
		a.scheduleMonitorOff(false)
	case cmdAutoNone:
		a.setAutoOffMinutes(0)
	case cmdAuto15:
		a.setAutoOffMinutes(15)
	case cmdAuto30:
		a.setAutoOffMinutes(30)
	case cmdAuto45:
		a.setAutoOffMinutes(45)
	case cmdAuto60:
		a.setAutoOffMinutes(60)
	case cmdAuto120:
		a.setAutoOffMinutes(120)
	case cmdToggleLeftClick:
		a.toggleLeftClick()
	case cmdLanguageJapanese:
		a.setLanguage(languageJapanese)
	case cmdLanguageEnglish:
		a.setLanguage(languageEnglish)
	case cmdPowerSettings:
		a.openPowerSettings()
	case cmdStartup:
		a.toggleStartup()
	case cmdExit:
		a.shutdown()
	}
}

func (a *application) appendAutoOffMenuItem(menu uintptr, command uintptr, minutes int, text string) {
	flags := uint32(mfString)
	if a.autoOffMinutes == minutes {
		flags |= mfChecked
	}
	appendMenu(menu, flags, command, text)
}

func (a *application) setAutoOffMinutes(minutes int) {
	if !isValidAutoOffMinutes(minutes) {
		return
	}

	previous := a.autoOffMinutes
	a.autoOffMinutes = minutes
	if err := a.saveSettings(); err != nil {
		a.autoOffMinutes = previous
		messageBox(a.tr("settingsSaveFailed") + "\n\n" + err.Error())
		return
	}

	// 「なし」へ変更した時点で、自動消灯の2秒待機中なら取り消します。
	// 左クリックなどによる手動消灯の待機は取り消しません。
	if minutes == 0 && a.monitorOffPending && a.pendingAutomatic {
		procKillTimer.Call(a.hwnd, timerMonitorOffDelay)
		a.monitorOffPending = false
		a.pendingAutomatic = false
		a.releaseMonitorOffLock()
	}

	a.updateTooltip()
	a.showNotification(a.tr("notifyAutoTitle"), fmt.Sprintf(a.tr("notifyAutoChanged"), formatMinutes(a.language, minutes)))
	if minutes != 0 {
		a.showPowerConflictNoticeOnce()
	}
}

func (a *application) toggleLeftClick() {
	previous := a.leftClickEnabled
	a.leftClickEnabled = !a.leftClickEnabled
	if err := a.saveSettings(); err != nil {
		a.leftClickEnabled = previous
		messageBox(a.tr("settingsSaveFailed") + "\n\n" + err.Error())
		return
	}

	text := a.tr("notifyLeftClickDisabled")
	if a.leftClickEnabled {
		text = a.tr("notifyLeftClickEnabled")
	}
	a.updateTooltip()
	a.showNotification(a.tr("notifyLeftClickTitle"), text)
}

func (a *application) setLanguage(language string) {
	if !isValidLanguage(language) || a.language == language {
		return
	}

	previous := a.language
	a.language = language
	if err := a.saveSettings(); err != nil {
		a.language = previous
		messageBox(a.tr("settingsSaveFailed") + "\n\n" + err.Error())
		return
	}

	a.updateTooltip()
	a.showNotification(a.tr("notifyLanguageTitle"), a.tr("notifyLanguageChanged"))
}

func (a *application) saveSettings() error {
	return saveAppSettings(a.settingsPath, appSettings{
		Language:                 a.language,
		LeftClickEnabled:         a.leftClickEnabled,
		AutoOffMinutes:           a.autoOffMinutes,
		PowerConflictNoticeShown: a.powerConflictNoticeShown,
	})
}

func (a *application) toggleStartup() {
	var err error
	if startupEnabled() {
		err = disableStartup()
		if err == nil {
			a.showNotification(a.tr("notifyStartupTitle"), a.tr("notifyStartupRemoved"))
		}
	} else {
		err = enableStartup()
		if err == nil {
			a.showNotification(a.tr("notifyStartupTitle"), a.tr("notifyStartupAdded"))
		}
	}

	if err != nil {
		messageBox(a.tr("startupChangeFailed") + "\n\n" + err.Error())
	}
}

func (a *application) shutdown() {
	procKillTimer.Call(a.hwnd, timerMonitorOffDelay)
	procKillTimer.Call(a.hwnd, timerProtection)
	procKillTimer.Call(a.hwnd, timerWakeDetection)
	procKillTimer.Call(a.hwnd, timerAutoOffCheck)
	procKillTimer.Call(a.hwnd, timerMonitorOffUnlock)
	procDestroyWindow.Call(a.hwnd)
}

func (a *application) statusText() string {
	if a.monitorOffPending {
		return a.tr("statusPending")
	}
	if a.protectionActive {
		return a.tr("statusProtection")
	}
	return fmt.Sprintf(a.tr("statusIdle"), formatMinutes(a.language, a.autoOffMinutes))
}

func (a *application) tooltipText() string {
	if a.monitorOffPending {
		return fmt.Sprintf(a.tr("tooltipPending"), appName)
	}
	if a.protectionActive {
		return fmt.Sprintf(a.tr("tooltipProtection"), appName)
	}
	return fmt.Sprintf(a.tr("tooltipIdle"), appName, formatMinutes(a.language, a.autoOffMinutes))
}

func (a *application) tr(key string) string {
	return translate(a.language, key)
}

func (a *application) addTrayIcon() {
	if a.hwnd == 0 {
		return
	}

	nid := notifyIconData{
		CbSize:       uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:         a.hwnd,
		UID:          1,
		UFlags:       nifMessage | nifIcon | nifTip,
		UCallbackMsg: wmTrayCallback,
		HIcon:        a.icon,
	}
	copyUTF16(nid.SzTip[:], a.tooltipText())
	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))
}

func (a *application) removeTrayIcon() {
	if a.hwnd == 0 {
		return
	}
	nid := notifyIconData{
		CbSize: uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:   a.hwnd,
		UID:    1,
	}
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
}

func (a *application) updateTooltip() {
	if a.hwnd == 0 {
		return
	}
	nid := notifyIconData{
		CbSize: uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:   a.hwnd,
		UID:    1,
		UFlags: nifTip,
	}
	copyUTF16(nid.SzTip[:], a.tooltipText())
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (a *application) showNotification(title, text string) {
	a.showNotificationWithFlags(title, text, niifInfo)
}

func (a *application) showSilentNotification(title, text string) {
	a.showNotificationWithFlags(title, text, niifInfo|niifNoSound)
}

func (a *application) showNotificationWithFlags(title, text string, flags uint32) {
	if a.hwnd == 0 {
		return
	}
	nid := notifyIconData{
		CbSize:      uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:        a.hwnd,
		UID:         1,
		UFlags:      nifInfo,
		DwInfoFlags: flags,
	}
	copyUTF16(nid.SzInfoTitle[:], title)
	copyUTF16(nid.SzInfo[:], text)
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (a *application) clearNotification() {
	if a.hwnd == 0 {
		return
	}

	// NIF_INFOを指定し、通知本文を空にして現在のバルーンを閉じます。
	nid := notifyIconData{
		CbSize:      uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:        a.hwnd,
		UID:         1,
		UFlags:      nifInfo,
		DwInfoFlags: niifNoSound,
	}
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func turnOffAllMonitors() {
	var result uintptr
	procSendMessageTimeoutW.Call(
		uintptr(0xFFFF),
		wmSysCommand,
		scMonitorPower,
		monitorPowerOff,
		smtoAbortIfHung|smtoErrorOnExit,
		1000,
		uintptr(unsafe.Pointer(&result)),
	)
}

func getLastInputTick() (uint32, bool) {
	info := lastInputInfo{CbSize: uint32(unsafe.Sizeof(lastInputInfo{}))}
	ret, _, _ := procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&info)))
	return info.DwTime, ret != 0
}

func appendMenu(menu uintptr, flags uint32, id uintptr, text string) {
	var textPtr uintptr
	if text != "" {
		textPtr = uintptr(unsafe.Pointer(mustUTF16Ptr(text)))
	}
	procAppendMenuW.Call(menu, uintptr(flags), id, textPtr)
}

func loadMonitorIcon() (uintptr, bool) {
	data, err := embeddedFiles.ReadFile("app.ico")
	if err == nil {
		if icon := createIconFromICO(data, 32); icon != 0 {
			return icon, true
		}
	}

	icon, _, _ := procLoadIconW.Call(0, 32512) // IDI_APPLICATION
	return icon, false
}

func createIconFromICO(data []byte, target int) uintptr {
	if len(data) < 6 || binary.LittleEndian.Uint16(data[2:4]) != 1 {
		return 0
	}

	count := int(binary.LittleEndian.Uint16(data[4:6]))
	bestScore := int(^uint(0) >> 1)
	bestOffset := 0
	bestSize := 0

	for i := 0; i < count; i++ {
		entry := 6 + i*16
		if entry+16 > len(data) {
			break
		}

		width := int(data[entry])
		height := int(data[entry+1])
		if width == 0 {
			width = 256
		}
		if height == 0 {
			height = 256
		}

		size := int(binary.LittleEndian.Uint32(data[entry+8 : entry+12]))
		offset := int(binary.LittleEndian.Uint32(data[entry+12 : entry+16]))
		if size <= 0 || offset < 0 || offset+size > len(data) {
			continue
		}

		score := abs(width-target) + abs(height-target)
		if score < bestScore {
			bestScore = score
			bestOffset = offset
			bestSize = size
		}
	}

	if bestSize == 0 {
		return 0
	}

	icon, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&data[bestOffset])),
		uintptr(bestSize),
		1,
		0x00030000,
		uintptr(target),
		uintptr(target),
		lrDefaultColor,
	)
	runtime.KeepAlive(data)
	return icon
}

func loadArrowCursor() uintptr {
	cursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW
	return cursor
}

func registerWindowMessage(name string) uint32 {
	ret, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(mustUTF16Ptr(name))))
	return uint32(ret)
}

func createSingleInstanceMutex() (uintptr, bool) {
	mutex, _, callErr := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(mustUTF16Ptr(mutexName))))
	if mutex == 0 {
		return 0, false
	}
	alreadyRunning := callErr == syscall.Errno(errorAlreadyExists)
	return mutex, alreadyRunning
}

func getTickCount() uint32 {
	ret, _, _ := procGetTickCount.Call()
	return uint32(ret)
}

func isValidAutoOffMinutes(minutes int) bool {
	switch minutes {
	case 0, 15, 30, 45, 60, 120:
		return true
	default:
		return false
	}
}

func (a *application) showPowerConflictNoticeOnce() {
	if a.powerConflictNoticeShown {
		return
	}

	a.showNotification(a.tr("powerConflictTitle"), a.tr("powerConflictBody"))
	a.powerConflictNoticeShown = true
	if err := a.saveSettings(); err != nil {
		a.powerConflictNoticeShown = false
		messageBox(a.tr("settingsSaveFailed") + "\n\n" + err.Error())
	}
}

func readSettingsDWORDAt(keyPath, valueName string) (uint32, bool) {
	var key uintptr
	ret, _, _ := procRegOpenKeyExW.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(mustUTF16Ptr(keyPath))),
		0,
		keyQueryValue,
		uintptr(unsafe.Pointer(&key)),
	)
	if ret != 0 {
		return 0, false
	}
	defer procRegCloseKey.Call(key)

	var valueType uint32
	var value uint32
	size := uint32(unsafe.Sizeof(value))
	ret, _, _ = procRegQueryValueExW.Call(
		key,
		uintptr(unsafe.Pointer(mustUTF16Ptr(valueName))),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		uintptr(unsafe.Pointer(&value)),
		uintptr(unsafe.Pointer(&size)),
	)
	return value, ret == 0 && valueType == regDWORD && size == 4
}

func (a *application) openPowerSettings() {
	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(mustUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(mustUTF16Ptr("ms-settings:powersleep"))),
		0,
		0,
		1,
	)
	if ret <= 32 {
		messageBox(a.tr("openPowerSettingsFailed"))
	}
}

func startupEnabled() bool {
	return startupValueExists(runValueName) || startupValueExists(legacyRunValueName)
}

func startupValueExists(valueName string) bool {
	var key uintptr
	ret, _, _ := procRegOpenKeyExW.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(mustUTF16Ptr(runKeyPath))),
		0,
		keyQueryValue,
		uintptr(unsafe.Pointer(&key)),
	)
	if ret != 0 {
		return false
	}
	defer procRegCloseKey.Call(key)

	var valueType uint32
	var size uint32
	ret, _, _ = procRegQueryValueExW.Call(
		key,
		uintptr(unsafe.Pointer(mustUTF16Ptr(valueName))),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	return ret == 0 && valueType == regSZ && size > 2
}

func migrateLegacyStartup() {
	if !startupValueExists(runValueName) && startupValueExists(legacyRunValueName) {
		_ = enableStartup()
	}
}

func enableStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	var key uintptr
	var disposition uint32
	ret, _, _ := procRegCreateKeyExW.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(mustUTF16Ptr(runKeyPath))),
		0,
		0,
		regOptionNonVolatile,
		keySetValue,
		0,
		uintptr(unsafe.Pointer(&key)),
		uintptr(unsafe.Pointer(&disposition)),
	)
	if ret != 0 {
		return fmt.Errorf("RegCreateKeyExW error: %d", ret)
	}
	defer procRegCloseKey.Call(key)

	value := utf16.Encode([]rune(`"` + exePath + `"` + "\x00"))
	ret, _, _ = procRegSetValueExW.Call(
		key,
		uintptr(unsafe.Pointer(mustUTF16Ptr(runValueName))),
		0,
		regSZ,
		uintptr(unsafe.Pointer(&value[0])),
		uintptr(len(value)*2),
	)
	if ret != 0 {
		return fmt.Errorf("RegSetValueExW error: %d", ret)
	}

	// 旧名のスタートアップ項目があれば、新しい項目へ移行します。
	_ = deleteStartupValue(legacyRunValueName)
	return nil
}

func disableStartup() error {
	if err := deleteStartupValue(runValueName); err != nil {
		return err
	}
	return deleteStartupValue(legacyRunValueName)
}

func deleteStartupValue(valueName string) error {
	var key uintptr
	ret, _, _ := procRegOpenKeyExW.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(mustUTF16Ptr(runKeyPath))),
		0,
		keySetValue,
		uintptr(unsafe.Pointer(&key)),
	)
	if ret != 0 {
		return nil
	}
	defer procRegCloseKey.Call(key)

	ret, _, _ = procRegDeleteValueW.Call(key, uintptr(unsafe.Pointer(mustUTF16Ptr(valueName))))
	if ret != 0 && ret != 2 { // ERROR_FILE_NOT_FOUND
		return fmt.Errorf("RegDeleteValueW error: %d", ret)
	}
	return nil
}

func messageBox(text string) {
	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(mustUTF16Ptr(text))),
		uintptr(unsafe.Pointer(mustUTF16Ptr(appName))),
		0x00000010, // MB_ICONERROR
	)
}

func mustUTF16Ptr(s string) *uint16 {
	ptr, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return ptr
}

func copyUTF16(dst []uint16, text string) {
	encoded := utf16.Encode([]rune(text))
	if len(encoded) >= len(dst) {
		encoded = encoded[:len(dst)-1]
	}
	copy(dst, encoded)
	dst[len(encoded)] = 0
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
