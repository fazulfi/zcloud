import { apiClient } from './client'

export interface SupportContactPayload {
  name: string
  email: string
  subject: string
  message: string
}

export async function sendSupportContact(payload: SupportContactPayload): Promise<void> {
  await apiClient.post('/support/contact', payload)
}
