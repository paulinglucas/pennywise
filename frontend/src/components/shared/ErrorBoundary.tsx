import { Component, type ReactNode } from "react";

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (!this.state.hasError) {
      return this.props.children;
    }

    return (
      <div
        className="flex h-screen flex-col items-center justify-center text-center"
        style={{ backgroundColor: "var(--color-background)" }}
      >
        <h1 className="mb-2 text-xl font-semibold" style={{ color: "var(--color-text-primary)" }}>
          Something went wrong
        </h1>
        <p className="mb-6 max-w-md text-sm" style={{ color: "var(--color-text-secondary)" }}>
          An unexpected error occurred. Please reload the page and try again.
        </p>
        <button
          onClick={this.handleReload}
          className="rounded-md px-4 py-2 text-sm font-medium transition-all"
          style={{
            backgroundColor: "var(--color-accent)",
            color: "var(--color-background)",
            boxShadow: "var(--glow-accent)",
          }}
        >
          Reload Page
        </button>
      </div>
    );
  }
}
