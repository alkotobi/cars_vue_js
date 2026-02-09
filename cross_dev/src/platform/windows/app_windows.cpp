// Windows application lifecycle implementation
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <windows.h>
#include <string>
#include <iostream>
#include <io.h>
#include <fcntl.h>
#include <objbase.h>

#ifdef PLATFORM_WINDOWS

namespace platform {

// Helper function to convert std::string to std::wstring
// Note: Currently unused but kept for potential future use
// static std::wstring stringToWString(const std::string& str) {
//     if (str.empty()) return std::wstring();
//     int size_needed = MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), NULL, 0);
//     std::wstring wstrTo(size_needed, 0);
//     MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), &wstrTo[0], size_needed);
//     return wstrTo;
// }

static HINSTANCE g_hInstance = nullptr;
static const wchar_t* g_className = L"NativeWindowClass";
static bool g_comInitialized = false;

LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);

void initApplication() {
    if (!g_hInstance) {
        g_hInstance = GetModuleHandle(nullptr);
        
        // Initialize COM for WebView2 (required for WebView2 to work)
        HRESULT hr = CoInitializeEx(nullptr, COINIT_APARTMENTTHREADED);
        if (SUCCEEDED(hr)) {
            g_comInitialized = true;
            std::wcout << L"COM initialized successfully" << std::endl;
        } else if (hr == RPC_E_CHANGED_MODE) {
            // COM was already initialized with a different mode - this is OK
            std::wcout << L"COM already initialized with different mode" << std::endl;
        } else {
            std::wcout << L"COM initialization warning: 0x" << std::hex << hr << std::dec << std::endl;
        }
        
        // Allocate console for debug output (optional - can be removed for release)
        #ifdef _DEBUG
        AllocConsole();
        FILE* pCout;
        FILE* pCerr;
        FILE* pCin;
        freopen_s(&pCout, "CONOUT$", "w", stdout);
        freopen_s(&pCerr, "CONOUT$", "w", stderr);
        freopen_s(&pCin, "CONIN$", "r", stdin);
        std::cout.clear();
        std::cerr.clear();
        std::cin.clear();
        #endif
        
        WNDCLASSW wc = {};
        wc.lpfnWndProc = WindowProc;
        wc.hInstance = g_hInstance;
        wc.lpszClassName = g_className;
        wc.hbrBackground = (HBRUSH)(COLOR_WINDOW + 1);
        wc.hCursor = LoadCursor(nullptr, IDC_ARROW);
        
        RegisterClassW(&wc);
    }
}

void runApplication() {
    MSG msg = {};
    while (GetMessage(&msg, nullptr, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}

void quitApplication() {
    // Uninitialize COM if we initialized it
    if (g_comInitialized) {
        CoUninitialize();
        g_comInitialized = false;
    }
    PostQuitMessage(0);
}

HINSTANCE getInstance() {
    return g_hInstance;
}

const wchar_t* getClassName() {
    return g_className;
}

// Forward declaration - implemented in button_windows.cpp
LRESULT CALLBACK MainWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);

LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    return MainWindowProc(hwnd, uMsg, wParam, lParam);
}

} // namespace platform

#endif // PLATFORM_WINDOWS
