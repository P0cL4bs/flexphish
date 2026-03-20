export interface SMTPProfile {
  id: number
  user_id?: number
  is_global: boolean
  name: string
  host: string
  port: number
  username: string
  from_name?: string
  from_email?: string
  is_active: boolean
  created_at?: string
  updated_at?: string
}

export interface SMTPProfilePayload {
  name: string
  is_global: boolean
  host: string
  port: number
  username: string
  password: string
  from_name?: string
  from_email?: string
  is_active: boolean
}
