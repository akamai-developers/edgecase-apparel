# 3. Use Golang SDK for IaC

Date: 2026-05-21

## Status

Accepted

## Context

Per [2. Infrastructure as Code], Pulumi supports codifyng infrastructure in multiple markup and programming languages. 

## Decision

[Go]() was selected as the IaC language, via the Pulumi Golang SDK. In addition to being the language that powers Pulumi itself, this enables us to more easily expand or embed Pulumi functionality into other Go-based projects, libraries, and frameworks. Go is a type safe and memory safe language, often noted as having less of a learning curve compared to low-level "systems programming" languages, as well as those used primarirly for high-level application code.

## Consequences

Languages such as Python or Java dominate in some areas such as AI or data science. Some scenarios may require engineers in these fields to become more familiar with the basics of Go.
