---
name: send-release-info
description: Announce a released Cato Terraform provider version on Slack, or send a private test announcement. Normal announcements go to #rn-kb-opensource-api-announcements with a link in #management-integrations; tests go only to the requested recipient.
---

# Terraform Provider Cato Send Release Info Skill

## Repository Boundary

Before using this skill:

1. Confirm the current workspace is the `terraform-provider-cato` repository.
2. If the workspace is not this repository, do not apply this skill. Continue with general agent behavior and mention that the skill is repo-scoped.

## Goal

Announce a released version of the Cato Terraform provider on Slack. This skill:

1. Posts a release announcement to `#rn-kb-opensource-api-announcements`.
2. Posts a link to that announcement in `#management-integrations`.

For a requested test, send the announcement only to the requested recipient,
label it as a test, and skip both channel posts.

This skill only communicates a release that already exists. It does not bump
versions, edit the changelog, tag, or publish — use `prepare-release` and
`publish-release` for those steps.

## Slack Targets

- Announcement channel `#rn-kb-opensource-api-announcements` — channel ID `C08QPHRURP1`.
- Internal notification channel `#management-integrations` — channel ID `C0A7P8V7WFR`.

## Required Grounding

Before drafting the message, gather the release facts from the repository, not
from memory:

1. Determine the version to announce. Default to the latest released version
   represented in `changelog.md`; verify the actual release/tag. If the user
   names a version, preserve it exactly, including any prerelease suffix.
2. Read that version's changelog and release metadata for the date and changes.
   A prerelease tag may contain a changelog headed with the target stable version
   (for example, `v2.0.0-prerelease.1` contains a `2.0.0` entry). In that case,
   verify the exact released tag and use the changelog at that tag; take the
   publication date from release metadata. Do not announce the stable version
   as released or substitute the working tree's newer content.
3. Select the most notable 2–4 highlights from that entry (prefer `Added` and
   user-facing `Fixed`/`Changed` items; omit purely internal `Tests`/chore
   items unless nothing else is notable).
4. Rewrite each selected item for customers:
   - Say what users gain or what customer-visible problem is avoided.
   - Name affected resources or workflows when relevant.
   - Do not expose internal Terraform state, planning, API implementation, or
     dependency details unless they explain the user impact.
   - Refer to the product as the "Cato Terraform provider", never as a
     "Terraform module".
   - Describe profile authentication as "named credential profiles" or "named
     profiles in a credentials file". Do not call it "catocli profiles", "CLI
     authentication", or add "without requiring the Cato CLI". The provider's
     authentication feature must not imply a dependency on the CLI or community
     tooling.
   - Do not use a Go SDK version as a highlight by itself. If the update keeps
     the provider aligned with the latest public API schema, say so and explain
     the practical impact. Example: `Updated the provider to the latest API
     schema. This is typically a routine update; in rare cases, staying on an
     older provider version could cause API call errors due to schema
     mismatches.` Do not claim fixes, performance gains, or new capabilities
     without evidence.

## Provider Resources, Modules, and Major Releases

- Separate core provider resources from separately published Terraform modules.
  Modules build on provider resources; they are not resources bundled in the
  provider release. Do not imply that community-maintained modules have the
  same ownership or support commitment as core provider resources.
- List new resource types and, when modules are mentioned, list those modules
  separately. Give each resource a short, customer-facing description of what
  it covers, grounded in its documentation at the announced tag.
  Keep the main announcement concise; put long inventories in a
  thread or link to a verified inventory. Derive resource names/counts from
  registered resources at the release tag versus the previous stable tag, not
  all source files or documentation files (which may include unregistered types).
  Use published module sources for module names/counts; do not invent a list
  from the provider changelog.
- Lead with what users can manage and the number of added resources when useful.
  Omit historical coverage totals such as "from 55 to 112".
- For a major release, call out breaking changes and link to the upgrade guide
  at the announced tag. Do not invent a 1.x support or end-of-support policy.
- Major releases need coordination and knowledge transfer with Product,
  Knowledge/technical writers, and Support before going public. Surface this
  preparation item to the user; do not schedule meetings or contact participants
  unless asked. Any exception agreed for a beta does not automatically apply to
  stable releases.

