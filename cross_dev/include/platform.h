#ifndef PLATFORM_H
#define PLATFORM_H

// Platform detection
#if defined(__APPLE__) && defined(__MACH__)
    #include <TargetConditionals.h>
    #if TARGET_OS_IPHONE || TARGET_OS_SIMULATOR
        #define PLATFORM_IOS
        #define PLATFORM_NAME "iOS"
    #else
        #define PLATFORM_MACOS
        #define PLATFORM_NAME "macOS"
    #endif
#elif defined(_WIN32) || defined(_WIN64)
    #define PLATFORM_WINDOWS
    #define PLATFORM_NAME "Windows"
#elif defined(__linux__)
    #define PLATFORM_LINUX
    #define PLATFORM_NAME "Linux"
#else
    #error "Unsupported platform"
#endif

#endif // PLATFORM_H
