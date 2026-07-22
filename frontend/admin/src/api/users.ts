// User API module — CRUD operations for user management.
// Calls backend endpoints under /api/v1/users.

import client from './client'
import type { User, ListResponse, CreateUserRequest, UpdateUserRequest, UpdateUserStatusRequest } from './types'

export async function listUsers(params: {
  page: number
  page_size: number
  status?: string
}): Promise<ListResponse<User>> {
  const { data } = await client.get<ListResponse<User>>('/users', { params })
  return data
}

export async function createUser(input: CreateUserRequest): Promise<User> {
  const { data } = await client.post<User>('/users', input)
  return data
}

export async function getUser(id: string): Promise<User> {
  const { data } = await client.get<User>(`/users/${id}`)
  return data
}

export async function updateUser(id: string, input: UpdateUserRequest): Promise<User> {
  const { data } = await client.put<User>(`/users/${id}`, input)
  return data
}

export async function updateUserStatus(id: string, input: UpdateUserStatusRequest): Promise<User> {
  const { data } = await client.put<User>(`/users/${id}/status`, input)
  return data
}
