# Tennis Arena — Setup (run AWK v0.14.0 on this spec)

This project is an **empty scaffold seeded with a spec**. Follow these steps to
have AWK read the spec and build the game from 0→1. The spec lives at
`.ai/specs/tennis-arena/` (`requirements.md`, `design.md`, `tasks.md`).

## 0. Prerequisites

- `awkit` v0.14.0 installed (`awkit version`)
- `gh` authenticated (`gh auth login`) — AWK uses GitHub issues/PRs as its state machine
- A Worker backend on PATH: `codex` (default) **or** `claude` (if you set
  `worker.backend: claude-code`)
- `git`, `go` 1.25+, Node 20+ (the game's own toolchains)
- (later tasks) Docker + docker-compose for the deployment steps

## 1. Initialize the project

```bash
cd D:/projects/tennis-arena

git init
git add .
git commit -m "chore: seed tennis-arena spec"

# Create the GitHub repo AWK will drive (issues/PRs live here)
gh repo create tennis-arena --private --source=. --push
```

## 2. Install AWK with the react-go monorepo scaffold

```bash
# React (frontend) + Go (backend) monorepo scaffold + .ai/ kit files
awkit init --preset react-go --scaffold
```

This creates `backend/` (Go), `frontend/` (React/TS), `.ai/config/`, rules,
skills, and a starter `.ai/config/workflow.yaml`. Your seeded
`.ai/specs/tennis-arena/` is preserved.

> If `awkit init` ever reports it would overwrite the spec, re-copy
> `.ai/specs/tennis-arena/` back in afterward — it is the source of truth.

## 3. Configure `workflow.yaml`

Replace the generated `.ai/config/workflow.yaml` with **`workflow.recommended.yaml`**
from this repo (or merge its key settings). The important parts:

- **Activate the spec**: `specs.active: ["tennis-arena"]`
- **Repos + verify commands** (what the review evidence gate re-runs):
  - `backend`: `go build ./...` / `go test ./...`
  - `frontend`: `npm ci` (setup) / `npm test` (vitest)
- **Turn on the v0.14.0 capabilities you want to exercise** (see below).

```bash
cp workflow.recommended.yaml .ai/config/workflow.yaml
awkit validate
awkit evaluate --offline
```

## 4. Kick off

```bash
awkit kickoff            # or: awkit kickoff --dry-run  to preview first
```

AWK will now, task by task (37 of them, in dependency order):
1. `analyze-next` → sees an uncompleted task in `tasks.md`
2. `create-task` → writes a ticket (Summary/Scope/Acceptance Criteria/Testing) from
   the task line + `design.md`, and opens a GitHub issue
3. `dispatch-worker` → the Worker implements it and runs the verify commands
4. `review_pr` → the pr-reviewer subagent submits a **structured `review.json`**;
   the **evidence gate** re-runs the tests and checks each acceptance criterion
   maps to a passing test with a real assertion; **multi-model consensus** and the
   **severity gate** apply
5. merge or send back — and every rejection is **distilled into a lesson** that is
   injected into the next Worker's prompt

Stop anytime with `touch .ai/state/STOP` (or `awkit stop-workflow`).

## 5. Watch v0.14.0 do its thing

```bash
awkit lessons list        # lessons accumulating from review rejections
awkit lessons stats       # hit/miss + candidate→active→proven progression
awkit events              # unified event stream incl. session_usage (token/cost)
gh pr list                # PRs opened/reviewed/merged
gh issue list --label ai-task
```

Signs the marquee features are working:
- **Learning loop**: after a few rejections, `awkit lessons list` shows lessons; the
  same class of mistake stops recurring in later tasks.
- **ACI structured review**: PR review comments are rendered from `review.json`; a
  reviewer format slip shows `SUBMISSION INVALID` and is fixed in-session (not a
  wasted review round).
- **Multi-model consensus**: review comments carry a consensus section (if you set
  `review.multi_model: true`).
- **Knowledge-graph grounding**: if you run `/understand` (Understand-Anything) in
  the repo to produce `.understand-anything/knowledge-graph.json`, later Worker
  prompts get a CODEBASE MAP of the files they touch.

## Notes & tips

- **Frontend test runner**: the react scaffold may not ship vitest. Task 18 (the
  first frontend task) is expected to establish `npm test` (vitest) so later
  frontend tasks verify. Keep `frontend` `verify.test: "npm test"`.
- **Determinism first**: Tasks 1–10 build the deterministic sim + golden vectors.
  They are the foundation; let them go green before the netcode/client tasks.
- **Postgres**: Task 24 introduces migrations. For local verification you can point
  tests at a disposable Postgres (docker) or an in-memory fake behind the repo
  interface — the design allows both.
- **Cost**: this is a large, real 0→1 build (37 tasks, many with retries). Watch
  `awkit events` / `awkit lessons stats`; use `awkit kickoff --dry-run` first.
- **Parallelism/ordering**: independent tasks (e.g. 18, 22, 23, 24) can proceed in
  parallel; the `Depends on` metadata in `tasks.md` + the table in `design.md`
  encode the order.
