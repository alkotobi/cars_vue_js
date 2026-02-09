// Linux web view implementation
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>
#include <fstream>
#include <sstream>
#include <cstring>

#ifdef PLATFORM_LINUX

namespace platform {

struct WindowData {
    Display* display;
    Window window;
    GtkWidget* gtkWindow;
    bool visible;
    std::string title;
};

// Callback type for window creation
typedef void (*CreateWindowCallback)(const std::string& title, void* userData);

struct WebViewData {
    WebKitWebView* webView;
    GtkWidget* container;
    CreateWindowCallback createWindowCallback;
    void* createWindowUserData;
    MessageCallback messageCallback;
    void* messageUserData;
};

void* createWebView(void* windowHandle, int x, int y, int width, int height) {
    if (!windowHandle) {
        return nullptr;
    }
    
    WindowData* windowData = static_cast<WindowData*>(windowHandle);
    
    // Create GTK window if not exists
    if (!windowData->gtkWindow) {
        windowData->gtkWindow = gtk_window_new(GTK_WINDOW_TOPLEVEL);
        gtk_window_set_default_size(GTK_WINDOW(windowData->gtkWindow), width + x, height + y);
    }
    
    WebViewData* webViewData = new WebViewData;
    webViewData->webView = WEBKIT_WEB_VIEW(webkit_web_view_new());
    webViewData->container = GTK_WIDGET(webViewData->webView);
    webViewData->createWindowCallback = nullptr;
    webViewData->createWindowUserData = nullptr;
    webViewData->messageCallback = nullptr;
    webViewData->messageUserData = nullptr;
    
    gtk_widget_set_size_request(webViewData->container, width, height);
    gtk_fixed_put(GTK_FIXED(gtk_window_get_child(GTK_WINDOW(windowData->gtkWindow))), 
                  webViewData->container, x, y);
    gtk_widget_show(webViewData->container);
    
    return webViewData;
}

void destroyWebView(void* webViewHandle) {
    if (webViewHandle) {
        WebViewData* data = static_cast<WebViewData*>(webViewHandle);
        if (data->container) {
            gtk_widget_destroy(data->container);
        }
        delete data;
    }
}

void loadHTMLFile(void* webViewHandle, const std::string& filePath) {
    if (!webViewHandle) {
        return;
    }
    
    WebViewData* data = static_cast<WebViewData*>(webViewHandle);
    
    std::ifstream file(filePath);
    if (file.is_open()) {
        std::stringstream buffer;
        buffer << file.rdbuf();
        std::string htmlContent = buffer.str();
        file.close();
        
        webkit_web_view_load_html(data->webView, htmlContent.c_str(), 
                                  ("file://" + filePath).c_str());
    }
}

void loadHTMLString(void* webViewHandle, const std::string& html) {
    if (!webViewHandle) {
        return;
    }
    
    WebViewData* data = static_cast<WebViewData*>(webViewHandle);
    webkit_web_view_load_html(data->webView, html.c_str(), nullptr);
}

void loadURL(void* webViewHandle, const std::string& url) {
    if (!webViewHandle) {
        return;
    }
    
    WebViewData* data = static_cast<WebViewData*>(webViewHandle);
    webkit_web_view_load_uri(data->webView, url.c_str());
}

// Script message handler callback for WebKitGTK
static void onScriptMessageReceived(WebKitUserContentManager* manager, WebKitJavascriptResult* result, gpointer userData) {
    WebViewData* data = static_cast<WebViewData*>(userData);
    if (!data) {
        return;
    }
    
    JSCValue* value = webkit_javascript_result_get_js_value(result);
    if (!value) {
        return;
    }
    
    // Convert to JSON string
    char* jsonStr = jsc_value_to_json_string(value, 0);
    std::string jsonMessage = jsonStr ? jsonStr : "";
    g_free(jsonStr);
    
    // Try new message callback first (for MessageRouter)
    if (data->messageCallback) {
        data->messageCallback(jsonMessage, data->messageUserData);
        g_object_unref(value);
        return;
    }
    
    // Fallback to old createWindow callback for backward compatibility
    if (data->createWindowCallback && jsc_value_is_object(value)) {
        JSCValue* typeValue = jsc_value_object_get_property(value, "type");
        if (typeValue && jsc_value_is_string(typeValue)) {
            gchar* typeStr = jsc_value_to_string(typeValue);
            if (typeStr && strcmp(typeStr, "createWindow") == 0) {
                JSCValue* titleValue = jsc_value_object_get_property(value, "title");
                if (titleValue && jsc_value_is_string(titleValue)) {
                    gchar* titleStr = jsc_value_to_string(titleValue);
                    std::string title = titleStr ? titleStr : "New Window";
                    if (titleStr) g_free(titleStr);
                    
                    data->createWindowCallback(title, data->createWindowUserData);
                }
                if (titleValue) g_object_unref(titleValue);
            }
            if (typeStr) g_free(typeStr);
        }
        if (typeValue) g_object_unref(typeValue);
    }
    
    g_object_unref(value);
}

void setWebViewCreateWindowCallback(void* webViewHandle, void (*callback)(const std::string& title, void* userData), void* userData) {
    if (!webViewHandle || !callback) {
        return;
    }
    
    WebViewData* data = static_cast<WebViewData*>(webViewHandle);
    data->createWindowCallback = callback;
    data->createWindowUserData = userData;
    
    // Get user content manager
    WebKitUserContentManager* manager = webkit_web_view_get_user_content_manager(data->webView);
    
    // Register script message handler
    webkit_user_content_manager_register_script_message_handler(manager, "createWindow");
    
    // Connect to script message received signal
    g_signal_connect(manager, "script-message-received::createWindow",
                     G_CALLBACK(onScriptMessageReceived), data);
    
    // Inject JavaScript bridge for window.chrome.webview.postMessage
    const char* script = R"(
        window.chrome = window.chrome || {};
        window.chrome.webview = window.chrome.webview || {};
        window.chrome.webview.postMessage = function(message) {
            if (typeof message === 'object' && message.type === 'createWindow') {
                window.webkit.messageHandlers.createWindow.postMessage(message);
            }
        };
    )";
    
