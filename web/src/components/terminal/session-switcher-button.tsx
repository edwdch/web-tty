import { createPortal } from "react-dom"
import { LayersIcon } from "lucide-react"

import { Button } from "@/components/ui/button"

type Props = {
  onClick: () => void
}

export function SessionSwitcherButton({ onClick }: Props) {
  return createPortal(
    <Button
      type="button"
      variant="secondary"
      size="icon-sm"
      title="Sessions"
      aria-label="Sessions"
      onClick={onClick}
      className="fixed top-3 right-3 z-40 border-border/60 bg-background/40 opacity-50 shadow-sm ring-1 ring-foreground/15 backdrop-blur-sm hover:bg-background/70 hover:opacity-100"
    >
      <LayersIcon />
    </Button>,
    document.body
  )
}
