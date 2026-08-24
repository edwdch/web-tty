import { Renderer, WTerm } from "@wterm/dom"

type Arms = readonly [u: number, d: number, l: number, r: number]

const L = 1
const H = 2
const D = 3

const BOX_ARMS: Array<Arms | null> = Array.from({ length: 128 }, () => null)

function setBox(cp: number, u: number, d: number, l: number, r: number): void {
  BOX_ARMS[cp - 0x2500] = [u, d, l, r]
}

setBox(0x2500, 0, 0, L, L)
setBox(0x2501, 0, 0, H, H)
setBox(0x2502, L, L, 0, 0)
setBox(0x2503, H, H, 0, 0)
setBox(0x2504, 0, 0, L, L)
setBox(0x2505, 0, 0, H, H)
setBox(0x2506, L, L, 0, 0)
setBox(0x2507, H, H, 0, 0)
setBox(0x2508, 0, 0, L, L)
setBox(0x2509, 0, 0, H, H)
setBox(0x250a, L, L, 0, 0)
setBox(0x250b, H, H, 0, 0)
setBox(0x250c, 0, L, 0, L)
setBox(0x250d, 0, L, 0, H)
setBox(0x250e, 0, H, 0, L)
setBox(0x250f, 0, H, 0, H)
setBox(0x2510, 0, L, L, 0)
setBox(0x2511, 0, L, H, 0)
setBox(0x2512, 0, H, L, 0)
setBox(0x2513, 0, H, H, 0)
setBox(0x2514, L, 0, 0, L)
setBox(0x2515, L, 0, 0, H)
setBox(0x2516, H, 0, 0, L)
setBox(0x2517, H, 0, 0, H)
setBox(0x2518, L, 0, L, 0)
setBox(0x2519, L, 0, H, 0)
setBox(0x251a, H, 0, L, 0)
setBox(0x251b, H, 0, H, 0)
setBox(0x251c, L, L, 0, L)
setBox(0x251d, L, L, 0, H)
setBox(0x251e, H, L, 0, L)
setBox(0x251f, L, H, 0, L)
setBox(0x2520, H, H, 0, L)
setBox(0x2521, H, L, 0, H)
setBox(0x2522, L, H, 0, H)
setBox(0x2523, H, H, 0, H)
setBox(0x2524, L, L, L, 0)
setBox(0x2525, L, L, H, 0)
setBox(0x2526, H, L, L, 0)
setBox(0x2527, L, H, L, 0)
setBox(0x2528, H, H, L, 0)
setBox(0x2529, H, L, H, 0)
setBox(0x252a, L, H, H, 0)
setBox(0x252b, H, H, H, 0)
setBox(0x252c, 0, L, L, L)
setBox(0x252d, 0, L, H, L)
setBox(0x252e, 0, L, L, H)
setBox(0x252f, 0, L, H, H)
setBox(0x2530, 0, H, L, L)
setBox(0x2531, 0, H, H, L)
setBox(0x2532, 0, H, L, H)
setBox(0x2533, 0, H, H, H)
setBox(0x2534, L, 0, L, L)
setBox(0x2535, L, 0, H, L)
setBox(0x2536, L, 0, L, H)
setBox(0x2537, L, 0, H, H)
setBox(0x2538, H, 0, L, L)
setBox(0x2539, H, 0, H, L)
setBox(0x253a, H, 0, L, H)
setBox(0x253b, H, 0, H, H)
setBox(0x253c, L, L, L, L)
setBox(0x253d, L, L, H, L)
setBox(0x253e, L, L, L, H)
setBox(0x253f, L, L, H, H)
setBox(0x2540, H, L, L, L)
setBox(0x2541, L, H, L, L)
setBox(0x2542, H, H, L, L)
setBox(0x2543, H, L, H, L)
setBox(0x2544, H, L, L, H)
setBox(0x2545, L, H, H, L)
setBox(0x2546, L, H, L, H)
setBox(0x2547, H, L, H, H)
setBox(0x2548, L, H, H, H)
setBox(0x2549, H, H, H, L)
setBox(0x254a, H, H, L, H)
setBox(0x254b, H, H, H, H)
setBox(0x254c, 0, 0, L, L)
setBox(0x254d, 0, 0, H, H)
setBox(0x254e, L, L, 0, 0)
setBox(0x254f, H, H, 0, 0)
setBox(0x2550, 0, 0, D, D)
setBox(0x2551, D, D, 0, 0)
setBox(0x2552, 0, L, 0, D)
setBox(0x2553, 0, D, 0, L)
setBox(0x2554, 0, D, 0, D)
setBox(0x2555, 0, L, D, 0)
setBox(0x2556, 0, D, L, 0)
setBox(0x2557, 0, D, D, 0)
setBox(0x2558, L, 0, 0, D)
setBox(0x2559, D, 0, 0, L)
setBox(0x255a, D, 0, 0, D)
setBox(0x255b, L, 0, D, 0)
setBox(0x255c, D, 0, L, 0)
setBox(0x255d, D, 0, D, 0)
setBox(0x255e, L, L, 0, D)
setBox(0x255f, D, D, 0, L)
setBox(0x2560, D, D, 0, D)
setBox(0x2561, L, L, D, 0)
setBox(0x2562, D, D, L, 0)
setBox(0x2563, D, D, D, 0)
setBox(0x2564, 0, L, D, D)
setBox(0x2565, 0, D, L, L)
setBox(0x2566, 0, D, D, D)
setBox(0x2567, L, 0, D, D)
setBox(0x2568, D, 0, L, L)
setBox(0x2569, D, 0, D, D)
setBox(0x256a, L, L, D, D)
setBox(0x256b, D, D, L, L)
setBox(0x256c, D, D, D, D)
setBox(0x256d, 0, L, 0, L)
setBox(0x256e, 0, L, L, 0)
setBox(0x256f, L, 0, L, 0)
setBox(0x2570, L, 0, 0, L)
setBox(0x2574, 0, 0, L, 0)
setBox(0x2575, L, 0, 0, 0)
setBox(0x2576, 0, 0, 0, L)
setBox(0x2577, 0, L, 0, 0)
setBox(0x2578, 0, 0, H, 0)
setBox(0x2579, H, 0, 0, 0)
setBox(0x257a, 0, 0, 0, H)
setBox(0x257b, 0, H, 0, 0)
setBox(0x257c, 0, 0, L, H)
setBox(0x257d, L, H, 0, 0)
setBox(0x257e, 0, 0, H, L)
setBox(0x257f, H, L, 0, 0)

