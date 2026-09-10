create table golauth_login_attempt
(
    user_id      uuid primary key references golauth_user (id) on delete cascade,
    failed_count integer   not null default 0,
    last_failure timestamp,
    locked_until timestamp
);
