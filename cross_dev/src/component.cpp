#include "../include/component.h"
#include <algorithm>
#include <sstream>
#include <stdexcept>
#include <typeinfo>
#include <cstring>

Component::Component(Component* owner)
    : owner_(nullptr), name_(""), destroying_(false) {
    SetOwner(owner);
}

Component::~Component() {
    destroying_ = true;
    
    // Destroy all owned components (in reverse order)
    // This ensures dependencies are cleaned up properly
    while (!ownedComponents_.empty()) {
        Component* comp = ownedComponents_.back();
        ownedComponents_.pop_back();
        delete comp;
    }
    
    // Remove self from owner's list
    if (owner_) {
        owner_->removeOwnedComponent(this);
    }
}

Component::Component(Component&& other) noexcept
    : owner_(other.owner_),
      name_(std::move(other.name_)),
      ownedComponents_(std::move(other.ownedComponents_)),
      destroying_(other.destroying_) {
    // Update owner's list
    if (owner_) {
        owner_->removeOwnedComponent(&other);
        owner_->addOwnedComponent(this);
    }
    
    // Update owned components' owner pointers
    for (auto* comp : ownedComponents_) {
        comp->owner_ = this;
    }
    
    other.owner_ = nullptr;
    other.ownedComponents_.clear();
    other.destroying_ = false;
}

Component& Component::operator=(Component&& other) noexcept {
    if (this != &other) {
        // Clean up current state
        destroying_ = true;
        while (!ownedComponents_.empty()) {
            Component* comp = ownedComponents_.back();
            ownedComponents_.pop_back();
            delete comp;
        }
        if (owner_) {
            owner_->removeOwnedComponent(this);
        }
        
        // Move from other
        owner_ = other.owner_;
        name_ = std::move(other.name_);
        ownedComponents_ = std::move(other.ownedComponents_);
        destroying_ = other.destroying_;
        
        // Update owner's list
        if (owner_) {
            owner_->removeOwnedComponent(&other);
            owner_->addOwnedComponent(this);
        }
        
        // Update owned components' owner pointers
        for (auto* comp : ownedComponents_) {
            comp->owner_ = this;
        }
        
        other.owner_ = nullptr;
        other.ownedComponents_.clear();
        other.destroying_ = false;
    }
    return *this;
}

void Component::SetOwner(Component* owner) {
    if (owner_ == owner) {
        return;
    }
    
    // Validate: can't own self or create circular ownership
    if (owner == this) {
        throw std::runtime_error("Component cannot own itself");
    }
    
    // Check for circular ownership
    Component* check = owner;
    while (check) {
        if (check == this) {
            throw std::runtime_error("Circular ownership detected");
        }
        check = check->owner_;
    }
    
    // Remove from old owner
    if (owner_) {
        owner_->removeOwnedComponent(this);
        Removed();
    }
    
    // Add to new owner
    owner_ = owner;
    if (owner_) {
        owner_->addOwnedComponent(this);
        Inserted();
    }
}

void Component::SetName(const std::string& name) {
    if (name_ == name) {
        return;
    }
    
    validateName(name);
    
    // Check if name is already used by sibling
    if (owner_) {
        Component* existing = owner_->FindComponent(name);
        if (existing && existing != this) {
            throw std::runtime_error("Component name already exists: " + name);
        }
    }
    
    name_ = name;
}

Component* Component::FindComponent(const std::string& name) const {
    if (name.empty()) {
        return nullptr;
    }
    
    // Search in owned components
    for (Component* comp : ownedComponents_) {
        if (comp->name_ == name) {
            return comp;
        }
        // Recursive search
        Component* found = comp->FindComponent(name);
        if (found) {
            return found;
        }
    }
    
    return nullptr;
}

int Component::GetComponentCount() const {
    return static_cast<int>(ownedComponents_.size());
}

Component* Component::GetComponent(int index) const {
    if (index < 0 || index >= static_cast<int>(ownedComponents_.size())) {
        return nullptr;
    }
    return ownedComponents_[index];
}

std::vector<Component*> Component::GetComponents() const {
    return ownedComponents_;
}

void Component::Notification(Component* component, bool inserting) {
    // Override in derived classes if needed
    (void)component;
    (void)inserting;
}

void Component::Inserted() {
    // Override in derived classes if needed
}

void Component::Removed() {
    // Override in derived classes if needed
}

void Component::addOwnedComponent(Component* component) {
    if (!component) {
        return;
    }
    
    // Check if already added
    auto it = std::find(ownedComponents_.begin(), ownedComponents_.end(), component);
    if (it != ownedComponents_.end()) {
        return;
    }
    
    ownedComponents_.push_back(component);
    Notification(component, true);
}

void Component::removeOwnedComponent(Component* component) {
    if (!component) {
        return;
    }
    
    auto it = std::find(ownedComponents_.begin(), ownedComponents_.end(), component);
    if (it != ownedComponents_.end()) {
        Notification(component, false);
        ownedComponents_.erase(it);
    }
}

void Component::validateName(const std::string& name) const {
    // Basic validation: non-empty, no special characters that might cause issues
    if (name.empty()) {
        return; // Empty name is allowed (will use default)
    }
    
    // Check for invalid characters (basic check)
    for (char c : name) {
        if (c == '\0' || c == '\n' || c == '\r') {
            throw std::runtime_error("Component name contains invalid characters");
        }
    }
}

std::string Component::generateDefaultName() const {
    // Generate default name based on class name
    // Use typeid to get class name (requires RTTI)
    const char* typeName = typeid(*this).name();
    
    // Demangle name (platform-specific)
    #ifdef __GNUC__
        // GCC/Clang: typeid returns mangled name like "6Button"
        // Skip numbers at the start
        while (*typeName >= '0' && *typeName <= '9') {
            typeName++;
        }
    #elif defined(_MSC_VER)
        // MSVC: typeid returns class name like "class Button"
        // Skip "class " prefix if present
        if (strncmp(typeName, "class ", 6) == 0) {
            typeName += 6;
        }
    #endif
    
    // Find next available number for this type
    std::string baseName = typeName;
    int counter = 1;
    std::string candidateName;
    
    do {
        std::ostringstream oss;
        oss << baseName << counter;
        candidateName = oss.str();
        counter++;
        
        // Check if name already exists (only check up to reasonable limit)
        if (owner_) {
            Component* existing = owner_->FindComponent(candidateName);
            if (!existing) {
                break; // Found available name
            }
        } else {
            break; // No owner, use first available
        }
    } while (counter < 10000); // Safety limit
    
    return candidateName;
}
