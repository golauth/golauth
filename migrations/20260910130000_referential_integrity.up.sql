-- Referential integrity for the join tables, plus the two columns whose entity
-- fields never had a column behind them.
--
-- ORPHAN CLEANUP: the foreign keys below will not apply while a join row points
-- at a user, role or authority that no longer exists. Inspect the offending
-- rows on a real installation before running this migration:
--
--   SELECT ur.* FROM golauth_user_role ur
--    WHERE NOT EXISTS (SELECT 1 FROM golauth_user u WHERE u.id = ur.user_id)
--       OR NOT EXISTS (SELECT 1 FROM golauth_role r WHERE r.id = ur.role_id);
--
--   SELECT ra.* FROM golauth_role_authority ra
--    WHERE NOT EXISTS (SELECT 1 FROM golauth_role r      WHERE r.id = ra.role_id)
--       OR NOT EXISTS (SELECT 1 FROM golauth_authority a WHERE a.id = ra.authority_id);
--
-- The DELETEs remove exactly that set -- a deliberate delete, not a cascade.

DELETE FROM golauth_user_role ur
 WHERE NOT EXISTS (SELECT 1 FROM golauth_user u WHERE u.id = ur.user_id)
    OR NOT EXISTS (SELECT 1 FROM golauth_role r WHERE r.id = ur.role_id);

DELETE FROM golauth_role_authority ra
 WHERE NOT EXISTS (SELECT 1 FROM golauth_role r WHERE r.id = ra.role_id)
    OR NOT EXISTS (SELECT 1 FROM golauth_authority a WHERE a.id = ra.authority_id);

-- Missing columns. golauth_user_role.enabled backs entity.UserRole.Enabled and
-- is read by FindAuthoritiesByUserID so a disabled membership grants nothing.
-- golauth_role_authority.creation_date brings the table in line with every
-- sibling.
ALTER TABLE golauth_user_role
    ADD COLUMN enabled boolean NOT NULL DEFAULT true;

ALTER TABLE golauth_role_authority
    ADD COLUMN creation_date timestamp NOT NULL DEFAULT current_timestamp;

-- Foreign keys. Cascade on the user side (a deleted user has no memberships);
-- restrict on the role and authority sides (removing one that is still in use
-- must be a deliberate act, not a silent side effect).
ALTER TABLE golauth_user_role
    ADD CONSTRAINT fk_golauth_user_role_user
        FOREIGN KEY (user_id) REFERENCES golauth_user (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_golauth_user_role_role
        FOREIGN KEY (role_id) REFERENCES golauth_role (id) ON DELETE RESTRICT;

ALTER TABLE golauth_role_authority
    ADD CONSTRAINT fk_golauth_role_authority_role
        FOREIGN KEY (role_id) REFERENCES golauth_role (id) ON DELETE RESTRICT,
    ADD CONSTRAINT fk_golauth_role_authority_authority
        FOREIGN KEY (authority_id) REFERENCES golauth_authority (id) ON DELETE RESTRICT;
