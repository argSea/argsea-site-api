# AGENTS.md

## Purpose
The argSea portfolio/blog backend API: a Go HTTP service exposing portfolio and
blog resources. It owns the API surface and its persistence; it is not the
frontend/site and does not own deployment infrastructure.

## Instruction Priority
Resolve instructions in this order:
1. an external session path assigned by the caravan primary integrator
2. local `SESSION.md`, if present
3. this `AGENTS.md`
4. task-relevant repo docs
5. source code
If instructions conflict, pause and ask.

## Boot Sequence
1. `AGENTS.md`
2. assigned external session path, if one was given
3. local `SESSION.md`, if present
4. `main.go` (wiring/entrypoint) and the relevant `argHex/` port+adapter pair
5. only the files the task requires
Read narrowly. Do not wander the repo.

## Hierarchical Workflow
- An assigned external session is the authoritative task contract.
- Keep edits scoped to this repo unless the session explicitly allows more.
- Prefer a worktree over the primary checkout for a new branch.
- Branch name: `type/scope/short-desc`.
- Return implementation + verification evidence to the primary integrator.

## Operating Rules
- Stay inside the declared scope and exclusions.
- Preserve existing behavior unless the task changes it.
- Keep diffs reviewable and tied to the task.
- Update durable docs only when architecture/contracts materially change.
- Plain English in responses and session notes.
- This is a hexagonal (ports/adapters) codebase: a new capability adds an
  interface in the relevant `*_port` package and an implementation in the
  matching `*_adapter` package; don't collapse the two.
- `config.json` is local-only (gitignored). Never commit it or secrets.

## Repo Map
- `main.go` (entrypoint): parses `--config`/`--log`, wires adapters→services→
  routes (gorilla/mux, viper for config).
- `argHex/` (the hexagon):
  - `domain/`, `data_objects/`: core types and DTOs.
  - `in_port/`, `in_adapter/`: inbound interfaces and their HTTP handlers.
  - `out_port/`, `out_adapter/`: outbound interfaces and their implementations.
  - `service/`: application/business logic.
  - `stores/`: persistence wiring.
  - `utility/`: shared helpers.

## Architecture Defaults
Ports & adapters. Dependencies point inward: adapters depend on ports, services
depend on ports, never the reverse. Keep HTTP/transport concerns in adapters.

## Verification Rules
For touched behavior, run the smallest useful check for the changed surface:
- `go build ./...`: compiles the module.
- `go vet ./...`: static checks.
- `go test ./...`: tests (add coverage for changed behavior).
- `gofmt -l .`: formatting (should print nothing).
If an expected command doesn't exist, say so. Don't claim a run you didn't do.

## Session Discipline
- Small tasks: one agent.
- Multi-agent: the assigned external session (or local `SESSION.md`) is the
  parent contract; one primary integrator owns consolidation.
- With no session and no explicit implementation request, stay in planning mode.

## Final Output Expectations
Report: what changed, files changed, verification run, known limitations/
follow-ups, and assumptions a human should review. Plain English.
