import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"

type Props = {
  open: boolean
  onReconnect: () => void
  onClosePage: () => void
}

export function DisconnectDialog({ open, onReconnect, onClosePage }: Props) {
  return (
    <AlertDialog open={open}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Disconnected</AlertDialogTitle>
          <AlertDialogDescription>
            The connection was lost. The session may still be running in the
            background. Reconnect reloads this page; if this tab's session is
            still running it attaches automatically.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <Button type="button" variant="outline" onClick={onClosePage}>
            Close page
          </Button>
          <Button type="button" onClick={onReconnect}>
            Reconnect
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
