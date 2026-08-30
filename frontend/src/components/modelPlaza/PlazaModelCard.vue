<template>
  <article
    class="group relative flex min-w-0 flex-col overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md dark:border-dark-600 dark:bg-dark-800"
    :style="accentStyle"
  >
    <div class="h-1 w-full" :style="{ backgroundColor: platformAccentColor(platform ?? '') }" />
    <div class="flex flex-1 flex-col p-5">
      <header>
        <div class="flex items-start justify-between gap-3">
          <button type="button" class="min-w-0 break-all text-left font-mono text-sm font-semibold text-primary-600 hover:underline dark:text-primary-400" :title="t('modelPlaza.table.buy')" @click="goToModel(model.name)">
            {{ model.name }}
          </button>
          <span v-if="period" class="shrink-0 rounded-md bg-gray-100 px-1.5 py-0.5 font-mono text-[10px] font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-300" :title="timePricingRowHint(model)">
            <span v-if="model.time_pricing?.weekdays_only" class="mr-1 font-sans">{{ t('modelPlaza.table.timePricingWeekdays') }}</span>{{ formatTimeWindow(period) }}
          </span>
        </div>
        <div class="mt-2 flex flex-wrap gap-1.5">
          <span v-if="platform && model.platform !== platform" :class="['rounded-md px-1.5 py-0.5 text-[10px] font-medium', platformBadgeLightClass(model.platform)]">{{ platformLabel(model.platform) }}</span>
          <span v-if="billingMode(model) !== BILLING_MODE_TOKEN" class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-300">{{ billingModeLabel(model) }}</span>
          <span v-if="model.long_context_basis === 'marginal'" class="rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-700 dark:text-dark-300" :title="t('modelPlaza.table.tierHintMarginal')">{{ t('modelPlaza.table.marginalBadge') }}</span>
        </div>
      </header>

      <div class="mt-5 space-y-3 text-xs tabular-nums">
        <section class="rounded-lg border border-gray-100 bg-gray-50/70 p-3 dark:border-dark-700 dark:bg-dark-900/30">
          <div class="mb-2 flex items-center justify-between gap-2">
            <h3 class="font-semibold uppercase tracking-wider text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.officialPrice') }}</h3>
            <span class="text-[10px] text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.unitPerMillion') }}</span>
          </div>
          <div class="grid grid-cols-3 gap-2 text-gray-500 dark:text-dark-400">
            <div>
              <p class="mb-1 text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.input') }}</p>
              <template v-if="officialIntervals(model).length">
                <p v-for="iv in officialIntervals(model)" :key="iv.min_tokens" class="leading-5">
                  <span class="mr-1 font-sans text-gray-400" :title="t('modelPlaza.table.tierHint')">{{ tierLabel(iv) }}</span>{{ official(iv.input_price) }}
                </p>
              </template>
              <p v-else>{{ official(model.official_pricing?.input_price) }}</p>
            </div>
            <div>
              <p class="mb-1 text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.output') }}</p>
              <template v-if="officialIntervals(model).length">
                <p v-for="iv in officialIntervals(model)" :key="iv.min_tokens" class="leading-5">{{ official(iv.output_price) }}</p>
              </template>
              <p v-else>{{ official(model.official_pricing?.output_price) }}</p>
            </div>
            <div>
              <p class="mb-1 text-gray-400 dark:text-dark-500">{{ t('modelPlaza.table.cache') }}</p>
              <template v-if="hasTierCachePricing(officialIntervals(model))">
                <p v-for="iv in officialIntervals(model)" :key="iv.min_tokens" class="leading-5">
                  <template v-if="iv.cache_write_price != null || iv.cache_read_price != null">{{ t('modelPlaza.table.cacheWriteShort') }} {{ official(iv.cache_write_price) }} {{ t('modelPlaza.table.cacheReadShort') }} {{ official(iv.cache_read_price) }}</template>
                  <span v-else>-</span>
                </p>
              </template>
              <div v-else-if="model.official_pricing && hasOfficialCache(model.official_pricing)" class="space-y-0.5">
                <p>{{ t('modelPlaza.table.cacheWrite') }} {{ official(model.official_pricing.cache_write_price) }}<template v-if="model.official_pricing.cache_write_1h_price != null"> <span class="font-sans text-gray-400">{{ t('modelPlaza.table.perHour') }}</span> {{ official(model.official_pricing.cache_write_1h_price) }}</template></p>
                <p>{{ t('modelPlaza.table.cacheRead') }} {{ official(model.official_pricing.cache_read_price) }}</p>
              </div>
              <span v-else>-</span>
            </div>
          </div>
        </section>
      </div>

      <footer class="mt-auto flex items-center justify-between gap-3 border-t border-gray-100 pt-4 dark:border-dark-700">
        <div class="font-mono text-xs">
          <span v-if="period" class="font-bold text-primary-600 dark:text-primary-400" :title="t('modelPlaza.table.timePricingRateHint', { rate: effectiveRate, multiplier: period.multiplier })">{{ periodRate(period) }}x</span>
          <span v-else-if="usesIndependentImageRate(model)" class="font-bold text-gray-700 dark:text-gray-300">{{ requestRate(model) }}x</span>
          <template v-else-if="hasCustomRate">
            <span class="mr-1 text-gray-400 line-through dark:text-dark-500">{{ rateMultiplier }}x</span>
            <span class="font-bold text-primary-600 dark:text-primary-400">{{ effectiveRate }}x</span>
          </template>
          <span v-else class="font-bold text-gray-700 dark:text-gray-300">{{ effectiveRate }}x</span>
          <span class="ml-1 text-gray-400">{{ t('modelPlaza.table.rate') }}</span>
        </div>
        <button type="button" :class="['inline-flex items-center rounded-lg px-3 py-1.5 text-xs font-semibold text-white shadow-sm transition hover:opacity-90 active:scale-[.97]', platformButtonClass(platform ?? '')]" @click="goToModel(model.name)">{{ t('modelPlaza.table.buy') }}</button>
      </footer>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { formatScaled } from '@/utils/pricing'
