// Basic Usage Example
// Demonstrates simple SQLite connection and query execution

#include "../sqlite_client.h"
#include <iostream>

int main() {
    SQLiteClient client;
    
    // Connect to SQLite database file
    std::string db_path = "test.db";
    
    std::cout << "Connecting to SQLite database: " << db_path << std::endl;
    
    if (!client.connect(db_path, false, true)) {  // read_only=false, create_if_missing=true
        std::cerr << "Failed to connect: " << client.getLastError() << std::endl;
        return 1;
    }
    
    std::cout << "Connected successfully!" << std::endl;
    
    // Create a test table
    std::cout << "\nCreating test table..." << std::endl;
    client.executeUpdate(
        "CREATE TABLE IF NOT EXISTS users ("
        "id INTEGER PRIMARY KEY AUTOINCREMENT,"
        "username TEXT NOT NULL,"
        "email TEXT NOT NULL,"
        "age INTEGER"
        ")"
    );
    
    // Insert some data
    std::cout << "Inserting test data..." << std::endl;
    client.executeUpdate("INSERT INTO users (username, email, age) VALUES ('alice', 'alice@example.com', 30)");
    client.executeUpdate("INSERT INTO users (username, email, age) VALUES ('bob', 'bob@example.com', 25)");
    client.executeUpdate("INSERT INTO users (username, email, age) VALUES ('charlie', 'charlie@example.com', 35)");
    
    std::cout << "Last insert row ID: " << client.lastInsertRowId() << std::endl;
    
    // Query data
    std::cout << "\nQuerying users table..." << std::endl;
    auto result = client.executeQuery("SELECT id, username, email, age FROM users");
    
    if (result) {
        std::cout << "\nResults:" << std::endl;
        std::cout << "Columns: " << result->columnCount() << std::endl;
        
        // Print column names
        for (size_t i = 0; i < result->columnCount(); i++) {
            std::cout << result->columnName(i);
            if (i < result->columnCount() - 1) std::cout << " | ";
        }
        std::cout << std::endl;
        std::cout << std::string(50, '-') << std::endl;
        
        // Fetch and print rows
        int row_count = 0;
        while (result->fetch()) {
            row_count++;
            for (size_t i = 0; i < result->columnCount(); i++) {
                auto value = result->getValue(i);
                if (value.is_null) {
                    std::cout << "NULL";
                } else {
                    std::cout << value.asString();
                }
                if (i < result->columnCount() - 1) std::cout << " | ";
            }
            std::cout << std::endl;
        }
        std::cout << "\nTotal rows: " << row_count << std::endl;
    }
    
    // Test prepared statement
    std::cout << "\nTesting prepared statement..." << std::endl;
    auto stmt = client.prepareStatement("INSERT INTO users (username, email, age) VALUES (?, ?, ?)");
    if (stmt) {
        stmt->bindText(1, "david");
        stmt->bindText(2, "david@example.com");
        stmt->bindInt(3, 28);
        if (stmt->executeUpdate()) {
            std::cout << "Inserted via prepared statement, row ID: " << client.lastInsertRowId() << std::endl;
        }
    }
    
    // Test transaction
    std::cout << "\nTesting transaction..." << std::endl;
    if (client.beginTransaction()) {
        client.executeUpdate("INSERT INTO users (username, email, age) VALUES ('eve', 'eve@example.com', 32)");
        client.executeUpdate("INSERT INTO users (username, email, age) VALUES ('frank', 'frank@example.com', 27)");
        
        std::cout << "Changes in transaction: " << client.changes() << std::endl;
        
        if (client.commit()) {
            std::cout << "Transaction committed successfully" << std::endl;
        } else {
            std::cerr << "Failed to commit transaction" << std::endl;
        }
    }
    
    // Final count
    auto count_result = client.executeQuery("SELECT COUNT(*) FROM users");
    if (count_result && count_result->fetch()) {
        std::cout << "\nTotal users in database: " << count_result->getValue(0).asInt64() << std::endl;
    }
    
    client.disconnect();
    std::cout << "\nDisconnected" << std::endl;
    
    return 0;
}
