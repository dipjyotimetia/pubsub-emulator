# Security Policy

## Supported Versions

This project is a local development and testing tool. Security fixes are applied
to the latest released image and the `main` branch only.

## Reporting a Vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, report privately via GitHub's
[private vulnerability reporting](https://github.com/dipjyotimetia/pubsub-emulator/security/advisories/new)

Include:

- A description of the issue and its impact
- Steps to reproduce (proof of concept if possible)
- Affected version or commit

You can expect an acknowledgement within a few days. Once a fix is available, a
new release and image will be published.

## Scope

This emulator is intended for **local, trusted environments**. It ships with no
authentication and binds to all interfaces inside its container by design. Do
not expose it to untrusted networks or use it in production.
