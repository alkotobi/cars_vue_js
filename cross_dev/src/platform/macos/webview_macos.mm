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
+ (NSString *)crossdevBridgeScript;
@end

// Custom WKWebView that accepts first mouse click when window is inactive.
// By default, the first click only activates the window; returning YES here
// lets the click also be delivered to the WebView (one-click button activation).
@interface ClickThroughWebView : WKWebView
@end
@implementation ClickThroughWebView
- (BOOL)acceptsFirstMouse:(NSEvent *)event {
    return YES;
}
@end

@implementation WebViewMessageHandler
+ (NSString *)crossdevBridgeScript {
    return @"(function(){"
        "var _pending=new Map();"
        "var _eventListeners={};"
        "function _ab2b64(ab){var u8=new Uint8Array(ab);var bin='';for(var i=0;i<u8.length;i++)bin+=String.fromCharCode(u8[i]);return btoa(bin);}"
        "function _b642ab(s){var bin=atob(s);var u8=new Uint8Array(bin.length);for(var i=0;i<bin.length;i++)u8[i]=bin.charCodeAt(i);return u8.buffer;}"
        "function _toWire(obj){"
        "if(obj===null||typeof obj!=='object')return obj;"
        "if(obj instanceof ArrayBuffer){return {__base64:_ab2b64(obj)};}"
        "if(ArrayBuffer.isView(obj)){return {__base64:_ab2b64(obj.buffer.slice(obj.byteOffset,obj.byteOffset+obj.byteLength))};}"
        "if(Array.isArray(obj)){return obj.map(_toWire);}"
        "var out={};for(var k in obj)if(obj.hasOwnProperty(k))out[k]=_toWire(obj[k]);return out;"
        "}"
        "window.addEventListener('message',function(e){"
        "var d=e.data;"
        "if(!d)return;"
        "if(d.type==='crossdev:event'){"
        "var name=d.name,payload=d.payload||{};"
        "var list=_eventListeners[name];"
        "if(list)list.forEach(function(fn){try{fn(payload)}catch(err){console.error(err)}});"
        "return;"
        "}"
        "if(d.requestId){"
        "var h=_pending.get(d.requestId);"
        "if(h){_pending.delete(d.requestId);"
        "var res=d.result;"
        "if(h.binary&&res&&typeof res.data==='string'){res=Object.assign({},res);res.data=_b642ab(res.data);}"
        "d.error?h.reject(new Error(d.error)):h.resolve(res);"
        "}"
        "}"
        "});"
        "function _post(msg){"
        "if(window.webkit&&window.webkit.messageHandlers&&window.webkit.messageHandlers.nativeMessage){"
        "window.webkit.messageHandlers.nativeMessage.postMessage(msg);"
        "}"
        "}"
        "var CrossDev={"
        "invoke:function(type,payload,opts){"
        "var opt=opts||{};"
        "return new Promise(function(resolve,reject){"
        "var rid=Date.now()+'-'+Math.random();"
        "_pending.set(rid,{resolve:resolve,reject:reject,binary:!!opt.binaryResponse});"
        "setTimeout(function(){if(_pending.has(rid)){_pending.delete(rid);reject(new Error('Request timeout'));}},10000);"
        "_post({type:type,payload:_toWire(payload||{}),requestId:rid});"
        "});"
        "},"
        "events:{"
        "on:function(name,fn){"
        "if(!_eventListeners[name])_eventListeners[name]=[];"
        "_eventListeners[name].push(fn);"
        "return function(){var i=_eventListeners[name].indexOf(fn);if(i>=0)_eventListeners[name].splice(i,1);};"
        "}"
        "}"
        "};"
        "Object.freeze(CrossDev.events);"
        "Object.freeze(CrossDev);"
        "Object.defineProperty(window,'CrossDev',{value:CrossDev,configurable:false,writable:false});"
        "window.chrome=window.chrome||{};"
        "window.chrome.webview=window.chrome.webview||{};"
        "window.chrome.webview.postMessage=function(m){var msg=typeof m==='string'?JSON.parse(m):m;_post(msg);};"
        "})();";
}
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

void setWebViewPreloadScript(void* webViewHandle, const std::string& scriptContent) {
    @autoreleasepool {
        if (!webViewHandle) return;
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        if (scriptContent.empty()) {
            objc_setAssociatedObject(webView, @"customPreloadScript", nil, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
        } else {
            NSString *script = [NSString stringWithUTF8String:scriptContent.c_str()];
            objc_setAssociatedObject(webView, @"customPreloadScript", script, OBJC_ASSOCIATION_RETAIN_NONATOMIC);
        }
    }
}

static NSString* getPreloadScriptForWebView(WKWebView *webView) {
    NSString *custom = objc_getAssociatedObject(webView, @"customPreloadScript");
    if (custom && [custom length] > 0) {
        return custom;
    }
    return [WebViewMessageHandler crossdevBridgeScript];
}

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
        WKWebView *webView = [[ClickThroughWebView alloc] initWithFrame:webViewRect];
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
        
        // Re-inject bridge after page loads (in case WKUserScript ran before callback was set)
        dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.1 * NSEC_PER_SEC)), dispatch_get_main_queue(), ^{
            NSString *script = getPreloadScriptForWebView(webView);
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
            
            // Inject JavaScript bridge (custom preload or built-in) in page world
            NSString *script = getPreloadScriptForWebView(webView);
            WKUserScript *userScript;
            if (@available(macOS 11.0, *)) {
                userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES inContentWorld:WKContentWorld.pageWorld];
            } else {
                userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES];
            }
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
            
            // Inject JavaScript bridge (custom preload or built-in) in page world
            NSString *script = getPreloadScriptForWebView(webView);
            WKUserScript *userScript;
            if (@available(macOS 11.0, *)) {
                userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES inContentWorld:WKContentWorld.pageWorld];
            } else {
                userScript = [[WKUserScript alloc] initWithSource:script injectionTime:WKUserScriptInjectionTimeAtDocumentStart forMainFrameOnly:YES];
            }
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

void executeWebViewScript(void* webViewHandle, const std::string& script) {
    @autoreleasepool {
        if (!webViewHandle || script.empty()) return;
        WKWebView *webView = (__bridge WKWebView*)webViewHandle;
        NSString *nsScript = [NSString stringWithUTF8String:script.c_str()];
        [webView evaluateJavaScript:nsScript completionHandler:nil];
    }
}

} // namespace platform

#endif // PLATFORM_MACOS
