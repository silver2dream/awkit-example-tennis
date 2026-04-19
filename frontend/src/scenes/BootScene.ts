import Phaser from 'phaser'

import { BOOT_SCENE_KEY } from '../gameSettings'

export class BootScene extends Phaser.Scene {
  constructor() {
    super(BOOT_SCENE_KEY)
  }

  create(): void {
    // Intentionally empty: this step only needs a bootable Phaser canvas.
  }
}
