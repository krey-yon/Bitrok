import assert from "node:assert/strict";
import test from "node:test";

import type { LogDTO, LogsResponse, TunnelDTO } from "./server-api.ts";
import { dashboardView } from "./dashboard-view.ts";

const tunnel: TunnelDTO = {
  id: "tun_1",
  user_id: "usr_1",
  name: "myapp",
  host: "myapp-vikas.bitrok.tech",
  port: 3000,
  active: true,
  created_at: "2026-01-01T00:00:00.000Z",
  updated_at: "2026-01-01T00:00:00.000Z",
};

const log: LogDTO = {
  id: 1,
  tunnel_id: "tun_1",
  tunnel_name: "myapp",
  method: "GET",
  path: "/",
  status: 200,
  latency_ms: 12,
  bytes_in: 0,
  bytes_out: 128,
  ts: "2026-01-01T00:00:01.000Z",
};

const logsOk: LogsResponse = { total: 41, logs: [log] };

function fulfilled<T>(value: T): PromiseFulfilledResult<T> {
  return { status: "fulfilled", value };
}

function rejected(reason = new Error("relay unreachable")): PromiseRejectedResult {
  return { status: "rejected", reason };
}

test("rejected tunnels fetch is relay-down, not an empty account", () => {
  const view = dashboardView("vikas", rejected(), fulfilled({ total: 0, logs: [] }));
  assert.deepEqual(view, { kind: "relay-down", username: "vikas" });
});

test("fulfilled empty tunnels is first-run even if logs failed", () => {
  const view = dashboardView("vikas", fulfilled([]), rejected());
  assert.deepEqual(view, { kind: "first-run", username: "vikas" });
});

test("fulfilled nonempty tunnels is operating and keeps logs", () => {
  const view = dashboardView("vikas", fulfilled([tunnel]), fulfilled(logsOk));
  assert.deepEqual(view, {
    kind: "operating",
    username: "vikas",
    tunnels: [tunnel],
    logs: [log],
    totalRequests: 41,
    logsUnavailable: false,
  });
});

test("operating marks logs unavailable when the logs fetch rejects", () => {
  const view = dashboardView("vikas", fulfilled([tunnel]), rejected());
  assert.equal(view.kind, "operating");
  if (view.kind !== "operating") return;
  assert.equal(view.logsUnavailable, true);
  assert.deepEqual(view.logs, []);
  assert.equal(view.totalRequests, 0);
  assert.deepEqual(view.tunnels, [tunnel]);
});
