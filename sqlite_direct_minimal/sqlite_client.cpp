#include "sqlite_client.h"
#include "sqlite_resultset.h"
#include "sqlite_prepared_statement.h"
#include <iostream>
#include <cstring>
#include <cstdlib>

#ifdef _WIN32
#include <windows.h>
#else
#include <dlfcn.h>
#endif

// SQLite constants
#define SQLITE_OK           0
#define SQLITE_ERROR        1
#define SQLITE_ROW          100
#define SQLITE_DONE         101
#define SQLITE_OPEN_READONLY        0x00000001
#define SQLITE_OPEN_READWRITE       0x00000002
#define SQLITE_OPEN_CREATE          0x00000004

// Forward declarations for static linking
#ifdef SQLITE_STATIC
extern "C" {
    int sqlite3_open_v2(const char*, void**, int, const char*);
    int sqlite3_close(void*);
    int sqlite3_exec(void*, const char*, int (*)(void*, int, char**, char**), void*, char**);
    int sqlite3_prepare_v2(void*, const char*, int, void**, const char**);
    int sqlite3_step(void*);
    int sqlite3_reset(void*);
    int sqlite3_finalize(void*);
    int sqlite3_column_count(void*);
    const char* sqlite3_column_name(void*, int);
    int sqlite3_column_type(void*, int);
    const void* sqlite3_column_blob(void*, int);
    int sqlite3_column_bytes(void*, int);
    int64_t sqlite3_column_int64(void*, int);
    double sqlite3_column_double(void*, int);
    const unsigned char* sqlite3_column_text(void*, int);
    int sqlite3_bind_int64(void*, int, int64_t);
    int sqlite3_bind_double(void*, int, double);
    int sqlite3_bind_text(void*, int, const char*, int, void*);
    int sqlite3_bind_blob(void*, int, const void*, int, void*);
    int sqlite3_bind_null(void*, int);
    int64_t sqlite3_last_insert_rowid(void*);
    int sqlite3_changes(void*);
    const char* sqlite3_errmsg(void*);
    int sqlite3_errcode(void*);
    int sqlite3_get_autocommit(void*);
}
#endif // SQLITE_STATIC

SQLiteClient::SQLiteClient()
    : db_handle_(nullptr)
    , connected_(false)
    , in_transaction_(false)
    , sqlite_lib_handle_(nullptr)
    , sqlite3_open_v2_(nullptr)
    , sqlite3_close_(nullptr)
    , sqlite3_exec_(nullptr)
    , sqlite3_prepare_v2_(nullptr)
    , sqlite3_step_(nullptr)
    , sqlite3_reset_(nullptr)
    , sqlite3_finalize_(nullptr)
    , sqlite3_column_count_(nullptr)
    , sqlite3_column_name_(nullptr)
    , sqlite3_column_type_(nullptr)
    , sqlite3_column_blob_(nullptr)
    , sqlite3_column_bytes_(nullptr)
    , sqlite3_column_int64_(nullptr)
    , sqlite3_column_double_(nullptr)
    , sqlite3_column_text_(nullptr)
    , sqlite3_bind_int64_(nullptr)
    , sqlite3_bind_double_(nullptr)
    , sqlite3_bind_text_(nullptr)
    , sqlite3_bind_blob_(nullptr)
    , sqlite3_bind_null_(nullptr)
    , sqlite3_last_insert_rowid_(nullptr)
    , sqlite3_changes_(nullptr)
    , sqlite3_errmsg_(nullptr)
    , sqlite3_errcode_(nullptr)
    , sqlite3_get_autocommit_(nullptr)
{
    loadSQLiteLibrary();
}

SQLiteClient::~SQLiteClient() {
    disconnect();
    unloadSQLiteLibrary();
}

