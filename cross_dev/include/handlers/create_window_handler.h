#ifndef CREATE_WINDOW_HANDLER_H
#define CREATE_WINDOW_HANDLER_H

#include "../message_handler.h"
#include "../webview_window.h"
#include <memory>
#include <functional>
#include <string>

// Factory function to create CreateWindowHandler
// Callback receives (name, title, contentType, content, isSingleton).
// name: auto-generated as {className}-{incremental} or className for singleton.
// title: user-facing window title.
// contentType: WebViewContentType (Url, Html, File, or Default).
// content: URL string, HTML code, or file path depending on type.
// isSingleton: if true, show existing window with same className; default false.
// Payload: className (required), title, url/html/file, isSingleton.
std::shared_ptr<MessageHandler> createCreateWindowHandler(
    std::function<void(const std::string& name, const std::string& title, WebViewContentType contentType, const std::string& content, bool isSingleton)> onCreateWindow);

#endif // CREATE_WINDOW_HANDLER_H
