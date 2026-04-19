# Tennis Battle - Requirements

## Overview

Tennis Battle is a real-time online multiplayer tennis game with "battle" elements (energy system, super shots, court power-ups). It runs on mobile (iOS/Android via PWA) and PC (desktop browser), sharing a single codebase.

## Target Platforms

- iOS Safari (PWA, landscape)
- Android Chrome (PWA, landscape)
- Desktop browsers (Chrome, Firefox, Safari, Edge)

## Game Modes

### 1. Online 1v1
- Public matchmaking (FIFO queue, first-come-first-served)
- Private rooms (6-character room code, share with friend)
- Real-time play over WebSocket

### 2. Single Player vs AI
- 3 difficulty levels: Easy, Medium, Hard
- Instant start (no waiting for opponent)
- Practice mode for learning controls

### 3. Local Multiplayer (PC only)
- Two players on same keyboard (split controls: WASD vs Arrow keys)
- Not available on mobile (screen too small for split view)

## Core Gameplay

### Court & View
- Isometric/angled top-down view (~45 degrees)
- Full court visible at all times
- Player's character at bottom, opponent at top
- Net clearly visible in center

### Movement
- 8-directional movement
- Auto-positioning assist: character auto-adjusts to reach ball within "reach zone"
- Mobile: virtual joystick (left thumb, floating)
- PC: WASD or Arrow keys

### Shot Types
- **Flat** (default): balanced speed and accuracy
- **Topspin** (hold longer): more curve and bounce
- **Slice/Drop** (short tap): slower ball, tight angle
- **Lob** (swipe up / modifier key): high arc over opponent
- Mobile: 3 shot buttons (right side, triangle layout) + swipe-up for lob
- PC: J/K/L keys or Z/X/C keys

### Serving
- Tap to toss ball, tap again to hit
- Timing window determines power/accuracy (moving meter like golf games)
- Service area alternates (left/right) each point

### Scoring (Simplified)
- First to 3 points wins a game
- First to 3 games wins a set
- Single set per match (no deuce/advantage)
- Matches last ~3-5 minutes (mobile-friendly session length)

## Battle Elements

### Energy System
- Energy meter fills as player rallies and hits good shots
- When full, player can activate a **Super Shot**
- Each character has a unique Super Shot

### Power-Ups
- Spawn randomly near the net area during rallies
- Player must approach net to collect (risk/reward)
- Types:
  - **Speed Boost**: temporary movement speed increase
  - **Big Racket**: wider hit zone for 10 seconds
  - **Ice Patch**: slows opponent movement for 5 seconds

### Characters
- 4 playable characters at launch
- Each has distinct stats: Speed / Power / Control (rated 1-5)
- Each has a unique Super Shot:
  - **Blaze**: Fireball flat shot (unreturable if opponent is far)
  - **Frost**: Freeze shot (slows ball on opponent's side, hard to time)
  - **Phantom**: Teleport return (player teleports to optimal position)
  - **Titan**: Shockwave serve (pushes opponent back on landing)

## Technical Requirements

### Performance
- 60 FPS rendering on mid-range mobile devices
- 30 tick/sec server simulation
- < 100ms perceived input latency

### Networking
- WebSocket for real-time game state sync
- Server-authoritative (all physics on server, client only renders)
- Client-side prediction for own player movement only
- Binary protocol for game state (minimize bandwidth)
- 10-second reconnection window on disconnect

### PWA
- Installable on mobile home screen
- Fullscreen landscape mode
- Orientation lock
- Offline: show "no connection" (game requires server)

### API
- REST endpoints for lobby/matchmaking/room management
- Single WebSocket endpoint for in-game communication
- Ticket-based WebSocket auth (REST returns short-lived token)

## Non-Functional Requirements

- No user accounts at v1 (anonymous play, session-based)
- No persistent progression/ranking at v1
- No in-app purchases
- No chat system at v1
