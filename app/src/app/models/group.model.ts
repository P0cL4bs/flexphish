export interface GroupTarget {
  id: number
  first_name: string
  last_name: string
  email: string
  position?: string
  created_at?: string
  updated_at?: string
}

export interface Group {
  id: number
  user_id?: number
  name: string
  is_global: boolean
  targets?: GroupTarget[]
  created_at?: string
  updated_at?: string
}

export interface GroupTargetPayload {
  first_name: string
  last_name: string
  email: string
  position?: string
}

export interface CreateGroupRequest {
  name: string
  is_global: boolean
  targets: GroupTargetPayload[]
}

export interface UpdateGroupRequest {
  name: string
  is_global: boolean
}
