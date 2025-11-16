-- Remove email column from invitations table
ALTER TABLE invitations DROP COLUMN IF EXISTS email;
