// Linux window implementation
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <string>
#include <X11/Xlib.h>
#include <X11/Xutil.h>

namespace platform {
    Display* getDisplay();
    int getScreen();
    void initApplication();
}

#ifdef PLATFORM_LINUX

namespace platform {

typedef void (*ResizeCallback)(int width, int height, void* userData);

struct WindowData {
    Display* display;
    Window window;
    GtkWidget* gtkWindow;
    bool visible;
    std::string title;
    ResizeCallback resizeCallback;
    void* resizeUserData;
};

void* createWindow(int x, int y, int width, int height, const std::string& title, void* userData) {
    initApplication();
    
    if (!getDisplay()) {
        return nullptr;
    }
    
    WindowData* data = new WindowData;
    data->display = getDisplay();
    data->title = title;
    data->visible = false;
    data->gtkWindow = nullptr;
    data->resizeCallback = nullptr;
    data->resizeUserData = nullptr;
    
    Window root = RootWindow(data->display, getScreen());
    
    XSetWindowAttributes attrs;
    attrs.event_mask = ExposureMask | KeyPressMask | ButtonPressMask | StructureNotifyMask;
    attrs.background_pixel = WhitePixel(data->display, getScreen());
    attrs.border_pixel = BlackPixel(data->display, getScreen());
    
    unsigned long attrmask = CWBackPixel | CWBorderPixel | CWEventMask;
    
    data->window = XCreateWindow(
        data->display,
        root,
        x, y, width, height,
        1,
        DefaultDepth(data->display, getScreen()),
        InputOutput,
        DefaultVisual(data->display, getScreen()),
        attrmask,
        &attrs
    );
    
    XStoreName(data->display, data->window, title.c_str());
    XMapWindow(data->display, data->window);
    
    return data;
}

void destroyWindow(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        if (data->display && data->window) {
            XDestroyWindow(data->display, data->window);
        }
        delete data;
    }
}

void showWindow(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        if (data->display && data->window) {
            XMapRaised(data->display, data->window);
            XFlush(data->display);
            data->visible = true;
        }
    }
}

void hideWindow(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        if (data->display && data->window) {
            XUnmapWindow(data->display, data->window);
            XFlush(data->display);
            data->visible = false;
        }
    }
}

void setWindowTitle(void* handle, const std::string& title) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        data->title = title;
        if (data->display && data->window) {
            XStoreName(data->display, data->window, title.c_str());
            XFlush(data->display);
        }
    }
}

bool isWindowVisible(void* handle) {
    if (handle) {
        WindowData* data = static_cast<WindowData*>(handle);
        if (data->display && data->window) {
            XWindowAttributes attrs;
            if (XGetWindowAttributes(data->display, data->window, &attrs)) {
                return data->visible && (attrs.map_state == IsViewable);
            }
        }
    }
    return false;
}

void setWindowResizeCallback(void* windowHandle, void (*callback)(int width, int height, void* userData), void* userData) {
    if (windowHandle && callback) {
        WindowData* data = static_cast<WindowData*>(windowHandle);
        data->resizeCallback = callback;
        data->resizeUserData = userData;
        
        // Note: For X11, resize events are handled via ConfigureNotify
        // The application's event loop should check for ConfigureNotify and call the callback
        // For GTK, we can use size-allocate signal if gtkWindow is available
        if (data->gtkWindow) {
            // GTK-based resize handling would go here if needed
        }
    }
}

void setWindowCloseCallback(void*, void (*)(void*), void*) {
    // Linux: Close handling would need WM_DELETE_WINDOW - stub for now
}

} // namespace platform

#endif // PLATFORM_LINUX
