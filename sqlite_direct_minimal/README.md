# SQLite Direct Client - Minimal Implementation

A minimal C++ implementation of SQLite direct connection, following UniDAC's approach.

## Features

- **Dynamic Library Loading** - Loads SQLite at runtime (no compile-time dependency)
- **Static Library Support** - Can optionally use statically linked SQLite
- **File-Based Connection** - Connects to SQLite database files (not network)
- **Basic Query Execution** - Execute SQL queries and retrieve results
- **Prepared Statements** - Support for parameterized queries
- **Transaction Support** - BEGIN, COMMIT, ROLLBACK

## Architecture

Following UniDAC's SQLite implementation:
- Dynamic loading using `dlopen`/`LoadLibraryEx`
- Function pointer resolution using `dlsym`/`GetProcAddress`
- No compile-time dependency on SQLite headers
- Supports both static and dynamic modes

## Building

### Static Linking (Default - Recommended)

Static linking embeds SQLite directly into the executable using pre-compiled object files from UniDAC:

```bash
mkdir build && cd build
cmake -DUSE_STATIC_SQLITE=ON ..
make
```

Or simply (static is default):
```bash
cmake ..
make
```

### Dynamic Loading

To use dynamic loading instead (loads SQLite at runtime):

```bash
cmake -DUSE_STATIC_SQLITE=OFF ..
make
```

See [STATIC_LINKING.md](STATIC_LINKING.md) for more details.

## Usage

```cpp
#include "sqlite_client.h"

SQLiteClient client;

// Connect to database file
if (client.connect("test.db")) {
    // Execute query
    auto result = client.executeQuery("SELECT * FROM users");
    
    while (result->fetch()) {
        std::cout << result->getValue(0).asString() << std::endl;
    }
    
    client.disconnect();
}
```

## Dependencies

### Static Linking Mode (Default)
- **Compile-time**: SQLite object files from `../Source/sqlite3/` (automatically selected by CMake)
- **Runtime**: None (SQLite embedded in executable)

### Dynamic Loading Mode
- **Compile-time**: None
- **Runtime**: SQLite library (`libsqlite3.so`, `libsqlite3.dylib`, or `sqlite3.dll`)
