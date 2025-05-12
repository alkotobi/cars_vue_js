<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import CarsStock from './CarsStock.vue'
import ClientsView from './ClientsView.vue'
import BrokersView from './BrokersView.vue'
import SuppliersView from './SuppliersView.vue'

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
      <button 
        @click="returnToDashboard"
        class="sidebar-btn return-btn"
      >
        Return to Dashboard
      </button>
      <button 
        @click="navigateTo('stock')"
        :class="{ active: activeView === 'stock' }"
        class="sidebar-btn"
      >
        Cars Stock
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
    <div class="main-content">
      <h1>Cars Management</h1>
      <div class="content">
        <div v-if="!activeView" class="empty-state">
          Please select an option from the sidebar
        </div>
        <CarsStock v-if="activeView === 'stock'" />
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
}

.main-content {
  flex: 1;
  padding: 20px;
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
</style>