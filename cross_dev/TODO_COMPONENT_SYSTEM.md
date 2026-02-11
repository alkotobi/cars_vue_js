# Todo List: Delphi-Style Component Ownership System

This document tracks the implementation of a Delphi-style component ownership and parent system for the CrossDev project.

## ✅ Implementation Complete!

The component system has been successfully implemented. All core tasks are complete. See [COMPONENT_SYSTEM.md](COMPONENT_SYSTEM.md) for usage documentation.

## Status Summary

- ✅ **All Core Features Implemented**: Component, Control, Window, Button, InputField, WebView, Container
- ✅ **All Platform Support**: macOS, Windows, Linux, iOS
- ✅ **All Tests Passing**: 5 test suites, 100% pass rate
- ✅ **Documentation Complete**: COMPONENT_SYSTEM.md and README updated
- ⏸️ **Optional Enhancements**: ComponentCollection class (optimization, not critical)

## Architecture Overview

Delphi uses a two-level system:

- **Owner** (`TComponent.Owner`): Memory management - determines who owns the component and frees it
- **Parent** (`TControl.Parent`): Visual hierarchy - determines where the control is visually displayed

## Implementation Phases

### Phase 1: Foundation - Base Classes

#### ✅ Task 1: Create base Component class (TComponent equivalent)

- [ ] Create `include/component.h` - base class for all components
- [ ] Add `Owner` property (Component\*)
- [ ] Add `Name` property (std::string) for component identification
- [ ] Implement ownership chain management
- [ ] Add virtual destructor that cleans up owned components
- [ ] Add `FindComponent(const std::string& name)` method
- [ ] Add `GetOwner()` and `SetOwner()` methods
- [ ] Create `src/component.cpp` with implementation

**Files to create:**

- `cross_dev/include/component.h`
- `cross_dev/src/component.cpp`

**Key features:**

- Automatic cleanup when owner is destroyed
- Component name lookup
- Ownership chain traversal

---

#### ✅ Task 2: Create base Control class (TControl equivalent)

- [ ] Create `include/control.h` - inherits from Component
- [ ] Add `Parent` property (Control\*)
- [ ] Add position/size properties (Left, Top, Width, Height)
- [ ] Add `Visible` property
- [ ] Implement parent change notifications
- [ ] Add `GetParent()` and `SetParent()` methods
- [ ] Add virtual methods for visual operations
- [ ] Create `src/control.cpp` with implementation

**Files to create:**

- `cross_dev/include/control.h`
- `cross_dev/src/control.cpp`

**Key features:**

- Visual hierarchy management
- Parent change callbacks
- Coordinate system relative to parent

---

### Phase 2: Refactor Existing Classes

#### ✅ Task 3: Refactor Window class

- [ ] Make Window inherit from Control
- [ ] Window becomes both a Component (can own others) and Control (can be parented)
- [ ] Update constructor to accept Owner parameter
- [ ] Add methods to manage child controls
- [ ] Update platform implementations if needed
- [ ] Ensure backward compatibility

**Files to modify:**

- `cross_dev/include/window.h`
- `cross_dev/src/window.cpp`
- Platform-specific window implementations

**Breaking changes:**

- Constructor signature may change
- Need to handle Owner parameter

---

#### ✅ Task 4: Refactor Button class

- [ ] Make Button inherit from Control
- [ ] Update constructor: `Button(Component* owner, Control* parent, ...)`
- [ ] Remove direct Window\* dependency
- [ ] Use Parent for visual hierarchy
- [ ] Update platform implementations to use Parent instead of Window\*

**Files to modify:**

- `cross_dev/include/button.h`
- `cross_dev/src/button.cpp`
- `cross_dev/src/platform/*/button_*.cpp|mm`

**Key changes:**

- Owner can be Window or Container
- Parent determines visual placement
- Can be parented to Window or Container

---

#### ✅ Task 5: Refactor InputField class

- [ ] Make InputField inherit from Control
- [ ] Update constructor: `InputField(Component* owner, Control* parent, ...)`
- [ ] Remove direct Window\* dependency
- [ ] Use Parent for visual hierarchy
- [ ] Update platform implementations

**Files to modify:**

- `cross_dev/include/inputfield.h`
- `cross_dev/src/inputfield.cpp`
- `cross_dev/src/platform/*/input_*.cpp|mm`

---

#### ✅ Task 6: Refactor WebView class

- [ ] Make WebView inherit from Control
- [ ] Update constructor: `WebView(Component* owner, Control* parent, ...)`
- [ ] Remove direct Window\* dependency
- [ ] Use Parent for visual hierarchy
- [ ] Update platform implementations

**Files to modify:**

- `cross_dev/include/webview.h`
- `cross_dev/src/webview.cpp`
- `cross_dev/src/platform/*/webview_*.cpp|mm`

---

### Phase 3: New Container Class

