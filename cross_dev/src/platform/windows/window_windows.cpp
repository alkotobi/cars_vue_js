// Windows window implementation
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include "windows_common.h"
#include <string>

// Forward declarations
namespace platform {
    HINSTANCE getInstance();
    const wchar_t* getClassName();
    void initApplication();
    LRESULT CALLBACK MainWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
}

// Helper function to convert std::string to std::wstring
static std::wstring stringToWString(const std::string& str) {
    if (str.empty()) return std::wstring();
    int size_needed = MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), NULL, 0);
    std::wstring wstrTo(size_needed, 0);
    MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), &wstrTo[0], size_needed);
    return wstrTo;
}

#ifdef PLATFORM_WINDOWS

namespace platform {

// Global map to track all windows (exported for use in button_windows.cpp)
std::map<HWND, WindowData*> g_windowMap;
WindowData* g_mainWindow = nullptr;  // Track the main window

void* createWindow(int x, int y, int width, int height, const std::string& title, void* userData) {
    initApplication();
    
    WindowData* data = new WindowData;
    data->title = title;
    data->visible = false;
    data->isMainWindow = (g_mainWindow == nullptr);  // First window is the main window
    data->userData = userData;  // Store Window* pointer for cleanup
    data->resizeCallback = nullptr;
    data->resizeUserData = nullptr;
    data->closeCallback = nullptr;
    data->closeUserData = nullptr;
    data->beingDestroyed = false;
    
    std::wstring wtitle = stringToWString(title);
    HWND hwnd = CreateWindowExW(
        0,
        getClassName(),
        wtitle.c_str(),
        WS_OVERLAPPEDWINDOW,
        x, y, width, height,
        nullptr,
        nullptr,
        getInstance(),
        nullptr
    );
    
    if (!hwnd) {
        delete data;
        return nullptr;
    }
    
    data->hwnd = hwnd;
    g_windowMap[hwnd] = data;
    
    if (data->isMainWindow) {
        g_mainWindow = data;
    }
    
    return data;
}

void destroyWindow(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        if (data->hwnd) {
            g_windowMap.erase(data->hwnd);
            if (g_mainWindow == data) {
                g_mainWindow = nullptr;
            }
            if (!data->beingDestroyed) {
                DestroyWindow(data->hwnd);
            }
        }
        delete data;
    }
}

void showWindow(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        ShowWindow(data->hwnd, SW_SHOW);
        UpdateWindow(data->hwnd);
        data->visible = true;
    }
}

void hideWindow(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        ShowWindow(data->hwnd, SW_HIDE);
        data->visible = false;
    }
}

void setWindowTitle(void* handle, const std::string& title) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        data->title = title;
        std::wstring wtitle = stringToWString(title);
        SetWindowTextW(data->hwnd, wtitle.c_str());
    }
}

bool isWindowVisible(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        return data->visible && IsWindowVisible(data->hwnd);
    }
    return false;
}

void setWindowResizeCallback(void* windowHandle, void (*callback)(int width, int height, void* userData), void* userData) {
    if (windowHandle && callback) {
        WindowData* data = static_cast<WindowData*>(windowHandle);
        data->resizeCallback = callback;
        data->resizeUserData = userData;
    }
}

void setWindowCloseCallback(void* windowHandle, void (*callback)(void* userData), void* userData) {
    if (windowHandle) {
        WindowData* data = static_cast<WindowData*>(windowHandle);
        data->closeCallback = callback;
        data->closeUserData = userData;
    }
}

} // namespace platform

#endif // PLATFORM_WINDOWS
