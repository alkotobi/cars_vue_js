<script setup>
import { useRouter } from 'vue-router'
import UsersCrud from '../components/UsersCrud.vue'

const router = useRouter()
const goToRoles = () => {
  router.push('/roles')
}
const goToDashboard = () => {
  router.push('/')
}
</script>

<template>
  <div class="users-page">
    <div class="header-buttons">
      <button @click="goToDashboard" class="dashboard-btn">← Return to Dashboard</button>
    </div>
    <UsersCrud />
    <button @click="goToRoles" class="roles-btn">Manage Roles</button>
  </div>
</template>

<style scoped>
.users-page {
  padding: 20px;
}

.header-buttons {
  margin-bottom: 20px;
}

.dashboard-btn {
  padding: 10px 20px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1em;
  transition: background-color 0.3s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dashboard-btn:hover {
  background-color: #2563eb;
}

.roles-btn {
  margin-top: 20px;
  padding: 10px 20px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1em;
  transition: background-color 0.3s;
}

.roles-btn:hover {
  background-color: #388e3c;
}
</style>