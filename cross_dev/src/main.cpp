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
        
        // Create a window using the platform-agnostic interface
        Window window(100, 100, 800, 600, "Web View Test");
        
        // Create components - each component is a separate class
        InputField inputField(&window, 10, 10, 500, 30, "Enter URL or file path");
        
        // Create web view (positioned below input/button area)
        WebView webView(&window, 10, 50, 780, 540);
        
        // Create event handler to manage all HTML/webview events
        EventHandler eventHandler(&window, &inputField, &webView);
        
        // Register button click event
        Button loadButton(&window, 520, 10, 150, 30, "Load File");
        loadButton.setCallback([&eventHandler](Window* w) {
            eventHandler.onLoadButtonClicked();
        });
        
        // Register webview create window event
        eventHandler.onWebViewCreateWindow([](const std::string& title) {
            try {
                // Create a new window
                // The Window object will be automatically deleted when the window is closed
                // (handled in WM_DESTROY message handler on Windows)
                Window* newWindow = new Window(150, 150, 600, 400, title);
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
        
        // Load the demo.html file - read it as a string to avoid path issues
        std::ifstream htmlFile("demo.html");
        if (!htmlFile.is_open()) {
            // Try absolute path from project root
            htmlFile.open("/Users/merhab/dev/native dev/cpp test/demo.html");
        }
        if (htmlFile.is_open()) {
            std::stringstream buffer;
            buffer << htmlFile.rdbuf();
            std::string htmlContent = buffer.str();
            htmlFile.close();
            webView.loadHTMLString(htmlContent);
            std::cout << "Loaded demo.html successfully" << std::endl;
        } else {
            std::cerr << "Warning: Could not open demo.html, loading empty page" << std::endl;
            webView.loadHTMLString("<html><body><h1>Error: Could not load demo.html</h1><p>Please ensure demo.html is in the current directory or project root.</p></body></html>");
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
