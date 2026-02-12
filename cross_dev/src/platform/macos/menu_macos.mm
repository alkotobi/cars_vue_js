// macOS menu implementation (main menu bar + context menu)
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include "nlohmann/json.hpp"
#include <string>

#ifdef PLATFORM_MACOS

using json = nlohmann::json;

typedef void (*MenuItemCallback)(const std::string& itemId, void* userData);

@interface MenuItemTarget2 : NSObject
@property (assign) MenuItemCallback callback;
@property (assign) void* userData;
@end
@implementation MenuItemTarget2
- (void)menuItemClicked:(NSMenuItem*)sender {
    if (self.callback) {
        NSString* itemId = objc_getAssociatedObject(sender, "itemId");
        if (itemId) {
            std::string idStr([itemId UTF8String]);
            self.callback(idStr, self.userData);
        }
    }
}
@end

static void addMenuItems2(NSMenu* menu, const json& items, MenuItemTarget2* target) {
    for (const auto& item : items) {
        std::string idStr = item.value("id", "");
        std::string label = item.value("label", "");
        if (idStr == "-" || label == "-") {
            [menu addItem:[NSMenuItem separatorItem]];
            continue;
        }
        if (item.contains("items")) {
            NSMenuItem* subItem = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:label.c_str()]
                                                            action:nil keyEquivalent:@""];
            NSMenu* subMenu = [[NSMenu alloc] initWithTitle:[NSString stringWithUTF8String:label.c_str()]];
            addMenuItems2(subMenu, item["items"], target);
            [menu addItem:subItem];
            [subItem setSubmenu:subMenu];
        } else {
            std::string shortcut = item.value("shortcut", "");
            NSMenuItem* nsItem = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:label.c_str()]
                                                             action:@selector(menuItemClicked:)
                                                      keyEquivalent:[NSString stringWithUTF8String:shortcut.c_str()]];
            [nsItem setTarget:target];
            objc_setAssociatedObject(nsItem, "itemId", [NSString stringWithUTF8String:idStr.c_str()], OBJC_ASSOCIATION_RETAIN_NONATOMIC);
            [menu addItem:nsItem];
        }
    }
}

namespace platform {

void setWindowMainMenu(void* windowHandle, const std::string& menuJson,
                       void (*itemCallback)(const std::string& itemId, void* userData), void* userData) {
    @autoreleasepool {
        if (!itemCallback || menuJson.empty()) return;
        try {
            json menus = json::parse(menuJson);
            if (!menus.is_array()) return;
            MenuItemTarget2* target = [[MenuItemTarget2 alloc] init];
            target.callback = itemCallback;
            target.userData = userData;
            NSMenu* mainMenu = [[NSMenu alloc] initWithTitle:@"MainMenu"];
            for (const auto& top : menus) {
                std::string label = top.value("label", "");
                if (label.empty()) continue;
                NSMenuItem* topItem = [[NSMenuItem alloc] initWithTitle:[NSString stringWithUTF8String:label.c_str()]
                                                                action:nil keyEquivalent:@""];
                NSMenu* subMenu = [[NSMenu alloc] initWithTitle:[NSString stringWithUTF8String:label.c_str()]];
                if (top.contains("items")) {
                    addMenuItems2(subMenu, top["items"], target);
                }
                [mainMenu addItem:topItem];
                [topItem setSubmenu:subMenu];
            }
            [NSApp setMainMenu:mainMenu];
        } catch (...) {}
    }
}

void showContextMenu(void* parentHandle, int x, int y, const std::string& itemsJson,
                     void (*itemCallback)(const std::string& itemId, void* userData), void* userData) {
    @autoreleasepool {
        if (!parentHandle || !itemCallback || itemsJson.empty()) return;
        NSView* view = nil;
        if ([(__bridge id)parentHandle isKindOfClass:[NSWindow class]]) {
            view = [(__bridge NSWindow*)parentHandle contentView];
        } else if ([(__bridge id)parentHandle isKindOfClass:[NSView class]]) {
            view = (__bridge NSView*)parentHandle;
        }
        if (!view) return;
        try {
            json items = json::parse(itemsJson);
            if (!items.is_array()) return;
            MenuItemTarget2* target = [[MenuItemTarget2 alloc] init];
            target.callback = itemCallback;
            target.userData = userData;
            NSMenu* menu = [[NSMenu alloc] initWithTitle:@"ContextMenu"];
            addMenuItems2(menu, items, target);
            NSPoint pt = NSMakePoint((CGFloat)x, (CGFloat)y);
            [menu popUpMenuPositioningItem:nil atLocation:pt inView:view];
        } catch (...) {}
    }
}

} // namespace platform

#endif // PLATFORM_MACOS
