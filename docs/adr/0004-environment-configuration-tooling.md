# 4. Environment configuration tooling

Date: 2026-05-21

## Status

Accepted

## Context

The [Twelve-Factor App](https://12factor.net/) represents a set of design principals that decouple application components from the runtime. Application _config_ describes anything that is subject to change between deployment targets (i.e. dev, staging, prod). Adherence requires [separating the app's config](https://12factor.net/config) from its internal logic, state, and other dependecies.  

## Decision

We are implementing [Viper](https://github.com/spf13/viper)―a battle-tested config management library used by several enterprise scale [Go](https://go.dev/)-based projects. Viper offers production-ready feature sets and flexibility for working with CLIs, environment variables, and config file formats.

## Consequences

Viper is not language-agnostic. It's specifically for Go projects.
