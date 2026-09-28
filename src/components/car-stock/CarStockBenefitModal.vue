<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  selectedCars: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['close'])

const { t } = useI18n()

const toNumber = (value) => {
  if (value === null || value === undefined || value === '') return null
  const parsed = parseFloat(value)
  return isNaN(parsed) ? null : parsed
}

const soldCars = computed(() =>
  props.selectedCars.filter((car) => car.id_sell !== null && car.id_sell !== undefined),
)

const skippedCars = computed(() => props.selectedCars.length - soldCars.value.length)

const rows = computed(() =>
  soldCars.value.map((car) => {
    const sellPrice = toNumber(car.price_cell)
    const costPrice = toNumber(car.cost_price)
    const benefit = sellPrice !== null && costPrice !== null ? sellPrice - costPrice : null
    return {
      id: car.id,
      clientName: car.client_name,
      vin: car.vin,
      sellBillRef: car.sell_bill_ref,
      sellPrice,
      costPrice,
      benefit,
    }
  }),
)

const countedRows = computed(() => rows.value.filter((row) => row.benefit !== null))
const missingPriceCount = computed(() => rows.value.length - countedRows.value.length)

const totalSell = computed(() => countedRows.value.reduce((sum, row) => sum + row.sellPrice, 0))
const totalCost = computed(() => countedRows.value.reduce((sum, row) => sum + row.costPrice, 0))
const totalBenefit = computed(() => countedRows.value.reduce((sum, row) => sum + row.benefit, 0))

const formatMoney = (value) => {
  if (value === null) return '—'
  return '$' + value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click="emit('close')">
      <div class="modal-content" @click.stop>
        <div class="benefit-form">
          <div class="form-header">
            <h3>
              <i class="fas fa-chart-line"></i>
              {{ t('carStockBenefitModal.title') }}
            </h3>
            <button class="close-btn" @click="emit('close')">
              <i class="fas fa-times"></i>
            </button>
          </div>

          <div class="form-content">
            <div class="report-summary">
              <div class="summary-item">
                <span class="summary-label">{{ t('carStockBenefitModal.cars_sold') }}</span>
                <span class="summary-value">{{ soldCars.length }}</span>
              </div>
              <div v-if="skippedCars > 0" class="summary-item">
                <span class="summary-label">{{
                  t('carStockBenefitModal.cars_not_sold_ignored')
                }}</span>
                <span class="summary-value">{{ skippedCars }}</span>
              </div>
              <div v-if="missingPriceCount > 0" class="summary-item">
                <span class="summary-label">{{
                  t('carStockBenefitModal.cars_without_price')
                }}</span>
                <span class="summary-value">{{ missingPriceCount }}</span>
              </div>
            </div>

            <p v-if="missingPriceCount > 0" class="summary-note">
              <i class="fas fa-info-circle"></i>
              {{ t('carStockBenefitModal.missing_price_note') }}
            </p>

            <div class="table-wrapper">
              <table class="benefit-table">
                <thead>
                  <tr>
                    <th>#</th>
                    <th class="client-col">{{ t('carStock.client_name') }}</th>
                    <th>{{ t('carStockBenefitModal.vin') }}</th>
                    <th>{{ t('carStockBenefitModal.sell_bill') }}</th>
                    <th class="numeric">{{ t('carStockBenefitModal.sell_price') }}</th>
                    <th class="numeric">{{ t('carStockBenefitModal.buy_cost') }}</th>
                    <th class="numeric">{{ t('carStockBenefitModal.benefit') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, index) in rows" :key="row.id">
                    <td>{{ index + 1 }}</td>
                    <td class="client-col">{{ row.clientName || '—' }}</td>
                    <td>{{ row.vin || '—' }}</td>
                    <td>{{ row.sellBillRef || '—' }}</td>
                    <td class="numeric">{{ formatMoney(row.sellPrice) }}</td>
                    <td class="numeric">{{ formatMoney(row.costPrice) }}</td>
                    <td
                      class="numeric benefit-cell"
                      :class="{
                        positive: row.benefit > 0,
                        negative: row.benefit < 0,
                        unknown: row.benefit === null,
                      }"
                    >
                      {{ formatMoney(row.benefit) }}
                    </td>
                  </tr>
                </tbody>
                <tfoot>
                  <tr>
                    <td colspan="4">{{ t('carStockBenefitModal.total') }}</td>
                    <td class="numeric">{{ formatMoney(totalSell) }}</td>
                    <td class="numeric">{{ formatMoney(totalCost) }}</td>
                    <td
                      class="numeric benefit-cell"
                      :class="{ positive: totalBenefit > 0, negative: totalBenefit < 0 }"
                    >
                      {{ formatMoney(totalBenefit) }}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>

            <div class="form-actions">
              <button class="close-form-btn" @click="emit('close')">
                {{ t('carStockBenefitModal.close') }}
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
  max-width: 1200px;
  width: 95%;
  max-height: 90vh;
  overflow-y: auto;
}

.benefit-form {
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
  color: #16a34a;
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

.close-btn:hover {
  background-color: #f3f4f6;
  color: #1f2937;
}

.form-content {
  padding: 24px;
}

.report-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background-color: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
}

.summary-label {
  font-size: 13px;
  color: #6b7280;
}

.summary-value {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.summary-note {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 16px;
  padding: 10px 14px;
  background-color: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 6px;
  font-size: 13px;
  color: #92400e;
}

.table-wrapper {
  max-height: 60vh;
  max-height: calc(90vh - 250px);
  min-height: 120px;
  overflow: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.benefit-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 14px;
}

.benefit-table th,
.benefit-table td {
  padding: 10px 10px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
  white-space: nowrap;
}

.benefit-table th.client-col,
.benefit-table td.client-col {
  white-space: normal;
  min-width: 140px;
  max-width: 220px;
  word-break: break-word;
}

.benefit-table thead th {
  position: sticky;
  top: 0;
  z-index: 2;
  background-color: #f8fafc;
  font-weight: 600;
  color: #374151;
  box-shadow: inset 0 -1px 0 #e5e7eb;
}

.benefit-table tbody tr:hover {
  background-color: #f9fafb;
}

.benefit-table tfoot td {
  position: sticky;
  bottom: 0;
  z-index: 1;
  font-weight: 700;
  color: #1f2937;
  background-color: #f8fafc;
  border-bottom: none;
  box-shadow: inset 0 1px 0 #e5e7eb;
}

.benefit-table .numeric {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.benefit-cell.positive {
  color: #16a34a;
  font-weight: 600;
}

.benefit-cell.negative {
  color: #dc2626;
  font-weight: 600;
}

.benefit-cell.unknown {
  color: #9ca3af;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.close-form-btn {
  padding: 10px 20px;
  background-color: #4f46e5;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.close-form-btn:hover {
  background-color: #4338ca;
}
</style>
