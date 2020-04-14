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

insert into golauth_user (username, first_name, last_name, email, password)
values ('admin', 'Admin', 'Admin', 'admin@goauth.org',
        '$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22'); -- password = admin123

insert into golauth_user_role (role_id, user_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'), (select u.id from golauth_user u where u.username = 'admin');
