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
      <button @click="goToUsers" class="back-btn">
        <i class="fas fa-arrow-left"></i>
        Return to Users
      </button>
      <h1>
        <i class="fas fa-shield-alt"></i>
        Roles & Permissions Management
      </h1>
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
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 24px;
  background: white;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.header h1 {
  margin: 0;
  font-size: 24px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h1 i {
  color: #3b82f6;
  font-size: 28px;
}

.back-btn {
  padding: 10px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.back-btn i {
  font-size: 14px;
  transition: transform 0.2s ease;
}

.back-btn:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.back-btn:hover i {
  transform: translateX(-2px);
}

.back-btn:active {
  transform: translateY(0);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.crud-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 24px;
  margin-top: 24px;
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .back-btn {
    width: 100%;
    justify-content: center;
  }

  .crud-section {
    grid-template-columns: 1fr;
  }
}
</style>
