# Tennis Arena — Tasks

Repo: backend, frontend
Coordination: sequential-with-parallelism
Sync: independent (directory monorepo)

> One GitHub issue is created per **main task** below, top to bottom. Each task's
> full acceptance criteria and testing approach live in `design.md` (the
> Step Dependencies table, keyed by the same number). Every acceptance criterion
> must be proven by a named automated test with a real assertion — the review
> evidence gate re-runs the suite and verifies each one.

## Tasks

- [ ] 1. Backend: fixed-point math, seeded PRNG, and golden-vector loader <!-- Issue #1 -->
  - Repo: backend
  - Depends on: -
  - [ ] 1.1 Implement Q32.32 fixed-point type (add/sub/mul/div/sqrt) and seeded PRNG
    - _Requirements: R1.1, R1.2, R1.4_
  - [ ] 1.2 Implement the `backend/testdata/vectors` golden-vector loader
    - _Requirements: R1.5_
  - [ ] 1.3 Unit tests: fixed-point round-trip precision, PRNG reproducibility, loader parsing
    - _Requirements: R1.1, R1.4, R1.5_

- [ ] 2. Backend: vector math, court geometry, and in/out classifier <!-- Issue #2 -->
  - Repo: backend
  - Depends on: Step 1
  - [ ] 2.1 Implement court geometry (singles court, service boxes, net line) and `classifyBounce`
    - _Requirements: R3.1, R3.2_
  - [ ] 2.2 Unit tests: boundary and interior points classify as In/Out/ServiceBox/Let correctly
    - _Requirements: R3.2_

- [ ] 3. Backend: ball integrator — gravity + drag (golden vectors) <!-- Issue #3 -->
  - Repo: backend
  - Depends on: Steps 1,2
  - [ ] 3.1 Implement fixed sub-step semi-implicit Euler integration with gravity and quadratic drag
    - _Requirements: R2.1, R2.2, R1.1_
  - [ ] 3.2 Add gravity+drag golden vectors and tests asserting exact trajectory reproduction
    - _Requirements: R2.1, R2.2, R1.5_

- [ ] 4. Backend: aerodynamics — Magnus / spin (golden vectors) <!-- Issue #4 -->
  - Repo: backend
  - Depends on: Step 3
  - [ ] 4.1 Add Magnus force from angular velocity; topspin dips, backspin extends, sidespin deviates
    - _Requirements: R2.3_
  - [ ] 4.2 Add spin golden vectors and tests asserting each spin behavior
    - _Requirements: R2.3, R1.5_

- [ ] 5. Backend: bounce model — restitution + spin-coupled friction <!-- Issue #5 -->
  - Repo: backend
  - Depends on: Steps 3,4
  - [ ] 5.1 Implement bounce: restitution on normal, spin↔tangential-velocity coupling, spin decay
    - _Requirements: R2.4_
  - [ ] 5.2 Add bounce golden vectors and tests (topspin kick, backspin check, sidespin deviation)
    - _Requirements: R2.4, R1.5_

- [ ] 6. Backend: net collision <!-- Issue #6 -->
  - Repo: backend
  - Depends on: Step 3
  - [ ] 6.1 Implement net plane with center/post height; deflect+dampen below, pass above
    - _Requirements: R2.5_
  - [ ] 6.2 Unit tests at center and posts: clearing vs clipping the net
    - _Requirements: R2.5_

- [ ] 7. Backend: stroke model — swing types, timing, imparted spin
  - Repo: backend
  - Depends on: Step 4
  - [ ] 7.1 Implement stroke → outgoing velocity+spin from (swing type, aim, power, timing)
    - _Requirements: R3.3_
  - [ ] 7.2 Unit tests: flat/topspin/slice output profiles; off-timing reduces power/accuracy
    - _Requirements: R3.3_

- [ ] 8. Backend: scoring state machine (point → game → set → match)
  - Repo: backend
  - Depends on: Step 1
  - [ ] 8.1 Implement pure `applyPoint`: 0/15/30/40, deuce, advantage, set (first-to-6, win-by-2, tiebreak), best-of-3
    - _Requirements: R4.1, R4.2_
  - [ ] 8.2 Tests: known scorelines reach correct terminals; property test over legal point sequences
    - _Requirements: R4.3_

