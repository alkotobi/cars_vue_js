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
import SellBillsView from './SellBillsView.vue'
import WarehousesView from './WarehousesView.vue'
import ContainersView from './ContainersView.vue'
import StatisticsView from './StatisticsView.vue'
import LoadingView from './LoadingView.vue'
import { useApi } from '../composables/useApi'
import FinishedOrdersTable from '../components/FinishedOrdersTable.vue'

const router = useRouter()
const activeView = ref(null)
const isLoading = ref(false)
const isProcessing = ref({
  dashboard: false,
  buy: false,
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
  containers: false,
  statistics: false,
  load: false,
  loadingInquiry: false,
})

const { callApi } = useApi()
const showFinishedOrders = ref(false)
const finishedOrders = ref([])
const finishedOrdersLoading = ref(false)

// Mobile navigation state
const showMobileNav = ref(false)

const toggleMobileNav = () => {
  showMobileNav.value = !showMobileNav.value
}

const closeMobileNav = () => {
  showMobileNav.value = false
}

// Sidebar hide/show state
const sidebarCollapsed = ref(true)

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

const navigateTo = async (view) => {
  if (isProcessing.value[view]) return
  isProcessing.value[view] = true
  try {
    activeView.value = view
    // Close mobile nav when navigating
    closeMobileNav()
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
  return currentUser.value?.permissions?.some((p) => p.permission_name === 'can_purchase_cars')
})

const canSellCars = computed(() => {
  if (isAdmin.value) {
    return true
  }
  return currentUser.value?.permissions?.some((p) => p.permission_name === 'can_sell_cars')
})

const canCCarStock = computed(() => {
  if (isAdmin.value) {
    return true
  }
  return currentUser.value?.permissions?.some((p) => p.permission_name === 'can_c_car_stock')
})

const canLoadCar = computed(() => {
  if (isAdmin.value) {
    return true
  }
  return currentUser.value?.permissions?.some((p) => p.permission_name === 'can_load_car')
})

const handleStatisticsClick = async () => {
  if (isProcessing.value.statistics) return
  isProcessing.value.statistics = true
  try {
    activeView.value = 'statistics'
  } finally {
    isProcessing.value.statistics = false
  }
}

const fetchFinishedOrders = async () => {
  finishedOrdersLoading.value = true
  try {
    const result = await callApi({
      query: `SELECT * FROM cars_stock WHERE date_send_documents IS NOT NULL ORDER BY date_send_documents DESC`,
      params: [],
    })
    if (result.success) {
      finishedOrders.value = result.data
    }
  } finally {
    finishedOrdersLoading.value = false
  }
}

const handleFinishedOrdersClick = async () => {
  showFinishedOrders.value = true
  await fetchFinishedOrders()
}
const handleReturnToMain = () => {
  showFinishedOrders.value = false
}

const openCarsStockInNewTab = () => {
  const route = router.resolve({ name: 'cars-stock' })
  window.open(route.href, '_blank')
}

const handleLoadClick = async () => {
  if (isProcessing.value.load) return
  isProcessing.value.load = true
  try {
    activeView.value = 'load'
  } finally {
    isProcessing.value.load = false
  }
}

const handleLoadingInquiryClick = async () => {
  if (isProcessing.value.loadingInquiry) return
  isProcessing.value.loadingInquiry = true
  try {
    activeView.value = 'loading-inquiry'
  } finally {
    isProcessing.value.loadingInquiry = false
  }
}
</script>

