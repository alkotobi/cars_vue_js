#include "../include/window.h"
#include "../include/webview.h"
#include "../include/button.h"
#include "../include/inputfield.h"
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
        
        // Create a window using the platform-agnostic interface
        // Window(owner, parent, x, y, width, height, title)
        Window window(nullptr, nullptr, 100, 100, 800, 600, "Web View Test");
        
        // Create components - each component is a separate class
        // InputField(owner, parent, x, y, width, height, placeholder)
        InputField inputField(&window, &window, 10, 10, 500, 30, "Enter URL or file path");
        
        // Create web view (positioned below input/button area)
        // WebView(owner, parent, x, y, width, height)
        WebView webView(&window, &window, 10, 50, 780, 540);
        
        // Create event handler to manage all HTML/webview events
        EventHandler eventHandler(&window, &inputField, &webView);
        
        // Register button click event
        // Button(owner, parent, x, y, width, height, label)
        Button loadButton(&window, &window, 520, 10, 150, 30, "Load File");
        loadButton.setCallback([&eventHandler](Control* parent) {
            eventHandler.onLoadButtonClicked();
        });
        
        // Register webview create window event
        eventHandler.onWebViewCreateWindow([](const std::string& title) {
            try {
                // Create a new window
                // The Window object will be automatically deleted when the window is closed
                // (handled in WM_DESTROY message handler on Windows)
                Window* newWindow = new Window(nullptr, nullptr, 150, 150, 600, 400, title);
                newWindow->show();
                std::cout << "Created new window: " << title << std::endl;
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
        auto fileDialogHandler = createFileDialogHandler(&window);
        router->registerHandler(fileDialogHandler);
        std::cout << "Registered FileDialog handler" << std::endl;
        
        std::cout << "All handlers registered successfully!" << std::endl;
        
        // Load HTML based on options.json configuration
        std::string loadingMethod = config.getHtmlLoadingMethod();
        std::cout << "HTML loading method: " << loadingMethod << std::endl;
        
        if (loadingMethod == "file") {
            // Load from file
            std::string filePath = config.getHtmlFilePath();
            std::ifstream htmlFile(filePath);
            if (!htmlFile.is_open()) {
                // Try current directory
                htmlFile.open("demo.html");
            }
            if (htmlFile.is_open()) {
                std::stringstream buffer;
                buffer << htmlFile.rdbuf();
                std::string htmlContent = buffer.str();
                htmlFile.close();
                webView.loadHTMLString(htmlContent);
                inputField.setText(filePath);
                std::cout << "Loaded HTML file: " << filePath << std::endl;
            } else {
                std::cerr << "Warning: Could not open HTML file: " << filePath << std::endl;
                webView.loadHTMLString("<html><body><h1>Error: Could not load HTML file</h1><p>File: " + filePath + "</p><p>Please check the file path in options.json</p></body></html>");
            }
        } else if (loadingMethod == "url") {
            // Load from URL
            std::string url = config.getHtmlUrl();
            if (!url.empty()) {
                webView.loadURL(url);
                inputField.setText(url);
                std::cout << "Loading URL: " << url << std::endl;
            } else {
                std::cerr << "Warning: URL is empty in options.json" << std::endl;
                webView.loadHTMLString("<html><body><h1>Error: No URL specified</h1><p>Please set the URL in options.json</p></body></html>");
            }
        } else if (loadingMethod == "html") {
            // Load from HTML content in JSON
            std::string htmlContent = config.getHtmlContent();
            if (!htmlContent.empty()) {
                webView.loadHTMLString(htmlContent);
                inputField.setText("(HTML from options.json)");
                std::cout << "Loaded HTML content from options.json" << std::endl;
            } else {
                std::cerr << "Warning: HTML content is empty in options.json" << std::endl;
                webView.loadHTMLString("<html><body><h1>Error: No HTML content specified</h1><p>Please set the htmlContent in options.json</p></body></html>");
            }
        } else {
            // Fallback to default behavior
            std::cerr << "Warning: Unknown HTML loading method: " << loadingMethod << std::endl;
            std::ifstream htmlFile("demo.html");
            if (htmlFile.is_open()) {
                std::stringstream buffer;
                buffer << htmlFile.rdbuf();
                std::string htmlContent = buffer.str();
                htmlFile.close();
                webView.loadHTMLString(htmlContent);
                std::cout << "Loaded demo.html (fallback)" << std::endl;
            } else {
                webView.loadHTMLString("<html><body><h1>Error: Could not load HTML</h1><p>Please configure options.json</p></body></html>");
            }
        }
        
        // Show the window
        window.show();
        
        std::cout << "Window created with web view and button." << std::endl;
        std::cout << "Click the button to load an HTML file." << std::endl;
        
        // Run the application event loop
        Application::getInstance().run();
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}
