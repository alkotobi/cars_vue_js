# Windows Build Guide

## Avoid CMake Download (use local WebView2 SDK)

If CMake hangs at "Using manually downloaded nlohmann/json" or shortly after, the next step is **WebView2 FetchContent** which downloads from NuGet (can hang on slow/firewalled networks).

**Solution: use a local WebView2 SDK (no download):**

1. Download the [WebView2 NuGet package](https://www.nuget.org/packages/Microsoft.Web.WebView2/1.0.2592.51) (or run `nuget install Microsoft.Web.WebView2 -Version 1.0.2592.51`).
2. Extract the `.nupkg` (it's a zip) so that `build/native/include/WebView2.h` exists.
3. Place the extracted folder as `cross_dev/WebView2/`:
   ```
   cross_dev/
     WebView2/
       build/native/include/WebView2.h
       build/native/x64/WebView2LoaderStatic.lib
       ...
   ```
4. Re-run CMake. You should see: `WebView2: using LOCAL SDK from .../WebView2 (no download)`

Or pass a custom path: `cmake -DWEBVIEW2_SDK_PATH=C:\path\to\extracted\WebView2 ..`

## Quick Build

```cmd
rmdir /s /q build 2>nul
mkdir build
cd build
cmake -A x64 ..
cmake --build .
```

## Troubleshooting LNK2019 (Unresolved Externals)

If you see linker errors for symbols like:
- `WebViewWindow::setOnDestroyCallback`
- `EventHandler::onWebViewCreateWindow`, `attachWebView`
- `createContextMenuHandler`
- `SingletonWebViewWindowManager`

these usually mean the object files that define them are missing. Common causes:

### 1. Out-of-date sources

Ensure your `cross_dev` folder matches the latest repo:

```cmd
cd cross_dev
git status
git pull
```

### 2. Clean reconfigure

CMake cache can point at old paths or exclude files. Do a clean build:

```cmd
cd build
rmdir /s /q * 2>nul
cd ..
cmake -A x64 -B build
cmake --build build
```

### 3. Verify required files exist

These must be present:

| File | Defines |
|------|---------|
| `src/webview_window.cpp` | `WebViewWindow::setOnDestroyCallback` |
| `src/event_handler.cpp` | `EventHandler` methods |
| `src/handlers/context_menu_handler.cpp` | `createContextMenuHandler` |
| `src/singleton_webview_window_manager.cpp` | uses `setOnDestroyCallback` |

### 4. Verbose build

See which sources are compiled:

```cmd
cmake --build build --verbose
```

Confirm `webview_window.cpp`, `event_handler.cpp`, `context_menu_handler.cpp` appear in the build commands.

### 5. Build configuration

If using Visual Studio, make sure both Debug and Release are built, or specify:

```cmd
cmake --build build --config Release
```

## Double .exe (CrossDev.exe.exe)

If the output is `CrossDev.exe.exe`, do a clean reconfigure and rebuild. The project sets `RUNTIME_OUTPUT_NAME` to `CrossDev` (no `.exe`); CMake adds `.exe`. A stale cache can cause a duplicate suffix.

## Debugging Car Stock / createWindow Timeout

Logging was added to trace the createWindow flow. To view it:

1. **JavaScript (browser console)**: Open DevTools in the WebView (right-click → Inspect, or use `--enable-webview-devtools`). Look for `[CrossDev]` logs: invoke start, message sent, response received, or timeout.

2. **Native (OutputDebugString)**: Run [DebugView](https://learn.microsoft.com/en-us/sysinternals/downloads/debugview) as Administrator, or run the app from Visual Studio. Look for:
   - `WebView2MessageHandler::Invoke` – message received from JS
   - `MessageRouter::routeMessage` – message routed
   - `CreateWindowHandler` – handler invoked
   - `SingletonWebViewWindowManager` – window creation
   - `MessageRouter::sendResponse` – response sent back

3. **Console**: If the app has a console window, `std::cout` will show the same native logs.

If you see `messageCallback is NULL`, the MessageRouter is not attached. If you see the handler run but no response, check `postMessageToJavaScript` / `PostWebMessageAsJson`.

If you see `WebView2MessageHandler::Invoke CALLED` but never `Received message:` or `Processing message`, the native code was receiving postMessage objects. WebView2's `TryGetWebMessageAsString` fails for objects; the fix uses `get_WebMessageAsJson` as fallback and the built-in bridge now sends `JSON.stringify(...)` so both paths work.

### WebView2 message format (string vs object)

| Direction | API | Behavior |
|-----------|-----|----------|
| **Page → Native** | `page.postMessage(string)` | `TryGetWebMessageAsString` works |
| **Page → Native** | `page.postMessage(object)` | `TryGetWebMessageAsString` FAILS; use `get_WebMessageAsJson` |
| **Native → Page** | `PostWebMessageAsJson(jsonString)` | Page receives `event.data` as parsed **object** |

New debug logs: `[WebView2] TryGetWebMessageAsString: OK` = page sent string. `[WebView2] get_WebMessageAsJson: OK` = page sent object. Duplicate "No handler found" after a successful response is a known quirk (message delivered twice).

## MinGW (alternative)

If MSVC builds keep failing, try MinGW:

```cmd
rmdir /s /q build 2>nul
mkdir build
cd build
cmake -G "MinGW Makefiles" -DCMAKE_PREFIX_PATH="C:\msys64\mingw64" ..
cmake --build .
```

(Adjust `CMAKE_PREFIX_PATH` to your MinGW installation if needed.)
