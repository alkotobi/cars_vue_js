#include "../include/webview_window.h"
#include "../include/application.h"
#include "../include/config_manager.h"
#include "platform/platform_impl.h"
#include <stdexcept>
#include <iostream>

namespace {
    const char* DEFAULT_HTML = R"(<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>CrossDev</title></head>
<body style="font-family:sans-serif;padding:2em;margin:0;">
<h1>CrossDev</h1>
<p>Welcome. Configure content via options.json or pass HTML, URL, or file path to WebViewWindow.</p>
</body></html>)";
}

// Global: the first WebViewWindow created with owner=nullptr
static WebViewWindow* g_mainWebViewWindow = nullptr;

WebViewWindow* WebViewWindow::GetMainWebViewWindow() {
    return g_mainWebViewWindow;
}

WebViewWindow::WebViewWindow(Component* owner, int x, int y, int width, int height, const std::string& title,
                             WebViewContentType type, const std::string& content)
    : Component(owner), window_(nullptr), webView_(nullptr) {
    // First one with null owner becomes the main window
    if (!owner && !g_mainWebViewWindow) {
        g_mainWebViewWindow = this;
    }
    // Create the window (resizable by default in platform implementations)
    window_ = std::make_unique<Window>(nullptr, nullptr, x, y, width, height, title);
    
    if (!window_ || !window_->getNativeHandle()) {
        throw std::runtime_error("Failed to create window for WebViewWindow");
    }
    
    // Create WebView that fills the entire window
    webView_ = std::make_unique<WebView>(window_.get(), window_.get(), 0, 0, width, height);
    
    if (!webView_ || !webView_->getNativeHandle()) {
        throw std::runtime_error("Failed to create WebView for WebViewWindow");
    }
    
    // Load initial content based on type
    if (type == WebViewContentType::Html && !content.empty()) {
        webView_->loadHTMLString(content);
    } else if (type == WebViewContentType::Url && !content.empty()) {
        webView_->loadURL(content);
    } else if (type == WebViewContentType::File && !content.empty()) {
        std::string htmlContent = ConfigManager::tryLoadFileContent(content);
        if (!htmlContent.empty()) {
            webView_->loadHTMLString(htmlContent);
        } else {
            webView_->loadHTMLFile(content);  // Try direct path as fallback
        }
    } else {
        webView_->loadHTMLString(DEFAULT_HTML);
    }
    
    registerResizeCallback();
    if (owner) {
        registerCloseCallback();  // Child: delete self when user closes window
    } else {
        registerMainWindowCloseCallback();  // Main: quit app when user closes window
    }
    resetDefaultNameCache();
    debugLogLifecycleCreation(this, GetOwner(), nullptr);
}

WebViewWindow::~WebViewWindow() {
#ifndef NDEBUG
    const char* trigger = "?";
    if (g_mainWebViewWindow == this) {
        trigger = "main window (Application::quit -> main exit)";
    } else if (!GetOwner()) {
        trigger = "close callback (user clicked X)";
    } else {
        trigger = "owner destructor (Component cascade)";
    }
    std::cout << "[WebViewWindow] DESTROYED this=" << (void*)this
              << " owner=" << (void*)GetOwner()
              << " | triggered by: " << trigger << std::endl;
#endif
    if (g_mainWebViewWindow == this) {
        g_mainWebViewWindow = nullptr;
    }
    // Unique pointers and Component will clean up owned components
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

void WebViewWindow::registerCloseCallback() {
    if (window_ && window_->getNativeHandle() && GetOwner()) {
        platform::setWindowCloseCallback(
            window_->getNativeHandle(),
            [](void* userData) {
                WebViewWindow* self = static_cast<WebViewWindow*>(userData);
                if (self) {
                    self->SetOwner(nullptr);
                    delete self;
                }
            },
            this
        );
    }
}

void WebViewWindow::closeAllOwnedWebViewWindows() {
    // Collect child WebViewWindows first (we'll modify the list by deleting)
    std::vector<WebViewWindow*> children;
    for (int i = 0; i < GetComponentCount(); ++i) {
        Component* comp = GetComponent(i);
        if (WebViewWindow* w = dynamic_cast<WebViewWindow*>(comp)) {
            children.push_back(w);
        }
    }
    for (WebViewWindow* w : children) {
        if (w->getWindow() && w->getWindow()->getNativeHandle()) {
            // Clear close callback so destroyWindow won't queue a redundant delete
            platform::setWindowCloseCallback(w->getWindow()->getNativeHandle(), nullptr, nullptr);
        }
        w->SetOwner(nullptr);
        delete w;
    }
}

void WebViewWindow::registerMainWindowCloseCallback() {
    if (window_ && window_->getNativeHandle() && !GetOwner()) {
        platform::setWindowCloseCallback(
            window_->getNativeHandle(),
            [](void* userData) {
                WebViewWindow* mainWin = static_cast<WebViewWindow*>(userData);
                if (mainWin) {
                    // Destroy children while run loop is active (faster WebKit teardown)
                    mainWin->closeAllOwnedWebViewWindows();
                }
                Application::getInstance().quit();
            },
            this
        );
    }
}
