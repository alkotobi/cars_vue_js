<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import TransfersInterTable from '../components/cashier/TransfersInterTable.vue'
import MoneyMovements from '../components/cashier/MoneyMovements.vue'
import ChinaCash from '../components/cashier/ChinaCash.vue'

const router = useRouter()
const isProcessing = ref({
  dashboard: false,
  transfers: false,
  moneyMovements: false,
  chinaCash: false,
})

const goToDashboard = async () => {
  if (isProcessing.value.dashboard) return
  isProcessing.value.dashboard = true
  try {
    await router.push('/dashboard')
  } finally {
    isProcessing.value.dashboard = false
  }
}

const goToMoneyMovements = async () => {
  if (isProcessing.value.moneyMovements) return
  isProcessing.value.moneyMovements = true
  try {
    await router.push('/cashier/money-movements')
  } finally {
    isProcessing.value.moneyMovements = false
  }
}

const goToTransfers = async () => {
  if (isProcessing.value.transfers) return
  isProcessing.value.transfers = true
  try {
    await router.push('/cashier/transfers')
  } finally {
    isProcessing.value.transfers = false
  }
}

const goToChinaCash = async () => {
  if (isProcessing.value.chinaCash) return
  isProcessing.value.chinaCash = true
  try {
    await router.push('/cashier/china-cash')
  } finally {
    isProcessing.value.chinaCash = false
  }
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
      <h1 class="sidebar-title">
        <i class="fas fa-cash-register"></i>
        Cashier
      </h1>

      <button
        @click="goToDashboard"
        class="sidebar-btn"
        :disabled="isProcessing.dashboard"
        :class="{ processing: isProcessing.dashboard }"
      >
        <i class="fas fa-arrow-left"></i>
        <span>Dashboard</span>
        <i v-if="isProcessing.dashboard" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>

      <button
        @click="goToTransfers"
        class="sidebar-btn transfers-btn"
        :disabled="isProcessing.transfers"
        :class="{ processing: isProcessing.transfers }"
      >
        <i class="fas fa-exchange-alt"></i>
        <span>Transfers</span>
        <i v-if="isProcessing.transfers" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>

      <button
        @click="goToMoneyMovements"
        class="sidebar-btn movements-btn"
        :disabled="isProcessing.moneyMovements"
        :class="{ processing: isProcessing.moneyMovements }"
      >
        <i class="fas fa-money-bill-wave"></i>
        <span>Money Movements</span>
        <i v-if="isProcessing.moneyMovements" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>

      <button
        v-if="isAdmin"
        @click="goToChinaCash"
        class="sidebar-btn china-btn"
        :disabled="isProcessing.chinaCash"
        :class="{ processing: isProcessing.chinaCash }"
      >
        <i class="fas fa-yen-sign"></i>
        <span>China Cash</span>
        <i v-if="isProcessing.chinaCash" class="fas fa-spinner fa-spin loading-indicator"></i>
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
  width: 280px;
  background-color: #1f2937;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  box-shadow: 4px 0 8px rgba(0, 0, 0, 0.1);
}

.sidebar-title {
  color: white;
  font-size: 1.5rem;
  margin: 0 0 16px 0;
  padding-bottom: 16px;
  border-bottom: 2px solid #374151;
  display: flex;
  align-items: center;
  gap: 12px;
}

.sidebar-title i {
  color: #60a5fa;
}

.main-content {
  flex: 1;
  padding: 24px;
  background-color: #f3f4f6;
  overflow-y: auto;
}

.sidebar-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 14px 16px;
  border: none;
  border-radius: 8px;
  background-color: transparent;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

.sidebar-btn:not(:disabled):hover {
  background-color: #374151;
  transform: translateX(4px);
}

.sidebar-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.sidebar-btn i:not(.loading-indicator) {
  width: 20px;
  text-align: center;
  font-size: 1.1em;
}

.sidebar-btn span {
  flex: 1;
  text-align: left;
}

.loading-indicator {
  margin-left: auto;
  font-size: 0.9em;
}

.sidebar-btn.processing {
  position: relative;
  pointer-events: none;
}

.sidebar-btn.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

/* Button variants */
.transfers-btn:not(:disabled):hover {
  background-color: #1d4ed8;
}

.movements-btn:not(:disabled):hover {
  background-color: #047857;
}

.china-btn:not(:disabled):hover {
  background-color: #b91c1c;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}

@media (max-width: 768px) {
  .cashier-view {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    padding: 16px;
  }

  .sidebar-btn {
    padding: 12px;
  }

  .main-content {
    padding: 16px;
  }
}
</style>
