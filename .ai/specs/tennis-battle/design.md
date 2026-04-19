# Tennis Battle - Technical Design

## Tech Stack

| Layer | Technology | Reason |
|-------|-----------|--------|
| Backend | Go 1.24+ | High-performance game server, goroutine-per-match |
| Frontend | Phaser 3 + TypeScript | Most documented HTML5 game framework, built-in physics/input/scenes |
| Build | Vite | Fast TS bundling, HMR for dev |
| WebSocket | nhooyr.io/websocket | Context-aware, cleaner API than gorilla (archived) |
| Deployment | PWA (static frontend) + Go binary | Single web app serves all platforms |
| Logging | zerolog | Structured JSON logging |

## Architecture Overview

```
Mobile/PC Browser (Phaser 3 PWA)
    |
    | REST (lobby, rooms, matchmaking)
    | WebSocket (game state sync, 30 tick/sec)
    |
Go Server
    ├── HTTP Router (REST + WS upgrade)
    ├── Lobby (rooms, matchmaking queue)
    ├── Match Engine (goroutine per match)
    │   ├── Physics (ball, collision, court)
    │   ├── Player (input processing, state)
    │   ├── AI (rule-based opponent)
    │   └── Power-ups (spawn, collect, effects)
    └── Transport (WebSocket conn wrapper)
```

## Backend Structure

```
backend/
  cmd/server/main.go              # HTTP server, signal handling, graceful shutdown
  internal/
    server/server.go              # HTTP routes, WebSocket upgrade, DI wiring
    match/
      match.go                    # Match struct, game loop (30 tick/sec), lifecycle
      physics.go                  # Ball movement, collision detection, court bounds
      player.go                   # Player state, input processing, position update
      ai.go                       # AI opponent (rule-based, 3 difficulty levels)
      powerup.go                  # Power-up spawn, collection, effect application
      character.go                # Character stats and super shot definitions
    lobby/
      lobby.go                    # Room registry, matchmaking FIFO queue
      room.go                     # Room struct (waiting/playing/finished states)
    protocol/
      messages.go                 # All message types (ClientInput, ServerState, etc.)
      codec.go                    # Binary encode/decode (encoding/binary)
    transport/
      conn.go                     # WebSocket connection wrapper, read/write pumps
  go.mod
  go.sum
```

## Frontend Structure

```
frontend/
  index.html
  src/
    main.ts                       # Phaser game config, boot
    scenes/
      BootScene.ts                # Asset preloading
      MenuScene.ts                # Main menu (play online, vs AI, local)
      LobbyScene.ts               # Room creation/joining, matchmaking queue
      GameScene.ts                # Main gameplay (court render, input, HUD)
      ResultScene.ts              # Match result screen
    game/
      Court.ts                    # Court rendering (isometric)
      Player.ts                   # Player sprite, animation
      Ball.ts                     # Ball sprite, interpolation
      PowerUp.ts                  # Power-up sprites and effects
      HUD.ts                      # Score, energy meter, timer UI
    network/
      WebSocketClient.ts          # WS connection, reconnect logic
      Protocol.ts                 # Binary decode/encode (matching server codec)
      StateBuffer.ts              # Server state interpolation buffer
    input/
      TouchInput.ts               # Virtual joystick + shot buttons (mobile)
      KeyboardInput.ts            # Keyboard controls (PC)
      InputManager.ts             # Unified input abstraction
    characters/
      CharacterDefs.ts            # Character stats and sprite mappings
  public/
    manifest.json                 # PWA manifest
    sw.js                         # Service worker (asset caching)
    assets/                       # Sprites, sounds
  vite.config.ts
  tsconfig.json
  package.json
```

## Protocol Design

### Client → Server (Input only)
```
ClientInput {
  tick:      uint32    // echo server tick for latency calc
  moveX:     int8      // -1, 0, 1
  moveY:     int8      // -1, 0, 1
  shotType:  uint8     // 0=none, 1=flat, 2=topspin, 3=slice, 4=lob
  superShot: uint8     // 0=no, 1=activate
}
// Total: 7 bytes
```

### Server → Client (State broadcast, 30 Hz)
```
ServerState {
  tick:        uint32
  ballX:       float32
  ballY:       float32
  ballVelX:    float32
  ballVelY:    float32
  player1X:    float32
  player1Y:    float32
  player2X:    float32
  player2Y:    float32
  score1:      uint8
  games1:      uint8
  score2:      uint8
  games2:      uint8
  energy1:     uint8     // 0-100
  energy2:     uint8     // 0-100
  gameState:   uint8     // serving, playing, point_scored, game_over
  powerUpType: uint8     // 0=none, 1=speed, 2=big_racket, 3=ice
  powerUpX:    float32
  powerUpY:    float32
  flags:       uint16    // bitfield: who has what active effects
}
// Total: ~52 bytes
```

