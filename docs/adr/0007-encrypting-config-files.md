# 7. Encrypting config files

Date: 2026-08-13

## Status

Accepted

## Context

After giving a live test run of how portable the Flox environment and dependency management is, it was realized how much the Pulumi code relies on certain environment variables to be set on the user's local machine. The `${ECA_PULUMI_CONFIG_SECRET}` variable references `${PULUMI_CONFIG_PASSPHRASE}`. The `${ECA_OBJ_ACCESS_KEY}` and `${ECA_OBJ_SECRET_KEY}` variables reference `${AWS_ACCESS_KEY_ID}` and `${AWS_SECRET_ACCESS_KEY}` respectively. The `${ECA_LINODE_TOKEN}` variable references `${LINODE_TOKEN}`. All of the reference variables could at any point be used by other programs running on a user's machine, so the decision to utilize Edgecase Apparel specific environment variables (prefixed with `ECA`) was done to avoid conflict. This means a couple things:

- For a user to share the same development experience, that user needs to manually set all the same `ECA` variables in their environment
- The secrets contained in these environment variables will need to be shared

Both of these points introduce more margin for error and friction than is necessary. To alleviate these concerns, we discussed implementing [SOPS](https://github.com/getsops/sops) for encrypting and committing a common `.env` file to the git repository, with [Age](https://github.com/FiloSottile/age) as the encryption and key management backend. The combination of SOPS and Age provides a safe way to encrypt and then commit a shared `.env` file to the git repository, which other team members can then decrypt on their local machine.

## Decision

We put all the secrets which Pulumi requires in a `.env` file, that is then encrypted via SOPS with an Age key, into a file called `.secret.env.enc`. The plaintext `.env` file is excluded in the `.gitignore` file via glob pattern matching, and the encrypted `.secret.env.enc` file is committed to the repository. The only secret we need to concern ourselves with sharing at the moment (via shared vault in password manager) is just the Age **private** key. More detailed documentation on the process can be found [here](../sops-age-encryption.md).

## Consequences

The local development environment is more portable, with less friction points during setup, and the error surface for secrets sharing is greatly reduced.   
