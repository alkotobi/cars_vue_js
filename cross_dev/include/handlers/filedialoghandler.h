#ifndef FILEDIALOGHANDLER_H
#define FILEDIALOGHANDLER_H

#include "../messagehandler.h"
#include <memory>

class Window;

// Factory function to create FileDialogHandler
std::shared_ptr<MessageHandler> createFileDialogHandler(Window* window);

#endif // FILEDIALOGHANDLER_H
