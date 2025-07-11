<script setup>
import { ref, defineProps, defineEmits } from 'vue'

const props = defineProps({
  selectedCount: {
    type: Number,
    default: 0,
  },
  totalCount: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['print-selected', 'loading-order'])

const isDropdownOpen = ref(false)

const toggleDropdown = () => {
  isDropdownOpen.value = !isDropdownOpen.value
}

const closeDropdown = () => {
  isDropdownOpen.value = false
}

const handlePrintSelected = () => {
  emit('print-selected')
  closeDropdown()
}

const handleLoadingOrder = () => {
  emit('loading-order')
  closeDropdown()
}

// Close dropdown when clicking outside
const handleClickOutside = (event) => {
  if (!event.target.closest('.print-dropdown')) {
    closeDropdown()
  }
}

// Add event listener when component mounts
import { onMounted, onUnmounted } from 'vue'

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="print-dropdown">
    <button
      @click="toggleDropdown"
      class="print-dropdown-toggle"
      :disabled="totalCount === 0"
      title="Print Options"
    >
      <i class="fas fa-print"></i>
      <span>Print</span>
      <i class="fas fa-chevron-down dropdown-arrow" :class="{ rotated: isDropdownOpen }"></i>
    </button>

    <div v-if="isDropdownOpen" class="print-dropdown-menu">
      <button
        @click="handlePrintSelected"
        class="print-dropdown-item"
        :disabled="selectedCount === 0"
        :title="
          selectedCount === 0
            ? 'No cars selected'
            : `Print ${selectedCount} selected car${selectedCount === 1 ? '' : 's'}`
        "
      >
        <i class="fas fa-check-square"></i>
        <span>Print Selected ({{ selectedCount }})</span>
      </button>
      <button
        @click="handleLoadingOrder"
        class="print-dropdown-item"
        :disabled="selectedCount === 0"
        :title="
          selectedCount === 0
            ? 'No cars selected'
            : `Loading order for ${selectedCount} selected car${selectedCount === 1 ? '' : 's'}`
        "
      >
        <i class="fas fa-spinner"></i>
        <span>Loading Order ({{ selectedCount }})</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.print-dropdown {
  position: relative;
  display: inline-block;
}

.print-dropdown-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #3b82f6;
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

.print-dropdown-toggle:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.print-dropdown-toggle:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.dropdown-arrow {
  font-size: 12px;
  transition: transform 0.2s ease;
}

.dropdown-arrow.rotated {
  transform: rotate(180deg);
}

.print-dropdown-menu {
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

.print-dropdown-item {
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

.print-dropdown-item:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.print-dropdown-item:disabled {
  color: #9ca3af;
  cursor: not-allowed;
  background-color: #f9fafb;
}

.print-dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.print-dropdown-item:hover:not(:disabled) i {
  color: #374151;
}

.print-dropdown-item:disabled i {
  color: #d1d5db;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .print-dropdown-menu {
    right: auto;
    left: 0;
    min-width: 180px;
  }

  .print-dropdown-toggle {
    min-width: 80px;
    padding: 10px 12px;
  }

  .print-dropdown-toggle span {
    display: none;
  }

  .print-dropdown-toggle i:first-child {
    margin-right: 0;
  }
}
</style>