bool SQLiteClient::loadSQLiteLibrary() {
    if (sqlite_lib_handle_ != nullptr) {
        return true; // Already loaded
    }
    
#ifdef SQLITE_STATIC
    // Static linking mode - directly assign function pointers to SQLite functions
    // SQLite functions are linked directly, so we can get their addresses
    // (Forward declarations are at the top of the file)
    
    // Assign function pointers directly (static linking)
    sqlite3_open_v2_ = sqlite3_open_v2;
    sqlite3_close_ = sqlite3_close;
    sqlite3_exec_ = sqlite3_exec;
    sqlite3_prepare_v2_ = sqlite3_prepare_v2;
    sqlite3_step_ = sqlite3_step;
    sqlite3_reset_ = sqlite3_reset;
    sqlite3_finalize_ = sqlite3_finalize;
    sqlite3_column_count_ = sqlite3_column_count;
    sqlite3_column_name_ = sqlite3_column_name;
    sqlite3_column_type_ = sqlite3_column_type;
    sqlite3_column_blob_ = sqlite3_column_blob;
    sqlite3_column_bytes_ = sqlite3_column_bytes;
    sqlite3_column_int64_ = sqlite3_column_int64;
    sqlite3_column_double_ = sqlite3_column_double;
    sqlite3_column_text_ = sqlite3_column_text;
    sqlite3_bind_int64_ = sqlite3_bind_int64;
    sqlite3_bind_double_ = sqlite3_bind_double;
    sqlite3_bind_text_ = sqlite3_bind_text;
    sqlite3_bind_blob_ = sqlite3_bind_blob;
    sqlite3_bind_null_ = sqlite3_bind_null;
    sqlite3_last_insert_rowid_ = sqlite3_last_insert_rowid;
    sqlite3_changes_ = sqlite3_changes;
    sqlite3_errmsg_ = sqlite3_errmsg;
    sqlite3_errcode_ = sqlite3_errcode;
    sqlite3_get_autocommit_ = sqlite3_get_autocommit;
    
    // Mark as loaded (use a dummy value for static mode)
    sqlite_lib_handle_ = reinterpret_cast<void*>(1);
    
    return true;
#else
    // Dynamic loading mode
#ifdef _WIN32
    // Try common SQLite DLL names
    const char* lib_names[] = {
        "sqlite3.dll",
        "libsqlite3.dll",
        nullptr
    };
    
    for (int i = 0; lib_names[i] != nullptr; i++) {
        sqlite_lib_handle_ = LoadLibraryExA(lib_names[i], nullptr, LOAD_WITH_ALTERED_SEARCH_PATH);
        if (sqlite_lib_handle_ != nullptr) {
            break;
        }
    }
    
    if (sqlite_lib_handle_ == nullptr) {
        std::cerr << "Failed to load SQLite library" << std::endl;
        return false;
    }
    
    auto getProc = [this](const char* name) -> void* {
        return GetProcAddress((HMODULE)sqlite_lib_handle_, name);
    };
#else
    // Try common SQLite library names
#ifdef __APPLE__
    const char* lib_names[] = {
        "/opt/homebrew/lib/libsqlite3.dylib",
        "/usr/local/lib/libsqlite3.dylib",
        "/usr/lib/libsqlite3.dylib",
        "libsqlite3.dylib",
        nullptr
    };
#else
    const char* lib_names[] = {
        "libsqlite3.so.0",
        "libsqlite3.so",
        nullptr
    };
#endif
    
    for (int i = 0; lib_names[i] != nullptr; i++) {
        sqlite_lib_handle_ = dlopen(lib_names[i], RTLD_LAZY);
        if (sqlite_lib_handle_ != nullptr) {
            break;
        }
    }
    
    if (sqlite_lib_handle_ == nullptr) {
        const char* err = dlerror();
        std::cerr << "Failed to load SQLite library: " << (err ? err : "unknown error") << std::endl;
        return false;
    }
    
    auto getProc = [this](const char* name) -> void* {
        return dlsym(sqlite_lib_handle_, name);
    };
#endif
    
    // Load all required functions
    sqlite3_open_v2_ = (sqlite3_open_v2_func)getProc("sqlite3_open_v2");
    sqlite3_close_ = (sqlite3_close_func)getProc("sqlite3_close");
    sqlite3_exec_ = (sqlite3_exec_func)getProc("sqlite3_exec");
    sqlite3_prepare_v2_ = (sqlite3_prepare_v2_func)getProc("sqlite3_prepare_v2");
    sqlite3_step_ = (sqlite3_step_func)getProc("sqlite3_step");
    sqlite3_reset_ = (sqlite3_reset_func)getProc("sqlite3_reset");
    sqlite3_finalize_ = (sqlite3_finalize_func)getProc("sqlite3_finalize");
    sqlite3_column_count_ = (sqlite3_column_count_func)getProc("sqlite3_column_count");
    sqlite3_column_name_ = (sqlite3_column_name_func)getProc("sqlite3_column_name");
    sqlite3_column_type_ = (sqlite3_column_type_func)getProc("sqlite3_column_type");
    sqlite3_column_blob_ = (sqlite3_column_blob_func)getProc("sqlite3_column_blob");
    sqlite3_column_bytes_ = (sqlite3_column_bytes_func)getProc("sqlite3_column_bytes");
    sqlite3_column_int64_ = (sqlite3_column_int64_func)getProc("sqlite3_column_int64");
    sqlite3_column_double_ = (sqlite3_column_double_func)getProc("sqlite3_column_double");
    sqlite3_column_text_ = (sqlite3_column_text_func)getProc("sqlite3_column_text");
    sqlite3_bind_int64_ = (sqlite3_bind_int64_func)getProc("sqlite3_bind_int64");
    sqlite3_bind_double_ = (sqlite3_bind_double_func)getProc("sqlite3_bind_double");
    sqlite3_bind_text_ = (sqlite3_bind_text_func)getProc("sqlite3_bind_text");
    sqlite3_bind_blob_ = (sqlite3_bind_blob_func)getProc("sqlite3_bind_blob");
    sqlite3_bind_null_ = (sqlite3_bind_null_func)getProc("sqlite3_bind_null");
    sqlite3_last_insert_rowid_ = (sqlite3_last_insert_rowid_func)getProc("sqlite3_last_insert_rowid");
    sqlite3_changes_ = (sqlite3_changes_func)getProc("sqlite3_changes");
    sqlite3_errmsg_ = (sqlite3_errmsg_func)getProc("sqlite3_errmsg");
    sqlite3_errcode_ = (sqlite3_errcode_func)getProc("sqlite3_errcode");
    sqlite3_get_autocommit_ = (sqlite3_get_autocommit_func)getProc("sqlite3_get_autocommit");
    
    // Check if all critical functions are loaded
    if (!sqlite3_open_v2_ || !sqlite3_close_ || !sqlite3_prepare_v2_ || 
        !sqlite3_step_ || !sqlite3_finalize_ || !sqlite3_column_count_ ||
        !sqlite3_column_name_ || !sqlite3_column_type_) {
        std::cerr << "Failed to load required SQLite functions" << std::endl;
        unloadSQLiteLibrary();
        return false;
    }
    
    return true;
#endif // SQLITE_STATIC
}

