# Generating iOS Xcode Project with CMake

## Quick Start

Due to sandbox restrictions, CMake's automatic project generation may have issues. Here's the recommended approach:

### Option 1: Generate from Terminal (Outside Sandbox)

Run this command from a terminal (not from within the sandboxed environment):

```bash
cd ios
mkdir -p build_cmake
cd build_cmake
cmake ../.. \
    -G Xcode \
    -DCMAKE_SYSTEM_NAME=iOS \
    -DCMAKE_OSX_DEPLOYMENT_TARGET=12.0 \
    -DCMAKE_OSX_ARCHITECTURES="arm64" \
    -DIOS=ON

# Then open the generated project
open CrossDev.xcodeproj
```

### Option 2: Use the Script (May Require Permissions)

```bash
cd ios
./create_xcode_project_cmake.sh
```

**Note:** If you get permission errors about DerivedData, you may need to run this outside of a sandboxed environment, or manually create the Xcode project using Xcode's "New Project" wizard and add the source files.

### Option 3: Manual Xcode Project Creation

1. Open Xcode
2. File → New → Project
3. Choose "App" template
4. Configure:
   - Product Name: `NativeWindow`
   - Team: Your development team
   - Organization Identifier: `com.nativewindow`
   - Language: Objective-C++
   - Interface: Storyboard (or None)
5. Add all source files from:
   - `src/` (all .cpp files)
   - `src/platform/ios/` (all .mm files)
   - `include/` (all .h files)
6. Add frameworks:
   - UIKit.framework
   - Foundation.framework
   - WebKit.framework
7. Set deployment target to iOS 12.0
8. Configure Info.plist (copy from `ios/Info.plist`)

This manual approach is the most reliable if automated generation fails.
