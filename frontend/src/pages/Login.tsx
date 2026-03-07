import { useState, type FormEvent } from "react";
import { useLogin } from "@/hooks/useAuth";
import { ApiError } from "@/api/client";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const loginMutation = useLogin();

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    loginMutation.mutate({ email, password });
  }

  const errorMessage =
    loginMutation.error instanceof ApiError
      ? loginMutation.error.message
      : loginMutation.error?.message;

  return (
    <div
      className="flex min-h-screen items-center justify-center"
      style={{ backgroundColor: "var(--color-background)" }}
    >
      <div
        className="w-full max-w-sm rounded-lg p-8"
        style={{ backgroundColor: "var(--color-surface)" }}
      >
        <h1
          className="mb-8 text-center text-2xl font-semibold"
          style={{ color: "var(--color-text-primary)" }}
        >
          Pennywise
        </h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="email"
              className="mb-1 block text-sm font-medium"
              style={{ color: "var(--color-text-secondary)" }}
            >
              Email
            </label>
            <input
              id="email"
              type="email"
              required
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              className="w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2"
              style={{
                backgroundColor: "var(--color-background)",
                borderColor: "var(--color-border)",
                color: "var(--color-text-primary)",
              }}
              autoComplete="email"
            />
          </div>
          <div>
            <label
              htmlFor="password"
              className="mb-1 block text-sm font-medium"
              style={{ color: "var(--color-text-secondary)" }}
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              required
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2"
              style={{
                backgroundColor: "var(--color-background)",
                borderColor: "var(--color-border)",
                color: "var(--color-text-primary)",
              }}
              autoComplete="current-password"
            />
          </div>
          {errorMessage && (
            <p className="text-sm" style={{ color: "var(--color-error)" }} role="alert">
              {errorMessage}
            </p>
          )}
          <button
            type="submit"
            disabled={loginMutation.isPending}
            className="w-full rounded-md px-4 py-2 text-sm font-medium text-white transition-colors disabled:opacity-50"
            style={{ backgroundColor: "var(--color-accent)" }}
          >
            {loginMutation.isPending ? "Signing in..." : "Sign in"}
          </button>
        </form>
      </div>
    </div>
  );
}
