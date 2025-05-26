<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import CarsStock from './CarsStock.vue'
import ClientsView from './ClientsView.vue'
import BrokersView from './BrokersView.vue'
import SuppliersView from './SuppliersView.vue'
import CarModelsView from './CarModelsView.vue'
import ColorsView from './ColorsView.vue'
import DischargePortsView from './DischargePortsView.vue'
import LoadingPortsView from './LoadingPortsView.vue'
import BuyView from './BuyView.vue'
import SellView from './SellView.vue'
import SellBillsView from './SellBillsView.vue'
import WarehousesView from './WarehousesView.vue' // Add this import

const router = useRouter()
const activeView = ref(null)
const isLoading = ref(false)
const isProcessing = ref({
  dashboard: false,
  buy: false,
  sell: false,
  sellBills: false,
  stock: false,
  models: false,
  colors: false,
  dischargePorts: false,
  loadingPorts: false,
  clients: false,
  brokers: false,
  suppliers: false,
  warehouses: false,
})

const navigateTo = async (view) => {
  if (isProcessing.value[view]) return
  isProcessing.value[view] = true
  try {
    activeView.value = view
  } finally {
    isProcessing.value[view] = false
  }
}

const returnToDashboard = async () => {
  if (isProcessing.value.dashboard) return
  isProcessing.value.dashboard = true
  try {
    await router.push('/')
  } finally {
    isProcessing.value.dashboard = false
  }
}
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  if (currentUser.value) {
    return currentUser.value?.role_id === 1
  } else {
    router.push('/login')
    return false
  }
})

const canPurchaseCars = computed(() => {
  if (isAdmin.value) {
    return true
  }
  return user.value.permissions?.some((p) => p.permission_name === 'can_purchase_cars')
})

const canSellCars = computed(() => {
  if (isAdmin.value) {
    return true
  }
  return user.value.permissions?.some((p) => p.permission_name === 'can_sell_cars')
})

const canCCarStock = computed(() => {
  if (isAdmin.value) {
    return true
  }
  return user.value.permissions?.some((p) => p.permission_name === 'can_c_car_stock')
})
</script>

