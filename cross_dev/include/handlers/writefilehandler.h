#ifndef WRITEFILEHANDLER_H
#define WRITEFILEHANDLER_H

#include "../messagehandler.h"
#include <memory>

std::shared_ptr<MessageHandler> createWriteFileHandler();

#endif // WRITEFILEHANDLER_H