<template>
  <div class="cars-view" :class="{ 'is-loading': isLoading }">
    <!-- Mobile Navigation Toggle -->
    <div class="mobile-nav-toggle">
      <button @click="toggleMobileNav" class="mobile-nav-btn">
        <i class="fas fa-bars"></i>
      </button>
    </div>

    <!-- Mobile Navigation Overlay -->
    <div v-if="showMobileNav" class="mobile-nav-overlay" @click="closeMobileNav"></div>

    <!-- Sidebar -->
    <div class="sidebar" :class="{ 'mobile-open': showMobileNav, collapsed: sidebarCollapsed }">
      <div class="sidebar-top">
        <button
          @click="toggleSidebar"
          class="sidebar-toggle-btn"
          :title="sidebarCollapsed ? 'Expand Sidebar' : 'Collapse Sidebar'"
        >
          <i :class="sidebarCollapsed ? 'fas fa-chevron-right' : 'fas fa-chevron-left'"></i>
        </button>
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
          :disabled="!canCCarStock"
          @click="openCarsStockInNewTab"
          class="sidebar-btn stock-btn"
        >
          <i class="fas fa-warehouse"></i>
          <span>Cars Stock</span>
          <i class="fas fa-external-link-alt" style="margin-left: auto; font-size: 0.8em"></i>
        </button>

        <button
          v-if="canLoadCar"
          @click="handleLoadClick"
          :class="{ active: activeView === 'load', processing: isProcessing.load }"
          class="sidebar-btn load-btn"
          :disabled="isProcessing.load"
        >
          <i class="fas fa-truck-loading"></i>
          <span>Load</span>
          <i v-if="isProcessing.load" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>

        <button
          v-if="canLoadCar"
          @click="handleLoadingInquiryClick"
          :class="{
            active: activeView === 'loading-inquiry',
            processing: isProcessing.loadingInquiry,
          }"
          class="sidebar-btn loading-inquiry-btn"
          :disabled="isProcessing.loadingInquiry"
        >
          <i class="fas fa-list-alt" title="Loading Order"></i>
          <span>Loading Order</span>
          <i
            v-if="isProcessing.loadingInquiry"
            class="fas fa-spinner fa-spin loading-indicator"
          ></i>
        </button>

        <button
          v-if="isAdmin"
          @click="handleStatisticsClick"
          :class="{ processing: isProcessing.statistics }"
          class="sidebar-btn statistics-btn"
          :disabled="isProcessing.statistics"
        >
          <i class="fas fa-chart-bar"></i>
          <span>Statistics</span>
          <i v-if="isProcessing.statistics" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <button
          @click="showFinishedOrders = true"
          :class="{ active: showFinishedOrders }"
          class="sidebar-btn finished-orders-btn"
        >
          <i class="fas fa-check-circle"></i>
          <span>Finished Orders</span>
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
        <button
          v-if="canPurchaseCars"
          @click="navigateTo('containers')"
          :class="{ active: activeView === 'containers', processing: isProcessing.containers }"
          class="sidebar-btn"
          :disabled="isProcessing.containers"
        >
          <i class="fas fa-box"></i>
          <span>Containers</span>
          <i v-if="isProcessing.containers" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>
    <div class="main-content">
      <h1>
        <i class="fas fa-car-side"></i>
        Cars Management
      </h1>
      <div class="content">
        <FinishedOrdersTable v-if="showFinishedOrders" @close="showFinishedOrders = false" />
        <div v-else>
          <div v-if="!activeView" class="empty-state">
            <i class="fas fa-hand-point-left fa-2x"></i>
            <p>Please select an option from the sidebar</p>
          </div>
          <BuyView v-if="activeView === 'buy'" />
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
          <ContainersView v-if="activeView === 'containers'" />
          <StatisticsView v-if="activeView === 'statistics'" />
          <LoadingView v-if="activeView === 'load'" />
        </div>
      </div>
      <div class="copyright">© Merhab Noureddine 2025</div>
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
  transition: width 0.3s ease;
  position: relative;
}

.sidebar.collapsed {
  width: 60px;
  padding: 20px 10px;
}

.sidebar-toggle-btn {
  position: absolute;
  top: 20px;
  right: -15px;
  width: 30px;
  height: 30px;
  background-color: #2c3e50;
  border: 2px solid #34495e;
  border-radius: 50%;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  transition: all 0.2s ease;
}

.sidebar-toggle-btn:hover {
  background-color: #34495e;
  transform: scale(1.1);
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
  transition: opacity 0.3s ease;
}

