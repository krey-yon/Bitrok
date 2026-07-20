type GithubEmail = {
  email?: string;
  primary?: boolean;
  verified?: boolean;
};

export function selectVerifiedGithubEmail(emails: GithubEmail[]): string | null {
  const selected =
    emails.find((entry) => entry.primary && entry.verified && entry.email) ??
    emails.find((entry) => entry.verified && entry.email);
  return selected?.email?.trim().toLowerCase() || null;
}
