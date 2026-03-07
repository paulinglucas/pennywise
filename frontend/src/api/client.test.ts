import { describe, it, expect, vi, beforeEach } from "vitest";
import { ApiError, login, getMe, logout } from "./client";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

describe("ApiError", () => {
  it("stores status, code, message, and requestId", () => {
    const err = new ApiError(400, "VALIDATION_FAILED", "bad input", "req-123");
    expect(err.status).toBe(400);
    expect(err.code).toBe("VALIDATION_FAILED");
    expect(err.message).toBe("bad input");
    expect(err.requestId).toBe("req-123");
    expect(err.name).toBe("ApiError");
  });
});

describe("login", () => {
  it("sends credentials and returns login response", async () => {
    const responseBody = { user: { id: "u1", email: "test@test.com", name: "Test" } };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(responseBody),
    });

    const result = await login({ email: "test@test.com", password: "pass" });
    expect(result).toEqual(responseBody);

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/v1/auth/login");
    expect(options.method).toBe("POST");
    expect(options.credentials).toBe("include");
    expect(JSON.parse(options.body as string)).toEqual({
      email: "test@test.com",
      password: "pass",
    });
  });

  it("throws ApiError on 401", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: "Unauthorized",
      json: () =>
        Promise.resolve({
          error: { code: "UNAUTHORIZED", message: "Invalid credentials", request_id: "r1" },
        }),
    });

    await expect(login({ email: "bad@test.com", password: "wrong" })).rejects.toThrow(ApiError);
  });
});

describe("getMe", () => {
  it("sends GET with credentials", async () => {
    const user = { id: "u1", email: "test@test.com", name: "Test" };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(user),
    });

    const result = await getMe();
    expect(result).toEqual(user);

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/v1/auth/me");
    expect(options.credentials).toBe("include");
  });
});

describe("logout", () => {
  it("sends POST to logout", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
    });

    await logout();

    const [url, options] = mockFetch.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/v1/auth/logout");
    expect(options.method).toBe("POST");
  });
});

describe("error handling", () => {
  it("handles non-JSON error responses", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
      json: () => Promise.reject(new Error("not json")),
    });

    try {
      await getMe();
      expect.fail("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      const apiErr = err as ApiError;
      expect(apiErr.status).toBe(500);
      expect(apiErr.code).toBe("INTERNAL_ERROR");
    }
  });
});
