#include "../include/eventhandler.h"
#include "../include/window.h"
#include "../include/webview.h"
#include "../include/messagerouter.h"
#include "../include/handlers/createwindowhandler.h"
#include <iostream>

EventHandler::EventHandler(Window* window, WebView* webView)
    : window_(window), webView_(webView) {
    if (!window_ || !webView_) {
        throw std::runtime_error("EventHandler requires valid window and webView");
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

void EventHandler::onWebViewCreateWindow(std::function<void(const std::string& title)> callback) {
    if (messageRouter_ && callback) {
        // Register CreateWindowHandler with the router
        auto handler = createCreateWindowHandler(callback);
        messageRouter_->registerHandler(handler);
    }
}
