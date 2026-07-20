-- Bind each CLI authorization state to the validated callback that created it.
ALTER TABLE "cli_auth_request" ADD COLUMN "callbackUrl" TEXT;
