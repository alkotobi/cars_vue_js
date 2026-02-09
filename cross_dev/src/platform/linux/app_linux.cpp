// Linux application lifecycle implementation
#include "../../../include/platform.h"
#include "../platform_impl.h"
#include <X11/Xlib.h>
#include <gtk/gtk.h>

#ifdef PLATFORM_LINUX

namespace platform {

static Display* g_display = nullptr;
static int g_screen = 0;

void initApplication() {
    if (!g_display) {
        g_display = XOpenDisplay(nullptr);
        if (g_display) {
            g_screen = DefaultScreen(g_display);
        }
    }
    // Initialize GTK for WebKit
    if (!gtk_is_initialized()) {
        gtk_init(nullptr, nullptr);
    }
}

void runApplication() {
    if (!g_display) {
        return;
    }
    
    XEvent event;
    while (true) {
        XNextEvent(g_display, &event);
        
        if (event.type == KeyPress) {
            KeySym keysym = XLookupKeysym(&event.xkey, 0);
            if (keysym == XK_Escape) {
                break;
            }
        }
        
        if (event.type == ClientMessage) {
            break;
        }
    }
}

void quitApplication() {
    if (g_display) {
        XCloseDisplay(g_display);
        g_display = nullptr;
    }
}

Display* getDisplay() {
    return g_display;
}

int getScreen() {
    return g_screen;
}

} // namespace platform

#endif // PLATFORM_LINUX