function isBoxCp(cp: number): boolean {
  return cp >= 0x2500 && cp <= 0x257f
}

function isBrailleCp(cp: number): boolean {
  return cp >= 0x2800 && cp <= 0x28ff
}

function armThickness(weight: number): string {
  if (weight >= 3) {
    return "2px"
  }
  if (weight === 2) {
    return "2px"
  }
  return "1px"
}

function armLayer(
  fg: string,
  x: string,
  y: string,
  w: string,
  h: string,
): string {
  return `linear-gradient(${fg},${fg}) ${x} ${y} / ${w} ${h} no-repeat`
}

function getBoxBackground(cp: number, fg: string, bg: string): string | null {
  const arms = BOX_ARMS[cp - 0x2500]
  if (!arms) {
    return null
  }
  const [u, d, l, r] = arms
  const layers: string[] = []
  if (u && d && !l && !r && u === d) {
    layers.push(armLayer(fg, "50%", "50%", armThickness(u), "100%"))
  } else if (l && r && !u && !d && l === r) {
    layers.push(armLayer(fg, "50%", "50%", "100%", armThickness(l)))
  } else {
    if (u) {
      layers.push(armLayer(fg, "50%", "0", armThickness(u), "50%"))
    }
    if (d) {
      layers.push(armLayer(fg, "50%", "100%", armThickness(d), "50%"))
    }
    if (l) {
      layers.push(armLayer(fg, "0", "50%", "50%", armThickness(l)))
    }
    if (r) {
      layers.push(armLayer(fg, "100%", "50%", "50%", armThickness(r)))
    }
  }
  layers.push(bg)
  return layers.join(",")
}

function spanColors(el: HTMLElement): { fg: string; bg: string } {
  return {
    fg: el.style.color || "var(--term-fg)",
    bg: el.style.background || el.style.backgroundColor || "var(--term-bg)",
  }
}

function setSpanCols(el: HTMLElement, cols: number): void {
  el.style.setProperty("--term-span-cols", String(cols))
}

