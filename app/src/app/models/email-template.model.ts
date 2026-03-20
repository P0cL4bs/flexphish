export interface EmailTemplate {
  id: number
  user_id?: number
  is_global: boolean
  name: string
  subject: string
  body: string
  created_at?: string
  updated_at?: string
}

export interface EmailTemplatePayload {
  name: string
  is_global: boolean
  subject: string
  body: string
}
