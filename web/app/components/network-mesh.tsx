/**
 * NetworkMesh — animated SVG of the bitrok relay mesh.
 *
 * Nodes (tunnels + relay) connected by edges, with cyan + violet pulses
 * traveling along the links via SMIL `<animateMotion>` and nodes gently
 * breathing with halos. Pure SVG + SMIL → zero client JS.
 */

type Node = { id: string; x: number; y: number; hub?: boolean };
type Edge = { from: string; to: string };

const NODES: Node[] = [
  { id: "relay", x: 600, y: 180, hub: true },
  { id: "a", x: 180, y: 90 },
  { id: "b", x: 180, y: 270 },
  { id: "c", x: 420, y: 70 },
  { id: "d", x: 420, y: 300 },
  { id: "e", x: 820, y: 70 },
  { id: "f", x: 820, y: 300 },
  { id: "g", x: 1040, y: 100 },
  { id: "h", x: 1040, y: 260 },
];

const EDGES: Edge[] = [
  { from: "a", to: "relay" },
  { from: "b", to: "relay" },
  { from: "c", to: "relay" },
  { from: "d", to: "relay" },
  { from: "relay", to: "e" },
  { from: "relay", to: "f" },
  { from: "e", to: "g" },
  { from: "f", to: "h" },
];

const PULSES = [
  { edge: "a->relay", dur: "3.2s", delay: "0s" },
  { edge: "relay->e", dur: "2.8s", delay: "0.4s" },
  { edge: "d->relay", dur: "3.6s", delay: "0.9s" },
  { edge: "relay->f", dur: "3s", delay: "1.3s" },
  { edge: "b->relay", dur: "3.4s", delay: "1.8s" },
];

const byId = Object.fromEntries(NODES.map((n) => [n.id, n]));

export function NetworkMesh({ className = "" }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 1200 360"
      preserveAspectRatio="xMidYMid slice"
      aria-hidden
    >
      {/* edges */}
      {EDGES.map((e, i) => {
        const from = byId[e.from]!;
        const to = byId[e.to]!;
        const d = `M${from.x} ${from.y} L${to.x} ${to.y}`;
        return (
          <path
            key={`e${i}`}
            id={`mesh-edge-${i}`}
            d={d}
            fill="none"
            stroke="var(--border)"
            strokeWidth={1}
          />
        );
      })}

      {/* pulses riding select edges */}
      {PULSES.map((p, i) => {
        const [fromId, toId] = p.edge.split("->");
        const from = byId[fromId]!;
        const to = byId[toId]!;
        const d = `M${from.x} ${from.y} L${to.x} ${to.y}`;
        return (
          <g key={`p${i}`}>
            {/* glow trail */}
            <circle r={5} fill="var(--accent)" opacity={0.2}>
              <animateMotion
                dur={p.dur}
                begin={p.delay}
                repeatCount="indefinite"
                path={d}
                calcMode="linear"
              />
            </circle>
            {/* core pulse */}
            <circle r={2.5} fill="var(--accent)">
              <animateMotion
                dur={p.dur}
                begin={p.delay}
                repeatCount="indefinite"
                path={d}
                calcMode="linear"
              />
            </circle>
          </g>
        );
      })}

      {/* nodes with halos */}
      {NODES.map((n) => (
        <g key={n.id}>
          <circle
            cx={n.x}
            cy={n.y}
            r={n.hub ? 12 : 7}
            fill={n.hub ? "var(--accent)" : "var(--secondary)"}
            opacity={n.hub ? 0.15 : 0.1}
          >
            <animate
              attributeName="r"
              values={n.hub ? "12;16;12" : "7;10;7"}
              dur="3.5s"
              repeatCount="indefinite"
            />
          </circle>
          <circle
            cx={n.x}
            cy={n.y}
            r={n.hub ? 5 : 3}
            fill={n.hub ? "var(--accent)" : "var(--muted-foreground)"}
            stroke="var(--background)"
            strokeWidth={1.5}
          />
        </g>
      ))}
    </svg>
  );
}
