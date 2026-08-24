import { XIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  formatCreatedAt,
  shortSessionId,
  type SessionInfo,
} from "@/lib/sessions"
import { cn } from "@/lib/utils"

export type PickerMode = "boot" | "switch"

type Props = {
  open: boolean
  mode: PickerMode
  sessions: SessionInfo[]
  currentId: string | null
  selectedId: string | null
  onSelectedId: (id: string) => void
  onContinue: () => void
  onNew: () => void
  onRequestCloseSession: (session: SessionInfo) => void
  onOpenChange: (open: boolean) => void
}

export function SessionPickerDialog({
  open,
  mode,
  sessions,
  currentId,
  selectedId,
  onSelectedId,
  onContinue,
  onNew,
  onRequestCloseSession,
  onOpenChange,
}: Props) {
  const dismissable = mode === "switch"

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next && !dismissable) {
          return
        }
        onOpenChange(next)
      }}
    >
      <DialogContent
        showCloseButton={dismissable}
        className="sm:max-w-md"
        onPointerDownOutside={(event) => {
          if (!dismissable) {
            event.preventDefault()
          }
        }}
        onEscapeKeyDown={(event) => {
          if (!dismissable) {
            event.preventDefault()
          }
        }}
      >
        <DialogHeader>
          <DialogTitle>
            {mode === "boot" ? "Resume a session" : "Sessions"}
          </DialogTitle>
          <DialogDescription>
            {mode === "boot"
              ? "A session is still running. Continue one or start a new shell."
              : "Switch to another session or close a shell. Closing kills the process even if no tab is attached."}
          </DialogDescription>
        </DialogHeader>
        <ul className="max-h-64 overflow-y-auto rounded-lg ring-1 ring-foreground/10">
          {sessions.length === 0 ? (
            <li className="px-3 py-6 text-center text-sm text-muted-foreground">
              No running sessions
            </li>
          ) : (
            sessions.map((item) => {
              const selected = item.id === selectedId
              const current = item.id === currentId
              return (
                <li
                  key={item.id}
                  className="border-b border-foreground/10 last:border-b-0"
                >
                  <div
                    className={cn(
                      "flex items-stretch gap-1",
                      selected && "bg-muted"
                    )}
                  >
                    <button
                      type="button"
                      className="flex min-w-0 flex-1 flex-col items-start gap-0.5 px-3 py-2 text-left"
                      onClick={() => onSelectedId(item.id)}
                    >
                      <span className="font-mono text-sm">
                        {shortSessionId(item.id)}
                        {current ? " · current" : ""}
                      </span>
                      <span className="w-full truncate text-xs text-muted-foreground">
                        {item.cwd || "shell"}
                        {" · "}
                        {item.clients} attached
                        {" · "}
                        {formatCreatedAt(item.createdAt)}
                      </span>
                    </button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      className="mt-1.5 mr-1 text-destructive hover:bg-destructive/10 hover:text-destructive"
                      title="Close session"
                      onClick={() => onRequestCloseSession(item)}
                    >
                      <XIcon />
                      <span className="sr-only">Close session</span>
                    </Button>
                  </div>
                </li>
              )
            })
          )}
        </ul>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onNew}>
            New session
          </Button>
          <Button
            type="button"
            onClick={onContinue}
            disabled={!selectedId || sessions.length === 0}
          >
            Continue
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
