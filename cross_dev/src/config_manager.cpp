#include "../include/config_manager.h"
#include "../include/platform.h"
#include <fstream>
#include <iostream>
#include <sstream>

#ifdef _WIN32
    #include <windows.h>
    #include <shlobj.h>
    #include <direct.h>
    #define mkdir(path, mode) _mkdir(path)
#elif __APPLE__
    #include <TargetConditionals.h>
    #if TARGET_OS_MAC
        #include <unistd.h>
        #include <sys/stat.h>
    #endif
#elif __linux__
    #include <unistd.h>
    #include <sys/stat.h>
    #include <pwd.h>
#endif

ConfigManager& ConfigManager::getInstance() {
    static ConfigManager instance;
    return instance;
}

std::string ConfigManager::getConfigDirectory() {
    std::string configDir;
    
#ifdef _WIN32
    // Windows: %APPDATA%\CrossDev
    char appDataPath[MAX_PATH];
    if (SHGetFolderPathA(NULL, CSIDL_APPDATA, NULL, SHGFP_TYPE_CURRENT, appDataPath) == S_OK) {
        configDir = std::string(appDataPath) + "\\CrossDev";
    } else {
        // Fallback to current directory
        configDir = ".\\CrossDev";
    }
#elif __APPLE__
    #if TARGET_OS_MAC
        // macOS: ~/Library/Application Support/CrossDev
        const char* home = getenv("HOME");
        if (home) {
            configDir = std::string(home) + "/Library/Application Support/CrossDev";
        } else {
            configDir = "./CrossDev";
        }
    #elif TARGET_OS_IPHONE
        // iOS: App's Documents directory (managed by iOS)
        // For iOS, we'll use a relative path that works in the app sandbox
        configDir = "./Documents/CrossDev";
    #endif
#elif __linux__
    // Linux: ~/.config/CrossDev
    const char* home = getenv("HOME");
    if (home) {
        configDir = std::string(home) + "/.config/CrossDev";
    } else {
        // Fallback for systems without HOME (shouldn't happen normally)
        struct passwd* pw = getpwuid(getuid());
        if (pw && pw->pw_dir) {
            configDir = std::string(pw->pw_dir) + "/.config/CrossDev";
        } else {
            configDir = "./CrossDev";
        }
    }
#else
    // Unknown platform - use current directory
    configDir = "./CrossDev";
#endif
    
    return configDir;
}

std::string ConfigManager::getOptionsFilePath() {
    return getConfigDirectory() + 
#ifdef _WIN32
           "\\options.json"
#else
           "/options.json"
#endif
           ;
}

bool ConfigManager::ensureConfigDirectory() {
    std::string configDir = getConfigDirectory();
    
#ifdef _WIN32
    // Create directory on Windows
    if (_mkdir(configDir.c_str()) != 0) {
        // Check if directory already exists
        DWORD dwAttrib = GetFileAttributesA(configDir.c_str());
        if (dwAttrib == INVALID_FILE_ATTRIBUTES || !(dwAttrib & FILE_ATTRIBUTE_DIRECTORY)) {
            std::cerr << "Failed to create config directory: " << configDir << std::endl;
            return false;
        }
    }
#else
    // Create directory on Unix-like systems
    struct stat info;
    if (stat(configDir.c_str(), &info) != 0) {
        // Directory doesn't exist, create it
        if (mkdir(configDir.c_str(), 0755) != 0) {
            std::cerr << "Failed to create config directory: " << configDir << std::endl;
            return false;
        }
    } else if (!(info.st_mode & S_IFDIR)) {
        // Path exists but is not a directory
        std::cerr << "Config path exists but is not a directory: " << configDir << std::endl;
        return false;
    }
#endif
    
    return true;
}

