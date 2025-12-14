ALTER TABLE users
  ADD COLUMN password_hash TEXT;

UPDATE users
  SET password_hash = repeat('x', 60)
  WHERE password_hash IS NULL;

ALTER TABLE users
  ALTER COLUMN password_hash SET NOT NULL;
