import type { paths, components } from "./generated";

type Schemas = components["schemas"];
export type ErrorCode = Schemas["ErrorCode"];
export type UserResponse = Schemas["UserResponse"];
export type LoginRequest = Schemas["LoginRequest"];
export type LoginResponse = Schemas["LoginResponse"];
export type TransactionResponse = Schemas["TransactionResponse"];
export type TransactionListResponse = Schemas["TransactionListResponse"];
export type CreateTransactionRequest = Schemas["CreateTransactionRequest"];
export type UpdateTransactionRequest = Schemas["UpdateTransactionRequest"];
export type ImportResponse = Schemas["ImportResponse"];
export type AccountResponse = Schemas["AccountResponse"];
export type AccountListResponse = Schemas["AccountListResponse"];
export type TransactionType = Schemas["TransactionType"];
export type PaginationMeta = Schemas["PaginationMeta"];
export type TransactionGroupResponse = Schemas["TransactionGroupResponse"];
export type TransactionGroupListResponse = Schemas["TransactionGroupListResponse"];
export type CreateTransactionGroupRequest = Schemas["CreateTransactionGroupRequest"];
export type UpdateTransactionGroupRequest = Schemas["UpdateTransactionGroupRequest"];
export type TransactionGroupMemberInput = Schemas["TransactionGroupMemberInput"];
export type TransactionGroupMemberUpdate = Schemas["TransactionGroupMemberUpdate"];

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

export type DashboardResponse = Schemas["DashboardResponse"];
type DashboardResult =
  paths["/dashboard"]["get"]["responses"]["200"]["content"]["application/json"];

export type SpendingPeriod = NonNullable<
  paths["/dashboard"]["get"]["parameters"]["query"]
>["spending_period"];

export function getDashboard(spendingPeriod?: SpendingPeriod): Promise<DashboardResult> {
  const params = spendingPeriod ? `?spending_period=${spendingPeriod}` : "";
  return request<DashboardResult>(`/api/v1/dashboard${params}`);
}

export type NetWorthHistoryResponse = Schemas["NetWorthHistoryResponse"];
type NetWorthHistoryResult =
  paths["/dashboard/net-worth-history"]["get"]["responses"]["200"]["content"]["application/json"];
type NetWorthPeriod = NonNullable<
  paths["/dashboard/net-worth-history"]["get"]["parameters"]["query"]
>["period"];

export function getNetWorthHistory(period: NetWorthPeriod = "1y"): Promise<NetWorthHistoryResult> {
  return request<NetWorthHistoryResult>(`/api/v1/dashboard/net-worth-history?period=${period}`);
}

export interface TransactionFilters {
  page?: number;
  per_page?: number;
  account_id?: string;
  category?: string;
  type?: TransactionType;
  date_from?: string;
  date_to?: string;
  amount_min?: number;
  amount_max?: number;
  tags?: string;
  search?: string;
  group_id?: string;
}

function buildQueryString(params: Record<string, string | number | undefined>): string {
  const entries = Object.entries(params).filter(
    (entry): entry is [string, string | number] => entry[1] !== undefined && entry[1] !== "",
  );
  if (entries.length === 0) return "";
  return "?" + entries.map(([key, val]) => `${key}=${encodeURIComponent(val)}`).join("&");
}

type TransactionListResult =
  paths["/transactions"]["get"]["responses"]["200"]["content"]["application/json"];

export function listTransactions(filters?: TransactionFilters): Promise<TransactionListResult> {
  const query = buildQueryString((filters ?? {}) as Record<string, string | number | undefined>);
  return request<TransactionListResult>(`/api/v1/transactions${query}`);
}

type CreateTransactionBody =
  paths["/transactions"]["post"]["requestBody"]["content"]["application/json"];
type CreateTransactionResult =
  paths["/transactions"]["post"]["responses"]["201"]["content"]["application/json"];

export function createTransaction(body: CreateTransactionBody): Promise<CreateTransactionResult> {
  return request<CreateTransactionResult>("/api/v1/transactions", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

type UpdateTransactionBody =
  paths["/transactions/{id}"]["put"]["requestBody"]["content"]["application/json"];
type UpdateTransactionResult =
  paths["/transactions/{id}"]["put"]["responses"]["200"]["content"]["application/json"];

export function updateTransaction(
  id: string,
  body: UpdateTransactionBody,
): Promise<UpdateTransactionResult> {
  return request<UpdateTransactionResult>(`/api/v1/transactions/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteTransaction(id: string): Promise<void> {
  return request<void>(`/api/v1/transactions/${id}`, { method: "DELETE" });
}

type ImportResult =
  paths["/transactions/import"]["post"]["responses"]["201"]["content"]["application/json"];

export function importTransactions(file: File, accountId: string): Promise<ImportResult> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("account_id", accountId);
  return requestRaw<ImportResult>("/api/v1/transactions/import", {
    method: "POST",
    body: formData,
  });
}

async function requestRaw<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    credentials: "include",
  });

  if (!response.ok) {
    throw await parseErrorResponse(response);
  }

  return response.json() as Promise<T>;
}

type AccountListResult =
  paths["/accounts"]["get"]["responses"]["200"]["content"]["application/json"];

export function listAccounts(page = 1, perPage = 100): Promise<AccountListResult> {
  return request<AccountListResult>(`/api/v1/accounts?page=${page}&per_page=${perPage}`);
}

type CategoriesResult =
  paths["/categories"]["get"]["responses"]["200"]["content"]["application/json"];

export function listCategories(): Promise<string[]> {
  return request<CategoriesResult>("/api/v1/categories").then((r) => r.categories);
}

type GroupListResult =
  paths["/transaction-groups"]["get"]["responses"]["200"]["content"]["application/json"];

export function listTransactionGroups(page = 1, perPage = 20): Promise<GroupListResult> {
  return request<GroupListResult>(`/api/v1/transaction-groups?page=${page}&per_page=${perPage}`);
}

type GroupResult =
  paths["/transaction-groups/{id}"]["get"]["responses"]["200"]["content"]["application/json"];

export function getTransactionGroup(id: string): Promise<GroupResult> {
  return request<GroupResult>(`/api/v1/transaction-groups/${id}`);
}

type CreateGroupBody =
  paths["/transaction-groups"]["post"]["requestBody"]["content"]["application/json"];
type CreateGroupResult =
  paths["/transaction-groups"]["post"]["responses"]["201"]["content"]["application/json"];

export function createTransactionGroup(body: CreateGroupBody): Promise<CreateGroupResult> {
  return request<CreateGroupResult>("/api/v1/transaction-groups", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

type UpdateGroupBody =
  paths["/transaction-groups/{id}"]["put"]["requestBody"]["content"]["application/json"];
type UpdateGroupResult =
  paths["/transaction-groups/{id}"]["put"]["responses"]["200"]["content"]["application/json"];

export function updateTransactionGroup(
  id: string,
  body: UpdateGroupBody,
): Promise<UpdateGroupResult> {
  return request<UpdateGroupResult>(`/api/v1/transaction-groups/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteTransactionGroup(id: string): Promise<void> {
  return request<void>(`/api/v1/transaction-groups/${id}`, {
    method: "DELETE",
  });
}
