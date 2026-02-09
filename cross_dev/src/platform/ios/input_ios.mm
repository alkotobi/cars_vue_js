// iOS input field implementation
#import <UIKit/UIKit.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>

#ifdef PLATFORM_IOS

namespace platform {

void* createInputField(void* windowHandle, int x, int y, int width, int height, const std::string& placeholder) {
    @autoreleasepool {
        if (!windowHandle) {
            return nullptr;
        }
        
        UIWindow* window = (__bridge UIWindow*)windowHandle;
        UIViewController* viewController = window.rootViewController;
        
        if (!viewController) {
            return nullptr;
        }
        
        CGRect inputFrame = CGRectMake(x, y, width, height);
        UITextField* textField = [[UITextField alloc] initWithFrame:inputFrame];
        textField.placeholder = [NSString stringWithUTF8String:placeholder.c_str()];
        textField.borderStyle = UITextBorderStyleRoundedRect;
        textField.autoresizingMask = UIViewAutoresizingFlexibleRightMargin | UIViewAutoresizingFlexibleBottomMargin;
        textField.keyboardType = UIKeyboardTypeURL;
        textField.autocapitalizationType = UITextAutocapitalizationTypeNone;
        textField.autocorrectionType = UITextAutocorrectionTypeNo;
        
        [viewController.view addSubview:textField];
        
        return (void*)CFBridgingRetain(textField);
    }
}

void destroyInputField(void* inputHandle) {
    @autoreleasepool {
        if (inputHandle) {
            UITextField* textField = (__bridge_transfer UITextField*)inputHandle;
            [textField removeFromSuperview];
        }
    }
}

void setInputText(void* inputHandle, const std::string& text) {
    @autoreleasepool {
        if (inputHandle) {
            UITextField* textField = (__bridge UITextField*)inputHandle;
            textField.text = [NSString stringWithUTF8String:text.c_str()];
        }
    }
}

std::string getInputText(void* inputHandle) {
    @autoreleasepool {
        if (inputHandle) {
            UITextField* textField = (__bridge UITextField*)inputHandle;
            if (textField.text) {
                return [textField.text UTF8String];
            }
        }
        return "";
    }
}

} // namespace platform

#endif // PLATFORM_IOS
