#include "../include/webview.h"
#include "../include/window.h"
#include "platform/platform_impl.h"
#include <stdexcept>
#include <functional>

WebView::WebView(Window* parent, int x, int y, int width, int height)
    : parent_(parent), nativeHandle_(nullptr), messageCallback_(nullptr) {
    if (!parent || !parent->getNativeHandle()) {
        throw std::runtime_error("Parent window must be created before creating web view");
    }
    nativeHandle_ = platform::createWebView(parent->getNativeHandle(), x, y, width, height);
    if (!nativeHandle_) {
        throw std::runtime_error("Failed to create web view");
    }
}

WebView::~WebView() {
    if (nativeHandle_) {
        platform::destroyWebView(nativeHandle_);
    }
}

WebView::WebView(WebView&& other) noexcept
    : parent_(other.parent_), nativeHandle_(other.nativeHandle_), 
      createWindowCallback_(std::move(other.createWindowCallback_)),
      messageCallback_(std::move(other.messageCallback_)) {
    other.nativeHandle_ = nullptr;
    other.createWindowCallback_ = nullptr;
    other.messageCallback_ = nullptr;
}

WebView& WebView::operator=(WebView&& other) noexcept {
    if (this != &other) {
        if (nativeHandle_) {
            platform::destroyWebView(nativeHandle_);
        }
        
        parent_ = other.parent_;
        nativeHandle_ = other.nativeHandle_;
        createWindowCallback_ = std::move(other.createWindowCallback_);
        messageCallback_ = std::move(other.messageCallback_);
        
        other.nativeHandle_ = nullptr;
        other.createWindowCallback_ = nullptr;
        other.messageCallback_ = nullptr;
    }
    return *this;
}

void WebView::loadHTMLFile(const std::string& filePath) {
    if (!nativeHandle_) {
        throw std::runtime_error("Web view is not initialized");
    }
    platform::loadHTMLFile(nativeHandle_, filePath);
}

void WebView::loadHTMLString(const std::string& html) {
    if (!nativeHandle_) {
        throw std::runtime_error("Web view is not initialized");
    }
    platform::loadHTMLString(nativeHandle_, html);
}

void WebView::loadURL(const std::string& url) {
    if (!nativeHandle_) {
        throw std::runtime_error("Web view is not initialized");
    }
    platform::loadURL(nativeHandle_, url);
}

void WebView::setCreateWindowCallback(std::function<void(const std::string& title)> callback) {
    if (!nativeHandle_) {
        return;
    }
    createWindowCallback_ = callback;
    // Pass this WebView instance as userData so we can access the callback
    platform::setWebViewCreateWindowCallback(nativeHandle_, 
        createWindowCallbackWrapper, this);
}

void WebView::setMessageCallback(std::function<void(const std::string& jsonMessage)> callback) {
    if (!nativeHandle_) {
        return;
    }
    messageCallback_ = callback;
    // Pass this WebView instance as userData so we can access the callback
    platform::setWebViewMessageCallback(nativeHandle_, 
        messageCallbackWrapper, this);
}

void WebView::postMessageToJavaScript(const std::string& jsonMessage) {
    if (!nativeHandle_) {
        return;
    }
    platform::postMessageToJavaScript(nativeHandle_, jsonMessage);
}

void WebView::createWindowCallbackWrapper(const std::string& title, void* userData) {
    WebView* webview = static_cast<WebView*>(userData);
    if (webview && webview->createWindowCallback_) {
        webview->createWindowCallback_(title);
    }
}

void WebView::messageCallbackWrapper(const std::string& jsonMessage, void* userData) {
    WebView* webview = static_cast<WebView*>(userData);
    if (webview && webview->messageCallback_) {
        webview->messageCallback_(jsonMessage);
    }
}
