#include "../include/messagerouter.h"
#include "../include/webview.h"
#include "../include/messagehandler.h"
#include "platform/platform_impl.h"
#include <nlohmann/json.hpp>
#include <iostream>
#include <sstream>

MessageRouter::MessageRouter(WebView* webView) : webView_(webView) {
    if (!webView_) {
        throw std::runtime_error("MessageRouter requires a valid WebView");
    }
}

MessageRouter::~MessageRouter() {
    handlers_.clear();
}

void MessageRouter::registerHandler(const std::string& messageType, std::shared_ptr<MessageHandler> handler) {
    if (!handler) {
        return;
    }
    handlers_[messageType] = handler;
}

void MessageRouter::registerHandler(std::shared_ptr<MessageHandler> handler) {
    if (!handler) {
        return;
    }
    // Register for all types this handler supports
    for (const auto& type : handler->getSupportedTypes()) {
        handlers_[type] = handler;
    }
}

void MessageRouter::routeMessage(const std::string& jsonMessage) {
    std::cout << "=== MessageRouter::routeMessage called ===" << std::endl;
    std::cout << "  jsonMessage: " << jsonMessage << std::endl;
    
    if (jsonMessage.empty()) {
        std::cout << "  Message is empty, returning" << std::endl;
        return;
    }
    
    std::string type, payload, requestId;
    if (!parseMessage(jsonMessage, type, payload, requestId)) {
        std::cerr << "Failed to parse message: " << jsonMessage << std::endl;
        if (!requestId.empty()) {
            sendResponse(requestId, "", "Failed to parse message");
        }
        return;
    }
    
    std::cout << "  Parsed - type: " << type << ", requestId: " << requestId << std::endl;
    
    // Find handler for this message type
    auto it = handlers_.find(type);
    if (it == handlers_.end()) {
        std::cerr << "No handler registered for message type: " << type << std::endl;
        if (!requestId.empty()) {
            sendResponse(requestId, "", "Unknown message type: " + type);
        }
        return;
    }
    
    // Parse payload JSON
    nlohmann::json payloadJson;
    try {
        if (!payload.empty()) {
            payloadJson = nlohmann::json::parse(payload);
        }
    } catch (const nlohmann::json::exception& e) {
        std::cerr << "Failed to parse payload JSON: " << e.what() << std::endl;
        if (!requestId.empty()) {
            sendResponse(requestId, "", "Invalid payload JSON: " + std::string(e.what()));
        }
        return;
    }
    
    // Call handler
    std::cout << "Calling handler for type: " << type << std::endl;
    try {
        nlohmann::json result = it->second->handle(payloadJson, requestId);
        std::cout << "Handler returned result: " << result.dump() << std::endl;
        
        // Send response if requestId was provided
        if (!requestId.empty()) {
            std::string resultStr = result.is_null() ? "null" : result.dump();
            std::cout << "Sending response for requestId: " << requestId << std::endl;
            sendResponse(requestId, resultStr, "");
        } else {
            std::cout << "No requestId, skipping response" << std::endl;
        }
    } catch (const std::exception& e) {
        std::cerr << "Handler error: " << e.what() << std::endl;
        if (!requestId.empty()) {
            sendResponse(requestId, "", "Handler error: " + std::string(e.what()));
        }
    }
}

void MessageRouter::sendResponse(const std::string& requestId, const std::string& resultJson, const std::string& error) {
    std::cout << "=== MessageRouter::sendResponse called ===" << std::endl;
    std::cout << "  requestId: " << requestId << std::endl;
    std::cout << "  resultJson: " << resultJson << std::endl;
    std::cout << "  error: " << error << std::endl;
    
    if (!webView_) {
        std::cerr << "ERROR: webView_ is null!" << std::endl;
        return;
    }
    
    // Create response JSON
    nlohmann::json response;
    response["requestId"] = requestId;
    if (!error.empty()) {
        response["error"] = error;
        response["result"] = nullptr;
    } else {
        response["error"] = nullptr;
        if (!resultJson.empty() && resultJson != "null") {
            try {
                response["result"] = nlohmann::json::parse(resultJson);
            } catch (...) {
                response["result"] = resultJson; // Fallback to string
            }
        } else {
            response["result"] = nullptr;
        }
    }
    
    // Send to JavaScript via platform API
    std::string responseStr = response.dump();
    std::cout << "  Sending response: " << responseStr << std::endl;
    platform::postMessageToJavaScript(webView_->getNativeHandle(), responseStr);
    std::cout << "  Response sent!" << std::endl;
}

bool MessageRouter::parseMessage(const std::string& jsonMessage, std::string& type, 
                                 std::string& payload, std::string& requestId) {
    try {
        nlohmann::json msg = nlohmann::json::parse(jsonMessage);
        
        // Extract type (required)
        if (!msg.contains("type") || !msg["type"].is_string()) {
            return false;
        }
        type = msg["type"].get<std::string>();
        
        // Extract payload (optional)
        if (msg.contains("payload")) {
            payload = msg["payload"].dump();
        }
        
        // Extract requestId (optional)
        if (msg.contains("requestId") && msg["requestId"].is_string()) {
            requestId = msg["requestId"].get<std::string>();
        }
        
        return true;
    } catch (const nlohmann::json::exception& e) {
        std::cerr << "JSON parse error: " << e.what() << std::endl;
        return false;
    }
}
