# outline-to-chapter procedure

Reviewed source clients: claude, codex, opencode. Check the actual client tools before following this procedure. Resolve its scripts, references and assets from this directory, not from the outer skill root.

# outline-to-chapter

## Use this skill

Autonomously draft and polish chapters through multi-skill editorial passes. Use when you have a complete outline and want to produce a polished first draft with iterative refinement.

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

## Companion skills

Load [required steps and optional helpers](references/skill-routes.md) only when their stated condition applies.
