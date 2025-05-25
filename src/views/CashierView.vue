<script setup>
import { ref,computed } from 'vue'
import { useRouter } from 'vue-router'
import TransfersInterTable from '../components/cashier/TransfersInterTable.vue'
import MoneyMovements from '../components/cashier/MoneyMovements.vue'
import ChinaCash from '../components/cashier/ChinaCash.vue'

const router = useRouter()

const goToDashboard = () => {
  router.push('/dashboard')
}

const goToMoneyMovements = () => {
  router.push('/cashier/money-movements')
}

const goToTransfers = () => {
  router.push('/cashier/transfers')
}

const goToChinaCash = () => {
  router.push('/cashier/china-cash')
}

const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})
</script>

<template>
  <div class="cashier-view">
    <div class="sidebar">
      <button @click="goToDashboard" class="sidebar-btn">
        <span class="btn-icon">←</span>
        Dashboard
      </button>
      <button @click="goToTransfers" class="sidebar-btn">
        <span class="btn-icon">↔</span>
        Transfers
      </button>
      <button @click="goToMoneyMovements" class="sidebar-btn">
        <span class="btn-icon">💰</span>
        Money Movements
      </button>
      <button v-if="isAdmin" @click="goToChinaCash" class="sidebar-btn">
        <span class="btn-icon">💴</span>
        China Cash
      </button>
    </div>
    <div class="main-content">
      <router-view v-slot="{ Component }">
        <component :is="Component" />
      </router-view>
    </div>
  </div>
</template>

<style scoped>
.cashier-view {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 250px;
  background-color: #1f2937;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.main-content {
  flex: 1;
  padding: 20px;
  background-color: #f3f4f6;
}

.sidebar-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: 6px;
  background-color: transparent;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.sidebar-btn:hover {
  background-color: #374151;
}

.btn-icon {
  font-size: 1.2em;
}
</style> 