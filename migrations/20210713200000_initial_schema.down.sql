-- Drop in dependency order. The original `drop schema golauth cascade` targeted
-- a schema that is never created (the tables live in the default schema), so the
-- rollback always failed and the down path was never exercised.
drop table if exists golauth_user_role;
drop table if exists golauth_role_authority;
drop table if exists golauth_authority;
drop table if exists golauth_role;
drop table if exists golauth_user;