nlohmann::json ConfigManager::createDefaultOptions() {
    nlohmann::json defaultOptions;
    
    // HTML loading configuration
    defaultOptions["htmlLoading"] = nlohmann::json::object();
    defaultOptions["htmlLoading"]["method"] = "file";  // "file", "url", or "html"
    defaultOptions["htmlLoading"]["filePath"] = "demo.html";  // Path to HTML file
    defaultOptions["htmlLoading"]["url"] = "";  // URL to load
    defaultOptions["htmlLoading"]["htmlContent"] = "";  // HTML content as string
    
    return defaultOptions;
}

bool ConfigManager::loadOptions() {
    if (optionsLoaded_) {
        return true;
    }
    
    // Ensure config directory exists
    if (!ensureConfigDirectory()) {
        std::cerr << "Warning: Could not create config directory, using defaults" << std::endl;
        options_ = createDefaultOptions();
        optionsLoaded_ = true;
        return false;
    }
    
    std::string optionsPath = getOptionsFilePath();
    std::ifstream file(optionsPath);
    
    if (file.is_open()) {
        try {
            file >> options_;
            file.close();
            std::cout << "Loaded options from: " << optionsPath << std::endl;
            optionsLoaded_ = true;
            return true;
        } catch (const std::exception& e) {
            std::cerr << "Error parsing options.json: " << e.what() << std::endl;
            file.close();
        }
    } else {
        std::cout << "Options file not found, creating default: " << optionsPath << std::endl;
    }
    
    // Create default options
    options_ = createDefaultOptions();
    optionsLoaded_ = true;
    
    // Save default options to file
    saveOptions();
    
    return true;
}

bool ConfigManager::saveOptions() {
    if (!ensureConfigDirectory()) {
        std::cerr << "Error: Could not create config directory" << std::endl;
        return false;
    }
    
    std::string optionsPath = getOptionsFilePath();
    std::ofstream file(optionsPath);
    
    if (!file.is_open()) {
        std::cerr << "Error: Could not open options file for writing: " << optionsPath << std::endl;
        return false;
    }
    
    try {
        // Write with pretty printing (indentation)
        file << options_.dump(4);
        file.close();
        std::cout << "Saved options to: " << optionsPath << std::endl;
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Error writing options.json: " << e.what() << std::endl;
        file.close();
        return false;
    }
}

std::string ConfigManager::getHtmlLoadingMethod() const {
    if (options_.contains("htmlLoading") && 
        options_["htmlLoading"].contains("method") &&
        options_["htmlLoading"]["method"].is_string()) {
        std::string method = options_["htmlLoading"]["method"].get<std::string>();
        if (method == "file" || method == "url" || method == "html") {
            return method;
        }
    }
    return "file";  // Default
}

void ConfigManager::setHtmlLoadingMethod(const std::string& method) {
    if (method != "file" && method != "url" && method != "html") {
        std::cerr << "Warning: Invalid HTML loading method: " << method 
                  << ". Must be 'file', 'url', or 'html'" << std::endl;
        return;
    }
    
    if (!options_.contains("htmlLoading")) {
        options_["htmlLoading"] = nlohmann::json::object();
    }
    options_["htmlLoading"]["method"] = method;
}

std::string ConfigManager::getHtmlFilePath() const {
    if (options_.contains("htmlLoading") && 
        options_["htmlLoading"].contains("filePath") &&
        options_["htmlLoading"]["filePath"].is_string()) {
        return options_["htmlLoading"]["filePath"].get<std::string>();
    }
    return "demo.html";  // Default
}

std::string ConfigManager::getHtmlUrl() const {
    if (options_.contains("htmlLoading") && 
        options_["htmlLoading"].contains("url") &&
        options_["htmlLoading"]["url"].is_string()) {
        return options_["htmlLoading"]["url"].get<std::string>();
    }
    return "";  // Default
}

std::string ConfigManager::getHtmlContent() const {
    if (options_.contains("htmlLoading") && 
        options_["htmlLoading"].contains("htmlContent") &&
        options_["htmlLoading"]["htmlContent"].is_string()) {
        return options_["htmlLoading"]["htmlContent"].get<std::string>();
    }
    return "";  // Default
}
