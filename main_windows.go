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
	protectionPeriodMS   = 30 * 60 * 1_000
	wakeDetectionMS      = 250
	autoOffCheckMS       = 1_000

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nifInfo    = 0x00000010
	niifInfo   = 0x00000001

	mfString    = 0x00000000
	mfGrayed    = 0x00000001
	mfChecked   = 0x00000008
	mfPopup     = 0x00000010
	mfSeparator = 0x00000800

	tpmRightButton = 0x0002
	tpmNonotify    = 0x0080
	tpmReturnCmd   = 0x0100

	cmdTurnOff       = 1001
	cmdStatus        = 1002
	cmdStartup       = 1003
	cmdExit          = 1004
	cmdPowerSettings = 1005

	cmdAutoNone = 1100
	cmdAuto15   = 1101
	cmdAuto30   = 1102
	cmdAuto45   = 1103
	cmdAuto60   = 1104
	cmdAuto120  = 1105

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
	hwnd                uintptr
	hInstance           uintptr
	icon                uintptr
	iconOwned           bool
	monitorOffPending   bool
	monitorOffLocked    bool
	pendingAutomatic    bool
	protectionActive    bool
	inputTickAtOff      uint32
	inputTickAtSchedule uint32
	autoOffMinutes      int
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
		messageBox("起動に失敗しました。\n\n" + err.Error())
	}
}

func (a *application) run() error {
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	if hInstance == 0 {
		return fmt.Errorf("GetModuleHandleW failed")
	}
	a.hInstance = hInstance
	a.icon, a.iconOwned = loadMonitorIcon()
	a.autoOffMinutes = loadAutoOffMinutes()
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
			app.scheduleMonitorOff(false)
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
		a.showNotification("モニターオフ", "すべてのモニターをオフにします。")
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
		a.showNotification(appName, "入力状態を取得できないため、再点灯防止監視を開始できませんでした。")
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
		a.stopProtection()
	}
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
	if a.monitorOffLocked {
		turnOffFlags |= mfGrayed
	}
	appendMenu(menu, turnOffFlags, cmdTurnOff, "モニターをオフ")
	appendMenu(menu, mfString|mfGrayed, cmdStatus, a.statusText())
	appendMenu(menu, mfSeparator, 0, "")

	autoMenu, _, _ := procCreatePopupMenu.Call()
	if autoMenu != 0 {
		a.appendAutoOffMenuItem(autoMenu, cmdAutoNone, 0, "なし")
		a.appendAutoOffMenuItem(autoMenu, cmdAuto15, 15, "15分")
		a.appendAutoOffMenuItem(autoMenu, cmdAuto30, 30, "30分")
		a.appendAutoOffMenuItem(autoMenu, cmdAuto45, 45, "45分")
		a.appendAutoOffMenuItem(autoMenu, cmdAuto60, 60, "1時間")
		a.appendAutoOffMenuItem(autoMenu, cmdAuto120, 120, "2時間")
		appendMenu(menu, mfPopup, autoMenu, "自動モニターオフ")
	}

	appendMenu(menu, mfString, cmdPowerSettings, "Windowsの画面オフ設定を開く")
	appendMenu(menu, mfSeparator, 0, "")

	startupText := "スタートアップに登録"
	if startupEnabled() {
		startupText = "スタートアップ登録を解除"
	}
	appendMenu(menu, mfString, cmdStartup, startupText)
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, cmdExit, "終了")

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
	case cmdPowerSettings:
		openPowerSettings()
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

	a.autoOffMinutes = minutes

	// 「なし」へ変更した時点で、自動消灯の2秒待機中なら取り消します。
	// 左クリックなどによる手動消灯の待機は取り消しません。
	if minutes == 0 && a.monitorOffPending && a.pendingAutomatic {
		procKillTimer.Call(a.hwnd, timerMonitorOffDelay)
		a.monitorOffPending = false
		a.pendingAutomatic = false
		a.releaseMonitorOffLock()
	}

	if err := saveAutoOffMinutes(minutes); err != nil {
		messageBox("自動モニターオフ設定の保存に失敗しました。\n\n" + err.Error())
		return
	}

	a.updateTooltip()
	a.showNotification("自動モニターオフ", formatMinutes(minutes)+"に変更しました。")
	if minutes != 0 {
		a.showPowerConflictNoticeOnce()
	}
}