- [ ] 9. Backend: serve/rally flow, faults, and lets
  - Repo: backend
  - Depends on: Steps 2,8
  - [ ] 9.1 Implement serve validation (diagonal service box), fault/double-fault, let replay
    - _Requirements: R3.4, R4.2_
  - [ ] 9.2 Tests: double fault loses point; out-of-box serve is fault; let replays
    - _Requirements: R3.4_

- [ ] 10. Backend: match sim — pure `Step(state, inputs)` integrating physics + rules
  - Repo: backend
  - Depends on: Steps 5,6,7,9
  - [ ] 10.1 Compose integrator, bounce, net, stroke, serve, and scoring into a pure `Step`
    - _Requirements: R1.3, R2.*, R3.*, R4.*_
  - [ ] 10.2 Tests: identical state+input log yields bit-identical output; a scripted rally reaches the expected score
    - _Requirements: R1.1, R4.3_

- [ ] 11. Backend: Transport interface + deterministic in-memory transport
  - Repo: backend
  - Depends on: Step 1
  - [ ] 11.1 Define `Transport` interface; implement in-memory transport with configurable order/latency
    - _Requirements: R5.4_
  - [ ] 11.2 Tests: deterministic delivery; latency/reorder honored
    - _Requirements: R5.4_

- [ ] 12. Backend: server tick loop, room model, input application
  - Repo: backend
  - Depends on: Steps 10,11
  - [ ] 12.1 Implement 60 Hz room loop: drain inputs by tick, `Step`, clamp/ignore invalid inputs
    - _Requirements: R5.1, R5.2_
  - [ ] 12.2 Tests: scripted match over in-memory transport ends with correct score; bad inputs rejected
    - _Requirements: R5.1, R5.2_

- [ ] 13. Backend: snapshot generation + delta compression (golden vectors)
  - Repo: backend
  - Depends on: Step 12
  - [ ] 13.1 Implement snapshot + `encode(base,full)->delta` / `apply(base,delta)->full` against acked baselines
    - _Requirements: R5.3_
  - [ ] 13.2 Golden-vector tests: delta round-trip reconstructs state exactly; delta < full for typical frames
    - _Requirements: R5.3, R1.5_

- [ ] 14. Backend: WebSocket transport (production)
  - Repo: backend
  - Depends on: Step 11
  - [ ] 14.1 Implement WebSocket `Transport` with lossless frame encode/decode and close/error handling
    - _Requirements: R5.5_
  - [ ] 14.2 Tests: frame round-trip lossless; close/error handled without panic
    - _Requirements: R5.5_

- [ ] 15. Frontend: TS sim port + client prediction & reconciliation
  - Repo: frontend
  - Depends on: Steps 10,13
  - [ ] 15.1 Port the fixed-point `Step` to TS; add `sync-vectors` and verify against the synced golden vectors
    - _Requirements: R1.2, R14.3_
  - [ ] 15.2 Implement predict + reconcile (replay unacked inputs from acked state)
    - _Requirements: R6.1_
  - [ ] 15.3 Tests: TS sim matches golden vectors; post-correction state matches authoritative snapshot
    - _Requirements: R6.1, R14.3_

- [ ] 16. Frontend: entity interpolation + clock sync
  - Repo: frontend
  - Depends on: Step 15
  - [ ] 16.1 Implement snapshot buffer + interpolation for ball/opponent; RTT/offset clock sync
    - _Requirements: R6.2, R6.4_
  - [ ] 16.2 Tests: interpolated position lies between buffered snapshots; offset estimate converges under jitter
    - _Requirements: R6.2, R6.4_

- [ ] 17. Backend: lag compensation (server rewind)
  - Repo: backend
  - Depends on: Step 12
  - [ ] 17.1 Implement server rewind to the striker's view time for contact detection
    - _Requirements: R6.3_
  - [ ] 17.2 Tests: hit valid at the client's view time is accepted; stale hit rejected
    - _Requirements: R6.3_