import { platformAccentColor, platformBadgeLightClass, platformButtonClass, platformLabel } from '@/utils/platformColors'
import { BILLING_MODE_TOKEN, BILLING_MODE_IMAGE, type BillingMode } from '@/constants/channel'
import type { PlazaModel, PlazaTimePricingPeriod } from '@/api/modelPlaza'
import type { UserPricingInterval } from '@/api/channels'
const props = defineProps<{ model: PlazaModel; platform?: string; rateMultiplier: number; userRateMultiplier?: number | null; imageRateIndependent?: boolean; imageRateMultiplier?: number | null; period?: PlazaTimePricingPeriod | null; peakWindow?: string; peakRateMultiplier?: number | null }>()
const { t } = useI18n(); const router = useRouter(); const PER_MILLION = 1_000_000; const MIN_DECIMALS = 2
const accentStyle = computed(() => ({ '--plaza-accent': platformAccentColor(props.platform ?? '') }))
const effectiveRate = computed(() => props.userRateMultiplier ?? props.rateMultiplier)
const hasCustomRate = computed(() => props.userRateMultiplier != null && props.userRateMultiplier !== props.rateMultiplier)
function goToModel(name: string): void { void router.push(`/model-plaza/model/${encodeURIComponent(name)}`) }
function billingMode(m: PlazaModel): BillingMode { return (m.pricing?.billing_mode || BILLING_MODE_TOKEN) as BillingMode }
function billingModeLabel(m: PlazaModel): string { return billingMode(m) === BILLING_MODE_IMAGE ? t('modelPlaza.table.perImage') : t('modelPlaza.table.perRequest') }
function periodRate(p: PlazaTimePricingPeriod): number { return Math.round(effectiveRate.value * p.multiplier * 1000) / 1000 }
function usesIndependentImageRate(m: PlazaModel): boolean { return billingMode(m) === BILLING_MODE_IMAGE && props.imageRateIndependent === true }
function requestRate(m: PlazaModel): number { return usesIndependentImageRate(m) ? (props.imageRateMultiplier ?? 1) : effectiveRate.value }
function official(value: number | null | undefined): string { return value == null ? '-' : formatScaled(value, PER_MILLION, MIN_DECIMALS) }
function hasOfficialCache(o: NonNullable<PlazaModel['official_pricing']>): boolean { return o.cache_write_price != null || o.cache_read_price != null || o.cache_write_1h_price != null }
function sortByContext(i: UserPricingInterval[]): UserPricingInterval[] { return [...i].sort((a, b) => a.min_tokens - b.min_tokens) }
function officialIntervals(m: PlazaModel): UserPricingInterval[] { return sortByContext(m.official_pricing?.intervals ?? []) }
function hasTierCachePricing(i: UserPricingInterval[]): boolean { return i.some((iv) => iv.cache_write_price != null || iv.cache_read_price != null) }
function tierLabel(iv: UserPricingInterval): string { if (iv.tier_label) return iv.tier_label; const max = iv.max_tokens; return max == null ? `>${formatTokenCount(iv.min_tokens)}` : `≤${formatTokenCount(max)}` }
function formatTokenCount(n: number): string { if (n >= 1_000_000) return `${trimZero(n / 1_000_000)}M`; if (n >= 1_000) return `${trimZero(n / 1_000)}K`; return String(n) }
function trimZero(n: number): string { return String(Math.round(n * 100) / 100) }
function formatTimeWindow(p: PlazaTimePricingPeriod): string { const clock = (v: string) => v.replace(/^(\d{2}:\d{2}):00$/, '$1'); return `${clock(p.start_time)}–${clock(p.end_time)}` }
function timePricingRowHint(m: PlazaModel): string { const key = m.time_pricing?.weekdays_only ? 'modelPlaza.table.timePricingRowHintWeekdays' : 'modelPlaza.table.timePricingRowHint'; let hint = t(key, { timezone: m.time_pricing?.timezone }); if (props.peakWindow) hint += t('modelPlaza.table.timePricingRowHintPeak', { window: props.peakWindow, multiplier: props.peakRateMultiplier ?? 1 }); return hint }
</script>
