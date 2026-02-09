#ifndef WINDOW_H
#define WINDOW_H

#include <string>
#include <memory>

// Platform-agnostic Window interface
class Window {
public:
    Window(int x, int y, int width, int height, const std::string& title);
    ~Window();
    
    // Non-copyable, movable
    Window(const Window&) = delete;
    Window& operator=(const Window&) = delete;
    Window(Window&&) noexcept;
    Window& operator=(Window&&) noexcept;
    
    void show();
    void hide();
    void setTitle(const std::string& title);
    bool isVisible() const;
    
    // Platform-specific handle (opaque pointer) - exposed for components
    void* getNativeHandle() const { return nativeHandle_; }
    
private:
    void* nativeHandle_;
    int x_, y_, width_, height_;
    std::string title_;
    bool visible_;
    
    // Platform-specific implementation
    void createNativeWindow();
    void destroyNativeWindow();
    void showNativeWindow();
    void hideNativeWindow();
    void setNativeTitle(const std::string& title);
};

#endif // WINDOW_H
