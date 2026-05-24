# 2. Infrastructure as code

Date: 2026-05-21

## Status

Accepted

## Context

The issue motivating this decision, and any context that influences or constrains the decision.
Cloud infrastructure for this project should remain as be cloud provider agnostic as possible, deploy reproducibly across environments (dev, staging, production) and cloud hosting. It should follow common practices around using Infrastructure as Code (IaC) to declare, version, and document the cloud resources in use, and to reconcile drift from a desired state.

## Decision

Pulumi was selected as the IaC tooling due to its support for multiple programming languages, structured config/markup languages, and related testing frameworks. As architecture decisions evolve through time, Pulumi offers the most flexibility for ongoing changes and refactoring. It also provides native support for MCP and codegen.

## Consequences

What becomes easier or more difficult to do and any risks introduced by the change that will need to be mitigated.
Pulumi is lesser known compared to Terraform/Open Tofu, and thus we may see less IaC PRs from external contributors.

There may be less documentation, examples, tutorials and other training material compared to Terraform/Open Tofu.
