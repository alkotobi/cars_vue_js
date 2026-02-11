# SQLite Static Linking

This document explains how static linking works in the SQLite Direct Minimal implementation, following UniDAC's approach.

## Overview

The project supports two modes:
1. **Static Linking** (default): Embeds SQLite directly into the executable using pre-compiled object files
2. **Dynamic Loading**: Loads SQLite library at runtime (like MySQL direct client)

## Static Linking Mode

### How It Works

1. **Object Files**: Uses pre-compiled SQLite object files (`.o` files) from UniDAC's source tree
   - Location: `../Source/sqlite3/`
   - Platform-specific files:
     - macOS ARM64: `sqlite3osxarm64.o`
     - macOS x64: `sqlite3osx64.o`
     - Linux x64: `sqlite3linux64.o`
     - Linux x32: `sqlite3linux32.o`
     - Windows x64: `sqlite3win64_xe5.o`
     - Windows x32: `sqlite3win32.o`

2. **CMake Configuration**: 
   - Option `USE_STATIC_SQLITE` (default: ON)
   - Automatically selects the correct object file for the platform
   - Adds `-DSQLITE_STATIC` preprocessor definition
   - Links the object file directly into the executable

3. **Code Implementation**:
   - When `SQLITE_STATIC` is defined, function pointers are assigned directly to SQLite functions
   - No dynamic library loading required
   - Functions are linked at compile time

### Building with Static Linking

```bash
cd sqlite_direct_minimal
mkdir build && cd build
cmake -DUSE_STATIC_SQLITE=ON ..
make
```

Or simply (static is default):
```bash
cmake ..
make
```

### Building with Dynamic Loading

```bash
cmake -DUSE_STATIC_SQLITE=OFF ..
make
```

## Comparison with UniDAC

UniDAC uses Pascal's `{$L}` directive to link object files:
```pascal
{$L 'sqlite3\sqlite3win64_xe5.o'}
```

Our CMakeLists.txt does the equivalent:
```cmake
list(APPEND SOURCES "${SQLITE_OBJECT_FILE}")
```

## Benefits of Static Linking

1. **No Runtime Dependencies**: Executable is self-contained
2. **No Library Path Issues**: No need to find `libsqlite3.so` or `sqlite3.dll`
3. **Version Control**: Guaranteed SQLite version
4. **Simplified Deployment**: Single executable file

## Benefits of Dynamic Loading

1. **Smaller Executable**: SQLite library not embedded
2. **System Library Updates**: Can use updated system SQLite
3. **Flexibility**: Can switch SQLite versions without recompiling

## Code Structure

The `SQLiteClient` class handles both modes:

```cpp
#ifdef SQLITE_STATIC
    // Static: Direct function pointer assignment
    extern "C" {
        int sqlite3_open_v2(...);
        // ... other functions
    }
    sqlite3_open_v2_ = sqlite3_open_v2;
    // ...
#else
    // Dynamic: Load library and resolve symbols
    sqlite_lib_handle_ = dlopen("libsqlite3.so", RTLD_LAZY);
    sqlite3_open_v2_ = (sqlite3_open_v2_func)dlsym(sqlite_lib_handle_, "sqlite3_open_v2");
    // ...
#endif
```

## Platform Support

- ✅ macOS (x64 and ARM64)
- ✅ Linux (x64 and x32)
- ✅ Windows (x64 and x32)

The CMakeLists.txt automatically detects the platform and selects the appropriate object file.
