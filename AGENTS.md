# Agent Guidance

> **Purpose.** This file is the primary onboarding and guardrail document for any LLM
> (Claude, GPT, Gemini, Cursor, Copilot, etc.) that reads, writes, or reviews code in
> this repository. Read it before making changes. It is intentionally rule-oriented and
> self-contained.

## Git

Split commit message to a meaningful scope!

**Commit message format**

- Prefix with exactly one of:
  - `[fix]` — fixes a bug
  - `[feature]` — adds new functionality
  - `[improve]` — improves existing functionality
  - `[meta]` — changes outside the code base (e.g. deployment setup)
  - `[docs]` — documentation (README, these docs, etc.)
  - `[refactor]` — formatting, renaming, structural-only changes
- Capitalize the first letter after the prefix.
- **Title only — no body.**
