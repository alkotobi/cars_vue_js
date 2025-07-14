<template>
  <div class="alert-cars-view">
    <div class="alert-header">
      <h2>{{ alertTitle }}</h2>
      <p class="alert-description">{{ alertDescription }}</p>
    </div>

    <CarStockTable :alertType="alertType" :alertDays="alertDays" :showAlertFilter="true" />
  </div>
</template>

<script>
import CarStockTable from '@/components/car-stock/CarStockTable.vue'

export default {
  name: 'AlertCarsView',
  components: {
    CarStockTable,
  },
  data() {
    return {
      alertType: '',
      alertDays: 0,
    }
  },
  computed: {
    alertTitle() {
      const titles = {
        unloaded: 'Cars Not Unloaded',
        not_arrived: 'Cars Not Arrived',
        no_licence: 'Cars Without Export License',
        no_docs_sent: 'Cars Without Documents Sent',
      }
      return titles[this.alertType] || 'Alert Cars'
    },
    alertDescription() {
      const descriptions = {
        unloaded: `Cars sold more than ${this.alertDays} days ago that haven't been unloaded`,
        not_arrived: `Cars purchased more than ${this.alertDays} days ago that haven't arrived at port`,
        no_licence: `Cars purchased more than ${this.alertDays} days ago without export license`,
        no_docs_sent: `Cars sold more than ${this.alertDays} days ago without documents sent`,
      }
      return descriptions[this.alertType] || ''
    },
  },
  mounted() {
    // Get alert type and days from route params
    this.alertType = this.$route.params.alertType || ''
    this.alertDays = parseInt(this.$route.params.alertDays) || 0
  },
}
</script>

<style scoped>
.alert-cars-view {
  padding: 20px;
}

.alert-header {
  margin-bottom: 20px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #dc3545;
}

.alert-header h2 {
  margin: 0 0 10px 0;
  color: #dc3545;
  font-size: 1.5rem;
}

.alert-description {
  margin: 0;
  color: #6c757d;
  font-size: 0.9rem;
}
</style>
