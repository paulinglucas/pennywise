import { useState, type FormEvent } from "react";
import { useLogin, useRegister } from "@/hooks/useAuth";
import { ApiError } from "@/api/client";
import BrandLogo from "@/components/shared/BrandLogo";

export default function Login() {
  const [isRegistering, setIsRegistering] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const loginMutation = useLogin();
  const registerMutation = useRegister();

  const activeMutation = isRegistering ? registerMutation : loginMutation;

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (isRegistering) {
      registerMutation.mutate({ email, password, name });
    } else {
      loginMutation.mutate({ email, password });
    }
  }

  function toggleMode() {
    setIsRegistering((prev) => !prev);
    loginMutation.reset();
    registerMutation.reset();
  }

  const errorMessage =
    activeMutation.error instanceof ApiError
      ? activeMutation.error.message
      : activeMutation.error?.message;

  return (
    <div
      className="flex min-h-screen items-center justify-center"
      style={{ backgroundColor: "var(--color-background)" }}
    >
      <div
        className="w-full max-w-sm rounded-lg p-8"
        style={{
          backgroundColor: "var(--color-surface)",
          border: "1px solid var(--color-accent-muted)",
          boxShadow: "var(--glow-lg)",
        }}
      >
        <h1 className="mb-8 flex justify-center">
          <BrandLogo size="lg" />
        </h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          {isRegistering && (
            <div>
              <label
                htmlFor="name"
                className="mb-1 block text-sm font-medium"
                style={{ color: "var(--color-text-secondary)" }}
              >
                Name
              </label>
              <input
                id="name"
                type="text"
                required
                value={name}
                onChange={(event) => setName(event.target.value)}
                className="w-full rounded-md border px-3 py-2 text-sm transition-shadow"
                style={{
                  backgroundColor: "var(--color-background)",
                  borderColor: "var(--color-border)",
                  color: "var(--color-text-primary)",
                }}
                autoComplete="name"
              />
            </div>
          )}
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
              className="w-full rounded-md border px-3 py-2 text-sm transition-shadow"
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
              minLength={isRegistering ? 8 : undefined}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="w-full rounded-md border px-3 py-2 text-sm transition-shadow"
              style={{
                backgroundColor: "var(--color-background)",
                borderColor: "var(--color-border)",
                color: "var(--color-text-primary)",
              }}
              autoComplete={isRegistering ? "new-password" : "current-password"}
            />
          </div>
          {errorMessage && (
            <p className="text-sm" style={{ color: "var(--color-error)" }} role="alert">
              {errorMessage}
            </p>
          )}
          <button
            type="submit"
            disabled={activeMutation.isPending}
            className="btn-primary w-full rounded-md px-4 py-2 text-sm font-medium transition-all disabled:opacity-50"
            style={{
              backgroundColor: "var(--color-accent)",
              color: "var(--color-background)",
              boxShadow: "var(--glow-accent)",
            }}
          >
            {activeMutation.isPending
              ? isRegistering
                ? "Creating account..."
                : "Signing in..."
              : isRegistering
                ? "Create account"
                : "Sign in"}
          </button>
        </form>
        <p className="mt-4 text-center text-sm" style={{ color: "var(--color-text-secondary)" }}>
          {isRegistering ? "Already have an account?" : "Need an account?"}{" "}
          <button
            type="button"
            onClick={toggleMode}
            className="font-medium underline"
            style={{ color: "var(--color-accent)" }}
          >
            {isRegistering ? "Sign in" : "Sign up"}
          </button>
        </p>
      </div>
    </div>
  );
}
