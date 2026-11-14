---
name: diagnose-and-generate-names
description: Naming diagnosis and generation for brands, products, characters, places and titles. Use when a name feels off or a set of names needs a systematic method.
license: MIT
metadata:
  blacktower-pack: creative
  blacktower-kind: workflow
  blacktower-tier: standard
  blacktower-capabilities: worldbuilding-fiction,naming
  author: jwynia
  version: "1.0"
  type: diagnostic
  mode: diagnostic+generative
  domain: creativity
---

# diagnose-and-generate-names

## Use this skill

Diagnose why names don't work and guide creation of names that do. Use for brand names, product names, character names, place names, and titles when something feels off or when systematic naming is needed.

## Workflow

1. Confirm the request matches this skill's stated intent; do not activate it
   for a keyword-only near miss.
2. Inspect the user's inputs, repository conventions, and available tools before
   selecting an approach.
3. Load only the relevant section of the detailed guide when the task needs
   domain-specific patterns or edge cases.
4. Perform the smallest complete unit of work and preserve existing behavior or
   user data unless change is explicitly requested.
5. Validate the result with the project's own checks plus any skill-specific
   verification described in the detailed guide.

## Constraints

- Prefer the project's existing stack and conventions over generic examples.
- State required credentials, network access, CLIs, and destructive side effects
  before relying on them.
- Use current authoritative documentation when provider or framework behavior is
  version-sensitive.
- Report concrete evidence, remaining risks, and any unverified assumptions.

## Resources

- `references/detailed-guide.md` — detailed framework and edge cases; load only the sections needed for the task.
- `scripts/` — deterministic helpers; inspect --help and side effects before use.
- `templates/` — reusable output skeletons.

Load [companion procedures](references/skill-routes.md) only when their stated condition applies.
