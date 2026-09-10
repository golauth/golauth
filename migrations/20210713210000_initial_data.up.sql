-- Seed the ADMIN and USER roles and authorities, but NO user.
--
-- A shipped default administrator credential (its bcrypt hash lives in version
-- control) means every untouched deployment is compromised. Instead the service
-- creates the first administrator at start-up from BOOTSTRAP_ADMIN_USER /
-- BOOTSTRAP_ADMIN_PASSWORD when no user holds the ADMIN authority, and refuses
-- to boot if neither an admin nor those variables exist.
insert into golauth_authority (name, description)
values ('ADMIN', 'Authority ADMIN');
insert into golauth_authority (name, description)
values ('USER', 'Authority USER');

insert into golauth_role (name, description)
values ('ADMIN', 'Role ADMIN');
insert into golauth_role (name, description)
values ('USER', 'Role USER');

insert into golauth_role_authority (role_id, authority_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'), (select a.id from golauth_authority a where a.name = 'ADMIN');
insert into golauth_role_authority (role_id, authority_id)
select (select r.id from golauth_role r where r.name = 'USER'), (select a.id from golauth_authority a where a.name = 'USER');
