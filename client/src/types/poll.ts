export interface PollOption {
  id: string;
  poll_id: string;
  option_text: string;
  vote_count: number;
  position: number;
  created_at?: string;
  percentage?: number;
}

export interface Poll {
  id: string;
  question: string;
  description?: string | null;
  is_active: boolean;
  expires_at?: string | null;
  created_at: string;
  total_votes: number;
  options?: PollOption[];
}

export interface CreatePollRequest {
  question: string;
  description?: string;
  options: string[];
  expires_at?: string;
}

export interface VoteRequest {
  option_id: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  error?: string;
}

export interface PaginatedResponse<T> {
  polls: T[];
  total: number;
  limit: number;
  offset: number;
}
