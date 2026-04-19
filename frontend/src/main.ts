import Phaser from 'phaser'

const config: Phaser.Types.Core.GameConfig = {
  type: Phaser.AUTO,
  width: 960,
  height: 540,
  parent: 'game',
  backgroundColor: '#2d8b46',
  scene: {
    create() {
      this.add.text(480, 270, 'Tennis Battle', {
        fontSize: '32px',
        color: '#ffffff',
      }).setOrigin(0.5)
    },
  },
  scale: {
    mode: Phaser.Scale.FIT,
    autoCenter: Phaser.Scale.CENTER_BOTH,
  },
}

new Phaser.Game(config)
