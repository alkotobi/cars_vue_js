#ifndef COMPONENT_H
#define COMPONENT_H

#include <string>
#include <vector>
#include <memory>

// Forward declaration
class Component;

// Base class for all components (equivalent to TComponent in Delphi)
// Handles ownership and automatic cleanup
class Component {
public:
    Component(Component* owner = nullptr);
    virtual ~Component();
    
    // Non-copyable, movable
    Component(const Component&) = delete;
    Component& operator=(const Component&) = delete;
    Component(Component&&) noexcept;
    Component& operator=(Component&&) noexcept;
    
    // Owner management
    Component* GetOwner() const { return owner_; }
    void SetOwner(Component* owner);
    
    // Component naming
    const std::string& GetName() const { return name_; }
    void SetName(const std::string& name);
    
    // Component lookup
    Component* FindComponent(const std::string& name) const;
    
    // Component enumeration
    int GetComponentCount() const;
    Component* GetComponent(int index) const;
    std::vector<Component*> GetComponents() const;
    
    // Check if component is being destroyed
    bool IsDestroying() const { return destroying_; }
    
protected:
    // Called when component is being destroyed (before owned components are destroyed)
    virtual void Notification(Component* component, bool inserting);
    
    // Called when component is inserted/removed from owner
    virtual void Inserted();
    virtual void Removed();
    
private:
    Component* owner_;
    std::string name_;
    std::vector<Component*> ownedComponents_;
    bool destroying_;
    
    void addOwnedComponent(Component* component);
    void removeOwnedComponent(Component* component);
    void validateName(const std::string& name) const;
    std::string generateDefaultName() const;
};

#endif // COMPONENT_H
