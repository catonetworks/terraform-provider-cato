---
name: publish-release
description: Use when the user asks to publish the release of a new prepared version of the terraform provider living in this repository. It will create a new release in the repository with the appropriate tag and release notes based on the changelog entry created during the prepare-release step.
---

# Terraform Provider Cato Publish Release Skill

## Repository Boundary

Before using this skill:

1. Confirm the current workspace is the `terraform-provider-cato` repository.
2. If the workspace is not this repository, do not apply this skill. Continue with general agent behavior and mention that the skill is repo-scoped.
3. If the last commit in the current branch is not a release commit, do not apply this skill. Instead, ask the user if they want to prepare a new release.

## Goal

Use this skill to publish the release of the latest prepared version of the terraform provider by following the repository's release process.

## Required Grounding

Before editing or advising:

1. Make sure the workspace is on the correct branch for releasing (e.g., `main` or `master`).
2. Make sure the last commit in the current branch is a release commit with changes in the changelog file and a commit message indicating the release version.

## Workflow

Follow this order:

1. Identify the last release commit.
2. Check if the version within the release commit message has already been published by looking for the version tag on the remote repository. If it has, inform the user that the release is already published and do not proceed with publishing again.
3. Check if the version within the release commit message exists as a tag in the local repository. If it does, notify the user that a tag with the release version already exists locally and ask if they want to push the existing tag to the remote repository.
4. If the version does not exist as a tag in the local repository, create a new tag with the release version from the release commit in the format `vX.Y.Z`.
5. Push the release tag to the remote repository.

## Constraints

Always:

- Make sure the tag versions points to the release commit with the correct version in the commit message.

Never:

- Never create a new release commit or modify existing release commits. This skill is only responsible for publishing existing release commits by creating and pushing tags to the remote repository.


## Examples

User request that should trigger this skill:

```text
Publish the latest created version of the terraform provider with the latest changes.
```
