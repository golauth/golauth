insert into golauth_authority (id, name, description)
values (1, 'ADMIN', 'Authority ADMIN');
insert into golauth_authority (id, name, description)
values (2, 'USER', 'Authority USER');

insert into golauth_role (id, name, description)
values (1,'ADMIN', 'Role ADMIN');
insert into golauth_role (id, name, description)
values (2, 'USER', 'Role USER');

insert into golauth_role_authority (role_id, authority_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'), (select a.id from golauth_authority a where a.name = 'ADMIN');
insert into golauth_role_authority (role_id, authority_id)
select (select r.id from golauth_role r where r.name = 'USER'), (select a.id from golauth_authority a where a.name = 'USER');

insert into golauth_user (id, username, first_name, last_name, email, document, password)
values (1, 'admin', 'Admin', 'Admin', 'admin@goauth.org', '000',
        '$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22'); -- password = admin123

insert into golauth_user_role (role_id, user_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'), (select u.id from golauth_user u where u.username = 'admin');

insert into golauth_user_role (role_id, user_id)
select (select r.id from golauth_role r where r.name = 'USER'), (select u.id from golauth_user u where u.username = 'admin');

insert into golauth_user (id, username, first_name, last_name, email, document, password)
values (2, 'admin2', 'Admin2', 'Admin2', 'admin2@goauth.org', '001',
        '$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22'); -- password = admin123
