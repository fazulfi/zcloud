import { describe, expect, it } from 'vitest'
import { currencySymbol, formatPaymentAmount } from '../currency'

describe('formatPaymentAmount', () => {
  it('formats IDR as whole rupiah', () => {
    expect(formatPaymentAmount(160000.75, 'IDR', 'en-US')).not.toContain('.75')
    expect(formatPaymentAmount(160000, 'IDR', 'en-US')).toContain('160,000')
  })

  it('uses the currency default fraction digits', () => {
    expect(formatPaymentAmount(100, 'JPY', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'KRW', 'en-US')).not.toContain('.00')
    expect(formatPaymentAmount(100, 'HKD', 'en-US')).toContain('.00')
  })
})

describe('currencySymbol', () => {
  it('maps common payment currencies and falls back safely', () => {
    expect(currencySymbol('USD')).toBe('$')
    expect(currencySymbol('cny')).toBe('¥')
    expect(currencySymbol('EUR')).toBe('€')
    expect(currencySymbol('IDR')).toBe('Rp')
    expect(currencySymbol('')).toBe('¥')
    expect(currencySymbol('XYZ')).toBe('XYZ')
  })
})
