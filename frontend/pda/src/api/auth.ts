// Auth API — login, profile, and token management endpoints.
// Mirrors the admin frontend pattern for auth operations.

import client from './client'
import type { LoginRequest, LoginResponse, UserProfile } from './types'

export async function login(input: LoginRequest): Promise<LoginResponse> {
  const { data } = await client.post<LoginResponse>('/auth/login', input)
  return data
}

export async function getProfile(): Promise<UserProfile> {
  const { data } = await client.get<UserProfile>('/auth/me')
  return data
}
