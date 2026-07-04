# Tennis Arena — Requirements

> A commercial-grade, real-time **networked** tennis game with **realistic ball
> physics**, built server-authoritative and **verifiable end-to-end without a GUI**.
>
> These requirements are written **ACI-first** (agent-friendly): every requirement
> is phrased so that a single automated test can decide whether it is met. The
> guiding rule — **if it can't be checked by a deterministic test, it isn't a
> requirement here, it's a note.**

## Goal

Ship a playable, operable 1-v-1 online tennis game where two players compete in
real time, the ball obeys believable physics (gravity, drag, spin/Magnus,
realistic bounce), the server is authoritative (cheat-resistant), and the entire
simulation is **deterministic and headless-testable** so quality can be enforced
automatically.

## Definitions

- **Sim** — the deterministic simulation core: `Step(state, inputs) -> state`.
- **Tick** — one authoritative server simulation step (fixed rate).
- **Snapshot** — server-authoritative world state sent to clients.
- **Golden vector** — a committed input→output fixture a test replays for an
  exact-match assertion (the backbone of determinism verification).
- **In-memory transport** — a `Transport` implementation with no sockets, used to
  run full server+client matches inside one test process, deterministically.

---

## R1: Deterministic simulation core

- R1.1 THE sim SHALL advance on a **fixed timestep** (no wall-clock, no
  frame-rate dependence); given the same initial state and input sequence it
  SHALL produce **bit-identical** output.
- R1.2 THE sim SHALL use **deterministic fixed-point arithmetic** (integer
  Q-format) for all state that affects gameplay, so results are identical across
  OS, CPU, and language (Go server ⇔ TypeScript client prediction).
- R1.3 THE sim SHALL expose `Step` as a **pure function** of `(state, inputs)`
  with no I/O, no goroutines, and no global mutable state.
- R1.4 ALL randomness (serve toss variance, coin toss) SHALL derive from a
  **seeded PRNG** carried in the state; the seed SHALL be reproducible.
- R1.5 THE repository SHALL provide **golden-vector fixtures** and a loader; the
  Go sim and the TS prediction sim SHALL both reproduce every physics/rules
  golden vector exactly.

## R2: Realistic ball physics

- R2.1 THE ball SHALL be integrated under **gravity** (9.81 m/s² downward) at a
  fixed sub-step; a vertically-launched ball SHALL return to launch height within
  a documented tolerance (energy-consistent trajectory).
- R2.2 THE ball SHALL experience **aerodynamic drag** proportional to the square
  of speed; a struck ball SHALL decelerate along its path per the drag model.
- R2.3 THE ball SHALL experience **Magnus force** from spin: topspin SHALL curve
  the trajectory downward (shorter, dipping arc), backspin SHALL extend it, and
  sidespin SHALL deviate it laterally — each verified against golden vectors.
- R2.4 ON contact with the court plane THE ball SHALL **bounce** with a
  restitution coefficient and **spin-coupled friction**: topspin SHALL kick
  forward with a lower rebound angle, backspin SHALL skid/check, sidespin SHALL
  deviate laterally — verified against golden vectors.
- R2.5 THE ball SHALL collide with the **net** (correct height at center vs
  posts), losing energy and deflecting; balls clearing the net SHALL not.
- R2.6 THE physics constants (ball mass 57 g, diameter 6.7 cm, drag/lift
  coefficients, restitution, friction) SHALL be centralized and documented.

## R3: Court, strokes, and in/out rules

- R3.1 THE court geometry SHALL model a regulation **singles court**
  (23.77 m × 8.23 m), service boxes, baselines, and sidelines.
- R3.2 ON each bounce THE sim SHALL classify the landing position as **in / out /
  let region / service-box** per the current rally phase.
- R3.3 A **stroke** SHALL be produced from player input: swing type (flat /
  topspin / slice), aim, power, and timing; timing quality SHALL affect resulting
  ball speed, spin, and accuracy.
- R3.4 A **serve** SHALL require the ball to land in the diagonally-opposite
  service box; a serve into the net or outside the box SHALL be a fault; two
  faults SHALL lose the point.

## R4: Tennis scoring & match flow

- R4.1 THE scoring engine SHALL implement point → game (0/15/30/40, deuce,
  advantage) → set (first to 6, win-by-2, **tiebreak at 6–6**) → match
  (best-of-3 sets) as a pure state machine.
- R4.2 THE engine SHALL rotate **serve** each game and **ends** per the rules,
  and SHALL handle lets, double faults, and tiebreak serve rotation.
- R4.3 GIVEN any legal sequence of point outcomes THE engine SHALL produce the
  correct score and terminal match result (property-tested against known
  scorelines).

## R5: Networked real-time play (server-authoritative)

- R5.1 THE **server** SHALL run the authoritative sim at a fixed tick rate and be
  the single source of truth for world state and outcomes.
- R5.2 CLIENTS SHALL send **timestamped, sequence-numbered inputs**; the server
  SHALL apply them by tick and SHALL ignore/clamp out-of-range or out-of-order
  inputs.
