// macOS web view implementation
#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#import <objc/runtime.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>
#include <fstream>
#include <sstream>

#ifdef PLATFORM_MACOS

// Callback types
typedef void (*CreateWindowCallback)(const std::string& title, void* userData);
typedef void (*MessageCallback)(const std::string& jsonMessage, void* userData);

// Store callback in WebView's user content controller
@interface WebViewMessageHandler : NSObject <WKScriptMessageHandler>
@property (assign) CreateWindowCallback createWindowCallback;
@property (assign) void* createWindowUserData;
@property (assign) MessageCallback messageCallback;
@property (assign) void* messageUserData;
@end

@implementation WebViewMessageHandler
- (void)userContentController:(WKUserContentController *)userContentController
      didReceiveScriptMessage:(WKScriptMessage *)message {
    // Convert message to JSON string
    NSError *error = nil;
    NSData *jsonData = [NSJSONSerialization dataWithJSONObject:message.body
                                                       options:0
                                                         error:&error];
    if (jsonData && !error) {
        NSString *jsonString = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];
        std::string jsonMsg = [jsonString UTF8String];
        
        // Try new message callback first (for MessageRouter)
        if (self.messageCallback) {
            self.messageCallback(jsonMsg, self.messageUserData);
            return;
        }
        
        // Fallback to old createWindow callback for backward compatibility
        if ([message.name isEqualToString:@"createWindow"] && self.createWindowCallback) {
            NSDictionary *body = message.body;
            NSString *title = body[@"title"] ?: @"New Window";
            std::string titleStr = [title UTF8String];
            self.createWindowCallback(titleStr, self.createWindowUserData);
        }
    }
}
@end

namespace platform {

void* createWebView(void* parentHandle, int x, int y, int width, int height) {
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
        
        NSRect webViewRect = NSMakeRect(x, y, width, height);
        WKWebView *webView = [[WKWebView alloc] initWithFrame:webViewRect];
        [webView setAutoresizingMask:NSViewWidthSizable | NSViewHeightSizable];
        
        [parentView addSubview:webView];
        
        return (void*)CFBridgingRetain(webView);
    }
}

void destroyWebView(void* webViewHandle) {
    @autoreleasepool {
        if (webViewHandle) {
            WKWebView *webView = (__bridge WKWebView*)webViewHandle;
            [webView removeFromSuperview];
            CFBridgingRelease(webViewHandle);
        }
    }
}

void resizeWebView(void* webViewHandle, int width, int height) {
    @autoreleasepool {
        if (webViewHandle) {
            WKWebView *webView = (__bridge WKWebView*)webViewHandle;
            NSRect newFrame = NSMakeRect(0, 0, width, height);
            [webView setFrame:newFrame];
        }
    }
}

void loadHTMLFile(void* webViewHandle, const std::string& filePath) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        
        // Read file content
        std::ifstream file(filePath);
        if (!file.is_open()) {
            NSLog(@"Failed to open file: %s", filePath.c_str());
            return;
        }
        
        std::stringstream buffer;
        buffer << file.rdbuf();
        std::string htmlContent = buffer.str();
        file.close();
        
        NSString *htmlString = [NSString stringWithUTF8String:htmlContent.c_str()];
        NSURL *baseURL = [NSURL fileURLWithPath:[NSString stringWithUTF8String:filePath.c_str()]];
        
        [webView loadHTMLString:htmlString baseURL:baseURL];
    }
}

