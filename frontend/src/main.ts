import Phaser from 'phaser'

import {
  GAME_BACKGROUND_COLOR,
  GAME_HEIGHT,
  GAME_PARENT_ID,
  GAME_WIDTH,
} from './gameSettings'
import { BootScene } from './scenes/BootScene'

export const gameConfig: Phaser.Types.Core.GameConfig = {
  type: Phaser.AUTO,
  width: GAME_WIDTH,
  height: GAME_HEIGHT,
  parent: GAME_PARENT_ID,
  backgroundColor: GAME_BACKGROUND_COLOR,
  scene: [BootScene],
  scale: {
    mode: Phaser.Scale.FIT,
    autoCenter: Phaser.Scale.CENTER_BOTH,
  },
}

new Phaser.Game(gameConfig)
