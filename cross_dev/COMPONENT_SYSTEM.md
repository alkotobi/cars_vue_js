# Component System - Owner/Parent Architecture

This document explains the Delphi-style component ownership and parent system implemented in CrossDev.

## Overview

The component system uses a two-level hierarchy similar to Delphi:

- **Owner** (`Component.Owner`): Memory management - determines who owns the component and automatically frees it
- **Parent** (`Control.Parent`): Visual hierarchy - determines where the control is visually displayed

## Key Concepts

### Component Class
Base class for all components. Handles:
- Ownership chain management
- Automatic cleanup when owner is destroyed
- Component naming and lookup
- Component enumeration

### Control Class
Inherits from Component. Adds:
- Parent relationship for visual hierarchy
- Position and size properties (Left, Top, Width, Height)
- Visibility management
- Control enumeration (visual children)

## Usage Examples

### Basic Ownership

```cpp
Window window(nullptr, nullptr, 100, 100, 800, 600, "My App");
Button* btn = new Button(&window, &window, 10, 10, 100, 30, "Click");
// When window is destroyed, btn is automatically destroyed
```

### Owner and Parent Can Be Different

```cpp
Window window(nullptr, nullptr, 100, 100, 800, 600, "My App");
Container panel(&window, &window, 10, 10, 300, 200);

// Button owned by window, but parented to panel
Button btn(&window, &panel, 5, 5, 100, 30, "Click Me");
// - Owner: window (automatic cleanup)
// - Parent: panel (visual placement)
```

### Nested Hierarchy

```cpp
Window window(nullptr, nullptr, 100, 100, 800, 600, "My App");
Container outerPanel(&window, &window, 10, 10, 500, 400);
Container innerPanel(&window, &outerPanel, 10, 10, 200, 200);
Button btn(&window, &innerPanel, 5, 5, 100, 30, "Click");
```

### Component Lookup by Name

```cpp
Window window(nullptr, nullptr, 100, 100, 800, 600, "My App");
Button* btn = new Button(&window, &window, 10, 10, 100, 30, "OK");
btn->SetName("btnOK");

// Later...
Button* found = dynamic_cast<Button*>(window.FindComponent("btnOK"));
```

### Component Enumeration

```cpp
// Enumerate owned components
for (int i = 0; i < window.GetComponentCount(); i++) {
    Component* comp = window.GetComponent(i);
    // Process component
}

// Enumerate visual children
for (int i = 0; i < window.GetControlCount(); i++) {
    Control* ctrl = window.GetControl(i);
    // Process control
}
```

## Constructor Patterns

All UI components follow this pattern:

```cpp
Component(Component* owner, Control* parent, int x, int y, int width, int height, ...)
```

- `owner`: Who owns this component (for memory management). Can be `nullptr` for top-level components.
- `parent`: Where this control is visually displayed. Must be a `Control*` (Window, Container, etc.). Can be `nullptr` for top-level windows.
- `x, y, width, height`: Position and size relative to parent

## Available Components

- **Window**: Top-level window (can own and parent other components)
- **Container**: Panel-like container (can own and parent other components)
- **Button**: Clickable button
- **InputField**: Text input field
- **WebView**: Web browser view

## Key Features

1. **Automatic Cleanup**: When an owner is destroyed, all owned components are automatically destroyed
2. **Visual Hierarchy**: Parent determines where controls appear visually
3. **Flexible Relationships**: Owner and Parent can be different
4. **Component Naming**: Each component can have a name for easy lookup
5. **Enumeration**: Iterate through owned components or visual children

## Best Practices

1. **Always specify an Owner** for automatic memory management
2. **Use Parent for visual placement** - controls must have a parent to be visible
3. **Name important components** for easy lookup later
4. **Use Container for grouping** - create logical UI groups
5. **Owner typically matches Parent** for simple cases, but they can differ for complex layouts

## Example: Complex Layout

```cpp
Window window(nullptr, nullptr, 100, 100, 800, 600, "My App");

// Create panels
Container leftPanel(&window, &window, 10, 10, 200, 500);
Container rightPanel(&window, &window, 220, 10, 570, 500);

// Buttons in left panel
Button btn1(&window, &leftPanel, 10, 10, 180, 30, "Button 1");
Button btn2(&window, &leftPanel, 10, 50, 180, 30, "Button 2");

// Input in right panel
InputField input(&window, &rightPanel, 10, 10, 550, 30, "Enter text");

// All components owned by window (automatic cleanup)
// Visual hierarchy: window -> panels -> buttons/input
```

## Platform Support

All platform implementations support:
- Any Control as parent (Window, Container, etc.)
- Automatic native handle management
- Proper cleanup on destruction

The system works seamlessly across macOS, Windows, Linux, and iOS.

## ComponentCollection (Optional Optimization)

For applications with very large component hierarchies (1000+ components), you can use `ComponentCollection` for more efficient lookups:

```cpp
#include "component_collection.h"

ComponentCollection collection;
Button* btn1 = new Button(&window, &window, 10, 10, 100, 30, "OK");
Button* btn2 = new Button(&window, &window, 10, 50, 100, 30, "Cancel");

collection.Add(btn1, "btnOK");
collection.Add(btn2, "btnCancel");

// O(1) name lookup (vs O(n) with FindComponent)
Button* found = dynamic_cast<Button*>(collection.Find("btnOK"));

// O(1) index access
Component* first = collection.Get(0);
```

**Performance:**
- Name lookup: O(1) average case (vs O(n) with `FindComponent`)
- Index access: O(1)
- Memory overhead: Slightly higher due to hash maps

**When to use:**
- Large component hierarchies (1000+ components)
- Frequent name-based lookups
- Performance-critical code paths

For typical applications, the built-in `std::vector` implementation in `Component` is sufficient.
