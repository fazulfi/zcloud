import { apiClient } from './client'

export async function deleteAccount(): Promise<{ code: string; message: string }> {
  const { data } = await apiClient.post('/user/account/delete')
  return data
}
