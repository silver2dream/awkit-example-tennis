# Tennis Arena — Design

## Overview

Tennis Arena is a **server-authoritative**, real-time 1-v-1 online tennis game.
The heart of the system is a **deterministic simulation core** that both the Go
server and the TypeScript client run, enabling client-side prediction with exact
server reconciliation, input-log replays, and — crucially for this project —
**full headless verification** of gameplay by automated tests.

This document is the architectural contract every task builds against. It is
written to be **agent-friendly (ACI)**: it fixes the module boundaries, the data
shapes, the determinism policy, and the verification method up front, so each
task is a small, unambiguous, independently-testable slice.

## Repository layout (monorepo, `directory` type)

```
backend/    # Go — authoritative server, deterministic sim, netcode, services, persistence
  testdata/vectors/   # CANONICAL golden vectors (JSON) — Go loads natively via testdata
frontend/   # TypeScript + React + Canvas — thin client (render, input, predict, interpolate, UI)
  scripts/sync-vectors.mjs   # copies ../backend/testdata/vectors -> src/sim/__vectors__ before tests
```

- `backend/`: `go build ./...`, `go test ./...`
- `frontend/`: `npm ci && npm run build`, `npm test` (vitest; `pretest` runs sync-vectors)

**Why this layout (ACI + AWK scope safety):** each worker's changes stay inside
its own repo subdirectory (AWK's `directory` type enforces this). The golden
vectors are **owned by `backend/testdata/vectors/`** (Go's native fixture
location); the frontend keeps its determinism honest by **syncing a local copy**
and asserting against it, and the checkpoint (Step 37) verifies both sides agree.
There is intentionally **no cross-repo `shared/` directory** to avoid boundary
violations.

## Technology decisions

| Concern | Choice | Why |
|---|---|---|
| Authoritative server + sim | **Go** | deterministic, testable, strong concurrency for a game server |
| Client | **TypeScript + React + Canvas 2D** | React for lobby/HUD, canvas for the court view; vitest for logic tests |
| Determinism | **fixed-point (Q32.32) integer math** in the sim | bit-exact across OS/CPU/language; golden vectors become trivially checkable |
| Transport | **`Transport` interface** (in-memory + WebSocket impls) | run whole matches headless in tests; swap real sockets for prod |
| Persistence | **PostgreSQL** + SQL migrations | matches, users, ratings, replays |
| Ranking | **Glicko-2** (Elo acceptable if simpler) | standard, testable via golden vectors |
| Observability | **slog** + **Prometheus** | structured logs, metrics, health |
| Delivery | **Docker / docker-compose** | server + Postgres + Prometheus, env config |

## ACI & determinism strategy (the load-bearing idea)

1. **Fixed timestep, fixed-point sim.** The authoritative sim advances at 60 Hz
   ticks with 4 physics sub-steps (240 Hz). All gameplay state uses **Q32.32
   fixed-point** (`type Fixed int64`), so `Step` is bit-exact everywhere. Floats
   are allowed only in rendering.
2. **`Step(state, inputs) -> state` is pure.** No clocks, no goroutines, no
   globals inside the sim. This is what makes prediction/reconciliation exact and
   the whole game unit-testable.
3. **Golden vectors.** `backend/testdata/vectors/*.json` hold `{seed, initialState,
   inputs[], expectedState}` fixtures for physics, bounce, spin, scoring, delta
   encoding, and ranking. Tests replay them and assert **exact** equality. The Go
   sim and the TS prediction sim load the **same** vectors → cross-language
   determinism is enforced, not assumed.
4. **Transport-agnostic netcode.** Server and client talk through a `Transport`
   interface. An **in-memory transport** with a deterministic scheduler lets a
   test run server + two clients for a full match in-process and assert the final
   score — no sockets, no flakiness.
5. **Rendering behind an interface.** The client depends on a `Renderer`
   interface; tests inject a `MockRenderer` and assert the **scene** handed to it
   (positions, score) rather than pixels.
6. **Every acceptance criterion → a named test + real assertion.** Tasks state
   the test name and the assertion so the review evidence gate can confirm them.

## Physics model

State (fixed-point): ball `position(x,y,z)`, `velocity`, `angularVelocity(spin)`;
players `position`, `velocity`, `stance`.

Per physics sub-step (dt = 1/240 s), integrate with **semi-implicit Euler**:

- **Gravity**: `v.y -= g * dt`, `g = 9.81 m/s²`.
- **Drag**: `F_d = -½ · ρ · C_d · A · |v| · v` (quadratic; opposes motion).
- **Magnus**: `F_m = ½ · ρ · C_l · A · |v|² · (ŵ × v̂)` (spin curves the path).
- Integrate `v += (F/m) · dt`, then `position += v · dt`.

Constants (centralized in `backend/sim/constants`; the TS port reproduces them and is checked against the same vectors):
ball mass `0.057 kg`, diameter `0.067 m` (radius → area `A`), air density
`ρ = 1.21 kg/m³`, `C_d ≈ 0.55`, `C_l` from spin ratio, gravity `9.81`.

**Bounce** (ball crosses court plane `y = 0` with downward `v.y`):
- normal: `v.y' = -e · v.y`, restitution `e ≈ 0.73`.
- tangential + spin coupling: friction transfers between horizontal velocity and
  spin — topspin adds forward speed and lowers rebound angle; backspin reduces
  forward speed (skid/check); sidespin adds lateral velocity. Spin decays each
  bounce.

**Net**: plane at `z = 0`, height `0.914 m` center rising to `1.07 m` at posts.
A ball whose trajectory intersects the net below the local height loses energy
and deflects; above it passes.

**Stroke → ball**: a stroke sets the ball's outgoing velocity and spin from
(swing type, aim, power, timing quality). Timing quality is a fixed-point factor
in `[0,1]`; off-timing reduces power and adds aim error deterministically.

## Court & rules

- Singles court `23.77 m (length) × 8.23 m (width)`, service line `6.40 m` from
  net, service boxes, baselines/sidelines. `classifyBounce(pos, phase) ->
  {In, Out, ServiceBoxOK, ServiceFault, Let}`.
- **Scoring** is a pure state machine: `applyPoint(matchState, winner) ->
  matchState`. Points 0/15/30/40, deuce, advantage; game→set (first to 6, win by
  2, tiebreak to 7 at 6–6); match best-of-3. Serve/end rotation and tiebreak
  serve order included.

## Netcode model (server-authoritative + prediction + reconciliation)

- **Server**: 60 Hz tick. Per tick: drain input queue → apply inputs at their
  tick → `Step` the sim → produce a snapshot → delta-encode vs each client's
  acked baseline → send. Rooms are independent; one room = one match.
- **Client**: samples input each frame, tags it `{seq, clientTick}`, sends it,
  and **predicts** the local player with the shared sim. On each server snapshot
  it **reconciles**: reset to the authoritative state at the acked seq, then
  replay all still-unacked inputs. Remote entities (ball, opponent) are **buffered
  and interpolated** between snapshots. Clock offset/RTT estimated from timestamps.
- **Lag compensation**: for contact detection, the server rewinds the world to the
  striker's view time (RTT/interp-aware) before testing the hit.
- **Delta compression**: snapshot deltas reference a baseline sequence the client
  ACKs; `encode(base, full) -> delta`, `apply(base, delta) -> full` round-trips
  exactly (golden-vector verified).

## Data models (PostgreSQL)

- `users(id, email unique, password_hash, display_name, created_at)`
- `ratings(user_id, rating, rd, volatility, updated_at)` (Glicko-2 fields)
- `matches(id, player_a, player_b, winner, score_json, seed, started_at, ended_at)`
- `match_inputs(match_id, input_log)` — deterministic replay data
- `leaderboard` — view/query over `ratings`

## Module boundaries (so tasks don't collide)

- `backend/sim` — fixed-point math, physics, court, stroke, scoring. **Pure.**
- `backend/netcode` — snapshots, delta encode/decode, prediction contract,
  lag-comp. Depends on `sim` + `Transport` interface only.
- `backend/server` — room loop, matchmaking glue, WebSocket transport, session/tick
  orchestration, supervision, graceful shutdown.
- `backend/services` — accounts, matchmaking, ranking, persistence, replay.
- `backend/obs` — logging, metrics, health.
- `frontend/sim` — TS port of the fixed-point sim (client prediction), verified
  against shared golden vectors.
- `frontend/net` — client prediction/reconciliation/interpolation, clock sync.
- `frontend/render` — `Renderer` interface + canvas impl (thin).
- `frontend/ui` — React lobby, matchmaking, HUD/scoreboard.

## Step Dependencies

The Principal uses this table to sequence issue creation and worker dispatch.
Each row's **Acceptance Criteria** describe observable behavior verified by a
named test — never a test function name in place of the behavior.

| Step | Description | Repo | Depends On | Acceptance Criteria |
|------|-------------|------|-----------|---------------------|
| 1 | Fixed-point math + seeded PRNG + golden-vector loader | backend | - | Q32.32 add/mul/div/sqrt round-trip within documented precision; seeded PRNG reproduces a sequence; loader parses `backend/testdata/vectors` |
| 2 | Vector math + court geometry + in/out classifier | backend | 1 | `classifyBounce` returns correct In/Out/ServiceBox/Let for boundary and interior points |
| 3 | Ball integrator: gravity + drag (golden vectors) | backend | 1,2 | A launched ball's trajectory matches the gravity+drag golden vector exactly; apex/return within tolerance |
| 4 | Aerodynamics: Magnus/spin (golden vectors) | backend | 3 | Topspin dips, backspin extends, sidespin deviates — each matches its golden vector |
| 5 | Bounce model: restitution + spin-coupled friction | backend | 3,4 | Topspin kicks forward at lower angle; backspin checks; sidespin deviates — matches bounce golden vectors |
| 6 | Net collision | backend | 3 | Ball below local net height deflects and loses energy; ball above passes; verified at center and posts |
| 7 | Stroke model: swing types, timing, spin imparted | backend | 4 | Flat/topspin/slice produce documented speed+spin; off-timing reduces power/accuracy deterministically |
| 8 | Scoring state machine (point→game→set→match) | backend | 1 | Known scorelines (deuce, tiebreak, best-of-3) reach correct terminal results; property test over legal sequences |
| 9 | Serve/rally flow + faults + lets | backend | 2,8 | Double fault loses point; serve outside box is fault; let replays; service-box validation correct |
| 10 | Match sim: pure `Step(state, inputs)` integrating 3–9 | backend | 5,6,7,9 | Same state + input log yields bit-identical output; a scripted rally produces the expected score |
| 11 | `Transport` interface + in-memory transport | backend | 1 | In-memory transport delivers messages deterministically; ordering/latency configurable in tests |
| 12 | Server tick loop + room model + input application | backend | 10,11 | Room steps at fixed tick; out-of-order/out-of-range inputs clamped/ignored; scripted match ends correctly |
| 13 | Snapshot + delta compression (golden vectors) | backend | 12 | `encode`/`apply` round-trip reconstructs state exactly; delta smaller than full for typical frames |
| 14 | WebSocket transport (prod) | backend | 11 | Implements `Transport`; frames encode/decode losslessly; handles close/error without panic |
| 15 | Client prediction + reconciliation (shared sim) | frontend | 10,13 | After a correction, replaying unacked inputs yields state matching the authoritative snapshot |
| 16 | Entity interpolation + clock sync | frontend | 15 | Interpolated remote position lies between buffered snapshots; clock offset estimate converges under jitter |
| 17 | Lag compensation (server rewind) | backend | 12 | A hit valid at the striker's view time is accepted after server rewind; stale hits rejected |
| 18 | Canvas renderer behind `Renderer` interface | frontend | - | Given a scene, the renderer is called with correct entity positions and score (MockRenderer asserts) |
| 19 | Input handling → prediction inputs | frontend | 18 | Key/pointer input maps to move/swing/serve inputs with correct type and timing tag |
| 20 | Client netcode wiring (predict+reconcile+interp) | frontend | 15,16,19 | End-to-end: local client predicts, reconciles on snapshot, renders interpolated opponent/ball |
| 21 | HUD / scoreboard from game state | frontend | 8,18 | Scoreboard renders correct game/set/match score derived from state |
| 22 | Lobby + matchmaking UI (React) | frontend | - | Queue/ready/result screens render and transition on state; components unit-tested |
| 23 | Accounts: register/login (bcrypt/argon2 + JWT) | backend | - | Register hashes password; login issues JWT; protected route rejects invalid/expired token |
| 24 | Postgres persistence + migrations | backend | - | Migrations create schema; repository CRUD round-trips users/matches/ratings |
| 25 | Ranking (Glicko-2) with golden vectors | backend | 1 | Rating updates after a match match the ranking golden vectors |
| 26 | Matchmaking queue + pairing (seeded) | backend | 25 | Seeded queue pairs closest-MMR players deterministically; unpaired remain queued |
| 27 | Match lifecycle + result persistence + input log | backend | 12,24 | Completed match persists result, score, seed, and input log |
| 28 | Replay re-simulation + verification | backend | 10,27 | Re-simulating a stored input log reproduces the exact final state; mismatch is detected |
| 29 | Reconnection + state resync | backend | 12,27 | A client reconnecting within grace resyncs and the match continues; late reconnect rejected |
| 30 | Anti-cheat: input sanity + rate limiting | backend | 12,23 | Floods are rate-limited without crash; impossible inputs rejected; replay mismatch flagged |
| 31 | Observability: structured logs + metrics + health | backend | 12 | `/healthz`/`/readyz` respond; Prometheus exposes tick/RTT/room metrics; no secrets logged |
| 32 | Resilience: supervision + panic recovery + graceful shutdown | backend | 12 | A panicking room is isolated; shutdown drains/persists rooms; other rooms unaffected |
| 33 | Security hardening: authz on endpoints, input validation, DoS guards | backend | 23,30 | All mutating endpoints require auth; malformed payloads rejected; connection caps enforced |
| 34 | Deployment: Dockerfiles + docker-compose + env config | root | 24,31 | `docker-compose up` starts server+Postgres+Prometheus; config from env; migrations run |
| 35 | CI/CD + migration automation | root | 34 | CI builds/tests backend+frontend, runs migrations, and `awkit evaluate --offline --strict` passes |
| 36 | Load-test harness + performance budget | backend | 12 | Harness drives N concurrent matches; asserts p99 tick time under threshold at target room count |
| 37 | Checkpoint: headless full-match E2E + all green | root | 20,28,33,35,36 | Two in-memory clients play a full deterministic match to a valid result; `go test ./...`, `npm test`, and `awkit evaluate --offline` all pass |

### Dependency rules
- `-` : may start immediately.
- `Step N` : starts only after N is done.
- `Steps N,M` : all listed must complete first.
- Independent steps may run in parallel when workers are free.

## Verification strategy

- **Determinism**: golden vectors (`backend/testdata/vectors`, synced into the frontend for its tests) replayed by both Go and TS
  sims; exact-equality assertions.
- **Physics/rules**: unit tests per model against golden vectors + property tests
  for scoring.
- **Netcode**: in-memory transport E2E — server + two clients, full match, headless,
  deterministic; delta encode/decode round-trip vectors; reconciliation replay test.
- **Client logic**: vitest over sim port, prediction, interpolation, input, HUD;
  `MockRenderer` for the render boundary.
- **Services**: repository round-trips against a test Postgres (or an interface +
  in-memory fake for pure-logic tests); auth/ranking/matchmaking unit tests.
- **Ops**: health/metrics endpoint tests; load harness asserts the perf budget.
- **Per-repo commands**:
  - backend: `go build ./...` + `go test ./...`
  - frontend: `npm ci && npm run build` + `npm test`
  - root: `awkit evaluate --offline` (+ `--strict` in CI)

## Definition of Done (commercial checklist)

- [ ] Deterministic sim with golden vectors (Go ⇔ TS parity)
- [ ] Realistic physics: gravity, drag, Magnus/spin, bounce, net
- [ ] Full tennis rules (point→game→set→match, serve, faults, tiebreak)
- [ ] Server-authoritative netcode: prediction, reconciliation, interpolation, lag-comp
- [ ] Accounts + JWT; matchmaking + Glicko-2; leaderboard
- [ ] Postgres persistence + migrations; input-log replays + verification
- [ ] Reconnection + resync; anti-cheat + rate limiting
- [ ] Observability (logs/metrics/health); resilience (supervision/graceful shutdown)
- [ ] Security hardening; Docker/compose deploy
- [ ] CI green including `awkit evaluate --offline --strict`; headless full-match E2E
