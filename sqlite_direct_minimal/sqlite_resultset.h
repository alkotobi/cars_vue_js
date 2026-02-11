#pragma once
#include "sqlite_client.h"
#include <string>
#include <vector>
#include <memory>
#include <cstdint>

// SQLite result set implementation
class SQLiteResultSetImpl : public SQLiteResultSet {
public:
    SQLiteResultSetImpl(void* stmt, SQLiteClient* client);
    ~SQLiteResultSetImpl();
    
    bool fetch() override;
    size_t columnCount() const override;
    std::string columnName(size_t index) const override;
    SQLiteDataType columnType(size_t index) const override;
    SQLiteRowValue getValue(size_t index) const override;
    SQLiteRowValue getValue(const std::string& column_name) const override;
    
private:
    void* stmt_;  // SQLite statement handle
    SQLiteClient* client_;
    bool has_row_;
    int current_row_;
    
    // Function pointers from client
    typedef int (*sqlite3_step_func)(void*);
    typedef int (*sqlite3_column_count_func)(void*);
    typedef const char* (*sqlite3_column_name_func)(void*, int);
    typedef int (*sqlite3_column_type_func)(void*, int);
    typedef const void* (*sqlite3_column_blob_func)(void*, int);
    typedef int (*sqlite3_column_bytes_func)(void*, int);
    typedef int64_t (*sqlite3_column_int64_func)(void*, int);
    typedef double (*sqlite3_column_double_func)(void*, int);
    typedef const unsigned char* (*sqlite3_column_text_func)(void*, int);
    
    sqlite3_step_func sqlite3_step_;
    sqlite3_column_count_func sqlite3_column_count_;
    sqlite3_column_name_func sqlite3_column_name_;
    sqlite3_column_type_func sqlite3_column_type_;
    sqlite3_column_blob_func sqlite3_column_blob_;
    sqlite3_column_bytes_func sqlite3_column_bytes_;
    sqlite3_column_int64_func sqlite3_column_int64_;
    sqlite3_column_double_func sqlite3_column_double_;
    sqlite3_column_text_func sqlite3_column_text_;
};
