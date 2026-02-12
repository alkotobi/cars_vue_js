#include "../include/app_runner.h"
#include "../include/application.h"
#include "../include/eventhandler.h"
#include "../include/messagerouter.h"
#include "../include/config_manager.h"
#include "../include/platform.h"
#include "../include/handlers/createwindowhandler.h"
#include "../include/handlers/appinfohandler.h"
#include "../include/handlers/calculatorhandler.h"
#include "../include/handlers/filedialoghandler.h"
#include "../include/handlers/readfilehandler.h"
#include "../include/handlers/writefilehandler.h"
#include "../include/handlers/filesystemhandler.h"
#include "../include/handlers/contextmenuhandler.h"
#include "platform/platform_impl.h"
#include <iostream>

AppRunner::AppRunner(int argc, const char* argv[])
    : argc_(argc), argv_(argv) {
}

AppRunner::~AppRunner() {
    eventHandler_.reset();
    mainWindow_.reset();
}

void AppRunner::loadConfig() {
    ConfigManager& config = ConfigManager::getInstance();
    if (!config.loadOptions()) {
        std::cerr << "Warning: Failed to load options, using defaults" << std::endl;
    }

    loadingMethod_ = config.getHtmlLoadingMethod();
    contentType_ = WebViewContentType::Default;
    content_.clear();

    if (loadingMethod_ == "html") {
        contentType_ = WebViewContentType::Html;
        content_ = config.getHtmlContent();
        if (content_.empty()) {
            content_ = ConfigManager::tryLoadFileContent("demo.html");
        }
    } else if (loadingMethod_ == "url") {
        contentType_ = WebViewContentType::Url;
        content_ = config.getHtmlUrl();
    } else if (loadingMethod_ == "file") {
        contentType_ = WebViewContentType::File;
        content_ = config.getHtmlFilePath();
    } else {
        content_ = ConfigManager::tryLoadFileContent("demo.html");
    }
}

void AppRunner::createMainWindow() {
    mainWindow_ = std::make_unique<WebViewWindow>(
        nullptr, 100, 100, 800, 600,
        "Web View Test", contentType_, content_);
}

void AppRunner::setupEventHandler() {
    eventHandler_ = std::make_unique<EventHandler>(
        mainWindow_->getWindow(), mainWindow_->getWebView());

    eventHandler_->onWebViewCreateWindow(
        [this](const std::string& title, WebViewContentType type, const std::string& cnt) {
            try {
                WebViewWindow* child = new WebViewWindow(
                    mainWindow_.get(), 150, 150, 600, 400, title, type, cnt);
                eventHandler_->attachWebView(child->getWebView());
                child->show();
                std::cout << "Created new window (owned by main): " << title << std::endl;
            } catch (const std::exception& e) {
                std::cerr << "Error creating new window: " << e.what() << std::endl;
            }
        });
}

void AppRunner::registerHandlers() {
    MessageRouter* router = eventHandler_->getMessageRouter();

    router->registerHandler(createAppInfoHandler());
    router->registerHandler(createCalculatorHandler());
    router->registerHandler(createFileDialogHandler(mainWindow_->getWindow()));
    router->registerHandler(createReadFileHandler());
    router->registerHandler(createWriteFileHandler());
    router->registerHandler(createFileSystemHandler());
    router->registerHandler(createContextMenuHandler(mainWindow_.get(), router));
}

void AppRunner::setupContentFallbacks() {
    if (loadingMethod_ == "html" && content_.empty()) {
        mainWindow_->loadHTMLString(
            "<html><body><h1>Error: No HTML content</h1>"
            "<p>Set htmlContent in options.json</p></body></html>");
    } else if (loadingMethod_ == "url" && content_.empty()) {
        mainWindow_->loadHTMLString(
            "<html><body><h1>Error: No URL specified</h1>"
            "<p>Set url in options.json</p></body></html>");
    } else if (loadingMethod_ == "file" && content_.empty()) {
        std::string htmlContent = ConfigManager::tryLoadFileContent("demo.html");
        if (!htmlContent.empty()) {
            mainWindow_->loadHTMLString(htmlContent);
        } else {
            mainWindow_->loadHTMLString(
                "<html><body><h1>Error: Could not load file</h1>"
                "<p>Check filePath in options.json</p></body></html>");
        }
    }
}

int AppRunner::run() {
    std::cout << "Running on " << PLATFORM_NAME << std::endl;
    std::cout << "Config directory: " << ConfigManager::getConfigDirectory() << std::endl;
    std::cout << "Options file: " << ConfigManager::getOptionsFilePath() << std::endl;

    loadConfig();
    createMainWindow();
    setupEventHandler();
    registerHandlers();

    std::cout << "HTML loading method: " << loadingMethod_ << std::endl;

    setupContentFallbacks();

    mainWindow_->show();

    platform::deliverOpenFilePaths(argc_, argv_);

    std::cout << "Window created with web view." << std::endl;

    Application::getInstance().run();

    return 0;
}