For `2.0.0-prerelease.1`, the approved release-note framing is: 57 added provider
resources; named credential profiles; broader policy, Cloud Interconnect,
secondary AWS/Azure vSocket, and Wi-Fi SSID coverage. The 18 separately published
policy modules may be mentioned as a layer built on these resources, clearly
separated from the provider changes. These counts are specific to this release.

## Message Format

Match the format of the existing announcements, e.g.
`https://canu.slack.com/archives/C08QPHRURP1/p1782466704793209`.

Use this template (Slack markdown: italic title, `code` for version/date,
bulleted highlights, and named links):

```text
_Terraform Provider Cato `vX.Y.Z` is available_

Released: `YYYY-MM-DD`

Highlights:
- <highlight 1>
- <highlight 2>

Links:
- [Provider registry](https://registry.terraform.io/providers/catonetworks/cato/latest)
- [Changelog](https://github.com/catonetworks/terraform-provider-cato/blob/main/changelog.md)
```

Notes:
- Prefix the version with `v` (e.g. `v0.0.90`).
- Keep highlights concise, customer-facing, and actionable; reuse changelog
  wording only after trimming or rewriting implementation detail.
- Retain the provider registry and changelog links. For a prerelease, include
  the exact release link and tagged upgrade guide; `latest` alone does not
  identify the beta.

### Prereleases / Beta

Use "Beta" and "available for testing" explicitly. State that customers can
select the public beta by pinning the exact version, for example
`version = "= 2.0.0-prerelease.1"`; normal latest/stable selection does not
automatically select a prerelease. Recommend a non-production account and
mention that API availability depends on enabled account features. Do not
describe the beta as stable or generally available merely because the GitHub
release's prerelease checkbox is unset; preserve the semantic version suffix.

Suggested framing:

```text
_Cato Terraform provider v2 Beta — `v2.0.0-prerelease.1` is available for testing_

Released: `2026-11-16`

Manage more of your Cato configuration with 57 added provider resources and
authenticate using named credential profiles.

Provider capabilities:
- Expanded coverage includes policies, Cloud Interconnect, secondary AWS/Azure vSockets, and Wi-Fi SSIDs.
- Named profiles in a credentials file support authentication.

Separately published modules:
- 18 policy modules build on the added provider resources. <verified inventory link>

Testing and upgrades:
- Pin `version = "= 2.0.0-prerelease.1"` in a non-production account. Latest/stable selection does not automatically select this beta.
- Review the tagged upgrade guide for breaking changes. API availability depends on enabled account features.

Links: <provider registry>, <changelog>, <exact prerelease>, <tagged upgrade guide>
New provider resources: <verified inventory link or thread>
```

## Workflow

Follow this order:

1. Complete the Required Grounding above, prepare the separate inventories, and
   build the announcement text from the appropriate template.
   **Test mode:** if the user requests a test message to themselves or another
   specified recipient, resolve that Slack identity, send the labeled test
   directly using `slack_send_message`, and return its permalink. Send long
   inventories as replies in that same DM thread. A requested test send is
   already authorized; do not ask for public-announcement approval or proceed
   to the channel workflow below.
2. Present the drafted message to the user with `slack_send_message_draft` for
   `#rn-kb-opensource-api-announcements` (`C08QPHRURP1`) and get confirmation.
3. On approval, send it to `#rn-kb-opensource-api-announcements` with
   `slack_send_message` and capture the returned message permalink.
4. Post to `#management-integrations` (`C0A7P8V7WFR`) a short message linking to the announcement, e.g.
   `Terraform Provider Cato vX.Y.Z released — announcement: <permalink>`.
5. Report both message links back to the user.

## Constraints

Always:

- Base the exact version, date, and highlights on the released tag's changelog
  and release metadata, not assumptions.
- Ensure the announcement explains customer impact rather than internal causes.
- Get user approval of the announcement text before posting to the public
  announcement channel.
- Capture and reuse the real permalink of the posted announcement for the
  `#management-integrations` message.

Never:

- Post to `#management-integrations` before the announcement has been posted (the internal
  message must link to the real announcement).
- Edit repository files, tags, or releases as part of this skill.

## Examples

User request that should trigger this skill:

```text
Send the release info for the latest terraform provider version to Slack.
```
