import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Authentication Error",
  description: "An error occurred during authentication. Please try again.",
  robots: {
    index: false,
    follow: false,
  },
};

export default function ErrorLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
