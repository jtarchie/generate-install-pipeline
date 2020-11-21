# Introduction

This is an exploration of building a pipeline generator for testing installation and upgrades of Ops Manager and tiles.
These patterns are from Platform Automation.
Just extracted into a YAML generator, based on a config file.

# Usage

```bash
go run main.go -config config.yml
```

Please see `examples/` for actual examples.

## Pipeline design

Note: All state will be stored in a git repo, known as `deployments.uri` in the config.

1. Terraform using `paving` to create infrastructure in the desired IAAS.
1. Create an OpsManager VM.
1. Upload, staged, and configure the desired tile.
   If the configuration cannot be defaulted,
   the required vars will be searched for in `deployments/<env-name>/products/<product-slug>-<version>-vars.yml`.
1. Apply changes.
1. Repeat in order, the OpsManager, tile upgrade, or tile installation.
1. Cleanup everything, aggressively.

# Testing

This is using `ginkgo` to test the YAML that is generated.
The final test will be a successful run of the `examples/` in a running concourse.
