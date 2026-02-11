import { fetchApi } from '@/utils/request'

export interface UserAccount {
  id: string
  username: string
  email?: string
  role: string
  display_name: string
  avatar_url?: string
  is_active: boolean
  created_at: string
  updated_at: string
  last_login_at?: string
}

export interface UpdateMyProfileRequest {
  display_name: string
  avatar_url: string
}

export interface UpdateMyPasswordRequest {
  current_password: string
  new_password: string
}

export interface ChannelIdentity {
  id: string
  user_id?: string
  channel: string
  channel_subject_id: string
  display_name?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface ListMyIdentitiesResponse {
  user_id: string
  items: ChannelIdentity[]
}

export interface IssueBindCodeRequest {
  platform?: string
  ttl_seconds?: number
}

export interface IssueBindCodeResponse {
  token: string
  platform?: string
  expires_at: string
}

export async function getMyAccount(): Promise<UserAccount> {
  return fetchApi<UserAccount>('/users/me')
}

export async function updateMyProfile(data: UpdateMyProfileRequest): Promise<UserAccount> {
  return fetchApi<UserAccount>('/users/me', {
    method: 'PUT',
    body: data,
  })
}

export async function updateMyPassword(data: UpdateMyPasswordRequest): Promise<void> {
  return fetchApi<void>('/users/me/password', {
    method: 'PUT',
    body: data,
  })
}

export async function listMyIdentities(): Promise<ListMyIdentitiesResponse> {
  return fetchApi<ListMyIdentitiesResponse>('/users/me/identities')
}

export async function issueMyBindCode(data: IssueBindCodeRequest): Promise<IssueBindCodeResponse> {
  return fetchApi<IssueBindCodeResponse>('/users/me/bind_codes', {
    method: 'POST',
    body: data,
  })
}
