import { useEffect, useState } from "react"
import { Terminal, useTerminal } from "@wterm/react"
import type { GhosttyCore } from "@wterm/ghostty"
import "@wterm/react/css"

import { Skeleton } from "@/components/ui/skeleton"
import { ghosttyCore } from "@/lib/ghostty-core"
import type { Writer } from "@/hooks/use-terminal-session"

const termFont = '14px "JetBrainsMono NFM"'

type Props = {
  onData: (data: string) => void
  onResize: (cols: number, rows: number) => void
  onTitle: (title: string) => void
  registerWriter: (write: Writer) => void
}

async function loadTerminalFont(): Promise<void> {
  if (!document.fonts) {
    return
  }
  await Promise.all([
    document.fonts.load(`400 ${termFont}`),
    document.fonts.load(`700 ${termFont}`),
    document.fonts.load(`italic 400 ${termFont}`),
    document.fonts.load(`italic 700 ${termFont}`),
  ])
}

export function WTermView({
  onData,
  onResize,
  onTitle,
  registerWriter,
}: Props) {
  const { ref, write } = useTerminal()
  const [core, setCore] = useState<GhosttyCore | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    Promise.all([ghosttyCore, loadTerminalFont().catch(() => undefined)])
      .then(([loaded]) => {
        if (!cancelled) {
          setCore(loaded)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(
            err instanceof Error ? err.message : "failed to load terminal"
          )
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  if (error) {
    return (
      <div className="flex h-full w-full items-center justify-center bg-[#1e1e1e] p-6 text-sm text-[#d4d4d4]">
        {error}
      </div>
    )
  }

  if (!core) {
    return (
      <div className="flex h-full w-full items-center justify-center bg-[#1e1e1e]">
        <Skeleton className="h-4 w-48 bg-[#3c3c3c]" />
      </div>
    )
  }

  return (
    <Terminal
      ref={ref}
      core={core}
      autoResize
      cursorBlink
      className="h-full w-full"
      onData={onData}
      onResize={onResize}
      onTitle={onTitle}
      onReady={() => {
        registerWriter(write)
      }}
      onError={(err) => {
        setError(err instanceof Error ? err.message : "terminal error")
      }}
    />
  )
}
