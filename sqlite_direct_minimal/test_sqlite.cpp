// Test SQLite direct connection
#include "sqlite_client.h"
#include <iostream>
#include <cstdio>

int main() {
    std::cout << "SQLite Direct Client Test" << std::endl;
    std::cout << "Using dynamic loading (no compile-time dependency)" << std::endl;
    
    // Remove test database if exists
    std::remove("test_sqlite.db");
    
    SQLiteClient client;
    
    // Connect to database
    std::cout << "\nConnecting to test_sqlite.db..." << std::endl;
    if (!client.connect("test_sqlite.db", false, true)) {
        std::cerr << "Failed to connect: " << client.getLastError() << std::endl;
        return 1;
    }
    
    std::cout << "✓ Connected successfully!" << std::endl;
    
    // Create table
    std::cout << "\nCreating test table..." << std::endl;
    if (!client.executeUpdate(
        "CREATE TABLE IF NOT EXISTS test_table ("
        "id INTEGER PRIMARY KEY AUTOINCREMENT,"
        "name TEXT NOT NULL,"
        "value INTEGER,"
        "data BLOB"
        ")"
    )) {
        std::cerr << "Failed to create table: " << client.getLastError() << std::endl;
        return 1;
    }
    std::cout << "✓ Table created" << std::endl;
    
    // Insert data
    std::cout << "\nInserting test data..." << std::endl;
    client.executeUpdate("INSERT INTO test_table (name, value) VALUES ('test1', 100)");
    client.executeUpdate("INSERT INTO test_table (name, value) VALUES ('test2', 200)");
    client.executeUpdate("INSERT INTO test_table (name, value) VALUES ('test3', 300)");
    std::cout << "✓ Inserted 3 rows" << std::endl;
    std::cout << "  Last insert row ID: " << client.lastInsertRowId() << std::endl;
    
    // Query data
    std::cout << "\nQuerying data..." << std::endl;
    auto result = client.executeQuery("SELECT id, name, value FROM test_table");
    
    if (result) {
        std::cout << "✓ Query executed successfully" << std::endl;
        std::cout << "  Columns: " << result->columnCount() << std::endl;
        
        // Print column names
        std::cout << "\n  ";
        for (size_t i = 0; i < result->columnCount(); i++) {
            std::cout << result->columnName(i);
            if (i < result->columnCount() - 1) std::cout << " | ";
        }
        std::cout << std::endl;
        std::cout << "  " << std::string(30, '-') << std::endl;
        
        // Fetch and print rows
        int row_count = 0;
        while (result->fetch()) {
            row_count++;
            std::cout << "  ";
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
        std::cout << "\n  Total rows: " << row_count << std::endl;
    } else {
        std::cerr << "✗ Query failed: " << client.getLastError() << std::endl;
    }
    
    // Test prepared statement
    std::cout << "\nTesting prepared statement..." << std::endl;
    auto stmt = client.prepareStatement("INSERT INTO test_table (name, value) VALUES (?, ?)");
    if (stmt) {
        stmt->bindText(1, "prepared1");
        stmt->bindInt(2, 400);
        if (stmt->executeUpdate()) {
            std::cout << "✓ Prepared statement executed" << std::endl;
            std::cout << "  Last insert row ID: " << client.lastInsertRowId() << std::endl;
        }
    }
    
    // Test transaction
    std::cout << "\nTesting transaction..." << std::endl;
    if (client.beginTransaction()) {
        std::cout << "  Transaction started" << std::endl;
        client.executeUpdate("INSERT INTO test_table (name, value) VALUES ('tx1', 500)");
        client.executeUpdate("INSERT INTO test_table (name, value) VALUES ('tx2', 600)");
        std::cout << "  Changes: " << client.changes() << std::endl;
        
        if (client.commit()) {
            std::cout << "✓ Transaction committed" << std::endl;
        } else {
            std::cerr << "✗ Failed to commit" << std::endl;
        }
    }
    
    // Final count
    auto count_result = client.executeQuery("SELECT COUNT(*) FROM test_table");
    if (count_result && count_result->fetch()) {
        std::cout << "\n✓ Total rows in database: " << count_result->getValue(0).asInt64() << std::endl;
    }
    
    client.disconnect();
    std::cout << "\n✓ Disconnected" << std::endl;
    
    return 0;
}
