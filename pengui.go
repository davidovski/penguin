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

const GROUND_RESISTANCE = 0.8
const AIR_RESISTANCE = 0.99
var red = color.RGBA{0xcc, 0x66, 0x66, 0xff}
var green = color.RGBA{0xb5, 0xbd, 0x68, 0xff}
const screen_w, screen_h = 640, 480

type Penguin struct {
    img *ebiten.Image

    x, y float64
    xv, yv float64
    onGround bool
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

func (p *Penguin) Width() (float64) {
    return float64(p.img.Bounds().Dx())
}

func (p *Penguin) Height() (float64) {
    return float64(p.img.Bounds().Dy())
}

func (p *Penguin) Cx() (float64) {
    return p.x + float64(p.img.Bounds().Dx()) / 2
}

func (p *Penguin) Cy() (float64) {
    return p.y + float64(p.img.Bounds().Dy()) / 2
}

func (penguin *Penguin) draw(screen *ebiten.Image, game Game) {
    op := &ebiten.DrawImageOptions{}

    if penguin.onGround {
        angle := game.ground.angle(penguin.Cx(), 2)
        tx, ty := float64(penguin.Width()) / 2, float64(penguin.Height())

        op.GeoM.Translate(-tx, -ty)
        op.GeoM.Rotate(angle)
        op.GeoM.Translate(tx, ty)
    }

    op.GeoM.Translate(penguin.x, penguin.y)
    screen.DrawImage(penguin.img, op)
}

func (penguin *Penguin) drawBounds(screen *ebiten.Image, game Game) {

    var m float64 = 16

    vector.StrokeLine(screen, 
        float32(penguin.Cx()), float32(penguin.y+penguin.Height()), 
        float32(penguin.Cx() + penguin.xv*m), float32(penguin.y + penguin.Height() + penguin.yv*m), 
        4, red, true)

    if penguin.onGround{
        fx, fy := Normalize(game.ground.normal(game.penguin.Cx(), 2))
        vector.StrokeLine(screen, 
            float32(penguin.Cx()), float32(penguin.y+penguin.Height()), 
            float32(penguin.Cx() + fx*penguin.Height()), float32(penguin.y + penguin.Height() + fy*penguin.Height()), 
            4, green, true)
        return
    }

    vector.StrokeRect(screen,
        float32(penguin.x), float32(penguin.y),
        float32(penguin.Width()), float32(penguin.Height()),
        4, red, true)

    
}

func (penguin *Penguin) update(g *Game) {
    if penguin.onGround{
        penguin.yv = 0
        penguin.y = g.ground(penguin.Cx()) - float64(penguin.img.Bounds().Dy())
        penguin.x += penguin.xv
        penguin.y += penguin.yv

        dx, dy := Normalize(g.ground.normal(penguin.x, 1))
        penguin.xv += math.Sqrt(dx*dx + dy*dy) * math.Sin(g.ground.angle(penguin.Cx(), 2)) * 4
        //penguin.yv += math.Sqrt(dx*dx + dy*dy) * math.Cos(g.ground.angle(penguin.Cx(), 2)) * 4
        penguin.xv *= GROUND_RESISTANCE

        return
    }

    penguin.yv += 0.9

    penguin.x += penguin.xv
    penguin.y += penguin.yv

    penguin.xv *= AIR_RESISTANCE
    penguin.yv *= AIR_RESISTANCE

    if penguin.collideWithCurve(g.ground, 0, penguin.yv){
        penguin.yv = 0
        penguin.onGround = true
        return
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

func (curve Curve) delta(x, r float64) (float64, float64){
    x1, x2 := x + r, x - r
    y1, y2 := curve(x1), curve(x2)
    dx, dy := x2 - x1, y2 - y1
    return dx, dy
}

func (curve Curve) angle(x, r float64) (float64){
    dx, dy := curve.delta(x, r)
    return math.Atan(dy / dx)
}

func (curve Curve) normal(x, r float64) (float64, float64){
    dx, dy := curve.delta(x, r)
    //m := -(dx / dy)
    return -dy,dx 
}

func Normalize(x, y float64) (float64, float64) {
    m := math.Sqrt(x*x + y*y)
    return x/m, y/m
}

type Game struct{
    penguin Penguin
    ground Curve
}

func (g *Game) Update() error {
    if g.penguin.onGround{
        if ebiten.IsKeyPressed(ebiten.KeyA) {
            g.penguin.xv = -4
        }

        if ebiten.IsKeyPressed(ebiten.KeyD) {
            g.penguin.xv = 4
        }

        if ebiten.IsKeyPressed(ebiten.KeySpace) {
            g.penguin.onGround = false
            fx, fy := Normalize(g.ground.normal(g.penguin.Cx(), 2))
            const jumpHeight = 8
            g.penguin.xv += fx * jumpHeight
            g.penguin.yv += fy * jumpHeight
        }
    }

    g.penguin.update(g)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{0xfe, 0xfe, 0xfe, 0xff})

    g.ground.draw(screen, color.RGBA{0x19, 0x19, 0x19, 0xff}, 4, 16, true);
    g.penguin.draw(screen, *g)
    g.penguin.drawBounds(screen, *g)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("penguin")
    ebiten.SetTPS(10)

    game := Game{}

    game.ground = func(x float64)(y float64) {
        //return 200-math.Pow((x - 320)/320, 3)+math.Pow((x-320), 2)/300
        //return 400
        return 400+50*math.Sin(x/50)
    }

    game.init()
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
