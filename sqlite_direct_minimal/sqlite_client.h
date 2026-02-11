#pragma once
#include <string>
#include <vector>
#include <memory>
#include <cstdint>

// Forward declarations
class SQLiteResultSet;
class SQLitePreparedStatement;

// SQLite data types
enum class SQLiteDataType {
    INTEGER = 1,
    FLOAT = 2,
    TEXT = 3,
    BLOB = 4,
    NULL_TYPE = 5
};

// Row value structure
struct SQLiteRowValue {
    bool is_null;
    std::vector<uint8_t> data;
    SQLiteDataType data_type;
    
    SQLiteRowValue() : is_null(true), data_type(SQLiteDataType::NULL_TYPE) {}
    
    std::string asString() const;
    int64_t asInt64() const;
    double asDouble() const;
    const std::vector<uint8_t>& asBlob() const;
};

// Result set interface
class SQLiteResultSet {
public:
    virtual ~SQLiteResultSet() = default;
    virtual bool fetch() = 0;
    virtual size_t columnCount() const = 0;
    virtual std::string columnName(size_t index) const = 0;
    virtual SQLiteDataType columnType(size_t index) const = 0;
    virtual SQLiteRowValue getValue(size_t index) const = 0;
    virtual SQLiteRowValue getValue(const std::string& column_name) const = 0;
};

// Prepared statement interface
class SQLitePreparedStatement {
public:
    virtual ~SQLitePreparedStatement() = default;
    virtual bool bindInt(int index, int64_t value) = 0;
    virtual bool bindDouble(int index, double value) = 0;
    virtual bool bindText(int index, const std::string& value) = 0;
    virtual bool bindBlob(int index, const std::vector<uint8_t>& value) = 0;
    virtual bool bindNull(int index) = 0;
    virtual std::unique_ptr<SQLiteResultSet> executeQuery() = 0;
    virtual bool executeUpdate() = 0;
    virtual void reset() = 0;
};

// Main SQLite client class
class SQLiteClient {
public:
    SQLiteClient();
    ~SQLiteClient();
    
    // Connection management
    bool connect(const std::string& database_path, bool read_only = false, bool create_if_missing = false);
    void disconnect();
    bool isConnected() const;
    
    // Query execution
    std::unique_ptr<SQLiteResultSet> executeQuery(const std::string& sql);
    bool executeUpdate(const std::string& sql);
    int64_t lastInsertRowId() const;
    int changes() const;
    
    // Prepared statements
    std::unique_ptr<SQLitePreparedStatement> prepareStatement(const std::string& sql);
    
    // Transaction management
    bool beginTransaction();
    bool commit();
    bool rollback();
    bool isInTransaction() const;
    
    // Utility
    std::string getLastError() const;
    int getLastErrorCode() const;
    
    // Friend classes for accessing function pointers
    friend class SQLiteResultSetImpl;
    friend class SQLitePreparedStatementImpl;
    
private:
    // Function pointer types (must be defined before accessor methods)
    typedef int (*sqlite3_open_v2_func)(const char*, void**, int, const char*);
    typedef int (*sqlite3_close_func)(void*);
    typedef int (*sqlite3_exec_func)(void*, const char*, int (*)(void*, int, char**, char**), void*, char**);
    typedef int (*sqlite3_prepare_v2_func)(void*, const char*, int, void**, const char**);
    typedef int (*sqlite3_step_func)(void*);
    typedef int (*sqlite3_reset_func)(void*);
    typedef int (*sqlite3_finalize_func)(void*);
    typedef int (*sqlite3_column_count_func)(void*);
    typedef const char* (*sqlite3_column_name_func)(void*, int);
    typedef int (*sqlite3_column_type_func)(void*, int);
    typedef const void* (*sqlite3_column_blob_func)(void*, int);
    typedef int (*sqlite3_column_bytes_func)(void*, int);
    typedef int64_t (*sqlite3_column_int64_func)(void*, int);
    typedef double (*sqlite3_column_double_func)(void*, int);
    typedef const unsigned char* (*sqlite3_column_text_func)(void*, int);
    typedef int (*sqlite3_bind_int64_func)(void*, int, int64_t);
    typedef int (*sqlite3_bind_double_func)(void*, int, double);
    typedef int (*sqlite3_bind_text_func)(void*, int, const char*, int, void*);
    typedef int (*sqlite3_bind_blob_func)(void*, int, const void*, int, void*);
    typedef int (*sqlite3_bind_null_func)(void*, int);
    typedef int64_t (*sqlite3_last_insert_rowid_func)(void*);
    typedef int (*sqlite3_changes_func)(void*);
    typedef const char* (*sqlite3_errmsg_func)(void*);
    typedef int (*sqlite3_errcode_func)(void*);
    typedef int (*sqlite3_get_autocommit_func)(void*);
    
