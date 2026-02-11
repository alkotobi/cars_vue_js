# How to Run the Layout Demo

## Build and Run Instructions

The layout demo requires full platform libraries to build and run. Here's how to do it:

### macOS

```bash
cd cross_dev/build
cmake ..
make layout_demo
./examples/layout_demo
```

### Windows

```bash
cd cross_dev\build
cmake .. -G "Visual Studio 17 2022" -A x64
cmake --build . --config Release --target layout_demo
.\examples\Release\layout_demo.exe
```

### Linux

```bash
cd cross_dev/build
cmake ..
make layout_demo
./examples/layout_demo
```

## What the Demo Shows

When you run the demo, you'll see:

1. **Console Output** - Detailed information about each feature being demonstrated
2. **Visual Window** - A window with multiple panels showing:
   - Vertical layout with 4 buttons (left side)
   - Horizontal layout with 3 buttons (top right)
   - Nested layouts (right side)
   - Form with input fields (bottom)

## Expected Console Output

```
=== CrossDev Component & Layout System Demo ===

Example 1: Vertical Layout with Buttons
✓ Created vertical layout with 4 buttons
  - Layout spacing: 10px
  - Layout controls: 4

Example 2: Horizontal Layout with Buttons
✓ Created horizontal layout with 3 buttons
  - Layout spacing: 5px
  - Layout controls: 3

Example 3: Nested Layouts
✓ Created nested layouts
  - Outer panel: outerPanel
  - Inner panel: innerPanel
  - Inner panel owner: mainWindow
  - Inner panel parent: outerPanel

Example 4: Component Lookup by Name
✓ Found button by name: Button 1
✓ Found panel by name: buttonPanel
  - Panel has 0 owned components

Example 5: Owner vs Parent Separation
✓ Created button with different Owner and Parent
  - Button owner: mainWindow
  - Button parent: demoPanel
  - When window is destroyed, button is automatically cleaned up
  - Button is visually displayed in demoPanel

Example 6: Component Enumeration
Window owns X components:
  [0] buttonPanel (type: ...)
  [1] horizontalPanel (type: ...)
  ...

Example 7: Layout with Input Fields
✓ Created form with vertical layout
  - Name input field
  - Email input field
  - Submit button

=== Summary ===
✓ Component ownership system (automatic cleanup)
✓ Parent-based visual hierarchy
✓ Vertical and Horizontal layouts
✓ Nested layouts (containers in containers)
✓ Component naming and lookup
✓ Component enumeration
✓ Owner and Parent can be different
✓ Layout spacing and margins

Total components owned by window: X
All components will be automatically cleaned up when window is destroyed.

Demo UI created! The window should be visible.
Click buttons to see callbacks in action.
Try resizing the window to see layout updates.
```

## Troubleshooting

### Build Errors

If you get permission errors about system headers:
- Run the build command outside of any sandboxed environment
- Or use Xcode/Visual Studio directly to build

### Runtime Issues

- Make sure all platform libraries are properly linked
- On macOS, ensure Cocoa framework is available
- On Linux, ensure GTK and WebKit are installed
- On Windows, ensure proper Visual Studio runtime is installed

## Alternative: Review the Code

If you can't run the demo immediately, you can review the code in `examples/layout_demo.cpp` to see how all the features are used. The code is well-commented and demonstrates:

- Creating windows and containers
- Setting up layouts (vertical and horizontal)
- Adding controls to layouts
- Component naming and lookup
- Owner/Parent relationships
- Button callbacks
- Form handling
