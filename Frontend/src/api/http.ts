import { API_BASE_URL } from "./contracts.ts";

/** A failed request, carrying the HTTP status and the message the API sent. */
export class ApiError extends Error {
  readonly status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

/** True for the AbortError a cancelled request rejects with. */
export function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === "AbortError";
}

/** Picks the most useful message out of whatever the API answered with. */
function errorMessage(payload: unknown, status: number): string {
  if (typeof payload === "string" && payload.trim()) return payload.trim();
  if (payload && typeof payload === "object") {
    const body = payload as Record<string, unknown>;
    for (const key of ["error", "message", "Error", "Message"]) {
      const value = body[key];
      if (typeof value === "string" && value.trim()) return value.trim();
    }
  }
  return `The API responded with ${status}.`;
}

/** Reads a JSON body, tolerating an empty one. */
async function readPayload(response: Response): Promise<unknown> {
  const text = (await response.text()).trim();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    throw new ApiError(`The API did not answer with JSON (${text.slice(0, 80)}).`, response.status);
  }
}

interface RequestOptions {
  method?: "GET" | "POST" | "PATCH" | "DELETE";
  /** Sent as JSON when present; its presence also sets the content type. */
  body?: unknown;
  signal?: AbortSignal;
}

/**
 * The only place that speaks HTTP: one base URL, JSON in and out, status codes
 * turned into `ApiError`, and cancellation passed straight through so a caller
 * can drop a response it no longer wants.
 */
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: options.method ?? "GET",
      headers: options.body === undefined ? undefined : { "Content-Type": "application/json" },
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
      signal: options.signal,
    });
  } catch (cause) {
    if (isAbort(cause)) throw cause;
    throw new ApiError("Could not reach the API.", 0);
  }
  const payload = await readPayload(response);
  if (!response.ok) throw new ApiError(errorMessage(payload, response.status), response.status);
  return payload as T;
}
