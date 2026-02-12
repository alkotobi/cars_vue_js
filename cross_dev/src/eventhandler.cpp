#include "../include/eventhandler.h"
#include "../include/window.h"
#include "../include/webview.h"
#include "../include/messagerouter.h"
#include "../include/config_manager.h"
#include "../include/handlers/createwindowhandler.h"
#include "platform/platform_impl.h"
#include <iostream>

EventHandler::EventHandler(Window* window, WebView* webView)
    : window_(window), webView_(webView) {
    if (!window_ || !webView_) {
        throw std::runtime_error("EventHandler requires valid window and webView");
    }
    
    // Set custom preload script if configured (must be before message callback)
    std::string preload = ConfigManager::getPreloadScriptContent();
    if (!preload.empty() && webView_->getNativeHandle()) {
        platform::setWebViewPreloadScript(webView_->getNativeHandle(), preload);
    }
    
    // Create message router
    messageRouter_ = std::make_unique<MessageRouter>(webView_);
    
    // Set up message callback to route all messages through MessageRouter
    webView_->setMessageCallback([this](const std::string& jsonMessage) {
        if (messageRouter_) {
            messageRouter_->routeMessage(jsonMessage);
        }
    });
}

EventHandler::~EventHandler() {
    // MessageRouter will be automatically destroyed
}

void EventHandler::onWebViewCreateWindow(std::function<void(const std::string& title, WebViewContentType contentType, const std::string& content)> callback) {
    createWindowCallback_ = std::move(callback);
    if (messageRouter_ && createWindowCallback_) {
        auto handler = createCreateWindowHandler(createWindowCallback_);
        messageRouter_->registerHandler(handler);
    }
}

void EventHandler::attachWebView(WebView* webView) {
    if (!webView || !createWindowCallback_) {
        return;
    }
    auto router = std::make_unique<MessageRouter>(webView);
    router->registerHandler(createCreateWindowHandler(createWindowCallback_));
    MessageRouter* routerPtr = router.get();
    attachedRouters_.push_back(std::move(router));
    webView->setMessageCallback([routerPtr](const std::string& jsonMessage) {
        if (routerPtr) {
            routerPtr->routeMessage(jsonMessage);
        }
    });
}
