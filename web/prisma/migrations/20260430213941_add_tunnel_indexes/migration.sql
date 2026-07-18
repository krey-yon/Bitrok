-- CreateIndex
CREATE INDEX "tunnel_userId_createdAt_idx" ON "tunnel"("userId", "createdAt");

-- CreateIndex
CREATE INDEX "tunnel_log_tunnelId_createdAt_idx" ON "tunnel_log"("tunnelId", "createdAt");
