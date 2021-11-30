# Webhook HCL Playground

This repository contains a proof-of-concept playground for using HCL in
[webhook](https://github.com/adnanh/webhook/).

## Goals

- Use [HCLv2](https://github.com/hashicorp/hcl) to provide a custom, intuitive config DSL.
- Use functions and variables within the DSL for easy expression evaluation.
- Support all CLI options in the config file.

## Progress

See [TODO.md](TODO.md).

## Examples

- **[config-github.hcl](config-github.hcl)**: simply Github webhook
- **[config4.hcl](config4.hcl)**: everything imaginable in one file
