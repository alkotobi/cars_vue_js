<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import CarStockPrintDropdown from './CarStockPrintDropdown.vue'

const { t } = useEnhancedI18n()

const isCombineDropdownOpen = ref(false)
const isSelectionsDropdownOpen = ref(false)
const isDocumentsDropdownOpen = ref(false)
const isUpgradesDropdownOpen = ref(false)

const toggleCombineDropdown = () => {
  isCombineDropdownOpen.value = !isCombineDropdownOpen.value
  if (isCombineDropdownOpen.value) {
    isSelectionsDropdownOpen.value = false
    isDocumentsDropdownOpen.value = false
  }
}

const closeCombineDropdown = () => {
  isCombineDropdownOpen.value = false
}

const toggleSelectionsDropdown = () => {
  isSelectionsDropdownOpen.value = !isSelectionsDropdownOpen.value
  if (isSelectionsDropdownOpen.value) {
    isCombineDropdownOpen.value = false
    isDocumentsDropdownOpen.value = false
  }
}

const closeSelectionsDropdown = () => {
  isSelectionsDropdownOpen.value = false
}

const toggleDocumentsDropdown = () => {
  isDocumentsDropdownOpen.value = !isDocumentsDropdownOpen.value
  if (isDocumentsDropdownOpen.value) {
    isCombineDropdownOpen.value = false
    isSelectionsDropdownOpen.value = false
    isUpgradesDropdownOpen.value = false
  }
}

const closeDocumentsDropdown = () => {
  isDocumentsDropdownOpen.value = false
}

const toggleUpgradesDropdown = () => {
  isUpgradesDropdownOpen.value = !isUpgradesDropdownOpen.value
  if (isUpgradesDropdownOpen.value) {
    isCombineDropdownOpen.value = false
    isSelectionsDropdownOpen.value = false
    isDocumentsDropdownOpen.value = false
  }
}

const closeUpgradesDropdown = () => {
  isUpgradesDropdownOpen.value = false
}

const handleCombineByBuyRef = () => {
  emit('combine-by-buy-ref')
  closeCombineDropdown()
}

const handleCombineByContainer = () => {
  emit('combine-by-container')
  closeCombineDropdown()
}

const handleUncombine = () => {
  emit('uncombine')
  closeCombineDropdown()
}

