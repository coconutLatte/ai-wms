// Auth API module — login, refresh, and profile endpoints.

import client from './client'
import type { LoginRequest, LoginResponse, RefreshRequest, UserProfile } from './types'

export async function login(input: LoginRequest): Promise<LoginResponse> {
  const { data } = await client.post<LoginResponse>('/auth/login', input)
  return data
}

export async function refresh(input: RefreshRequest): Promise<LoginResponse> {
  const { data } = await client.post<LoginResponse>('/auth/refresh', input)
  return data
}

export async function getProfile(): Promise<UserProfile> {
  const { data } = await client.get<UserProfile>('/auth/me')
  return data
}
