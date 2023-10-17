package main

import (
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const screen_w, screen_h = 640, 480

type Penguin struct {
    img *ebiten.Image

    x, y float64
    xv, yv float64
}

func newPenguin() Penguin {
    penguin := Penguin{
        x: screen_w / 2,
        y: 0,
        xv: 0,
        yv: 0,
    }

	var err error
	penguin.img, _, err = ebitenutil.NewImageFromFile("assets/penguin.png")
	if err != nil {
		log.Fatal(err)
	}

    return penguin;
}

func (penguin *Penguin) draw(screen *ebiten.Image) {
    op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(penguin.x, penguin.y)
    screen.DrawImage(penguin.img, op)
}

func (penguin *Penguin) drawBounds(screen *ebiten.Image) {
    vector.StrokeRect(screen,
        float32(penguin.x), float32(penguin.y),
        float32(penguin.img.Bounds().Dx()), float32(penguin.img.Bounds().Dy()),
        4, color.RGBA{0xcc, 0x66, 0x66, 0xff}, true)
}

func (penguin *Penguin) update(g *Game) {
    if penguin.collideWithCurve(g.ground, penguin.xv, 0){
        penguin.xv = 0
        penguin.yv -= 4
    } else {
        penguin.x += penguin.xv
        penguin.xv *= 0.9
    }

    if penguin.collideWithCurve(g.ground, 0, penguin.yv){
        penguin.yv = 0
    } else {
        penguin.y += penguin.yv
        penguin.yv += 0.9
    }
}

func (p *Penguin) collideWithCurve(curve Curve, xv, yv float64) bool {
    cy1 := curve(p.x + xv)
    y1 := p.y + yv
    y2 := p.y + yv + float64(p.img.Bounds().Dy())

    if y1 < cy1 && cy1 < y2 {
        return true
    }

    cy2 := curve(p.x + xv + float64(p.img.Bounds().Dx()))
    if y1 < cy2 && cy2 < y2 {
        return true
    }

    return false
}

func (g *Game) init() {
    g.penguin = newPenguin()
}

type Curve func(x float64)(y float64)

func (curve Curve) draw(dst *ebiten.Image, clr color.Color, width float32, resolution int, antialias bool) {
    for x1 := 0; x1 < screen_w; x1 += resolution {
        x2 := x1 + resolution
        y1 := curve(float64(x1))
        y2 := curve(float64(x2))

        vector.StrokeLine(dst, float32(x1), float32(y1), float32(x2), float32(y2), width, color.RGBA{0x19, 0x19, 0x19, 0xff}, true)
    }
}

type Game struct{
    penguin Penguin
    ground Curve
}

func (g *Game) Update() error {
    if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
        g.penguin.xv = 4
    }

    if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
        g.penguin.xv = -4
    }

    if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
        g.penguin.yv = -4
    }

    if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
        g.penguin.yv = 4
    }

    g.penguin.update(g)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0xfe, 0xfe, 0xfe, 0xff})

    g.ground.draw(screen, color.RGBA{0x19, 0x19, 0x19, 0xff}, 4, 16, true);
    g.penguin.draw(screen)
    g.penguin.drawBounds(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("penguin")

    game := Game{}

    game.ground = func(x float64)(y float64) {
        return 200-math.Pow((x - 320)/320, 3)+math.Pow((x-320), 2)/300
    }

    game.init()
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
