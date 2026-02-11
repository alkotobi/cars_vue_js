#ifndef EVENTHANDLER_H
#define EVENTHANDLER_H

#include <string>
#include <functional>
#include <map>
#include <memory>

class Window;
class WebView;
class MessageRouter;

// Event handler for managing HTML/webview events
class EventHandler {
public:
    EventHandler(Window* window, WebView* webView);
    ~EventHandler();
    
    // Register callback for webview create window event
    void onWebViewCreateWindow(std::function<void(const std::string& title)> callback);
    
    // Get the message router (for registering additional handlers)
    MessageRouter* getMessageRouter() { return messageRouter_.get(); }
    
private:
    Window* window_;
    WebView* webView_;
    std::unique_ptr<MessageRouter> messageRouter_;
};

#endif // EVENTHANDLER_H
