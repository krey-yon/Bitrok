-- Add `user.username`: the URL slug used in deterministic tunnel hosts
-- (app-username.bitrok.tech). Populated from the GitHub login at signup going
-- forward; this migration backfills existing users.

ALTER TABLE "user" ADD COLUMN "username" TEXT;

-- Backfill: slugify the display name (better-auth stores GitHub `login` as
-- `name` when no profile name is set), falling back to the email local-part.
-- Lowercase, strip non-alphanumerics. ponytail: naive slug, no length cap —
-- fine for human names/logins; revisit if anyone's is >63 chars.
UPDATE "user"
SET "username" = lower(regexp_replace(coalesce("name", split_part("email", '@', 1)), '[^a-zA-Z0-9]+', '', 'g'))
WHERE "username" IS NULL;

-- De-duplicate any collisions by appending the first 4 chars of the cuid.
-- A normal UNIQUE index allows multiple NULLs in Postgres, so users who
-- somehow still have no username (empty slug) remain unconstrained.
WITH dups AS (
  SELECT id, username,
         row_number() OVER (PARTITION BY username ORDER BY "createdAt") AS rn
  FROM "user"
  WHERE "username" IS NOT NULL AND "username" <> ''
)
UPDATE "user" u
SET "username" = dups.username || '_' || substr(dups.id, 1, 4)
FROM dups
WHERE u.id = dups.id AND dups.rn > 1;

-- Drop any empty-string slugs back to NULL so the unique index doesn't trip.
UPDATE "user" SET "username" = NULL WHERE "username" = '';

CREATE UNIQUE INDEX "user_username_key" ON "user"("username");
