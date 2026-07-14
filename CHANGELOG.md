# Changelog

All notable changes to MonitorOffTools are documented in this file.

## [1.1.0] - 2026-07-14

### Added

- Japanese / English UI switching without restarting the app.
- Optional left-click monitor-off action to reduce accidental activation.
- Portable `MonitorOffTools.ini` settings file beside the executable.
- Automatic migration of compatible v1.0.0 registry settings when no INI file exists.

### Changed

- App preferences are now stored in the INI file instead of the registry.
- Windows startup registration continues to use the current user's Run registry key.
- GitHub Actions no longer enables an unnecessary Go module cache for this dependency-free project.

## [1.0.0] - 2026-07-14

### Added

- System-tray operation for Windows 11 x64.
- Left-click monitor-off action with a two-second delay.
- Shared action lock that prevents repeated clicks from queuing multiple monitor-off commands.
- Support for turning off all connected monitors.
- Keyboard and mouse wake detection.
- Thirty-minute monitor-off command re-send while no user input is detected.
- Idle auto-off settings: None, 15, 30, 45, 60, and 120 minutes.
- Startup registration and removal.
- Single-instance protection.
- Shortcut to Windows display power settings.
- Japanese and English documentation.

### Fixed

- Prevented a wake-up mouse click from being treated as another tray-icon monitor-off action.
- Prevented the monitor-off notification sound from being played after the displays wake.
