<script setup>
import { useRoute } from 'vue-router'
import CarPrintDocument from '../components/sells/CarPrintDocument.vue'

const route = useRoute()
const carId = parseInt(route.params.carId)
const billId = parseInt(route.query.billId)
const options = JSON.parse(decodeURIComponent(route.query.options || '{}'))
</script>

<template>
  <CarPrintDocument :carId="carId" :billId="billId" :options="options" />
</template>
