# iOS Toolchain for CMake
# This file helps CMake configure for iOS builds

set(CMAKE_SYSTEM_NAME iOS)
set(CMAKE_SYSTEM_VERSION 12.0)
set(CMAKE_OSX_DEPLOYMENT_TARGET 12.0)

# Find the iOS SDK
execute_process(
    COMMAND xcrun --sdk iphoneos --show-sdk-path
    OUTPUT_VARIABLE IOS_SDK_PATH
    OUTPUT_STRIP_TRAILING_WHITESPACE
)

if(NOT IOS_SDK_PATH)
    message(FATAL_ERROR "iOS SDK not found. Make sure Xcode is installed.")
endif()

set(CMAKE_OSX_SYSROOT ${IOS_SDK_PATH})
set(CMAKE_OSX_ARCHITECTURES "arm64")

# Find compilers
execute_process(
    COMMAND xcrun --find clang
    OUTPUT_VARIABLE CMAKE_C_COMPILER
    OUTPUT_STRIP_TRAILING_WHITESPACE
)

execute_process(
    COMMAND xcrun --find clang++
    OUTPUT_VARIABLE CMAKE_CXX_COMPILER
    OUTPUT_STRIP_TRAILING_WHITESPACE
)

set(CMAKE_C_COMPILER ${CMAKE_C_COMPILER})
set(CMAKE_CXX_COMPILER ${CMAKE_CXX_COMPILER})

# Set iOS-specific flags
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fobjc-arc")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fobjc-arc -std=c++17")

message(STATUS "iOS SDK: ${IOS_SDK_PATH}")
message(STATUS "C Compiler: ${CMAKE_C_COMPILER}")
message(STATUS "CXX Compiler: ${CMAKE_CXX_COMPILER}")
