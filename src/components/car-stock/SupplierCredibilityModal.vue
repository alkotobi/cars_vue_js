<script setup>
import { computed } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { CREDIBILITY_RETRYABLE_CODES, formatCheckedAt } from '../../lib/supplierCredibility'

/**
 * Shows one supplier's latest credibility check, its red flags, and every past
 * check.
 *
 * It renders only what the API sent; the wording of the assessment is the
 * model's, in the language it was asked for. What this component owns is the
 * framing around it: the score is a second opinion, not a verification, and the
 * modal says so where the number is.
 */
const props = defineProps({
  supplier: { type: Object, required: true },
  check: { type: Object, default: null },
  history: { type: Array, default: () => [] },
  error: { type: String, default: '' },
  errorCode: { type: String, default: '' },
  isLoading: { type: Boolean, default: false },
  isChecking: { type: Boolean, default: false },
})

const emit = defineEmits(['close', 'recheck'])

const { t, locale } = useEnhancedI18n()

// A busy model or a rate limit clears on its own, so the offer to run it again
// is the useful thing to show - unlike a rejected key, which needs a human.
const canRetryError = computed(() => CREDIBILITY_RETRYABLE_CODES.includes(props.errorCode))

const riskLabels = computed(() => ({
  low: t('supplierCredibility.risk.low'),
  medium: t('supplierCredibility.risk.medium'),
  high: t('supplierCredibility.risk.high'),
}))

const confidenceLabels = computed(() => ({
  low: t('supplierCredibility.confidence.low'),
  medium: t('supplierCredibility.confidence.medium'),
  high: t('supplierCredibility.confidence.high'),
}))

const riskLabel = (level) => riskLabels.value[level] ?? level
const confidenceLabel = (level) => confidenceLabels.value[level] ?? level
const checkedAt = (value) => formatCheckedAt(value, locale.value)

const scoreClass = computed(() => {
  const level = props.check?.risk_level
  return level ? `risk-${level}` : 'risk-unknown'
})

// History minus the check already shown above, so the newest is not repeated.
const olderChecks = computed(() =>
  props.check ? props.history.filter((row) => row.id !== props.check.id) : props.history,
)

// The model can only ever claim its own memory of a company's court history, so
// the modal labels that case rather than presenting it as something it checked.
const claimsRecollection = computed(() => props.check?.courts_basis === 'recollection')
</script>

<template>
  <div class="credibility-overlay" @click.self="emit('close')">
    <div class="credibility-dialog" role="dialog" aria-modal="true">
      <header class="credibility-header">
        <div>
          <h3>
            <i class="fas fa-sticky-note"></i>
            {{ t('supplierCredibility.title') }}
          </h3>
          <p class="supplier-name">{{ supplier.name }}</p>
        </div>
        <button
          class="close-btn"
          type="button"
          :aria-label="t('common.close')"
          @click="emit('close')"
        >
          <i class="fas fa-times"></i>
        </button>
      </header>

      <div class="credibility-content">
        <div v-if="isChecking" class="state-block">
          <i class="fas fa-spinner fa-spin"></i>
          <p>{{ t('supplierCredibility.running') }}</p>
        </div>

        <div v-else-if="isLoading" class="state-block">
          <i class="fas fa-spinner fa-spin"></i>
          <p>{{ t('supplierCredibility.loading') }}</p>
        </div>

        <div v-else-if="error" class="credibility-error" role="alert">
          <i class="fas fa-exclamation-triangle"></i>
          <span>{{ error }}</span>
          <button v-if="canRetryError" class="retry-inline" @click="emit('recheck')">
            {{ t('supplierCredibility.runAgain') }}
          </button>
        </div>

        <template v-else-if="check">
          <div class="score-row">
            <div class="score" :class="scoreClass">
              <span class="score-value">{{ check.score ?? '—' }}</span>
              <span class="score-max">/100</span>
            </div>
            <div class="score-detail">
              <span class="risk-pill" :class="`risk-${check.risk_level}`">
                {{ t('supplierCredibility.riskLabel') }}: {{ riskLabel(check.risk_level) }}
              </span>
              <p class="disclaimer">{{ t('supplierCredibility.disclaimer') }}</p>
            </div>
          </div>

          <section>
            <h4>
              <i class="fas fa-user-check"></i>
              {{ t('supplierCredibility.credibleQuestion') }}
            </h4>
            <p class="summary">{{ check.summary }}</p>
          </section>

          <section>
            <h4>
              <i class="fas fa-scale-balanced"></i>
              {{ t('supplierCredibility.courtsQuestion') }}
            </h4>
            <p class="summary">{{ check.court_records || '—' }}</p>
            <p v-if="claimsRecollection" class="basis-note">
              <i class="fas fa-exclamation-circle"></i>
              {{ t('supplierCredibility.recollectionNote') }}
            </p>
          </section>

          <section v-if="check.red_flags.length">
            <h4>{{ t('supplierCredibility.redFlags') }}</h4>
            <ul class="red-flags">
              <li v-for="(flag, index) in check.red_flags" :key="index">{{ flag }}</li>
            </ul>
          </section>

          <dl class="meta">
            <div>
              <dt>{{ t('supplierCredibility.checkedAt') }}</dt>
              <dd>{{ checkedAt(check.checked_at) }}</dd>
            </div>
            <div>
              <dt>{{ t('supplierCredibility.confidenceLabel') }}</dt>
              <dd>{{ confidenceLabel(check.confidence) }}</dd>
            </div>
            <div>
              <dt>{{ t('supplierCredibility.model') }}</dt>
              <dd>{{ check.model || '—' }}</dd>
            </div>
            <div v-if="check.username">
              <dt>{{ t('supplierCredibility.requestedBy') }}</dt>
              <dd>{{ check.username }}</dd>
            </div>
          </dl>

          <section v-if="olderChecks.length">
            <h4>{{ t('supplierCredibility.history') }}</h4>
            <ul class="history">
              <li v-for="row in olderChecks" :key="row.id">
                <span class="history-score" :class="`risk-${row.risk_level}`">
                  {{ row.score ?? '—' }}
                </span>
                <span class="history-date">{{ checkedAt(row.checked_at) }}</span>
                <span class="history-model">{{ row.model }}</span>
              </li>
            </ul>
          </section>
        </template>

        <div v-else class="state-block">
          <p>{{ t('supplierCredibility.neverChecked') }}</p>
        </div>
      </div>

      <footer class="credibility-actions">
        <button class="btn cancel-btn" @click="emit('close')">{{ t('common.close') }}</button>
        <button class="btn run-btn" :disabled="isChecking || isLoading" @click="emit('recheck')">
          <i class="fas fa-magnifying-glass-chart"></i>
          {{ check ? t('supplierCredibility.runAgain') : t('supplierCredibility.run') }}
        </button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.credibility-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10010;
  padding: 20px;
  backdrop-filter: blur(4px);
}

