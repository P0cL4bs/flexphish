export interface EmailTemplate {
  id: number
  user_id?: number
  is_global: boolean
  name: string
  category?: string
  track_opens: boolean
  subject: string
  body: string
  created_at?: string
  updated_at?: string
}

export interface EmailTemplateAttachment {
  id: number
  email_template_id: number
  filename: string
  mime_type: string
  size: number
  created_at?: string
  updated_at?: string
}

export interface EmailTemplatePayload {
  name: string
  category?: string
  is_global: boolean
  track_opens?: boolean
  subject: string
  body: string
}
