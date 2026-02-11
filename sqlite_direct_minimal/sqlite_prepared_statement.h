#pragma once
#include "sqlite_client.h"
#include <string>
#include <vector>
#include <memory>
#include <cstdint>

// SQLite prepared statement implementation
class SQLitePreparedStatementImpl : public SQLitePreparedStatement {
public:
    SQLitePreparedStatementImpl(void* stmt, SQLiteClient* client);
    ~SQLitePreparedStatementImpl();
    
    bool bindInt(int index, int64_t value) override;
    bool bindDouble(int index, double value) override;
    bool bindText(int index, const std::string& value) override;
    bool bindBlob(int index, const std::vector<uint8_t>& value) override;
    bool bindNull(int index) override;
    std::unique_ptr<SQLiteResultSet> executeQuery() override;
    bool executeUpdate() override;
    void reset() override;
    
private:
    void* stmt_;  // SQLite statement handle
    SQLiteClient* client_;
    
    // Function pointers from client
    typedef int (*sqlite3_bind_int64_func)(void*, int, int64_t);
    typedef int (*sqlite3_bind_double_func)(void*, int, double);
    typedef int (*sqlite3_bind_text_func)(void*, int, const char*, int, void*);
    typedef int (*sqlite3_bind_blob_func)(void*, int, const void*, int, void*);
    typedef int (*sqlite3_bind_null_func)(void*, int);
    typedef int (*sqlite3_step_func)(void*);
    typedef int (*sqlite3_reset_func)(void*);
    typedef int (*sqlite3_finalize_func)(void*);
    
    sqlite3_bind_int64_func sqlite3_bind_int64_;
    sqlite3_bind_double_func sqlite3_bind_double_;
    sqlite3_bind_text_func sqlite3_bind_text_;
    sqlite3_bind_blob_func sqlite3_bind_blob_;
    sqlite3_bind_null_func sqlite3_bind_null_;
    sqlite3_step_func sqlite3_step_;
    sqlite3_reset_func sqlite3_reset_;
    sqlite3_finalize_func sqlite3_finalize_;
};
