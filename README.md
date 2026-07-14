<div align="center">
  <img src="assets/icon.png" alt="MonitorOffTools icon" width="128">

# MonitorOffTools

Windows 11向けの軽量なタスクトレイ常駐モニターオフツールです。  
A lightweight system-tray monitor-off utility for Windows 11.

</div>

## 日本語

### 概要

MonitorOffToolsは、接続台数を問わず、すべてのモニターをタスクトレイからまとめて省電力オフにするWindows 11向けツールです。

.NETランタイムや外部ライブラリを必要としない、Windows x64向け単体EXEとして動作します。

### 主な機能

- タスクトレイに常駐
- トレイアイコンの左クリックから2秒後にすべてのモニターをオフ
- 左クリック操作を設定で無効化可能
- 日本語 / Englishを再起動なしで切り替え可能
- 左クリック連打と右クリックメニューからの重複実行を防止
- マウスで復帰させたクリックがトレイ操作として誤認されるのを防止
- 消灯前の通知は無音で表示し、復帰時に通知音だけ鳴る現象を防止
- キーボード入力またはマウス操作で復帰
- アプリがモニターをオフにした後、意図しない再点灯に備えて30分ごとにOFF命令を再送
- キーボード・マウス入力による復帰を検出すると、その回の再点灯防止監視を停止
- 無操作時の自動モニターオフ時間を選択可能
  - なし（初期値）
  - 15分
  - 30分
  - 45分
  - 1時間
  - 2時間
- 設定をEXEと同じフォルダーのINIファイルへ保存
- スタートアップ登録・解除
- 二重起動防止
- Windowsの画面オフ設定を右クリックメニューから直接表示

### ダウンロード

GitHubの **Releases** から最新版の `MonitorOffTools-vX.X.X.zip` をダウンロードしてください。

ZIPを展開し、`MonitorOffTools.exe` を書き込み可能な任意のフォルダーへ置いて実行します。インストールは不要です。

> [!NOTE]
> コード署名を行っていない個人配布アプリのため、初回起動時にWindows SmartScreenの警告が表示される場合があります。

### 操作

#### 左クリック

初期設定では、2秒後にすべてのモニターをオフにし、再点灯防止監視を開始します。

右クリックメニューの「左クリックでモニターをオフ」のチェックを外すと、左クリック操作を無効化できます。無効時も、右クリックメニューの「モニターをオフ」は利用できます。

最初の操作を受け付けてから、2秒の待機とモニターOFF命令送信後のクールダウンが終わるまで、追加の消灯操作は無視されます。連打によって複数の消灯処理が予約されることはありません。

マウスクリックでモニターを復帰させた場合、そのクリックがトレイアイコンの左クリックとして続けて届くことがあります。MonitorOffToolsは復帰直後のクリックを消費し、通知や新しい消灯予約が発生しないようにします。

#### 右クリック

- モニターをオフ
- 現在の状態を表示
- 自動モニターオフ時間を変更
- 左クリック操作の有効 / 無効
- 日本語 / Englishの切り替え
- Windowsの画面オフ設定を開く
- スタートアップに登録 / 登録解除
- 終了

### 言語切り替え

右クリックメニューの「言語 / Language」から、日本語またはEnglishを選択できます。

メニュー、ツールチップ、通知、エラーメッセージへ即時反映され、次回起動後も選択した言語を引き継ぎます。初期値は日本語です。

### 設定ファイル

初回起動時に、EXEと同じフォルダーへ次のファイルを自動生成します。

```text
MonitorOffTools.ini
```

内容は次の形式です。

```ini
[Settings]
Language=ja
LeftClickEnabled=true
AutoOffMinutes=0
PowerConflictNoticeShown=false
```

- `Language`: `ja` または `en`
- `LeftClickEnabled`: `true` または `false`
- `AutoOffMinutes`: `0` / `15` / `30` / `45` / `60` / `120`
- `PowerConflictNoticeShown`: Windows省電力設定の注意を表示済みかどうか

INIファイルを直接編集する場合は、MonitorOffToolsを終了してから行ってください。

> [!IMPORTANT]
> INIファイルをEXEと同じ場所へ保存するため、`C:\Program Files` など通常ユーザーが書き込めないフォルダーは避けてください。設定を保存できない場合はエラーを表示します。

v1.0.0から更新した場合、INIファイルがまだ存在しない初回起動時に、従来のレジストリから自動モニターオフ時間と省電力注意表示済み状態を移行します。

### 自動モニターオフ

無操作状態が選択した時間まで続くと、本アプリがモニターをオフにします。

初期値は「なし」です。「なし」を選択している間は、自動モニターオフだけが無効になります。手動消灯と、消灯後の再点灯防止監視は引き続き利用できます。

### Windows標準の省電力機能について

本アプリの自動モニターオフは、Windows標準の「画面をオフにする」タイマーとは別に動作します。両方が有効な場合、設定時間が短い側が先に実行されます。

本アプリの自動モニターオフを使用する場合は、Windows側の画面オフ時間を次のどちらかにすることを推奨します。

- 「なし」
- 本アプリの設定時間より長い時間

```text
設定 → システム → 電源とバッテリー → 画面とスリープ
```

PC本体がスリープすると本アプリも停止するため、モニターだけをオフにしたい場合は、Windowsのスリープ時間も本アプリの設定時間より長くしてください。

### 動画再生中の動作

自動モニターオフは、キーボード・マウスの最終入力時刻を基準に判定します。ブラウザ、動画プレイヤー、TVTestなどで動画を再生中でも、指定時間のあいだ入力がなければモニターがオフになる場合があります。

