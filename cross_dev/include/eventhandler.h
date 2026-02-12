#ifndef EVENTHANDLER_H
#define EVENTHANDLER_H

#include "webview_window.h"
#include <string>
#include <functional>
#include <map>
#include <memory>
#include <vector>

class Window;
class WebView;
class MessageRouter;

// Event handler for managing HTML/webview events
class EventHandler {
public:
    EventHandler(Window* window, WebView* webView);
    ~EventHandler();
    
    // Register callback for webview create window event
    // Callback receives (title, contentType, content). contentType is Url, Html, File, or Default.
    void onWebViewCreateWindow(std::function<void(const std::string& title, WebViewContentType contentType, const std::string& content)> callback);
    
    // Attach a child window's WebView so CrossDev.invoke (e.g. createWindow) works from it.
    // Call after creating a child WebViewWindow.
    void attachWebView(WebView* webView);
    
    // Get the message router (for registering additional handlers)
    MessageRouter* getMessageRouter() { return messageRouter_.get(); }
    
private:
    Window* window_;
    WebView* webView_;
    std::unique_ptr<MessageRouter> messageRouter_;
    std::function<void(const std::string&, WebViewContentType, const std::string&)> createWindowCallback_;
    std::vector<std::unique_ptr<MessageRouter>> attachedRouters_;
};

#endif // EVENTHANDLER_H
