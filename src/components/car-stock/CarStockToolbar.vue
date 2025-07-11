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
})

const emit = defineEmits(['print-selected', 'loading-order'])
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #f8f9fa;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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

@media (max-width: 768px) {
  .car-stock-toolbar {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .toolbar-left,
  .toolbar-right {
    justify-content: center;
  }
}
</style>
