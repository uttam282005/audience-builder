import {
  PreviewRequest,
  PreviewResponse,
  HealthResponse,
  ApiErrorResponse,
} from '../types/audience';

export const API_BASE_URL =
  import.meta.env.VITE_API_URL || 'http://localhost:8080';

export class ApiError extends Error {
  public status: number;
  public details?: { field: string; issue: string }[];
  public code: string;

  constructor(
    status: number,
    code: string,
    message: string,
    details?: { field: string; issue: string }[]
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

export async function checkHealth(): Promise<HealthResponse> {
  const response = await fetch(`${API_BASE_URL}/health`, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    throw new ApiError(
      response.status,
      'health_check_failed',
      `Health check failed with HTTP ${response.status}`
    );
  }

  return response.json();
}

export async function previewAudience(
  request: PreviewRequest
): Promise<PreviewResponse> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}/v1/audiences/preview`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify(request),
    });
  } catch (err) {
    throw new ApiError(
      0,
      'network_error',
      err instanceof Error
        ? `Network connection error: ${err.message}. Is the backend running on ${API_BASE_URL}?`
        : 'Network connection failed'
    );
  }

  if (!response.ok) {
    let errorData: ApiErrorResponse;
    try {
      errorData = await response.json();
    } catch {
      throw new ApiError(
        response.status,
        'http_error',
        `Server responded with HTTP ${response.status} ${response.statusText}`
      );
    }

    throw new ApiError(
      response.status,
      errorData.error || 'request_failed',
      errorData.message || 'Audience preview request failed',
      errorData.details
    );
  }

  return response.json();
}
