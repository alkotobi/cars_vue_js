// iOS window implementation
#import <UIKit/UIKit.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>

namespace platform {
    void initApplication();
    void setMainWindow(UIWindow* window);
}

#ifdef PLATFORM_IOS

namespace platform {

void* createWindow(int x, int y, int width, int height, const std::string& title, void* userData) {
    @autoreleasepool {
        initApplication();
        
        // On iOS, we create a UIWindow that fills the screen
        // x, y, width, height are ignored as iOS manages window size
        #pragma clang diagnostic push
        #pragma clang diagnostic ignored "-Wdeprecated-declarations"
        CGRect screenBounds = [[UIScreen mainScreen] bounds];
        UIWindow* window = [[UIWindow alloc] initWithFrame:screenBounds];
        #pragma clang diagnostic pop
        
        // Create a view controller to hold our content
        UIViewController* viewController = [[UIViewController alloc] init];
        viewController.view.backgroundColor = [UIColor whiteColor];
        
        // Create a label to display the title
        UILabel* titleLabel = [[UILabel alloc] init];
        titleLabel.text = [NSString stringWithUTF8String:title.c_str()];
        titleLabel.textAlignment = NSTextAlignmentCenter;
        titleLabel.font = [UIFont boldSystemFontOfSize:24];
        titleLabel.translatesAutoresizingMaskIntoConstraints = NO;
        
        [viewController.view addSubview:titleLabel];
        
        // Center the label
        [NSLayoutConstraint activateConstraints:@[
            [titleLabel.centerXAnchor constraintEqualToAnchor:viewController.view.centerXAnchor],
            [titleLabel.centerYAnchor constraintEqualToAnchor:viewController.view.centerYAnchor]
        ]];
        
        window.rootViewController = viewController;
        setMainWindow(window);
        
        return (void*)CFBridgingRetain(window);
    }
}

void destroyWindow(void* handle) {
    @autoreleasepool {
        if (handle) {
            UIWindow* window = (__bridge_transfer UIWindow*)handle;
            window.hidden = YES;
            setMainWindow(nullptr);
        }
    }
}

void showWindow(void* handle) {
    @autoreleasepool {
        if (handle) {
            UIWindow* window = (__bridge UIWindow*)handle;
            [window makeKeyAndVisible];
        }
    }
}

void hideWindow(void* handle) {
    @autoreleasepool {
        if (handle) {
            UIWindow* window = (__bridge UIWindow*)handle;
            window.hidden = YES;
        }
    }
}

void setWindowTitle(void* handle, const std::string& title) {
    @autoreleasepool {
        if (handle) {
            UIWindow* window = (__bridge UIWindow*)handle;
            if (window.rootViewController && window.rootViewController.view) {
                // Find the label and update it
                for (UIView* subview in window.rootViewController.view.subviews) {
                    if ([subview isKindOfClass:[UILabel class]]) {
                        UILabel* label = (UILabel*)subview;
                        label.text = [NSString stringWithUTF8String:title.c_str()];
                        break;
                    }
                }
            }
        }
    }
}

bool isWindowVisible(void* handle) {
    @autoreleasepool {
        if (handle) {
            UIWindow* window = (__bridge UIWindow*)handle;
            return !window.hidden && window.windowLevel == UIWindowLevelNormal;
        }
        return false;
    }
}

void setWindowResizeCallback(void* windowHandle, void (*callback)(int width, int height, void* userData), void* userData) {
    // iOS windows are typically full-screen and don't resize in the traditional sense
    // This is a stub implementation
    (void)windowHandle;
    (void)callback;
    (void)userData;
}

} // namespace platform

#endif // PLATFORM_IOS
