// Brand profile and style profile type definitions

export interface BrandProfileVO {
  id: string
  name: string
  brand_tone: string
  forbidden_words: string[]
  preferred_words: string[]
  cta_rules: Record<string, unknown>
  cover_style_rules: Record<string, unknown>
  image_style_rules: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateBrandProfileRequest {
  name: string
  brand_tone?: string
  forbidden_words?: string[]
  preferred_words?: string[]
  cta_rules?: Record<string, unknown>
  cover_style_rules?: Record<string, unknown>
  image_style_rules?: Record<string, unknown>
}

export interface UpdateBrandProfileRequest {
  name?: string
  brand_tone?: string
  forbidden_words?: string[]
  preferred_words?: string[]
  cta_rules?: Record<string, unknown>
  cover_style_rules?: Record<string, unknown>
  image_style_rules?: Record<string, unknown>
}

export interface StyleProfileVO {
  id: string
  name: string
  applicable_scene: string
  opening_template: string
  structure_template: Record<string, unknown>
  closing_template: string
  sentence_preference: Record<string, unknown>
  title_preference: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface CreateStyleProfileRequest {
  name: string
  applicable_scene?: string
  opening_template?: string
  structure_template?: Record<string, unknown>
  closing_template?: string
  sentence_preference?: Record<string, unknown>
  title_preference?: Record<string, unknown>
}

export interface UpdateStyleProfileRequest {
  name?: string
  applicable_scene?: string
  opening_template?: string
  structure_template?: Record<string, unknown>
  closing_template?: string
  sentence_preference?: Record<string, unknown>
  title_preference?: Record<string, unknown>
}
