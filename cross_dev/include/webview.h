#ifndef WEBVIEW_H
#define WEBVIEW_H

#include <string>
#include <functional>

class Window;

// Platform-agnostic WebView interface
class WebView {
public:
    WebView(Window* parent, int x, int y, int width, int height);
    ~WebView();
    
    // Non-copyable, movable
    WebView(const WebView&) = delete;
    WebView& operator=(const WebView&) = delete;
    WebView(WebView&&) noexcept;
    WebView& operator=(WebView&&) noexcept;
    
    void loadHTMLFile(const std::string& filePath);
    void loadHTMLString(const std::string& html);
    void loadURL(const std::string& url);
    void setCreateWindowCallback(std::function<void(const std::string& title)> callback);
    void setMessageCallback(std::function<void(const std::string& jsonMessage)> callback);
    void postMessageToJavaScript(const std::string& jsonMessage);
    
    // Platform-specific handle (opaque pointer)
    void* getNativeHandle() const { return nativeHandle_; }
    
private:
    Window* parent_;
    void* nativeHandle_;
    std::function<void(const std::string& title)> createWindowCallback_;
    std::function<void(const std::string& jsonMessage)> messageCallback_;
    
    // Static callback wrappers
    static void createWindowCallbackWrapper(const std::string& title, void* userData);
    static void messageCallbackWrapper(const std::string& jsonMessage, void* userData);
};

#endif // WEBVIEW_H
