export type SMTPSecurityMode = 'starttls' | 'implicit_tls' | 'none'

export interface SMTPProfile {
  id: number
  user_id?: number
  is_global: boolean
  name: string
  host: string
  port: number
  security_mode: SMTPSecurityMode
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
  security_mode: SMTPSecurityMode
  username: string
  password: string
  from_name?: string
  from_email?: string
  is_active: boolean
}

export interface SMTPTestPayload {
  smtp_profile_id?: number
  name: string
  host: string
  port: number
  security_mode: SMTPSecurityMode
  use_authentication?: boolean
  username: string
  password: string
  from_name?: string
  from_email?: string
  test_email: string
}
