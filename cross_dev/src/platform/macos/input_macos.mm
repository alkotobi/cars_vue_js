// macOS input field implementation
#import <Cocoa/Cocoa.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>

#ifdef PLATFORM_MACOS

namespace platform {

void* createInputField(void* windowHandle, int x, int y, int width, int height, const std::string& placeholder) {
    @autoreleasepool {
        if (!windowHandle) {
            return nullptr;
        }
        
        NSWindow *window = (__bridge NSWindow*)windowHandle;
        NSView *contentView = [window contentView];
        
        NSRect inputRect = NSMakeRect(x, y, width, height);
        NSTextField *textField = [[NSTextField alloc] initWithFrame:inputRect];
        [textField setPlaceholderString:[NSString stringWithUTF8String:placeholder.c_str()]];
        [textField setAutoresizingMask:NSViewMaxXMargin | NSViewMinYMargin];
        [textField setBezelStyle:NSTextFieldSquareBezel];
        
        [contentView addSubview:textField];
        
        return (void*)CFBridgingRetain(textField);
    }
}

void destroyInputField(void* inputHandle) {
    @autoreleasepool {
        if (inputHandle) {
            NSTextField *textField = (__bridge NSTextField*)inputHandle;
            [textField removeFromSuperview];
            CFBridgingRelease(inputHandle);
        }
    }
}

void setInputText(void* inputHandle, const std::string& text) {
    @autoreleasepool {
        if (inputHandle) {
            NSTextField *textField = (__bridge NSTextField*)inputHandle;
            [textField setStringValue:[NSString stringWithUTF8String:text.c_str()]];
        }
    }
}

std::string getInputText(void* inputHandle) {
    @autoreleasepool {
        if (inputHandle) {
            NSTextField *textField = (__bridge NSTextField*)inputHandle;
            NSString *value = [textField stringValue];
            if (value) {
                return [value UTF8String];
            }
        }
        return "";
    }
}

} // namespace platform

#endif // PLATFORM_MACOS