#### ✅ Task 7: Create Container class (Panel equivalent)

- [ ] Create `include/container.h` - inherits from Control
- [ ] Container can own and parent other controls
- [ ] Add background color property
- [ ] Add border style property
- [ ] Implement child control management
- [ ] Add layout management (optional)
- [ ] Create `src/container.cpp` with implementation
- [ ] Add platform implementations for all platforms

**Files to create:**

- `cross_dev/include/container.h`
- `cross_dev/src/container.cpp`
- `cross_dev/src/platform/macos/container_macos.mm`
- `cross_dev/src/platform/windows/container_windows.cpp`
- `cross_dev/src/platform/linux/container_linux.cpp`
- `cross_dev/src/platform/ios/container_ios.mm`

**Usage example:**

```cpp
Window window(nullptr, 100, 100, 800, 600, "My App");
Container panel(&window, &window, 10, 10, 200, 300);
Button btn(&window, &panel, 5, 5, 100, 30, "Click Me");
```

---

### Phase 4: Core Functionality

#### ✅ Task 8: Implement automatic cleanup

- [ ] In Component destructor, iterate through owned components
- [ ] Delete all owned components before destroying self
- [ ] Handle circular references (shouldn't happen with proper design)
- [ ] Add unit tests for cleanup scenarios
- [ ] Test with nested ownership (Window owns Container, Container owns Button)

**Files to modify:**

- `cross_dev/src/component.cpp`

**Test cases:**

- Window destroyed → all owned components destroyed
- Container destroyed → owned buttons destroyed
- Component removed from owner → not destroyed with owner

---

#### ✅ Task 9: Implement Parent change notifications

- [x] Add `OnParentChanged` virtual method
- [x] When Parent changes, update native handle parent
- [x] Recalculate coordinates relative to new parent
- [x] Update platform implementations to handle parent changes
- [x] Add validation (can't parent to self, etc.) - via validateParent()

**Files to modify:**

- `cross_dev/src/control.cpp`
- Platform-specific implementations

**Key features:**

- Automatic coordinate recalculation
- Native handle reparenting
- Validation and error handling

---

#### ✅ Task 10: Add component enumeration

- [x] Add `GetComponentCount()` method to Component
- [x] Add `GetComponent(int index)` method
- [x] Add `GetComponents()` method returning vector
- [x] Add `GetControlCount()` method to Control (for visual children)
- [x] Add `GetControl(int index)` method
- [x] Add iterator support for modern C++ - vectors work with range-based for loops

**Files to modify:**

- `cross_dev/include/component.h`
- `cross_dev/include/control.h`
- `cross_dev/src/component.cpp`
- `cross_dev/src/control.cpp`

**Usage:**

```cpp
for (auto* comp : window->GetComponents()) {
    // Iterate all owned components
}
for (auto* ctrl : window->GetControls()) {
    // Iterate visual children
}
```

---

### Phase 5: Platform Updates

#### ✅ Task 11: Update platform implementations

- [ ] Update platform_impl.h to support Control* as parent (not just Window*)
- [ ] Update macOS implementations to handle Control parent
- [ ] Update Windows implementations to handle Control parent
- [ ] Update Linux implementations to handle Control parent
- [ ] Update iOS implementations to handle Control parent
- [ ] Add Container platform implementations

**Files to modify:**

- `cross_dev/src/platform/platform_impl.h`
- All platform-specific implementation files

**Key changes:**

- `createButton(Control* parent, ...)` instead of `createButton(Window* parent, ...)`
- Support for Container as parent
- Native handle reparenting

---

### Phase 6: Enhanced Features

#### ✅ Task 12: Add component naming system

- [x] Add `Name` property to Component
- [x] Add `FindComponent(const std::string& name)` method
- [x] Add name validation (unique within owner)
- [x] Add name change notifications (via SetName validation)
- [x] Support for default names (Button1, Button2, etc.) - via generateDefaultName() using typeid

**Files to modify:**

- `cross_dev/include/component.h`
- `cross_dev/src/component.cpp`

**Usage:**

```cpp
Button* btn = new Button(&window, &panel, 10, 10, 100, 30, "OK");
btn->SetName("btnOK");
// Later...
Button* found = dynamic_cast<Button*>(window->FindComponent("btnOK"));
```

---

#### ✅ Task 13: Create ComponentCollection class (Optional Optimization)

- [x] Create `include/component_collection.h`
- [x] Efficient component storage and lookup
- [x] Support for name-based lookup (O(1) average case)
- [x] Support for index-based access (O(1))
- [x] Create `src/component_collection.cpp`
- [x] Add unit tests

**Status**: ✅ **Complete!** ComponentCollection provides O(1) name lookup and O(1) index access using `std::unordered_map` and `std::vector`. Useful for large component hierarchies (1000+ components) or performance-critical code paths.

**Files created:**

- `cross_dev/include/component_collection.h`
- `cross_dev/src/component_collection.cpp`
- `cross_dev/tests/test_component_collection.cpp`

**Files to create:**

- `cross_dev/include/component_collection.h`
- `cross_dev/src/component_collection.cpp`

**Design:**

- Use std::unordered_map for name lookup
- Use std::vector for index access
- Maintain both for efficiency

---

### Phase 7: Integration and Testing

#### ✅ Task 14: Update WebViewWindow

- [ ] Refactor WebViewWindow to use new component system
- [ ] Window owns WebView
- [ ] Window is parent of WebView
- [ ] Demonstrate Owner/Parent pattern
- [ ] Update main.cpp example

**Files to modify:**

- `cross_dev/include/webview_window.h`
- `cross_dev/src/webview_window.cpp`
- `cross_dev/src/main.cpp`

**Example:**

```cpp
WebViewWindow window(nullptr, 100, 100, 800, 600, "My App");
// Window owns WebView (automatic cleanup)
// Window is parent of WebView (visual hierarchy)
```

---

#### ✅ Task 15: Add unit tests

- [ ] Create test framework setup
- [ ] Test component ownership (creation, destruction)
- [ ] Test parent relationships
- [ ] Test component lookup by name
- [ ] Test automatic cleanup
- [ ] Test parent change notifications
- [ ] Test Container class

**Files to create:**

- `cross_dev/tests/test_component.cpp`
- `cross_dev/tests/test_control.cpp`
- `cross_dev/tests/test_container.cpp`
- `cross_dev/tests/test_ownership.cpp`

---

#### ✅ Task 16: Update documentation

- [ ] Create `docs/COMPONENT_SYSTEM.md` explaining the architecture
- [ ] Add usage examples for Owner/Parent pattern
- [ ] Document migration guide from old to new system
- [ ] Add API reference for Component and Control classes
- [ ] Update README.md with component system overview

**Files to create/modify:**

- `cross_dev/docs/COMPONENT_SYSTEM.md`
- `cross_dev/README.md`

---

## File Structure (After Implementation)

```
cross_dev/
├── include/
│   ├── component.h              # Base Component class (Owner)
│   ├── control.h                # Base Control class (Parent)
│   ├── container.h              # Container class (Panel equivalent)
│   ├── window.h                 # Window (inherits Control)
│   ├── button.h                 # Button (inherits Control)
│   ├── inputfield.h             # InputField (inherits Control)
│   ├── webview.h                # WebView (inherits Control)
│   ├── webview_window.h         # WebViewWindow (uses new system)
│   └── component_collection.h   # Component collection management
├── src/
│   ├── component.cpp
│   ├── control.cpp
│   ├── container.cpp
│   ├── component_collection.cpp
│   └── ... (other existing files)
└── tests/
    ├── test_component.cpp
    ├── test_control.cpp
    └── test_ownership.cpp
```

## Usage Examples

### Example 1: Simple Ownership

```cpp
// Window owns Button (automatic cleanup)
Window window(nullptr, 100, 100, 800, 600, "My App");
Button* btn = new Button(&window, &window, 10, 10, 100, 30, "Click");
// When window is destroyed, btn is automatically destroyed
```

### Example 2: Container with Buttons

```cpp
Window window(nullptr, 100, 100, 800, 600, "My App");
Container* panel = new Container(&window, &window, 10, 10, 300, 200);
Button* btn1 = new Button(&window, panel, 5, 5, 100, 30, "Button 1");
Button* btn2 = new Button(&window, panel, 5, 40, 100, 30, "Button 2");
// Window owns panel, btn1, btn2
// Panel is parent of btn1, btn2 (visual hierarchy)
```

### Example 3: Component Lookup

```cpp
Button* btn = new Button(&window, &window, 10, 10, 100, 30, "OK");
btn->SetName("btnOK");
// Later...
Button* found = dynamic_cast<Button*>(window->FindComponent("btnOK"));
```

## Migration Strategy

1. **Phase 1-2**: Create new base classes alongside existing code
2. **Phase 3**: Refactor one class at a time (Window first, then Button, etc.)
3. **Phase 4**: Add new features incrementally
4. **Phase 5**: Update all platform code
5. **Phase 6-7**: Testing and documentation

## Benefits

- ✅ Automatic memory management (no manual delete needed)
- ✅ Clear visual hierarchy (Parent determines layout)
- ✅ Flexible architecture (components can be owned by different owners)
- ✅ Component lookup by name
- ✅ Better code organization
- ✅ Similar to Delphi (familiar to many developers)

## Notes

- Owner and Parent can be different (like Delphi)
- Owner determines lifetime, Parent determines visual placement
- All components must have an Owner (for cleanup)
- Visual controls must have a Parent (for display)
- Container class enables nested UI hierarchies
