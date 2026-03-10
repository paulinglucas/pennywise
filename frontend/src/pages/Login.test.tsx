import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import Login from "./Login";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

beforeEach(() => {
  mockFetch.mockReset();
});

describe("Login", () => {
  it("renders email and password fields with labels", () => {
    renderWithProviders(<Login />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("submits credentials on form submit", async () => {
    const user = userEvent.setup();
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ user: { id: "u1", email: "a@b.com", name: "Test" } }),
    });

    renderWithProviders(<Login />);

    await user.type(screen.getByLabelText(/email/i), "a@b.com");
    await user.type(screen.getByLabelText(/password/i), "secret");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/auth/login",
        expect.objectContaining({ method: "POST" }),
      );
    });

    const body = JSON.parse((mockFetch.mock.calls[0] as [string, RequestInit])[1].body as string);
    expect(body).toEqual({ email: "a@b.com", password: "secret" });
  });

  it("displays error message on failed login", async () => {
    const user = userEvent.setup();
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: "Unauthorized",
      json: () =>
        Promise.resolve({
          error: {
            code: "UNAUTHORIZED",
            message: "Invalid email or password",
            request_id: "r1",
          },
        }),
    });

    renderWithProviders(<Login />);

    await user.type(screen.getByLabelText(/email/i), "bad@test.com");
    await user.type(screen.getByLabelText(/password/i), "wrong");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("Invalid email or password");
    });
  });

  it("shows loading state while submitting", async () => {
    const user = userEvent.setup();
    let resolveLogin: (value: unknown) => void;
    mockFetch.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveLogin = resolve;
      }),
    );

    renderWithProviders(<Login />);

    await user.type(screen.getByLabelText(/email/i), "a@b.com");
    await user.type(screen.getByLabelText(/password/i), "pass");
    await user.click(screen.getByRole("button", { name: /sign in/i }));

    const submitButton = screen.getByRole("button", { name: /signing in/i });
    expect(submitButton).toHaveTextContent("Signing in...");
    expect(submitButton).toBeDisabled();

    resolveLogin!({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ user: { id: "u1", email: "a@b.com", name: "Test" } }),
    });
  });

  it("toggles to registration form", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Login />);

    expect(screen.queryByLabelText(/name/i)).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /sign up/i }));

    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create account/i })).toBeInTheDocument();
  });

  it("submits registration form", async () => {
    const user = userEvent.setup();
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () => Promise.resolve({ user: { id: "u2", email: "new@test.com", name: "New" } }),
    });

    renderWithProviders(<Login />);

    await user.click(screen.getByRole("button", { name: /sign up/i }));
    await user.type(screen.getByLabelText(/name/i), "New");
    await user.type(screen.getByLabelText(/email/i), "new@test.com");
    await user.type(screen.getByLabelText(/password/i), "securepass123");
    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        "/api/v1/auth/register",
        expect.objectContaining({ method: "POST" }),
      );
    });

    const body = JSON.parse((mockFetch.mock.calls[0] as [string, RequestInit])[1].body as string);
    expect(body).toEqual({ email: "new@test.com", password: "securepass123", name: "New" });
  });

  it("displays error on failed registration", async () => {
    const user = userEvent.setup();
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 409,
      statusText: "Conflict",
      json: () =>
        Promise.resolve({
          error: {
            code: "CONFLICT",
            message: "Maximum number of users reached",
            request_id: "r2",
          },
        }),
    });

    renderWithProviders(<Login />);

    await user.click(screen.getByRole("button", { name: /sign up/i }));
    await user.type(screen.getByLabelText(/name/i), "Third");
    await user.type(screen.getByLabelText(/email/i), "third@test.com");
    await user.type(screen.getByLabelText(/password/i), "securepass123");
    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("Maximum number of users reached");
    });
  });

  it("toggles back to login from registration", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Login />);

    await user.click(screen.getByRole("button", { name: /sign up/i }));
    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /sign in/i }));
    expect(screen.queryByLabelText(/name/i)).not.toBeInTheDocument();
  });
});