// Close dropdown when clicking outside
const handleClickOutside = (event) => {
  if (!event.target.closest('.combine-dropdown')) {
    closeCombineDropdown()
  }
  if (!event.target.closest('.selections-dropdown')) {
    closeSelectionsDropdown()
  }
  if (!event.target.closest('.documents-dropdown')) {
    closeDocumentsDropdown()
  }
  if (!event.target.closest('.upgrades-dropdown')) {
    closeUpgradesDropdown()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

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
  canHideCar: {
    type: Boolean,
    default: false,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
  isCombineMode: {
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
  'cfr-da',
  'refresh',
  'delete-cars',
  'toggle-hidden',
  'mark-delivered',
  'combine-by-buy-ref',
  'combine-by-container',
  'uncombine',
  'save-selection',
  'show-selections',
  'transfer',
  'checkout',
  'add-upgrades',
  'mark-upgrades-done',
])
</script>

<template>
  <div class="car-stock-toolbar">
    <div class="toolbar-left">
      <div class="selection-info" v-if="selectedCars.size > 0">
        <i class="fas fa-check-circle"></i>
        <span
          >{{ selectedCars.size }} {{ t('carStockToolbar.car', { count: selectedCars.size }) }}
          {{ t('carStockToolbar.selected') }}</span
        >
      </div>
      <div class="total-info" v-else>
        <i class="fas fa-car"></i>
        <span
          >{{ totalCars }} {{ t('carStockToolbar.car', { count: totalCars }) }}
          {{ t('carStockToolbar.in_stock') }}</span
        >
      </div>
    </div>

    <div class="toolbar-right">
      <button @click="$emit('refresh')" class="refresh-btn" :title="t('carStockToolbar.refresh')">
        <i class="fas fa-sync-alt"></i>
        <span>{{ t('carStockToolbar.refresh') }}</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('ports')"
        class="ports-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.edit_ports_for_selected_cars')
        "
      >
        <i class="fas fa-anchor"></i>
        <span>{{ t('carStockToolbar.ports') }}</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('warehouse')"
        class="warehouse-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.edit_warehouse_for_selected_cars')
        "
      >
        <i class="fas fa-warehouse"></i>
        <span>{{ t('carStockToolbar.warehouse') }}</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('notes')"
        class="notes-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.edit_notes_for_selected_cars')
        "
      >
        <i class="fas fa-sticky-note"></i>
        <span>{{ t('carStockToolbar.notes') }}</span>
      </button>
      <button
        @click="$emit('task')"
        class="task-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.create_task_for_selected_cars', { count: selectedCars.size })
        "
      >
        <i class="fas fa-tasks"></i>
        <span>{{ t('carStockToolbar.task') }}</span>
      </button>
      <button
        v-if="isAdmin && canChangeColor"
        @click="$emit('color')"
        class="color-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.edit_color_for_selected_cars')
        "
      >
        <i class="fas fa-palette"></i>
        <span>{{ t('carStockToolbar.color') }}</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('vin')"
        class="vin-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.edit_vin_for_selected_cars')
        "
      >
        <i class="fas fa-barcode"></i>
        <span>{{ t('carStockToolbar.vin') }}</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('export-license')"
        class="export-license-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.edit_export_license_for_selected_cars')
        "
      >
        <i class="fas fa-file-signature"></i>
        <span>{{ t('carStockToolbar.export_license') }}</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('cfr-da')"
        class="cfr-da-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.edit_cfr_da_for_selected_cars')
        "
      >
        <i class="fas fa-coins"></i>
        <span>CFR DA</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('mark-delivered')"
        class="delivered-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.mark_delivered_for_selected_cars')
        "
      >
        <i class="fas fa-check-double"></i>
        <span>{{ t('carStockToolbar.delivered') }}</span>
      </button>
      <button
        v-if="isAdmin"
        @click="$emit('delete-cars')"
        class="delete-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.delete_selected_cars')
        "
      >
        <i class="fas fa-trash"></i>
        <span>{{ t('carStockToolbar.delete') }}</span>
      </button>
      <button
        v-if="isAdmin || canHideCar"
        @click="$emit('toggle-hidden')"
        class="toggle-hidden-btn"
        :disabled="selectedCars.size === 0"
        :title="
          selectedCars.size === 0
            ? t('carStockToolbar.no_cars_selected')
            : t('carStockToolbar.toggle_hidden_status')
        "
      >
        <i class="fas fa-eye-slash"></i>
        <span>{{ t('carStockToolbar.toggle_hidden') }}</span>
      </button>
      <CarStockPrintDropdown
        :selected-count="selectedCars.size"
        :total-count="totalCars"
        @print-selected="$emit('print-selected')"
        @loading-order="$emit('loading-order')"
      />
      <div class="combine-dropdown">
        <button
          @click="toggleCombineDropdown"
          class="combine-dropdown-toggle"
          :title="t('carStockToolbar.combine')"
        >
          <i class="fas fa-layer-group"></i>
          <span>{{ t('carStockToolbar.combine') }}</span>
          <i class="fas fa-chevron-down dropdown-arrow" :class="{ rotated: isCombineDropdownOpen }"></i>
        </button>

        <div v-if="isCombineDropdownOpen" class="combine-dropdown-menu">
          <button
            v-if="!isCombineMode"
            @click="handleCombineByBuyRef"
            class="combine-dropdown-item"
            :title="t('carStockToolbar.combine_by_buy_ref')"
          >
            <i class="fas fa-shopping-cart"></i>
            <span>{{ t('carStockToolbar.by_buy_ref') }}</span>
          </button>
          <button
            v-if="!isCombineMode"
            @click="handleCombineByContainer"
            class="combine-dropdown-item"
            :title="t('carStockToolbar.combine_by_container')"
          >
            <i class="fas fa-box"></i>
            <span>{{ t('carStockToolbar.by_container') }}</span>
          </button>
          <div v-if="isCombineMode" class="combine-dropdown-divider"></div>
          <button
            v-if="isCombineMode"
            @click="handleUncombine"
            class="combine-dropdown-item uncombine-item"
            :title="t('carStockToolbar.uncombine')"
          >
            <i class="fas fa-times-circle"></i>
            <span>{{ t('carStockToolbar.uncombine') }}</span>
          </button>
        </div>
      </div>

      <div class="selections-dropdown">
        <button
          @click="toggleSelectionsDropdown"
          class="selections-dropdown-toggle"
          :title="t('carStockToolbar.selections')"
        >
          <i class="fas fa-bookmark"></i>
          <span>{{ t('carStockToolbar.selections') || 'Selections' }}</span>
          <i class="fas fa-chevron-down dropdown-arrow" :class="{ rotated: isSelectionsDropdownOpen }"></i>
        </button>

        <div v-if="isSelectionsDropdownOpen" class="selections-dropdown-menu">
          <button
            @click="$emit('save-selection'); closeSelectionsDropdown()"
            class="selections-dropdown-item"
            :disabled="selectedCars.size === 0"
            :title="t('carStockToolbar.save_selection')"
          >
            <i class="fas fa-save"></i>
            <span>{{ t('carStockToolbar.save_selection') || 'Save Selection' }}</span>
          </button>
          <button
            @click="$emit('show-selections'); closeSelectionsDropdown()"
            class="selections-dropdown-item"
            :title="t('carStockToolbar.show_selections')"
          >
            <i class="fas fa-list"></i>
            <span>{{ t('carStockToolbar.show_selections') || 'Show Selections' }}</span>
          </button>
        </div>
      </div>

      <div class="documents-dropdown">
        <button
          @click="toggleDocumentsDropdown"
          class="documents-dropdown-toggle"
          :title="t('carStockToolbar.documents')"
        >
          <i class="fas fa-file-alt"></i>
          <span>{{ t('carStockToolbar.documents') || 'Documents' }}</span>
          <i class="fas fa-chevron-down dropdown-arrow" :class="{ rotated: isDocumentsDropdownOpen }"></i>
        </button>

        <div v-if="isDocumentsDropdownOpen" class="documents-dropdown-menu">
          <button
            @click="$emit('transfer'); closeDocumentsDropdown()"
            class="documents-dropdown-item"
            :disabled="selectedCars.size === 0"
            :title="t('carStockToolbar.transfer')"
          >
            <i class="fas fa-exchange-alt"></i>
            <span>{{ t('carStockToolbar.transfer') || 'Transfer' }}</span>
          </button>
          <button
            @click="$emit('checkout'); closeDocumentsDropdown()"
            class="documents-dropdown-item"
            :disabled="selectedCars.size === 0"
            :title="t('carStockToolbar.checkout')"
          >
            <i class="fas fa-shopping-cart"></i>
            <span>{{ t('carStockToolbar.checkout') || 'Checkout' }}</span>
          </button>
        </div>
      </div>

      <div class="upgrades-dropdown">
        <button
          @click="toggleUpgradesDropdown"
          class="upgrades-dropdown-toggle"
          :title="t('carStockToolbar.upgrades')"
        >
          <i class="fas fa-wrench"></i>
          <span>{{ t('carStockToolbar.upgrades') || 'Upgrades' }}</span>
          <i class="fas fa-chevron-down dropdown-arrow" :class="{ rotated: isUpgradesDropdownOpen }"></i>
        </button>

        <div v-if="isUpgradesDropdownOpen" class="upgrades-dropdown-menu">
          <button
            @click="$emit('add-upgrades'); closeUpgradesDropdown()"
            class="upgrades-dropdown-item"
            :disabled="selectedCars.size === 0"
            :title="t('carStockToolbar.add_upgrades_to_selected_cars')"
          >
            <i class="fas fa-plus"></i>
            <span>{{ t('carStockToolbar.add_upgrade') || 'Add Upgrade' }}</span>
          </button>
          <button
            @click="$emit('mark-upgrades-done'); closeUpgradesDropdown()"
            class="upgrades-dropdown-item"
            :disabled="selectedCars.size === 0"
            :title="t('carStockToolbar.mark_upgrades_done')"
          >
            <i class="fas fa-check"></i>
            <span>{{ t('carStockToolbar.make_upgrade_done') || 'Make Upgrade Done' }}</span>
          </button>
        </div>
      </div>
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
  flex-wrap: wrap;
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

.cfr-da-btn {
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

.cfr-da-btn:hover:not(:disabled) {
  background-color: #d97706;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.cfr-da-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.delivered-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #0ea5e9;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.delivered-btn:hover:not(:disabled) {
  background-color: #0284c7;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.delivered-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.delete-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}
.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.delete-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.toggle-hidden-btn {
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

.toggle-hidden-btn:hover:not(:disabled) {
  background-color: #7c3aed;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.toggle-hidden-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #2563eb;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}
.refresh-btn:hover:not(:disabled) {
  background-color: #1d4ed8;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.refresh-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* Combine Dropdown Styles */
.combine-dropdown {
  position: relative;
  display: inline-block;
}

.combine-dropdown-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #6366f1;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  min-width: 100px;
  justify-content: center;
}

.combine-dropdown-toggle:hover:not(:disabled) {
  background-color: #4f46e5;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.combine-dropdown-toggle:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.combine-dropdown .dropdown-arrow {
  font-size: 12px;
  transition: transform 0.2s ease;
}

.combine-dropdown .dropdown-arrow.rotated {
  transform: rotate(180deg);
}

.combine-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 1000;
  min-width: 200px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  margin-top: 4px;
  overflow: hidden;
}

.combine-dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
  transition: background-color 0.2s ease;
  text-align: left;
}

.combine-dropdown-item:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.combine-dropdown-item:disabled {
  color: #9ca3af;
  cursor: not-allowed;
  background-color: #f9fafb;
}

.combine-dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.combine-dropdown-item:hover:not(:disabled) i {
  color: #374151;
}

.combine-dropdown-item:disabled i {
  color: #d1d5db;
}

.combine-dropdown-divider {
  height: 1px;
  background-color: #e5e7eb;
  margin: 4px 0;
}

.combine-dropdown-item.uncombine-item {
  color: #ef4444;
}

.combine-dropdown-item.uncombine-item:hover:not(:disabled) {
  background-color: #fef2f2;
  color: #dc2626;
}

.combine-dropdown-item.uncombine-item i {
  color: #ef4444;
}

.combine-dropdown-item.uncombine-item:hover:not(:disabled) i {
  color: #dc2626;
}

/* Selections Dropdown Styles */
.selections-dropdown {
  position: relative;
  display: inline-block;
}

.selections-dropdown-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #8b5cf6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  min-width: 100px;
  justify-content: center;
}

.selections-dropdown-toggle:hover:not(:disabled) {
  background-color: #7c3aed;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.selections-dropdown-toggle:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.selections-dropdown .dropdown-arrow {
  font-size: 12px;
  transition: transform 0.2s ease;
}

.selections-dropdown .dropdown-arrow.rotated {
  transform: rotate(180deg);
}

.selections-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 1000;
  min-width: 200px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  margin-top: 4px;
  overflow: hidden;
}

.selections-dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
  transition: background-color 0.2s ease;
  text-align: left;
}

.selections-dropdown-item:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.selections-dropdown-item:disabled {
  color: #9ca3af;
  cursor: not-allowed;
  background-color: #f9fafb;
}

.selections-dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.selections-dropdown-item:hover:not(:disabled) i {
  color: #374151;
}

.selections-dropdown-item:disabled i {
  color: #d1d5db;
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

  .combine-dropdown-menu {
    right: auto;
    left: 0;
    min-width: 180px;
  }

  .combine-dropdown-toggle {
    min-width: 80px;
    padding: 10px 12px;
  }

  .combine-dropdown-toggle span {
    display: none;
  }

  .combine-dropdown-toggle i:first-child {
    margin-right: 0;
  }

  .selections-dropdown-menu {
    right: auto;
    left: 0;
    min-width: 180px;
  }

  .selections-dropdown-toggle {
    min-width: 80px;
    padding: 10px 12px;
  }

  .selections-dropdown-toggle span {
    display: none;
  }

  .selections-dropdown-toggle i:first-child {
    margin-right: 0;
  }

  .documents-dropdown-menu {
    right: auto;
    left: 0;
    min-width: 180px;
  }

  .documents-dropdown-toggle {
    min-width: 80px;
    padding: 10px 12px;
  }

  .documents-dropdown-toggle span {
    display: none;
  }

  .documents-dropdown-toggle i:first-child {
    margin-right: 0;
  }
}

/* Documents Dropdown Styles */
.documents-dropdown {
  position: relative;
  display: inline-block;
}

.documents-dropdown-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  min-width: 100px;
  justify-content: center;
}

