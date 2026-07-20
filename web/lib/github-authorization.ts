const GITHUB_AUTHORIZE_URL = "https://github.com/login/oauth/authorize";

const FORWARDED_PARAMETERS = [
  "response_type",
  "state",
  "code_challenge_method",
  "code_challenge",
  "login",
  "allow_signup",
  "prompt",
] as const;

/**
 * Build a GitHub App authorization URL without OAuth App scopes.
 *
 * Better Auth 1.6 emits `scope=` when `disableDefaultScope` is enabled. GitHub
 * Apps use configured fine-grained permissions and should receive no scope
 * parameter at all, so the local authorization bridge calls this helper before
 * redirecting to GitHub.
 */
export function buildGithubAppAuthorizationURL(
  source: URLSearchParams,
  clientId: string,
  callbackURL: string,
): URL {
  const target = new URL(GITHUB_AUTHORIZE_URL);
  target.searchParams.set("client_id", clientId);
  target.searchParams.set("redirect_uri", callbackURL);

  for (const parameter of FORWARDED_PARAMETERS) {
    const value = source.get(parameter);
    if (value) target.searchParams.set(parameter, value);
  }

  return target;
}
