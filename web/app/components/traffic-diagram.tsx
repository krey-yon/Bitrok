/**
 * TrafficDiagram — the hero centerpiece.
 *
 * An animated SVG showing a request flowing from a public URL, through the
 * bitrok relay, to localhost. Replaces the old ASCII-tunnel.
 *
 *   myapp.bitrok.tech  ──▶  bitrok relay  ──▶  localhost:3000
 *
 * Motion (all subtle, paused under prefers-reduced-motion via globals.css):
 *   - dashed connectors travel via the `flow` CSS keyframe (stroke-dashoffset)
 *   - an amber packet rides the full path left→right via SMIL animateMotion
 *   - a soft status dot pulses on the relay
 *
 * Pure SVG + CSS/SMIL, so it renders server-side with zero client JS.
 */

const NODE_Y = 110;
const NODE_H = 64;
const NODE_W = 150;
const NODE_RX = 14;

// node x-positions (left edge)
const AX = 40;
const BX = 305;
const CX = 570;
// connector endpoints (right edge of A → left edge of B, etc.)
const A_RIGHT = AX + NODE_W;
const B_LEFT = BX;
const B_RIGHT = BX + NODE_W;
const C_LEFT = CX;
// packet travels the full inbound path: URL → relay → localhost
const PACKET_PATH = `M${A_RIGHT} ${NODE_Y} L${C_LEFT} ${NODE_Y}`;
// response travels back: localhost → relay → URL
const REPLY_PATH = `M${C_LEFT} ${NODE_Y} L${A_RIGHT} ${NODE_Y}`;

export function TrafficDiagram({ className = "" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 760 240"
      className={className}
      role="img"
      aria-label="Animated diagram: a request flows from myapp.bitrok.tech through the bitrok relay to localhost:3000."
    >
      {/* connectors — animated dashed lines */}
      <line
        x1={A_RIGHT}
        y1={NODE_Y}
        x2={B_LEFT}
        y2={NODE_Y}
        stroke="var(--accent)"
        strokeWidth={2}
        strokeDasharray="6 6"
        strokeLinecap="round"
        className="animate-flow"
        opacity={0.55}
      />
      <line
        x1={B_RIGHT}
        y1={NODE_Y}
        x2={C_LEFT}
        y2={NODE_Y}
        stroke="var(--accent)"
        strokeWidth={2}
        strokeDasharray="6 6"
        strokeLinecap="round"
        className="animate-flow"
        opacity={0.55}
      />

      {/* arrowheads at each hop */}
      <Arrow x={B_LEFT - 4} y={NODE_Y} />
      <Arrow x={C_LEFT - 4} y={NODE_Y} />

      {/* nodes */}
      <Node x={AX} label="myapp.bitrok.tech" sub="public url">
        <GlobeIcon />
      </Node>
      <Node x={BX} label="bitrok relay" sub="your server">
        <RelayIcon />
        {/* live status pulse on the relay */}
        <circle
          cx={BX + NODE_W - 16}
          cy={NODE_Y - NODE_H / 2 + 14}
          r={3.5}
          fill="var(--success)"
          className="animate-pulse-dot"
        />
      </Node>
      <Node x={CX} label="localhost:3000" sub="your machine">
        <TerminalIcon />
      </Node>

      {/* the traveling packet — rides the full inbound path */}
      <circle r={5} fill="var(--accent)">
        <animateMotion
          dur="2.6s"
          repeatCount="indefinite"
          path={PACKET_PATH}
          calcMode="linear"
        />
      </circle>

      {/* a faint ghost trail under the packet for depth */}
      <circle r={9} fill="var(--accent)" opacity={0.18}>
        <animateMotion
          dur="2.6s"
          repeatCount="indefinite"
          path={PACKET_PATH}
          calcMode="linear"
        />
      </circle>

      {/* the response packet — rides back localhost → URL, lighter + delayed */}
      <circle r={4} fill="var(--accent-light)" opacity={0.85}>
        <animateMotion
          dur="2.6s"
          begin="1.3s"
          repeatCount="indefinite"
          path={REPLY_PATH}
          calcMode="linear"
        />
      </circle>
    </svg>
  );
}

/* ── node ──────────────────────────────────────────────────────────────── */

function Node({
  x,
  label,
  sub,
  children,
}: {
  x: number;
  label: string;
  sub: string;
  children: React.ReactNode;
}) {
  const cx = x + NODE_W / 2;
  return (
    <g>
      <rect
        x={x}
        y={NODE_Y - NODE_H / 2}
        width={NODE_W}
        height={NODE_H}
        rx={NODE_RX}
        fill="var(--card)"
        stroke="var(--border)"
        strokeWidth={1}
      />
      {/* icon centered in the upper area of the node */}
      <g transform={`translate(${cx - 12}, ${NODE_Y - 18})`}>{children}</g>
      {/* labels */}
      <text
        x={cx}
        y={NODE_Y + 10}
        textAnchor="middle"
        fontFamily="var(--font-mono)"
        fontSize={11}
        fill="var(--foreground)"
      >
        {label}
      </text>
      <text
        x={cx}
        y={NODE_Y + 24}
        textAnchor="middle"
        fontFamily="var(--font-mono)"
        fontSize={9}
        fill="var(--muted-foreground)"
      >
        {sub}
      </text>
    </g>
  );
}

/* ── arrowhead ─────────────────────────────────────────────────────────── */

function Arrow({ x, y }: { x: number; y: number }) {
  return (
    <path
      d={`M${x} ${y - 4} L${x + 5} ${y} L${x} ${y + 4}`}
      fill="none"
      stroke="var(--accent)"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      opacity={0.7}
    />
  );
}

/* ── minimal line icons (24x24 box, amber) ─────────────────────────────── */

function GlobeIcon() {
  return (
    <g
      stroke="var(--accent)"
      strokeWidth={1.6}
      fill="none"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx={12} cy={12} r={9} />
      <path d="M3 12h18M12 3c3 3 3 15 0 18M12 3c-3 3-3 15 0 18" />
    </g>
  );
}

function RelayIcon() {
  return (
    <g
      stroke="var(--accent)"
      strokeWidth={1.6}
      fill="none"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x={3} y={4} width={18} height={5} rx={1.5} />
      <rect x={3} y={15} width={18} height={5} rx={1.5} />
      <path d="M7 6.5h.01M7 17.5h.01" />
    </g>
  );
}

function TerminalIcon() {
  return (
    <g
      stroke="var(--accent)"
      strokeWidth={1.6}
      fill="none"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x={3} y={4} width={18} height={16} rx={2} />
      <path d="M7 9l3 3-3 3M13 15h4" />
    </g>
  );
}