- [ ] 18. Frontend: canvas renderer behind a `Renderer` interface
  - Repo: frontend
  - Depends on: -
  - [ ] 18.1 Define `Renderer` interface; implement a thin canvas renderer for court/ball/players
    - _Requirements: R14.2_
  - [ ] 18.2 Tests: given a scene, MockRenderer receives correct entity positions and score
    - _Requirements: R14.2_

- [ ] 19. Frontend: input handling → prediction inputs
  - Repo: frontend
  - Depends on: Step 18
  - [ ] 19.1 Map keyboard/pointer to move/swing/serve inputs with swing type and timing tag
    - _Requirements: R3.3, R5.2_
  - [ ] 19.2 Tests: input mapping produces correct input type and timing
    - _Requirements: R3.3_

- [ ] 20. Frontend: client netcode wiring (predict + reconcile + interpolate)
  - Repo: frontend
  - Depends on: Steps 15,16,19
  - [ ] 20.1 Wire prediction, reconciliation, interpolation, and input send into the game loop
    - _Requirements: R6.1, R6.2_
  - [ ] 20.2 Tests: end-to-end client loop predicts, reconciles on snapshot, renders interpolated remote state
    - _Requirements: R6.1, R6.2_

- [ ] 21. Frontend: HUD / scoreboard from game state
  - Repo: frontend
  - Depends on: Steps 8,18
  - [ ] 21.1 Derive and render game/set/match score from state
    - _Requirements: R4.1_
  - [ ] 21.2 Tests: scoreboard renders correct score for representative states
    - _Requirements: R4.1_

- [ ] 22. Frontend: lobby + matchmaking UI (React)
  - Repo: frontend
  - Depends on: -
  - [ ] 22.1 Implement queue / ready / result screens driven by state
    - _Requirements: R8.1_
  - [ ] 22.2 Component tests: screens render and transition on state changes
    - _Requirements: R8.1_

- [ ] 23. Backend: accounts — register/login (password hashing + JWT)
  - Repo: backend
  - Depends on: -
  - [ ] 23.1 Implement register (bcrypt/argon2) and login issuing signed JWT with expiry
    - _Requirements: R7.1, R7.2_
  - [ ] 23.2 Tests: password hashed (never plaintext); protected route rejects missing/invalid/expired token
    - _Requirements: R7.1, R7.2, R7.3_

- [ ] 24. Backend: PostgreSQL persistence + migrations
  - Repo: backend
  - Depends on: -
  - [ ] 24.1 Add schema migrations (users, ratings, matches, match_inputs) and repositories
    - _Requirements: R9.1_
  - [ ] 24.2 Tests: migrations apply; repository CRUD round-trips users/matches/ratings
    - _Requirements: R9.1_

- [ ] 25. Backend: ranking (Glicko-2) with golden vectors
  - Repo: backend
  - Depends on: Step 1
  - [ ] 25.1 Implement Glicko-2 rating update
    - _Requirements: R8.2_
  - [ ] 25.2 Golden-vector tests: rating/RD/volatility updates match expected values
    - _Requirements: R8.2_

- [ ] 26. Backend: matchmaking queue + pairing (seeded)
  - Repo: backend
  - Depends on: Step 25
  - [ ] 26.1 Implement a seeded queue that pairs closest-MMR players; unpaired stay queued
    - _Requirements: R8.1_
  - [ ] 26.2 Tests: deterministic pairing under a seed; leaderboard query ordered by rating
    - _Requirements: R8.1, R8.3_

- [ ] 27. Backend: match lifecycle + result persistence + input log
  - Repo: backend
  - Depends on: Steps 12,24
  - [ ] 27.1 Implement create/ready/play/result lifecycle; persist result, score, seed, and input log
    - _Requirements: R9.2_
  - [ ] 27.2 Tests: completed match persists all fields incl. input log
    - _Requirements: R9.2_

