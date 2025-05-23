<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import CarStockTable from '../components/car-stock/CarStockTable.vue'
import CarStockForm from '../components/car-stock/CarStockForm.vue'
import CarStockFilter from '../components/car-stock/CarStockFilter.vue'

const router = useRouter()
const showEditDialog = ref(false)
const editingCar = ref(null)
const carStockTableRef = ref(null)
const filters = ref({
  basic: '',
  advanced: null
})

const handleEditCar = (car) => {
  editingCar.value = { ...car }
  showEditDialog.value = true
}

const handleSave = () => {
  showEditDialog.value = false
  
  // Refresh the table
  if (carStockTableRef.value) {
    carStockTableRef.value.fetchCarsStock()
  }
}

const handleFilter = (filterData) => {
  filters.value = filterData
  
  // Refresh the table with the new filters
  if (carStockTableRef.value) {
    carStockTableRef.value.fetchCarsStock()
  }
}

const navigateToWarehouses = () => {
  router.push('/warehouses')
}
</script>

<template>
  <div class="cars-stock">
    <div class="header">
      <h2>Cars Stock Management</h2>
    </div>
    
    <div class="content">
      <!-- Add the filter component -->
      <CarStockFilter @filter="handleFilter" />
      
      <CarStockTable 
        ref="carStockTableRef"
        :onEdit="handleEditCar"
        :filters="filters"
      />
      
      <!-- Add Warehouses Button at the bottom -->
      <div class="warehouses-button-container">
        <button @click="navigateToWarehouses" class="warehouses-btn">
          Manage Warehouses
        </button>
      </div>
    </div>
    
    <!-- Edit Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <CarStockForm
          :carData="editingCar"
          @save="handleSave"
          @cancel="showEditDialog = false"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.cars-stock {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.dialog {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  width: 90%;
  max-width: 1200px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

/* Add these styles for the warehouses button */
.warehouses-button-container {
  margin-top: 20px;
  text-align: center;
}

.warehouses-btn {
  background-color: #4CAF50;
  color: white;
  padding: 12px 24px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.3s;
}

.warehouses-btn:hover {
  background-color: #45a049;
}
</style>