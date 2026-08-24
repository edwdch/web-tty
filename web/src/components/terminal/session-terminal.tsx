import { useCallback, useEffect, useRef, useState } from "react"

import { DisconnectDialog } from "@/components/terminal/disconnect-dialog"
import { SessionPickerDialog } from "@/components/terminal/session-picker-dialog"
import { SessionSwitcherButton } from "@/components/terminal/session-switcher-button"
import { WTermView } from "@/components/terminal/wterm-view"
import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { useTerminalSession } from "@/hooks/use-terminal-session"
import {
  deleteSession,
  fetchSessions,
  pickDefaultSession,
  saveLastSessionId,
  type SessionInfo,
} from "@/lib/sessions"

export function SessionTerminal() {
  const term = useTerminalSession()
  const [termKey, setTermKey] = useState(0)
  const [sessions, setSessions] = useState<SessionInfo[] | null>(null)
  const [pickerOpen, setPickerOpen] = useState(false)
  const [pickerMode, setPickerMode] = useState<"boot" | "switch">("boot")
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [confirmClose, setConfirmClose] = useState<SessionInfo | null>(null)
  const bootDecided = useRef(false)

  const refreshSessions = useCallback(async () => {
    try {
      const list = await fetchSessions()
      setSessions(list)
      return list
    } catch {
      setSessions([])
      return [] as SessionInfo[]
    }
  }, [])

  useEffect(() => {
    if (term.sessionId) {
      saveLastSessionId(term.sessionId)
    }
  }, [term.sessionId])

  const connect = term.connect
  const termReady = term.ready

  useEffect(() => {
    if (!termReady || bootDecided.current) {
      return
    }
    let cancelled = false
    fetchSessions()
      .then((list) => {
        if (cancelled || bootDecided.current) {
          return
        }
        bootDecided.current = true
        setSessions(list)
        if (list.length === 0) {
          connect("new")
          return
        }
        setSelectedId(pickDefaultSession(list))
        setPickerMode("boot")
        setPickerOpen(true)
      })
      .catch(() => {
        if (cancelled || bootDecided.current) {
          return
        }
        bootDecided.current = true
        setSessions([])
        connect("new")
      })
    return () => {
      cancelled = true
    }
  }, [termReady, connect])

  const remountAndConnect = (target: "new" | string) => {
    term.resetWriter()
    setTermKey((key) => key + 1)
    term.connect(target)
  }

  const handleContinue = () => {
    if (!selectedId) {
      return
    }
    setPickerOpen(false)
    if (selectedId === term.sessionId) {
      return
    }
    if (term.sessionId === null) {
      term.connect(selectedId)
      return
    }
    remountAndConnect(selectedId)
  }

  const handleNew = () => {
    setPickerOpen(false)
    if (pickerMode === "boot" && !term.sessionId) {
      term.connect("new")
      return
    }
    remountAndConnect("new")
  }

  const handleOpenSwitcher = async () => {
    const list = await refreshSessions()
    setSelectedId(term.sessionId ?? pickDefaultSession(list))
    setPickerMode("switch")
    setPickerOpen(true)
  }

  const closeSession = async (info: SessionInfo) => {
    const current = info.id === term.sessionId
    if (current) {
      term.disconnectExpected()
    }
    await deleteSession(info.id)
    const list = await refreshSessions()
    if (!current) {
      if (selectedId === info.id) {
        setSelectedId(pickDefaultSession(list))
      }
      return
    }
    term.resetWriter()
    setTermKey((key) => key + 1)
    if (list.length === 0) {
      setPickerOpen(false)
      term.connect("new")
      return
    }
    setSelectedId(pickDefaultSession(list))
    setPickerMode("boot")
    setPickerOpen(true)
  }

  const handleRequestCloseSession = (info: SessionInfo) => {
    if (info.clients > 0) {
      setConfirmClose(info)
      return
    }
    void closeSession(info)
  }

  const showSwitcher =
    term.sessionId !== null && !term.disconnected && !pickerOpen

  return (
    <div className="h-svh w-full overflow-hidden bg-[#1e1e1e]">
      <WTermView
        key={termKey}
        onData={term.sendInput}
        onResize={term.sendResize}
        onTitle={(title) => {
          document.title = title || "web-tty"
        }}
        registerWriter={term.registerWriter}
      />
      {showSwitcher ? (
        <SessionSwitcherButton onClick={() => void handleOpenSwitcher()} />
      ) : null}
      <SessionPickerDialog
        open={pickerOpen}
        mode={pickerMode}
        sessions={sessions ?? []}
        currentId={term.sessionId}
        selectedId={selectedId}
        onSelectedId={setSelectedId}
        onContinue={handleContinue}
        onNew={handleNew}
        onRequestCloseSession={handleRequestCloseSession}
        onOpenChange={setPickerOpen}
      />
      <AlertDialog
        open={confirmClose !== null}
        onOpenChange={(open) => {
          if (!open) {
            setConfirmClose(null)
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Close this session?</AlertDialogTitle>
            <AlertDialogDescription>
              This kills the shell even if other tabs are attached. It cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setConfirmClose(null)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              onClick={() => {
                const info = confirmClose
                setConfirmClose(null)
                if (info) {
                  void closeSession(info)
                }
              }}
            >
              Close session
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <DisconnectDialog
        open={term.disconnected}
        onReconnect={term.reconnect}
        onClosePage={term.closePage}
      />
    </div>
  )
}
