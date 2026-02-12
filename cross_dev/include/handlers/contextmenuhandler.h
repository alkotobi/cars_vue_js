#ifndef CONTEXTMENUHANDLER_H
#define CONTEXTMENUHANDLER_H

#include "../messagehandler.h"
#include <memory>

class WebViewWindow;
class MessageRouter;

std::shared_ptr<MessageHandler> createContextMenuHandler(WebViewWindow* window, MessageRouter* router);

#endif // CONTEXTMENUHANDLER_H
