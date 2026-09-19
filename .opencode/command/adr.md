---
description: Draft a new numbered ADR in docs/adr for the given decision
---

Draft a new Architecture Decision Record for: $ARGUMENTS

1. Find the next free number: list `docs/adr/` and take N+1 from the highest
   existing `NNNN-*.md`.
2. Read one existing ADR (e.g. docs/adr/0001-date-range-windows-3am-cutoff.md)
   and match its exact format: H1 `# ADR NNNN — Title`, a Status/Date header
   block, then Context, Decision, and Consequences sections. State the current
   design and its rationale — do not narrate past bugs or history.
3. Only write an ADR for a decision with lasting impact (schema, filtering
   semantics, storage formats, new surfaces, external contracts). If the topic
   is too small for an ADR, say so and stop instead of writing a trivial one.
4. Before writing, check whether existing ADRs already cover or contradict the
   topic; if it amends an existing one, update that ADR's Consequences instead
   of duplicating.
5. Cross-reference related ADRs by number where relevant.

Do not commit unless asked.
