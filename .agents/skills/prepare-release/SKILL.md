---
name: prepare-release
description: Use when the user asks to prepare the release of a new version of the terraform provider living in this repository.
---

# Terraform Provider Cato Prepare Release Skill

## Repository Boundary

Before using this skill:

1. Confirm the current workspace is the `terraform-provider-cato` repository.
2. If the workspace is not this repository, do not apply this skill. Continue with general agent behavior and mention that the skill is repo-scoped.
3. If the last commit in the current branch is a release commit, do not apply this skill. Instead, ask the user if they want to release real changes or if they want to update the release notes for the existing release commit.

## Goal

Use this skill to prepare the release of a new version of the terraform provider by following the repository's release process.

## Required Grounding

Before editing or advising:

1. Make sure there are no uncommitted changes in the workspace.
2. Make sure the workspace is on the correct branch for releasing (e.g., `main` or `master`).

## Workflow

Follow this order:

1. Identify the last release commit and the changes since that commit.
2. Determine the next version number based on the changes since the last release following a semantic versioning approach. If the user specified a version number, give it priority and validate that it is greater than the last release version.
3. Create a new local branch for the release (e.g., `release/vX.Y.Z`).
3. Update the version number in the appropriate files: `GNUmakefile`, and `main.go`.
4. Add a new entry to the changelog at the top of the file with the following suggestions:
   - Use the new version number and the current date in the format `YYYY-MM-DD` for the changelog entry header.
   - Use the messages from the commits since the last release to populate the changelog entry. If there are multiple commits, group them into categories (e.g., "Added", "Changed", "Fixed") based on the content of the commit messages. If the commit messages do not clearly indicate the type of change, you can use your judgment to categorize them appropriately.
5. Commit the changes with a message like `Release version X.Y.Z`.
6. Push the release branch to the remote repository.
7. If the github command line is available, use it to create a pull request from the release branch to the main branch with a title like `Release vX.Y.Z` and a description that includes the changelog entry for the new version.

## Constraints

Always:

- Make sure the version number is updated in both `GNUmakefile` and `main.go` files.
- Make sure the changelog entry is well-formatted and includes all relevant changes since the last release.
- Ask the user for confirmation while grouping commits into changelog categories if the commit messages are not clear.


Never:

- Modify commit changes to the main or master branch.
- Never approve or merge the release pull request. Only create the pull request and leave it to the user to review, approve, and merge.


## Examples

User request that should trigger this skill:

```text
Release a new version of the terraform provider with the latest changes.
```

Expected changes in code:

```diff
diff --git a/GNUmakefile b/GNUmakefile
index 713bb2e..afc68bf 100755
--- a/GNUmakefile
+++ b/GNUmakefile
@@ -4,7 +4,7 @@ NAMESPACE=catonetworks
 PKG_NAME=cato
 BINARY=terraform-provider-${PKG_NAME}
 # Whenever bumping provider version, please update the version in cato/client.go (line 27) as well.
-VERSION=0.0.73
+VERSION=0.0.75

diff --git a/changelog.md b/changelog.md
index bbafd82..2bfc61c 100644
--- a/changelog.md
+++ b/changelog.md
@@ -1,5 +1,13 @@
 # Changelog

+## 0.0.75 (2026-05-18)
+- Fixed license handling for accounts with more than 1,000 sites and added defensive unit coverage.
+- Fixed `translated_subnet` handling for network range, LAN interface, and socket site native ranges to submit nil values correctly when unset.
+- Added broader Terraform acceptance test coverage across provider resources and cleanup workflows.
+- Updated the Cato Go SDK dependency.
+- Hardened socket site update flows with bounded retries for transient backend conflicts and improved connection type hydration.
+- Updated internet and WAN firewall rule hydration to send empty API objects/lists instead of null values for service and action configuration fields.
+

diff --git a/main.go b/main.go
index 37d5587..24ec010 100644
--- a/main.go
+++ b/main.go
@@ -11,7 +11,7 @@ import (
 )

 var (
-       version string = "0.0.15"
+       version string = "0.0.75"
 )

```