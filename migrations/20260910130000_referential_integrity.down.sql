ALTER TABLE golauth_role_authority
    DROP CONSTRAINT IF EXISTS fk_golauth_role_authority_authority,
    DROP CONSTRAINT IF EXISTS fk_golauth_role_authority_role;

ALTER TABLE golauth_user_role
    DROP CONSTRAINT IF EXISTS fk_golauth_user_role_role,
    DROP CONSTRAINT IF EXISTS fk_golauth_user_role_user;

ALTER TABLE golauth_role_authority DROP COLUMN IF EXISTS creation_date;
ALTER TABLE golauth_user_role DROP COLUMN IF EXISTS enabled;
