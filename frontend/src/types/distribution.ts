// Distribution package types

export interface DistributionVO {
  public_id: string
  draft_id: number
  community_copy: string
  moments_copy: string
  summary_card: string
  comment_guide: string
  next_topic_suggestion: string
  created_at: string
  updated_at: string
}

export interface GenerateDistributionRequest {
  draft_public_id: string
}
