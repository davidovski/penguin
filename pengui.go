package main

import (
	_ "image/png"
    "image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const screen_w, screen_h = 640, 480

type Penguin struct {
    img *ebiten.Image

    x, y float64
    xv, yv float64
}

var penguin Penguin

func init() {
    penguin = Penguin{
        img: nil,
        x: screen_w / 2,
        y: screen_h / 2,
        xv: 0,
        yv: 0,
    }

	var err error
	penguin.img, _, err = ebitenutil.NewImageFromFile("assets/penguin.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct{}

func drawPenguin(screen *ebiten.Image, penguin Penguin) {
    op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(penguin.x, penguin.y)
    screen.DrawImage(penguin.img, op)
}

func updatePenguin(penguin *Penguin) {
    penguin.x = penguin.x + penguin.xv
    penguin.y = penguin.y + penguin.yv

    penguin.xv = penguin.xv / 2
    penguin.yv = penguin.yv / 2
}

func (g *Game) Update() error {
    if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
        penguin.xv = 4
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
        penguin.xv = -4
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
        penguin.yv = -4
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
        penguin.yv = 4
    }

    updatePenguin(&penguin)
        
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0xfe, 0xfe, 0xfe, 0xff})

    drawPenguin(screen, penguin)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("penguin")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
