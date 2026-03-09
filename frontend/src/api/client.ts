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
export type AssetResponse = Schemas["AssetResponse"];
export type AssetListResponse = Schemas["AssetListResponse"];
export type AssetType = Schemas["AssetType"];
export type CreateAssetRequest = Schemas["CreateAssetRequest"];
export type UpdateAssetRequest = Schemas["UpdateAssetRequest"];
export type PortfolioSummary = Schemas["PortfolioSummary"];
export type AllocationEntry = Schemas["AllocationEntry"];
export type AssetHistoryEntry = Schemas["AssetHistoryEntry"];
export type AllocationResponse = Schemas["AllocationResponse"];
export type GoalResponse = Schemas["GoalResponse"];
export type GoalListResponse = Schemas["GoalListResponse"];
export type GoalType = Schemas["GoalType"];
export type CreateGoalRequest = Schemas["CreateGoalRequest"];
export type UpdateGoalRequest = Schemas["UpdateGoalRequest"];
export type GoalReorderRequest = Schemas["GoalReorderRequest"];
export type GoalContributionResponse = Schemas["GoalContributionResponse"];
export type GoalContributionListResponse = Schemas["GoalContributionListResponse"];
export type CreateGoalContributionRequest = Schemas["CreateGoalContributionRequest"];

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

type AssetListResult = paths["/assets"]["get"]["responses"]["200"]["content"]["application/json"];

export function listAssets(page = 1, perPage = 100): Promise<AssetListResult> {
  return request<AssetListResult>(`/api/v1/assets?page=${page}&per_page=${perPage}`);
}

type AssetResult = paths["/assets/{id}"]["get"]["responses"]["200"]["content"]["application/json"];

export function getAsset(id: string): Promise<AssetResult> {
  return request<AssetResult>(`/api/v1/assets/${id}`);
}

type CreateAssetBody = paths["/assets"]["post"]["requestBody"]["content"]["application/json"];
type CreateAssetResult =
  paths["/assets"]["post"]["responses"]["201"]["content"]["application/json"];

export function createAsset(body: CreateAssetBody): Promise<CreateAssetResult> {
  return request<CreateAssetResult>("/api/v1/assets", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

type UpdateAssetBody = paths["/assets/{id}"]["put"]["requestBody"]["content"]["application/json"];
type UpdateAssetResult =
  paths["/assets/{id}"]["put"]["responses"]["200"]["content"]["application/json"];

export function updateAsset(id: string, body: UpdateAssetBody): Promise<UpdateAssetResult> {
  return request<UpdateAssetResult>(`/api/v1/assets/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteAsset(id: string): Promise<void> {
  return request<void>(`/api/v1/assets/${id}`, { method: "DELETE" });
}

type AssetHistoryResult =
  paths["/assets/{id}/history"]["get"]["responses"]["200"]["content"]["application/json"];

type HistoryPeriod = NonNullable<
  paths["/assets/{id}/history"]["get"]["parameters"]["query"]
>["period"];

export function getAssetHistory(id: string, period?: HistoryPeriod): Promise<AssetHistoryResult> {
  const query = period ? `?period=${period}` : "";
  return request<AssetHistoryResult>(`/api/v1/assets/${id}/history${query}`);
}

type AllocationResult =
  paths["/assets/allocation"]["get"]["responses"]["200"]["content"]["application/json"];

type AllocationPeriod = NonNullable<
  paths["/assets/allocation"]["get"]["parameters"]["query"]
>["period"];

export function getAssetAllocation(period?: AllocationPeriod): Promise<AllocationResult> {
  const query = period ? `?period=${period}` : "";
  return request<AllocationResult>(`/api/v1/assets/allocation${query}`);
}

type GoalListResult = paths["/goals"]["get"]["responses"]["200"]["content"]["application/json"];

export function listGoals(page = 1, perPage = 100): Promise<GoalListResult> {
  return request<GoalListResult>(`/api/v1/goals?page=${page}&per_page=${perPage}`);
}

type GoalResult = paths["/goals/{id}"]["get"]["responses"]["200"]["content"]["application/json"];

export function getGoal(id: string): Promise<GoalResult> {
  return request<GoalResult>(`/api/v1/goals/${id}`);
}

type CreateGoalBody = paths["/goals"]["post"]["requestBody"]["content"]["application/json"];
type CreateGoalResult = paths["/goals"]["post"]["responses"]["201"]["content"]["application/json"];

export function createGoal(body: CreateGoalBody): Promise<CreateGoalResult> {
  return request<CreateGoalResult>("/api/v1/goals", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

type UpdateGoalBody = paths["/goals/{id}"]["put"]["requestBody"]["content"]["application/json"];
type UpdateGoalResult =
  paths["/goals/{id}"]["put"]["responses"]["200"]["content"]["application/json"];

export function updateGoal(id: string, body: UpdateGoalBody): Promise<UpdateGoalResult> {
  return request<UpdateGoalResult>(`/api/v1/goals/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteGoal(id: string): Promise<void> {
  return request<void>(`/api/v1/goals/${id}`, { method: "DELETE" });
}

type ReorderGoalsBody =
  paths["/goals/reorder"]["put"]["requestBody"]["content"]["application/json"];
type ReorderGoalsResult =
  paths["/goals/reorder"]["put"]["responses"]["200"]["content"]["application/json"];

export function reorderGoals(body: ReorderGoalsBody): Promise<ReorderGoalsResult> {
  return request<ReorderGoalsResult>("/api/v1/goals/reorder", {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

type ContributionListResult =
  paths["/goals/{id}/contributions"]["get"]["responses"]["200"]["content"]["application/json"];

export function listGoalContributions(
  goalId: string,
  page = 1,
  perPage = 50,
): Promise<ContributionListResult> {
  return request<ContributionListResult>(
    `/api/v1/goals/${goalId}/contributions?page=${page}&per_page=${perPage}`,
  );
}

type CreateContributionBody =
  paths["/goals/{id}/contributions"]["post"]["requestBody"]["content"]["application/json"];
type CreateContributionResult =
  paths["/goals/{id}/contributions"]["post"]["responses"]["201"]["content"]["application/json"];

export function createGoalContribution(
  goalId: string,
  body: CreateContributionBody,
): Promise<CreateContributionResult> {
  return request<CreateContributionResult>(`/api/v1/goals/${goalId}/contributions`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function deleteGoalContribution(goalId: string, contributionId: string): Promise<void> {
  return request<void>(`/api/v1/goals/${goalId}/contributions/${contributionId}`, {
    method: "DELETE",
  });
}
