#include "../include/button.h"
#include "../include/window.h"
#include "platform/platform_impl.h"
#include <stdexcept>
#include <functional>

Button::Button(Window* parent, int x, int y, int width, int height, const std::string& label)
    : parent_(parent), nativeHandle_(nullptr) {
    if (!parent || !parent->getNativeHandle()) {
        throw std::runtime_error("Parent window must be created before creating button");
    }
    nativeHandle_ = platform::createButton(parent->getNativeHandle(), x, y, width, height, label, this);
    if (!nativeHandle_) {
        throw std::runtime_error("Failed to create button");
    }
    // Set callback wrapper
    platform::setButtonCallback(nativeHandle_, callbackWrapper);
}

Button::~Button() {
    if (nativeHandle_) {
        platform::destroyButton(nativeHandle_);
    }
}

Button::Button(Button&& other) noexcept
    : parent_(other.parent_), nativeHandle_(other.nativeHandle_), callback_(std::move(other.callback_)) {
    other.nativeHandle_ = nullptr;
    other.callback_ = nullptr;
}

Button& Button::operator=(Button&& other) noexcept {
    if (this != &other) {
        if (nativeHandle_) {
            platform::destroyButton(nativeHandle_);
        }
        
        parent_ = other.parent_;
        nativeHandle_ = other.nativeHandle_;
        callback_ = std::move(other.callback_);
        
        other.nativeHandle_ = nullptr;
        other.callback_ = nullptr;
    }
    return *this;
}

void Button::setCallback(std::function<void(Window*)> callback) {
    callback_ = callback;
}

void Button::setLabel(const std::string& label) {
    // Platform-specific implementation would go here
    // For now, this is a placeholder
}

std::string Button::getLabel() const {
    // Platform-specific implementation would go here
    // For now, this is a placeholder
    return "";
}

void Button::callbackWrapper(void* userData) {
    Button* button = static_cast<Button*>(userData);
    if (button && button->callback_ && button->parent_) {
        button->callback_(button->parent_);
    }
}
