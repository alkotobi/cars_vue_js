<script setup>
import { defineProps, defineEmits } from 'vue'
import { ref, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  buyBills: {
    type: Array,
    required: true,
  },
  selectedBill: {
    type: Object,
    default: null,
  },
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
          <i class="fas fa-file-invoice-dollar"></i>
          Purchase #{{ selectedBill.id }}
        </div>
        <div class="bill-details">
          <div class="detail-item">
            <i class="fas fa-building"></i>
            <span class="label">Supplier:</span>
            <span class="value">{{ selectedBill.supplier_name }}</span>
          </div>
          <div class="detail-item">
            <i class="fas fa-money-bill-wave"></i>
            <span class="label">Amount:</span>
            <span class="value">{{ formatNumber(selectedBill.amount) }}</span>
          </div>
          <div class="detail-item">
            <i class="fas fa-check-circle"></i>
            <span class="label">Paid:</span>
            <span class="value">{{ formatNumber(selectedBill.payed) }}</span>
          </div>
          <div class="detail-item">
            <i class="fas fa-balance-scale"></i>
            <span class="label">Balance:</span>
            <span class="value">{{ formatNumber(selectedBill.amount - selectedBill.payed) }}</span>
          </div>
          <div class="detail-item">
            <i
              class="fas"
              :class="
                selectedBill.is_stock_updated ? 'fa-check text-success' : 'fa-clock text-warning'
              "
            ></i>
            <span class="label">Status:</span>
            <span
              class="value"
              :class="selectedBill.is_stock_updated ? 'status-updated' : 'status-pending'"
            >
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
          <th><i class="fas fa-hashtag"></i> ID</th>
          <th><i class="fas fa-calendar"></i> Date</th>
          <th><i class="fas fa-building"></i> Supplier</th>
          <th><i class="fas fa-file-alt"></i> Bill Ref</th>
          <th><i class="fas fa-money-bill-wave"></i> Amount</th>
          <th><i class="fas fa-check-circle"></i> Paid</th>
          <th><i class="fas fa-balance-scale"></i> Balance</th>
          <th><i class="fas fa-info-circle"></i> Status</th>
          <th><i class="fas fa-file-pdf"></i> PI Document</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="bill in buyBills"
          :key="bill.id"
          :class="{ selected: selectedBill?.id === bill.id }"
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
              <i class="fas" :class="bill.is_stock_updated ? 'fa-check' : 'fa-clock'"></i>
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
              <i class="fas fa-file-pdf"></i>
              View PI
            </a>
            <span v-else class="no-document">
              <i class="fas fa-times-circle"></i>
              No document
            </span>
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
  padding: 20px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.bill-info {
  flex: 1;
}

.bill-id {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.bill-id i {
  color: #3b82f6;
}

.bill-details {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: white;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.detail-item i {
  color: #6b7280;
  width: 16px;
  text-align: center;
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
  border-collapse: separate;
  border-spacing: 0;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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

.data-table tbody tr {
  cursor: pointer;
  transition: all 0.2s ease;
}

.data-table tbody tr:hover {
  background-color: #f8fafc;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.data-table tbody tr.selected {
  background-color: #eff6ff;
  border-left: 4px solid #3b82f6;
}

.data-table td {
  color: #4b5563;
}

.status-updated {
  color: #059669;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-pending {
  color: #d97706;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.text-success {
  color: #059669;
}

.text-warning {
  color: #d97706;
}

.pi-document-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  padding: 6px 12px;
  border-radius: 6px;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: #eff6ff;
}

.pi-document-link:hover {
  background-color: #dbeafe;
  transform: translateY(-1px);
}

.pi-document-link i {
  color: #3b82f6;
}

.no-document {
  color: #9ca3af;
  font-size: 0.875rem;
  font-style: italic;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.no-document i {
  color: #9ca3af;
}

/* Responsive styles */
@media (max-width: 1024px) {
  .toolbar {
    flex-direction: column;
    gap: 16px;
  }

  .bill-details {
    flex-direction: column;
    gap: 12px;
  }

  .detail-item {
    width: 100%;
  }

  .toolbar-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

@media (max-width: 768px) {
  .data-table {
    display: block;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
}
</style>
