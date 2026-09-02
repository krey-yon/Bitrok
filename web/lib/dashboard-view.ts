import type { LogDTO, LogsResponse, TunnelDTO } from "./server-api";

export type DashboardView =
  | { kind: "relay-down"; username: string }
  | { kind: "first-run"; username: string }
  | {
      kind: "operating";
      username: string;
      tunnels: TunnelDTO[];
      logs: LogDTO[];
      totalRequests: number;
      logsUnavailable: boolean;
    };

export function dashboardView(
  username: string,
  tunnelsResult: PromiseSettledResult<TunnelDTO[]>,
  logsResult: PromiseSettledResult<LogsResponse>,
): DashboardView {
  if (tunnelsResult.status === "rejected") {
    return { kind: "relay-down", username };
  }

  if (tunnelsResult.value.length === 0) {
    return { kind: "first-run", username };
  }

  const logsUnavailable = logsResult.status === "rejected";
  return {
    kind: "operating",
    username,
    tunnels: tunnelsResult.value,
    logs: logsUnavailable ? [] : logsResult.value.logs,
    totalRequests: logsUnavailable ? 0 : logsResult.value.total,
    logsUnavailable,
  };
}
