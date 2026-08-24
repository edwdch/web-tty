import { useState } from "react"

import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"

type PingState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ok"; message: string }
  | { status: "error"; message: string }

export function App() {
  const [state, setState] = useState<PingState>({ status: "idle" })

  async function ping() {
    setState({ status: "loading" })
    try {
      const res = await fetch("/api/ping")
      if (!res.ok) {
        setState({ status: "error", message: `HTTP ${res.status}` })
        return
      }
      const data = (await res.json()) as { message?: string }
      setState({ status: "ok", message: data.message ?? "ok" })
    } catch (err) {
      setState({
        status: "error",
        message: err instanceof Error ? err.message : "request failed",
      })
    }
  }

  return (
    <div className="flex min-h-svh flex-col">
      <header className="flex h-12 shrink-0 items-center justify-end px-3">
        <ThemeToggle />
      </header>
      <div className="flex flex-1 items-center justify-center p-6">
        <div className="flex w-full max-w-sm flex-col gap-4">
          <div>
            <p className="text-muted-foreground text-xs font-medium tracking-wide uppercase">
              simple-app
            </p>
            <h1 className="mt-1 text-lg font-medium">API ping</h1>
            <p className="text-muted-foreground text-sm">
              Calls{" "}
              <code className="text-foreground font-mono">GET /api/ping</code>
            </p>
          </div>
          <Button onClick={ping} disabled={state.status === "loading"}>
            {state.status === "loading" ? "Pinging…" : "Ping"}
          </Button>
          <output className="border-border bg-muted/40 min-h-10 rounded-lg border px-3 py-2 font-mono text-sm">
            {state.status === "idle" && "waiting"}
            {state.status === "loading" && "…"}
            {state.status === "ok" && state.message}
            {state.status === "error" && state.message}
          </output>
          <p className="text-muted-foreground text-xs">
            Press <kbd>d</kbd> to toggle dark mode
          </p>
        </div>
      </div>
    </div>
  )
}

export default App
