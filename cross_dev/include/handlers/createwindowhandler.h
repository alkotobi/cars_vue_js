#ifndef CREATEWINDOWHANDLER_H
#define CREATEWINDOWHANDLER_H

#include "../messagehandler.h"
#include <memory>
#include <functional>
#include <string>

// Factory function to create CreateWindowHandler
std::shared_ptr<MessageHandler> createCreateWindowHandler(
    std::function<void(const std::string& title)> onCreateWindow);

#endif // CREATEWINDOWHANDLER_H
