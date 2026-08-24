import { useCallback, useEffect, useRef, useState } from "react"

import {
  Msg,
  decodeTitle,
  encodeHello,
  encodeInput,
  encodePing,
  encodeResize,
  heartbeatIntervalMs,
  sendFrame,
  wsURL,
  type Size,
} from "@/lib/protocol"

export type Writer = (data: Uint8Array | string) => void

export function useTerminalSession() {
  const wsRef = useRef<WebSocket | null>(null)
  const writeRef = useRef<Writer>(() => {})
  const sizeRef = useRef<Size>({ columns: 80, rows: 24 })
  const helloSentRef = useRef(false)
  const readyRef = useRef(false)
  const [ready, setReady] = useState(false)
  const [disconnected, setDisconnected] = useState(false)

  const registerWriter = useCallback((fn: Writer) => {
    writeRef.current = fn
    if (!readyRef.current) {
      readyRef.current = true
      setReady(true)
    }
  }, [])

  useEffect(() => {
    if (!ready) {
      return
    }

    let cancelled = false
    let pingTimer = 0
    helloSentRef.current = false

    const ws = new WebSocket(wsURL())
    ws.binaryType = "arraybuffer"
    wsRef.current = ws

    const sendHello = () => {
      if (ws.readyState !== WebSocket.OPEN || helloSentRef.current) {
        return
      }
      sendFrame(ws, encodeHello(sizeRef.current))
      helloSentRef.current = true
    }

    const sendPing = () => {
      if (ws.readyState !== WebSocket.OPEN || !helloSentRef.current) {
        return
      }
      sendFrame(ws, encodePing())
    }

    const startHeartbeat = () => {
      if (pingTimer !== 0) {
        return
      }
      pingTimer = window.setInterval(sendPing, heartbeatIntervalMs)
    }

    const onVisibility = () => {
      if (document.visibilityState === "visible") {
        sendPing()
      }
    }

    ws.onopen = () => {
      if (cancelled) {
        return
      }
      sendHello()
      startHeartbeat()
    }

    document.addEventListener("visibilitychange", onVisibility)

    ws.onmessage = (event) => {
      if (cancelled) {
        return
      }
      const buf = new Uint8Array(event.data as ArrayBuffer)
      if (buf.length === 0) {
        return
      }
      if (buf[0] === Msg.Output) {
        writeRef.current(buf.subarray(1))
        return
      }
      if (buf[0] === Msg.Title) {
        const title = decodeTitle(buf.subarray(1))
        document.title = title || "web-tty"
      }
    }

    ws.onclose = () => {
      if (cancelled) {
        return
      }
      setDisconnected(true)
    }

    return () => {
      cancelled = true
      window.clearInterval(pingTimer)
      document.removeEventListener("visibilitychange", onVisibility)
      ws.close()
      if (wsRef.current === ws) {
        wsRef.current = null
      }
    }
  }, [ready])

  const sendInput = useCallback((data: string) => {
    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      return
    }
    sendFrame(ws, encodeInput(data))
  }, [])

  const sendResize = useCallback((cols: number, rows: number) => {
    if (cols < 1 || rows < 1) {
      return
    }
    sizeRef.current = { columns: cols, rows }
    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      return
    }
    if (!helloSentRef.current) {
      sendFrame(ws, encodeHello(sizeRef.current))
      helloSentRef.current = true
      return
    }
    sendFrame(ws, encodeResize(sizeRef.current))
  }, [])

  const reconnect = useCallback(() => {
    window.location.reload()
  }, [])

  const closePage = useCallback(() => {
    window.close()
    window.location.replace("about:blank")
  }, [])

  return {
    registerWriter,
    sendInput,
    sendResize,
    disconnected,
    reconnect,
    closePage,
  }
}
