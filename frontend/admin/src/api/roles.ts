// Role API module — CRUD operations for role management.
// Calls backend endpoints under /api/v1/roles.

import client from './client'
import type { Role, ListResponse, CreateRoleRequest, UpdateRoleRequest } from './types'

export async function listRoles(): Promise<ListResponse<Role>> {
  const { data } = await client.get<ListResponse<Role>>('/roles')
  return data
}

export async function createRole(input: CreateRoleRequest): Promise<Role> {
  const { data } = await client.post<Role>('/roles', input)
  return data
}

export async function getRole(id: string): Promise<Role> {
  const { data } = await client.get<Role>(`/roles/${id}`)
  return data
}

export async function updateRole(id: string, input: UpdateRoleRequest): Promise<Role> {
  const { data } = await client.put<Role>(`/roles/${id}`, input)
  return data
}

export async function deleteRole(id: string): Promise<void> {
  await client.delete(`/roles/${id}`)
}
