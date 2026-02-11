#include "../include/webview_window.h"
#include "../include/application.h"
#include "../include/eventhandler.h"
#include "../include/messagerouter.h"
#include "../include/handlers/createwindowhandler.h"
#include "../include/handlers/appinfohandler.h"
#include "../include/handlers/calculatorhandler.h"
#include "../include/handlers/filedialoghandler.h"
#include "../include/config_manager.h"
#include "../include/platform.h"
#include "platform/platform_impl.h"
#include <iostream>
#include <exception>
#include <string>
#include <ctime>
#include <fstream>
#include <sstream>

int main(int /*argc*/, const char* /*argv*/[]) {
    try {
        std::cout << "Running on " << PLATFORM_NAME << std::endl;
        
        // Load application options
        ConfigManager& config = ConfigManager::getInstance();
        if (!config.loadOptions()) {
            std::cerr << "Warning: Failed to load options, using defaults" << std::endl;
        }
        
        std::cout << "Config directory: " << ConfigManager::getConfigDirectory() << std::endl;
        std::cout << "Options file: " << ConfigManager::getOptionsFilePath() << std::endl;
        
        // Determine content type and content from config
        std::string loadingMethod = config.getHtmlLoadingMethod();
        WebViewContentType contentType = WebViewContentType::Default;
        std::string content;

        if (loadingMethod == "html") {
            contentType = WebViewContentType::Html;
            content = config.getHtmlContent();
            if (content.empty()) {
                content = ConfigManager::tryLoadFileContent("demo.html");
            }
        } else if (loadingMethod == "url") {
            contentType = WebViewContentType::Url;
            content = config.getHtmlUrl();
        } else if (loadingMethod == "file") {
            contentType = WebViewContentType::File;
            content = config.getHtmlFilePath();
        } else {
            content = ConfigManager::tryLoadFileContent("demo.html");
        }

        // Main window: owner=nullptr makes it the unique "main" WebViewWindow.
        // All child windows created from JS will use this as owner and be auto-freed when main is destroyed.
        WebViewWindow mainWindow(nullptr, 100, 100, 800, 600, "Web View Test", contentType, content);

        // Create event handler
        EventHandler eventHandler(mainWindow.getWindow(), mainWindow.getWebView());

        // Child windows from JS use main as owner - they are auto-freed when main is destroyed
        eventHandler.onWebViewCreateWindow([&mainWindow](const std::string& title) {
            try {
                WebViewWindow* child = new WebViewWindow(&mainWindow, 150, 150, 600, 400, title);
                child->show();
                std::cout << "Created new window (owned by main): " << title << std::endl;
            } catch (const std::exception& e) {
                std::cerr << "Error creating new window: " << e.what() << std::endl;
            }
        });
        
        // Register additional demo handlers
        MessageRouter* router = eventHandler.getMessageRouter();
        
        // Register AppInfo handler (returns system/app information)
        auto appInfoHandler = createAppInfoHandler();
        router->registerHandler(appInfoHandler);
        std::cout << "Registered AppInfo handler" << std::endl;
        
        // Register Calculator handler (performs calculations and returns results)
        auto calculatorHandler = createCalculatorHandler();
        router->registerHandler(calculatorHandler);
        std::cout << "Registered Calculator handler" << std::endl;
        
        // Register FileDialog handler (opens native file dialog)
        auto fileDialogHandler = createFileDialogHandler(mainWindow.getWindow());
        router->registerHandler(fileDialogHandler);
        std::cout << "Registered FileDialog handler" << std::endl;
        
        std::cout << "All handlers registered successfully!" << std::endl;
        std::cout << "HTML loading method: " << loadingMethod << std::endl;

        // Content loaded by WebViewWindow constructor; fallbacks if empty
        if (loadingMethod == "html" && content.empty()) {
            mainWindow.loadHTMLString("<html><body><h1>Error: No HTML content</h1><p>Set htmlContent in options.json</p></body></html>");
        } else if (loadingMethod == "url" && content.empty()) {
            mainWindow.loadHTMLString("<html><body><h1>Error: No URL specified</h1><p>Set url in options.json</p></body></html>");
        } else if (loadingMethod == "file" && content.empty()) {
            std::string htmlContent = ConfigManager::tryLoadFileContent("demo.html");
            if (!htmlContent.empty()) {
                mainWindow.loadHTMLString(htmlContent);
            } else {
                mainWindow.loadHTMLString("<html><body><h1>Error: Could not load file</h1><p>Check filePath in options.json</p></body></html>");
            }
        }

        mainWindow.show();
        
        std::cout << "Window created with web view." << std::endl;
        
        // Run the application event loop
        Application::getInstance().run();
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}