void loadHTMLString(void* webViewHandle, const std::string& html) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        NSString *htmlString = [NSString stringWithUTF8String:html.c_str()];
        
        // Load HTML first
        [webView loadHTMLString:htmlString baseURL:nil];
        
        // Re-inject the bridge script after page loads to ensure it's available
        // This ensures the script is available even if the page loads before the callback is set
        NSString *script = @"window.chrome = window.chrome || {}; window.chrome.webview = window.chrome.webview || {}; window.chrome.webview.postMessage = function(message) { if (typeof message === 'object') { window.webkit.messageHandlers.createWindow.postMessage(message); } else if (typeof message === 'string') { try { var parsed = JSON.parse(message); window.webkit.messageHandlers.createWindow.postMessage(parsed); } catch(e) { window.webkit.messageHandlers.createWindow.postMessage({type: 'createWindow', title: message}); } } };";
        // Use a small delay to ensure page is ready
        dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.1 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
            [webView evaluateJavaScript:script completionHandler:nil];
        });
    }
}

void loadURL(void* webViewHandle, const std::string& url) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        NSString *urlString = [NSString stringWithUTF8String:url.c_str()];
        NSURL *nsURL = [NSURL URLWithString:urlString];
        if (nsURL) {
            NSURLRequest *request = [NSURLRequest requestWithURL:nsURL];
            [webView loadRequest:request];
        }
    }
}

void setWebViewCreateWindowCallback(void* webViewHandle, void (*callback)(const std::string& title, void* userData), void* userData) {
    @autoreleasepool {
        if (!webViewHandle || !callback) {
            return;
        }
        
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        WKUserContentController *userContentController = webView.configuration.userContentController;
        
        // Get or create message handler
        WebViewMessageHandler *handler = objc_getAssociatedObject(webView, @"messageHandler");
        if (!handler) {
            handler = [[WebViewMessageHandler alloc] init];
            objc_setAssociatedObject(webView, @"messageHandler", handler, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
            
            // Add script message handler for "nativeMessage" (generic handler)
            [userContentController addScriptMessageHandler:handler name:@"nativeMessage"];
            
            // Inject JavaScript bridge
            NSString *script = @"window.chrome = window.chrome || {}; window.chrome.webview = window.chrome.webview || {}; window.chrome.webview.postMessage = function(message) { var msg = typeof message === 'string' ? JSON.parse(message) : message; window.webkit.messageHandlers.nativeMessage.postMessage(msg); };";
            WKUserScript *userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES];
            [userContentController addUserScript:userScript];
            [webView evaluateJavaScript:script completionHandler:nil];
        }
        
        handler.createWindowCallback = callback;
        handler.createWindowUserData = userData;
    }
}

void setWebViewMessageCallback(void* webViewHandle, void (*callback)(const std::string& jsonMessage, void* userData), void* userData) {
    @autoreleasepool {
        if (!webViewHandle || !callback) {
            return;
        }
        
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        WKUserContentController *userContentController = webView.configuration.userContentController;
        
        // Get or create message handler
        WebViewMessageHandler *handler = objc_getAssociatedObject(webView, @"messageHandler");
        if (!handler) {
            handler = [[WebViewMessageHandler alloc] init];
            objc_setAssociatedObject(webView, @"messageHandler", handler, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
            
            // Add script message handler for "nativeMessage" (generic handler)
            [userContentController addScriptMessageHandler:handler name:@"nativeMessage"];
            
            // Inject JavaScript bridge
            NSString *script = @"window.chrome = window.chrome || {}; window.chrome.webview = window.chrome.webview || {}; window.chrome.webview.postMessage = function(message) { var msg = typeof message === 'string' ? JSON.parse(message) : message; window.webkit.messageHandlers.nativeMessage.postMessage(msg); };";
            WKUserScript *userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES];
            [userContentController addUserScript:userScript];
            [webView evaluateJavaScript:script completionHandler:nil];
        }
        
        handler.messageCallback = callback;
        handler.messageUserData = userData;
    }
}

void postMessageToJavaScript(void* webViewHandle, const std::string& jsonMessage) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        NSString *jsonString = [NSString stringWithUTF8String:jsonMessage.c_str()];
        NSString *jsCode = [NSString stringWithFormat:@"window.postMessage(%@, '*');", jsonString];
        [webView evaluateJavaScript:jsCode completionHandler:nil];
    }
}

} // namespace platform

#endif // PLATFORM_MACOS
