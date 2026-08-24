export const defaultDocumentTitle = "web-tty"

export function applyDocumentTitle(title: string): void {
  const next = title.trim()
  document.title = next || defaultDocumentTitle
}
