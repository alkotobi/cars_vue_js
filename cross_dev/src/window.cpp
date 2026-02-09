#include "../include/window.h"
#include "platform/platform_impl.h"
#include <stdexcept>

Window::Window(int x, int y, int width, int height, const std::string& title)
    : nativeHandle_(nullptr), x_(x), y_(y), width_(width), height_(height), title_(title), visible_(false) {
    createNativeWindow();
}

Window::~Window() {
    if (nativeHandle_) {
        destroyNativeWindow();
    }
}

Window::Window(Window&& other) noexcept
    : nativeHandle_(other.nativeHandle_),
      x_(other.x_), y_(other.y_),
      width_(other.width_), height_(other.height_),
      title_(std::move(other.title_)),
      visible_(other.visible_) {
    other.nativeHandle_ = nullptr;
    other.visible_ = false;
}

Window& Window::operator=(Window&& other) noexcept {
    if (this != &other) {
        if (nativeHandle_) {
            destroyNativeWindow();
        }
        
        nativeHandle_ = other.nativeHandle_;
        x_ = other.x_;
        y_ = other.y_;
        width_ = other.width_;
        height_ = other.height_;
        title_ = std::move(other.title_);
        visible_ = other.visible_;
        
        other.nativeHandle_ = nullptr;
        other.visible_ = false;
    }
    return *this;
}

void Window::createNativeWindow() {
    nativeHandle_ = platform::createWindow(x_, y_, width_, height_, title_, this);
    if (!nativeHandle_) {
        throw std::runtime_error("Failed to create native window");
    }
}

void Window::destroyNativeWindow() {
    if (nativeHandle_) {
        platform::destroyWindow(nativeHandle_);
        nativeHandle_ = nullptr;
    }
}

void Window::show() {
    if (!nativeHandle_) {
        return;
    }
    showNativeWindow();
    visible_ = true;
}

void Window::hide() {
    if (!nativeHandle_) {
        return;
    }
    hideNativeWindow();
    visible_ = false;
}

void Window::showNativeWindow() {
    platform::showWindow(nativeHandle_);
}

void Window::hideNativeWindow() {
    platform::hideWindow(nativeHandle_);
}

void Window::setTitle(const std::string& title) {
    title_ = title;
    if (nativeHandle_) {
        setNativeTitle(title);
    }
}

void Window::setNativeTitle(const std::string& title) {
    platform::setWindowTitle(nativeHandle_, title);
}

bool Window::isVisible() const {
    if (!nativeHandle_) {
        return false;
    }
    return platform::isWindowVisible(nativeHandle_);
}
