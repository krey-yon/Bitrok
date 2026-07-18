-- CreateTable
CREATE TABLE "cli_auth_request" (
    "id" TEXT NOT NULL,
    "state" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'pending',
    "token" TEXT,
    "userId" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "expiresAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "cli_auth_request_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "cli_auth_request_state_key" ON "cli_auth_request"("state");

-- CreateIndex
CREATE INDEX "cli_auth_request_state_idx" ON "cli_auth_request"("state");

-- CreateIndex
CREATE INDEX "cli_auth_request_expiresAt_idx" ON "cli_auth_request"("expiresAt");
