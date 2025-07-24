<script setup>
import { computed } from 'vue'
import { useEnhancedI18n } from '../composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  type: {
    type: String,
    default: 'info',
    validator: (value) => ['warning', 'error', 'info', 'confirm'].includes(value),
  },
  title: {
    type: String,
    default: '',
  },
  message: {
    type: String,
    default: '',
  },
  details: {
    type: String,
    default: '',
  },
  confirmText: {
    type: String,
    default: 'OK',
  },
  cancelText: {
    type: String,
    default: 'Cancel',
  },
  showCancel: {
    type: Boolean,
    default: true,
  },
  showConfirm: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['confirm', 'cancel', 'close'])

// Computed properties for dynamic styling based on type
const typeConfig = computed(() => {
  const configs = {
    warning: {
      icon: 'fas fa-exclamation-triangle',
      headerGradient: 'linear-gradient(135deg, #ffc107, #ff8f00)',
      borderColor: '#ffc107',
      overlayGradient: 'linear-gradient(135deg, rgba(255, 193, 7, 0.1), rgba(220, 53, 69, 0.1))',
      detailsBg: 'linear-gradient(135deg, #fff3cd, #ffeaa7)',
      detailsBorder: '#ffc107',
      detailsColor: '#856404',
      confirmGradient: 'linear-gradient(135deg, #28a745, #20c997)',
      confirmHoverGradient: 'linear-gradient(135deg, #20c997, #17a2b8)',
      confirmShadow: 'rgba(40, 167, 69, 0.3)',
    },
    error: {
      icon: 'fas fa-times-circle',
      headerGradient: 'linear-gradient(135deg, #dc3545, #c82333)',
      borderColor: '#dc3545',
      overlayGradient: 'linear-gradient(135deg, rgba(220, 53, 69, 0.1), rgba(255, 193, 7, 0.1))',
      detailsBg: 'linear-gradient(135deg, #f8d7da, #f5c6cb)',
      detailsBorder: '#dc3545',
      detailsColor: '#721c24',
      confirmGradient: 'linear-gradient(135deg, #dc3545, #c82333)',
      confirmHoverGradient: 'linear-gradient(135deg, #c82333, #bd2130)',
      confirmShadow: 'rgba(220, 53, 69, 0.3)',
    },
    info: {
      icon: 'fas fa-info-circle',
      headerGradient: 'linear-gradient(135deg, #17a2b8, #138496)',
      borderColor: '#17a2b8',
      overlayGradient: 'linear-gradient(135deg, rgba(23, 162, 184, 0.1), rgba(255, 193, 7, 0.1))',
      detailsBg: 'linear-gradient(135deg, #d1ecf1, #bee5eb)',
      detailsBorder: '#17a2b8',
      detailsColor: '#0c5460',
      confirmGradient: 'linear-gradient(135deg, #17a2b8, #138496)',
      confirmHoverGradient: 'linear-gradient(135deg, #138496, #117a8b)',
      confirmShadow: 'rgba(23, 162, 184, 0.3)',
    },
    confirm: {
      icon: 'fas fa-question-circle',
      headerGradient: 'linear-gradient(135deg, #6f42c1, #5a32a3)',
      borderColor: '#6f42c1',
      overlayGradient: 'linear-gradient(135deg, rgba(111, 66, 193, 0.1), rgba(255, 193, 7, 0.1))',
      detailsBg: 'linear-gradient(135deg, #e2d9f3, #d1c7e8)',
      detailsBorder: '#6f42c1',
      detailsColor: '#4a148c',
      confirmGradient: 'linear-gradient(135deg, #28a745, #20c997)',
      confirmHoverGradient: 'linear-gradient(135deg, #20c997, #17a2b8)',
      confirmShadow: 'rgba(40, 167, 69, 0.3)',
    },
  }
  return configs[props.type] || configs.info
})

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
}

const handleClose = () => {
  emit('close')
}
</script>