void SQLiteClient::unloadSQLiteLibrary() {
#ifdef SQLITE_STATIC
    // Nothing to unload in static mode
    sqlite_lib_handle_ = nullptr;
#else
#ifdef _WIN32
    if (sqlite_lib_handle_ != nullptr) {
        FreeLibrary((HMODULE)sqlite_lib_handle_);
        sqlite_lib_handle_ = nullptr;
    }
#else
    if (sqlite_lib_handle_ != nullptr) {
        dlclose(sqlite_lib_handle_);
        sqlite_lib_handle_ = nullptr;
    }
#endif
#endif
}

bool SQLiteClient::connect(const std::string& database_path, bool read_only, bool create_if_missing) {
    if (connected_) {
        return true;
    }
    
    if (!sqlite3_open_v2_) {
        std::cerr << "SQLite library not loaded" << std::endl;
        return false;
    }
    
    int flags = read_only ? SQLITE_OPEN_READONLY : SQLITE_OPEN_READWRITE;
    if (create_if_missing) {
        flags |= SQLITE_OPEN_CREATE;
    }
    
    void* db = nullptr;
    int result = sqlite3_open_v2_(database_path.c_str(), &db, flags, nullptr);
    
    if (result != SQLITE_OK) {
        std::cerr << "Failed to open database: " << (sqlite3_errmsg_ ? sqlite3_errmsg_(db) : "unknown error") << std::endl;
        if (db) {
            sqlite3_close_(db);
        }
        return false;
    }
    
    db_handle_ = db;
    connected_ = true;
    in_transaction_ = false;
    
    return true;
}

void SQLiteClient::disconnect() {
    if (db_handle_ && sqlite3_close_) {
        sqlite3_close_(db_handle_);
        db_handle_ = nullptr;
    }
    connected_ = false;
    in_transaction_ = false;
}

bool SQLiteClient::isConnected() const {
    return connected_ && db_handle_ != nullptr;
}

std::unique_ptr<SQLiteResultSet> SQLiteClient::executeQuery(const std::string& sql) {
    if (!connected_ || !db_handle_) {
        return nullptr;
    }
    
    void* stmt = nullptr;
    const char* tail = nullptr;
    
    int result = sqlite3_prepare_v2_(db_handle_, sql.c_str(), static_cast<int>(sql.length()), &stmt, &tail);
    
    if (result != SQLITE_OK || !stmt) {
        std::cerr << "Failed to prepare statement: " << getLastError() << std::endl;
        return nullptr;
    }
    
    return std::make_unique<SQLiteResultSetImpl>(stmt, this);
}

bool SQLiteClient::executeUpdate(const std::string& sql) {
    if (!connected_ || !db_handle_) {
        return false;
    }
    
    void* stmt = nullptr;
    const char* tail = nullptr;
    
    int result = sqlite3_prepare_v2_(db_handle_, sql.c_str(), static_cast<int>(sql.length()), &stmt, &tail);
    
    if (result != SQLITE_OK || !stmt) {
        return false;
    }
    
    result = sqlite3_step_(stmt);
    bool success = (result == SQLITE_DONE);
    
    sqlite3_finalize_(stmt);
    
    return success;
}