<template>
  <div class="cars-view" :class="{ 'is-loading': isLoading }">
    <div class="sidebar">
      <div class="sidebar-top">
        <button
          @click="returnToDashboard"
          class="sidebar-btn return-btn"
          :disabled="isProcessing.dashboard"
          :class="{ processing: isProcessing.dashboard }"
        >
          <i class="fas fa-arrow-left"></i>
          <span>Return to Dashboard</span>
          <i v-if="isProcessing.dashboard" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          :disabled="!canPurchaseCars || isProcessing.buy"
          @click="navigateTo('buy')"
          :class="{ active: activeView === 'buy', processing: isProcessing.buy }"
          class="sidebar-btn buy-btn"
        >
          <i class="fas fa-shopping-cart"></i>
          <span>Purchases</span>
          <i v-if="isProcessing.buy" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          v-if="false"
          @click="navigateTo('sell')"
          :class="{ active: activeView === 'sell', processing: isProcessing.sell }"
          class="sidebar-btn sell-btn"
          :disabled="isProcessing.sell"
        >
          <i class="fas fa-dollar-sign"></i>
          <span>Sell</span>
          <i v-if="isProcessing.sell" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          :disabled="!canSellCars || isProcessing.sellBills"
          @click="navigateTo('sell-bills')"
          :class="{ active: activeView === 'sell-bills', processing: isProcessing.sellBills }"
          class="sidebar-btn sell-bills-btn"
        >
          <i class="fas fa-file-invoice-dollar"></i>
          <span>Sells</span>
          <i v-if="isProcessing.sellBills" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          :disabled="!canCCarStock || isProcessing.stock"
          @click="navigateTo('stock')"
          :class="{ active: activeView === 'stock', processing: isProcessing.stock }"
          class="sidebar-btn stock-btn"
        >
          <i class="fas fa-warehouse"></i>
          <span>Cars Stock</span>
          <i v-if="isProcessing.stock" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>

      <div class="sidebar-bottom">
        <div class="sidebar-section-title"><i class="fas fa-cog"></i> Settings</div>
        <button
          v-if="canPurchaseCars"
          @click="navigateTo('models')"
          :class="{ active: activeView === 'models', processing: isProcessing.models }"
          class="sidebar-btn"
          :disabled="isProcessing.models"
        >
          <i class="fas fa-car"></i>
          <span>Car Models</span>
          <i v-if="isProcessing.models" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          v-if="canPurchaseCars"
          @click="navigateTo('colors')"
          :class="{ active: activeView === 'colors', processing: isProcessing.colors }"
          class="sidebar-btn"
          :disabled="isProcessing.colors"
        >
          <i class="fas fa-palette"></i>
          <span>Colors</span>
          <i v-if="isProcessing.colors" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          v-if="canPurchaseCars"
          @click="navigateTo('discharge-ports')"
          :class="{
            active: activeView === 'discharge-ports',
            processing: isProcessing.dischargePorts,
          }"
          class="sidebar-btn"
          :disabled="isProcessing.dischargePorts"
        >
          <i class="fas fa-anchor"></i>
          <span>Discharge Ports</span>
          <i
            v-if="isProcessing.dischargePorts"
            class="fas fa-spinner fa-spin loading-indicator"
          ></i>
        </button>
        <button
          v-if="canPurchaseCars"
          @click="navigateTo('loading-ports')"
          :class="{ active: activeView === 'loading-ports', processing: isProcessing.loadingPorts }"
          class="sidebar-btn"
          :disabled="isProcessing.loadingPorts"
        >
          <i class="fas fa-ship"></i>
          <span>Loading Ports</span>
          <i v-if="isProcessing.loadingPorts" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          v-if="canSellCars"
          @click="navigateTo('clients')"
          :class="{ active: activeView === 'clients', processing: isProcessing.clients }"
          class="sidebar-btn"
          :disabled="isProcessing.clients"
        >
          <i class="fas fa-users"></i>
          <span>Clients</span>
          <i v-if="isProcessing.clients" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          v-if="canSellCars"
          @click="navigateTo('brokers')"
          :class="{ active: activeView === 'brokers', processing: isProcessing.brokers }"
          class="sidebar-btn"
          :disabled="isProcessing.brokers"
        >
          <i class="fas fa-handshake"></i>
          <span>Brokers</span>
          <i v-if="isProcessing.brokers" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          v-if="canPurchaseCars"
          @click="navigateTo('suppliers')"
          :class="{ active: activeView === 'suppliers', processing: isProcessing.suppliers }"
          class="sidebar-btn"
          :disabled="isProcessing.suppliers"
        >
          <i class="fas fa-truck"></i>
          <span>Suppliers</span>
          <i v-if="isProcessing.suppliers" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          v-if="canPurchaseCars"
          @click="navigateTo('warehouses')"
          :class="{ active: activeView === 'warehouses', processing: isProcessing.warehouses }"
          class="sidebar-btn"
          :disabled="isProcessing.warehouses"
        >
          <i class="fas fa-building"></i>
          <span>Warehouses</span>
          <i v-if="isProcessing.warehouses" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>
    <div class="main-content">
      <h1>
        <i class="fas fa-car-side"></i>
        Cars Management
      </h1>
      <div class="content">
        <div v-if="!activeView" class="empty-state">
          <i class="fas fa-hand-point-left fa-2x"></i>
          <p>Please select an option from the sidebar</p>
        </div>
        <BuyView v-if="activeView === 'buy'" />
        <SellView v-if="activeView === 'sell'" />
        <SellBillsView v-if="activeView === 'sell-bills'" />
        <CarsStock v-if="activeView === 'stock'" />
        <CarModelsView v-if="activeView === 'models'" />
        <ColorsView v-if="activeView === 'colors'" />
        <DischargePortsView v-if="activeView === 'discharge-ports'" />
        <LoadingPortsView v-if="activeView === 'loading-ports'" />
        <ClientsView v-if="activeView === 'clients'" />
        <BrokersView v-if="activeView === 'brokers'" />
        <SuppliersView v-if="activeView === 'suppliers'" />
        <WarehousesView v-if="activeView === 'warehouses'" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.cars-view {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 220px;
  background-color: #2c3e50;
  padding: 20px;
  color: white;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.sidebar-top {
  border-bottom: 2px solid #34495e;
  padding-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.sidebar-bottom {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-section-title {
  font-size: 0.9em;
  text-transform: uppercase;
  color: #94a3b8;
  margin: 8px 0;
  padding-left: 10px;
  font-weight: 600;
}

.sidebar-btn {
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  color: white;
  text-align: left;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s ease;
  font-size: 0.95em;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
  outline: none;
}

.sidebar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.sidebar-btn:not(:disabled):hover {
  transform: translateX(4px);
}

.return-btn {
  background-color: #64748b;
  margin-bottom: 8px;
}

.return-btn:not(:disabled):hover {
  background-color: #475569;
}

.buy-btn {
  background-color: #10b981;
}

.buy-btn:not(:disabled):hover {
  background-color: #059669;
}

.buy-btn.active {
  background-color: #047857;
}

.sell-btn {
  background-color: #ef4444;
}

.sell-btn:not(:disabled):hover {
  background-color: #dc2626;
}

.sell-btn.active {
  background-color: #b91c1c;
}

.sell-bills-btn {
  background-color: #f59e0b;
}

.sell-bills-btn:not(:disabled):hover {
  background-color: #d97706;
}

.sell-bills-btn.active {
  background-color: #b45309;
}

.stock-btn {
  background-color: #3b82f6;
}

.stock-btn:not(:disabled):hover {
  background-color: #2563eb;
}

.stock-btn.active {
  background-color: #1d4ed8;
}

.main-content {
  flex: 1;
  padding: 24px;
  background-color: #f8fafc;
}

.sidebar-btn:not(.return-btn, .buy-btn, .sell-btn, .sell-bills-btn, .stock-btn):hover {
  background-color: #34495e;
}

.sidebar-btn:not(.return-btn, .buy-btn, .sell-btn, .sell-bills-btn, .stock-btn).active {
  background-color: #3498db;
}

.content {
  margin-top: 24px;
  background-color: white;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

h1 {
  margin: 0;
  color: #2c3e50;
  font-size: 1.75rem;
}

h2 {
  color: #2c3e50;
  margin-top: 0;
}

.empty-state {
  text-align: center;
  color: #64748b;
  padding: 48px;
  font-size: 1.1em;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.empty-state i {
  color: #94a3b8;
}

.empty-state p {
  margin: 0;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}

.cars-view.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.sidebar-btn i:not(.loading-indicator) {
  width: 20px;
  text-align: center;
}

.sidebar-btn span {
  flex: 1;
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

/* Remove duplicate styles */
.warehouses-btn {
  background-color: #10b981;
}

.warehouses-btn:not(:disabled):hover {
  background-color: #059669;
}

.warehouses-btn.active {
  background-color: #047857;
}
</style>
