import { ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { UserSubscription } from '@/types'
const activeSubscriptions = ref<UserSubscription[]>([])
const currentBalances = ref<Array<{
  model: string
  tokens_purchased: number
  tokens_consumed: number
  balance: number
  usage_percent: number
  status: string
}>>([])

vi.mock('@/stores', () => ({
  useSubscriptionStore: () => ({
    get activeSubscriptions() {
      return activeSubscriptions.value
    },
    get hasActiveSubscriptions() {
      return activeSubscriptions.value.length > 0
    },
    fetchActiveSubscriptions: vi.fn().mockResolvedValue(activeSubscriptions.value)
  })
}))

vi.mock('@/api/usage', () => ({
  default: {
    getModelBalances: vi.fn(() => Promise.resolve(currentBalances.value))
  }
}))

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackWarn: false,
  missingWarn: false,
  messages: {
    en: {
      subscriptionProgress: {
        title: () => 'My Subscriptions',
        activeCount: () => '1 active subscription(s)',
        unlimited: () => 'Unlimited',
        tokenBalance: () => 'Token balance',
        viewDetails: () => 'View details',
        viewAll: () => 'View all subscriptions'
      }
    }
  }
})

const subscription: UserSubscription = {
  id: 1,
  user_id: 1,
  group_id: 1,
  status: 'active',
  starts_at: '2026-01-01T00:00:00Z',
  daily_usage_usd: 0,
  weekly_usage_usd: 0,
  monthly_usage_usd: 0,
  daily_window_start: null,
  weekly_window_start: null,
  monthly_window_start: null,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  expires_at: null,
  group: {
    id: 1,
    name: 'Legacy unlimited',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null
  }
}

async function mountComponent() {
  vi.resetModules()
  const { default: SubscriptionProgressMini } = await import('../SubscriptionProgressMini.vue')
  activeSubscriptions.value = [subscription]
  const wrapper = mount(SubscriptionProgressMini, {
    global: {
      plugins: [i18n, createPinia()],
      stubs: { 'router-link': true }
    }
  })
  await flushPromises()
  await wrapper.get('button').trigger('click')
  await flushPromises()
  return wrapper
}

describe('SubscriptionProgressMini', () => {
  beforeEach(() => {
    currentBalances.value = []
  })

  it('shows token balance instead of Unlimited when active balances exist', async () => {
    currentBalances.value = [{
      model: 'gpt-5.5',
      tokens_purchased: 10_000_000,
      tokens_consumed: 0,
      balance: 10_000_000,
      usage_percent: 0,
      status: 'active'
    }]

    const wrapper = await mountComponent()

    expect(wrapper.text()).toContain('Token balance')
    expect(wrapper.text()).not.toContain('∞')
    expect(wrapper.text()).not.toContain('Unlimited')
  })

  it('shows Unlimited when no active token balances exist', async () => {
    const wrapper = await mountComponent()

    expect(wrapper.text()).toContain('Unlimited')
  })
})
