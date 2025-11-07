import type {
  Poll,
  CreatePollRequest,
  ApiResponse,
  PaginatedResponse,
} from "@/types/poll";

const API_BASE =
  import.meta.env.VITE_API_BASE || "http://localhost:6767/api/v1";

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.text();
    throw new Error(`API Error: ${response.status} - ${error}`);
  }
  const data: ApiResponse<T> = await response.json();
  return data.data;
}

export async function fetchPolls(
  limit = 20,
  offset = 0,
  activeOnly = true
): Promise<Poll[]> {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
    active: activeOnly.toString(),
  });

  const response = await fetch(`${API_BASE}/polls?${params}`);
  const paginatedData = await handleResponse<PaginatedResponse<Poll>>(response);
  return paginatedData.polls;
}

export async function fetchPoll(id: string): Promise<Poll> {
  const response = await fetch(`${API_BASE}/polls/${id}`);
  return handleResponse<Poll>(response);
}

export async function createPoll(poll: CreatePollRequest): Promise<Poll> {
  const response = await fetch(`${API_BASE}/polls`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(poll),
  });
  return handleResponse<Poll>(response);
}

export async function castVote(
  pollId: string,
  optionId: string
): Promise<void> {
  const response = await fetch(`${API_BASE}/polls/${pollId}/vote`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ option_id: optionId }),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`Vote failed: ${response.status} - ${error}`);
  }
}
