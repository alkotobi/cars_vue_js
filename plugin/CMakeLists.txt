cmake_minimum_required(VERSION 3.15)
project(CarsCrossDevPlugin LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_library(CrossDevAppPlugin SHARED
    app_handlers.cpp
)

target_include_directories(CrossDevAppPlugin PUBLIC
    /Users/merhab/dev/dev_in_progress/cross_dev/include
)

set_target_properties(CrossDevAppPlugin PROPERTIES
    OUTPUT_NAME "CrossDevAppPlugin"
)

if(APPLE)
    target_link_options(CrossDevAppPlugin PRIVATE "-undefined" "dynamic_lookup")
endif()
