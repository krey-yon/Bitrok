const EXPIRY_DAYS = 180;

export function buildSecurityTxt(baseUrl: string, contact: string, now = Date.now()): string {
  const expires = new Date(now + EXPIRY_DAYS * 24 * 60 * 60 * 1000).toISOString();

  return [
    `Contact: mailto:${contact}`,
    `Expires: ${expires}`,
    `Acknowledgments: ${baseUrl}/security`,
    `Canonical: ${baseUrl}/.well-known/security.txt`,
    `Policy: ${baseUrl}/security`,
    "Preferred-Languages: en",
  ].join("\n");
}
