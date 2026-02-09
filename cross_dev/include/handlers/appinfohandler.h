#ifndef APPINFOHANDLER_H
#define APPINFOHANDLER_H

#include "../messagehandler.h"
#include <memory>

// Factory function to create AppInfoHandler
std::shared_ptr<MessageHandler> createAppInfoHandler();

#endif // APPINFOHANDLER_H
