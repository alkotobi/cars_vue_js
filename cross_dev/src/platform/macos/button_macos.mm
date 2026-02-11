// macOS button implementation
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>

#ifdef PLATFORM_MACOS

// Button target class (must be outside namespace)
@interface ButtonTarget : NSObject
@property (assign) void* userData;
@property (assign) void (*callback)(void*);
- (void)buttonClicked:(id)sender;
@end

@implementation ButtonTarget
- (void)buttonClicked:(id)sender {
    if (self.callback && self.userData) {
        self.callback(self.userData);
    }
}
@end

namespace platform {

void* createButton(void* parentHandle, int x, int y, int width, int height, const std::string& label, void* userData) {
    @autoreleasepool {
        if (!parentHandle) {
            return nullptr;
        }
        
        NSView *parentView = nullptr;
        
        // Get parent view - could be NSWindow's contentView or an NSView (from Container)
        if ([(__bridge id)parentHandle isKindOfClass:[NSWindow class]]) {
            NSWindow *window = (__bridge NSWindow*)parentHandle;
            parentView = [window contentView];
        } else if ([(__bridge id)parentHandle isKindOfClass:[NSView class]]) {
            parentView = (__bridge NSView*)parentHandle;
        } else {
            return nullptr;
        }
        
        // Convert coordinates: parent view uses bottom-left origin, but we receive top-left origin
        NSRect parentBounds = [parentView bounds];
        NSRect buttonRect = NSMakeRect(x, parentBounds.size.height - y - height, width, height);
        NSButton *button = [[NSButton alloc] initWithFrame:buttonRect];
        [button setTitle:[NSString stringWithUTF8String:label.c_str()]];
        [button setButtonType:NSButtonTypeMomentaryPushIn];
        [button setBezelStyle:NSBezelStyleRounded];
        [button setAutoresizingMask:NSViewMaxXMargin | NSViewMinYMargin];
        
        // Create target for callback
        ButtonTarget *target = [[ButtonTarget alloc] init];
        target.userData = userData;
        [button setTarget:target];
        [button setAction:@selector(buttonClicked:)];
        
        // Retain target to keep it alive
        objc_setAssociatedObject(button, "target", target, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
        
        [parentView addSubview:button];
        
        return (void*)CFBridgingRetain(button);
    }
}

void destroyButton(void* buttonHandle) {
    @autoreleasepool {
        if (buttonHandle) {
            NSButton *button = (__bridge NSButton*)buttonHandle;
            [button removeFromSuperview];
            CFBridgingRelease(buttonHandle);
        }
    }
}

void setButtonCallback(void* buttonHandle, void (*callback)(void*)) {
    @autoreleasepool {
        if (!buttonHandle) {
            return;
        }
        
        NSButton *button = (__bridge NSButton*)buttonHandle;
        ButtonTarget *target = objc_getAssociatedObject(button, "target");
        if (target) {
            target.callback = callback;
        }
    }
}

} // namespace platform

#endif // PLATFORM_MACOS
