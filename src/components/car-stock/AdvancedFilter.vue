<script setup>
import { ref, defineProps, defineEmits, watch } from 'vue'

const props = defineProps({
  show: Boolean,
  initialFilters: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:filters', 'close'])

const filters = ref({
  id: '',
  car_name: '',
  color: '',
  vin: '',
  loading_port: '',
  discharge_port: '',
  freight_min: '',
  freight_max: '',
  price_min: '',
  price_max: '',
  loading_date_from: '',
  loading_date_to: '',
  status: '',
  client: '',
  warehouse: '',
  // New checkbox filters
  has_bl: false,
  freight_paid: false,
  has_supplier_docs: false,
  in_warehouse: false,
  has_export_license: false,
  documents_sent: false,
  is_loaded: false,
  has_vin: false,
  ...props.initialFilters
})

const applyFilters = () => {
  emit('update:filters', { ...filters.value })
}

const resetFilters = () => {
  filters.value = {
    id: '',
    car_name: '',
    color: '',
    vin: '',
    loading_port: '',
    discharge_port: '',
    freight_min: '',
    freight_max: '',
    price_min: '',
    price_max: '',
    loading_date_from: '',
    loading_date_to: '',
    status: '',
    client: '',
    warehouse: '',
    has_bl: false,
    freight_paid: false,
    has_supplier_docs: false,
    in_warehouse: false,
    has_export_license: false,
    documents_sent: false,
    is_loaded: false,
    has_vin: false
  }
  applyFilters()
}

watch(() => props.initialFilters, (newFilters) => {
  filters.value = { ...filters.value, ...newFilters }
}, { deep: true })
</script>

<template>
  <div v-if="show" class="advanced-filter-overlay">
    <div class="advanced-filter-modal">
      <div class="modal-header">
        <h3>Advanced Filters</h3>
        <button class="close-btn" @click="$emit('close')">&times;</button>
      </div>

      <div class="filter-content">
        <div class="filter-grid">
          <!-- Basic Information Section -->
          <div class="filter-section">
            <h4>Basic Information</h4>
            <div class="filter-group">
              <label>ID:</label>
              <input type="text" v-model="filters.id" placeholder="Car ID">
            </div>
            <div class="filter-group">
              <label>Car Name:</label>
              <input type="text" v-model="filters.car_name" placeholder="Car name">
            </div>
            <div class="filter-group">
              <label>Color:</label>
              <input type="text" v-model="filters.color" placeholder="Color">
            </div>
            <div class="filter-group">
              <label>VIN:</label>
              <input type="text" v-model="filters.vin" placeholder="VIN number">
            </div>
          </div>

          <!-- Ports Section -->
          <div class="filter-section">
            <h4>Ports</h4>
            <div class="filter-group">
              <label>Loading Port:</label>
              <input type="text" v-model="filters.loading_port" placeholder="Loading port">
            </div>
            <div class="filter-group">
              <label>Discharge Port:</label>
              <input type="text" v-model="filters.discharge_port" placeholder="Discharge port">
            </div>
          </div>

          <!-- Financial Section -->
          <div class="filter-section">
            <h4>Financial</h4>
            <div class="filter-group">
              <label>Freight Range:</label>
              <div class="range-inputs">
                <input type="number" v-model="filters.freight_min" placeholder="Min">
                <span>to</span>
                <input type="number" v-model="filters.freight_max" placeholder="Max">
              </div>
            </div>
            <div class="filter-group">
              <label>Price Range:</label>
              <div class="range-inputs">
                <input type="number" v-model="filters.price_min" placeholder="Min">
                <span>to</span>
                <input type="number" v-model="filters.price_max" placeholder="Max">
              </div>
            </div>
          </div>

          <!-- Status Section -->
          <div class="filter-section">
            <h4>Status</h4>
            <div class="filter-group">
              <label>Status:</label>
              <select v-model="filters.status">
                <option value="">All</option>
                <option value="available">Available</option>
                <option value="sold">Sold</option>
              </select>
            </div>
            <div class="filter-group">
              <label>Client:</label>
              <input type="text" v-model="filters.client" placeholder="Client name">
            </div>
            <div class="filter-group">
              <label>Warehouse:</label>
              <input type="text" v-model="filters.warehouse" placeholder="Warehouse">
            </div>
          </div>

          <!-- Document Status Section -->
          <div class="filter-section document-status">
            <h4>Document Status</h4>
            <div class="checkbox-grid">
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.has_bl">
                <span>We Got BL</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.freight_paid">
                <span>Freight Paid</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.has_supplier_docs">
                <span>Got Supplier Docs</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.in_warehouse">
                <span>In Warehouse</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.has_export_license">
                <span>Has Export License</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.documents_sent">
                <span>Documents Sent</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.is_loaded">
                <span>Car Loaded</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="filters.has_vin">
                <span>Has VIN</span>
              </label>
            </div>
          </div>

          <!-- Date Range Section -->
          <div class="filter-section">
            <h4>Date Range</h4>
            <div class="filter-group">
              <label>Loading Date:</label>
              <div class="date-range">
                <input type="date" v-model="filters.loading_date_from" placeholder="From">
                <span>to</span>
                <input type="date" v-model="filters.loading_date_to" placeholder="To">
              </div>
            </div>
          </div>
        </div>

        <div class="filter-actions">
          <button class="reset-btn" @click="resetFilters">Reset</button>
          <button class="apply-btn" @click="applyFilters">Apply Filters</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.advanced-filter-overlay {
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

.advanced-filter-modal {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 1200px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.filter-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.filter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.filter-section {
  background-color: #f9fafb;
  padding: 16px;
  border-radius: 6px;
}

.filter-section h4 {
  margin: 0 0 12px 0;
  color: #374151;
  font-size: 1.1rem;
}

.filter-group {
  margin-bottom: 12px;
}

.filter-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
  color: #374151;
}

.filter-group input,
.filter-group select {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.range-inputs,
.date-range {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 8px;
  align-items: center;
}

.range-inputs span,
.date-range span {
  color: #6b7280;
  font-size: 14px;
}

.document-status .checkbox-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.checkbox-label span {
  color: #374151;
  font-size: 14px;
}

.filter-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.reset-btn,
.apply-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.reset-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.reset-btn:hover {
  background-color: #e5e7eb;
}

.apply-btn {
  background-color: #3b82f6;
  color: white;
}

.apply-btn:hover {
  background-color: #2563eb;
}

@media (max-width: 768px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }
  
  .document-status .checkbox-grid {
    grid-template-columns: 1fr;
  }
}
</style> 