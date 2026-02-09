// Windows platform common definitions
#ifndef WINDOWS_COMMON_H
#define WINDOWS_COMMON_H

#include <windows.h>
#include <string>
#include <map>

#ifdef PLATFORM_WINDOWS

namespace platform {

struct WindowData {
    HWND hwnd;
    bool visible;
    std::string title;
    bool isMainWindow;  // Track if this is the main application window
    void* userData;     // Store Window* pointer for cleanup
};

// Forward declarations
extern std::map<HWND, WindowData*> g_windowMap;
extern WindowData* g_mainWindow;

} // namespace platform

#endif // PLATFORM_WINDOWS

#endif // WINDOWS_COMMON_H
