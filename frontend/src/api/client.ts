import type { paths, components } from "./generated";

type Schemas = components["schemas"];
export type ErrorCode = Schemas["ErrorCode"];
export type UserResponse = Schemas["UserResponse"];
export type LoginRequest = Schemas["LoginRequest"];
export type LoginResponse = Schemas["LoginResponse"];

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: ErrorCode,
    message: string,
    public requestId: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function parseErrorResponse(response: Response): Promise<ApiError> {
  try {
    const body = (await response.json()) as { error: Schemas["ErrorResponse"]["error"] };
    return new ApiError(
      response.status,
      body.error.code,
      body.error.message,
      body.error.request_id,
    );
  } catch {
    return new ApiError(
      response.status,
      "INTERNAL_ERROR",
      response.statusText || "Unknown error",
      "",
    );
  }
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

type LoginBody = paths["/auth/login"]["post"]["requestBody"]["content"]["application/json"];
type LoginResult = paths["/auth/login"]["post"]["responses"]["200"]["content"]["application/json"];

export function login(body: LoginBody): Promise<LoginResult> {
  return request<LoginResult>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

type MeResult = paths["/auth/me"]["get"]["responses"]["200"]["content"]["application/json"];

export function getMe(): Promise<MeResult> {
  return request<MeResult>("/api/v1/auth/me");
}

export function logout(): Promise<void> {
  return request<void>("/api/v1/auth/logout", { method: "POST" });
}