    // Inject script that will be executed when page loads
    WebKitUserScript* userScript = webkit_user_script_new(script, 
        WEBKIT_USER_CONTENT_INJECT_TOP_FRAME, 
        WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START, 
        nullptr, nullptr);
    webkit_user_content_manager_add_script(manager, userScript);
    g_object_unref(userScript);
}

void setWebViewMessageCallback(void* webViewHandle, void (*callback)(const std::string& jsonMessage, void* userData), void* userData) {
    if (!webViewHandle || !callback) {
        return;
    }
    
    WebViewData* data = static_cast<WebViewData*>(webViewHandle);
    if (!data || !data->webView) {
        return;
    }
    
    data->messageCallback = callback;
    data->messageUserData = userData;
    
    // Register message handler if not already registered
    WebKitUserContentManager* manager = webkit_web_view_get_user_content_manager(data->webView);
    if (!g_signal_has_handler_pending(manager, g_signal_lookup("script-message-received::nativeMessage", 
                                                               WEBKIT_TYPE_USER_CONTENT_MANAGER), 0, false)) {
        webkit_user_content_manager_register_script_message_handler(manager, "nativeMessage");
        g_signal_connect(manager, "script-message-received::nativeMessage",
                         G_CALLBACK(onScriptMessageReceived), data);
        
        // Inject JavaScript bridge
        const char* script = R"(
            window.chrome = window.chrome || {};
            window.chrome.webview = window.chrome.webview || {};
            window.chrome.webview.postMessage = function(message) {
                var msg = typeof message === 'string' ? JSON.parse(message) : message;
                window.webkit.messageHandlers.nativeMessage.postMessage(msg);
            };
        )";
        
        WebKitUserScript* userScript = webkit_user_script_new(script, 
            WEBKIT_USER_CONTENT_INJECT_TOP_FRAME, 
            WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START, 
            nullptr, nullptr);
        webkit_user_content_manager_add_script(manager, userScript);
        g_object_unref(userScript);
    }
}

void postMessageToJavaScript(void* webViewHandle, const std::string& jsonMessage) {
    if (!webViewHandle) {
        return;
    }
    
    WebViewData* data = static_cast<WebViewData*>(webViewHandle);
    if (!data || !data->webView) {
        return;
    }
    
    // Execute JavaScript to post message
    std::string script = "window.postMessage(" + jsonMessage + ", '*');";
    webkit_web_view_run_javascript(data->webView, script.c_str(), nullptr, nullptr, nullptr);
}

} // namespace platform

#endif // PLATFORM_LINUX
