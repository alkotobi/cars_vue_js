// Example: Create a database with 100 tables, each with 1000 records
// Demonstrates static linking and bulk operations

#include "../sqlite_client.h"
#include <iostream>
#include <iomanip>
#include <chrono>
#include <cstdio>
#include <random>
#include <fstream>

int main() {
    std::cout << "SQLite Direct Client - Massive Database Creation Example" << std::endl;
    std::cout << "========================================================" << std::endl;
    std::cout << std::endl;
    
    // Remove existing database if it exists
    const std::string db_path = "massive_test.db";
    std::remove(db_path.c_str());
    
    SQLiteClient client;
    
    // Connect to database
    std::cout << "Connecting to database: " << db_path << std::endl;
    auto start_time = std::chrono::high_resolution_clock::now();
    
    if (!client.connect(db_path, false, true)) {  // read_only=false, create_if_missing=true
        std::cerr << "Failed to connect: " << client.getLastError() << std::endl;
        return 1;
    }
    
    auto connect_time = std::chrono::high_resolution_clock::now();
    auto connect_duration = std::chrono::duration_cast<std::chrono::milliseconds>(connect_time - start_time);
    std::cout << "✓ Connected in " << connect_duration.count() << " ms" << std::endl;
    std::cout << std::endl;
    
    // Configuration
    const int num_tables = 100;
    const int records_per_table = 1000;
    
    std::cout << "Configuration:" << std::endl;
    std::cout << "  Tables: " << num_tables << std::endl;
    std::cout << "  Records per table: " << records_per_table << std::endl;
    std::cout << "  Total records: " << (num_tables * records_per_table) << std::endl;
    std::cout << std::endl;
    
    // Random number generator for sample data
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis_int(1, 1000000);
    std::uniform_real_distribution<> dis_double(0.0, 10000.0);
    
    // Start transaction for better performance
    std::cout << "Starting transaction..." << std::endl;
    if (!client.beginTransaction()) {
        std::cerr << "Failed to begin transaction: " << client.getLastError() << std::endl;
        return 1;
    }
    
    auto table_start = std::chrono::high_resolution_clock::now();
    
    // Create tables and insert data
    for (int table_num = 1; table_num <= num_tables; table_num++) {
        std::string table_name = "table_" + std::to_string(table_num);
        
        // Create table
        std::string create_sql = 
            "CREATE TABLE " + table_name + " ("
            "id INTEGER PRIMARY KEY AUTOINCREMENT,"
            "name TEXT NOT NULL,"
            "value INTEGER,"
            "price REAL,"
            "description TEXT,"
            "created_at TEXT DEFAULT (datetime('now'))"
            ")";
        
        if (!client.executeUpdate(create_sql)) {
            std::cerr << "Failed to create table " << table_name << ": " << client.getLastError() << std::endl;
            client.rollback();
            return 1;
        }
        
        // Insert records using prepared statement for better performance
        auto stmt = client.prepareStatement(
            "INSERT INTO " + table_name + " (name, value, price, description) VALUES (?, ?, ?, ?)"
        );
        
        if (!stmt) {
            std::cerr << "Failed to prepare statement for " << table_name << std::endl;
            client.rollback();
            return 1;
        }
        
        for (int record_num = 1; record_num <= records_per_table; record_num++) {
            std::string name = "item_" + std::to_string(table_num) + "_" + std::to_string(record_num);
            int value = dis_int(gen);
            double price = dis_double(gen);
            std::string description = "Description for " + name + " with value " + std::to_string(value);
            
            stmt->bindText(1, name);
            stmt->bindInt(2, value);
            stmt->bindDouble(3, price);
            stmt->bindText(4, description);
            
            if (!stmt->executeUpdate()) {
                std::cerr << "Failed to insert record " << record_num << " into " << table_name << std::endl;
                client.rollback();
                return 1;
            }
            
            stmt->reset();
        }
        
        // Progress indicator
        if (table_num % 10 == 0 || table_num == num_tables) {
            auto current_time = std::chrono::high_resolution_clock::now();
            auto elapsed = std::chrono::duration_cast<std::chrono::milliseconds>(current_time - table_start);
            double rate = (table_num * records_per_table) / (elapsed.count() / 1000.0);
            
            std::cout << "Progress: " << std::setw(3) << table_num << "/" << num_tables 
                      << " tables (" << std::setw(6) << (table_num * records_per_table) << " records)"
                      << " - " << std::fixed << std::setprecision(0) << rate << " records/sec" << std::endl;
        }
    }
    
    auto table_end = std::chrono::high_resolution_clock::now();
    auto table_duration = std::chrono::duration_cast<std::chrono::milliseconds>(table_end - table_start);
    
    // Commit transaction
    std::cout << std::endl << "Committing transaction..." << std::endl;
    if (!client.commit()) {
        std::cerr << "Failed to commit transaction: " << client.getLastError() << std::endl;
        return 1;
    }
    
    auto commit_time = std::chrono::high_resolution_clock::now();
    auto commit_duration = std::chrono::duration_cast<std::chrono::milliseconds>(commit_time - table_end);
    
    std::cout << "✓ Transaction committed in " << commit_duration.count() << " ms" << std::endl;
    std::cout << std::endl;
    
    // Verify data
    std::cout << "Verifying data..." << std::endl;
    
    // Check total tables
    auto count_result = client.executeQuery(
        "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name LIKE 'table_%'"
    );
    if (count_result && count_result->fetch()) {
        int table_count = static_cast<int>(count_result->getValue(0).asInt64());
        std::cout << "✓ Found " << table_count << " tables" << std::endl;
    }
    
    // Check total records in first table
    auto first_table_result = client.executeQuery("SELECT COUNT(*) FROM table_1");
    if (first_table_result && first_table_result->fetch()) {
        int record_count = static_cast<int>(first_table_result->getValue(0).asInt64());
        std::cout << "✓ Table 'table_1' has " << record_count << " records" << std::endl;
    }
    
    // Check total records in last table
    auto last_table_result = client.executeQuery("SELECT COUNT(*) FROM table_" + std::to_string(num_tables));
    if (last_table_result && last_table_result->fetch()) {
        int record_count = static_cast<int>(last_table_result->getValue(0).asInt64());
        std::cout << "✓ Table 'table_" << num_tables << "' has " << record_count << " records" << std::endl;
    }
    
    // Sample query from a random table
    std::cout << std::endl << "Sample query from table_50:" << std::endl;
    auto sample_result = client.executeQuery(
        "SELECT id, name, value, price FROM table_50 ORDER BY id LIMIT 5"
    );
    
    if (sample_result) {
        std::cout << std::left;
        std::cout << std::setw(6) << "ID" 
                  << std::setw(20) << "Name" 
                  << std::setw(10) << "Value" 
                  << std::setw(12) << "Price" << std::endl;
        std::cout << std::string(50, '-') << std::endl;
        
        while (sample_result->fetch()) {
            std::cout << std::setw(6) << sample_result->getValue(0).asInt64()
                      << std::setw(20) << sample_result->getValue(1).asString()
                      << std::setw(10) << sample_result->getValue(2).asInt64()
                      << std::fixed << std::setprecision(2) 
                      << std::setw(12) << sample_result->getValue(3).asDouble() << std::endl;
        }
    }
    
    // Statistics
    auto total_time = std::chrono::high_resolution_clock::now();
    auto total_duration = std::chrono::duration_cast<std::chrono::milliseconds>(total_time - start_time);
    
    std::cout << std::endl;
    std::cout << "========================================================" << std::endl;
    std::cout << "Statistics:" << std::endl;
    std::cout << "  Total time: " << total_duration.count() << " ms (" 
              << std::fixed << std::setprecision(2) << (total_duration.count() / 1000.0) << " seconds)" << std::endl;
    std::cout << "  Tables created: " << num_tables << std::endl;
    std::cout << "  Total records: " << (num_tables * records_per_table) << std::endl;
    std::cout << "  Average time per table: " << (table_duration.count() / num_tables) << " ms" << std::endl;
    std::cout << "  Average time per record: " << std::fixed << std::setprecision(3)
              << (table_duration.count() / (double)(num_tables * records_per_table)) << " ms" << std::endl;
    std::cout << "  Records per second: " << std::fixed << std::setprecision(0)
              << ((num_tables * records_per_table) / (total_duration.count() / 1000.0)) << std::endl;
    
    // Get database file size
    std::ifstream file(db_path, std::ios::binary | std::ios::ate);
    if (file.is_open()) {
        size_t file_size = file.tellg();
        file.close();
        std::cout << "  Database file size: " << std::fixed << std::setprecision(2)
                  << (file_size / 1024.0 / 1024.0) << " MB" << std::endl;
    }
    
    client.disconnect();
    std::cout << std::endl << "✓ Disconnected" << std::endl;
    
    return 0;
}