動画視聴中に自動消灯させたくない場合は、自動モニターオフを「なし」にしてください。

### スタートアップ登録

スタートアップ登録だけは、WindowsのRunレジストリを使用します。アプリの一般設定はINIファイルへ保存されます。

```text
HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run
```

管理者権限は不要です。EXEを移動する場合は、移動前にスタートアップ登録を解除し、移動後に再登録してください。

### 技術上の注意・既知の制限

- モニターをWindowsから論理的に切断するツールではありません。
- Windowsの `SC_MONITORPOWER` を使用して、省電力オフ命令を送信します。
- すべてのモニターが実際に消灯中かどうかを、GPUやモニターに依存せず確実に取得できる共通APIはないため、ユーザー入力がない間は30分ごとにOFF命令を再送します。
- 一部のモニター、GPUドライバー、USB入力機器、リモートデスクトップ環境では挙動が異なる場合があります。
- Windows 11 x64のみを対象としています。

### ビルド

Go 1.23以降をインストールし、`build-win-x64.bat` を実行します。

```text
MonitorOffTools.exe
```

リポジトリにはWindows用のアイコン・マニフェスト・バージョン情報を含むリソースファイルも収録しています。

---

## English

### Overview

MonitorOffTools is a lightweight Windows 11 utility that stays in the system tray and turns off all connected monitors at once, regardless of the number of displays.

It is distributed as a standalone Windows x64 executable and does not require the .NET runtime or external libraries.

### Features

- Runs in the system tray
- Left-click the tray icon to turn off all monitors after two seconds
- Optional left-click action to reduce accidental activation
- Switch between Japanese and English without restarting
- Prevents repeated clicks and duplicate monitor-off actions
- Prevents a mouse click used to wake the displays from being misinterpreted as a new tray action
- Shows the pre-off notification silently to prevent a delayed notification sound after wake
- Wake the monitors with keyboard or mouse input
- Re-sends the monitor-off command every 30 minutes after the app turns the monitors off, helping counter unintended wake-ups
- Stops the current wake-prevention monitoring session when keyboard or mouse input is detected
- Optional idle auto-off timer: None, 15, 30, 45, 60, or 120 minutes
- Stores app preferences in an INI file beside the executable
- Add to or remove from Windows startup
- Prevents multiple instances
- Opens Windows display power settings directly from the tray menu

### Download

Download the latest `MonitorOffTools-vX.X.X.zip` from **GitHub Releases**.

Extract the ZIP, place `MonitorOffTools.exe` in a writable folder, and run it. No installation is required.

> [!NOTE]
> Because this personal project is not code-signed, Windows SmartScreen may show a warning on first launch.

### Usage

#### Left click

By default, left-clicking the tray icon turns off all monitors after two seconds and starts wake-prevention monitoring.

Clear **Turn off monitors with left click** in the tray menu to disable this action. The **Turn off monitors** command remains available from the right-click menu.

Repeated clicks cannot queue multiple monitor-off actions. A mouse click used to wake the displays is consumed during the wake guard period, so it does not create another notification or monitor-off action.

#### Right click

- Turn off monitors
- Show current status
- Change the idle auto-off time
- Enable or disable the left-click action
- Switch between Japanese and English
- Open Windows display power settings
- Add to or remove from startup
- Exit

### Language

Choose Japanese or English from **Language / 言語** in the tray menu.

The menu, tooltip, notifications, and error messages update immediately. The selected language is saved for the next launch. Japanese is the default.

### Settings file

On first launch, the app creates this file beside the executable:

```text
MonitorOffTools.ini
```

```ini
[Settings]
Language=ja
LeftClickEnabled=true
AutoOffMinutes=0
PowerConflictNoticeShown=false
```

Close MonitorOffTools before editing the file manually.

> [!IMPORTANT]
> The settings file is stored beside the executable. Do not place the app in a normally protected location such as `C:\Program Files`. The app shows an error if it cannot save the file.

When upgrading from v1.0.0, the first v1.1.0 launch migrates the previous idle auto-off value and power-conflict notice state from the registry if no INI file exists yet.

### Idle auto-off

When there has been no keyboard or mouse input for the selected duration, the app turns off the monitors.

The default is **None**. This disables only automatic idle-based monitor-off behavior; manual operation and post-off wake-prevention monitoring remain available.

### Windows power settings

The app's idle auto-off timer works independently from the built-in Windows display-off timer. If both are enabled, whichever timer expires first will turn off the displays.

When using the app's idle auto-off feature, set the Windows display-off timer to either **Never** or a longer duration than the app's timer.

### Video playback

Idle detection is based only on keyboard and mouse input. A browser, media player, or TV application may continue playing video while the app still considers the system idle.

Set the auto-off option to **None** when you do not want video playback interrupted by automatic monitor-off behavior.

### Startup

Startup registration uses the current user's Windows Run registry key. General app preferences are stored in the INI file.

```text
HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run
```

No administrator privileges are required. Remove the startup entry before moving the executable, then add it again from the new location.

### Known limitations

- This tool does not logically disconnect displays from Windows.
- It sends the Windows `SC_MONITORPOWER` power-off command.
- There is no universal API that reliably reports the power state of every display across all GPUs and monitors, so the command is re-sent every 30 minutes while no user input is detected.
- Behavior may differ with some displays, GPU drivers, USB input devices, or Remote Desktop sessions.
- Windows 11 x64 only.

### Build

Install Go 1.23 or later and run `build-win-x64.bat`.
