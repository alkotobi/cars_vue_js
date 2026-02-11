# Layout Demo - Features Showcase

This document describes all the features demonstrated in `layout_demo.cpp`.

## Visual Layout

The demo creates a window with multiple panels demonstrating different layout features:

```
┌─────────────────────────────────────────────────────────┐
│  Component & Layout Demo                                 │
├──────────┬──────────────┬────────────────────────────────┤
│          │              │                                │
│ Button   │ OK Cancel    │ Settings                       │
│ Panel    │ Apply        │ ┌──────────────┐              │
│          │              │ │ Save  Reset  │              │
│ Button 1 │              │ └──────────────┘              │
│ Button 2 │              │                                │
│ Button 3 │              │                                │
│ Button 4 │              │                                │
│          │              │                                │
├──────────┴──────────────┴────────────────────────────────┤
│ Form Panel                                               │
│ ┌────────────────────┐                                  │
│ │ Enter name         │                                  │
│ │ Enter email        │                                  │
│ │ Submit             │                                  │
│ └────────────────────┘                                  │
└─────────────────────────────────────────────────────────┘
```

## Features Demonstrated

### 1. Vertical Layout
- **Location**: Left panel (buttonPanel)
- **Features**:
  - 4 buttons arranged vertically
  - 10px spacing between buttons
  - 10px margins around container
  - Automatic sizing based on content

**Code:**
```cpp
VerticalLayout* vLayout = new VerticalLayout(&window);
vLayout->setSpacing(10);
vLayout->setMargins(10, 10, 10, 10);
vLayout->addControl(btn1);
vLayout->addControl(btn2);
// ...
panel1->setLayout(vLayout);
```

### 2. Horizontal Layout
- **Location**: Top right panel (horizontalPanel)
- **Features**:
  - 3 buttons (OK, Cancel, Apply) arranged horizontally
  - 5px spacing between buttons
  - Equal distribution of space

**Code:**
```cpp
HorizontalLayout* hLayout = new HorizontalLayout(&window);
hLayout->setSpacing(5);
hLayout->addControl(okBtn);
hLayout->addControl(cancelBtn);
hLayout->addControl(applyBtn);
panel2->setLayout(hLayout);
```

### 3. Nested Layouts
- **Location**: Right panel (outerPanel)
- **Features**:
  - Outer container with vertical layout
  - Inner container with horizontal layout
  - Demonstrates containers within containers
  - Shows Owner/Parent relationships

**Code:**
```cpp
// Outer container
Container* outerPanel = new Container(&window, &window, ...);
VerticalLayout* outerLayout = new VerticalLayout(&window);
outerLayout->addControl(titleLabel);

// Inner container
Container* innerPanel = new Container(&window, outerPanel, ...);
HorizontalLayout* innerLayout = new HorizontalLayout(&window);
innerLayout->addControl(saveBtn);
innerLayout->addControl(resetBtn);
innerPanel->setLayout(innerLayout);

outerLayout->addControl(innerPanel);
outerPanel->setLayout(outerLayout);
```

### 4. Component Lookup
- **Feature**: Find components by name
- **Demonstrates**:
  - `FindComponent(name)` method
  - Dynamic casting to specific types
  - Component enumeration

**Code:**
```cpp
Component* foundBtn = window.FindComponent("btn1");
Button* btn = dynamic_cast<Button*>(foundBtn);
if (btn) {
    std::cout << "Found: " << btn->getLabel() << "\n";
}
```

### 5. Owner vs Parent Separation
- **Feature**: Different Owner and Parent
- **Demonstrates**:
  - Button owned by window (automatic cleanup)
  - Button parented to panel (visual placement)
  - Memory management vs visual hierarchy

**Code:**
```cpp
// Button owned by window, parented to panel
Button* demoBtn = new Button(&window, demoPanel, ...);
// Owner: window (cleanup)
// Parent: demoPanel (visual)
```

### 6. Component Enumeration
- **Feature**: Iterate through owned components
- **Demonstrates**:
  - `GetComponentCount()`
  - `GetComponent(index)`
  - Type information using RTTI

**Code:**
```cpp
for (int i = 0; i < window.GetComponentCount(); ++i) {
    Component* comp = window.GetComponent(i);
    std::cout << comp->GetName() << "\n";
}
```

### 7. Form with Input Fields
- **Location**: Bottom panel (formPanel)
- **Features**:
  - InputField controls in vertical layout
  - Button callback accessing input values
  - Form submission handling

**Code:**
```cpp
VerticalLayout* formLayout = new VerticalLayout(&window);
formLayout->addControl(nameField);
formLayout->addControl(emailField);
formLayout->addControl(submitBtn);

submitBtn->setCallback([nameField, emailField](Control* parent) {
    std::cout << "Name: " << nameField->getText() << "\n";
    std::cout << "Email: " << emailField->getText() << "\n";
});
```

## Key Concepts

### Component Ownership
- **Owner** determines who manages the component's memory
- When owner is destroyed, all owned components are automatically destroyed
- Prevents memory leaks

### Visual Hierarchy
- **Parent** determines where the control is visually displayed
- Controls are positioned relative to their parent
- Parent can be Window or Container

### Layout System
- **Layouts** automatically position and size controls
- Supports spacing and margins
- Updates automatically when container resizes
- Respects control size hints (min/preferred/max)

### Component Naming
- Components can have names for easy lookup
- Names must be unique within an owner
- Default names generated automatically using RTTI

## Automatic Cleanup

All components are automatically cleaned up when their owner is destroyed:

```cpp
{
    Window window(...);
    Container* panel = new Container(&window, &window, ...);
    Button* btn = new Button(&window, panel, ...);
    // ...
} // All components automatically destroyed here!
```

No manual `delete` needed - the component system handles it all!
