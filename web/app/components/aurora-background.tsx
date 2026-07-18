/**
 * AuroraBackground — animated SVG aurora for the hero/CTA.
 *
 * Three large, blurred gradient blobs in cyan + violet that drift
 * organically via SMIL `<animateTransform>`, plus a flowing ribbon.
 * Pure SVG + SMIL → server-rendered, zero client JS.
 */
export function AuroraBackground({ className = "" }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 1200 800"
      preserveAspectRatio="xMidYMid slice"
      aria-hidden
    >
      <defs>
        <radialGradient id="auroraA">
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0.5" />
          <stop offset="100%" stopColor="var(--accent)" stopOpacity="0" />
        </radialGradient>
        <radialGradient id="auroraB">
          <stop offset="0%" stopColor="var(--secondary)" stopOpacity="0.4" />
          <stop offset="100%" stopColor="var(--secondary)" stopOpacity="0" />
        </radialGradient>
        <radialGradient id="auroraC">
          <stop offset="0%" stopColor="var(--accent-light)" stopOpacity="0.25" />
          <stop offset="100%" stopColor="var(--accent-light)" stopOpacity="0" />
        </radialGradient>
        <linearGradient id="ribbon" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0" />
          <stop offset="50%" stopColor="var(--accent)" stopOpacity="0.4" />
          <stop offset="100%" stopColor="var(--secondary)" stopOpacity="0" />
        </linearGradient>
        <filter id="auroraBlur" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="60" />
        </filter>
      </defs>

      <g filter="url(#auroraBlur)">
        {/* primary cyan blob */}
        <ellipse cx="300" cy="220" rx="320" ry="260" fill="url(#auroraA)">
          <animateTransform
            attributeName="transform"
            type="translate"
            values="0 0; 70 45; -35 25; 25 -35; 0 0"
            keyTimes="0; 0.3; 0.55; 0.8; 1"
            dur="24s"
            repeatCount="indefinite"
          />
        </ellipse>

        {/* violet blob */}
        <ellipse cx="900" cy="180" rx="280" ry="240" fill="url(#auroraB)">
          <animateTransform
            attributeName="transform"
            type="translate"
            values="0 0; -55 35; 45 -25; -25 30; 0 0"
            keyTimes="0; 0.25; 0.5; 0.75; 1"
            dur="30s"
            repeatCount="indefinite"
          />
        </ellipse>

        {/* secondary cyan blob */}
        <ellipse cx="600" cy="450" rx="340" ry="220" fill="url(#auroraC)">
          <animateTransform
            attributeName="transform"
            type="translate"
            values="0 0; 35 -45; -45 15; 15 35; 0 0"
            keyTimes="0; 0.3; 0.6; 0.85; 1"
            dur="28s"
            repeatCount="indefinite"
          />
        </ellipse>
      </g>

      {/* flowing gradient ribbon */}
      <path
        d="M -100 360 Q 300 300 600 360 T 1300 360"
        fill="none"
        stroke="url(#ribbon)"
        strokeWidth="2"
        opacity="0.5"
      >
        <animateTransform
          attributeName="transform"
          type="translate"
          values="0 0; -80 0; 0 0"
          dur="16s"
          repeatCount="indefinite"
        />
      </path>
    </svg>
  );
}
