<template>
  <div class="space-y-6 p-6">
    <div class="flex flex-wrap items-end justify-between gap-4">
      <div><h1 class="text-2xl font-semibold text-gray-900 dark:text-white">Model margins</h1><p class="text-sm text-gray-500">Internal operational evidence</p></div>
      <div class="flex gap-2"><input v-model="start" type="date" class="input" /><input v-model="end" type="date" class="input" /><button class="btn-primary" @click="load">Refresh</button></div>
    </div>
    <section class="overflow-hidden rounded-lg bg-white shadow dark:bg-gray-800">
      <table class="min-w-full text-sm"><thead class="bg-gray-50 text-left dark:bg-gray-700"><tr><th class="px-4 py-3">Model</th><th class="px-4 py-3 text-right">Display</th><th class="px-4 py-3 text-right">Cost</th><th class="px-4 py-3 text-right">Margin</th><th class="px-4 py-3 text-right">Margin %</th><th class="px-4 py-3">Suppliers</th></tr></thead><tbody><tr v-for="item in models" :key="item.model" class="border-t dark:border-gray-700"><td class="px-4 py-3 font-medium">{{ item.model }}</td><td class="px-4 py-3 text-right">{{ money(item.display_total) }}</td><td class="px-4 py-3 text-right">{{ money(item.cost_total) }}</td><td class="px-4 py-3 text-right">{{ money(item.margin) }}</td><td class="px-4 py-3 text-right">{{ item.margin_percent.toFixed(2) }}%</td><td class="px-4 py-3">{{ item.supplier_breakdown.map(s => `${s.supplier_code || 'unattributed'}: ${money(s.cost_total)}`).join(', ') }}</td></tr><tr v-if="!models.length"><td colspan="6" class="px-4 py-8 text-center text-gray-500">No data</td></tr></tbody></table>
    </section>
    <section class="rounded-lg bg-white p-4 shadow dark:bg-gray-800"><div class="mb-3 flex gap-2"><input v-model.number="userId" type="number" min="1" placeholder="User ID" class="input" /><button class="btn-secondary" @click="loadBalances">Load model balances</button></div><table class="min-w-full text-sm"><thead><tr><th class="px-2 py-2 text-left">Model</th><th class="px-2 py-2 text-right">Balance</th><th class="px-2 py-2 text-right">Usage</th><th class="px-2 py-2 text-left">Status</th></tr></thead><tbody><tr v-for="item in balances" :key="item.model_id"><td class="px-2 py-2">{{ item.canonical_name }}</td><td class="px-2 py-2 text-right">{{ item.balance.toLocaleString() }}</td><td class="px-2 py-2 text-right">{{ item.usage_percent.toFixed(2) }}%</td><td class="px-2 py-2">{{ item.status }}</td></tr></tbody></table></section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getModelMargins, getUserModelBalances, type ModelBalance, type ModelMargin } from '@/api/admin/dashboard'
const models = ref<ModelMargin[]>([]); const balances = ref<ModelBalance[]>([]); const userId = ref<number | undefined>(); const start = ref(''); const end = ref('')
const money = (value: number) => value.toFixed(6)
async function load() { const data = await getModelMargins({ start_date: start.value || undefined, end_date: end.value || undefined }); models.value = data.models }
async function loadBalances() { if (userId.value) balances.value = (await getUserModelBalances(userId.value)).balances }
onMounted(load)
</script>
