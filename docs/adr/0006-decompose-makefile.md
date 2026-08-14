# 6. Decompose Makefile

Date: 2026-08-13

## Status

Accepted

## Context

As we anticipate the complexity and size of the overall project to grow, so will the amount of `Makefile` targets. Therefore it makes sense separate this into sub files that are grouped by common tasks. The project should still maintain a root `Makefile` for those tasks that _are_ and _should_ be more global, such as recursive code linting. Doing this will keep it leaner and more readable as needs arise to increase the number of targets.

## Decision

We have created a sub `Makefile` in the `tests` directory, which groups tasks and environment variables specific to unit, integration, and smoke testing. We have also put a sub `Makefile` in the `cmd` directory which contains the Pulumi apps, to group all IaC related tasks. Both sub files can be referenced from the root file.

## Consequences

Each additonal `Makefile` can introduce more back and forth navigation between directories when making large refactoring changes, but ultimately keeps cleanliness.
