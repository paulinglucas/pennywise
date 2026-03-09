import { useState, useEffect } from "react";
import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { computeProjection, type ProjectionRequest, type ProjectionResponse } from "@/api/client";

export interface ProjectionParams {
  monthlySavingsAdjustment: number;
  returnRate: number;
  yearsToProject: number;
  oneTimeEvents: Array<{
    amount: number;
    date: string;
    type: "windfall" | "expense";
  }>;
}

const DEFAULT_PARAMS: ProjectionParams = {
  monthlySavingsAdjustment: 0,
  returnRate: 7,
  yearsToProject: 10,
  oneTimeEvents: [],
};

function paramsToRequest(params: ProjectionParams): ProjectionRequest {
  const req: ProjectionRequest = {
    years_to_project: params.yearsToProject,
    monthly_savings_adjustment: params.monthlySavingsAdjustment,
    return_rate: params.returnRate,
  };
  if (params.oneTimeEvents.length > 0) {
    req.one_time_events = params.oneTimeEvents;
  }
  return req;
}

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}

export function useProjections(params: ProjectionParams = DEFAULT_PARAMS) {
  const debouncedParams = useDebouncedValue(params, 300);
  const requestBody = paramsToRequest(debouncedParams);

  return useQuery<ProjectionResponse>({
    queryKey: ["projections", requestBody],
    queryFn: () => computeProjection(requestBody),
    staleTime: 30 * 1000,
    placeholderData: keepPreviousData,
  });
}

export { DEFAULT_PARAMS };
