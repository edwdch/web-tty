import { DisconnectDialog } from "@/components/terminal/disconnect-dialog"
import { WTermView } from "@/components/terminal/wterm-view"
import { useTerminalSession } from "@/hooks/use-terminal-session"

export function SessionTerminal() {
  const session = useTerminalSession()

  return (
    <div className="h-svh w-full overflow-hidden bg-[#1e1e1e]">
      <WTermView
        onData={session.sendInput}
        onResize={session.sendResize}
        onTitle={(title) => {
          document.title = title || "web-tty"
        }}
        registerWriter={session.registerWriter}
      />
      <DisconnectDialog
        open={session.disconnected}
        onReconnect={session.reconnect}
        onClosePage={session.closePage}
      />
    </div>
  )
}