- [ ] 28. Backend: replay re-simulation + verification
  - Repo: backend
  - Depends on: Steps 10,27
  - [ ] 28.1 Re-simulate a stored input log and compare to the recorded final state
    - _Requirements: R9.3, R10.3_
  - [ ] 28.2 Tests: replay reproduces the exact final state; an altered log is flagged as mismatch
    - _Requirements: R9.3, R10.3_

- [ ] 29. Backend: reconnection + state resync
  - Repo: backend
  - Depends on: Steps 12,27
  - [ ] 29.1 Support reconnect within a grace window with full state resync; reject late reconnects
    - _Requirements: R11.1_
  - [ ] 29.2 Tests: reconnect-in-grace resyncs and match continues; out-of-grace rejected
    - _Requirements: R11.1_

- [ ] 30. Backend: anti-cheat — input sanity + rate limiting
  - Repo: backend
  - Depends on: Steps 12,23
  - [ ] 30.1 Validate/clamp inputs (rate, range, timing); rate-limit per connection
    - _Requirements: R10.1, R10.2_
  - [ ] 30.2 Tests: input floods rate-limited without crash; impossible inputs rejected
    - _Requirements: R10.1, R10.2_

- [ ] 31. Backend: observability — structured logs, metrics, health
  - Repo: backend
  - Depends on: Step 12
  - [ ] 31.1 Add slog structured logging, Prometheus metrics (tick/RTT/rooms), and health/readiness endpoints
    - _Requirements: R12.1, R12.2_
  - [ ] 31.2 Tests: health endpoints respond; metrics registered; no secrets logged
    - _Requirements: R12.1, R12.2_

- [ ] 32. Backend: resilience — supervision, panic recovery, graceful shutdown
  - Repo: backend
  - Depends on: Step 12
  - [ ] 32.1 Isolate a panicking room; drain/persist rooms on graceful shutdown
    - _Requirements: R11.2, R11.3_
  - [ ] 32.2 Tests: a panicking room does not affect others; shutdown persists active rooms
    - _Requirements: R11.2, R11.3_

- [ ] 33. Backend: security hardening — authz, input validation, DoS guards
  - Repo: backend
  - Depends on: Steps 23,30
  - [ ] 33.1 Require auth on all mutating endpoints; reject malformed payloads; enforce connection caps
    - _Requirements: R10.1, R10.2, R7.2_
  - [ ] 33.2 Tests: unauthenticated mutation rejected; malformed payload rejected; cap enforced
    - _Requirements: R7.2, R10.2_

- [ ] 34. Root: deployment — Dockerfiles, docker-compose, env config
  - Repo: root
  - Depends on: Steps 24,31
  - [ ] 34.1 Add backend/frontend Dockerfiles and a compose stack (server + Postgres + Prometheus) with env config
    - _Requirements: R13.1_
  - [ ] 34.2 Verify compose builds; config sourced from env; migrations run on start
    - _Requirements: R13.1_

- [ ] 35. Root: CI/CD + migration automation
  - Repo: root
  - Depends on: Step 34
  - [ ] 35.1 CI builds/tests backend + frontend, runs migrations, and runs `awkit evaluate --offline --strict`
    - _Requirements: R13.2_
  - [ ] 35.2 Verify the CI workflow passes on a clean checkout
    - _Requirements: R13.2_

- [ ] 36. Backend: load-test harness + performance budget
  - Repo: backend
  - Depends on: Step 12
  - [ ] 36.1 Add a harness driving N concurrent in-memory matches; measure tick time
    - _Requirements: R12.3_
  - [ ] 36.2 Tests: assert p99 tick time under threshold at the target room count
    - _Requirements: R12.3_

- [ ] 37. Root: checkpoint — headless full-match E2E and all-green
  - Repo: root
  - Depends on: Steps 20,28,33,35,36
  - [ ] 37.1 Headless E2E: two in-memory clients play a full deterministic match to a valid result
    - _Requirements: R5.4, R14.1, R14.4_
  - [ ] 37.2 Ensure `go test ./...`, `npm test`, and `awkit evaluate --offline` all pass
    - _Requirements: R13.2, R14.1_
