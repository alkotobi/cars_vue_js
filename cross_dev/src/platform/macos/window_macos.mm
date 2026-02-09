// macOS window implementation
#import <Cocoa/Cocoa.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>

// Forward declaration
namespace platform {
    NSApplication* getApplication();
    void initApplication();
}

#ifdef PLATFORM_MACOS

namespace platform {

void* createWindow(int x, int y, int width, int height, const std::string& title, void* userData) {
    @autoreleasepool {
        initApplication();
        
        NSRect windowRect = NSMakeRect(x, y, width, height);
        NSWindow *window = [[NSWindow alloc] 
            initWithContentRect:windowRect
            styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable | NSWindowStyleMaskMiniaturizable
            backing:NSBackingStoreBuffered
            defer:NO];
        
        NSString *nsTitle = [NSString stringWithUTF8String:title.c_str()];
        [window setTitle:nsTitle];
        [window center];
        
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

} // namespace platform

#endif // PLATFORM_MACOS
