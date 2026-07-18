import { auth } from "@/lib/auth";
import { NextRequest } from "next/server";

function handler(req: NextRequest) {
  return auth.handler(req);
}

export const GET = handler;
export const POST = handler;
export const PUT = handler;
export const PATCH = handler;
export const DELETE = handler;
export const HEAD = handler;
export const OPTIONS = handler;
