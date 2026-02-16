<script setup>
/**
 * Print report header for sell/car print documents.
 * Uses the shared LetterHeader component so logo and company info
 * are consistent everywhere we print.
 */
import LetterHeader from '../shared/LetterHeader.vue'

defineProps({
  logoUrl: { type: String, default: '' },
  companyName: { type: String, default: '' },
  companyAddress: { type: String, default: '' },
  contactItems: {
    type: Array,
    default: () => [],
  },
})
</script>

<template>
  <div class="print-report-header">
    <LetterHeader
      :logo-url="logoUrl"
      :company-name="companyName"
      :company-address="companyAddress"
      :contact-items="contactItems"
    />
  </div>
</template>

<style scoped>
.print-report-header {
  margin-bottom: 0; /* LetterHeader has its own margin */
}
</style>
