# Base procedure

Reviewed source clients: claude, codex, opencode.

Resolve command and resource paths from the owning skill root, unless a linked reference says otherwise.

# fiction-line-edit

## Use this skill

Diagnose sentence-level issues after structure is solid. Use when prose feels flat, sentences are monotonous, word choices are generic, or voice is inconsistent.

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
- `scripts/` — deterministic helpers; inspect --help and side effects before use. They use Deno APIs (`deno run --allow-read scripts/prose-check.ts <file>`), so Bun or Node will not run them. If `deno` is missing, ask before installing it (for example `mise use deno`) or diagnose by hand.
