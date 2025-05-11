<script setup>
import { useRouter } from 'vue-router'
import RolesCrud from '../components/RolesCrud.vue'
import PermissionsCrud from '../components/PermissionsCrud.vue'
import RolePermissionsCrud from '../components/RolePermissionsCrud.vue'

const router = useRouter()
const goToUsers = () => {
  router.push('/users')
}
</script>

<template>
  <div class="roles-page">
    <div class="header">
      <button @click="goToUsers" class="back-btn">← Return to Users</button>
      <h1>Roles, Permissions, and Role Permissions Management</h1>
    </div>
    <div class="crud-section">
      <RolesCrud />
      <PermissionsCrud />
      <RolePermissionsCrud />
    </div>
  </div>
</template>

<style scoped>
.roles-page {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 20px;
}

.back-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.back-btn:hover {
  background-color: #2563eb;
}

.crud-section {
  display: flex;
  flex-wrap: wrap;
  gap: 40px;
  margin-top: 30px;
}
</style>