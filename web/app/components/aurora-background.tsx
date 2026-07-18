/**
 * AuroraBackground — animated SVG aurora for the hero/CTA.
 *
 * Three large, blurred gradient blobs that drift organically via SMIL
 * `<animateTransform>`, plus one faint flowing ribbon. Replaces the simpler
 * CSS orbs with a richer, smoother animated gradient field.
 *
 * Pure SVG + SMIL → renders server-side, zero client JS. Paused under
 * prefers-reduced-motion (globals.css nukes all animations).
 *
 * Mount inside a `relative overflow-hidden` parent; this fills it via
 * preserveAspectRatio="xMidYMid slice".
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
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0.55" />
          <stop offset="100%" stopColor="var(--accent)" stopOpacity="0" />
        </radialGradient>
        <radialGradient id="auroraB">
          <stop offset="0%" stopColor="var(--accent-light)" stopOpacity="0.45" />
          <stop offset="100%" stopColor="var(--accent-light)" stopOpacity="0" />
        </radialGradient>
        <radialGradient id="auroraC">
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0.3" />
          <stop offset="100%" stopColor="var(--accent)" stopOpacity="0" />
        </radialGradient>
        <linearGradient id="ribbon" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0" />
          <stop offset="50%" stopColor="var(--accent)" stopOpacity="0.5" />
          <stop offset="100%" stopColor="var(--accent)" stopOpacity="0" />
        </linearGradient>
        <filter id="auroraBlur" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="55" />
        </filter>
      </defs>

      <g filter="url(#auroraBlur)">
        <ellipse cx="300" cy="220" rx="300" ry="240" fill="url(#auroraA)">
          <animateTransform
            attributeName="transform"
            type="translate"
            values="0 0; 60 40; -30 20; 20 -30; 0 0"
            keyTimes="0; 0.3; 0.55; 0.8; 1"
            dur="24s"
            repeatCount="indefinite"
          />
        </ellipse>

        <ellipse cx="900" cy="180" rx="260" ry="220" fill="url(#auroraB)">
          <animateTransform
            attributeName="transform"
            type="translate"
            values="0 0; -50 30; 40 -20; -20 25; 0 0"
            keyTimes="0; 0.25; 0.5; 0.75; 1"
            dur="30s"
            repeatCount="indefinite"
          />
        </ellipse>

        <ellipse cx="600" cy="420" rx="320" ry="200" fill="url(#auroraC)">
          <animateTransform
            attributeName="transform"
            type="translate"
            values="0 0; 30 -40; -40 10; 10 30; 0 0"
            keyTimes="0; 0.3; 0.6; 0.85; 1"
            dur="28s"
            repeatCount="indefinite"
          />
        </ellipse>
      </g>

      {/* a faint flowing ribbon — slow horizontal drift */}
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
