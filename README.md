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
- トレイアイコンを左クリックすると、2秒後にすべてのモニターをオフ
- 左クリック連打と右クリックメニューからの重複実行を防止
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
- スタートアップ登録・解除
- 二重起動防止
- Windowsの画面オフ設定を右クリックメニューから直接表示

### ダウンロード

GitHubの **Releases** から最新版の `MonitorOffTools-vX.X.X.zip` をダウンロードしてください。

ZIPを展開し、`MonitorOffTools.exe` を任意のフォルダーへ置いて実行します。インストールは不要です。

> [!NOTE]
> コード署名を行っていない個人配布アプリのため、初回起動時にWindows SmartScreenの警告が表示される場合があります。

### 操作

#### 左クリック

2秒後にすべてのモニターをオフにし、再点灯防止監視を開始します。

最初の操作を受け付けてから、2秒の待機とモニターOFF命令送信後のクールダウンが終わるまで、追加の左クリックや右クリックからの消灯操作は無視されます。連打によって複数の消灯処理が予約されることはありません。

#### 右クリック

- モニターをオフ
- 現在の状態を表示
- 自動モニターオフ時間を変更
- Windowsの画面オフ設定を開く
- スタートアップに登録 / 登録解除
- 終了

### 自動モニターオフ

無操作状態が選択した時間まで続くと、本アプリがモニターをオフにします。

初期値は「なし」です。「なし」を選択している間は、自動モニターオフだけが無効になります。左クリックによる手動消灯と、消灯後の再点灯防止監視は引き続き利用できます。

設定は次回起動後も引き継がれます。

```text
HKEY_CURRENT_USER\Software\MonitorOffTools
```

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

現在のユーザーの次のレジストリへ登録します。管理者権限は不要です。

```text
HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run
```

EXEを移動する場合は、移動前にスタートアップ登録を解除し、移動後に再登録してください。

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
- Prevents repeated left-clicks and duplicate monitor-off actions from the tray menu
- Wake the monitors with keyboard or mouse input
- Re-sends the monitor-off command every 30 minutes after the app turns the monitors off, helping counter unintended wake-ups
- Stops the current wake-prevention monitoring session when keyboard or mouse input is detected
- Optional idle auto-off timer
  - None (default)
  - 15 minutes
  - 30 minutes
  - 45 minutes
  - 1 hour
  - 2 hours
- Add to or remove from Windows startup
- Prevents multiple instances
- Opens the Windows display power settings directly from the tray menu

### Download

Download the latest `MonitorOffTools-vX.X.X.zip` from **GitHub Releases**.

Extract the ZIP, place `MonitorOffTools.exe` in any folder, and run it. No installation is required.

> [!NOTE]
> Because this personal project is not code-signed, Windows SmartScreen may show a warning on first launch.

### Usage

#### Left click

Turns off all monitors after two seconds and starts the wake-prevention monitoring session.

After the first action is accepted, additional left-clicks and tray-menu monitor-off commands are ignored until the two-second delay and the post-command cooldown have completed. Repeated clicks cannot queue multiple monitor-off actions.

#### Right click

- Turn off monitors
- Show current status
- Change idle auto-off time
- Open Windows display power settings
- Add to / remove from startup
- Exit

### Idle auto-off

When there has been no keyboard or mouse input for the selected duration, the app turns off the monitors.

The default is **None**. This disables only automatic idle-based monitor-off behavior; manual tray-icon operation and the post-off wake-prevention monitoring remain available.

The selected value is stored under:

```text
HKEY_CURRENT_USER\Software\MonitorOffTools
```

### Windows power settings

The app's idle auto-off timer works independently from the built-in Windows display-off timer. If both are enabled, whichever timer expires first will turn off the displays.

When using the app's idle auto-off feature, set the Windows display-off timer to either:

- Never
- A longer duration than the app's timer

If the PC enters sleep, this app also stops running. Set the Windows sleep timer longer than the app's auto-off timer when you want only the monitors to turn off.

### Video playback

Idle detection is based only on keyboard and mouse input. A browser, media player, or TV application may continue playing video while the app still considers the system idle.

Set the auto-off option to **None** when you do not want video playback interrupted by automatic monitor-off behavior.

### Known limitations

- This tool does not logically disconnect displays from Windows.
- It sends the Windows `SC_MONITORPOWER` power-off command.
- Windows does not provide one universal, reliable API for checking whether every monitor is physically off across all monitor and GPU combinations, so the app re-sends the command every 30 minutes while no user input is detected.
- Behavior may vary with certain monitors, GPU drivers, USB input devices, and Remote Desktop sessions.
- Windows 11 x64 only.

### Build

Install Go 1.23 or later and run:

```text
build-win-x64.bat
```

The resulting executable is:

```text
MonitorOffTools.exe
```

## License

MIT License. See [LICENSE](LICENSE).