.sidebar.collapsed .sidebar-btn span {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.sidebar.collapsed .sidebar-section-title {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.sidebar.collapsed .sidebar-btn {
  justify-content: center;
  padding: 12px 8px;
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

.copyright {
  text-align: center;
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 2rem;
  padding-bottom: 1rem;
}

.sidebar-btn.statistics-btn {
  background-color: #8b5cf6;
  color: white;
}

.sidebar-btn.statistics-btn:hover:not(:disabled) {
  background-color: #7c3aed;
}

.sidebar-btn.statistics-btn i:not(.loading-indicator) {
  color: #ddd6fe;
}

.finished-orders-btn {
  background: #e8f5e9;
  color: #388e3c;
}
.finished-orders-btn.active {
  background: #388e3c;
  color: #fff;
}
.finished-orders-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1.5rem;
}
.finished-orders-table th,
.finished-orders-table td {
  border: 1px solid #e0e0e0;
  padding: 0.75rem 1rem;
  text-align: left;
}
.finished-orders-table th {
  background: #f8f9fa;
}
.return-btn {
  margin-bottom: 1rem;
  background: #eee;
  color: #333;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
}
.return-btn:hover {
  background: #ddd;
}

/* Mobile Navigation Styles */
.mobile-nav-toggle {
  display: none;
  position: fixed;
  top: 20px;
  left: 20px;
  z-index: 1000;
}

.mobile-nav-btn {
  background: #2c3e50;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  font-size: 1.2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease;
}

.mobile-nav-btn:hover {
  background: #34495e;
  transform: translateY(-2px);
}

.mobile-nav-overlay {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
}

/* Responsive Design */
@media (max-width: 768px) {
  .cars-view {
    flex-direction: column;
  }

  .mobile-nav-toggle {
    display: block;
  }

  .mobile-nav-overlay {
    display: block;
  }

  .sidebar {
    position: fixed;
    top: 0;
    left: -280px;
    height: 100vh;
    width: 280px;
    z-index: 1000;
    transition: left 0.3s ease;
    overflow-y: auto;
  }

  .sidebar.mobile-open {
    left: 0;
  }

  .main-content {
    margin-left: 0;
    padding: 16px;
    min-height: 100vh;
  }

  .main-content h1 {
    font-size: 1.5rem;
    margin-top: 60px;
  }

  .content {
    margin-top: 16px;
    padding: 16px;
  }

  .sidebar-btn {
    padding: 16px 20px;
    font-size: 1rem;
  }

  .sidebar-btn span {
    font-size: 1rem;
  }

  .sidebar-section-title {
    font-size: 1rem;
    padding: 12px 16px;
    margin: 16px 0 8px 0;
  }

  .empty-state {
    padding: 32px 16px;
    font-size: 1rem;
  }

  .empty-state i {
    font-size: 3rem;
  }

  .copyright {
    margin-top: 1rem;
    padding: 1rem;
    font-size: 0.8rem;
  }
}

@media (max-width: 480px) {
  .mobile-nav-toggle {
    top: 16px;
    left: 16px;
  }

  .mobile-nav-btn {
    padding: 10px;
    font-size: 1rem;
  }

  .sidebar {
    width: 100%;
    left: -100%;
  }

  .main-content {
    padding: 12px;
  }

  .main-content h1 {
    font-size: 1.3rem;
    margin-top: 50px;
  }

  .content {
    padding: 12px;
  }

  .sidebar-btn {
    padding: 14px 16px;
    font-size: 0.95rem;
  }

  .sidebar-btn span {
    font-size: 0.95rem;
  }

  .empty-state {
    padding: 24px 12px;
  }

  .empty-state i {
    font-size: 2.5rem;
  }

  .empty-state p {
    font-size: 0.9rem;
  }
}

/* Landscape orientation adjustments */
@media (max-width: 768px) and (orientation: landscape) {
  .sidebar {
    height: 100vh;
    overflow-y: auto;
  }

  .sidebar-btn {
    padding: 12px 16px;
  }

  .main-content h1 {
    margin-top: 50px;
  }
}

/* High DPI displays */
@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
  .mobile-nav-btn {
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .sidebar-btn {
    min-height: 44px;
  }

  .mobile-nav-btn {
    min-height: 44px;
    min-width: 44px;
  }
}

/* Mobile Navigation Styles */
.mobile-nav-toggle {
  display: none;
  position: fixed;
  top: 20px;
  left: 20px;
  z-index: 1000;
}

.mobile-nav-btn {
  background: #2c3e50;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  font-size: 1.2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease;
}

.mobile-nav-btn:hover {
  background: #34495e;
  transform: translateY(-2px);
}

.mobile-nav-overlay {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
}

/* Responsive Design */
@media (max-width: 768px) {
  .cars-view {
    flex-direction: column;
  }

  .mobile-nav-toggle {
    display: block;
  }

  .mobile-nav-overlay {
    display: block;
  }

  .sidebar {
    position: fixed;
    top: 0;
    left: -280px;
    height: 100vh;
    width: 280px;
    z-index: 1000;
    transition: left 0.3s ease;
    overflow-y: auto;
  }

  .sidebar.mobile-open {
    left: 0;
  }

  .main-content {
    margin-left: 0;
    padding: 16px;
    min-height: 100vh;
  }

  .main-content h1 {
    font-size: 1.5rem;
    margin-top: 60px;
  }

  .content {
    margin-top: 16px;
    padding: 16px;
  }

  .sidebar-btn {
    padding: 16px 20px;
    font-size: 1rem;
  }

  .sidebar-btn span {
    font-size: 1rem;
  }

  .sidebar-section-title {
    font-size: 1rem;
    padding: 12px 16px;
    margin: 16px 0 8px 0;
  }

  .empty-state {
    padding: 32px 16px;
    font-size: 1rem;
  }

  .empty-state i {
    font-size: 3rem;
  }

  .copyright {
    margin-top: 1rem;
    padding: 1rem;
    font-size: 0.8rem;
  }
}

@media (max-width: 480px) {
  .mobile-nav-toggle {
    top: 16px;
    left: 16px;
  }

  .mobile-nav-btn {
    padding: 10px;
    font-size: 1rem;
  }

  .sidebar {
    width: 100%;
    left: -100%;
  }

  .main-content {
    padding: 12px;
  }

  .main-content h1 {
    font-size: 1.3rem;
    margin-top: 50px;
  }

  .content {
    padding: 12px;
  }

  .sidebar-btn {
    padding: 14px 16px;
    font-size: 0.95rem;
  }

  .sidebar-btn span {
    font-size: 0.95rem;
  }

  .empty-state {
    padding: 24px 12px;
  }

  .empty-state i {
    font-size: 2.5rem;
  }

  .empty-state p {
    font-size: 0.9rem;
  }
}

/* Landscape orientation adjustments */
@media (max-width: 768px) and (orientation: landscape) {
  .sidebar {
    height: 100vh;
    overflow-y: auto;
  }

  .sidebar-btn {
    padding: 12px 16px;
  }

  .main-content h1 {
    margin-top: 50px;
  }
}

/* High DPI displays */
@media (-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi) {
  .mobile-nav-btn {
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .sidebar-btn {
    min-height: 44px;
  }

  .mobile-nav-btn {
    min-height: 44px;
    min-width: 44px;
  }
}
</style>
