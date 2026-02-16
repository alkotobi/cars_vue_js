# Separating App-Specific Code from CrossDev Framework

## Goal

- **CrossDev** = general-purpose framework (webview, createWindow, file dialogs, options, etc.)
- **App code** = Cars-specific logic (invoice export, sell bills, etc.)
- Build CrossDev standalone (demo) **or** with Cars app (full product)

---

## Option 1: App Module + Registration Hook (Recommended)

### Structure

```
cars/                          # Project root
├── cross_dev/                 # Framework (general-purpose)
│   ├── include/
│   ├── src/
│   │   ├── app_runner.cpp
│   │   ├── handlers/          # Framework handlers only
│   │   └── ...
│   ├── CMakeLists.txt
│   └── apps/                  # App-specific code (optional)
│       └── cars/
│           ├── handlers/
│           │   └── create_invoice_handler.cpp
│           ├── app_handlers.cpp
│           └── CMakeLists.txt
├── src/                       # Vue frontend
└── CMakeLists.txt
```

### Mechanism

1. **CrossDev** defines a registration hook in `app_runner.cpp`:

   ```cpp
   // In app_runner.cpp
   void AppRunner::registerHandlers() {
       MessageRouter* router = eventHandler_->getMessageRouter();
       // ... framework handlers ...
       router->registerHandler(createAppInfoHandler());
       router->registerHandler(createOptionsHandler());
       // ... etc ...

       // App-specific handlers (optional, provided by app module)
       registerAppHandlers(router);
   }
   ```

2. **Hook declaration** in CrossDev (e.g. `include/app_handlers.h`):

   ```cpp
   // include/app_handlers.h
   void registerAppHandlers(class MessageRouter* router);
   ```

3. **Default (weak) implementation** – no-op when building vanilla CrossDev:

   ```cpp
   // src/app_handlers_stub.cpp (built by default)
   __attribute__((weak)) void registerAppHandlers(MessageRouter*) {}
   ```

4. **Cars app implementation** – in `apps/cars/app_handlers.cpp`:

   ```cpp
   void registerAppHandlers(MessageRouter* router) {
       router->registerHandler(createInvoiceHandler());
       // router->registerHandler(createSellBillHandler());
       // etc.
   }
   ```

5. **CMake** – when building cars app, exclude stub and link app module:

   ```cmake
   if(BUILD_CARS_APP)
       add_subdirectory(apps/cars)
       target_link_libraries(${PROJECT_NAME} PRIVATE app_cars)
       target_compile_definitions(${PROJECT_NAME} PRIVATE HAS_APP_HANDLERS=1)
   else()
       target_sources(${PROJECT_NAME} PRIVATE src/app_handlers_stub.cpp)
   endif()
   ```

---

## Option 2: Separate App Executable

Build CrossDev as a **library**, the app as a separate target:

```
cross_dev/
  CMakeLists.txt
  # Builds libcrossdev.a (no main, no app handlers)

apps/cars/
  CMakeLists.txt
  main.cpp              # Creates AppRunner, registers app handlers
  handlers/
```

- `libcrossdev.a` = framework only
- `CarsApp` executable = links libcrossdev + app_cars, owns `main()` and handler registration
- CrossDev has no knowledge of invoices

**Downside:** Two different entry points (CrossDev demo vs Cars app) require two build configs.

---

## Option 3: Handler List Injection

AppRunner receives app handlers at construction:

```cpp
// AppRunner constructor
AppRunner::AppRunner(int argc, const char* argv[],
    std::vector<std::shared_ptr<MessageHandler>> appHandlers = {})
    : appHandlers_(std::move(appHandlers)) {}

void AppRunner::registerHandlers() {
    // ... framework handlers ...
    for (auto& h : appHandlers_)
        router->registerHandler(h);
}
```

- `main.cpp` (or an app-specific `main_cars.cpp`) builds the handler list and passes it in
- Framework stays pure; app code lives in a separate directory and is only linked when building the app

**Downside:** main.cpp becomes app-specific, so you need different main files per app.

---

## Option 4: CMake Optional Sources

Keep one tree, use CMake to include app sources conditionally:

```cmake
set(APP_SOURCES "")
if(APP_NAME STREQUAL "cars")
    list(APPEND APP_SOURCES
        apps/cars/handlers/create_invoice_handler.cpp
    )
endif()
target_sources(${PROJECT_NAME} PRIVATE ${APP_SOURCES})
```

- `app_runner.cpp` uses `#ifdef HAS_INVOICE_HANDLER` to conditionally register
- Or better: `registerAppHandlers()` is always called; when `APP_SOURCES` is empty, the weak stub provides the no-op

---

## Recommended: Option 1 (Registration Hook)

| Aspect | Benefit |
|--------|---------|
| **Clear boundary** | Framework never includes app headers |
| **Single entry point** | Same main.cpp, same executable layout |
| **Weak symbol** | Vanilla CrossDev works without app module |
| **Per-app** | Each app has its own `apps/<name>/` |
| **Excel module** | Can stay in framework (general) or move to app (if only Cars uses it) |

### Excel/Invoice Placement

- **Excel library** (`excel_excel`) – keep in CrossDev if you expect multiple apps to use Excel; move to `apps/cars/` if only Cars needs it.
- **Invoice handler** – app-specific, lives in `apps/cars/handlers/`.
- **`createInvoice` logic** – app-specific wrapper in `apps/cars/`; core Excel API can remain in CrossDev.

---

## Migration Steps

1. Add `include/app_handlers.h` with `void registerAppHandlers(MessageRouter*);`
2. Add `src/app_handlers_stub.cpp` with weak no-op implementation
3. In `app_runner.cpp`, call `registerAppHandlers(router)` at end of `registerHandlers()`
4. Create `apps/cars/` with `create_invoice_handler`, `app_handlers.cpp`
5. Add CMake option `BUILD_CARS_APP` and wire `apps/cars` when enabled
6. Move or copy Excel/invoice code into `apps/cars` if you want it fully app-specific
