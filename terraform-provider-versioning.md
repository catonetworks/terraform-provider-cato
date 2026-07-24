# Terraform provider versioning

**Status:** Proposed  
**Applies to:** `registry.terraform.io/catonetworks/cato`  
**Effective from:** Cato Terraform Provider `1.0.0`

## Summary

The Cato Terraform Provider follows [Semantic Versioning](https://semver.org/) and the
[HashiCorp provider versioning guidelines](https://developer.hashicorp.com/terraform/plugin/best-practices/versioning):

- **Major (`X.0.0`)**: a practitioner may need to change Terraform configuration, state-handling procedures, or upgrade steps.
- **Minor (`X.Y.0`)**: backward-compatible functionality, including new resources, data sources, attributes, and deprecations.
- **Patch (`X.Y.Z`)**: backward-compatible bug or security fixes intended to be functionally equivalent to the previous release.
- **Prerelease (`X.Y.Z-beta.N`)**: an explicitly selected test build for targeted validation before a stable release.

The version number communicates compatibility, not urgency. Whether a release is relevant to a
particular customer still depends on the resources and functionality they use. Release notes are
the authoritative description of every change.

Provider upgrades are intentional. Terraform records the selected provider version in
`.terraform.lock.hcl`; a newer version is normally selected only when the configured constraint
allows it and the user runs `terraform init -upgrade`.

## Why we are adopting this policy

The provider is an API from the Terraform practitioner's perspective. Its public interface includes:

- provider, resource, and data-source schemas;
- documented attribute types, defaults, validation, and behavior;
- resource and import identifiers;
- state upgrade and migration behavior;
- plan, apply, update, and delete semantics;
- compatibility with supported Terraform versions and Cato APIs.

Predictable compatibility boundaries let customers upgrade deliberately and reduce the risk of
remaining on an old provider until a Cato API change makes it unusable.

This policy replaces the historical `0.0.x` convention. Version `1.0.0` establishes the first
stable compatibility contract. Unless separately documented, `1.0.0` should be built from the
then-current stable `0.0.x` release without bundling unrelated breaking changes. The change to
`1.0.0` is primarily a stability commitment, not permission to redesign the provider.

## Version selection

| Release type | Use when | Typical examples | Required communication |
|---|---|---|---|
| **Major `X.0.0`** | Existing valid configurations or documented workflows may require migration | Remove or rename a resource or attribute; add a required attribute without a compatible default; make an incompatible type/default change; change an ID or import format; change documented behavior in a way that requires configuration changes | Release notes, dedicated upgrade guide, advance announcement, KB/RN, and broad regression and migration testing |
| **Minor `X.Y.0`** | Add backward-compatible functionality or announce deprecation | Add a resource, data source, optional attribute, capability, or compatible validation; deprecate an existing interface while retaining it | Release notes; KB/RN for notable capabilities, deprecations, or operational impact |
| **Patch `X.Y.Z`** | Restore or correct behavior without requiring customer migration | Fix CRUD behavior, state drift, retries, API translation, diagnostics, documentation, or a backward-compatible security issue | Release notes; KB/RN only for high-impact, urgent, or security-sensitive fixes |
| **Beta `X.Y.Z-beta.N`** | Validate a specific upcoming release with selected users | Early implementation of a fix or capability that needs production-like validation | Targeted notes to participants; no general availability announcement |

Each release level may also contain changes from lower levels. For example, a minor release may
include bug fixes, and a major release may include features and fixes.

## How to decide whether a change is breaking

A change is breaking when a practitioner using documented behavior and valid configuration or
state cannot upgrade safely without taking additional migration action.

Examples include:

- removing or renaming a provider argument, resource, data source, or attribute;
- changing an attribute from optional to required without a backward-compatible default;
- changing an attribute type or structure incompatibly;
- narrowing accepted values so previously documented valid configuration is rejected;
- changing a default in a way that changes existing managed infrastructure;
- changing resource identifiers or import syntax incompatibly;
- causing replacement or materially different behavior solely because the provider was upgraded;
- dropping compatibility with a supported Terraform version when that requires customer action.

A fix is not automatically breaking merely because it produces a new plan. For example, correcting
drift detection so Terraform reports a real remote difference can be a patch when configuration
remains valid and the result matches the documented contract. If the old behavior was documented,
widely relied upon, or the correction requires configuration or state migration, use a major
release or first introduce a backward-compatible migration path.

When classification is uncertain, choose the higher compatibility boundary and explain the
decision in the release notes.

## Cato API changes

An upstream Cato API change does not change the meaning of Semantic Versioning:

| API impact on the provider | Provider release |
|---|---|
| The provider must change internally, but existing Terraform configuration and state continue to work | **Patch** |
| The provider adds a new backward-compatible API capability | **Minor** |
| Customers must change Terraform configuration, imports, or migration procedures | **Major** |

Urgency is handled through release and communication priority, not by publishing a breaking
change as a minor or patch release.

### Preferred deprecation flow

When an API retirement is known in advance:

1. Add support for the replacement in a minor release.
2. Deprecate the old provider interface without removing it.
3. Publish the affected resources, replacement, minimum compatible provider version, migration
   steps, and API end-of-life date.
4. Keep both paths working during the transition where the API permits it.
5. Remove the deprecated interface only in a major provider release.

For generally available functionality, target six months of notice and provide no less than three
months when Cato controls the schedule. For beta functionality, the notice period may be as short
as two weeks. If an upstream emergency makes these periods impossible, publish an emergency major
release when migration is required, give as much notice as possible, and use direct support
outreach in addition to normal release channels.

Planned major releases should normally occur no more than once per year. This does not prevent an
emergency major release when preserving compatibility is impossible.

## Beta releases

Beta versions are prereleases of a specific intended stable version:

```text
1.4.0-beta.1
1.4.0-beta.2
1.4.0
```

Rules:

- Increment `N` for every new beta build; never replace or republish an existing tag.
- Selected users must pin the exact beta version.
- Beta versions are not recommended for broad production use.
- A beta may change incompatibly before the corresponding stable release.
- Promote the tested build to the matching stable version after validation.
- The beta designation applies to the whole provider package. The maturity of an individual
  resource or Cato API must be stated separately in its documentation.
- Normal upgrade guidance and latest-stable comparisons exclude prereleases unless the user is
  explicitly participating in a beta.

Example:

```hcl
terraform {
  required_providers {
    cato = {
      source  = "catonetworks/cato"
      version = "= 1.4.0-beta.2"
    }
  }
}
```

## Customer upgrade guidance

Root Terraform configurations should accept compatible releases while excluding the next major
version:

```hcl
terraform {
  required_providers {
    cato = {
      source  = "catonetworks/cato"
      version = ">= 1.4.0, < 2.0.0"
    }
  }
}
```

Customers should:

1. Commit `.terraform.lock.hcl` to version control.
2. Review Cato provider release notes before upgrading.
3. Run `terraform init -upgrade` in a controlled branch.
4. Review the lock-file change and `terraform plan`.
5. Test significant minor and all major upgrades in a non-production environment.
6. Follow the major-version upgrade guide before applying.

An exact stable version may be appropriate for highly controlled environments, but it increases
the need for a regular upgrade process. Reusable child modules should declare only the minimum
provider version they require and should not unnecessarily constrain all callers to a narrow
release range.

See HashiCorp's documentation for
[provider requirements](https://developer.hashicorp.com/terraform/language/providers/requirements)
and the
[dependency lock file](https://developer.hashicorp.com/terraform/language/files/dependency-lock).

## Version availability notification

The provider may perform a best-effort check for a newer stable version at startup. This mechanism
is advisory:

- It may emit a non-blocking diagnostic containing the installed version, latest stable version,
  release type, and a link to release notes.
- A new major version may use more prominent wording, but it must not stop `plan` or `apply`.
- It must not claim that every newer release is mandatory; relevance depends on the resources and
  functionality used.
- Registry unavailability, timeouts, or malformed responses must never block provider operation.
- Prereleases must not be presented as the latest stable version.
- Customers must not need an override environment variable merely to continue using a valid
  pinned version.

The notification is not a substitute for version constraints, the dependency lock file, release
notes, KB announcements, or direct communication about an API end-of-life event.

## Release notes and communication

Every stable release must have a concise, practitioner-focused changelog using the applicable
sections:

- **Breaking changes** - major releases only, with a migration link;
- **Deprecations** - replacement and earliest removal version/date;
- **Features** - new resources, data sources, and significant capabilities;
- **Enhancements** - smaller backward-compatible additions;
- **Bug fixes**;
- **Security**;
- **Notes** - important operational or upstream API information.

Each entry should identify the affected provider component and explain customer impact. Internal
refactoring, tests, and build-system maintenance should be omitted unless they materially affect
provider users.

### Communication by release type

- **Major:** preannounce, publish a dedicated upgrade guide, include complete breaking-change and
  migration details, publish KB/RN, and notify affected customers through support channels.
- **Minor:** publish release notes; use KB/RN for notable capabilities, deprecations, or upcoming
  API deadlines.
- **Patch:** publish release notes; use KB/RN for security, urgent compatibility, or
  high-customer-impact fixes.
- **Beta:** provide targeted notes directly to participating users; do not announce as generally
  available.

### Current internal process

- Post the release summary to `#rn-kb-opensource-api-announcements` and tag
  a PM and a technical writer when KB/RN coordination is required.
- Knowledge-transfer meetings take place on Thursday. A change known by Thursday of week X can
  have its release note reviewed in week X+1, with the release note and change published no
  earlier than Sunday of week X+2.
- Urgent compatibility or security releases may use an accelerated process, but must still
  include complete release notes and explicit customer impact.

## Release readiness

Before publishing:

### Patch

- Targeted unit and acceptance tests pass.
- The change is backward-compatible with valid documented configuration and state.
- Release notes explain any newly visible drift or plan behavior.

### Minor

- Patch criteria pass.
- New functionality has documentation and examples.
- Existing supported functionality passes regression testing.
- Deprecations provide a replacement and migration guidance.

### Major

- Minor criteria pass.
- Every breaking change is listed in the upgrade guide.
- State migration and import behavior are tested.
- Representative existing configurations are tested across the upgrade boundary.
- The announcement, KB/RN, and support readiness are complete.
- The previous major's maintenance or end-of-support treatment is stated explicitly. This
  versioning policy does not by itself create a backport or support SLA.

### Beta

- The validation scope, participants, risks, and exit criteria are documented.
- Participants receive exact pinning and rollback instructions.
- Feedback is incorporated into a new beta number or the intended stable release.

## Industry alignment

| Provider or standard | Observed practice | What Cato adopts |
|---|---|---|
| **HashiCorp provider guidance** | Patch for functionally equivalent fixes, minor for backward-compatible features and deprecations, major for breaking changes; major releases recommended no more than annually | This policy uses the same compatibility contract and changelog categories |
| **AWS provider** | Weekly releases, best-effort backward compatibility within a major, version pinning recommended, breaking changes concentrated in major releases, and user-focused changelog categories | Frequent minor/patch delivery without weakening the major compatibility boundary |
| **AzureRM provider** | Frequent minor releases combine enhancements and bug fixes; major versions use dedicated upgrade guidance | Features remain minor, fixes remain patch, and breaking migrations receive an upgrade guide |
| **Google provider** | Breaking changes are scheduled and publicized in major releases; users are advised to constrain versions and follow an upgrade guide; preview functionality is clearly separated | Planned major announcements, explicit constraints, migration guides, and clearly identified prereleases |

## References

- [HashiCorp: Versioning and changelogs](https://developer.hashicorp.com/terraform/plugin/best-practices/versioning)
- [HashiCorp: Provider requirements](https://developer.hashicorp.com/terraform/language/providers/requirements)
- [HashiCorp: Dependency lock file](https://developer.hashicorp.com/terraform/language/files/dependency-lock)
- [HashiCorp: Publishing providers](https://developer.hashicorp.com/terraform/registry/providers/publishing)
- [AWS Provider: Backward compatibility and release cadence](https://hashicorp.github.io/terraform-provider-aws/faq/)
- [AWS Provider: Changelog process](https://hashicorp.github.io/terraform-provider-aws/changelog-process/)
- [AzureRM Provider releases](https://github.com/hashicorp/terraform-provider-azurerm/releases)
- [AzureRM 4.0 upgrade guide](https://registry.terraform.io/providers/hashicorp/azurerm/4.0.0/docs/guides/4.0-upgrade-guide)
- [Google Provider: Version 5 announcement and upgrade guidance](https://github.com/hashicorp/terraform-provider-google/issues/15582)
- [Google Provider repository and GA/beta model](https://github.com/hashicorp/terraform-provider-google)
- [Semantic Versioning 2.0.0](https://semver.org/)