- R5.3 THE server SHALL broadcast **snapshots** with **delta compression** against
  an acknowledged baseline; a snapshot delta round-trip SHALL reconstruct state
  exactly (golden-vector verified).
- R5.4 THE netcode SHALL depend on a **`Transport` interface**; an in-memory
  transport SHALL allow a **full two-client match to run headless in one test
  process** with deterministic results.
- R5.5 A real **WebSocket** transport SHALL implement the same interface for
  production play.

## R6: Client prediction, reconciliation, interpolation

- R6.1 THE client SHALL **predict** the local player using the shared sim and
  **reconcile** on server correction by replaying pending inputs from the acked
  state; after reconciliation the predicted state SHALL match the authoritative
  state (deterministic replay test).
- R6.2 THE client SHALL **interpolate** remote entities (ball, opponent) between
  buffered snapshots to render smoothly under jitter.
- R6.3 THE server SHALL apply **lag compensation** (rewind to the shooter's view
  time) for hit/contact detection; a documented test SHALL show a hit that is
  valid at the client's view time is accepted.
- R6.4 THE client and server clocks SHALL be synchronized (offset/RTT estimation)
  within a documented bound.

## R7: Accounts & sessions (commercial)

- R7.1 THE system SHALL support **registration and login** (email + password,
  passwords hashed with bcrypt/argon2); credentials SHALL never be stored or
  logged in plaintext.
- R7.2 AUTHENTICATED sessions SHALL use signed **JWT** tokens with expiry;
  protected endpoints SHALL reject missing/invalid/expired tokens.
- R7.3 A **player profile** (display name, rating, match history reference) SHALL
  be retrievable for an authenticated user.

## R8: Matchmaking & ranking

- R8.1 THE system SHALL provide a **matchmaking queue** that pairs players by
  skill (MMR); pairing SHALL be **seeded/deterministic under test**.
- R8.2 RATINGS SHALL update after each ranked match via a documented algorithm
  (Glicko-2 or Elo) verified against golden vectors.
- R8.3 A **leaderboard** SHALL rank players by rating and be queryable.

## R9: Persistence & replays

- R9.1 THE system SHALL persist users, matches, and ratings in **PostgreSQL**
  with **versioned migrations**.
- R9.2 EACH match SHALL store its **input log** (deterministic replay data), not
  just the score.
- R9.3 A stored match SHALL be **re-simulated from its input log** to reproduce
  the exact final state (replay verification) — this doubles as anti-cheat and
  spectating substrate.

## R10: Anti-cheat & integrity

- R10.1 BECAUSE the server is authoritative, clients SHALL NOT be trusted for
  outcomes; the server SHALL **validate and clamp** all inputs (rate, range,
  timing).
- R10.2 THE system SHALL **rate-limit** inputs and API calls per connection and
  reject floods without crashing.
- R10.3 REPLAY re-simulation mismatches SHALL be **detectable** and flagged.

## R11: Resilience & reconnection

- R11.1 A player who disconnects SHALL be able to **reconnect within a grace
  window** and receive a **state resync**; the match SHALL continue.
- R11.2 THE server SHALL handle a panicking room without taking down other rooms
  (**room supervision / panic recovery**).
- R11.3 THE server SHALL **shut down gracefully**, draining or persisting active
  rooms.

## R12: Observability & operability

- R12.1 THE server SHALL emit **structured logs** (leveled, JSON) and SHALL NOT
  log secrets.
- R12.2 THE server SHALL expose **Prometheus metrics** (tick duration, RTT, active
  rooms, matchmaking latency, error counts) and **health/readiness** endpoints.
- R12.3 A **load-test harness** SHALL drive N concurrent simulated matches and
  assert a **performance budget** (e.g. p99 tick time under a threshold at a
  target room count).

## R13: Deployment & delivery

- R13.1 THE server, client, and database SHALL be runnable via **Docker /
  docker-compose** (server + Postgres + Prometheus) with **env-based config**.
- R13.2 THE repository CI SHALL build and test both `backend/` and `frontend/`
  and run migrations; `awkit evaluate --offline --strict` SHALL pass.

## R14: ACI / testability (cross-cutting — the reason this spec exists)

- R14.1 EVERY acceptance criterion in `tasks.md` SHALL map to at least one named
  automated test with a **real assertion** (so the review evidence gate can
  verify it).
- R14.2 THE untestable surface (canvas rendering) SHALL be **isolated behind an
  interface**; game logic (prediction, interpolation, input, HUD derivation)
  SHALL be unit-testable without a browser.
- R14.3 CROSS-language determinism (Go sim ⇔ TS prediction sim) SHALL be enforced
  by shared golden vectors, not by inspection.
- R14.4 THE deterministic core + input-log replays SHALL make **any bug
  reproducible** from stored data.

---

## Non-goals (v1)

- 3D photoreal graphics, animation blending, audio design — rendering is a thin,
  functional 2.5D canvas view; polish is out of scope.
- Doubles (2-v-2), tournaments/brackets, in-game chat/voice, cosmetics/store.
- Mobile-native clients (web client only in v1).
- Cross-region dedicated-server orchestration beyond horizontal-scale design
  (the design documents the scaling model; global rollout is future work).
