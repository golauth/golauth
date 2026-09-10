# Security Policy

## Supported versions

Security fixes are released for the latest minor version. Older lines receive
fixes only for critical issues, at the maintainers' discretion.

| Version | Supported          |
| ------- | ------------------ |
| 0.4.x   | :white_check_mark: |
| < 0.4   | :x:                |

## Reporting a vulnerability

**Do not open a public issue for a security problem.**

Report it through GitHub's private vulnerability reporting:

1. Go to the repository's **Security** tab.
2. Choose **Report a vulnerability** (under *Advisories*).
3. Describe the issue, the affected version or commit, and a proof of concept or
   reproduction steps if you have them.

If you cannot use GitHub advisories, open a minimal public issue that says only
"security report, please provide a private contact" — with no detail — and a
maintainer will follow up privately.

### What to expect

| Stage                                                                 | Target                                                                         |
|-----------------------------------------------------------------------|--------------------------------------------------------------------------------|
| Acknowledgement of your report                                        | within 3 business days                                                         |
| Initial assessment (accepted / needs info / declined, with reasoning) | within 10 business days                                                        |
| Fix or mitigation for an accepted report                              | as soon as practical, prioritised by severity                                  |
| Public disclosure                                                     | coordinated with you, and by default within 90 days of the fix being available |

We will credit you in the advisory and the release notes unless you ask us not
to.

## Scope

In scope:

- authentication and token handling (`/auth/token`, `/auth/token/refresh`,
  `/auth/check_token`, `/auth/logout*`, the JWKS endpoint);
- authorization checks on the protected routes (see
  `docs/authorization-model.md`);
- the signing-key lifecycle, the refresh-token store, the login lockout and
  rate limiting;
- the database bootstrap, migrations and the start-up administrator flow;
- anything that lets a request read or change data it should not, or escalate
  its privileges.

Out of scope:

- findings that require a already-compromised host, database or signing key;
- denial of service from unbounded request volume against an unprotected
  deployment (put a proxy or gateway in front);
- missing hardening headers on responses that carry no HTML;
- the example `docker-compose.yml` and `.env.example`, which carry a clearly
  labelled development password and are not meant for production;
- social engineering and physical attacks.

