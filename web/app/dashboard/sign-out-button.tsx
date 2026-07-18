"use client";

import { useState } from "react";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { authClient } from "@/lib/auth-client";

export function SignOutButton() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const handleSignOut = async () => { setLoading(true); await authClient.signOut(); router.push("/"); router.refresh(); };
  return <button type="button" onClick={handleSignOut} disabled={loading} aria-label="Sign out" title="Sign out" className="inline-flex size-9 items-center justify-center rounded-md text-muted-foreground transition-[color,background-color] hover:bg-foreground/[0.05] hover:text-foreground disabled:opacity-50"><LogOut className="size-4" aria-hidden /></button>;
}
