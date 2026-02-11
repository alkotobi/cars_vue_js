// macOS window implementation
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>

// Forward declaration
namespace platform {
    NSApplication* getApplication();
    void initApplication();
}

#ifdef PLATFORM_MACOS

// Resize callback storage (must be at global scope for Objective-C)
typedef void (*ResizeCallback)(int width, int height, void* userData);

// Objective-C declarations must be at global scope, not inside C++ namespace
@interface WindowResizeDelegate : NSObject <NSWindowDelegate>
@property (assign) ResizeCallback resizeCallback;
@property (assign) void* resizeUserData;
@end

@implementation WindowResizeDelegate
- (void)windowDidResize:(NSNotification *)notification {
    if (self.resizeCallback) {
        NSWindow *window = notification.object;
        NSRect frame = [window contentRectForFrameRect:[window frame]];
        self.resizeCallback((int)frame.size.width, (int)frame.size.height, self.resizeUserData);
    }
}
@end

namespace platform {

void* createWindow(int x, int y, int width, int height, const std::string& title, void* userData) {
    @autoreleasepool {
        initApplication();
        
        // Convert screen coordinates (top-left origin) to Cocoa coordinates (bottom-left origin)
        NSScreen *mainScreen = [NSScreen mainScreen];
        NSRect screenRect = [mainScreen frame];
        NSRect windowRect = NSMakeRect(x, screenRect.size.height - y - height, width, height);
        
        NSWindow *window = [[NSWindow alloc] 
            initWithContentRect:windowRect
            styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable
            backing:NSBackingStoreBuffered
            defer:NO];
        
        NSString *nsTitle = [NSString stringWithUTF8String:title.c_str()];
        [window setTitle:nsTitle];
        
        // Center the window if coordinates are default
        if (x == 0 && y == 0) {
            [window center];
        }
        
        return (void*)CFBridgingRetain(window);
    }
}

void destroyWindow(void* handle) {
    @autoreleasepool {
        if (handle) {
            NSWindow *window = (__bridge NSWindow*)handle;
            [window close];
            CFBridgingRelease(handle);
        }
    }
}

void showWindow(void* handle) {
    @autoreleasepool {
        if (handle) {
            NSWindow *window = (__bridge NSWindow*)handle;
            [window makeKeyAndOrderFront:nil];
            NSApplication* app = getApplication();
            if (app) {
                [app activateIgnoringOtherApps:YES];
            }
        }
    }
}

void hideWindow(void* handle) {
    @autoreleasepool {
        if (handle) {
            NSWindow *window = (__bridge NSWindow*)handle;
            [window orderOut:nil];
        }
    }
}

void setWindowTitle(void* handle, const std::string& title) {
    @autoreleasepool {
        if (handle) {
            NSWindow *window = (__bridge NSWindow*)handle;
            NSString *nsTitle = [NSString stringWithUTF8String:title.c_str()];
            [window setTitle:nsTitle];
        }
    }
}

bool isWindowVisible(void* handle) {
    @autoreleasepool {
        if (handle) {
            NSWindow *window = (__bridge NSWindow*)handle;
            return [window isVisible];
        }
        return false;
    }
}

void setWindowResizeCallback(void* windowHandle, ResizeCallback callback, void* userData) {
    @autoreleasepool {
        if (!windowHandle || !callback) {
            return;
        }
        
        NSWindow *window = (__bridge NSWindow*)windowHandle;
        
        // Get or create delegate
        WindowResizeDelegate *delegate = objc_getAssociatedObject(window, @"resizeDelegate");
        if (!delegate) {
            delegate = [[WindowResizeDelegate alloc] init];
            objc_setAssociatedObject(window, @"resizeDelegate", delegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
            [window setDelegate:delegate];
        }
        
        delegate.resizeCallback = callback;
        delegate.resizeUserData = userData;
    }
}

} // namespace platform

#endif // PLATFORM_MACOS
