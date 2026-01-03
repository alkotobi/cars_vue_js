<script setup>
import { ref } from 'vue'
import UsersToolbar from '../components/users/UsersToolbar.vue'
import UsersCrud from '../components/UsersCrud.vue'
import AddUserForm from '../components/users/AddUserForm.vue'
import RolesViewPanel from '../components/users/RolesViewPanel.vue'

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
</style>
