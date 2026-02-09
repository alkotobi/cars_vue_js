#include "../include/eventhandler.h"
#include "../include/window.h"
#include "../include/inputfield.h"
#include "../include/webview.h"
#include "../include/messagerouter.h"
#include "../include/handlers/createwindowhandler.h"
#include "platform/platform_impl.h"
#include <iostream>
#include <algorithm>
#include <cctype>

EventHandler::EventHandler(Window* window, InputField* inputField, WebView* webView)
    : window_(window), inputField_(inputField), webView_(webView) {
    if (!window_ || !inputField_ || !webView_) {
        throw std::runtime_error("EventHandler requires valid window, inputField, and webView");
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

void EventHandler::onLoadButtonClicked() {
    handleLoadRequest();
}

void EventHandler::onWebViewCreateWindow(std::function<void(const std::string& title)> callback) {
    if (messageRouter_ && callback) {
        // Register CreateWindowHandler with the router
        auto handler = createCreateWindowHandler(callback);
        messageRouter_->registerHandler(handler);
    }
}

void EventHandler::handleLoadRequest() {
    if (!window_ || !inputField_ || !webView_) {
        return;
    }
    
    // Get URL/Path from input field
    std::string inputText = inputField_->getText();
    
    if (!inputText.empty()) {
        try {
            // Check if it's a URL
            bool isURL = false;
            std::string urlToLoad = inputText;
            
            // Check for explicit protocol
            if (inputText.find("http://") == 0 || inputText.find("https://") == 0) {
                isURL = true;
            }
            // Check for www. prefix (common URL pattern)
            else if (inputText.find("www.") == 0) {
                isURL = true;
                urlToLoad = "http://" + inputText;
            }
            // Check for common domain patterns (contains .com, .org, .net, etc.)
            else if (inputText.find('.') != std::string::npos && 
                     (inputText.find(".com") != std::string::npos ||
                      inputText.find(".org") != std::string::npos ||
                      inputText.find(".net") != std::string::npos ||
                      inputText.find(".edu") != std::string::npos ||
                      inputText.find(".io") != std::string::npos ||
                      inputText.find(".dev") != std::string::npos) &&
                     inputText.find(' ') == std::string::npos && // No spaces (unlikely in URLs)
                     inputText.find('/') == std::string::npos) { // No path separators (could be file path)
                isURL = true;
                urlToLoad = "http://" + inputText;
            }
            
            if (isURL) {
                // Load URL directly
                webView_->loadURL(urlToLoad);
                std::cout << "Loading URL: " << urlToLoad << std::endl;
            } else {
                // Assume it's a file path
                webView_->loadHTMLFile(inputText);
                std::cout << "Loaded HTML file: " << inputText << std::endl;
            }
        } catch (const std::exception& e) {
            std::cerr << "Error loading: " << e.what() << std::endl;
        }
    } else {
        // If input is empty, use file dialog
        std::string filePath;
        bool selected = platform::showOpenFileDialog(
            window_->getNativeHandle(),
            "Open HTML File",
            "HTML Files (*.html)|*.html|All Files (*.*)|*.*",
            filePath
        );
        
        if (selected && !filePath.empty()) {
            try {
                webView_->loadHTMLFile(filePath);
                inputField_->setText(filePath);
                std::cout << "Loaded HTML file: " << filePath << std::endl;
            } catch (const std::exception& e) {
                std::cerr << "Error loading HTML file: " << e.what() << std::endl;
            }
        }
    }
}
