#include "../include/webview_window.h"
#include "platform/platform_impl.h"
#include <stdexcept>
#include <iostream>

WebViewWindow::WebViewWindow(int x, int y, int width, int height, const std::string& title)
    : window_(nullptr), webView_(nullptr) {
    // Create the window (resizable by default in platform implementations)
    // Window(owner, parent, x, y, width, height, title)
    // For top-level window, owner and parent are nullptr
    window_ = std::make_unique<Window>(nullptr, nullptr, x, y, width, height, title);
    
    if (!window_ || !window_->getNativeHandle()) {
        throw std::runtime_error("Failed to create window for WebViewWindow");
    }
    
    // Create WebView that fills the entire window
    // WebView(owner, parent, x, y, width, height)
    // Owner is window_ (for automatic cleanup), Parent is window_ (for visual hierarchy)
    webView_ = std::make_unique<WebView>(window_.get(), window_.get(), 0, 0, width, height);
    
    if (!webView_ || !webView_->getNativeHandle()) {
        throw std::runtime_error("Failed to create WebView for WebViewWindow");
    }
    
    // Demonstrate Owner/Parent pattern:
    // - window_ owns webView_ (automatic cleanup when window_ is destroyed)
    // - window_ is parent of webView_ (visual hierarchy - WebView is displayed in Window)
    
    // Register resize callback to keep WebView sized to window
    registerResizeCallback();
}

WebViewWindow::~WebViewWindow() {
    // Unique pointers will automatically clean up
}

WebViewWindow::WebViewWindow(WebViewWindow&& other) noexcept
    : window_(std::move(other.window_)),
      webView_(std::move(other.webView_)) {
}

WebViewWindow& WebViewWindow::operator=(WebViewWindow&& other) noexcept {
    if (this != &other) {
        window_ = std::move(other.window_);
        webView_ = std::move(other.webView_);
    }
    return *this;
}

void WebViewWindow::show() {
    if (window_) {
        window_->show();
    }
}

void WebViewWindow::hide() {
    if (window_) {
        window_->hide();
    }
}

void WebViewWindow::setTitle(const std::string& title) {
    if (window_) {
        window_->setTitle(title);
    }
}

bool WebViewWindow::isVisible() const {
    if (window_) {
        return window_->isVisible();
    }
    return false;
}

void WebViewWindow::loadHTMLFile(const std::string& filePath) {
    if (webView_) {
        webView_->loadHTMLFile(filePath);
    }
}

void WebViewWindow::loadHTMLString(const std::string& html) {
    if (webView_) {
        webView_->loadHTMLString(html);
    }
}

void WebViewWindow::loadURL(const std::string& url) {
    if (webView_) {
        webView_->loadURL(url);
    }
}

void WebViewWindow::setCreateWindowCallback(std::function<void(const std::string& title)> callback) {
    if (webView_) {
        webView_->setCreateWindowCallback(callback);
    }
}

void WebViewWindow::setMessageCallback(std::function<void(const std::string& jsonMessage)> callback) {
    if (webView_) {
        webView_->setMessageCallback(callback);
    }
}

void WebViewWindow::postMessageToJavaScript(const std::string& jsonMessage) {
    if (webView_) {
        webView_->postMessageToJavaScript(jsonMessage);
    }
}

void WebViewWindow::onWindowResize(int newWidth, int newHeight) {
    if (webView_ && window_) {
        // Update WebView bounds using Control's SetBounds method
        // This demonstrates the component system - bounds are managed by Control
        webView_->SetBounds(0, 0, newWidth, newHeight);
        // The OnBoundsChanged() will call updateNativeWebViewBounds() which uses platform::resizeWebView
    }
}

void WebViewWindow::registerResizeCallback() {
    if (window_ && window_->getNativeHandle()) {
        // Register resize callback with platform layer
        platform::setWindowResizeCallback(
            window_->getNativeHandle(),
            [](int width, int height, void* userData) {
                WebViewWindow* self = static_cast<WebViewWindow*>(userData);
                if (self) {
                    self->onWindowResize(width, height);
                }
            },
            this
        );
    }
}
