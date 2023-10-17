package main

import (
	_ "image/png"
    "image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var img *ebiten.Image

var x float64 = 0
var y float64 = 0

func init() {
	var err error
	img, _, err = ebitenutil.NewImageFromFile("assets/penguin.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct{}

func (g *Game) Update() error {
    if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
        x = x + 1
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
        x = x - 1
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
        y = y - 1
    }
    if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
        y = y + 1
    }
    log.Printf("%f %f\n", x, y)
        
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0xfe, 0xfe, 0xfe, 0xff})

    op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
    screen.DrawImage(img, op)
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
