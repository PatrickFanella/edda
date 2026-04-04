import { Component, type ErrorInfo, type ReactNode } from 'react';

interface AppErrorBoundaryProps {
  readonly children: ReactNode;
}

interface AppErrorBoundaryState {
  readonly hasError: boolean;
}

export class AppErrorBoundary extends Component<AppErrorBoundaryProps, AppErrorBoundaryState> {
  override state: AppErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(): AppErrorBoundaryState {
    return { hasError: true };
  }

  override componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('frontend render failure', error, info);
  }

  override render() {
    if (this.state.hasError) {
      return (
        <main className="min-h-screen bg-obsidian px-6 py-16 text-champagne">
          <div className="mx-auto flex w-full max-w-3xl flex-col gap-4 border border-ruby/40 bg-ruby/10 p-8">
            <p className="font-heading text-sm font-semibold uppercase tracking-[0.32em] text-ruby">Frontend error</p>
            <h1 className="font-heading text-3xl font-semibold uppercase tracking-[0.12em] text-champagne">The interface hit an unexpected failure.</h1>
            <p className="text-sm leading-7 text-champagne/80">
              Reload the page to retry. If the problem persists, inspect the browser console for the captured component stack.
            </p>
          </div>
        </main>
      );
    }

    return this.props.children;
  }
}
