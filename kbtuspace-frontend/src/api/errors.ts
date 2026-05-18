import { AxiosError } from 'axios';

interface ErrorResponse {
  error?: string;
}

export function getApiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof AxiosError) {
    const data = error.response?.data as ErrorResponse | undefined;
    return data?.error || fallback;
  }
  if (error && typeof error === 'object') {
    const maybe = error as { response?: { data?: ErrorResponse } };
    if (maybe.response?.data?.error) {
      return maybe.response.data.error;
    }
  }
  return fallback;
}