<template>
  <div v-if="show" class="message-box-overlay" @click="handleClose">
    <div class="message-box" @click.stop>
      <div class="message-box-header" :style="{ background: typeConfig.headerGradient }">
        <div class="message-icon">
          <i :class="typeConfig.icon"></i>
        </div>
        <h3>{{ title }}</h3>
      </div>

      <div class="message-box-body">
        <div class="message-content">
          <p>{{ message }}</p>
          <div
            v-if="details"
            class="message-details"
            :style="{
              background: typeConfig.detailsBg,
              borderLeftColor: typeConfig.detailsBorder,
              color: typeConfig.detailsColor,
            }"
          >
            <i class="fas fa-info-circle"></i>
            <span>{{ details }}</span>
          </div>
        </div>
      </div>

      <div class="message-box-footer">
        <button v-if="showCancel" @click="handleCancel" class="cancel-btn">
          <i class="fas fa-times"></i>
          {{ cancelText }}
        </button>
        <button
          v-if="showConfirm"
          @click="handleConfirm"
          class="confirm-btn"
          :style="{
            background: typeConfig.confirmGradient,
            '--hover-gradient': typeConfig.confirmHoverGradient,
            '--shadow-color': typeConfig.confirmShadow,
          }"
        >
          <i class="fas fa-check"></i>
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-box-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
  padding: 20px;
  animation: fadeIn 0.3s ease-out;
}

.message-box-overlay[data-type='warning'] {
  background: linear-gradient(135deg, rgba(255, 193, 7, 0.1), rgba(220, 53, 69, 0.1));
}

.message-box-overlay[data-type='error'] {
  background: linear-gradient(135deg, rgba(220, 53, 69, 0.1), rgba(255, 193, 7, 0.1));
}

.message-box-overlay[data-type='info'] {
  background: linear-gradient(135deg, rgba(23, 162, 184, 0.1), rgba(255, 193, 7, 0.1));
}

.message-box-overlay[data-type='confirm'] {
  background: linear-gradient(135deg, rgba(111, 66, 193, 0.1), rgba(255, 193, 7, 0.1));
}

.message-box {
  background: linear-gradient(145deg, #ffffff, #f8f9fa);
  border-radius: 16px;
  box-shadow:
    0 20px 40px rgba(0, 0, 0, 0.1),
    0 0 0 1px rgba(0, 0, 0, 0.1),
    0 0 0 4px rgba(0, 0, 0, 0.05);
  width: 90%;
  max-width: 500px;
  overflow: hidden;
  animation: slideIn 0.4s ease-out;
  border: 2px solid;
}

.message-box-header {
  padding: 24px;
  text-align: center;
  position: relative;
}

.message-icon {
  margin-bottom: 16px;
}

.message-icon i {
  font-size: 48px;
  color: #ffffff;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  animation: pulse 2s infinite;
}

.message-box-header h3 {
  margin: 0;
  color: #ffffff;
  font-size: 24px;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.message-box-body {
  padding: 32px 24px;
}

.message-content p {
  font-size: 18px;
  color: #2c3e50;
  margin: 0 0 20px 0;
  line-height: 1.6;
  text-align: center;
  font-weight: 500;
}

.message-details {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-radius: 8px;
  border-left: 4px solid;
}

.message-details i {
  font-size: 18px;
}

.message-details span {
  font-size: 14px;
  line-height: 1.5;
}

.message-box-footer {
  padding: 24px;
  display: flex;
  gap: 16px;
  justify-content: center;
  background: #f8f9fa;
  border-top: 1px solid #e9ecef;
}

.cancel-btn,
.confirm-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 140px;
  justify-content: center;
}

.cancel-btn {
  background: linear-gradient(135deg, #6c757d, #5a6268);
  color: white;
}

.cancel-btn:hover {
  background: linear-gradient(135deg, #5a6268, #495057);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(108, 117, 125, 0.3);
}

.confirm-btn {
  color: white;
}

.confirm-btn:hover {
  background: var(--hover-gradient) !important;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--shadow-color);
}

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-50px) scale(0.9);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes pulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}
</style>
