<template>
  <section class="card overflow-hidden border border-blue-100 bg-white shadow-sm dark:border-dark-600 dark:bg-dark-800">
    <div class="border-b border-gray-100 bg-gradient-to-r from-blue-50 to-white px-6 py-5 dark:border-dark-700 dark:from-blue-950/30 dark:to-dark-800">
      <div class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p class="text-xs font-semibold uppercase tracking-wider text-blue-600 dark:text-blue-400">{{ t('dashboard.catalog.eyebrow') }}</p>
          <h2 class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ t('dashboard.catalog.title') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('dashboard.catalog.description') }}</p>
        </div>
        <RouterLink to="/purchase" class="text-sm font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300">
          {{ t('dashboard.catalog.browsePlans') }} <span aria-hidden="true">→</span>
        </RouterLink>
      </div>
    </div>

    <div v-if="modelBalances.length === 0" class="px-6 py-12 text-center">
      <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-blue-50 text-blue-600 dark:bg-blue-950/40 dark:text-blue-400">
        <Icon name="cube" size="md" />
      </div>
      <h3 class="mt-4 text-base font-semibold text-gray-900 dark:text-white">{{ t('dashboard.catalog.emptyTitle') }}</h3>
      <p class="mx-auto mt-1 max-w-md text-sm text-gray-500 dark:text-gray-400">{{ t('dashboard.catalog.emptyDescription') }}</p>
      <RouterLink to="/purchase" class="mt-5 inline-flex rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-blue-700">
        {{ t('dashboard.catalog.buyPlan') }}
      </RouterLink>
    </div>

    <div v-else class="grid gap-4 p-6 md:grid-cols-2 xl:grid-cols-3">
      <article v-for="balance in modelBalances" :key="balance.model" class="flex flex-col rounded-xl border border-gray-200 p-5 dark:border-dark-600">
        <div class="flex items-start justify-between gap-3">
          <h3 class="break-all font-mono text-sm font-semibold text-gray-900 dark:text-white">{{ balance.model }}</h3>
          <span :class="statusClass(balance.status)" class="shrink-0 rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-wide">
            {{ statusLabel(balance.status) }}
          </span>
        </div>
        <div class="mt-5 flex items-end justify-between gap-3">
          <div>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ formatTokens(balance.balance) }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('dashboard.catalog.tokensRemaining') }}</p>
          </div>
          <p class="text-right text-xs text-gray-500 dark:text-gray-400">
            {{ t('dashboard.catalog.purchased') }} {{ formatTokens(balance.tokens_purchased) }}<br />
            {{ t('dashboard.catalog.consumed') }} {{ formatTokens(balance.tokens_consumed) }}
          </p>
        </div>
        <div class="mt-4">
          <div class="mb-1 flex justify-between text-xs text-gray-500 dark:text-gray-400">
            <span>{{ t('dashboard.catalog.usage') }}</span><span>{{ usagePercent(balance.usage_percent) }}%</span>
          </div>
          <div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-600">
            <div class="h-full rounded-full bg-blue-600 transition-all" :style="{ width: `${usagePercent(balance.usage_percent)}%` }" />
          </div>
        </div>
        <div class="mt-5 flex flex-wrap gap-2">
          <RouterLink :to="purchaseLink(balance.model)" :class="balance.status === 'active' ? 'border border-gray-300 text-gray-700 hover:bg-gray-50 dark:border-dark-500 dark:text-gray-200 dark:hover:bg-dark-700' : 'bg-blue-600 text-white hover:bg-blue-700'" class="inline-flex flex-1 items-center justify-center rounded-lg px-3 py-2 text-xs font-semibold transition">
            {{ balance.status === 'active' ? t('dashboard.catalog.topUp') : t('dashboard.catalog.buyPlan') }}
          </RouterLink>
          <RouterLink v-if="balance.status === 'active'" to="/keys" class="inline-flex flex-1 items-center justify-center rounded-lg border border-blue-200 px-3 py-2 text-xs font-semibold text-blue-600 transition hover:bg-blue-50 dark:border-blue-900 dark:text-blue-400 dark:hover:bg-blue-950/30">
            {{ t('dashboard.catalog.useModel') }}
          </RouterLink>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ModelBalance } from '@/api/usage'

defineProps<{ modelBalances: ModelBalance[] }>()
const { t } = useI18n()

const formatTokens = (value: number): string => {
  const tokens = Math.max(0, Math.floor(value || 0))
  if (tokens >= 1_000_000) return `${Math.floor(tokens / 100_000) / 10}M`
  if (tokens >= 1_000) return `${Math.floor(tokens / 100) / 10}K`
  return tokens.toLocaleString()
}
const usagePercent = (value: number): number => Math.min(100, Math.max(0, Math.floor(value || 0)))
const purchaseLink = (model: string) => ({ path: '/purchase', query: { tab: 'subscription', model } })
const statusLabel = (status: ModelBalance['status']) => t(`dashboard.catalog.status.${status === 'not_purchased' ? 'notPurchased' : status}`)
const statusClass = (status: ModelBalance['status']) => status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : status === 'blocked' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-gray-100 text-gray-600 dark:bg-dark-600 dark:text-gray-300'
</script>
