<script setup>
import { useRoute } from 'vue-router'
import SellBillPrintDocument from '../components/sells/SellBillPrintDocument.vue'

const route = useRoute()
const billId = parseInt(route.params.billId)
const options = JSON.parse(decodeURIComponent(route.query.options || '{}'))
</script>

<template>
  <SellBillPrintDocument 
    :billId="billId"
    :options="options"
  />
</template> 