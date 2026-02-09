#ifndef BUTTON_H
#define BUTTON_H

#include <string>
#include <functional>

class Window;

// Platform-agnostic Button interface
class Button {
public:
    Button(Window* parent, int x, int y, int width, int height, const std::string& label);
    ~Button();
    
    // Non-copyable, movable
    Button(const Button&) = delete;
    Button& operator=(const Button&) = delete;
    Button(Button&&) noexcept;
    Button& operator=(Button&&) noexcept;
    
    // Use std::function to allow lambdas with captures
    void setCallback(std::function<void(Window*)> callback);
    void setLabel(const std::string& label);
    std::string getLabel() const;
    
    // Platform-specific handle (opaque pointer)
    void* getNativeHandle() const { return nativeHandle_; }
    
private:
    Window* parent_;
    void* nativeHandle_;
    std::function<void(Window*)> callback_;
    
    // Static callback wrapper
    static void callbackWrapper(void* userData);
};

#endif // BUTTON_H
