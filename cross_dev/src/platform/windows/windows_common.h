// Windows platform common definitions
#ifndef WINDOWS_COMMON_H
#define WINDOWS_COMMON_H

#include <windows.h>
#include <string>
#include <map>

#ifdef PLATFORM_WINDOWS

namespace platform {

typedef void (*ResizeCallback)(int width, int height, void* userData);

typedef void (*CloseCallback)(void* userData);

struct WindowData {
    HWND hwnd;
    bool visible;
    std::string title;
    bool isMainWindow;  // Track if this is the main application window
    void* userData;     // Store Window* pointer for cleanup
    ResizeCallback resizeCallback;  // Callback for window resize
    void* resizeUserData;  // User data for resize callback
    CloseCallback closeCallback;    // Callback when user closes window (X button)
    void* closeUserData;            // WebViewWindow* to delete when closed
    bool beingDestroyed;            // True when in WM_DESTROY - skip DestroyWindow in destroyWindow
};

// Forward declarations
extern std::map<HWND, WindowData*> g_windowMap;
extern WindowData* g_mainWindow;

} // namespace platform

#endif // PLATFORM_WINDOWS

#endif // WINDOWS_COMMON_H
