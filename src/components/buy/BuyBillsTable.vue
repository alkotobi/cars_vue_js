<script setup>
import { defineProps, defineEmits } from 'vue'
import { ref, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  buyBills: {
    type: Array,
    required: true
  },
  selectedBill: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['select-bill'])

const { getFileUrl } = useApi()

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Date(dateStr).toLocaleDateString()
}

const formatNumber = (value) => {
  const num = Number(value)
  return !isNaN(num) ? num.toFixed(2) : 'N/A'
}
</script>

<template>
  <div class="master-section">
    <!-- Toolbar -->
    <div class="toolbar" v-if="selectedBill">
      <div class="bill-info">
        <div class="bill-id">
          Purchase #{{ selectedBill.id }}
        </div>
        <div class="bill-details">
          <div class="detail-item">
            <span class="label">Supplier:</span>
            <span class="value">{{ selectedBill.supplier_name }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Amount:</span>
            <span class="value">{{ formatNumber(selectedBill.amount) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Paid:</span>
            <span class="value">{{ formatNumber(selectedBill.payed) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Balance:</span>
            <span class="value">{{ formatNumber(selectedBill.amount - selectedBill.payed) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">Status:</span>
            <span class="value" :class="selectedBill.is_stock_updated ? 'status-updated' : 'status-pending'">
              {{ selectedBill.is_stock_updated ? 'Updated' : 'Pending' }}
            </span>
          </div>
        </div>
      </div>
      <div class="toolbar-actions">
        <slot name="actions" :bill="selectedBill"></slot>
      </div>
    </div>

    <!-- Table -->
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Date</th>
          <th>Supplier</th>
          <th>Bill Ref</th>
          <th>Amount</th>
          <th>Paid</th>
          <th>Balance</th>
          <th>Status</th>
          <th>PI Document</th>
        </tr>
      </thead>
      <tbody>
        <tr 
          v-for="bill in buyBills" 
          :key="bill.id"
          :class="{ 'selected': selectedBill?.id === bill.id }"
          @click="$emit('select-bill', bill)"
        >
          <td>{{ bill.id }}</td>
          <td>{{ formatDate(bill.date_buy) }}</td>
          <td>{{ bill.supplier_name }}</td>
          <td>{{ bill.bill_ref || 'N/A' }}</td>
          <td>{{ formatNumber(bill.amount) }}</td>
          <td>{{ formatNumber(bill.payed) }}</td>
          <td>{{ formatNumber(bill.amount - bill.payed) }}</td>
          <td>
            <span :class="bill.is_stock_updated ? 'status-updated' : 'status-pending'">
              {{ bill.is_stock_updated ? 'Updated' : 'Pending' }}
            </span>
          </td>
          <td>
            <a 
              v-if="bill.pi_path" 
              :href="getFileUrl(bill.pi_path)"
              target="_blank"
              class="pi-document-link"
              @click.stop
            >
              View PI
            </a>
            <span v-else class="no-document">No document</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.master-section {
  width: 100%;
}

.toolbar {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.bill-info {
  flex: 1;
}

.bill-id {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 8px;
}

.bill-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.label {
  color: #64748b;
  font-size: 0.875rem;
}

.value {
  font-weight: 500;
  color: #1e293b;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
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
  color: #374151;
}

.data-table tbody tr {
  cursor: pointer;
  transition: background-color 0.2s;
}

.data-table tbody tr:hover {
  background-color: #f3f4f6;
}

.data-table tbody tr.selected {
  background-color: #e5e7eb;
}

.data-table td {
  color: #4b5563;
}

.status-updated {
  color: #059669;
  font-weight: 500;
}

.status-pending {
  color: #d97706;
  font-weight: 500;
}

.pi-document-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}

.pi-document-link:hover {
  background-color: #dbeafe;
  text-decoration: underline;
}

.no-document {
  color: #9ca3af;
  font-size: 0.875rem;
  font-style: italic;
}

/* Responsive styles */
@media (max-width: 1024px) {
  .toolbar {
    flex-direction: column;
    gap: 16px;
  }

  .toolbar-actions {
    width: 100%;
    justify-content: flex-start;
  }

  .bill-details {
    flex-direction: column;
    gap: 8px;
  }
}
</style>