create table golauth_user
(
    id            uuid PRIMARY KEY       DEFAULT gen_random_uuid(),
    username      varchar(255)  not null,
    first_name    varchar(255)  not null,
    last_name     varchar(255)  not null,
    email         varchar(1000) not null,
    document      varchar(100)  not null,
    password      varchar(1000) not null,
    enabled       boolean       not null default true,
    creation_date timestamp     not null default current_timestamp
);

create unique index ui_golauth_user_username
    on golauth_user (username);

create unique index ui_golauth_user_email
    on golauth_user (email);

create table golauth_role
(
    id            uuid PRIMARY KEY       DEFAULT gen_random_uuid(),
    name          varchar(255)  not null,
    description   varchar(1000) not null,
    enabled       boolean       not null default true,
    creation_date timestamp     not null default current_timestamp
);

create unique index ui_golauth_role_name
    on golauth_role (name);

create table golauth_authority
(
    id            uuid PRIMARY KEY       DEFAULT gen_random_uuid(),
    name          varchar(255)  not null,
    description   varchar(1000) not null,
    enabled       boolean       not null default true,
    creation_date timestamp     not null default current_timestamp
);

create unique index ui_golauth_authority_name
    on golauth_authority (name);

create table golauth_role_authority
(
    role_id      uuid not null,
    authority_id uuid not null,
    constraint pk_golauth_role_authority primary key (role_id, authority_id)
);

create table golauth_user_role
(
    user_id       uuid      not null,
    role_id       uuid      not null,
    creation_date timestamp not null default current_timestamp,
    constraint pk_golauth_user_role primary key (user_id, role_id)
);
