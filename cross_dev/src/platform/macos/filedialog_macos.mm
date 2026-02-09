// macOS file dialog implementation
#import <Cocoa/Cocoa.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>
#include <dispatch/dispatch.h>

#ifdef PLATFORM_MACOS

namespace platform {

static bool showOpenFileDialogImpl(void* windowHandle, const std::string& title, const std::string& /*filter*/, std::string& selectedPath) {
    @autoreleasepool {
        NSOpenPanel* panel = [NSOpenPanel openPanel];
        
        // Set dialog title
        if (!title.empty()) {
            [panel setTitle:[NSString stringWithUTF8String:title.c_str()]];
        }
        
        // Set allowed file types using modern API (macOS 11.0+)
        if (@available(macOS 11.0, *)) {
            UTType* htmlType = [UTType typeWithFilenameExtension:@"html"];
            UTType* htmType = [UTType typeWithFilenameExtension:@"htm"];
            NSMutableArray* contentTypes = [[NSMutableArray alloc] init];
            if (htmlType) {
                [contentTypes addObject:htmlType];
            }
            if (htmType) {
                [contentTypes addObject:htmType];
            }
            if (contentTypes.count > 0) {
                [panel setAllowedContentTypes:contentTypes];
            }
        } else {
            // Fallback for older macOS versions (10.15 and earlier)
            #pragma clang diagnostic push
            #pragma clang diagnostic ignored "-Wdeprecated-declarations"
            NSMutableArray* allowedTypes = [[NSMutableArray alloc] init];
            [allowedTypes addObject:@"html"];
            [allowedTypes addObject:@"htm"];
            [panel setAllowedFileTypes:allowedTypes];
            #pragma clang diagnostic pop
        }
        
        // Configure panel
        [panel setAllowsMultipleSelection:NO];
        [panel setCanChooseFiles:YES];
        [panel setCanChooseDirectories:NO];
        
        // Set default directory
        [panel setDirectoryURL:[NSURL fileURLWithPath:[NSHomeDirectory() stringByAppendingPathComponent:@"Documents"]]];
        
        // Activate the app to ensure dialog is visible
        [[NSApplication sharedApplication] activateIgnoringOtherApps:YES];
        
        // Get parent window if provided
        NSWindow* parentWindow = nil;
        if (windowHandle) {
            parentWindow = (__bridge NSWindow*)windowHandle;
            // Make sure parent window is key
            [parentWindow makeKeyWindow];
        }
        
        // Show modal dialog (as sheet if we have a parent window, otherwise standalone)
        NSInteger result;
        if (parentWindow) {
            // Show as sheet attached to parent window
            result = [panel runModal];
        } else {
            // Show as standalone modal dialog
            result = [panel runModal];
        }
        
        if (result == NSModalResponseOK) {
            NSURL* url = [panel URL];
            if (url) {
                selectedPath = [[url path] UTF8String];
                return true;
            }
        }
        
        return false;
    }
}

bool showOpenFileDialog(void* windowHandle, const std::string& title, const std::string& filter, std::string& selectedPath) {
    // Ensure we're on the main thread
    if (![NSThread isMainThread]) {
        __block bool result = false;
        __block std::string path;
        dispatch_sync(dispatch_get_main_queue(), ^{
            result = showOpenFileDialogImpl(windowHandle, title, filter, path);
        });
        selectedPath = path;
        return result;
    }
    
    return showOpenFileDialogImpl(windowHandle, title, filter, selectedPath);
}

} // namespace platform

#endif // PLATFORM_MACOS
