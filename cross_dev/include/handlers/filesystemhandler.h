#ifndef FILESYSTEMHANDLER_H
#define FILESYSTEMHANDLER_H

#include "../messagehandler.h"
#include <memory>

std::shared_ptr<MessageHandler> createFileSystemHandler();

#endif // FILESYSTEMHANDLER_H
