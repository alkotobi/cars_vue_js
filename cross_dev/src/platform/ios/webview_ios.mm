// iOS web view implementation
#import <UIKit/UIKit.h>
#import <WebKit/WebKit.h>
#import <objc/runtime.h>
#import <Foundation/Foundation.h>
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>
#include <fstream>
#include <sstream>

#ifdef PLATFORM_IOS

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

void* createWebView(void* windowHandle, int x, int y, int width, int height) {
    @autoreleasepool {
        if (!windowHandle) {
            return nullptr;
        }
        
        UIWindow* window = (__bridge UIWindow*)windowHandle;
        UIViewController* viewController = window.rootViewController;
        
        if (!viewController) {
            return nullptr;
        }
        
        CGRect webViewFrame = CGRectMake(x, y, width, height);
        WKWebView* webView = [[WKWebView alloc] initWithFrame:webViewFrame];
        webView.autoresizingMask = UIViewAutoresizingFlexibleWidth | UIViewAutoresizingFlexibleHeight;
        
        [viewController.view addSubview:webView];
        
        return (void*)CFBridgingRetain(webView);
    }
}

void destroyWebView(void* webViewHandle) {
    @autoreleasepool {
        if (webViewHandle) {
            WKWebView* webView = (__bridge_transfer WKWebView*)webViewHandle;
            [webView removeFromSuperview];
        }
    }
}

void loadHTMLFile(void* webViewHandle, const std::string& filePath) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView* webView = (__bridge WKWebView*)webViewHandle;
        
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
        
        NSString* htmlString = [NSString stringWithUTF8String:htmlContent.c_str()];
        NSURL* baseURL = [NSURL fileURLWithPath:[NSString stringWithUTF8String:filePath.c_str()]];
        
        [webView loadHTMLString:htmlString baseURL:baseURL];
    }
}

void loadHTMLString(void* webViewHandle, const std::string& html) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView* webView = (__bridge WKWebView*)webViewHandle;
        NSString* htmlString = [NSString stringWithUTF8String:html.c_str()];
        [webView loadHTMLString:htmlString baseURL:nil];
    }
}

void loadURL(void* webViewHandle, const std::string& url) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView* webView = (__bridge WKWebView*)webViewHandle;
        NSString* urlString = [NSString stringWithUTF8String:url.c_str()];
        NSURL* nsURL = [NSURL URLWithString:urlString];
        if (nsURL) {
            NSURLRequest* request = [NSURLRequest requestWithURL:nsURL];
            [webView loadRequest:request];
        }
    }
}

void setWebViewCreateWindowCallback(void* webViewHandle, void (*callback)(const std::string& title, void* userData), void* userData) {
    @autoreleasepool {
        if (!webViewHandle || !callback) {
            return;
        }
        
        WKWebView* webView = (__bridge WKWebView*)webViewHandle;
        WKUserContentController* userContentController = webView.configuration.userContentController;
        
        // Create message handler
        WebViewMessageHandler* handler = [[WebViewMessageHandler alloc] init];
        handler.createWindowCallback = callback;
        handler.createWindowUserData = userData;
        
        // Add script message handler
        [userContentController addScriptMessageHandler:handler name:@"createWindow"];
        
        // Inject JavaScript to bridge window.chrome.webview.postMessage to WKWebView
        NSString* script = @"window.chrome = window.chrome || {}; window.chrome.webview = window.chrome.webview || {}; window.chrome.webview.postMessage = function(message) { if (typeof message === 'object') { window.webkit.messageHandlers.createWindow.postMessage(message); } };";
        WKUserScript* userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES];
        [userContentController addUserScript:userScript];
        
        // Retain handler to keep it alive
        objc_setAssociatedObject(webView, (__bridge const void*)@"messageHandler", handler, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
    }
}

void setWebViewMessageCallback(void* webViewHandle, void (*callback)(const std::string& jsonMessage, void* userData), void* userData) {
    @autoreleasepool {
        if (!webViewHandle || !callback) {
            return;
        }
        
        WKWebView* webView = (__bridge WKWebView*)webViewHandle;
        WKUserContentController* userContentController = webView.configuration.userContentController;
        
        // Get or create message handler
        WebViewMessageHandler* handler = objc_getAssociatedObject(webView, (__bridge const void*)@"messageHandler");
        if (!handler) {
            handler = [[WebViewMessageHandler alloc] init];
            objc_setAssociatedObject(webView, (__bridge const void*)@"messageHandler", handler, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
        }
        
        // Set message callback
        handler.messageCallback = callback;
        handler.messageUserData = userData;
        
        // Add script message handler for generic messages
        [userContentController addScriptMessageHandler:handler name:@"nativeMessage"];
        
        // Inject JavaScript bridge
        NSString* script = @"window.webkit = window.webkit || {}; window.webkit.messageHandlers = window.webkit.messageHandlers || {}; window.webkit.messageHandlers.nativeMessage = { postMessage: function(message) { if (typeof message === 'object') { window.webkit.messageHandlers.nativeMessage.postMessage(message); } } }; window.postMessage = function(message) { if (typeof message === 'object') { window.webkit.messageHandlers.nativeMessage.postMessage(message); } };";
        WKUserScript* userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES];
        [userContentController addUserScript:userScript];
    }
}

void postMessageToJavaScript(void* webViewHandle, const std::string& jsonMessage) {
    @autoreleasepool {
        if (!webViewHandle) {
            return;
        }
        
        WKWebView* webView = (__bridge WKWebView*)webViewHandle;
        NSString* messageString = [NSString stringWithUTF8String:jsonMessage.c_str()];
        
        // Execute JavaScript to post message
        NSString* script = [NSString stringWithFormat:@"if (window.postMessage) { window.postMessage(%@); }", messageString];
        [webView evaluateJavaScript:script completionHandler:^(id result, NSError* error) {
            if (error) {
                NSLog(@"Error posting message to JavaScript: %@", error);
            }
        }];
    }
}

} // namespace platform

#endif // PLATFORM_IOS
