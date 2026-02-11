#include "sqlite_prepared_statement.h"
#include "sqlite_client.h"
#include "sqlite_resultset.h"
#include <cstring>

SQLitePreparedStatementImpl::SQLitePreparedStatementImpl(void* stmt, SQLiteClient* client)
    : stmt_(stmt)
    , client_(client)
{
    // Get function pointers from client
    if (client_) {
        sqlite3_bind_int64_ = client_->getBindInt64Func();
        sqlite3_bind_double_ = client_->getBindDoubleFunc();
        sqlite3_bind_text_ = client_->getBindTextFunc();
        sqlite3_bind_blob_ = client_->getBindBlobFunc();
        sqlite3_bind_null_ = client_->getBindNullFunc();
        sqlite3_step_ = client_->getStepFunc();
        sqlite3_reset_ = client_->getResetFunc();
        sqlite3_finalize_ = client_->getFinalizeFunc();
    }
}

SQLitePreparedStatementImpl::~SQLitePreparedStatementImpl() {
    if (stmt_ && sqlite3_finalize_) {
        sqlite3_finalize_(stmt_);
    }
}

bool SQLitePreparedStatementImpl::bindInt(int index, int64_t value) {
    if (!stmt_ || !sqlite3_bind_int64_) {
        return false;
    }
    return sqlite3_bind_int64_(stmt_, index, value) == 0; // SQLITE_OK = 0
}

bool SQLitePreparedStatementImpl::bindDouble(int index, double value) {
    if (!stmt_ || !sqlite3_bind_double_) {
        return false;
    }
    return sqlite3_bind_double_(stmt_, index, value) == 0;
}

bool SQLitePreparedStatementImpl::bindText(int index, const std::string& value) {
    if (!stmt_ || !sqlite3_bind_text_) {
        return false;
    }
    // SQLITE_STATIC = 0, SQLITE_TRANSIENT = -1
    // We use SQLITE_TRANSIENT to let SQLite copy the string
    return sqlite3_bind_text_(stmt_, index, value.c_str(), static_cast<int>(value.length()), 
                              reinterpret_cast<void*>(-1)) == 0;
}

bool SQLitePreparedStatementImpl::bindBlob(int index, const std::vector<uint8_t>& value) {
    if (!stmt_ || !sqlite3_bind_blob_) {
        return false;
    }
    return sqlite3_bind_blob_(stmt_, index, value.data(), static_cast<int>(value.size()),
                              reinterpret_cast<void*>(-1)) == 0;
}

bool SQLitePreparedStatementImpl::bindNull(int index) {
    if (!stmt_ || !sqlite3_bind_null_) {
        return false;
    }
    return sqlite3_bind_null_(stmt_, index) == 0;
}

std::unique_ptr<SQLiteResultSet> SQLitePreparedStatementImpl::executeQuery() {
    if (!stmt_ || !sqlite3_step_) {
        return nullptr;
    }
    
    // Reset statement first
    if (sqlite3_reset_) {
        sqlite3_reset_(stmt_);
    }
    
    // Create result set (it will handle stepping)
    return std::make_unique<SQLiteResultSetImpl>(stmt_, client_);
}

bool SQLitePreparedStatementImpl::executeUpdate() {
    if (!stmt_ || !sqlite3_step_) {
        return false;
    }
    
    int result = sqlite3_step_(stmt_);
    return (result == 101); // SQLITE_DONE = 101
}

void SQLitePreparedStatementImpl::reset() {
    if (stmt_ && sqlite3_reset_) {
        sqlite3_reset_(stmt_);
    }
}
