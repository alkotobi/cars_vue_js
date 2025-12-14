<template>
  <div class="sidebar" :class="{ collapsed: isCollapsed }">
    <div class="sidebar-header">
      <h3 v-show="!isCollapsed">Database Manager</h3>
      <button @click="toggleSidebar" class="toggle-button" :title="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'">
        <i :class="isCollapsed ? 'fas fa-chevron-right' : 'fas fa-chevron-left'"></i>
      </button>
    </div>
    <nav class="sidebar-nav">
      <button
        @click="selectMenuItem('sql')"
        :class="{ active: activeItem === 'sql' }"
        class="nav-button"
        :title="isCollapsed ? 'SQL' : ''"
      >
        <i class="fas fa-database"></i>
        <span v-show="!isCollapsed">SQL</span>
      </button>
      <button
        @click="selectMenuItem('databases')"
        :class="{ active: activeItem === 'databases' }"
        class="nav-button"
        :title="isCollapsed ? 'Databases' : ''"
      >
        <i class="fas fa-server"></i>
        <span v-show="!isCollapsed">Databases</span>
      </button>
      <button
        @click="selectMenuItem('update-structure')"
        :class="{ active: activeItem === 'update-structure' }"
        class="nav-button"
        :title="isCollapsed ? 'Update DB Structure' : ''"
      >
        <i class="fas fa-sync-alt"></i>
        <span v-show="!isCollapsed">Update DB Structure</span>
      </button>
    </nav>
    <div class="sidebar-footer">
      <button @click="logout" class="logout-button" :title="isCollapsed ? 'Logout' : ''">
        <i class="fas fa-sign-out-alt"></i>
        <span v-show="!isCollapsed">Logout</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, defineEmits, defineProps } from 'vue'

const props = defineProps({
  activeItem: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['select-item', 'logout', 'toggle'])

const isCollapsed = ref(true) // Default to collapsed (minimal)

const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value
  emit('toggle', isCollapsed.value)
}

const selectMenuItem = (item) => {
  emit('select-item', item)
}

const logout = () => {
  localStorage.removeItem('db_manager_user')
  emit('logout')
}
</script>

<style scoped>
.sidebar {
  width: 250px;
  height: calc(100vh - 70px);
  background-color: #2c3e50;
  color: white;
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 70px;
  box-shadow: 2px 0 5px rgba(0, 0, 0, 0.1);
  transition: width 0.3s ease;
  overflow: hidden;
  z-index: 1000;
}

.sidebar.collapsed {
  width: 70px;
}

@media (max-width: 768px) {
  .sidebar {
    height: calc(100vh - 60px);
    top: 60px;
  }
}

.sidebar-header {
  padding: 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  min-height: 60px;
}

.sidebar.collapsed .sidebar-header {
  padding: 1.5rem 0.75rem;
  justify-content: center;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: white;
  white-space: nowrap;
  opacity: 1;
  transition: opacity 0.2s ease;
}

.sidebar.collapsed .sidebar-header h3 {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.toggle-button {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.8);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.toggle-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: white;
}

.sidebar.collapsed .toggle-button {
  margin: 0 auto;
}

.sidebar-nav {
  flex: 1;
  padding: 1rem 0;
  overflow-y: auto;
}

.nav-button {
  width: 100%;
  padding: 1rem 1.5rem;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.8);
  text-align: left;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  transition: all 0.3s ease;
  font-size: 1rem;
  position: relative;
}

.sidebar.collapsed .nav-button {
  padding: 1rem;
  justify-content: center;
}

.nav-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: white;
}

.nav-button.active {
  background-color: #409eff;
  color: white;
}

.nav-button i {
  width: 20px;
  text-align: center;
  flex-shrink: 0;
}

.nav-button span {
  white-space: nowrap;
  opacity: 1;
  transition: opacity 0.2s ease;
}

.sidebar.collapsed .nav-button span {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.logout-button {
  width: 100%;
  padding: 0.75rem 1.5rem;
  background-color: #e74c3c;
  border: none;
  color: white;
  text-align: left;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  transition: all 0.3s ease;
  border-radius: 4px;
  font-size: 1rem;
  position: relative;
}

.sidebar.collapsed .logout-button {
  padding: 0.75rem;
  justify-content: center;
}

.logout-button:hover {
  background-color: #c0392b;
}

.logout-button i {
  width: 20px;
  text-align: center;
  flex-shrink: 0;
}

.logout-button span {
  white-space: nowrap;
  opacity: 1;
  transition: opacity 0.2s ease;
}

.sidebar.collapsed .logout-button span {
  opacity: 0;
  width: 0;
  overflow: hidden;
}
</style>