function makeLockedCell(
  template: HTMLElement,
  text: string,
  className?: string,
): HTMLElement {
  const el = template.cloneNode(false) as HTMLElement
  if (className) {
    el.classList.add(className)
  }
  el.textContent = text
  setSpanCols(el, 1)
  return el
}

function makeBoxCell(template: HTMLElement, ch: string, cp: number): HTMLElement {
  const el = makeLockedCell(template, ch, "term-box")
  const { fg, bg } = spanColors(template)
  const drawn = getBoxBackground(cp, fg, bg)
  if (drawn) {
    el.style.background = drawn
    el.style.color = "transparent"
  }
  return el
}

function alignSpan(span: HTMLElement): number {
  if (span.classList.contains("term-wide")) {
    setSpanCols(span, 2)
    return 2
  }
  if (span.classList.contains("term-block") || span.classList.contains("term-box")) {
    setSpanCols(span, 1)
    return 1
  }

  const text = span.textContent ?? ""
  if (!text) {
    setSpanCols(span, 1)
    return 1
  }

  let needsSplit = false
  for (const ch of text) {
    const cp = ch.codePointAt(0) ?? 0
    if (isBoxCp(cp) || isBrailleCp(cp)) {
      needsSplit = true
      break
    }
  }

  if (!needsSplit) {
    const n = [...text].length
    setSpanCols(span, n)
    return n
  }

  const frag = document.createDocumentFragment()
  let buf = ""
  let total = 0

  const flushText = () => {
    if (!buf) {
      return
    }
    const n = [...buf].length
    const el = span.cloneNode(false) as HTMLElement
    el.textContent = buf
    setSpanCols(el, n)
    frag.appendChild(el)
    total += n
    buf = ""
  }

  for (const ch of text) {
    const cp = ch.codePointAt(0) ?? 0
    if (isBoxCp(cp)) {
      flushText()
      frag.appendChild(makeBoxCell(span, ch, cp))
      total += 1
    } else if (isBrailleCp(cp)) {
      flushText()
      frag.appendChild(makeLockedCell(span, ch, "term-box"))
      total += 1
    } else {
      buf += ch
    }
  }
  flushText()
  span.replaceWith(frag)
  return total
}

function alignTerminalRow(rowEl: HTMLElement): void {
  for (const node of [...rowEl.children]) {
    if (!(node instanceof HTMLElement)) {
      continue
    }
    if (node.classList.contains("term-link")) {
      for (const child of [...node.children]) {
        if (child instanceof HTMLElement) {
          alignSpan(child)
        }
      }
      continue
    }
    if (node.tagName === "SPAN") {
      alignSpan(node)
    }
  }
}

type RendererPrivate = {
  _buildRowContent: (
    rowEl: HTMLElement,
    getCell: (col: number) => unknown,
    lineLen: number,
    cursorCol: number,
    rowIndex: number,
  ) => void
}

type WTermPrivate = {
  element: HTMLElement
  cols: number
  _charWidth: number
  _measureCharSize: () => { charWidth: number; rowHeight: number } | null
  resize: (cols: number, rows: number) => void
}

function applyGridMetrics(term: WTermPrivate): void {
  term.element.style.setProperty("--term-cols", String(term.cols))
  if (term._charWidth > 0) {
    term.element.style.setProperty("--term-char-width", `${term._charWidth}px`)
  }
}

let patched = false

export function patchWtermGrid(): void {
  if (patched) {
    return
  }
  patched = true

  const rendererProto = Renderer.prototype as unknown as RendererPrivate
  const originalBuild = rendererProto._buildRowContent
  rendererProto._buildRowContent = function (
    this: unknown,
    rowEl: HTMLElement,
    getCell: (col: number) => unknown,
    lineLen: number,
    cursorCol: number,
    rowIndex: number,
  ) {
    originalBuild.call(this, rowEl, getCell, lineLen, cursorCol, rowIndex)
    alignTerminalRow(rowEl)
  }

  const termProto = WTerm.prototype as unknown as WTermPrivate
  const originalMeasure = termProto._measureCharSize
  termProto._measureCharSize = function (this: WTermPrivate) {
    const result = originalMeasure.call(this)
    applyGridMetrics(this)
    return result
  }

  const originalResize = termProto.resize
  termProto.resize = function (this: WTermPrivate, cols: number, rows: number) {
    originalResize.call(this, cols, rows)
    applyGridMetrics(this)
  }
}

patchWtermGrid()
