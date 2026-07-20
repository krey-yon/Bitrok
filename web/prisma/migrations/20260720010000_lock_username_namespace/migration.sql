-- Remove legacy backfills that do not match the canonical username format.
-- Those users will be sent through the normal one-time claim flow.
UPDATE "user"
SET "username" = NULL
WHERE "username" IS NOT NULL
  AND (
    char_length("username") NOT BETWEEN 2 AND 32
    OR "username" !~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    OR "username" !~ '[a-z]'
    OR "username" IN (
      'api', 'www', 'app', 'admin', 'dashboard', 'static', 'assets', 'cdn',
      'mail', 'ftp', 'status', 'support', 'help', 'docs', 'blog', 'auth',
      'login', 'register', 'bitrok', 'tunnel', 'tunnels', 'cli', 'null',
      'undefined'
    )
  );

ALTER TABLE "user"
ADD CONSTRAINT "user_username_canonical"
CHECK (
  "username" IS NULL
  OR (
    char_length("username") BETWEEN 2 AND 32
    AND "username" ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    AND "username" ~ '[a-z]'
    AND "username" NOT IN (
      'api', 'www', 'app', 'admin', 'dashboard', 'static', 'assets', 'cdn',
      'mail', 'ftp', 'status', 'support', 'help', 'docs', 'blog', 'auth',
      'login', 'register', 'bitrok', 'tunnel', 'tunnels', 'cli', 'null',
      'undefined'
    )
  )
);

CREATE FUNCTION bitrok_prevent_username_change()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF OLD."username" IS NOT NULL AND NEW."username" IS DISTINCT FROM OLD."username" THEN
    RAISE EXCEPTION 'username cannot be changed after it is claimed'
      USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END;
$$;

CREATE TRIGGER "user_username_immutable"
BEFORE UPDATE OF "username" ON "user"
FOR EACH ROW
EXECUTE FUNCTION bitrok_prevent_username_change();
