<script setup>
import { useRoute } from 'vue-router'
import { onMounted, nextTick } from 'vue'
import SellBillPrintDocument from '../components/sells/SellBillPrintDocument.vue'
import { usePrintOnlyStyles } from '../composables/usePrintOnlyStyles'

const route = useRoute()
const billId = parseInt(route.params.billId)
const options = JSON.parse(decodeURIComponent(route.query.options || '{}'))

const { applyPrintOnlyStyles } = usePrintOnlyStyles()

onMounted(() => {
  if (route.query.printOnly === '1') {
    nextTick(() => {
      setTimeout(applyPrintOnlyStyles, 100)
    })
  }
})
</script>

<template>
  <SellBillPrintDocument 
    :billId="billId"
    :options="options"
  />
</template> 