insert into golauth_authority (id, name, description)
values ('8ae4420b-760c-47a6-ab7a-1cb2f9f07c16', 'ADMIN', 'Authority ADMIN');
insert into golauth_authority (id, name, description)
values ('df30f1c0-9a0c-4095-a14b-13d44d39ec15', 'USER', 'Authority USER');

insert into golauth_role (id, name, description)
values ('7f68301e-df80-45bd-9532-23a58733ef2c','ADMIN', 'Role ADMIN');
insert into golauth_role (id, name, description)
values ('c12b415b-c3ad-487f-9800-f548aa18cc58', 'USER', 'Role USER');

insert into golauth_role_authority (role_id, authority_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'), (select a.id from golauth_authority a where a.name = 'ADMIN');
insert into golauth_role_authority (role_id, authority_id)
select (select r.id from golauth_role r where r.name = 'USER'), (select a.id from golauth_authority a where a.name = 'USER');

insert into golauth_user (id, username, first_name, last_name, email, document, password)
values ('8c61f220-8bb8-48b9-b225-d54dfa6503db', 'admin', 'Admin', 'Admin', 'admin@goauth.org', '000',
        '$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22'); -- password = admin123

insert into golauth_user_role (role_id, user_id)
select (select r.id from golauth_role r where r.name = 'ADMIN'), (select u.id from golauth_user u where u.username = 'admin');

insert into golauth_user_role (role_id, user_id)
select (select r.id from golauth_role r where r.name = 'USER'), (select u.id from golauth_user u where u.username = 'admin');

insert into golauth_user (id, username, first_name, last_name, email, document, password)
values ('e227d878-b5d6-4902-a500-3357955c962d', 'admin2', 'Admin2', 'Admin2', 'admin2@goauth.org', '001',
        '$2a$10$VNkiJ40.00IfVjxo8ILyauLUbnxMcKK2G/FbbwdsTYb.lCuZEbh22'); -- password = admin123
