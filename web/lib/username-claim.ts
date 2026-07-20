export type UsernameClaimDecision =
  | { action: "claim" }
  | { action: "unchanged"; username: string }
  | { action: "reject" };

/** Decide whether an already normalized username may be claimed by this user. */
export function decideUsernameClaim(
  currentUsername: string | null,
  requestedUsername: string,
): UsernameClaimDecision {
  if (!currentUsername) return { action: "claim" };
  if (currentUsername === requestedUsername) {
    return { action: "unchanged", username: currentUsername };
  }
  return { action: "reject" };
}
