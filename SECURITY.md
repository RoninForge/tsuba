# Security Policy

## Scope

tsuba is a local CLI tool that writes files. Given a name and a kind (skill, plugin, hook, agent), it renders embedded Go templates and writes the output to paths under the current working directory or a user-specified target directory. It does not open network sockets (except for the explicit `tsuba publish` flow, which is opt-in), does not phone home, does not collect telemetry, does not require elevated privileges, and refuses to write outside the target directory.

A security issue is anything that violates the above: unintended file writes (path traversal in `--name` inputs), network egress in non-publish flows, privilege escalation, arbitrary code execution during scaffolding, or anything that leaks data off the machine.

## Reporting a vulnerability

**Do not file a public issue.**

Send the details to **security@roninforge.org** with:

- A description of the issue and its impact
- Steps to reproduce (ideally a minimal `tsuba new ...` invocation that triggers it)
- Affected versions
- Your name and whether you want credit in the advisory

You will get an acknowledgement within 72 hours. We aim to have a patched release available within 14 days for high-severity issues and 30 days for lower-severity ones. The embargo window is 90 days.

## Supported versions

Only the latest minor release on the `main` branch receives security fixes.

## No bug bounty

tsuba is a small OSS project. We cannot pay bounties. We will credit responsible disclosures in release notes and the advisory.
