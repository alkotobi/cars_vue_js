// Windows button implementation
#include "../../../include/platform.h"
#include "../../../include/window.h"
#include "../platform_impl.h"
#include "windows_common.h"
#include <string>
#include <windows.h>

namespace platform {
    HINSTANCE getInstance();
    const wchar_t* getClassName();
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

struct ButtonData {
    HWND hwnd;
    void* userData;
    void (*callback)(void*);
};

// Global button map for callback handling
std::map<HWND, ButtonData*> g_buttonMap;

void* createButton(void* parentHandle, int x, int y, int width, int height, const std::string& label, void* userData) {
    if (!parentHandle) {
        return nullptr;
    }
    
    HWND parentHwnd = nullptr;
    
    // Get parent HWND - could be from WindowData or direct HWND (from Container)
    if (IsWindow((HWND)parentHandle)) {
        parentHwnd = (HWND)parentHandle;
    } else {
        // Try to get from WindowData
        WindowData* windowData = static_cast<WindowData*>(parentHandle);
        if (windowData && windowData->hwnd) {
            parentHwnd = windowData->hwnd;
        } else {
            return nullptr;
        }
    }
    
    ButtonData* buttonData = new ButtonData;
    buttonData->userData = userData;
    buttonData->callback = nullptr;
    
    std::wstring wlabel = stringToWString(label);
    buttonData->hwnd = CreateWindowExW(
        0,
        L"BUTTON",
        wlabel.c_str(),
        WS_VISIBLE | WS_CHILD | BS_PUSHBUTTON,
        x, y, width, height,
        parentHwnd,
        nullptr,
        getInstance(),
        nullptr
    );
    
    // Store button in map for callback lookup
    g_buttonMap[buttonData->hwnd] = buttonData;
    
    return buttonData;
}

void destroyButton(void* buttonHandle) {
    if (buttonHandle) {
        ButtonData* data = static_cast<ButtonData*>(buttonHandle);
        if (data->hwnd) {
            g_buttonMap.erase(data->hwnd);
            DestroyWindow(data->hwnd);
        }
        delete data;
    }
}

void setButtonCallback(void* buttonHandle, void (*callback)(void*)) {
    if (buttonHandle) {
        ButtonData* data = static_cast<ButtonData*>(buttonHandle);
        data->callback = callback;
    }
}

// Window procedure - handles button callbacks and window messages
LRESULT CALLBACK MainWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    if (uMsg == WM_SIZE) {
        // Handle window resize
        auto it = platform::g_windowMap.find(hwnd);
        if (it != platform::g_windowMap.end()) {
            WindowData* windowData = it->second;
            if (windowData->resizeCallback) {
                int width = LOWORD(lParam);
                int height = HIWORD(lParam);
                windowData->resizeCallback(width, height, windowData->resizeUserData);
            }
        }
        return 0;
    }
    if (uMsg == WM_COMMAND) {
        HWND buttonHwnd = (HWND)lParam;
        auto it = g_buttonMap.find(buttonHwnd);
        if (it != g_buttonMap.end() && it->second->callback) {
            it->second->callback(it->second->userData);
            return 0;
        }
    }
    if (uMsg == WM_DESTROY) {
        // Look up the window in the map
        auto it = platform::g_windowMap.find(hwnd);
        if (it != platform::g_windowMap.end()) {
            WindowData* windowData = it->second;
            
            // If this is not the main window and has userData (Window*), delete it
            if (windowData != platform::g_mainWindow && windowData->userData) {
                // This is a dynamically created window - delete the Window object
                Window* window = static_cast<Window*>(windowData->userData);
                delete window;
            }
            
            // Only quit the application if the main window is being destroyed
            if (windowData == platform::g_mainWindow) {
                // Main window is being destroyed - quit the application
                PostQuitMessage(0);
            }
        }
        // For other windows, just destroy them (no PostQuitMessage)
        return 0;
    }
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

} // namespace platform

#endif // PLATFORM_WINDOWS
