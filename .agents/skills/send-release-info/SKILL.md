---
name: send-release-info
description: Use when the user asks to send/announce release info or release notes for the terraform-provider-cato to Slack. It posts a release announcement to #rn-kb-opensource-api-announcements using the standard release format and then posts a link to that announcement in #tf-internal.
---

# Terraform Provider Cato Send Release Info Skill

## Repository Boundary

Before using this skill:

1. Confirm the current workspace is the `terraform-provider-cato` repository.
2. If the workspace is not this repository, do not apply this skill. Continue with general agent behavior and mention that the skill is repo-scoped.

## Goal

Announce a released version of the terraform provider on Slack. This skill:

1. Posts a release announcement to `#rn-kb-opensource-api-announcements`.
2. Posts a link to that announcement in `#tf-internal`.

This skill only communicates a release that already exists. It does not bump
versions, edit the changelog, tag, or publish — use `prepare-release` and
`publish-release` for those steps.

## Slack Targets

- Announcement channel `#rn-kb-opensource-api-announcements` — channel ID `C08QPHRURP1`.
- Internal notification channel `#tf-internal` — resolve the channel ID at runtime
  with `slack_search_channels` (search `public_channel,private_channel`). If it
  cannot be resolved, ask the user for the channel ID before posting.

## Required Grounding

Before drafting the message, gather the release facts from the repository, not
from memory:

1. Determine the version to announce. Default to the latest (topmost) entry in
   `changelog.md`. If the user names a version, use that one and confirm it
   exists in `changelog.md`.
2. Read that version's `changelog.md` entry to get the release date (the
   `## X.Y.Z (YYYY-MM-DD)` header) and the list of changes.
3. Select the most notable 2–4 highlights from that entry (prefer `Added` and
   user-facing `Fixed`/`Changed` items; omit purely internal `Tests`/chore
   items unless nothing else is notable).

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
- Keep highlights concise; reuse the changelog wording, trimming trailing
  implementation detail where helpful.
- Keep the two standard links unchanged.

## Workflow

Follow this order:

1. Complete the Required Grounding above and build the announcement text from the
   template.
2. Present the drafted message to the user with `slack_send_message_draft` for
   `#rn-kb-opensource-api-announcements` (`C08QPHRURP1`) and get confirmation.
3. On approval, send it to `#rn-kb-opensource-api-announcements` with
   `slack_send_message` and capture the returned message permalink.
4. Resolve the `#tf-internal` channel ID (see Slack Targets).
5. Post to `#tf-internal` a short message linking to the announcement, e.g.
   `Terraform Provider Cato vX.Y.Z released — announcement: <permalink>`.
6. Report both message links back to the user.

## Constraints

Always:

- Base the version, date, and highlights on `changelog.md`, not assumptions.
- Get user approval of the announcement text before posting to the public
  announcement channel.
- Capture and reuse the real permalink of the posted announcement for the
  `#tf-internal` message.

Never:

- Post to `#tf-internal` before the announcement has been posted (the internal
  message must link to the real announcement).
- Invent a channel ID for `#tf-internal`; resolve it or ask the user.
- Edit repository files, tags, or releases as part of this skill.

## Examples

User request that should trigger this skill:

```text
Send the release info for the latest terraform provider version to Slack.
```
