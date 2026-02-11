#ifndef PLATFORM_IMPL_H
#define PLATFORM_IMPL_H

#include <string>

// Platform-specific implementation interface
// Each platform implements these functions

namespace platform {
    // Window management
    void* createWindow(int x, int y, int width, int height, const std::string& title, void* userData = nullptr);
    void destroyWindow(void* handle);
    void showWindow(void* handle);
    void hideWindow(void* handle);
    void setWindowTitle(void* handle, const std::string& title);
    bool isWindowVisible(void* handle);
    
    // Web view management
    // parentHandle can be a Window or Container native handle
    void* createWebView(void* parentHandle, int x, int y, int width, int height);
    void destroyWebView(void* webViewHandle);
    void resizeWebView(void* webViewHandle, int width, int height);
    void loadHTMLFile(void* webViewHandle, const std::string& filePath);
    void loadHTMLString(void* webViewHandle, const std::string& html);
    void loadURL(void* webViewHandle, const std::string& url);
    void setWebViewCreateWindowCallback(void* webViewHandle, void (*callback)(const std::string& title, void* userData), void* userData);
    void setWebViewMessageCallback(void* webViewHandle, void (*callback)(const std::string& jsonMessage, void* userData), void* userData);
    void postMessageToJavaScript(void* webViewHandle, const std::string& jsonMessage);
    
    // Window resize callback management
    void setWindowResizeCallback(void* windowHandle, void (*callback)(int width, int height, void* userData), void* userData);
    
    // Button management
    // parentHandle can be a Window or Container native handle
    void* createButton(void* parentHandle, int x, int y, int width, int height, const std::string& label, void* userData);
    void destroyButton(void* buttonHandle);
    void setButtonCallback(void* buttonHandle, void (*callback)(void*));
    
    // File dialog management
    bool showOpenFileDialog(void* windowHandle, const std::string& title, const std::string& filter, std::string& selectedPath);
    
    // Input field management
    // parentHandle can be a Window or Container native handle
    void* createInputField(void* parentHandle, int x, int y, int width, int height, const std::string& placeholder);
    void destroyInputField(void* inputHandle);
    void setInputText(void* inputHandle, const std::string& text);
    std::string getInputText(void* inputHandle);
    
    // Container functions
    void* createContainer(void* parentHandle, int x, int y, int width, int height);
    void destroyContainer(void* containerHandle);
    void resizeContainer(void* containerHandle, int x, int y, int width, int height);
    void showContainer(void* containerHandle);
    void hideContainer(void* containerHandle);
    void setContainerBackgroundColor(void* containerHandle, int red, int green, int blue);
    void setContainerBorderStyle(void* containerHandle, int borderStyle);
    
    // Application lifecycle
    void initApplication();
    void runApplication();
    void quitApplication();
}

#endif // PLATFORM_IMPL_H
