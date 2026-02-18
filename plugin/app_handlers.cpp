#include "app_handlers.h"
#include "message_router.h"
#include "message_handler.h"
#include <nlohmann/json.hpp>
#include <memory>
#include <iostream>

class CarsInfoHandler : public MessageHandler {
public:
    bool canHandle(const std::string& messageType) const override {
        return messageType == "carsInfo";
    }

    nlohmann::json handle(const nlohmann::json& payload, const std::string& requestId) override {
        (void)payload;
        (void)requestId;
        nlohmann::json result;
        result["success"] = true;
        result["app"] = "Cars";
        result["message"] = "Cars CrossDev plugin loaded";
        return result;
    }

    std::vector<std::string> getSupportedTypes() const override {
        return {"carsInfo"};
    }
};

extern "C" void registerAppHandlers(MessageRouter* router) {
    if (!router) {
        return;
    }
    std::cout << "[Cars plugin] registerAppHandlers called" << std::endl;
    router->registerHandler(std::make_shared<CarsInfoHandler>());
}

