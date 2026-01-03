<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import UsersToolbar from '../components/users/UsersToolbar.vue'
import UsersCrud from '../components/UsersCrud.vue'
import AddUserForm from '../components/users/AddUserForm.vue'
import RolesViewPanel from '../components/users/RolesViewPanel.vue'

const router = useRouter()

const showAddForm = ref(false)
const showRoles = ref(false)
const usersCrudRef = ref(null)

const handleAddUser = () => {
  showAddForm.value = !showAddForm.value
  if (showAddForm.value) {
    showRoles.value = false
  }
}

const handleShowRoles = () => {
  showRoles.value = !showRoles.value
  if (showRoles.value) {
    showAddForm.value = false
  }
}

const handleUserAdded = () => {
  // Refresh users list
  if (usersCrudRef.value && usersCrudRef.value.fetchUsers) {
    usersCrudRef.value.fetchUsers()
  }
  showAddForm.value = false
}

const handleCloseAddForm = () => {
  showAddForm.value = false
}

const handleCloseRoles = () => {
  showRoles.value = false
}
</script>

<template>
  <div class="users-view">
    <div class="header-section">
      <button @click="router.push('/dashboard')" class="btn-return-dashboard">
        <i class="fas fa-arrow-left"></i>
        Return to Dashboard
      </button>
    </div>

    <UsersToolbar
      :show-add-form="showAddForm"
      :show-roles="showRoles"
      @add-user="handleAddUser"
      @show-roles="handleShowRoles"
    />

    <AddUserForm
      v-if="showAddForm"
      @user-added="handleUserAdded"
      @close="handleCloseAddForm"
    />

    <RolesViewPanel v-if="showRoles" @close="handleCloseRoles" />

    <UsersCrud ref="usersCrudRef" />
  </div>
</template>

<style scoped>
.users-view {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.header-section {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
}

.btn-return-dashboard {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-return-dashboard:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.btn-return-dashboard i {
  font-size: 14px;
}
</style>
