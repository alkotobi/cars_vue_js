#ifndef CREATEWINDOWHANDLER_H
#define CREATEWINDOWHANDLER_H

#include "../messagehandler.h"
#include "../webview_window.h"
#include <memory>
#include <functional>
#include <string>

// Factory function to create CreateWindowHandler
// Callback receives (title, contentType, content).
// contentType: WebViewContentType (Url, Html, File, or Default).
// content: URL string, HTML code, or file path depending on type.
std::shared_ptr<MessageHandler> createCreateWindowHandler(
    std::function<void(const std::string& title, WebViewContentType contentType, const std::string& content)> onCreateWindow);

#endif // CREATEWINDOWHANDLER_H
