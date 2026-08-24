export const Cmd = {
  Input: 0x30,
  Resize: 0x31,
  Pause: 0x32,
  Resume: 0x33,
  Ping: 0x34,
  Hello: 0x7b,
} as const

export const heartbeatIntervalMs = 15_000

export const Msg = {
  Output: 0x30,
  Title: 0x31,
  Info: 0x32,
} as const

export type Size = {
  columns: number
  rows: number
}

const encoder = new TextEncoder()
const decoder = new TextDecoder()

export function wsURL(id?: string): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:"
  const url = `${protocol}//${window.location.host}/ws`
  if (!id) {
    return url
  }
  return `${url}?id=${encodeURIComponent(id)}`
}

export function decodeInfo(payload: Uint8Array): { id: string } | null {
  try {
    const value = JSON.parse(decoder.decode(payload)) as { id?: unknown }
    if (typeof value?.id === "string" && value.id !== "") {
      return { id: value.id }
    }
  } catch {
    return null
  }
  return null
}

export function encodeHello(size: Size): Uint8Array {
  return encoder.encode(
    JSON.stringify({ columns: size.columns, rows: size.rows })
  )
}

export function encodeInput(data: string): Uint8Array {
  const body = encoder.encode(data)
  const msg = new Uint8Array(1 + body.length)
  msg[0] = Cmd.Input
  msg.set(body, 1)
  return msg
}

export function encodeResize(size: Size): Uint8Array {
  const body = encoder.encode(
    JSON.stringify({ columns: size.columns, rows: size.rows })
  )
  const msg = new Uint8Array(1 + body.length)
  msg[0] = Cmd.Resize
  msg.set(body, 1)
  return msg
}

export function encodePing(): Uint8Array {
  return new Uint8Array([Cmd.Ping])
}

export function decodeTitle(payload: Uint8Array): string {
  return decoder.decode(payload)
}

export function sendFrame(ws: WebSocket, data: Uint8Array): void {
  const copy = new ArrayBuffer(data.byteLength)
  new Uint8Array(copy).set(data)
  ws.send(copy)
}
