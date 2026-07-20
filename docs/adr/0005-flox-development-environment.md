# 5. Flox development environment

Date: 2026-07-17

## Status

Accepted

## Context

[Flox](https://flox.dev/docs) is a cross-platform, virtual environment that declaratively packages all dependencies for a reproducible developer experience throughout the SDLC. When a Flox environment is active, those declared dependencies are isolated in a layer on top of the host OS, while still allowing the developer to access locally installed software that is _not_ declared. In addition, Flox can [export an environment](https://flox.dev/docs/man/flox-containerize) to OCI compliant container images. Lastly, the [Flox Hub](https://flox.dev/docs/concepts/floxhub) provides a registry for pushing and pulling remote environments, written by ourselves or the community.

## Decision

We are implementing Flox for the purpose of declarative dependency management, and maintaining as much portability as possible across development and testing environments. Additionally this allows us to publish reproducible builds for the larger community, and remotely [include](https://flox.dev/docs/man/manifest.toml#include) them where there is need to compose a merged environment.

## Consequences

Requires a developer to first install Flox. Components such as setup scripts or makefiles may need to be updated accordingly.
