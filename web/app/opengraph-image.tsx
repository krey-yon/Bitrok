import { ImageResponse } from "next/og";

export const alt = "Bitrok deterministic tunneling infrastructure";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpenGraphImage() {
  return new ImageResponse(
    <div
      style={{
        width: "100%",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        justifyContent: "space-between",
        background: "#0c0f0a",
        color: "#f2f6e9",
        padding: "64px 72px",
        fontFamily: "Arial, sans-serif",
        letterSpacing: 0,
      }}
    >
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 18 }}>
          <div
            style={{
              width: 52,
              height: 52,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              border: "2px solid #b8f34a",
              borderRadius: 8,
              color: "#b8f34a",
              fontSize: 28,
              fontWeight: 700,
            }}
          >
            B
          </div>
          <div style={{ display: "flex", fontSize: 34, fontWeight: 700 }}>Bitrok</div>
        </div>
        <div style={{ display: "flex", color: "#a5ad9b", fontSize: 22 }}>
          HTTP tunnels
        </div>
      </div>

      <div style={{ display: "flex", flexDirection: "column", maxWidth: 960 }}>
        <div style={{ display: "flex", color: "#b8f34a", fontSize: 24, fontWeight: 700 }}>
          ONE URL. EVERY SESSION.
        </div>
        <div
          style={{
            display: "flex",
            marginTop: 22,
            fontSize: 72,
            lineHeight: 1.05,
            fontWeight: 700,
          }}
        >
          Deterministic tunnels to localhost.
        </div>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          borderTop: "1px solid #343b30",
          paddingTop: 28,
          fontSize: 23,
        }}
      >
        <div style={{ display: "flex", color: "#c8cec0", fontFamily: "monospace" }}>
          myapp-you.bitrok.tech
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 12, color: "#b8f34a" }}>
          <div style={{ display: "flex", width: 10, height: 10, borderRadius: 5, background: "#b8f34a" }} />
          Stable across restarts
        </div>
      </div>
    </div>,
    size,
  );
}
