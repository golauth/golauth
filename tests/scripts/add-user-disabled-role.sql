-- The admin user has exactly one role, ADMIN, and that role starts disabled.
-- FindAuthoritiesByUserID must return nothing until the role is enabled.
insert into golauth_authority (id, name, description)
values ('8ae4420b-760c-47a6-ab7a-1cb2f9f07c16', 'ADMIN', 'Authority ADMIN');

insert into golauth_role (id, name, description, enabled)
values ('7f68301e-df80-45bd-9532-23a58733ef2c', 'ADMIN', 'Role ADMIN', false);

insert into golauth_role_authority (role_id, authority_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'),
       (select a.id from golauth_authority a where a.name = 'ADMIN');

insert into golauth_user (id, username, first_name, last_name, email, document, password)
values ('8c61f220-8bb8-48b9-b225-d54dfa6503db', 'admin', 'Admin', 'Admin', 'admin@goauth.org', '000',
        '$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22'); -- password = admin123

insert into golauth_user_role (role_id, user_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'),
       (select u.id from golauth_user u where u.username = 'admin');
