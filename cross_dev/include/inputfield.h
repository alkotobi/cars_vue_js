#ifndef INPUTFIELD_H
#define INPUTFIELD_H

#include <string>

class Window;

// Platform-agnostic InputField interface
class InputField {
public:
    InputField(Window* parent, int x, int y, int width, int height, const std::string& placeholder);
    ~InputField();
    
    // Non-copyable, movable
    InputField(const InputField&) = delete;
    InputField& operator=(const InputField&) = delete;
    InputField(InputField&&) noexcept;
    InputField& operator=(InputField&&) noexcept;
    
    void setText(const std::string& text);
    std::string getText() const;
    void setPlaceholder(const std::string& placeholder);
    
    // Platform-specific handle (opaque pointer)
    void* getNativeHandle() const { return nativeHandle_; }
    
private:
    Window* parent_;
    void* nativeHandle_;
};

#endif // INPUTFIELD_H
