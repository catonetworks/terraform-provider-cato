# Coding Guidelines

## Codebase structure

### STRUCT-000: resource-specific-file-placement

**Level:** MUST

Golang files related to a specific resource are placed under `./internal/provider/{domain}/{resource}/` (e.g. `./internal/provider/network/netrange/`). The files include Terraform resource and data-source code, related type definitions, validators, plan modifiers, and unit tests.

**Rationale:**

Go packages provide clear boundaries that simplify code reviews and prevent introduction of accidental unrelated changes.

### STRUCT-001: shared-file-placement

**Level:** MUST

Golang files that are shared by multiple resources are placed under `./internal/provider/common/{package-name}/` (e.g. `./internal/provider/common/dhcp/`). The files may be either defining a common object shared by multiple resources (e.g. dhcp), or generic utilities, parsers, etc.

**Rationale:**

The 'common' folder make it explicit that multiple resources can be affected.

### STRUCT-002: example-placement

**Level:** MUST

Resource examples are located in examples/resources/{resource-name}/resource.tf (e.g. `./examples/resources/cato_if_rule/resource.tf`). Data-source examples are located in examples/data-sources/{data-source-name}/data-source.tf (e.g. `./examples/data-sources/cato_group/data-source.tf`).

**Rationale:**

Standardized place used also by documentation generator.

### STRUCT-003: acceptance-test-placement

**Level:** MUST

Acceptance test files are located in internal/acctests/{resource-name}/ (e.g. `./internal/acctests/group/group_test.go`).

**Rationale:**

The 'acctests' directory contains the whole test suite.

## File names

### FILE-NAME-000: file-name-lowercase

**Level:** MUST

file names are lower case with underscores, (e.g. `resource_network_range.go`)

**Rationale:**

Avoid problems on POSIX filesystems.

### FILE-NAME-001: file-name-datasource

**Level:** MUST

data-source files are prefixed with `datasource_` (e.g. `datasource_network_range.go`)

**Rationale:**

Standardized naming convention simplifies code base maintenance.

### FILE-NAME-002: file-name-resource

**Level:** MUST

resource files are prefixed with `resource_` (e.g. `resource_network_range.go`)

**Rationale:**

Standardized naming convention simplifies code base maintenance.

### FILE-NAME-003: file-name-types

**Level:** MUST

terraform type definition files (used for terraform state or plan) are prefixed with `type_` (e.g. `type_network_range.go`)

**Rationale:**

Standardized naming convention simplifies code base maintenance.

### FILE-NAME-004: file-name-validator

**Level:** MUST

validator files are prefixed with `validator_` (e.g. `validator_network_range.go`)

**Rationale:**

Standardized naming convention simplifies code base maintenance.

### FILE-NAME-005: file-name-plan-modifier

**Level:** MUST

plan modifier files are prefixed with `plan_` (e.g. `plan_dhcp_settings.go`)

**Rationale:**

Standardized naming convention simplifies code base maintenance.

### FILE-NAME-006: file-name-unit-tests

**Level:** MUST

unit test files end with `_test.go` (e.g. `resource_network_range_test.go`)

**Rationale:**

Go Test framework conventions.

## Type system

### TYPES-000: types-model-struct

**Level:** MUST

TF type model (state, plan) must be a struct containing fields with terraform-plugin-framework types (e.g. types.String, types.Object, ...)

**Rationale:**

TF Framework needs to be able to work with null and unknown values
