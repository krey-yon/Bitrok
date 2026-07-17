# bitrok CLI UI Overhaul — A pi.dev-Inspired Plan

> **Goal:** Transform the bitrok CLI from a functional cobra+lipgloss tool into a polished, premium-feeling TUI that rivals pi.dev's UX — while staying true to bitrok's identity (tunnel management, not a coding agent).

---

## Table of Contents

1. [Current State Audit](#1-current-state-audit)
2. [pi.dev Reference Analysis](#2-pidev-reference-analysis)
3. [Design Principles](#3-design-principles)
4. [Phase 1 — Theme System Overhaul](#phase-1--theme-system-overhaul)
5. [Phase 2 — Rendering Engine](#phase-2--rendering-engine)
6. [Phase 3 — Component Library](#phase-3--component-library)
7. [Phase 4 — Onboarding Wizard](#phase-4--onboarding-wizard)
8. [Phase 5 — Boot Sequence 2.0](#phase-5--boot-sequence-20)
9. [Phase 6 — Interactive Command Experiences](#phase-6--interactive-command-experiences)
10. [Phase 7 — Dashboard 2.0](#phase-7--dashboard-20)
11. [Phase 8 — Terminal Protocol Integration](#phase-8--terminal-protocol-integration)
12. [Phase 9 — Keybinding System](#phase-9--keybinding-system)
13. [Phase 10 — Polish & Micro-interactions](#phase-10--polish--micro-interactions)
14. [File Structure](#file-structure)
15. [Implementation Order](#implementation-order)

---

## 1. Current State Audit

### What bitrok has today

| Area | Current State | Quality |
|---|---|---|
| **Colors** | 9 hardcoded lipgloss.Color vars (colors.go) | Basic |
| **Banner** | ASCII box-drawing logo with gradient amber, line-by-line boot animation | Good — keep and refine |
| **Boot sequence** | 4 fake steps, braille spinner → green check (boot.go) | Good — enhance |
| **Spinner** | braille/moon/pulse/bounce frames via bubbles/spinner | Good |
| **Dashboard** | bubbletea TUI with header, URL, stats, logs, pet mascot | Good — needs redesign |
| **Pet** | ASCII fox with idle/happy/alert moods | Unique — keep and elevate |
| **Output kit** | Success/Warn/Error/Info/Hint/Section/KV/DetailCard/Confirm | Good — needs consistency pass |
| **Table** | Aligned table with gradient header, status colors | Good |
| **Commands** | 12 cobra commands — all print text, only `up`/`http` show TUI | Functional |
| **Borders** | Rounded border on cards, dashed dividers in dashboard | Inconsistent |

### What pi.dev has that bitrok doesn't

| Feature | pi.dev | bitrok gap |
|---|---|---|
| **Theme system** | 45+ tokens, JSON-driven, variable refs, schema-validated, auto-detection | Only 9 hardcoded colors |
| **Rendering engine** | Differential line-diff, CSI 2026 sync output, 60fps cap, per-line SGR reset | Raw fmt.Print everywhere |
| **Border language** | DynamicBorder (full-width `─` in theme color) used as section dividers everywhere | Only used inside dashboard |
| **Interactive components** | SelectList (`→`/`  ` prefix, wrap, two-col), SettingsList (cycle/submenu), Editor (kill-ring, autocomplete, paste-markers) | Only Confirm (y/n) and dashboard |
| **Onboarding** | Bordered ASCII logo + 2-step wizard (theme pick, analytics) with dim hint bar | None |
| **Hint bars** | `dim(keys) + muted(" desc")` consistent pattern, reads from keybinding registry | Only `[o] open [q] quit` in dashboard |
| **Keybindings** | Registry, user-configurable, conflict detection, j/k vim-style | Only q/ctrl+c/o in dashboard |
| **Terminal protocols** | Kitty keyboard, bracketed paste, OSC 0 title, OSC 11 theme detection, OSC 9;4 progress, CSI 2026 sync | None beyond bubbletea defaults |
| **256-color fallback** | Nearest-cube/grayscale with hue preservation | None |
| **Scroll indicators** | `─── ↑ N more ` / `─── ↓ N more ` built into editor borders | None |
| **Footer telemetry** | Stats with color thresholds (green→yellow>70%→red>90%) | Stats exist but no thresholding |
| **Markdown rendering** | Full marked parser with tables, code blocks, blockquotes | N/A for bitrok use case |

---

## 2. pi.dev Reference Analysis

### Architecture (packages/tui)

```
Component interface {
  render(width) → string[]
  handleInput?(data)
  invalidate?()
}
TUI extends Container { children: Component[] }
```

**Rendering pipeline** (`doRender()`):
1. First render → emit all lines, no clear
2. Width/height changed → `ESC[2J ESC[H ESC[3J` (clear + scrollback) → full redraw
3. Normal update → find first/last changed line by string compare → move cursor (`ESC[NA`/`ESC[NB`) → `ESC[J` clear-to-end → emit only changed lines

Every frame wrapped in `ESC[?2026h … ESC[?2026l` (synchronized output). 16ms throttle (60fps cap). Per-line `ESC[0m ESC]8;;ESC[07` (SGR + hyperlink reset).

### Theme system (dark.json)

```
{
  "name": "dark",
  "vars": { "cyan": "#00d7ff", "blue": "#5f87ff", ... },
  "colors": {
    "accent": "accent",        // var ref → resolved recursively
    "border": "blue",
    "mdHeading": "#f0c674"      // literal hex also allowed
  }
}
```

45+ tokens. Categories: base palette (cyan/blue/green/red/yellow/text/gray/dimGray), surfaces (selectedBg, userMsgBg, customMsgBg, toolPendingBg, toolSuccessBg, toolErrorBg), markdown (mdHeading, mdLink, mdCode, mdCodeBlock, mdQuote, mdListBullet), syntax highlighting (comment/keyword/function/variable/string/number/type), thinking gradient (6 levels: off→xhigh).

Border tokens: `border` (blue) for dialogs, `borderAccent` (cyan) for announcements, `borderMuted` (dim) for editor — three levels of visual importance.

### Visual identity
- Restraint is the signature. No splash animation on every launch. One ASCII logo block during onboarding.
- DynamicBorder = `theme.fg("border", "─".repeat(width))` — the single most-used component. Visual rhyme across every screen.
- Editor has NO side borders — only top/bottom `─` lines. Breathing room without boxing.
- Hint bars: `dim(keys) + muted(" desc")`, reads from keybinding registry so rebinds reflect immediately.
- Footer: 2-3 dim lines. Color thresholds gamify cost/context awareness.

### Animations
Only ONE continuous animation: 80ms braille spinner. Everything else is "iota of state" — collapsed labels, paste markers, scroll indicators in borders, color swaps via differential rendering. Polished ≠ animated = polished = thoughtful.

---

## 3. Design Principles

Adapting pi.dev's philosophy for bitrok:

1. **Restraint over animation.** One spinner. Borders, spacing, and typography do the heavy lifting. Don't add typing effects, bouncing dots, or rainbow gradients to everything.
2. **Dynamic border as visual rhyme.** Every section transition gets a full-width `─` line in the accent border color. This single component creates cohesion across all screens.
3. **Three border tiers.** `border` (amber, for cards/dialogs), `borderAccent` (amberLight, for announcements/highlights), `borderMuted` (darkGray, for inline editors/input). Reuse everywhere.
4. **Hint bars are cheap polish.** Every interactive screen ends with `  dim(keys) muted(desc) dim(keys) muted(desc)`. Reads from a keybinding registry. Costs 10 lines of code, feels like +1000.
5. **Pet is bitrok's signature.** Pi has no mascot. Bit's fox is a differentiator — keep it, give it more personality, more mood states, more sayings.
6. **Theme-aware everything.** No raw color constants in component code. Everything goes through the theme struct. This unlocks light mode + custom themes later.
7. **Sync output everywhere.** Wrap every multi-line redraw in `ESC[?2026h ... ESC[?2026l`. Zero flicker, zero cost.
8. **Progressive enhancement.** Detect Kitty, OSC support, truecolor support. Gracefully degrade. Never block on a protocol timeout.

---

## Phase 1 — Theme System Overhaul

**Files:** `internal/ui/theme.go`, `internal/ui/colors.go` (refactor)

### 1.1 Theme struct

Replace the 9 hardcoded `lipgloss.Color` vars with a structured theme:

```go
type Theme struct {
    Name      string
    // Base palette
    Amber       lipgloss.Color // #fbbf24 — primary accent
    AmberLight  lipgloss.Color // #fcd34d — highlight
    AmberDim    lipgloss.Color // #b45309 — muted accent
    Bg          lipgloss.Color // #0a0a0a — near-black
    BgCard      lipgloss.Color // #171717 — card surface
    BgSelected  lipgloss.Color // #1a1a2a — selected row bg
    BgAlt       lipgloss.Color // #121217 — alt row bg
    White       lipgloss.Color // #ededed — foreground
    LightGray   lipgloss.Color // #a3a3a3 — neutral-400
    Gray        lipgloss.Color // #737373 — muted
    DarkGray    lipgloss.Color // #404040 — borders (borderMuted)
    BorderAccent lipgloss.Color // #fbbf24 — border for dialogs (border)
    BorderMuted  lipgloss.Color // #404040 — border for inputs (borderMuted)
    Green       lipgloss.Color // #22c55e — up/success
    Red         lipgloss.Color // #ef4444 — down/error
    Yellow      lipgloss.Color // #facc15 — warning threshold
    // Semantic
    Success     lipgloss.Color
    Error       lipgloss.Color
    Warning     lipgloss.Color
    // Status pill backgrounds
    UpBg        lipgloss.Color // green-tinted bg for "up" pill
    DownBg      lipgloss.Color // gray-tinted bg for "down" pill
    ErrorBg     lipgloss.Color // red-tinted bg for error pill
    PendingBg   lipgloss.Color // amber-tinted bg for loading pill
    // Pet colors
    PetIdle     lipgloss.Color
    PetHappy    lipgloss.Color
    PetAlert    lipgloss.Color
}

var DefaultTheme = Theme{
    Name:         "dark",
    Amber:        lipgloss.Color("#fbbf24"),
    AmberLight:   lipgloss.Color("#fcd34d"),
    AmberDim:     lipgloss.Color("#b45309"),
    Bg:           lipgloss.Color("#0a0a0a"),
    BgCard:       lipgloss.Color("#171717"),
    BgSelected:   lipgloss.Color("#1a1a2a"),
    BgAlt:        lipgloss.Color("#121217"),
    White:        lipgloss.Color("#ededed"),
    LightGray:    lipgloss.Color("#a3a3a3"),
    Gray:         lipgloss.Color("#737373"),
    DarkGray:     lipgloss.Color("#404040"),
    BorderAccent: lipgloss.Color("#fbbf24"),
    BorderMuted:  lipgloss.Color("#404040"),
    Green:        lipgloss.Color("#22c55e"),
    Red:          lipgloss.Color("#ef4444"),
    Yellow:       lipgloss.Color("#facc15"),
    Success:      lipgloss.Color("#22c55e"),
    Error:        lipgloss.Color("#ef4444"),
    Warning:      lipgloss.Color("#facc15"),
    UpBg:         lipgloss.Color("#1a2a1a"),
    DownBg:       lipgloss.Color("#2a2a2a"),
    ErrorBg:      lipgloss.Color("#2a1a1a"),
    PendingBg:    lipgloss.Color("#2a2a1a"),
    PetIdle:      lipgloss.Color("#fcd34d"),
    PetHappy:     lipgloss.Color("#22c55e"),
    PetAlert:     lipgloss.Color("#ef4444"),
}
```

### 1.2 Theme helper functions

```go
func (t Theme) Fg(color lipgloss.Color, text string) string  // wraps with FG color + ESC[39m
func (t Theme) Bg(color lipgloss.Color, text string) string  // wraps with BG color + ESC[49m
func (t Theme) Bold(color lipgloss.Color, text string) string
func (t Theme) Dim(color lipgloss.Color, text string) string
func (t Theme) Italic(color lipgloss.Color, text string) string
func (t Theme) BorderLine(width int) string  // ─×width in BorderMuted color
func (t Theme) AccentBorder(width int) string  // ─×width in BorderAccent color
```

### 1.3 Theme auto-detection (dark/light)

```go
func DetectTheme() string {
    // 1. Check COLORFGBG env: "fg;bg" → bg luminance ≥ 0.5 → "light"
    // 2. Query OSC 11 (ESC]11;?ESC\) → parse RGB response → luminance
    // 3. Fallback: "dark"
}
```

### 1.4 256-color fallback

```go
func nearest256(r, g, b uint8) lipgloss.Color {
    // 6×6×6 color cube (indices 16-51) + 24 grayscale (232-255)
    // Weighted Euclidean distance, preserve hue
}
func resolveColor(hex string, supportsTruecolor bool) lipgloss.Color {
    if supportsTruecolor { return lipgloss.Color(hex) }
    r,g,b := parseHex(hex)
    return nearest256(r,g,b)
}
```

### 1.5 Migration

- Keep `colors.go` as a thin compatibility layer: `var Amber = DefaultTheme.Amber` etc. — so existing code compiles during migration. Gradually replace direct color refs with `theme.Fg(t.Amber, ...)`.

---

## Phase 2 — Rendering Engine

**Files:** `internal/ui/render.go`

### 2.1 Sync output wrapper

```go
const SyncStart = "\x1b[?2026h"
const SyncEnd   = "\x1b[?2026l"

func SyncPrint(lines []string) {
    fmt.Print(SyncStart)
    for _, line := range lines {
        fmt.Println(line + "\x1b[0m") // SGR reset per line
    }
    fmt.Print(SyncEnd)
}
```

### 2.2 Line-diff renderer

```go
type Renderer struct {
    prevLines []string
}

func (r *Renderer) Render(newLines []string) {
    // 1. If first render → print all
    // 2. Find first changed line index (string compare)
    // 3. Find last changed line index
    // 4. Move cursor to firstChanged: ESC[firstChanged - len(prevLines)]A (if negative = up)
    // 5. ESC[J (clear to end)
    // 6. Print changed range
    // 7. Update prevLines
}
```

### 2.3 Clear and redraw utilities

```go
func ClearScreen()        // ESC[2J ESC[H ESC[3J (clear + scrollback)
func ClearLine()          // ESC[2K
func MoveTo(row, col int) // ESC[row;colH
func MoveUp(n int)        // ESC[nA
func MoveDown(n int)      // ESC[nB
func HideCursor()         // ESC[?25l
func ShowCursor()         // ESC[?25h
func SaveCursor()         // ESC7
func RestoreCursor()      // ESC8
```

### 2.4 Where to use

- **Boot sequence** (boot.go): replace manual `clearLines()` with the diff renderer — eliminates the current janky cursor manipulation.
- **Dashboard** (dashboard.go): bubbletea already handles this, but wrap manual output in sync blocks.
- **All command output**: when a command produces multi-line output (list, inspect, config get), wrap in sync output so terminal scrollback stays clean.

---

## Phase 3 — Component Library

**Files:** `internal/ui/components/` (new directory)

Build reusable Go components that mirror pi.dev's TUI patterns, adapted for bubbletea.

### 3.1 DynamicBorder

```go
// internal/ui/components/border.go

type BorderModel struct {
    width int
    color lipgloss.Color  // theme.BorderMuted by default
    char  rune            // '─' by default
}

func NewDynamicBorder(width int, color lipgloss.Color) BorderModel
func (m BorderModel) View() string  // strings.Repeat("─", width) in color
```

**Usage:** Every section transition in every screen gets a DynamicBorder. This is the visual rhyme.

### 3.2 SelectList

```go
// internal/ui/components/selectlist.go

type SelectItem struct {
    Label   string
    Subtext string  // dim description (shown if width > 40)
    Value   string
}

type SelectListModel struct {
    Items       []SelectItem
    selected    int
    width       int
    maxVisible  int
    scrollOffset int
    wrapping    bool  // wrap top↔bottom
    done         bool
    cancelled    bool
}

func NewSelectList(items []SelectItem, maxVisible int) SelectListModel
func (m SelectListModel) View() string
// Renders: "→ label   subtext" (selected, accent prefix)
//          "  label   subtext" (unselected, text)
//          "  (N/M)" scroll indicator when more items than visible
// Wraps top↔bottom on arrow up/down

func (m *SelectListModel) Selected() SelectItem
func (m SelectListModel) Done() bool
func (m SelectListModel) Cancelled() bool
```

**Usage:** `bitrok up` with no args → interactive tunnel picker. `bitrok delete` with no args → picker with confirm. `bitrok update` with no args → picker → then edit`.

### 3.3 SettingsList

```go
// internal/ui/components/settingslist.go

type SettingsItem struct {
    ID          string
    Label       string
    Description string  // wrapped to width-4, indented 2 spaces
    CurrentValue string
    Values      []string  // cycle through on Enter/Space
}

type SettingsListModel struct {
    Items      []SettingsItem
    selected   int
    width      int
    maxVisible int
    onChange   func(id, newValue string)
    done       bool
}

func NewSettingsList(items []SettingsItem, maxVisible int, onChange func(id, string)) SettingsListModel
func (m SettingsListModel) View() string
// Renders: "→ Label     value" (selected row, accent prefix + accent cursor)
//          "  Label     value" (unselected)
//          "  Description wrapped..." (indented 2, muted)
// Hint bar: "  Enter/Space to change · Esc to cancel"
```

**Usage:** `bitrok config edit` (new) → interactive settings editor replacing `config set key value`. Also used in onboarding wizard step 2.

### 3.4 ProgressBar

```go
// internal/ui/components/progress.go

type ProgressModel struct {
    current int
    total   int
    width   int
    label   string
}

// Renders: "  ●━━━━━━━━○──────  12/50 requests"
// Filled portion in Green, empty in DarkGray, handle character: filled ●, empty ○
// For indeterminate: "  ◌━━━━━━━━━━━━━  …" animated dash cycle
```

**Usage:** During `bitrok up` boot sequence, replace the fake spinner with a real progress bar if we can track actual connection steps. Also useful for `bitrok update --self` (binary update).

### 3.5 SpinLoader (BorderedLoader)

```go
// internal/ui/components/loader.go

type LoaderModel struct {
    spinner   spinner.Model
    message   string
    width     int
    cancelHint bool  // show "Esc to cancel" hint
    cancelled  bool
}

func NewBorderedLoader(msg string, width int) LoaderModel
func (m LoaderModel) View() string
// Renders:
// ─────────────────────────────────────────  (DynamicBorder)
//                                           (spacer)
//   ⠋ Loading local registry...             (spinner + message)
//                                           (spacer)
//   Esc to cancel                            (dim hint, if cancelHint)
// ─────────────────────────────────────────  (DynamicBorder)
```

**Usage:** Replace the current boot.go BootSequence's raw spinner+check with bordered loaders. Each step shows a bordered loader, then transitions to a bordered success card.

### 3.6 StatusPill

```go
// internal/ui/components/pill.go

func StatusPill(text string, color lipgloss.Color, bg lipgloss.Color) string
// Renders: " ● LIVE " with fg=color, bg=bgColor, bold
func UpPill() string       // green pill: " ● UP "
func DownPill() string      // gray pill: " ○ DOWN "
func ErrorPill(msg string)  // red pill
func PendingPill(msg string) // amber pill
```

**Usage:** `bitrok status`, `bitrok list` table, dashboard header, inspect output.

### 3.7 HintBar

```go
// internal/ui/components/hints.go

type Hint struct {
    Key  string  // "↑↓" or "q"
    Desc string  // "navigate" or "quit"
}

func RenderHints(hints []Hint) string
// Renders: "  dim(↑↓) muted( navigate )  dim(⏎) muted( confirm )  dim(⎋) muted( cancel )"
// macOS-aware: Alt→Option
```

**Usage:** Every interactive screen. Every dashboard view. Every onboarding step. This is the #1 cheap polish upgrade.

---

## Phase 4 — Onboarding Wizard

**Files:** `internal/ui/onboarding.go`, `internal/cli/onboard.go`

### 4.1 Trigger condition

When `config.Load()` returns no token AND no `BITROK_TOKEN` env var AND stdin is a TTY → show onboarding wizard instead of erroring.

### 4.2 Layout (pi.dev-inspired)

```
────────────────────────────────────────────────────────────  (DynamicBorder, BorderAccent)
                                                              (spacer)

██████╗ ██╗████████╗██████╗  ██████╗ ██╗  ██╗                  (ASCII logo, GradientAmber)
██╔══██╗██║╚══██╔══╝██╔══██╗██╔═══██╗██║ ██╔╝
██████╔╝██║   ██║   ██████╔╝██║   ██║█████╔╝
██╔══██╗██║   ██║   ██╔══██╗██║   ██║██╔═██╗
██████╔╝██║   ██║   ██╔  ██║╚██████╔╝██║  ██╗
╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝

  Welcome to bitrok — deterministic tunnels, zero bullshit.   (bold, AmberLight)
                                                              (spacer)
  Let's get you set up. It takes 30 seconds.                   (Gray)
                                                              (spacer)
────────────────────────────────────────────────────────────  (DynamicBorder)
 Lund ved 1: Choose your serer
────────────────────────────────────────────────────────────

  → bitrok.tech  (default, free)                               (accent selected)
    Self-hosted  (your own server)                             (text)
    Custom URL                                                (text)

  Enter your server URL if you chose 'Custom': ___________     (inline prompt on selection)

────────────────────────────────────────────────────────────

  ↑↓ navigate   ⏎ confirm   ⎋ skip setup                      (HintBar)

────────────────────────────────────────────────────────────
```

### 4.3 Steps

**Step 1 — Server selection**
- SelectList: bitrok.tech (default) / Self-hosted / Custom URL
- If "Self-hosted" or "Custom URL" → inline text input for server URL

**Step 2 — Authentication**
- Two options:
  - **Browser login** (calls existing copyPasteLogin flow but in-TUI with bordered loader)
  - **Paste token** (inline editor masked input — show `••••••••••••` as you type)
- BorderedLoader while validating token against server
- If auth fails → red error card with retry option

**Step 3 — Theme preference** (future: when light theme exists)
- Dark / Light / Auto-detect
- Currently: just note "Dark mode detected" and skip

**Step 4 — Confirmation**
- Bordered success card:
  - Green check
  - "You're all set! Your tunnels will be on bitrok.tech"
  - Next steps: "Run `bitrok http 3000` to start a quick tunnel"
  - Pet wave: Bit fox with `( ^.^ ) editing`

### 4.4 Key handling

- `↑`/`↓` or `j`/`k` — navigate
- `Enter` — confirm/advance step
- `Esc` — skip (saves nothing, shows help text)
- `Ctrl+C` — abort

### 4.5 Persistence

On completion: save to `~/.config/bitrok/config.json` (or whatever the current config path is). Set `DefaultDomain` and `ServerURL` and `Token`.

---

## Phase 5 — Boot Sequence 2.0

**Files:** `internal/ui/boot.go` (rewrite), `internal/cli/up.go` (update)

### 5.1 Current problems

- Fake timing (knows it's lying, comments say so)
- Manual cursor manipulation is janky on some terminals
- No sync output → minor flicker
- No visual hierarchy — just spinner + check, all flat

### 5.2 New boot sequence

Two-phase: **banner animation** → **stepped init with real state where possible**.

**Phase A — Banner reveal (already good, refine):**

```
  (line-by-line gradient logo reveal, 55ms per line — keep this)
  (tagline appears with a subtle fade-in: dim→full over 200ms)
```

**Phase B — Init steps using BorderedLoader:**

Each step shows a bordered box while running, then collapses to a one-line check:

```
────────────────────────────────────────────────  (DynamicBorder, BorderMuted)
  ⠋ Loading local registry...                      (spinner + message, amber)
────────────────────────────────────────────────
```

When step completes, box is replaced by:
```
  ✓ Local registry loaded (3 tunnels found)       (green check, White)
```

### 5.3 Steps with real callbacks

Where possible, make boot steps real:

```go
type BootStep struct {
    Label  string
    Action func() (string, error)  // returns status detail, or error
}

// If Action is nil → fake timing (current behavior, fallback)
// If Action is set → run it, show spinner while running, check on success
```

Steps for `bitrok up`:
1. "Loading local registry" → `config.LoadRegistry()` (real, instant)
2. "Resolving tunnel {name}" → `reg.FindByName(name)` + validate (real)
3. "Authenticating with server" → `client.NewAPIClient()` + validate token (real HTTP)
4. "Establishing WebSocket connection" → `session.Start()` partial (real — connecting)
5. "Waking Bit — {greeting}" → pet animation (fake, delightful)

### 5.4 Error in boot

If a real step fails:
```
────────────────────────────────────────────────
  ✗ Authenticating with server                   (red ✗)
    Server returned 401: invalid token            (red, indented)
    Run 'bitrok login' to authenticate.           (AmberLight hint)
────────────────────────────────────────────────
```

### 5.5 Sync output

Wrap the entire boot sequence redraw in `ESC[?2026h ... ESC[?2026l`.

---

## Phase 6 — Interactive Command Experiences

Currently every command just prints text. Upgrade the key commands to have interactive fallbacks.

### 6.1 `bitrok up` (no args) → Tunnel picker

```
────────────────────────────────────────────────
  ⠋ Loading tunnels...                           (BorderedLoader while fetching)
────────────────────────────────────────────────

(then)

────────────────────────────────────────────────
  Select a tunnel to start                       (title, AmberLight bold)
────────────────────────────────────────────────

  → my-api       api.myapp.bitrok.tech  :3000    ● up    (accent selected, status pill)
    staging      staging.bitrok.tech     :8080    ○ down
    webhook      hooks.bitrok.tech       :4000    ○ down
    preview      preview.bitrok.tech     :5173    ○ down

  (4/4)

────────────────────────────────────────────────

  ↑↓ navigate   ⏎ start tunnel   ⎋ cancel        (HintBar)
────────────────────────────────────────────────
```

If there are active tunnels, show a warning pill. If zero tunnels, show empty state with CTA:
```
  No tunnels yet. Run 'bitrok create' to make one.
  Or: 'bitrok http 3000' for a quick ad-hoc tunnel.
```

### 6.2 `bitrok delete` (no args) → Picker + confirm

```
  Select a tunnel to delete

  → my-api       api.myapp.bitrok.tech  :3000
    staging      staging.bitrok.tech     :8080

  ⏎ select   ⎋ cancel
  
(selected)

  Delete tunnel 'my-api'?                          (Amber bold)
  This will permanently remove the tunnel and stop any active session.

  → ⏎ Yes, delete it    ⎋ No, keep it             (inline confirm SelectList)

  ⏎ confirm   ⎋ cancel
```

### 6.3 `bitrok list` → Enhanced table

Current: plain aligned table. Upgrade to:

```
────────────────────────────────────────────────

  YOUR TUNNELS  (4 total, 1 active)                (GradientAmber title + count in Gray)

────────────────────────────────────────────────

  ●  my-api       api.myapp.bitrok.tech  :3000    [● UP]    2m ago
  ○  staging      staging.bitrok.tech     :8080    [○ DOWN] 1h ago
  ○  webhook      hooks.bitrok.tech       :4000    [○ DOWN] 3d ago
  ○  preview      preview.bitrok.tech     :5173    [○ DOWN] 1w ago

────────────────────────────────────────────────

  Run 'bitrok up <name>' to start a tunnel.       (HintBar-style)
```

Changes:
- DynamicBorder top/bottom instead of dashed line
- Status pills `[● UP]` / `[○ DOWN]` instead of inline text
- Relative time column refined
- Footer hint for next action
- `--json` flag unchanged for scripting

### 6.4 `bitrok inspect <name>` → Enhanced detail card

Current: rounded bordered card with KV rows. Upgrade to sectioned layout:

```
────────────────────────────────────────────────

  my-api                                           (title, bold AmberLight)
  api.myapp.bitrok.tech → localhost:3000           (subtitle, Gray)

────────────────────────────────────────────────

  Status     ● UP                                   (StatusPill inline)
  Uptime     12m 34s                               (if active)
  Tunnel ID  550e8400-e29b-41d4-a716-446655440000   (Gray)
  Created    2024-07-06 14:22
  Updated    2024-07-06 14:35

────────────────────────────────────────────────

  ● 172 requests   ↑ 1.2 MB   ↓ 847 KB   p50 45ms   (stats if active)

────────────────────────────────────────────────

  Bit says: "running smooth — no complaints!"      (pet saying, Gray italic)

────────────────────────────────────────────────
```

### 6.5 `bitrok status` → Status overview

Current: one-line summary. Upgrade to mini-dashboard:

```
────────────────────────────────────────────────

  BITROK STATUS

────────────────────────────────────────────────

  ●  1/4 tunnels active                            (big status)

────────────────────────────────────────────────

  ●  my-api       api.myapp.bitrok.tech  :3000     [● UP]
  ○  staging      staging.bitrok.tech     :8080     [○ DOWN]
  ○  webhook      hooks.bitrok.tech       :4000     [○ DOWN]
  ○  preview      preview.bitrok.tech     :5173     [○ DOWN]

────────────────────────────────────────────────

  Bit is watching the wire...                      (pet idle)
```

### 6.6 `bitrok config` → Interactive editor

New subcommand: `bitrok config edit` (in addition to `config get/set/reset`).

Uses SettingsListModel:
```
────────────────────────────────────────────────

  BITROK CONFIG

────────────────────────────────────────────────

  → Server        bitrok.tech
    Token         ••••••••••••••••5f2a
    Domain        bitrok.tech
    Theme         dark

  Enter/Space to change · Esc to cancel

(selected item opens inline editor or submenu)
```

### 6.7 `bitrok login` → In-TUI flow

Current: prints text, opens browser, reads from stdin. Upgrade to full TUI:

```
────────────────────────────────────────────────
  Authenticate with bitrok                         (title)
────────────────────────────────────────────────

  Opening your browser...                          (BorderedLoader)
  If it doesn't open, visit:
  bitrok.tech/dashboard/cli-token                  (AmberLight underline)

────────────────────────────────────────────────

  Paste your token:                                (prompt)
  ••••••••••••••••••••••••••••••••                (mask input, show cursor)

  ⏎ submit   ⎋ cancel

────────────────────────────────────────────────
```

Token verification via bordered loader, then success card with pet.

---

## Phase 7 — Dashboard 2.0

**Files:** `internal/ui/dashboard.go` (redesign)

### 7.1 Layout overhaul

Current dashboard is good but feels cramped. Restructure with clear sections divided by DynamicBorders:

```
╭──────────────────────────────────────────────────────────────────╮
│                                                                  │  ← Rounded border, BorderAccent
│  BITROK  │  ● LIVE            up 12m 34s                        │  ← Header: logo + live pill + uptime
│                                                                  │
│  ──────────────────────────────────────────────────────────────  │  ← DynamicBorder (muted)
│                                                                  │
│  ⠋  https://api.myapp.bitrok.tech → localhost:3000               │  ← URL line with spinner
│                                                                  │
│  ──────────────────────────────────────────────────────────────  │  ← DynamicBorder (muted)
│                                                                  │
│  Requests  172    ↑ Out  1.2MB    ↓ In  847KB    p50 45ms       │  ← Stats row
│  2xx 150   4xx 18    5xx 4                                        │  ← Status breakdown
│                                                                  │
│  ──────────────────────────────────────────────────────────────  │  ← DynamicBorder (muted)
│                                                                  │
│  TRAFFIC  ────────────────────────────────────────────────      │  ← GradientAmber section + fill
│                                                                  │
│  14:22:03  GET    /api/health         200   12ms                  │  ← Log lines
│  14:22:05  POST   /api/webhooks         201   45ms                  │
│  14:22:10  GET    /api/users            200    8ms                  │
│  14:22:12  GET    /static/app.js      304    2ms                  │
│  14:22:15  POST   /api/orders           500  230ms  ← red          │
│  ...                                                             │
│                                                                  │
│  ──────────────────────────────────────────────────────────────  │  ← DynamicBorder (muted)
│                                                                  │
│   /\_/\     Bit  nice one!  (152 reqs)                          │  ← Pet + saying + key hints
│  ( ^.^ )                                                        │
│   > w <    [o] open  [l] logs  [s] stats  [q] quit              │  ← Expanded key hints
│                                                                  │
╰──────────────────────────────────────────────────────────────────╯
```

### 7.2 Specific changes

| Element | Current | New |
|---|---|---|
| Header | Flat logo + pill | Logo + status pill with bg color + uptime in Gray |
| Borders | `─` in DarkGray | Three tiers: BorderAccent (main), BorderMuted (sections) |
| Stats | One flat line | Two lines: counts + status breakdown with color-coded numbers |
| Logs | Plain colored lines | Same, but alternating bg (BgCard/BgAlt) for readability |
| Footer | Pet + 2-line keys | Pet + saying + full HintBar from keybinding registry |
| Border color | Breathing Amber↔AmberDim every 4 ticks | Solid BorderAccent (remove breathing — it's not adding value) |

### 7.3 New key bindings

| Key | Action |
|---|---|
| `q` / `Ctrl+C` | Quit (exit tunnel) |
| `o` | Open public URL in browser |
| `c` | Copy URL to clipboard (OSC 52) |
| `l` | Toggle log view (compact/verbose) |
| `s` | Toggle stats panel (show/hide) |
| `r` | Refresh / clear stats |
| `?` | Show full keybinding help overlay (BorderedLoader-style popup) |
| `↑`/`↓` | Scroll through logs |
| `PgUp`/`PgDn` | Page through logs |
| `Home`/`End` | Jump to top/bottom of logs |
| `f` | Freeze/unfreeze log scroll (follow mode) |

### 7.4 Log rendering improvements

- Alternating row backgrounds (Bg/BgAlt) for readability — like a table
- Status 5xx rows get a subtle ErrorBg tint
- Verbose mode shows: `time method path status latency reqSize respSize userAgent`
- Compact mode shows: `time method path status latency` (current)
- `f` follow mode: auto-scrolls to latest (default on)
- Scroll indicator in border: `─── ↑ N more ───` when scrolled up from bottom

### 7.5 Pet enhancements

More mood states:

```go
const (
    PetIdle     PetMood = iota  // breathing, occasional blink
    PetHappy                     // request landed (2xx)
    PetAlert                     // 5xx error
    PetSleeping                  // no traffic for 60+ seconds (zzz)
    PetThinking                  // connecting / authenticating
    PetExcited                   // traffic spike (>10 reqs in 5 seconds)
)
```

More idle frames (breathing animation):
```go
var petFramesIdle = []string{
    " /\\_/\\   ",
    "( o.o )  ",
    " > ^ <   ",
}
var petFramesBreathe = []string{  // slow breathing (subtle scale)
    " /\\_/\\   ",
    " /\\_/\\   ",
    "( o.o )  ",
    "(  o  )  ",  // eyes slightly wider
    " > ^ <   ",
    " > ^ <   ",
}
```

Sleeping state with Z's:
```go
var petFramesSleeping = []string{
    " /\\_/\\    Z",
    "( -.- )  z",
    " > ~ <   z",
}
```

### 7.6 Connection state indicator

The border color reflects actual session state:
- Connecting → amber pulsing border (subtle, every 800ms)
- Connected (LIVE) → solid green border
- Reconnecting → amber fast pulse (every 300ms) + ⚠ in header
- Error → red border for 3 seconds, then solid border

This replaces the fake "breathing border" with real state — which was the original design intent anyway.

---

## Phase 8 — Terminal Protocol Integration

**Files:** `internal/ui/terminal.go`

### 8.1 Capability detection

```go
type TermCapabilities struct {
    Truecolor   bool
    Kitty       bool
    OSCSupport  bool
    Width       int
    Height      int
    BgColor     string  // for theme detection
}

func DetectCapabilities() TermCapabilities {
    // 1. COLORTERM=truecolor → Truecolor
    // 2. TERM=xterm-256color → 256-color
    // 3. Check TERM_PROGRAM for iTerm2/Kitty
    // 4. Query cell size ESC[16t for Kitty image support
}
```

### 8.2 OSC sequences

```go
func SetWindowTitle(title string)  // ESC]0;titleESC\
func SetProgress(completed int, total int)  // ESC]9;4;0;completed;totalESC\
func SetProgressIndeterminate()    // ESC]9;4;3ESC\
func ClearProgress()               // ESC]9;4;0;0ESC\
func QueryBgColor() (r,g,b int, ok bool)  // ESC]11;?ESC\ → parse response
func CopyToClipboard(text string)  // ESC]52;c;textESC\
```

**Usage:**
- During `bitrok up`: set window title to "bitrok — my-api ● LIVE"
- Boot sequence: set progress indeterminate
- When tunnel connects: set progress to 100%, then clear
- `c` key in dashboard: copy URL to clipboard via OSC 52 (works in SSH too!)

### 8.3 Bracketed paste

```go
func EnableBracketedPaste()   // ESC[?2004h
func DisableBracketedPaste()  // ESC[?2004l
// Paste arrives wrapped in ESC[200~ ... ESC[201~
```

For the login token paste → detect bracketed paste to accept multi-line tokens safely.

### 8.4 Sync output (from Phase 2)

Already covered — `ESC[?2026h ... ESC[?2026l` around all multi-line redraws.

---

## Phase 9 — Keybinding System

**Files:** `internal/ui/keybindings.go`, `internal/config/keybindings.go`

### 9.1 Registry

```go
type Keybinding struct {
    ID    string  // "dashboard.quit", "select.navigate_up"
    Key   string  // "q", "ctrl+c", "up"
    Action string // "quit", "navigate_up"
    Desc  string  // "quit", "navigate up"
}

type KeybindingRegistry struct {
    bindings map[string]Keybinding  // keyed by ID
}

func DefaultKeybindings() *KeybindingRegistry
func (r *KeybindingRegistry) Get(id string) Keybinding
func (r *KeybindingRegistry) Set(id, key string)
func (r *KeybindingRegistry) GetConflicts() []string  // same key → multiple IDs
```

### 9.2 Default bindings

```go
var Defaults = []Keybinding{
    // Global
    {ID: "global.quit", Key: "q", Desc: "quit"},
    {ID: "global.force_quit", Key: "ctrl+c", Desc: "force quit"},
    {ID: "global.cancel", Key: "esc", Desc: "cancel"},

    // Navigation
    {ID: "nav.up", Key: "up", Desc: "up"},
    {ID: "nav.down", Key: "down", Desc: "down"},
    {ID: "nav.up_alt", Key: "k", Desc: "up (vim)"},
    {ID: "nav.down_alt", Key: "j", Desc: "down (vim)"},
    {ID: "nav.confirm", Key: "enter", Desc: "confirm"},
    {ID: "nav.page_up", Key: "pgup", Desc: "page up"},
    {ID: "nav.page_down", Key: "pgdown", Desc: "page down"},
    {ID: "nav.top", Key: "home", Desc: "jump to top"},
    {ID: "nav.bottom", Key: "end", Desc: "jump to bottom"},

    // Dashboard
    {ID: "dashboard.open_url", Key: "o", Desc: "open URL in browser"},
    {ID: "dashboard.copy_url", Key: "c", Desc: "copy URL to clipboard"},
    {ID: "dashboard.toggle_logs", Key: "l", Desc: "toggle log view"},
    {ID: "dashboard.toggle_stats", Key: "s", Desc: "toggle stats"},
    {ID: "dashboard.refresh", Key: "r", Desc: "refresh stats"},
    {ID: "dashboard.freeze", Key: "f", Desc: "freeze/unfreeze logs"},
    {ID: "dashboard.help", Key: "?", Desc: "show all keybindings"},
}
```

### 9.3 User configuration

```json
// ~/.config/bitrok/config.json (existing file, add field)
{
  "keybindings": {
    "dashboard.quit": "q",
    "nav.up": "k"
  }
}
```

### 9.4 Hint bar integration

`HintBar` reads from the registry:

```go
func RenderHints(ids []string, registry *KeybindingRegistry) string {
    var parts []string
    for _, id := range ids {
        kb := registry.Get(id)
        parts = append(parts, theme.Dim(kb.Key) + " " + theme.Muted(kb.Desc))
    }
    return "  " + strings.Join(parts, "   ")
}
```

Now every screen's hints update automatically when the user rebinds.

---

## Phase 10 — Polish & Micro-interactions

### 10.1 Empty states

Every command that can have zero results gets a beautiful empty state:

```
────────────────────────────────────────────────

  No tunnels yet.                                  (Gray, italic)

  Get started:                                    (AmberLight)

  bitrok http 3000        Quick ad-hoc tunnel     (Dim keys + muted desc)
  bitrok create           Register a named tunnel

────────────────────────────────────────────────

  Bit is napping...                               (sleeping pet)
```

### 10.2 Error states

Consistent error rendering across all commands:

```
────────────────────────────────────────────────

  ✗ Failed to create tunnel                        (Red bold)

  Server returned 409: host already in use         (Red)

  Try:                                             (AmberLight)
  → Use a different --host value
  → Run 'bitrok list' to see existing tunnels

────────────────────────────────────────────────

  ⎋ dismiss                                       (HintBar)
```

### 10.3 Success states

```
────────────────────────────────────────────────

  ✓ Tunnel 'my-api' is live                        (Green check, White)

  https://api.myapp.bitrok.tech → localhost:3000  (AmberLight underline)

  [Press 'c' to copy URL]                          (Gray hint, ephemeral)

────────────────────────────────────────────────

  /\_/\    Bit says: "let's go!"                   (happy pet)
```

### 10.4 Config validation feedback

When config is invalid:

```
────────────────────────────────────────────────

  ⚠ Configuration issues found                     (Amber bold)

  → Missing auth token                             (Amber →, text)
  → Server URL is empty
  → Default domain not set (will use 'bitrok.tech')

────────────────────────────────────────────────

  Run 'bitrok login' to fix auth.                  (HintBar-style)
  Run 'bitrok config edit' for full settings.
```

### 10.5 Update experience (`bitrok update --self`)

```
────────────────────────────────────────────────

  Checking for updates...                          (BorderedLoader)

────────────────────────────────────────────────

  Current: v0.1.0                                   (Gray)
  Latest:  v0.2.0                                   (Green bold)

━━━━━━━━━━━━━━━━○──────────────────  Downloading  (ProgressBar)
  45%  2.3MB / 5.1MB

  ⏎ continue in background   ⎋ cancel             (HintBar)
```

### 10.6 `bitrok http <port>` quick start

Currently goes straight to boot. Add a brief "choosing your URL" moment:

```
────────────────────────────────────────────────

  Starting ad-hoc tunnel...                         (title)

  Local port:  3000                                (KV)
  Subdomain:   auto-assigned                       (Gray)
  Domain:      bitrok.tech                         (Gray)

────────────────────────────────────────────────

  ⠋ Reserving subdomain...                         (BorderedLoader)

────────────────────────────────────────────────

  ✓ Reserved: https://a1b2c3.bitrok.tech           (success)

  (boot sequence continues normally)
```

### 10.7 Consistent indentation

Every single output line starts with 2-space indent. No exceptions. This is pi.dev-level consistency. Audit every `fmt.Print` in the codebase and ensure 2-space indent + consistent spacing.

### 10.8 Color threshold for stats

In dashboard, stats get color-coded when they cross thresholds:

| Metric | Good | Warning | Critical |
|---|---|---|---|
| Error rate | <5% green | 5-10% yellow | >10% red |
| p50 latency | <100ms green | 100-500ms yellow | >500ms red |
| Total requests | show as-is | shows amber at >1000 | red at >5000 (rate limit) |

This replaces the static Green/AmberLight/Red coloring in the current dashboard.

---

## File Structure

New and modified files:

```
cli/internal/ui/
├── theme.go           (NEW — Phase 1: Theme struct, helpers, auto-detection)
├── colors.go          (MODIFY — thin compat layer over Theme)
├── terminal.go        (NEW — Phase 8: capability detection, OSC sequences)
├── render.go          (NEW — Phase 2: sync output, line-diff renderer)
├── keybindings.go     (NEW — Phase 9: keybinding registry)
├── banner.go          (MODIFY — Phase 5: refine animation)
├── boot.go            (REWRITE — Phase 5: real callbacks, bordered loaders)
├── spinner.go         (KEEP — minor: tie to theme)
├── gradient.go        (KEEP)
├── pet.go             (MODIFY — Phase 7: more moods, more frames)
├── dashboard.go       (REWRITE — Phase 7: new layout, scroll, keybindings)
├── output.go          (MODIFY — Phase 10: empty/error/success states, consistency)
├── table.go           (MODIFY — Phase 6: status pills, dynamic borders)
├── onboarding.go      (NEW — Phase 4: first-time setup wizard)
└── components/        (NEW directory)
    ├── border.go      (Phase 3: DynamicBorder)
    ├── selectlist.go   (Phase 3: SelectList)
    ├── settingslist.go (Phase 3: SettingsList)
    ├── progress.go     (Phase 3: ProgressBar)
    ├── loader.go       (Phase 3: BorderedLoader)
    ├── pill.go         (Phase 3: StatusPill)
    └── hints.go        (Phase 3: HintBar)

cli/internal/cli/
├── root.go            (MODIFY — Phase 4: trigger onboarding when no config)
├── up.go              (MODIFY — Phase 5+6: boot 2.0, tunnel picker)
├── login.go           (MODIFY — Phase 6: in-TUI login flow)
├── create.go          (MODIFY — Phase 10: success state)
├── delete.go          (MODIFY — Phase 6: picker + confirm)
├── list.go            (MODIFY — Phase 6: enhanced table)
├── inspect.go         (MODIFY — Phase 6: sectioned layout)
├── status.go          (MODIFY — Phase 6: mini-dashboard)
├── config.go          (MODIFY — Phase 6: interactive editor)
├── http.go            (MODIFY — Phase 10: quick-start moment)
├── down.go            (MODIFY — Phase 10: consistent states)
├── logs.go            (MODIFY — Phase 10: nice "not implemented" state)
├── update.go          (MODIFY — Phase 10: update experience)
└── auth.go            (MODIFY — Phase 10: consistent output)

cli/internal/config/
├── config.go          (MODIFY — Phase 9: keybindings field)
```

---

## Implementation Order

Each phase builds on the previous. Phases 1-3 are foundational infrastructure. Phases 4-7 are the user-facing payoffs. Phases 8-10 are polish.

### Week 1: Foundation (Phases 1-3)

| Priority | Task | Effort | Dependencies |
|---|---|---|---|
| P0 | **Phase 1:** Theme struct + helpers | 2h | none |
| P0 | **Phase 2.1-2.2:** SyncPrint + simple renderer | 1h | Phase 1 |
| P0 | **Phase 3.1:** DynamicBorder | 30m | Phase 1 |
| P0 | **Phase 3.7:** HintBar | 30m | Phase 1 |
| P0 | **Phase 3.5:** BorderedLoader | 1h | Phase 3.1 |
| P1 | **Phase 3.2:** SelectList | 2h | Phase 1 |
| P1 | **Phase 3.3:** SettingsList | 2h | Phase 1 |
| P1 | **Phase 3.6:** StatusPill | 30m | Phase 1 |
| P2 | **Phase 3.4:** ProgressBar | 1h | Phase 1 |
| P2 | **Phase 1.3:** Theme auto-detection (OSC 11) | 1h | Phase 1 |
| P2 | **Phase 1.4:** 256-color fallback | 1h | Phase 1 |

### Week 2: Boot + Onboarding (Phases 4-5)

| Priority | Task | Effort | Dependencies |
|---|---|---|---|
| P0 | **Phase 5:** Boot sequence 2.0 rewrite | 3h | Phase 3.5 |
| P0 | **Phase 5.3:** Real boot step callbacks | 1h | Phase 5 |
| P0 | **Phase 4:** Onboarding wizard | 4h | Phase 3.2, 3.5, 3.7 |
| P0 | **Phase 4:** Onboarding trigger in root.go | 30m | Phase 4 |

### Week 3: Command UX (Phase 6)

| Priority | Task | Effort | Dependencies |
|---|---|---|---|
| P0 | **Phase 6.1:** `up` tunnel picker | 2h | Phase 3.2 |
| P0 | **Phase 6.7:** `login` in-TUI flow | 2h | Phase 3.5 |
| P1 | **Phase 6.2:** `delete` picker + confirm | 1h | Phase 3.2 |
| P1 | **Phase 6.3:** `list` enhanced table | 1h | Phase 3.6 |
| P1 | **Phase 6.4:** `inspect` sectioned layout | 1h | Phase 3.6 |
| P1 | **Phase 6.5:** `status` mini-dashboard | 1h | Phase 3.6 |
| P2 | **Phase 6.6:** `config edit` interactive | 2h | Phase 3.3 |

### Week 4: Dashboard 2.0 (Phase 7)

| Priority | Task | Effort | Dependencies |
|---|---|---|---|
| P0 | **Phase 7.1:** Dashboard layout overhaul | 3h | Phase 1, 3.1 |
| P0 | **Phase 7.2:** Stats with color thresholds | 1h | Phase 7.1 |
| P0 | **Phase 7.3:** New key bindings | 1h | Phase 9 |
| P0 | **Phase 7.4:** Log rendering (alt rows, freeze, scroll) | 2h | Phase 7.1 |
| P1 | **Phase 7.5:** Pet enhanced moods | 1h | Phase 7.1 |
| P1 | **Phase 7.6:** Connection state border | 1h | Phase 7.1 |

### Week 5: Terminal + Keybindings + Polish (Phases 8-10)

| Priority | Task | Effort | Dependencies |
|---|---|---|---|
| P0 | **Phase 9:** Keybinding registry | 2h | none |
| P0 | **Phase 8.2:** OSC title + progress + clipboard | 1h | Phase 8.1 |
| P0 | **Phase 8.1:** Capability detection | 1h | none |
| P0 | **Phase 8.3:** Bracketed paste | 30m | Phase 8.1 |
| P1 | **Phase 10.1-10.3:** Empty/error/success states | 2h | Phase 1, 3.1 |
| P1 | **Phase 10.7:** Indentation audit | 1h | none |
| P1 | **Phase 10.8:** Color thresholds | 30m | Phase 1 |
| P2 | **Phase 10.4-10.6:** Config validation, update, http quick-start | 2h | various |

**Total estimated effort: ~45 hours**

---

## Summary: The 10 Highest-Impact Changes

If you only do 10 things, do these — in order:

1. **Theme struct** (Phase 1) — foundation for everything else
2. **DynamicBorder + HintBar** (Phase 3.1, 3.7) — instant visual cohesion across all screens
3. **BorderedLoader** (Phase 3.5) — every loading moment feels intentional
4. **Boot sequence 2.0** (Phase 5) — real callbacks + bordered loaders
5. **Onboarding wizard** (Phase 4) — first impression = last impression
6. **SelectList** (Phase 3.2) — enables interactive `up`, `delete`, `update`
7. **Dashboard layout** (Phase 7.1) — the hero screen, needs to shine
8. **Keybinding registry** (Phase 9) — every hint bar becomes dynamic
9. **OSC sequences** (Phase 8.2) — window title, progress bar, clipboard copy
10. **Consistency audit** (Phase 10.7) — 2-space indent everywhere, no exceptions

These 10 changes will transform bitrok from "functional CLI with some animations" to "polished TUI product that feels premium."