import { describe, expect, it } from 'vitest'

import {
  BOOT_SCENE_KEY,
  GAME_BACKGROUND_COLOR,
  GAME_HEIGHT,
  GAME_PARENT_ID,
  GAME_WIDTH,
} from './gameSettings'

describe('game scaffold', () => {
  it('exposes a landscape boot configuration', () => {
    expect(GAME_WIDTH).toBeGreaterThan(GAME_HEIGHT)
    expect(GAME_PARENT_ID).toBe('game')
    expect(GAME_BACKGROUND_COLOR).toMatch(/^#/)
    expect(BOOT_SCENE_KEY).toBe('boot')
  })
})
