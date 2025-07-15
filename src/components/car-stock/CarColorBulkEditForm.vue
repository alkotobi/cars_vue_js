<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  selectedCars: {
    type: Array,
    default: () => [],
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])

const { callApi } = useApi()

const selectedColor = ref(null)
const colors = ref([])
const isProcessing = ref(false)

// Watch for modal visibility to fetch colors
watch(
  () => props.show,
  (newValue) => {
    if (newValue) {
      fetchColors()
    }
  },
)

onMounted(async () => {
  await fetchColors()
})

const fetchColors = async () => {
  try {
    console.log('Fetching colors...')
    const result = await callApi({
      query: `
        SELECT id, color as name, hexa 
        FROM colors 
        ORDER BY color ASC
      `,
      requiresAuth: true,
    })

    console.log('Colors fetch result:', result)
    if (result.success && result.data && result.data.length > 0) {
      colors.value = result.data
      console.log('Colors loaded:', colors.value)
    } else {
      console.log('No colors found or error:', result)
      // Fallback to basic colors if database query fails
      colors.value = [
        { id: 1, name: 'Red', hexa: '#ff0000' },
        { id: 2, name: 'Blue', hexa: '#0000ff' },
        { id: 3, name: 'Green', hexa: '#00ff00' },
        { id: 4, name: 'Yellow', hexa: '#ffff00' },
        { id: 5, name: 'Black', hexa: '#000000' },
        { id: 6, name: 'White', hexa: '#ffffff' },
        { id: 7, name: 'Silver', hexa: '#c0c0c0' },
        { id: 8, name: 'Gray', hexa: '#808080' },
      ]
      console.log('Using fallback colors:', colors.value)
    }
  } catch (err) {
    console.error('Error fetching colors:', err)
    // Fallback to basic colors if database query fails
    colors.value = [
      { id: 1, name: 'Red', hexa: '#ff0000' },
      { id: 2, name: 'Blue', hexa: '#0000ff' },
      { id: 3, name: 'Green', hexa: '#00ff00' },
      { id: 4, name: 'Yellow', hexa: '#ffff00' },
      { id: 5, name: 'Black', hexa: '#000000' },
      { id: 6, name: 'White', hexa: '#ffffff' },
      { id: 7, name: 'Silver', hexa: '#c0c0c0' },
      { id: 8, name: 'Gray', hexa: '#808080' },
    ]
    console.log('Using fallback colors due to error:', colors.value)
  }
}

const handleSave = async () => {
  if (!selectedColor.value) {
    alert('Please select a color')
    return
  }

  isProcessing.value = true
  try {
    const carIds = props.selectedCars.map((car) => car.id)
    const colorId = selectedColor.value

    const result = await callApi({
      query: `
        UPDATE cars_stock
        SET id_color = ?
        WHERE id IN (${carIds.map(() => '?').join(',')})
      `,
      params: [colorId, ...carIds],
      requiresAuth: true,
    })

    if (result.success) {
      console.log('Colors updated for cars:', carIds)
      emit(
        'save',
        props.selectedCars.map((car) => ({ ...car, id_color: colorId })),
      )
    } else {
      alert('Failed to update car colors. Please try again.')
    }
  } catch (err) {
    console.error('Error updating car colors:', err)
    alert('Error updating car colors. Please try again.')
  } finally {
    isProcessing.value = false
  }
}

const handleRevert = async () => {
  if (!props.isAdmin) return

  if (!confirm('Are you sure you want to remove color from all selected cars?')) {
    return
  }

  isProcessing.value = true
  try {
    const carIds = props.selectedCars.map((car) => car.id)

    const result = await callApi({
      query: `
        UPDATE cars_stock
        SET id_color = NULL
        WHERE id IN (${carIds.map(() => '?').join(',')})
      `,
      params: carIds,
      requiresAuth: true,
    })

    if (result.success) {
      console.log('Colors removed from cars:', carIds)
      emit(
        'save',
        props.selectedCars.map((car) => ({ ...car, id_color: null })),
      )
    } else {
      alert('Failed to remove car colors. Please try again.')
    }
  } catch (err) {
    console.error('Error removing car colors:', err)
    alert('Error removing car colors. Please try again.')
  } finally {
    isProcessing.value = false
  }
}

const handleClose = () => {
  selectedColor.value = null
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click="handleClose">
      <div class="modal-content" @click.stop>
        <div class="color-form">
          <div class="form-header">
            <h3>
              <i class="fas fa-palette"></i>
              Edit Color for {{ selectedCars.length }} Car{{ selectedCars.length === 1 ? '' : 's' }}
            </h3>
            <button class="close-btn" @click="handleClose" :disabled="isProcessing">
              <i class="fas fa-times"></i>
            </button>
          </div>

          <div class="form-content">
            <div class="selected-cars-info">
              <p>Selected cars:</p>
              <div class="cars-list">
                <div v-for="car in selectedCars" :key="car.id" class="car-item">
                  <span class="car-id">#{{ car.id }}</span>
                  <span class="car-name">{{ car.car_name }}</span>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label for="color">Select Color</label>
              <select id="color" v-model="selectedColor" class="color-select">
                <option value="">Choose a color</option>
                <option v-for="color in colors" :key="color.id" :value="color.id">
                  {{ color.name }}
                </option>
              </select>
              <div v-if="selectedColor" class="selected-color-preview">
                <span
                  class="color-preview"
                  :style="{
                    backgroundColor: colors.find((c) => c.id == selectedColor)?.hexa || '#000000',
                  }"
                ></span>
                <span>{{ colors.find((c) => c.id == selectedColor)?.name }}</span>
              </div>
            </div>

            <div class="form-actions">
              <button
                @click="handleSave"
                class="btn btn-primary"
                :disabled="!selectedColor || isProcessing"
              >
                <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
                <span v-else>Save Color</span>
              </button>
              <button @click="handleClose" class="btn btn-secondary" :disabled="isProcessing">
                Cancel
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1002;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.color-form {
  padding: 0;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
  border-top-left-radius: 12px;
  border-top-right-radius: 12px;
}

.form-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
}

.form-header h3 i {
  color: #4f46e5;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.form-content {
  padding: 24px;
}

.selected-cars-info {
  margin-bottom: 20px;
}

.selected-cars-info p {
  margin: 0 0 12px 0;
  font-weight: 500;
  color: #374151;
}

.cars-list {
  max-height: 120px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 8px;
  background-color: #f9fafb;
}

.car-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  font-size: 14px;
}

.car-id {
  font-weight: 600;
  color: #4f46e5;
  min-width: 40px;
}

.car-name {
  color: #374151;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.color-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  background-color: white;
}

.color-select:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.selected-color-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 8px;
  background-color: #f9fafb;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
}

.color-preview {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 2px solid #d1d5db;
  display: inline-block;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background-color: #4f46e5;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #4338ca;
}

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #4b5563;
}

.btn-danger {
  background-color: #dc2626;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background-color: #b91c1c;
}
</style>
