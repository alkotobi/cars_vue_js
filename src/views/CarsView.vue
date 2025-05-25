<script setup>
import { ref ,computed} from 'vue'
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
import WarehousesView from './WarehousesView.vue'  // Add this import

const router = useRouter()
const activeView = ref(null)

const navigateTo = (view) => {
  activeView.value = view
}

const returnToDashboard = () => {
  router.push('/')
}
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  if (currentUser.value) {
  return currentUser.value?.role_id === 1
}else
  {
    router.push('/login')
    return false
  }
}
)

const canPurchaseCars = computed(() => {
  if (isAdmin.value){
    return true
  }
  return user.value.permissions?.some(p => p.permission_name === 'can_purchase_cars')
})

const canSellCars = computed(() => {
  if (isAdmin.value){
    return true
  }
  return user.value.permissions?.some(p => p.permission_name === 'can_sell_cars')
})

const canCCarStock = computed(() => {
  if (isAdmin.value){
    return true
  }
  return user.value.permissions?.some(p => p.permission_name === 'can_c_car_stock')
})
</script>

<template>
  <div class="cars-view">
    <div class="sidebar">
      <div class="sidebar-top">
        <button 
          @click="returnToDashboard"
          class="sidebar-btn return-btn"
        >
          Return to Dashboard
        </button>
        <button 
          :disabled="!canPurchaseCars"
          @click="navigateTo('buy')"
          :class="{ active: activeView === 'buy' }"
          class="sidebar-btn buy-btn"
        >
          Purchases
        </button>
        <button v-if="false"
          @click="navigateTo('sell')"
          :class="{ active: activeView === 'sell' }"
          class="sidebar-btn sell-btn"
        >
          Sell
        </button>
        <button 
        :disabled="!canSellCars"
          @click="navigateTo('sell-bills')"
          :class="{ active: activeView === 'sell-bills' }"
          class="sidebar-btn sell-bills-btn"
        >
          Sells
        </button>
        <button 
          :disabled="!canCCarStock"
          @click="navigateTo('stock')"
          :class="{ active: activeView === 'stock' }"
          class="sidebar-btn stock-btn"
        >
          Cars Stock
        </button>
      </div>

      <div class="sidebar-bottom">
        <div class="sidebar-section-title">Settings</div>
        <button 
        v-if="canPurchaseCars"
          @click="navigateTo('models')"
          :class="{ active: activeView === 'models' }"
          class="sidebar-btn"
        >
          Car Models
        </button>
        <button 
        v-if="canPurchaseCars"
          @click="navigateTo('colors')"
          :class="{ active: activeView === 'colors' }"
          class="sidebar-btn"
        >
          Colors
        </button>
        <button 
        v-if="canPurchaseCars"
          @click="navigateTo('discharge-ports')"
          :class="{ active: activeView === 'discharge-ports' }"
          class="sidebar-btn"
        >
          Discharge Ports
        </button>
        <button 
        v-if="canPurchaseCars"
          @click="navigateTo('loading-ports')"
          :class="{ active: activeView === 'loading-ports' }"
          class="sidebar-btn"
        >
          Loading Ports
        </button>
        <button 
        v-if="canSellCars"
          @click="navigateTo('clients')"
          :class="{ active: activeView === 'clients' }"
          class="sidebar-btn"
        >
          Clients
        </button>
        <button 
        v-if="canSellCars"
          @click="navigateTo('brokers')"
          :class="{ active: activeView === 'brokers' }"
          class="sidebar-btn"
        >
          Brokers
        </button>
        <button 
        v-if="canPurchaseCars"
          @click="navigateTo('suppliers')"
          :class="{ active: activeView === 'suppliers' }"
          class="sidebar-btn"
        >
          Suppliers
        </button>
        <!-- Add Warehouses button at the bottom of the sidebar -->
        <button 
        v-if="canPurchaseCars"
          @click="navigateTo('warehouses')"
          :class="{ active: activeView === 'warehouses' }"
          class="sidebar-btn"
        >
          Warehouses
        </button>
      </div>
    </div>
    <div class="main-content">
      <h1>Cars Management</h1>
      <div class="content">
        <div v-if="!activeView" class="empty-state">
          Please select an option from the sidebar
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
        <!-- Add conditional rendering for Warehouses view -->
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