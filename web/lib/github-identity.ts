type GithubEmail = {
  email?: string;
  primary?: boolean;
  verified?: boolean;
};

type GithubProfile = {
  id?: number | string;
  login?: string;
  name?: string | null;
  email?: string | null;
  avatar_url?: string;
};

type GithubFetch = (
  input: string | URL | Request,
  init?: RequestInit,
) => Promise<Response>;

export function selectVerifiedGithubEmail(emails: GithubEmail[]): string | null {
  const selected =
    emails.find((entry) => entry.primary && entry.verified && entry.email) ??
    emails.find((entry) => entry.verified && entry.email);
  return selected?.email?.trim().toLowerCase() || null;
}

function githubNoreplyEmail(id: number | string, login: string): string {
  return `${id}+${login.toLowerCase()}@users.noreply.github.com`;
}

export async function fetchVerifiedGithubIdentity(
  accessToken: string,
  githubFetch: GithubFetch = fetch,
) {
  const headers = {
    Accept: "application/vnd.github+json",
    Authorization: `Bearer ${accessToken}`,
    "User-Agent": "Bitrok-Web",
    "X-GitHub-Api-Version": "2022-11-28",
  };

  const profileResponse = await githubFetch("https://api.github.com/user", { headers });
  if (!profileResponse.ok) {
    console.error("GitHub profile request failed", profileResponse.status);
    return null;
  }

  const profile = (await profileResponse.json()) as GithubProfile;
  if (!profile.id || !profile.login) return null;

  const emailsResponse = await githubFetch("https://api.github.com/user/emails", {
    headers,
  });
  let verifiedEmail: string | null = null;
  if (emailsResponse.ok) {
    const emails = (await emailsResponse.json()) as GithubEmail[];
    verifiedEmail = selectVerifiedGithubEmail(emails);
  } else {
    console.warn("GitHub email request unavailable", emailsResponse.status);
  }

  const email =
    verifiedEmail ||
    profile.email?.trim().toLowerCase() ||
    githubNoreplyEmail(profile.id, profile.login);

  return {
    user: {
      id: String(profile.id),
      email,
      emailVerified: true,
      name: profile.name || profile.login,
      image: profile.avatar_url,
    },
    data: profile,
  };
}
