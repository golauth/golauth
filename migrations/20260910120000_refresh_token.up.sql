create table golauth_refresh_token
(
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid         not null references golauth_user (id) on delete cascade,
    -- SHA-256 hex digest of the opaque token. Only the digest is stored, so a
    -- database dump does not hand over live sessions.
    token_hash  varchar(64)  not null,
    issued_at   timestamp    not null default current_timestamp,
    expires_at  timestamp    not null,
    revoked_at  timestamp,
    -- the token that superseded this one on rotation; a non-null value here on a
    -- token that is presented again is refresh-token reuse.
    replaced_by uuid         references golauth_refresh_token (id) on delete set null,
    user_agent  varchar(512),
    client_ip   varchar(64)
);

create unique index ui_golauth_refresh_token_hash
    on golauth_refresh_token (token_hash);

create index ix_golauth_refresh_token_user
    on golauth_refresh_token (user_id);

create index ix_golauth_refresh_token_expires_at
    on golauth_refresh_token (expires_at);
