<script setup>
import { defineProps, computed } from 'vue'
import { useApi } from '../../composables/useApi'

// Import the letter head image using the proper Vue.js way
const letterHeadUrl = new URL('../../assets/letter_head.png', import.meta.url).href

const { getFileUrl } = useApi()

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  cars: {
    type: Array,
    default: () => [],
  },
  columns: {
    type: Array,
    default: () => [],
  },
  contentBeforeTable: {
    type: String,
    default: '',
  },
  contentAfterTable: {
    type: String,
    default: '',
  },
  showHeader: {
    type: Boolean,
    default: true,
  },
})

const today = computed(() => {
  return new Date().toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
})

const getCarValue = (car, columnKey) => {
  const value = car[columnKey]

  // Handle different data types
  if (value === null || value === undefined) return '-'

  // Handle client ID picture - render as image
  if (columnKey === 'client_id_picture' && value) {
    return `<img src="${getFileUrl(value)}" alt="Client ID" style="max-width: 100px; max-height: 60px; object-fit: contain;" />`
  }

  // Handle dates
  if (columnKey.includes('date') && value) {
    return new Date(value).toLocaleDateString()
  }

  // Handle numbers
  if (['price_cell', 'freight', 'rate', 'cfr_usd', 'cfr_dza'].includes(columnKey) && value) {
    return parseFloat(value).toLocaleString()
  }

  return value.toString()
}
</script>

<template>
  <div class="print-report">
    <!-- Header Image -->
    <div v-if="showHeader" class="report-header">
      <img
        :src="letterHeadUrl"
        alt="Letter Head"
        class="letter-head"
        @error="$event.target.style.display = 'none'"
      />
    </div>

    <!-- Date -->
    <div class="report-date">
      <span>{{ today }}</span>
    </div>

    <!-- Title -->
    <h1 class="report-title">{{ title }}</h1>

    <!-- Content Before Table -->
    <div v-if="contentBeforeTable" class="content-before-table" v-html="contentBeforeTable"></div>

    <!-- Table -->
    <div class="table-container">
      <table class="report-table">
        <thead>
          <tr>
            <th v-for="column in columns" :key="column.key" class="table-header">
              {{ column.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(car, index) in cars" :key="car.id || index" class="table-row">
            <td v-for="column in columns" :key="column.key" class="table-cell">
              <span
                v-if="column.key === 'client_id_picture'"
                v-html="getCarValue(car, column.key)"
              ></span>
              <span v-else>{{ getCarValue(car, column.key) }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Content After Table -->
    <div v-if="contentAfterTable" class="content-after-table" v-html="contentAfterTable"></div>
  </div>
</template>

<style scoped>
.print-report {
  font-family: 'Arial', sans-serif;
  max-width: 210mm; /* A4 width */
  margin: 0 auto;
  padding: 20px;
  background: white;
  color: #333;
}

.report-header {
  text-align: center;
  margin-bottom: 20px;
}

.letter-head {
  max-width: 100%;
  height: auto;
  max-height: 120px;
}

.report-date {
  text-align: right;
  margin-bottom: 20px;
  font-size: 14px;
  color: #666;
  float: right;
  clear: both;
  width: 100%;
}

.report-title {
  text-align: center;
  margin-bottom: 30px;
  font-size: 24px;
  font-weight: bold;
  color: #333;
  text-transform: uppercase;
  clear: both;
}

.content-before-table {
  margin-bottom: 20px;
  line-height: 1.6;
  font-size: 14px;
}

.table-container {
  margin-bottom: 20px;
  overflow-x: auto;
}

.report-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.table-header {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  padding: 8px 12px;
  text-align: left;
  font-weight: bold;
  color: #495057;
  white-space: nowrap;
}

.table-row:nth-child(even) {
  background-color: #f8f9fa;
}

.table-cell {
  border: 1px solid #dee2e6;
  padding: 8px 12px;
  text-align: left;
  vertical-align: top;
}

.table-row:hover {
  background-color: #e9ecef;
}

.content-after-table {
  margin-top: 20px;
  line-height: 1.6;
  font-size: 14px;
}

/* Print styles */
@media print {
  .print-report {
    padding: 0;
    margin: 0;
  }

  .report-table {
    font-size: 10px;
  }

  .table-header,
  .table-cell {
    padding: 6px 8px;
  }

  .letter-head {
    max-height: 80px;
  }

  .report-title {
    font-size: 20px;
    margin-bottom: 20px;
  }
}

/* Mobile responsive */
@media (max-width: 768px) {
  .print-report {
    padding: 10px;
  }

  .report-table {
    font-size: 10px;
  }

  .table-header,
  .table-cell {
    padding: 4px 6px;
  }
}
</style>
