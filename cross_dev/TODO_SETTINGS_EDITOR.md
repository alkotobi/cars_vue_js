# Settings (options.json) Editor - Implementation Plan

## Overview

Add a "Settings" menu item that opens a WebView window for viewing and editing CrossDev's `options.json`. Provide a user-friendly UI with form controls per key and/or raw JSON editing (VS Code style).

## Todo List

### 1. Native / C++ Side

- [x] **Add getOptionsPath handler** – Expose options.json path to JS via `CrossDev.invoke('getOptionsPath')`
- [x] **Add readOptions / writeOptions handlers** – Direct read/write of options.json via `OptionsHandler` (options_handler.cpp)
- [x] **Add "Settings" menu item** – In View menu (below Zoom), triggers opening the Settings window
- [x] **Wire menu → createWindow** – When Settings is clicked, executes script that calls `createWindow` with url `/settings`

### 2. Checkbox Control (Native)

- [ ] **Create Checkbox C++ class** – `include/checkbox.h`, `src/checkbox.cpp` (pattern: Button, InputField)
- [ ] **Platform API** – `createCheckbox`, `destroyCheckbox`, `getChecked`, `setChecked`, `setCallback`
- [ ] **Windows** – BS_AUTOCHECKBOX style
- [ ] **macOS** – NSButtonTypeSwitch or NSButtonTypeMomentaryPushIn with checkbox appearance
- [ ] **Linux** – GtkCheckButton
- [ ] **Add to CMakeLists** – checkbox.cpp + platform-specific sources

### 3. Vue / Frontend

- [ ] **Add /settings route** – In router
- [ ] **Create SettingsView.vue** – Main settings editor component
- [ ] **Form UI** – Per-key controls:
  - htmlLoading.method → dropdown (file/url/html)
  - htmlLoading.filePath → text input
  - htmlLoading.url → text input
  - htmlLoading.preloadPath → text input
  - Checkboxes for boolean options if any
- [ ] **Raw JSON editor** – Textarea with syntax highlighting (e.g. simple `JSON.stringify(obj, null, 2)`), or tab to switch between Form / Raw
- [ ] **Load on mount** – `CrossDev.invoke('getOptionsPath')` then `readFile` to fetch content
- [ ] **Save** – Validate JSON, then `writeFile` with path
- [ ] **Reload hint** – After save, show "Restart app to apply changes" or similar

### 4. Integration

- [ ] **Menu handler** – On Settings click, call createWindow with `url: baseUrl + '/settings'`, `title: 'Settings'`, `isSingleton: true`
- [ ] **Ensure Settings route is in known routes** – For router base path detection

## options.json Structure (Reference)

```json
{
  "htmlLoading": {
    "method": "file",
    "filePath": "demo.html",
    "url": "",
    "htmlContent": "...",
    "preloadPath": ""
  }
}
```

## Notes

- Settings edits the CrossDev config, not Cars app config
- Path is platform-specific: `%APPDATA%\CrossDev\options.json` (Windows), `~/Library/Application Support/CrossDev/options.json` (macOS)
- Checkbox control is for CrossDev native UI (e.g. layout_demo, future dialogs); Settings editor uses HTML form controls
