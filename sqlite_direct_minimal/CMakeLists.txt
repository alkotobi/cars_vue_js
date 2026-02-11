cmake_minimum_required(VERSION 3.10)
project(SQLiteDirectMinimal)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Option to use static linking (like UniDAC)
option(USE_STATIC_SQLITE "Use static SQLite linking (embed SQLite in executable)" ON)

# Source files
set(SOURCES
    sqlite_client.cpp
    sqlite_resultset.cpp
    sqlite_prepared_statement.cpp
)

# Header files
set(HEADERS
    sqlite_client.h
    sqlite_resultset.h
    sqlite_prepared_statement.h
)

# SQLite object file path (from UniDAC source)
set(SQLITE_OBJECT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../Source/sqlite3")

if(USE_STATIC_SQLITE)
    # Static linking mode - link SQLite object file
    message(STATUS "Using static SQLite linking")
    
    # Determine platform-specific object file (like UniDAC)
    if(APPLE)
        if(CMAKE_OSX_ARCHITECTURES MATCHES "arm64")
            set(SQLITE_OBJECT_FILE "${SQLITE_OBJECT_DIR}/sqlite3osxarm64.o")
        else()
            set(SQLITE_OBJECT_FILE "${SQLITE_OBJECT_DIR}/sqlite3osx64.o")
        endif()
    elseif(UNIX AND NOT APPLE)
        if(CMAKE_SIZEOF_VOID_P EQUAL 8)
            set(SQLITE_OBJECT_FILE "${SQLITE_OBJECT_DIR}/sqlite3linux64.o")
        else()
            set(SQLITE_OBJECT_FILE "${SQLITE_OBJECT_DIR}/sqlite3linux32.o")
        endif()
    elseif(WIN32)
        if(CMAKE_SIZEOF_VOID_P EQUAL 8)
            set(SQLITE_OBJECT_FILE "${SQLITE_OBJECT_DIR}/sqlite3win64_xe5.o")
        else()
            set(SQLITE_OBJECT_FILE "${SQLITE_OBJECT_DIR}/sqlite3win32.o")
        endif()
    endif()
    
    if(EXISTS "${SQLITE_OBJECT_FILE}")
        message(STATUS "Using SQLite object file: ${SQLITE_OBJECT_FILE}")
        add_definitions(-DSQLITE_STATIC)
        list(APPEND SOURCES "${SQLITE_OBJECT_FILE}")
        set(USE_DYNAMIC_LOADING OFF)
    else()
        message(WARNING "SQLite object file not found: ${SQLITE_OBJECT_FILE}")
        message(WARNING "Falling back to dynamic loading")
        set(USE_STATIC_SQLITE OFF)
        set(USE_DYNAMIC_LOADING ON)
    endif()
else()
    set(USE_DYNAMIC_LOADING ON)
endif()

# Platform libraries
if(WIN32)
    set(PLATFORM_LIBS ws2_32)
else()
    if(USE_STATIC_SQLITE)
        set(PLATFORM_LIBS pthread)
    else()
        set(PLATFORM_LIBS pthread dl)
    endif()
endif()

# Main library
add_library(sqlite_direct_minimal STATIC ${SOURCES} ${HEADERS})
target_link_libraries(sqlite_direct_minimal ${PLATFORM_LIBS})

# Example executables
add_executable(example_basic examples/basic_usage.cpp)
target_link_libraries(example_basic sqlite_direct_minimal ${PLATFORM_LIBS})

add_executable(example_massive_db examples/create_massive_db.cpp)
target_link_libraries(example_massive_db sqlite_direct_minimal ${PLATFORM_LIBS})

# Test executable
add_executable(test_sqlite test_sqlite.cpp)
target_link_libraries(test_sqlite sqlite_direct_minimal ${PLATFORM_LIBS})

# Include directories
target_include_directories(sqlite_direct_minimal PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})
target_include_directories(example_basic PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_include_directories(test_sqlite PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
