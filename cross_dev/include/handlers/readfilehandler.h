#ifndef READFILEHANDLER_H
#define READFILEHANDLER_H

#include "../messagehandler.h"
#include <memory>

std::shared_ptr<MessageHandler> createReadFileHandler();

#endif // READFILEHANDLER_H
