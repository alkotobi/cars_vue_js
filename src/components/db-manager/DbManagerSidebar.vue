<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <h3>Database Manager</h3>
    </div>
    <nav class="sidebar-nav">
      <button
        @click="selectMenuItem('sql')"
        :class="{ active: activeItem === 'sql' }"
        class="nav-button"
      >
        <i class="fas fa-database"></i>
        <span>SQL</span>
      </button>
      <button
        @click="selectMenuItem('databases')"
        :class="{ active: activeItem === 'databases' }"
        class="nav-button"
      >
        <i class="fas fa-server"></i>
        <span>Databases</span>
      </button>
    </nav>
    <div class="sidebar-footer">
      <button @click="logout" class="logout-button">
        <i class="fas fa-sign-out-alt"></i>
        <span>Logout</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { defineEmits, defineProps } from 'vue'

const props = defineProps({
  activeItem: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['select-item', 'logout'])

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
}

.sidebar-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: white;
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
}

.logout-button:hover {
  background-color: #c0392b;
}

.logout-button i {
  width: 20px;
  text-align: center;
}
</style>