    void* db_handle_;  // Opaque SQLite database handle
    bool connected_;
    bool in_transaction_;
    
    // Library loading
    bool loadSQLiteLibrary();
    void unloadSQLiteLibrary();
    void* sqlite_lib_handle_;
    
    // Accessors for function pointers (for friend classes)
    sqlite3_step_func getStepFunc() const { return sqlite3_step_; }
    sqlite3_column_count_func getColumnCountFunc() const { return sqlite3_column_count_; }
    sqlite3_column_name_func getColumnNameFunc() const { return sqlite3_column_name_; }
    sqlite3_column_type_func getColumnTypeFunc() const { return sqlite3_column_type_; }
    sqlite3_column_blob_func getColumnBlobFunc() const { return sqlite3_column_blob_; }
    sqlite3_column_bytes_func getColumnBytesFunc() const { return sqlite3_column_bytes_; }
    sqlite3_column_int64_func getColumnInt64Func() const { return sqlite3_column_int64_; }
    sqlite3_column_double_func getColumnDoubleFunc() const { return sqlite3_column_double_; }
    sqlite3_column_text_func getColumnTextFunc() const { return sqlite3_column_text_; }
    sqlite3_bind_int64_func getBindInt64Func() const { return sqlite3_bind_int64_; }
    sqlite3_bind_double_func getBindDoubleFunc() const { return sqlite3_bind_double_; }
    sqlite3_bind_text_func getBindTextFunc() const { return sqlite3_bind_text_; }
    sqlite3_bind_blob_func getBindBlobFunc() const { return sqlite3_bind_blob_; }
    sqlite3_bind_null_func getBindNullFunc() const { return sqlite3_bind_null_; }
    sqlite3_reset_func getResetFunc() const { return sqlite3_reset_; }
    sqlite3_finalize_func getFinalizeFunc() const { return sqlite3_finalize_; }
    
    sqlite3_open_v2_func sqlite3_open_v2_;
    sqlite3_close_func sqlite3_close_;
    sqlite3_exec_func sqlite3_exec_;
    sqlite3_prepare_v2_func sqlite3_prepare_v2_;
    sqlite3_step_func sqlite3_step_;
    sqlite3_reset_func sqlite3_reset_;
    sqlite3_finalize_func sqlite3_finalize_;
    sqlite3_column_count_func sqlite3_column_count_;
    sqlite3_column_name_func sqlite3_column_name_;
    sqlite3_column_type_func sqlite3_column_type_;
    sqlite3_column_blob_func sqlite3_column_blob_;
    sqlite3_column_bytes_func sqlite3_column_bytes_;
    sqlite3_column_int64_func sqlite3_column_int64_;
    sqlite3_column_double_func sqlite3_column_double_;
    sqlite3_column_text_func sqlite3_column_text_;
    sqlite3_bind_int64_func sqlite3_bind_int64_;
    sqlite3_bind_double_func sqlite3_bind_double_;
    sqlite3_bind_text_func sqlite3_bind_text_;
    sqlite3_bind_blob_func sqlite3_bind_blob_;
    sqlite3_bind_null_func sqlite3_bind_null_;
    sqlite3_last_insert_rowid_func sqlite3_last_insert_rowid_;
    sqlite3_changes_func sqlite3_changes_;
    sqlite3_errmsg_func sqlite3_errmsg_;
    sqlite3_errcode_func sqlite3_errcode_;
    sqlite3_get_autocommit_func sqlite3_get_autocommit_;
};
