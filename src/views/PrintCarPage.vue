<script setup>
import { useRoute } from 'vue-router'
import { onMounted, nextTick } from 'vue'
import CarPrintDocument from '../components/sells/CarPrintDocument.vue'
import { usePrintOnlyStyles } from '../composables/usePrintOnlyStyles'

const route = useRoute()
const carId = parseInt(route.params.carId)
const billId = parseInt(route.query.billId)
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
  <CarPrintDocument :carId="carId" :billId="billId" :options="options" />
</template>
