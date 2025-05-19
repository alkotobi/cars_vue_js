<script setup>
import { ref } from 'vue'
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

const router = useRouter()
const activeView = ref(null)

const navigateTo = (view) => {
  activeView.value = view
}

const returnToDashboard = () => {
  router.push('/')
}
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
          @click="navigateTo('buy')"
          :class="{ active: activeView === 'buy' }"
          class="sidebar-btn buy-btn"
        >
          Buy
        </button>
        <button 
          @click="navigateTo('sell')"
          :class="{ active: activeView === 'sell' }"
          class="sidebar-btn sell-btn"
        >
          Sell
        </button>
        <button 
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
          @click="navigateTo('models')"
          :class="{ active: activeView === 'models' }"
          class="sidebar-btn"
        >
          Car Models
        </button>
        <button 
          @click="navigateTo('colors')"
          :class="{ active: activeView === 'colors' }"
          class="sidebar-btn"
        >
          Colors
        </button>
        <button 
          @click="navigateTo('discharge-ports')"
          :class="{ active: activeView === 'discharge-ports' }"
          class="sidebar-btn"
        >
          Discharge Ports
        </button>
        <button 
          @click="navigateTo('loading-ports')"
          :class="{ active: activeView === 'loading-ports' }"
          class="sidebar-btn"
        >
          Loading Ports
        </button>
        <button 
          @click="navigateTo('clients')"
          :class="{ active: activeView === 'clients' }"
          class="sidebar-btn"
        >
          Clients
        </button>
        <button 
          @click="navigateTo('brokers')"
          :class="{ active: activeView === 'brokers' }"
          class="sidebar-btn"
        >
          Brokers
        </button>
        <button 
          @click="navigateTo('suppliers')"
          :class="{ active: activeView === 'suppliers' }"
          class="sidebar-btn"
        >
          Suppliers
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
        <CarsStock v-if="activeView === 'stock'" />
        <CarModelsView v-if="activeView === 'models'" />
        <ColorsView v-if="activeView === 'colors'" />
        <DischargePortsView v-if="activeView === 'discharge-ports'" />
        <LoadingPortsView v-if="activeView === 'loading-ports'" />
        <ClientsView v-if="activeView === 'clients'" />
        <BrokersView v-if="activeView === 'brokers'" />
        <SuppliersView v-if="activeView === 'suppliers'" />
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
  width: 200px;
  background-color: #2c3e50;
  padding: 20px;
  color: white;
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.sidebar-top {
  border-bottom: 2px solid #34495e;
  padding-bottom: 20px;
}

.sidebar-bottom {
  flex: 1;
}

.sidebar-section-title {
  font-size: 0.9em;
  text-transform: uppercase;
  color: #94a3b8;
  margin-bottom: 10px;
  padding-left: 10px;
}

.sidebar-btn {
  width: 100%;
  padding: 10px;
  margin: 5px 0;
  background: none;
  border: none;
  color: white;
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.buy-btn {
  background-color: #10b981;
}

.buy-btn:hover {
  background-color: #059669;
}

.buy-btn.active {
  background-color: #047857;
}

.stock-btn {
  background-color: #3b82f6;
}

.stock-btn:hover {
  background-color: #2563eb;
}

.stock-btn.active {
  background-color: #1d4ed8;
}

.main-content {
  flex: 1;
  padding: 20px;
}

.sidebar-btn:hover {
  background-color: #34495e;
}

.sidebar-btn.active {
  background-color: #3498db;
}

.content {
  margin-top: 20px;
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

h1 {
  margin: 0;
  color: #2c3e50;
}

h2 {
  color: #2c3e50;
  margin-top: 0;
}
.empty-state {
  text-align: center;
  color: #666;
  padding: 40px;
  font-size: 1.2em;
}

.return-btn {
  margin-bottom: 20px;
  background-color: #64748b;
}

.return-btn:hover {
  background-color: #475569;
}

.buy-btn {
  background-color: #10b981;
  margin-bottom: 20px;
}

.buy-btn:hover {
  background-color: #059669;
}

.buy-btn.active {
  background-color: #047857;
}

.sell-btn {
  background-color: #ef4444;
  margin-bottom: 20px;
}

.sell-btn:hover {
  background-color: #dc2626;
}

.sell-btn.active {
  background-color: #b91c1c;
}
</style>