.documents-dropdown-toggle:hover:not(:disabled) {
  background-color: #0891b2;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.documents-dropdown-toggle:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.documents-dropdown .dropdown-arrow {
  font-size: 12px;
  transition: transform 0.2s ease;
}

.documents-dropdown .dropdown-arrow.rotated {
  transform: rotate(180deg);
}

.documents-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 1000;
  min-width: 200px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  margin-top: 4px;
  overflow: hidden;
}

.documents-dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
  transition: background-color 0.2s ease;
  text-align: left;
}

.documents-dropdown-item:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.documents-dropdown-item:disabled {
  color: #9ca3af;
  cursor: not-allowed;
  background-color: #f9fafb;
}

.documents-dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.documents-dropdown-item:hover:not(:disabled) i {
  color: #374151;
}

.documents-dropdown-item:disabled i {
  color: #d1d5db;
}

/* Upgrades Dropdown Styles */
.upgrades-dropdown {
  position: relative;
  display: inline-block;
}

.upgrades-dropdown-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #f59e0b;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  min-width: 100px;
  justify-content: center;
}

.upgrades-dropdown-toggle:hover:not(:disabled) {
  background-color: #d97706;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.upgrades-dropdown-toggle:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.upgrades-dropdown .dropdown-arrow {
  font-size: 12px;
  transition: transform 0.2s ease;
}

.upgrades-dropdown .dropdown-arrow.rotated {
  transform: rotate(180deg);
}

.upgrades-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 1000;
  min-width: 200px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  margin-top: 4px;
  overflow: hidden;
}

.upgrades-dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
  transition: background-color 0.2s ease;
  text-align: left;
}

.upgrades-dropdown-item:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.upgrades-dropdown-item:disabled {
  color: #9ca3af;
  cursor: not-allowed;
  background-color: #f9fafb;
}

.upgrades-dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.upgrades-dropdown-item:hover:not(:disabled) i {
  color: #374151;
}

.upgrades-dropdown-item:disabled i {
  color: #d1d5db;
}

@media (max-width: 768px) {
  .upgrades-dropdown-menu {
    right: auto;
    left: 0;
    min-width: 180px;
  }

  .upgrades-dropdown-toggle {
    min-width: 80px;
    padding: 10px 12px;
  }

  .upgrades-dropdown-toggle span {
    display: none;
  }

  .upgrades-dropdown-toggle i:first-child {
    margin-right: 0;
  }
}
</style>
