#ifndef WEBVIEW_WINDOW_H
#define WEBVIEW_WINDOW_H

#include "window.h"
#include "webview.h"
#include <string>
#include <functional>
#include <memory>

// A window that contains a WebView that fills the entire window
// The WebView automatically resizes when the window is resized
class WebViewWindow {
public:
    WebViewWindow(int x, int y, int width, int height, const std::string& title);
    ~WebViewWindow();
    
    // Non-copyable, movable
    WebViewWindow(const WebViewWindow&) = delete;
    WebViewWindow& operator=(const WebViewWindow&) = delete;
    WebViewWindow(WebViewWindow&&) noexcept;
    WebViewWindow& operator=(WebViewWindow&&) noexcept;
    
    // Window operations
    void show();
    void hide();
    void setTitle(const std::string& title);
    bool isVisible() const;
    
    // WebView operations
    void loadHTMLFile(const std::string& filePath);
    void loadHTMLString(const std::string& html);
    void loadURL(const std::string& url);
    
    // Message handling
    void setCreateWindowCallback(std::function<void(const std::string& title)> callback);
    void setMessageCallback(std::function<void(const std::string& jsonMessage)> callback);
    void postMessageToJavaScript(const std::string& jsonMessage);
    
    // Get underlying window and webview (for advanced usage)
    Window* getWindow() { return window_.get(); }
    WebView* getWebView() { return webView_.get(); }
    
private:
    std::unique_ptr<Window> window_;
    std::unique_ptr<WebView> webView_;
    
    // Handle window resize to update WebView size
    void onWindowResize(int newWidth, int newHeight);
    
    // Platform-specific resize callback registration
    void registerResizeCallback();
};

#endif // WEBVIEW_WINDOW_H
