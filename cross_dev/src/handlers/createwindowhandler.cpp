#include "../../include/handlers/createwindowhandler.h"
#include "../../include/messagehandler.h"
#include "../../include/window.h"
#include <nlohmann/json.hpp>
#include <iostream>
#include <vector>
#include <functional>

// Handler for creating new windows from JavaScript
class CreateWindowHandler : public MessageHandler {
public:
    CreateWindowHandler(std::function<void(const std::string& title)> onCreateWindow)
        : onCreateWindow_(onCreateWindow) {}
    
    bool canHandle(const std::string& messageType) const override {
        return messageType == "createWindow";
    }
    
    nlohmann::json handle(const nlohmann::json& payload, const std::string& requestId) override {
        // Extract title from payload
        std::string title = "New Window";
        if (payload.contains("title") && payload["title"].is_string()) {
            title = payload["title"].get<std::string>();
        }
        
        // Call the callback to create the window
        if (onCreateWindow_) {
            try {
                onCreateWindow_(title);
                
                // Return success response
                nlohmann::json result;
                result["success"] = true;
                result["title"] = title;
                return result;
            } catch (const std::exception& e) {
                // Return error response
                nlohmann::json result;
                result["success"] = false;
                result["error"] = e.what();
                return result;
            }
        }
        
        // No callback registered
        nlohmann::json result;
        result["success"] = false;
        result["error"] = "No create window callback registered";
        return result;
    }
    
    std::vector<std::string> getSupportedTypes() const override {
        return {"createWindow"};
    }
    
private:
    std::function<void(const std::string& title)> onCreateWindow_;
};

// Factory function to create handler
std::shared_ptr<MessageHandler> createCreateWindowHandler(
    std::function<void(const std::string& title)> onCreateWindow) {
    return std::make_shared<CreateWindowHandler>(onCreateWindow);
}