### Events (variable length, JSON)
```
// Server → Client events (infrequent, JSON is fine)
MatchStart   { yourSide: "bottom"|"top", opponent: string, character: string }
PointScored  { scorer: 1|2, newScore: [int,int] }
GameWon      { winner: 1|2, newGames: [int,int] }
MatchEnd     { winner: 1|2, finalScore: string }
PowerUpSpawn { type: string, x: float, y: float }
SuperShot    { player: 1|2, type: string }
Disconnected { reason: string, reconnectSec: int }
```

## Match Lifecycle

```
1. Room created (REST) → room in "waiting" state
2. Player 2 joins (REST) → room in "ready" state
3. Both connect via WebSocket → room transitions to "playing"
4. Server creates Match goroutine with 30Hz ticker
5. Game loop: read inputs → simulate physics → broadcast state
6. Match ends (someone wins set) → write result → close connections
7. Match goroutine exits, room removed from registry
```

## AI Design (Rule-Based)

```go
func (ai *AIPlayer) DecideInput(ball, self Position, ballVel Velocity) PlayerInput {
    // 1. Move toward predicted ball landing position
    target := predictLanding(ball, ballVel)
    // 2. Apply reaction delay (Easy: 200ms, Med: 100ms, Hard: 30ms)
    // 3. Apply positional error (Easy: ±40px, Med: ±20px, Hard: ±5px)
    // 4. Swing when ball is in range (timing accuracy varies by difficulty)
}
```

## Step Dependencies

| Step | Description | Depends On | Acceptance Criteria |
|------|-------------|------------|---------------------|
| 1 | Project scaffolding (Go mod, Vite, Phaser boot) | - | Both `go build` and `npm run build` succeed; Phaser shows empty canvas |
| 2 | Binary protocol codec (shared encode/decode) | - | Encode → decode round-trip tests pass for all message types |
| 3 | WebSocket transport layer | Step 1 | Server accepts WS connections; client connects and receives ping |
| 4 | Court rendering (isometric view) | Step 1 | Phaser renders tennis court with net, lines, correct perspective |
| 5 | Player movement + input system | Steps 3, 4 | Player sprite moves via keyboard/touch; input sent over WS; server echoes position |
| 6 | Ball physics engine (server-side) | Step 2 | Ball bounces, collides with court bounds, net interaction; unit tests cover edge cases |
| 7 | Shot mechanics (flat, topspin, slice, lob) | Steps 5, 6 | Player can hit ball with 4 shot types; ball trajectory differs per type |
| 8 | Serving system | Step 7 | Serve toss + timing meter; service alternates sides; fault detection |
| 9 | Scoring system | Step 8 | Points/games tracked; match ends at 3 games; score displayed in HUD |
| 10 | Lobby + matchmaking | Step 3 | REST endpoints for room create/join/queue; private room codes work |
| 11 | Character system (4 characters with stats) | Step 5 | Character selection screen; stats affect gameplay (speed/power/control) |
| 12 | Energy system + super shots | Steps 7, 11 | Energy fills on rallies; super shot activates when full; each character has unique effect |
| 13 | Power-up system | Steps 6, 7 | Power-ups spawn near net; collectible on approach; effects apply correctly |
| 14 | AI opponent | Steps 6, 7 | 3 difficulty levels; AI makes believable decisions; beatable on Easy |
| 15 | PWA setup (manifest, service worker, orientation lock) | Step 4 | Installable on mobile; fullscreen landscape; passes Lighthouse PWA audit |
| 16 | Client-side prediction + interpolation | Steps 5, 6 | Own player movement feels instant; opponent/ball interpolated smoothly |
| 17 | Reconnection handling | Steps 3, 10 | 10-sec reconnect window; match pauses; forfeit on timeout |
| 18 | Match result screen + polish | Steps 9, 10 | Result screen shows winner/stats; return to menu; animations/transitions |
| 19 | Mobile touch controls (virtual joystick + buttons) | Step 5 | Joystick and shot buttons render; responsive on touch devices; no overlap with gameplay |
| 20 | Local multiplayer (PC) | Steps 5, 7 | Two players on same keyboard (WASD + Arrows); split shot buttons |

## Verification

### Backend
```bash
cd backend && go build ./... && go test ./...
```

### Frontend
```bash
cd frontend && npm install && npm run build && npm run test -- --run
```
