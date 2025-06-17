<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  bills: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['view', 'edit', 'delete'])

const filteredBills = computed(() => {
  const bills = props.bills || []
  return bills.slice(0, 5) // Limit to 5 rows
})

const formatCurrency = (value) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value || 0)
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const handleView = (bill) => {
  emit('view', bill)
}

const handleEdit = (bill) => {
  emit('edit', bill)
}

const handleDelete = (bill) => {
  emit('delete', bill)
}
</script>

<template>
  <div class="table-container">
    <table class="cars-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Client</th>
          <th>Car</th>
          <th>Price</th>
          <th>Date</th>
          <th>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="bill in filteredBills" :key="bill.id">
          <td>#{{ bill.id }}</td>
          <td>{{ bill.client_name }}</td>
          <td>{{ bill.car_name }}</td>
          <td>{{ formatCurrency(bill.price) }}</td>
          <td>{{ formatDate(bill.date) }}</td>
          <td>
            <span :class="['status-badge', bill.status.toLowerCase()]">
              {{ bill.status }}
            </span>
          </td>
          <td>
            <div class="actions">
              <button @click="handleView(bill)" class="action-btn view">
                <i class="fas fa-eye"></i>
              </button>
              <button @click="handleEdit(bill)" class="action-btn edit">
                <i class="fas fa-edit"></i>
              </button>
              <button @click="handleDelete(bill)" class="action-btn delete">
                <i class="fas fa-trash"></i>
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-container {
  width: 100%;
  overflow-y: auto;
  height: 300px; /* Height for exactly 5 rows */
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.cars-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.cars-table thead {
  position: sticky;
  top: 0;
  background: white;
  z-index: 1;
}

.cars-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
  background: white;
}

.cars-table td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
  height: 48px; /* Fixed height for each row */
  box-sizing: border-box;
}

.cars-table tbody tr:last-child td {
  border-bottom: none;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
}

.status-badge.pending {
  background-color: #fef3c7;
  color: #92400e;
}

.status-badge.completed {
  background-color: #d1fae5;
  color: #065f46;
}

.status-badge.cancelled {
  background-color: #fee2e2;
  color: #991b1b;
}

.actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  padding: 6px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.action-btn.view {
  background-color: #e3f2fd;
  color: #1976d2;
}

.action-btn.edit {
  background-color: #fff3e0;
  color: #f57c00;
}

.action-btn.delete {
  background-color: #ffebee;
  color: #d32f2f;
}

.action-btn:hover {
  opacity: 0.8;
}
</style>
 