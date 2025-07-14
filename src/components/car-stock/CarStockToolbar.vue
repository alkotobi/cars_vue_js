<script setup>
import { defineProps, defineEmits } from 'vue'
import CarStockPrintDropdown from './CarStockPrintDropdown.vue'

const props = defineProps({
  selectedCars: {
    type: Set,
    default: () => new Set(),
  },
  totalCars: {
    type: Number,
    default: 0,
  },
  canChangeColor: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits([
  'print-selected',
  'loading-order',
  'vin',
  'warehouse',
  'notes',
  'task',
  'color',
  'export-license',
])
</script>

<template>
  <div class="car-stock-toolbar">
    <div class="toolbar-left">
      <div class="selection-info" v-if="selectedCars.size > 0">
        <i class="fas fa-check-circle"></i>
        <span>{{ selectedCars.size }} car{{ selectedCars.size === 1 ? '' : 's' }} selected</span>
      </div>
      <div class="total-info" v-else>
        <i class="fas fa-car"></i>
        <span>{{ totalCars }} car{{ totalCars === 1 ? '' : 's' }} in stock</span>
      </div>
    </div>

    <div class="toolbar-right">
      <button
        @click="$emit('ports')"
        class="ports-btn"
        :disabled="selectedCars.size === 0"
        :title="selectedCars.size === 0 ? 'No cars selected' : 'Edit ports for selected cars'"
      >
        <i class="fas fa-anchor"></i>
        <span>Ports</span>
      </button>
      <button
        @click="$emit('warehouse')"
        class="warehouse-btn"
        :disabled="selectedCars.size === 0"
        :title="selectedCars.size === 0 ? 'No cars selected' : 'Edit warehouse for selected cars'"
      >
        <i class="fas fa-warehouse"></i>
        <span>Warehouse</span>
      </button>
      <button
        @click="$emit('notes')"
        class="notes-btn"
        :disabled="selectedCars.size === 0"
        :title="selectedCars.size === 0 ? 'No cars selected' : 'Edit notes for selected cars'"
      >
        <i class="fas fa-sticky-note"></i>
        <span>Notes</span>
      </button>
      <button
        @click="$emit('task')"
        class="task-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? 'No cars selected'
            : `Create task for ${selectedCars.size} selected car${selectedCars.size === 1 ? '' : 's'}`
        "
      >
        <i class="fas fa-tasks"></i>
        <span>Task</span>
      </button>
      <button
        v-if="canChangeColor"
        @click="$emit('color')"
        class="color-btn"
        :disabled="selectedCars.size === 0"
        :title="selectedCars.size === 0 ? 'No cars selected' : 'Edit color for selected cars'"
      >
        <i class="fas fa-palette"></i>
        <span>Color</span>
      </button>
      <button
        @click="$emit('vin')"
        class="vin-btn"
        :disabled="selectedCars.size === 0"
        :title="selectedCars.size === 0 ? 'No cars selected' : 'Edit VIN for selected cars'"
      >
        <i class="fas fa-barcode"></i>
        <span>VIN</span>
      </button>
      <button
        @click="$emit('export-license')"
        class="export-license-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0 ? 'No cars selected' : 'Edit export license for selected cars'
        "
      >
        <i class="fas fa-file-signature"></i>
        <span>Export License</span>
      </button>
      <CarStockPrintDropdown
        :selected-count="selectedCars.size"
        :total-count="totalCars"
        @print-selected="$emit('print-selected')"
        @loading-order="$emit('loading-order')"
      />
    </div>
  </div>
</template>

<style scoped>
.car-stock-toolbar {
  position: sticky;
  top: 0;
  z-index: 1000;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #f8f9fa;
  border-bottom: 1px solid #e5e7eb;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  background-color: rgba(248, 249, 250, 0.95);
  border-radius: 8px 8px 0 0;
  margin-bottom: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selection-info,
.total-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.selection-info {
  color: #059669;
}

.selection-info i {
  color: #10b981;
}

.total-info i {
  color: #6b7280;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.warehouse-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #8b5cf6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.warehouse-btn:hover:not(:disabled) {
  background-color: #7c3aed;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.warehouse-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.notes-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #f59e0b;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.notes-btn:hover:not(:disabled) {
  background-color: #d97706;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.notes-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.vin-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.vin-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.vin-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.ports-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.ports-btn:hover:not(:disabled) {
  background-color: #059669;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.ports-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.task-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.task-btn:hover:not(:disabled) {
  background-color: #45a049;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.task-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.color-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #4f46e5;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.color-btn:hover:not(:disabled) {
  background-color: #4338ca;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.color-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.export-license-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}
.export-license-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.export-license-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

@media (max-width: 768px) {
  .car-stock-toolbar {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
    padding: 8px 12px;
  }

  .toolbar-left,
  .toolbar-right {
    justify-content: center;
  }
}
</style>