int64_t SQLiteClient::lastInsertRowId() const {
    if (!connected_ || !db_handle_ || !sqlite3_last_insert_rowid_) {
        return 0;
    }
    return sqlite3_last_insert_rowid_(db_handle_);
}

int SQLiteClient::changes() const {
    if (!connected_ || !db_handle_ || !sqlite3_changes_) {
        return 0;
    }
    return sqlite3_changes_(db_handle_);
}

std::unique_ptr<SQLitePreparedStatement> SQLiteClient::prepareStatement(const std::string& sql) {
    if (!connected_ || !db_handle_) {
        return nullptr;
    }
    
    void* stmt = nullptr;
    const char* tail = nullptr;
    
    int result = sqlite3_prepare_v2_(db_handle_, sql.c_str(), static_cast<int>(sql.length()), &stmt, &tail);
    
    if (result != SQLITE_OK || !stmt) {
        return nullptr;
    }
    
    return std::make_unique<SQLitePreparedStatementImpl>(stmt, this);
}

bool SQLiteClient::beginTransaction() {
    if (in_transaction_) {
        return true;
    }
    
    if (executeUpdate("BEGIN TRANSACTION")) {
        in_transaction_ = true;
        return true;
    }
    return false;
}

bool SQLiteClient::commit() {
    if (!in_transaction_) {
        return false;
    }
    
    if (executeUpdate("COMMIT")) {
        in_transaction_ = false;
        return true;
    }
    return false;
}

bool SQLiteClient::rollback() {
    if (!in_transaction_) {
        return false;
    }
    
    if (executeUpdate("ROLLBACK")) {
        in_transaction_ = false;
        return true;
    }
    return false;
}

bool SQLiteClient::isInTransaction() const {
    return in_transaction_;
}

std::string SQLiteClient::getLastError() const {
    if (!db_handle_ || !sqlite3_errmsg_) {
        return "Not connected";
    }
    const char* err = sqlite3_errmsg_(db_handle_);
    return err ? std::string(err) : "Unknown error";
}

int SQLiteClient::getLastErrorCode() const {
    if (!db_handle_ || !sqlite3_errcode_) {
        return -1;
    }
    return sqlite3_errcode_(db_handle_);
}

// SQLiteRowValue implementation
std::string SQLiteRowValue::asString() const {
    if (is_null || data.empty()) {
        return "";
    }
    
    // Convert based on data type
    switch (data_type) {
        case SQLiteDataType::INTEGER: {
            if (data.size() == sizeof(int64_t)) {
                int64_t int_val;
                std::memcpy(&int_val, data.data(), sizeof(int64_t));
                return std::to_string(int_val);
            }
            break;
        }
        case SQLiteDataType::FLOAT: {
            if (data.size() == sizeof(double)) {
                double double_val;
                std::memcpy(&double_val, data.data(), sizeof(double));
                return std::to_string(double_val);
            }
            break;
        }
        case SQLiteDataType::TEXT:
        case SQLiteDataType::BLOB:
            return std::string(reinterpret_cast<const char*>(data.data()), data.size());
        default:
            return "";
    }
    
    return "";
}

int64_t SQLiteRowValue::asInt64() const {
    if (is_null) {
        return 0;
    }
    
    switch (data_type) {
        case SQLiteDataType::INTEGER: {
            if (data.size() == sizeof(int64_t)) {
                int64_t int_val;
                std::memcpy(&int_val, data.data(), sizeof(int64_t));
                return int_val;
            }
            break;
        }
        case SQLiteDataType::TEXT: {
            try {
                return std::stoll(asString());
            } catch (...) {
                return 0;
            }
        }
        case SQLiteDataType::FLOAT: {
            return static_cast<int64_t>(asDouble());
        }
        default:
            return 0;
    }
    
    return 0;
}

double SQLiteRowValue::asDouble() const {
    if (is_null) {
        return 0.0;
    }
    
    switch (data_type) {
        case SQLiteDataType::FLOAT: {
            if (data.size() == sizeof(double)) {
                double double_val;
                std::memcpy(&double_val, data.data(), sizeof(double));
                return double_val;
            }
            break;
        }
        case SQLiteDataType::INTEGER: {
            return static_cast<double>(asInt64());
        }
        case SQLiteDataType::TEXT: {
            try {
                return std::stod(asString());
            } catch (...) {
                return 0.0;
            }
        }
        default:
            return 0.0;
    }
    
    return 0.0;
}

const std::vector<uint8_t>& SQLiteRowValue::asBlob() const {
    return data;
}
