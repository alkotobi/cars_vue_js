<script setup>
import { defineProps, defineEmits, ref, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi, error, loading } = useApi()

const props = defineProps({
  buyDetails: {
    type: Array,
    required: true,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['add-detail', 'delete-detail', 'update-detail'])

const showEditDialog = ref(false)
const editingDetail = ref(null)

// Compute if stock is updated from the first detail
const isStockUpdated = computed(() => {
  return props.buyDetails[0]?.is_stock_updated === 1
})

// Add loading state
const isProcessing = ref(false)

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

  if (isProcessing.value) return // Prevent double-click

  if (confirm('Are you sure you want to update the stock? This action cannot be undone.')) {
    isProcessing.value = true

    try {
      // Process each detail
      for (const detail of props.buyDetails) {
        // Create stock entries based on quantity
        for (let i = 0; i < detail.QTY; i++) {
          const result = await callApi({
            query: `
              INSERT INTO cars_stock 
              (id_buy_details, price_cell, notes, is_used_car, is_big_car)
              VALUES (?, ?, ?, ?, ?)
            `,
            params: [
              detail.id,
              detail.price_sell,
              detail.notes,
              detail.is_used_car,
              detail.is_big_car,
            ],
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
        params: [props.buyDetails[0].id_buy_bill],
      })

      if (result.success) {
        alert('Stock has been successfully updated')
        emit('update-detail', props.buyDetails[0]) // Trigger refresh of parent component
      } else {
        console.error('Error updating is_stock_updated flag:', result.error)
        alert('Error updating stock')
      }
    } catch (err) {
      console.error('Error updating stock:', err)
      alert('Error updating stock: ' + err.message)
    } finally {
      isProcessing.value = false
    }
  }
}
</script>

<template>
  <div class="detail-section">
    <!-- Loading Overlay -->
    <div v-if="isProcessing" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Processing stock update...</span>
    </div>

    <div class="detail-header">
      <h3>
        <i class="fas fa-list-alt"></i>
        Purchase Details
      </h3>
      <div class="button-group">
        <button
          @click="showStockAlert"
          class="stock-btn"
          :class="{ disabled: isStockUpdated || isProcessing }"
          :disabled="isStockUpdated || isProcessing"
          :title="isStockUpdated ? 'Stock has already been updated' : 'Update stock'"
        >
          <i class="fas" :class="isProcessing ? 'fa-spinner fa-spin' : 'fa-boxes'"></i>
          {{ isProcessing ? 'Processing...' : 'Update Stock' }}
        </button>
        <button
          @click="$emit('add-detail')"
          class="add-btn"
          :class="{ disabled: isStockUpdated }"
          :disabled="isStockUpdated"
          :title="isStockUpdated ? 'Cannot add details - Stock has been updated' : 'Add new detail'"
        >
          <i class="fas fa-plus"></i>
          Add Detail
        </button>
      </div>
    </div>
    <table class="data-table">
      <thead>
        <tr>
          <th><i class="fas fa-car"></i> Car</th>
          <th><i class="fas fa-palette"></i> Color</th>
          <th><i class="fas fa-hashtag"></i> Quantity</th>
          <th><i class="fas fa-money-bill-wave"></i> Amount</th>
          <th><i class="fas fa-calendar-alt"></i> Year</th>
          <th><i class="fas fa-calendar-day"></i> Month</th>
          <th><i class="fas fa-tag"></i> Price Sell</th>
          <th><i class="fas fa-sticky-note"></i> Notes</th>
          <th><i class="fas fa-car-alt"></i> Used Car</th>
          <th><i class="fas fa-truck"></i> Big Car</th>
          <th><i class="fas fa-cog"></i> Actions</th>
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
          <td>{{ detail.notes }}</td>
          <td>
            <i
              :class="detail.is_used_car ? 'fas fa-check text-success' : 'fas fa-times text-danger'"
              :title="detail.is_used_car ? 'Used Car' : 'New Car'"
            ></i>
          </td>
          <td>
            <i
              :class="detail.is_big_car ? 'fas fa-check text-success' : 'fas fa-times text-danger'"
              :title="detail.is_big_car ? 'Big Car' : 'Small Car'"
            ></i>
          </td>
          <td class="actions">
            <button
              @click="openEditDialog(detail)"
              class="edit-btn"
              :class="{ disabled: detail.is_stock_updated }"
              :disabled="detail.is_stock_updated"
              :title="
                detail.is_stock_updated
                  ? 'Cannot edit - Stock has been updated'
                  : 'Edit this detail'
              "
            >
              <i class="fas fa-edit"></i>
              Edit
            </button>
            <button
              @click="confirmDelete(detail.id, detail.is_stock_updated)"
              class="delete-btn"
              :class="{ disabled: detail.is_stock_updated }"
              :disabled="detail.is_stock_updated"
              :title="
                detail.is_stock_updated
                  ? 'Cannot delete - Stock has been updated'
                  : 'Delete this detail'
              "
            >
              <i class="fas fa-trash-alt"></i>
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
      <div class="dialog-header">
        <h3>
          <i class="fas fa-edit"></i>
          Edit Purchase Detail
        </h3>
        <button class="close-btn" @click="showEditDialog = false">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <form @submit.prevent="handleEditSubmit">
        <div class="form-group">
          <label for="edit-qty">
            <i class="fas fa-hashtag"></i>
            Quantity
          </label>
          <input type="number" id="edit-qty" v-model="editingDetail.QTY" min="1" required />
        </div>

        <div class="form-group">
          <label for="edit-amount">
            <i class="fas fa-money-bill-wave"></i>
            Amount
          </label>
          <input
            type="number"
            id="edit-amount"
            v-model="editingDetail.amount"
            step="0.01"
            required
          />
        </div>

        <div class="form-row">
          <div class="form-group half">
            <label for="edit-year">
              <i class="fas fa-calendar-alt"></i>
              Year
            </label>
            <input type="number" id="edit-year" v-model="editingDetail.year" required />
          </div>

          <div class="form-group half">
            <label for="edit-month">
              <i class="fas fa-calendar-day"></i>
              Month
            </label>
            <input
              type="number"
              id="edit-month"
              v-model="editingDetail.month"
              min="1"
              max="12"
              required
            />
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
          />
        </div>

        <div class="form-group">
          <label for="edit-notes">Notes</label>
          <textarea id="edit-notes" v-model="editingDetail.notes" rows="3"></textarea>
        </div>

        <div class="form-group">
          <label for="edit-is-used-car">
            <i class="fas fa-car"></i>
            Used Car
          </label>
          <input type="checkbox" id="edit-is-used-car" v-model="editingDetail.is_used_car" />
        </div>

        <div class="form-group">
          <label for="edit-is-big-car">
            <i class="fas fa-truck"></i>
            Big Car
          </label>
          <input type="checkbox" id="edit-is-big-car" v-model="editingDetail.is_big_car" />
        </div>

        <div class="dialog-buttons">
          <button type="button" @click="showEditDialog = false" class="cancel-btn">Cancel</button>
          <button type="submit" class="submit-btn">Save Changes</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.detail-section {
  width: 100%;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  position: relative;
  min-height: 200px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.detail-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-header h3 i {
  color: #3b82f6;
}

.button-group {
  display: flex;
  gap: 12px;
}

.stock-btn,
.add-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.stock-btn {
  background-color: #059669;
  color: white;
}

.stock-btn:hover:not(:disabled) {
  background-color: #047857;
  transform: translateY(-1px);
}

.add-btn {
  background-color: #3b82f6;
  color: white;
}

.add-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #374151;
  white-space: nowrap;
}

.data-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.data-table td {
  color: #4b5563;
}

.actions {
  display: flex;
  gap: 8px;
}

.edit-btn,
.delete-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
  transform: translateY(-1px);
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Dialog styles */
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
  width: 90%;
  max-width: 500px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dialog-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  font-size: 1.25rem;
  padding: 4px;
  transition: color 0.2s ease;
}

.close-btn:hover {
  color: #374151;
}

form {
  padding: 20px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 500;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.form-group.half {
  flex: 1;
  margin-bottom: 0;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.dialog-buttons button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
}

.submit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
}

.submit-btn:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
}

/* Responsive styles */
@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .button-group {
    width: 100%;
    justify-content: flex-end;
  }

  .data-table {
    display: block;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .form-row {
    flex-direction: column;
    gap: 16px;
  }

  .form-group.half {
    width: 100%;
    margin-bottom: 16px;
  }
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 12px;
  z-index: 10;
}

.loading-overlay i {
  color: #3b82f6;
}

.loading-overlay span {
  color: #4b5563;
  font-size: 0.875rem;
}

/* Checkbox styling */
input[type='checkbox'] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

/* Status icons */
.text-success {
  color: #059669;
}

.text-danger {
  color: #dc2626;
}

/* Notes cell */
td:nth-child(8) {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

td:nth-child(8):hover {
  white-space: normal;
  overflow: visible;
  background-color: #f8fafc;
}
</style>
