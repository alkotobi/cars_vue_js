<script setup>
defineProps({
  showAddForm: {
    type: Boolean,
    default: false,
  },
  showRoles: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['add-user', 'show-roles'])
</script>

<template>
  <div class="users-toolbar">
    <div class="toolbar-left">
      <h2 class="toolbar-title">
        <i class="fas fa-users"></i>
        Users Management
      </h2>
    </div>
    <div class="toolbar-right">
      <button
        @click="emit('add-user')"
        class="toolbar-btn add-user-btn"
        :class="{ active: showAddForm }"
      >
        <i class="fas fa-user-plus"></i>
        Add User
      </button>
      <button
        @click="emit('show-roles')"
        class="toolbar-btn roles-btn"
        :class="{ active: showRoles }"
      >
        <i class="fas fa-shield-alt"></i>
        Roles
      </button>
    </div>
  </div>
</template>

<style scoped>
.users-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 24px;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-title i {
  color: #3b82f6;
}

.toolbar-right {
  display: flex;
  gap: 12px;
}

.toolbar-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.toolbar-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.toolbar-btn:active {
  transform: translateY(0);
}

.add-user-btn {
  background-color: #10b981;
  color: white;
}

.add-user-btn:hover {
  background-color: #059669;
}

.add-user-btn.active {
  background-color: #047857;
  box-shadow: 0 2px 4px rgba(16, 185, 129, 0.3);
}

.roles-btn {
  background-color: #6366f1;
  color: white;
}

.roles-btn:hover {
  background-color: #4f46e5;
}

.roles-btn.active {
  background-color: #4338ca;
  box-shadow: 0 2px 4px rgba(99, 102, 241, 0.3);
}

@media (max-width: 768px) {
  .users-toolbar {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .toolbar-right {
    flex-direction: column;
  }

  .toolbar-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>

