<template>
  <div class="language-switcher">
    <div class="language-dropdown" @click="toggleDropdown" ref="dropdown">
      <div class="language-current">
        <span class="language-flag">{{ getLanguageFlag(currentLocale) }}</span>
        <span class="language-name">{{ getLanguageName(currentLocale) }}</span>
        <i class="fas fa-chevron-down" :class="{ rotated: isOpen }"></i>
      </div>

      <div class="language-options" v-if="isOpen" @click.stop>
        <div
          v-for="locale in availableLocales"
          :key="locale.code"
          class="language-option"
          :class="{ active: locale.code === currentLocale }"
          @click="selectLanguage(locale.code)"
        >
          <span class="language-flag">{{ locale.flag }}</span>
          <span class="language-name">{{ locale.name }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useEnhancedI18n } from '../composables/useI18n'
import { changeLanguage } from '../i18n'

const { locale } = useEnhancedI18n()

const isOpen = ref(false)
const dropdown = ref(null)

const currentLocale = computed(() => {
  const current = locale.value
  console.log('Current locale:', current)
  return current
})

const availableLocales = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'ar', name: 'العربية', flag: '🇸🇦' },
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
  { code: 'zh', name: '中文', flag: '🇨🇳' },
]

const getLanguageFlag = (code) => {
  const lang = availableLocales.find((l) => l.code === code)
  return lang ? lang.flag : '🇺🇸'
}

const getLanguageName = (code) => {
  const lang = availableLocales.find((l) => l.code === code)
  return lang ? lang.name : 'English'
}

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const selectLanguage = (code) => {
  console.log('Changing language to:', code)
  changeLanguage(code)
  isOpen.value = false
}

// Watch for locale changes
watch(currentLocale, (newLocale) => {
  console.log('Locale changed to:', newLocale)
})

// Close dropdown when clicking outside
const handleClickOutside = (event) => {
  if (dropdown.value && !dropdown.value.contains(event.target)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  console.log('LanguageSwitcher mounted, current locale:', currentLocale.value)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.language-switcher {
  position: relative;
}

.language-dropdown {
  position: relative;
  cursor: pointer;
  user-select: none;
}

.language-current {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  transition: all 0.2s;
}

.language-current:hover {
  background: #f1f5f9;
  border-color: #d1d5db;
}

.language-flag {
  font-size: 16px;
}

.language-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.language-current i {
  font-size: 12px;
  color: #6b7280;
  transition: transform 0.2s;
}

.language-current i.rotated {
  transform: rotate(180deg);
}

.language-options {
  position: absolute;
  top: 100%;
  right: 0;
  min-width: 160px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  margin-top: 4px;
  overflow: hidden;
}

.language-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.language-option:hover {
  background: #f8fafc;
}

.language-option.active {
  background: #eff6ff;
  color: #2563eb;
}

.language-option.active .language-name {
  color: #2563eb;
  font-weight: 600;
}

/* RTL Support */
[dir='rtl'] .language-options {
  right: auto;
  left: 0;
}

[dir='rtl'] .language-current {
  flex-direction: row-reverse;
}

[dir='rtl'] .language-option {
  flex-direction: row-reverse;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .language-name {
    display: none;
  }

  .language-current {
    padding: 8px;
  }

  .language-options {
    min-width: 120px;
  }

  .language-option {
    padding: 10px 12px;
  }
}
</style>
