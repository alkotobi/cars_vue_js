<script setup>
import { defineProps, defineEmits, ref, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi, error, loading } = useApi()

const props = defineProps({
  buyDetails: {
    type: Array,
    required: true
  },
  isAdmin: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['add-detail', 'delete-detail', 'update-detail'])

const showEditDialog = ref(false)
const editingDetail = ref(null)

// Compute if stock is updated from the first detail
const isStockUpdated = computed(() => {
  return props.buyDetails[0]?.is_stock_updated === 1
})

const confirmDelete = (detailId, isStockUpdated) => {
  if (isStockUpdated) {
    alert('Cannot delete details - Stock has already been updated')
    return
  }
  if (confirm('Are you sure you want to delete this detail?')) {
    emit('delete-detail', detailId)
  }
}

const openEditDialog = (detail) => {
  editingDetail.value = { ...detail }
  showEditDialog.value = true
}

const handleEditSubmit = () => {
  emit('update-detail', editingDetail.value)
  showEditDialog.value = false
  editingDetail.value = null
}

const showStockAlert = async () => {
  if (!props.buyDetails.length) {
    alert('No details to process')
    return
  }

  if (confirm('Are you sure you want to update the stock? This action cannot be undone.')) {
    // Process each detail
    for (const detail of props.buyDetails) {
      // Create stock entries based on quantity
      for (let i = 0; i < detail.QTY; i++) {
        const result = await callApi({
          query: `
            INSERT INTO cars_stock 
            (id_buy_details,price_cell)
            VALUES (?,?)
          `,
          params: [detail.id,detail.price_sell]
        })
        
        if (!result.success) {
          alert('Error creating stock entry')
          console.error('Error creating stock entry:', result.error)
          return
        }
      }
    }

    // Update the is_stock_updated flag in buy_bill
    const result = await callApi({
      query: 'UPDATE buy_bill SET is_stock_updated = 1 WHERE id = ?',
      params: [props.buyDetails[0].id_buy_bill]
    })

    if (result.success) {
      alert('Stock has been successfully updated')
      emit('update-detail', props.buyDetails[0]) // Trigger refresh of parent component
    } else {
      console.error('Error updating is_stock_updated flag:', result.error)
      alert('Error updating stock')
    }
  }
}
</script>

<template>
  <div class="detail-section">
    <div class="detail-header">
      <h3>Purchase Details</h3>
      <div class="button-group">
        <button 
          @click="showStockAlert" 
          class="stock-btn"
          :class="{ 'disabled': isStockUpdated }"
          :disabled="isStockUpdated"
          :title="isStockUpdated ? 'Stock has already been updated' : 'Update stock'"
        >
          Update Stock
        </button>
        <button 
          @click="$emit('add-detail')" 
          class="add-btn"
          :class="{ 'disabled': isStockUpdated }"
          :disabled="isStockUpdated"
          :title="isStockUpdated ? 'Cannot add details - Stock has been updated' : 'Add new detail'"
        >
          Add Detail
        </button>
      </div>
    </div>
    <table class="data-table">
      <thead>
        <tr>
          <th>Car</th>
          <th>Color</th>
          <th>Quantity</th>
          <th>Amount</th>
          <th>Year</th>
          <th>Month</th>
          <th>Price Sell</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="detail in buyDetails" :key="detail.id">
          <td>{{ detail.car_name }}</td>
          <td>{{ detail.color }}</td>
          <td>{{ detail.QTY }}</td>
          <td>{{ detail.amount }}</td>
          <td>{{ detail.year }}</td>
          <td>{{ detail.month }}</td>
          <td>{{ detail.price_sell }}</td>
          <td class="actions">
            <button 
              @click="openEditDialog(detail)"
              class="edit-btn"
              :class="{ 'disabled': detail.is_stock_updated }"
              :disabled="detail.is_stock_updated"
              :title="detail.is_stock_updated ? 'Cannot edit - Stock has been updated' : 'Edit this detail'"
            >
              Edit
            </button>
            <button 
              @click="confirmDelete(detail.id, detail.is_stock_updated)"
              class="delete-btn"
              :class="{ 'disabled': detail.is_stock_updated }"
              :disabled="detail.is_stock_updated"
              :title="detail.is_stock_updated ? 'Cannot delete - Stock has been updated' : 'Delete this detail'"
            >
              Delete
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  <!-- Edit Dialog -->
<div v-if="showEditDialog && editingDetail" class="dialog-overlay">
  <div class="dialog">
    <h3>Edit Purchase Detail</h3>
    <form @submit.prevent="handleEditSubmit">
      <div class="form-group">
        <label for="edit-qty">Quantity</label>
        <input 
          type="number" 
          id="edit-qty"
          v-model="editingDetail.QTY" 
          min="1" 
          required
        >
      </div>
      
      <div class="form-group">
        <label for="edit-amount">Amount</label>
        <input 
          type="number" 
          id="edit-amount"
          v-model="editingDetail.amount" 
          step="0.01" 
          required
        >
      </div>
      
      <div class="form-row">
        <div class="form-group half">
          <label for="edit-year">Year</label>
          <input 
            type="number" 
            id="edit-year"
            v-model="editingDetail.year" 
            required
          >
        </div>
        
        <div class="form-group half">
          <label for="edit-month">Month</label>
          <input 
            type="number" 
            id="edit-month"
            v-model="editingDetail.month" 
            min="1" 
            max="12" 
            required
          >
        </div>
      </div>
      
      <!-- Add price_sell field -->
      <div class="form-group">
        <label for="edit-price-sell">Price Sell</label>
        <input 
          type="number" 
          id="edit-price-sell"
          v-model="editingDetail.price_sell"
          step="0.01" 
          required
        >
      </div>
      
      <div class="form-group">
        <label for="edit-notes">Notes</label>
        <textarea 
          id="edit-notes"
          v-model="editingDetail.notes" 
          rows="3"
        ></textarea>
      </div>
      
      <div class="dialog-buttons">
        <button 
          type="button" 
          @click="showEditDialog = false" 
          class="cancel-btn"
        >
          Cancel
        </button>
        <button type="submit" class="submit-btn">
          Save Changes
        </button>
      </div>
    </form>
  </div>
</div>
</template>

<style scoped>
.detail-section {
  width: 100%;
  border-top: 1px solid #e5e7eb;
  padding-top: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.add-btn:hover {
  background-color: #059669;
}

.delete-btn {
  padding: 4px 8px;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.delete-btn.disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.delete-btn.disabled:hover {
  background-color: #9ca3af;
}

.actions-cell {
  display: flex;
  gap: 8px;
}

.edit-btn {
  padding: 4px 8px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.edit-btn:hover {
  background-color: #2563eb;
}

.edit-btn.disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.dialog {
  background: white;
  padding: 24px;
  border-radius: 8px;
  min-width: 400px;
  max-width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-group.half {
  flex: 1;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  padding: 8px 16px;
  background-color: #9ca3af;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.submit-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.cancel-btn:hover {
  background-color: #6b7280;
}

.submit-btn:hover {
  background-color: #059669;
}

.button-group {
  display: flex;
  gap: 12px;
}

.stock-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.stock-btn:hover {
  background-color: #2563eb;
}

.stock-btn.disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
}

.stock-btn.disabled:hover {
  background-color: #9ca3af;
}

.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

button.disabled {
  background-color: #ccc;
  pointer-events: none;
}

</style>