func (a *application) toggleStartup() {
	var err error
	if startupEnabled() {
		err = disableStartup()
		if err == nil {
			a.showNotification("スタートアップ", "スタートアップ登録を解除しました。")
		}
	} else {
		err = enableStartup()
		if err == nil {
			a.showNotification("スタートアップ", "Windowsログオン時に起動するよう登録しました。")
		}
	}

	if err != nil {
		messageBox("スタートアップ設定の変更に失敗しました。\n\n" + err.Error())
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
		return "状態: 2秒後にモニターオフ"
	}
	if a.protectionActive {
		return "状態: 再点灯防止監視中（30分固定）"
	}
	return "状態: 待機中 / 自動オフ " + formatMinutes(a.autoOffMinutes)
}

func (a *application) tooltipText() string {
	if a.monitorOffPending {
		return appName + " - 2秒後に消灯"
	}
	if a.protectionActive {
		return appName + " - 再点灯防止監視中"
	}
	return appName + " - 自動オフ " + formatMinutes(a.autoOffMinutes)
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
	if a.hwnd == 0 {
		return
	}
	nid := notifyIconData{
		CbSize:      uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:        a.hwnd,
		UID:         1,
		UFlags:      nifInfo,
		DwInfoFlags: niifInfo,
	}
	copyUTF16(nid.SzInfoTitle[:], title)
	copyUTF16(nid.SzInfo[:], text)
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

func formatMinutes(minutes int) string {
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

func loadAutoOffMinutes() int {
	value, ok := readSettingsDWORD(autoOffValueName)
	if ok && isValidAutoOffMinutes(int(value)) {
		return int(value)
	}
	return defaultAutoOffMinutes
}

func saveAutoOffMinutes(minutes int) error {
	return writeSettingsDWORD(autoOffValueName, uint32(minutes))
}

func (a *application) showPowerConflictNoticeOnce() {
	if value, ok := readSettingsDWORD(conflictNoticeValueName); ok && value != 0 {
		return
	}

	a.showNotification(
		"Windows省電力との競合防止",
		"本アプリの自動オフを使う場合、Windows標準の画面オフは「なし」または本アプリより長い時間を推奨します。",
	)
	_ = writeSettingsDWORD(conflictNoticeValueName, 1)
}

func readSettingsDWORD(valueName string) (uint32, bool) {
	if value, ok := readSettingsDWORDAt(settingsKeyPath, valueName); ok {
		return value, true
	}

	// v1.0.0の旧設定を引き継ぎます。
	value, ok := readSettingsDWORDAt(legacySettingsKeyPath, valueName)
	if ok {
		_ = writeSettingsDWORD(valueName, value)
	}
	return value, ok
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

func writeSettingsDWORD(valueName string, value uint32) error {
	var key uintptr
	var disposition uint32
	ret, _, _ := procRegCreateKeyExW.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(mustUTF16Ptr(settingsKeyPath))),
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

	ret, _, _ = procRegSetValueExW.Call(
		key,
		uintptr(unsafe.Pointer(mustUTF16Ptr(valueName))),
		0,
		regDWORD,
		uintptr(unsafe.Pointer(&value)),
		4,
	)
	if ret != 0 {
		return fmt.Errorf("RegSetValueExW error: %d", ret)
	}
	return nil
}

func openPowerSettings() {
	ret, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(mustUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(mustUTF16Ptr("ms-settings:powersleep"))),
		0,
		0,
		1,
	)
	if ret <= 32 {
		messageBox("Windowsの電源設定を開けませんでした。")
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
