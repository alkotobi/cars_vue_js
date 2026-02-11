#include "sqlite_resultset.h"
#include "sqlite_client.h"
#include <cstring>

#define SQLITE_ROW  100
#define SQLITE_DONE 101

SQLiteResultSetImpl::SQLiteResultSetImpl(void* stmt, SQLiteClient* client)
    : stmt_(stmt)
    , client_(client)
    , has_row_(false)
    , current_row_(0)
{
    // Get function pointers from client
    if (client_) {
        sqlite3_step_ = client_->getStepFunc();
        sqlite3_column_count_ = client_->getColumnCountFunc();
        sqlite3_column_name_ = client_->getColumnNameFunc();
        sqlite3_column_type_ = client_->getColumnTypeFunc();
        sqlite3_column_blob_ = client_->getColumnBlobFunc();
        sqlite3_column_bytes_ = client_->getColumnBytesFunc();
        sqlite3_column_int64_ = client_->getColumnInt64Func();
        sqlite3_column_double_ = client_->getColumnDoubleFunc();
        sqlite3_column_text_ = client_->getColumnTextFunc();
    }
    
    // Fetch first row
    if (stmt_ && sqlite3_step_) {
        int result = sqlite3_step_(stmt_);
        has_row_ = (result == SQLITE_ROW);
    }
}

SQLiteResultSetImpl::~SQLiteResultSetImpl() {
    if (stmt_ && client_) {
        auto finalize = client_->getFinalizeFunc();
        if (finalize) {
            finalize(stmt_);
        }
    }
}

bool SQLiteResultSetImpl::fetch() {
    if (!stmt_ || !sqlite3_step_) {
        return false;
    }
    
    if (has_row_) {
        // Already have a row, fetch next
        int result = sqlite3_step_(stmt_);
        has_row_ = (result == SQLITE_ROW);
        if (has_row_) {
            current_row_++;
        }
        return has_row_;
    }
    
    return false;
}

size_t SQLiteResultSetImpl::columnCount() const {
    if (!stmt_ || !sqlite3_column_count_) {
        return 0;
    }
    return static_cast<size_t>(sqlite3_column_count_(stmt_));
}

std::string SQLiteResultSetImpl::columnName(size_t index) const {
    if (!stmt_ || !sqlite3_column_name_) {
        return "";
    }
    const char* name = sqlite3_column_name_(stmt_, static_cast<int>(index));
    return name ? std::string(name) : "";
}

SQLiteDataType SQLiteResultSetImpl::columnType(size_t index) const {
    if (!stmt_ || !sqlite3_column_type_) {
        return SQLiteDataType::NULL_TYPE;
    }
    int type = sqlite3_column_type_(stmt_, static_cast<int>(index));
    return static_cast<SQLiteDataType>(type);
}

SQLiteRowValue SQLiteResultSetImpl::getValue(size_t index) const {
    SQLiteRowValue value;
    
    if (!stmt_ || !has_row_) {
        return value;
    }
    
    if (!sqlite3_column_type_) {
        return value;
    }
    
    int type = sqlite3_column_type_(stmt_, static_cast<int>(index));
    
    if (type == 5) { // SQLITE_NULL
        value.is_null = true;
        value.data_type = SQLiteDataType::NULL_TYPE;
        return value;
    }
    
    value.is_null = false;
    value.data_type = static_cast<SQLiteDataType>(type);
    
    switch (type) {
        case 1: { // SQLITE_INTEGER
            if (sqlite3_column_int64_) {
                int64_t int_val = sqlite3_column_int64_(stmt_, static_cast<int>(index));
                value.data.resize(sizeof(int64_t));
                std::memcpy(value.data.data(), &int_val, sizeof(int64_t));
            }
            break;
        }
        case 2: { // SQLITE_FLOAT
            if (sqlite3_column_double_) {
                double double_val = sqlite3_column_double_(stmt_, static_cast<int>(index));
                value.data.resize(sizeof(double));
                std::memcpy(value.data.data(), &double_val, sizeof(double));
            }
            break;
        }
        case 3: { // SQLITE_TEXT
            if (sqlite3_column_text_) {
                const unsigned char* text = sqlite3_column_text_(stmt_, static_cast<int>(index));
                if (sqlite3_column_bytes_) {
                    int bytes = sqlite3_column_bytes_(stmt_, static_cast<int>(index));
                    if (text && bytes > 0) {
                        value.data.resize(bytes);
                        std::memcpy(value.data.data(), text, bytes);
                    }
                }
            }
            break;
        }
        case 4: { // SQLITE_BLOB
            if (sqlite3_column_blob_) {
                const void* blob = sqlite3_column_blob_(stmt_, static_cast<int>(index));
                if (sqlite3_column_bytes_) {
                    int bytes = sqlite3_column_bytes_(stmt_, static_cast<int>(index));
                    if (blob && bytes > 0) {
                        value.data.resize(bytes);
                        std::memcpy(value.data.data(), blob, bytes);
                    }
                }
            }
            break;
        }
    }
    
    return value;
}

SQLiteRowValue SQLiteResultSetImpl::getValue(const std::string& column_name) const {
    size_t count = columnCount();
    for (size_t i = 0; i < count; i++) {
        if (columnName(i) == column_name) {
            return getValue(i);
        }
    }
    return SQLiteRowValue();
}
