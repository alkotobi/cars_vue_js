#include "../include/inputfield.h"
#include "../include/window.h"
#include "platform/platform_impl.h"
#include <stdexcept>

InputField::InputField(Window* parent, int x, int y, int width, int height, const std::string& placeholder)
    : parent_(parent), nativeHandle_(nullptr) {
    if (!parent || !parent->getNativeHandle()) {
        throw std::runtime_error("Parent window must be created before creating input field");
    }
    nativeHandle_ = platform::createInputField(parent->getNativeHandle(), x, y, width, height, placeholder);
    if (!nativeHandle_) {
        throw std::runtime_error("Failed to create input field");
    }
}

InputField::~InputField() {
    if (nativeHandle_) {
        platform::destroyInputField(nativeHandle_);
    }
}

InputField::InputField(InputField&& other) noexcept
    : parent_(other.parent_), nativeHandle_(other.nativeHandle_) {
    other.nativeHandle_ = nullptr;
}

InputField& InputField::operator=(InputField&& other) noexcept {
    if (this != &other) {
        if (nativeHandle_) {
            platform::destroyInputField(nativeHandle_);
        }
        
        parent_ = other.parent_;
        nativeHandle_ = other.nativeHandle_;
        
        other.nativeHandle_ = nullptr;
    }
    return *this;
}

void InputField::setText(const std::string& text) {
    if (nativeHandle_) {
        platform::setInputText(nativeHandle_, text);
    }
}

std::string InputField::getText() const {
    if (nativeHandle_) {
        return platform::getInputText(nativeHandle_);
    }
    return "";
}

void InputField::setPlaceholder(const std::string& placeholder) {
    // Platform-specific implementation would go here
    // For now, this is a placeholder
}
