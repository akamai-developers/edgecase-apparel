# Architecture Decision Records

This project recognizes the importance of maintaining an [architecture decision record (ADR)](https://github.com/architecture-decision-record/architecture-decision-record), using the [template by Michael Nygard](https://github.com/architecture-decision-record/architecture-decision-record/tree/main/locales/en/templates/decision-record-template-by-michael-nygard). The `adr-tools` CLI should be installed and used for generating ADRs whenever a new decision is to be made.


## Install ADR Tools

The simplest and presumably most common method is to install with [Homebrew](https://brew.sh/). See the [install guide](https://github.com/npryce/adr-tools/blob/master/INSTALL.md) if manual installation is required.

```bash
brew install adr-tools
```

## Creating ADRs

For a new project (or to change up and existing project), you can initialize a new `adr-tools` schema with the `init` command.

```bash
adr init [path]
```

This project has already done that by running `adr init docs/adr` when this repo was created. All new ADRs created by this tool will populate the `docs/adr` subdirectory as markdown files. Invocation of the `init` command also writes that first record, titled `0001-record-architecture-decisions.md`.

Creating a new ADR is as simple as running `adr-tools` with the `new` command, followed by the title of the document. For example, we need to create a decision record for our choice of IaC tooling. Running the below command produces the new templated markdown file.  

```bash
adr new Infrastructure as code
...
docs/adr/0002-infrastructure-as-code.md
```

At the time of this writing, the `docs/adr` subdirectory contains the following ADRs:
```bash
docs
└── adr
    ├── 0001-record-architecture-decisions.md
    ├── 0002-infrastructure-as-code.md
    ├── 0003-use-golang-sdk-for-iac.md
    └── 0004-environment-configuration-tooling.md
```

Use the `-s` option when creating a new ADR that supersedes a previous one. If for example, we wanted to supersede the `0002-infrastructure-as-code.md` we created earlier, we supply the record's number as an argument. 

```bash
adr new -s 9 Use Rust for performance-critical functionality
...
docs/adr/0005-infrastructure-as-code-revised.md
```

```bash
docs
└── adr
    ├── 0001-record-architecture-decisions.md
    ├── 0002-infrastructure-as-code.md
    ├── 0003-use-golang-sdk-for-iac.md
    ├── 0004-environment-configuration-tooling.md
    └── 0005-infrastructure-as-code-revised.md
```

The newly generated file will already have notated which record it supersedes. See the [README](https://github.com/npryce/adr-tools) for examples and more.
