#ifndef CALCULATORHANDLER_H
#define CALCULATORHANDLER_H

#include "../messagehandler.h"
#include <memory>

// Factory function to create CalculatorHandler
std::shared_ptr<MessageHandler> createCalculatorHandler();

#endif // CALCULATORHANDLER_H
