import { computed, ref } from 'vue'
import usageAPI, { type ModelBalance } from '@/api/usage'

const CACHE_TTL_MS = 60_000

let cachedBalances: ModelBalance[] = []
let lastFetchedAt: number | null = null
let inFlightPromise: Promise<ModelBalance[]> | null = null

export function useModelBalances() {
  const modelBalances = ref<ModelBalance[]>(cachedBalances)
  const loading = ref(false)

  const activeBalances = computed(() =>
    modelBalances.value.filter((balance) => balance.status === 'active' && balance.balance > 0)
  )
  const totalBalance = computed(() =>
    activeBalances.value.reduce((total, balance) => total + Math.max(0, balance.balance), 0)
  )
  const hasTokenBalances = computed(() => activeBalances.value.length > 0)
  const balanceModelNames = computed(() => activeBalances.value.map((balance) => balance.model).join(', '))

  async function loadModelBalances(force = false): Promise<ModelBalance[]> {
    const now = Date.now()
    if (!force && lastFetchedAt !== null && now - lastFetchedAt < CACHE_TTL_MS) {
      modelBalances.value = cachedBalances
      return cachedBalances
    }

    if (inFlightPromise && !force) {
      const balances = await inFlightPromise
      modelBalances.value = balances
      return balances
    }

    const requestPromise = usageAPI.getModelBalances()
      .then((balances) => {
        cachedBalances = balances
        lastFetchedAt = Date.now()
        return balances
      })
      .catch((error) => {
        console.error('Failed to load model balances:', error)
        return []
      })
      .finally(() => {
        if (inFlightPromise === requestPromise) {
          inFlightPromise = null
        }
      })

    inFlightPromise = requestPromise
    loading.value = true
    try {
      const balances = await requestPromise
      modelBalances.value = balances
      return balances
    } finally {
      loading.value = false
    }
  }

  return {
    modelBalances,
    loading,
    activeBalances,
    totalBalance,
    hasTokenBalances,
    balanceModelNames,
    loadModelBalances
  }
}
