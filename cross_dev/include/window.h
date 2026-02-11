#ifndef WINDOW_H
#define WINDOW_H

#include "control.h"
#include <string>
#include <memory>

// Platform-agnostic Window interface
// Window inherits from Control, so it can own and parent other components
class Window : public Control {
public:
    // Constructor: Window(owner, parent, x, y, width, height, title)
    // For top-level windows, owner and parent are typically nullptr
    Window(Component* owner = nullptr, Control* parent = nullptr, 
           int x = 0, int y = 0, int width = 800, int height = 600, 
           const std::string& title = "");
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
    
protected:
    // Override Control virtual methods
    void OnParentChanged(Control* oldParent, Control* newParent) override;
    void OnBoundsChanged() override;
    void OnVisibleChanged() override;
    
private:
    void* nativeHandle_;
    std::string title_;
    
    // Platform-specific implementation
    void createNativeWindow();
    void destroyNativeWindow();
    void showNativeWindow();
    void hideNativeWindow();
    void setNativeTitle(const std::string& title);
    void updateNativeWindowBounds();
};

#endif // WINDOW_H
