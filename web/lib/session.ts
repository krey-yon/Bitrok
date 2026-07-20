import { auth } from "@/lib/auth";
import { headers } from "next/headers";
import { redirect } from "next/navigation";

export async function getSession() {
  const session = await auth.api.getSession({
    headers: await headers(),
  });
  return session;
}

export async function requireAuth() {
  let session;
  try {
    session = await getSession();
  } catch {
    redirect("/error?error=service_unavailable");
  }
  if (!session) {
    redirect("/login");
  }
  return session;
}