.credibility-dialog {
  background: #fff;
  border-radius: 12px;
  width: 100%;
  max-width: 800px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.credibility-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.credibility-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.credibility-header h3 i {
  color: #3b82f6;
}

.supplier-name {
  margin: 0.25rem 0 0;
  color: #6b7280;
  font-size: 0.9rem;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
  font-size: 1.25rem;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #ef4444;
}

/* The body scrolls on its own so a long history cannot push the buttons the
   user needs to reach off the bottom of the card. */
.credibility-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.state-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px 0;
  color: #6b7280;
}

.credibility-error {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border: 1px solid #fecaca;
  border-radius: 6px;
  background: #fef2f2;
  color: #b91c1c;
}

.retry-inline {
  margin-left: auto;
  background: none;
  border: 1px solid currentColor;
  border-radius: 4px;
  color: inherit;
  padding: 4px 10px;
  cursor: pointer;
  white-space: nowrap;
}

.score-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.score {
  display: flex;
  align-items: baseline;
  gap: 2px;
  padding: 12px 16px;
  border-radius: 8px;
  color: #fff;
  font-weight: 700;
}

.score-value {
  font-size: 1.75rem;
}

.score-max {
  font-size: 0.85rem;
  opacity: 0.85;
}

.risk-low,
.history-score.risk-low {
  background-color: #10b981;
}

.risk-medium,
.history-score.risk-medium {
  background-color: #f59e0b;
}

.risk-high,
.history-score.risk-high {
  background-color: #dc3545;
}

.risk-unknown,
.history-score.risk-unknown {
  background-color: #9ca3af;
}

.score-detail {
  flex: 1;
}

.risk-pill {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 999px;
  color: #fff;
  font-size: 0.8rem;
  font-weight: 600;
}

.disclaimer {
  margin: 8px 0 0;
  font-size: 0.82rem;
  color: #6b7280;
  line-height: 1.4;
}

h4 {
  margin: 0 0 6px;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

h4 i {
  color: #3b82f6;
}

/* The recorded legal issues read as a quoted note rather than as a heading, so
   nobody mistakes a person's own record for something the model verified. */
.basis-note {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px 14px;
  background: #f8fafc;
}

.summary {
  margin: 0;
  line-height: 1.5;
}

.red-flags {
  margin: 0;
  padding-left: 20px;
  line-height: 1.5;
}

.meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 10px;
  margin: 0;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
}

.meta dt {
  font-size: 0.75rem;
  text-transform: uppercase;
  color: #6b7280;
}

.meta dd {
  margin: 2px 0 0;
  font-size: 0.9rem;
}

.history {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.history li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.85rem;
}

.history-score {
  color: #fff;
  border-radius: 4px;
  padding: 1px 8px;
  font-weight: 600;
  min-width: 40px;
  text-align: center;
}

.history-date {
  color: #374151;
}

.history-model {
  color: #9ca3af;
  margin-left: auto;
}

.credibility-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  border-top: 1px solid #e5e7eb;
  padding: 1rem 1.5rem;
}

.btn {
  padding: 8px 14px;
  border: none;
  border-radius: 4px;
  color: #fff;
  cursor: pointer;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cancel-btn {
  background: #6b7280;
}

.run-btn {
  background: #2563eb;
}

.run-btn i {
  margin-right: 6px;
}
</style>
