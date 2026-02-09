// macOS application lifecycle implementation
#import <Cocoa/Cocoa.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"

#ifdef PLATFORM_MACOS

namespace platform {

static NSApplication* g_app = nullptr;

void initApplication() {
    @autoreleasepool {
        if (!g_app) {
            g_app = [NSApplication sharedApplication];
            [g_app setActivationPolicy:NSApplicationActivationPolicyRegular];
        }
    }
}

void runApplication() {
    @autoreleasepool {
        if (g_app) {
            [g_app run];
        }
    }
}

void quitApplication() {
    @autoreleasepool {
        if (g_app) {
            [g_app terminate:nil];
        }
    }
}

NSApplication* getApplication() {
    return g_app;
}

} // namespace platform

#endif // PLATFORM_MACOS
