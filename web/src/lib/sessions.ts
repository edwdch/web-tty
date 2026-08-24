export type SessionInfo = {
  id: string
  pid: number
  cwd?: string
  createdAt: string
  clients: number
}

// sessionStorage is per-tab and survives refresh (unlike localStorage, which is
// shared). A new tab starts empty. Duplicating a tab copies the value.
const lastKey = "web-tty.lastSessionId"

export function loadLastSessionId(): string | null {
  try {
    return sessionStorage.getItem(lastKey)
  } catch {
    return null
  }
}

export function saveLastSessionId(id: string): void {
  try {
    sessionStorage.setItem(lastKey, id)
  } catch {
    // ignore quota / private mode
  }
}

/** This tab's last session, if that PTY is still running. */
export function resumeSessionId(list: SessionInfo[]): string | null {
  const last = loadLastSessionId()
  if (last && list.some((s) => s.id === last)) {
    return last
  }
  return null
}

export async function fetchSessions(): Promise<SessionInfo[]> {
  const res = await fetch("/api/sessions")
  if (!res.ok) {
    throw new Error(`list sessions: ${res.status}`)
  }
  const body = (await res.json()) as { sessions?: SessionInfo[] }
  return body.sessions ?? []
}

export async function deleteSession(id: string): Promise<void> {
  const res = await fetch(`/api/sessions/${encodeURIComponent(id)}`, {
    method: "DELETE",
  })
  if (!res.ok && res.status !== 404) {
    throw new Error(`delete session: ${res.status}`)
  }
}

export function pickDefaultSession(list: SessionInfo[]): string | null {
  return resumeSessionId(list) ?? list[0]?.id ?? null
}

export function shortSessionId(id: string): string {
  return id.length > 8 ? id.slice(0, 8) : id
}

export function formatCreatedAt(iso: string): string {
  const t = new Date(iso).getTime()
  if (Number.isNaN(t)) {
    return ""
  }
  const seconds = Math.max(0, Math.floor((Date.now() - t) / 1000))
  if (seconds < 60) {
    return "just now"
  }
  if (seconds < 3600) {
    return `${Math.floor(seconds / 60)}m ago`
  }
  if (seconds < 86400) {
    return `${Math.floor(seconds / 3600)}h ago`
  }
  return `${Math.floor(seconds / 86400)}d ago`
}
