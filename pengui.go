package main

import (
	"image/color"
	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var RED = color.RGBA{0xcc, 0x66, 0x66, 0xff}
var GREEN = color.RGBA{0xb5, 0xbd, 0x68, 0xff}
var BLACK = color.RGBA{0x19, 0x19, 0x19, 0xff}
var WHITE = color.RGBA{0xfe, 0xfe, 0xfe, 0xff}
var BLUE = color.RGBA{0x81, 0xa2, 0xbe, 0xff}

const screen_w, screen_h = 960, 540

const GRAVITY, FRICTION, AIR_RESISTANCE, PUSH, JUMP_HEIGHT = 1.2, 0.1, 0.02, 0.8, 18
const CAMERA_DELAY, DELTA_RESOLUTION, DRAW_RESOLUTION, STROKE_WIDTH = 0.7, 1.0, 2.0, 4.0

var debug = false

var sx, sy, sxv, syv = 0.0, 0.0, 0.0, 0.0

type Penguin struct {
    img *ebiten.Image

    x, y float64
    xv, yv float64
    onGround bool
    fx, fy []float64
    airTime float64
    angle float64
}

func newPenguin() Penguin {
    penguin := Penguin{
        x: -200,
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

func (penguin *Penguin) Draw(screen *ebiten.Image, game Game) {
    op := &ebiten.DrawImageOptions{}

    if penguin.airTime < 60 {
        tx, ty := float64(penguin.Width()) / 2, float64(penguin.Height())

        op.GeoM.Translate(-tx, -ty)
        op.GeoM.Rotate(penguin.angle)
        op.GeoM.Translate(tx, ty)
    }

    op.GeoM.Translate(sx + penguin.x, sy + penguin.y)
    screen.DrawImage(penguin.img, op)
}

func (penguin *Penguin) DrawForces(screen *ebiten.Image, game Game) {

    const s float64 = 60

    tx, ty := 0.0, 0.0
    for i := range penguin.fx{
        x, y := penguin.fx[i], penguin.fy[i]
        tx += x
        ty += y

        vector.StrokeLine(screen,
            float32(sx + penguin.Cx()), float32(sy + penguin.y+penguin.Height()),
            float32(sx + penguin.Cx() + x*s), float32(sy + penguin.y + penguin.Height() + y*s),
            STROKE_WIDTH, GREEN, true)
    }

    vector.StrokeLine(screen,
        float32(sx + penguin.Cx()), float32(sy + penguin.y+penguin.Height()),
        float32(sx + penguin.Cx() + tx*s), float32(sy + penguin.y + penguin.Height() + ty*s),
        STROKE_WIDTH, RED, true)

}

func (penguin *Penguin) CalculateForces(game *Game) {
    gx, gy := 0.0, GRAVITY
    penguin.fx = []float64{gx}
    penguin.fy = []float64{gy}

    if penguin.onGround{

        angle := game.ground.angle(penguin.Cx())
        penguin.angle = angle
        dx, dy := Normalize(game.ground.normal(penguin.Cx()))

        nx, ny := Multiply(dx, dy, GRAVITY * math.Cos(angle))

        sd := 0.0
        if penguin.xv > 0 {
            sd = 1
        } else if penguin.xv < 0 {
            sd = -1
        }

        rx, ry := Multiply(-sd*dy, sd*dx, -FRICTION * GRAVITY)
        penguin.fx = append(penguin.fx, nx, rx)
        penguin.fy = append(penguin.fy, ny, ry)
    }
    ax, ay := Multiply(penguin.xv, penguin.yv, -AIR_RESISTANCE)
    penguin.fx = append(penguin.fx, ax)
    penguin.fy = append(penguin.fy, ay)
}

func (penguin *Penguin) Update(game *Game) {
    // add the forces to the current velocity
    for _, v := range penguin.fx{
        penguin.xv += v
    }
    for _, v := range penguin.fy{
        penguin.yv += v
    }

    // add the current velocity to the penguin
    penguin.x += penguin.xv
    penguin.y += penguin.yv

    // check if the penguin is on the ground
    if penguin.y+penguin.Height() > game.ground(penguin.Cx()){
        penguin.y = game.ground(penguin.Cx()) - penguin.Height()
        penguin.onGround = true
        penguin.airTime = 0
    } else {
        penguin.onGround = false
        penguin.airTime += 1
    }

}

func (p *Penguin) collideWithCurve(curve Curve, xv, yv float64) bool {
    cy1 := curve(p.Cx() + xv)
    y1 := p.y + yv
    y2 := yv + p.y + p.Height()

    if y1 < cy1 && cy1 < y2 {
        return true
    }

    return false
}

func (g *Game) init() {
    g.penguin = newPenguin()
}

type Curve func(x float64)(y float64)

func (curve Curve) draw(dst *ebiten.Image, clr color.Color, width float32, resolution float64, antialias bool) {
    for x1 := float64(-sx); x1 < float64(screen_w - sx); x1 += resolution {
        x2 := x1 + resolution
        y1 := curve(x1)
        y2 := curve(x2)

        ry := y1
        if y2 > y1{
            ry = y2
        }
        vector.DrawFilledRect(dst, float32(sx + x1), float32(sy + ry), float32(resolution), float32(screen_h - (sy + ry)), WHITE, true)
        vector.StrokeLine(dst, float32(sx + x1), float32(sy + y1), float32(sx + x2), float32(sy + y2), width, clr, true)
    }
}

func (curve Curve) delta(x float64) (float64, float64){
    x1, x2 := x + DELTA_RESOLUTION, x - DELTA_RESOLUTION
    y1, y2 := curve(x1), curve(x2)
    dx, dy := x2 - x1, y2 - y1
    return dx, dy
}

func (curve Curve) angle(x float64) (float64){
    dx, dy := curve.delta(x)
    return Angle(dx, dy)
}

func (curve Curve) normal(x float64) (float64, float64){
    dx, dy := curve.delta(x)
    return -dy,dx
}

func Normalize(x, y float64) (float64, float64) {
    m := math.Sqrt(x*x + y*y)
    return x/m, y/m
}

func Angle(x, y float64) (float64) {
    if x == 0 {
        return 0
    }
    return math.Atan(y / x)
}

func Multiply(x, y, c float64) (float64, float64) {
    return x*c, y*c
}

func Magnitude(x, y float64) (float64) {
    return math.Sqrt(x*x + y*y)
}

type Game struct{
    penguin Penguin
    ground Curve
}

func (g *Game) Update() error {

    if inpututil.IsKeyJustPressed(ebiten.KeyB) {
        debug = !debug;
    }

    if ebiten.IsMouseButtonPressed(ebiten.MouseButton0){
        cx, cy := ebiten.CursorPosition()
        g.penguin.x = float64(cx) - g.penguin.Width()/2
        g.penguin.y = float64(cy) - g.penguin.Height()/2
        g.penguin.xv = 0
        g.penguin.yv = 0
    }

    g.penguin.CalculateForces(g)
    if g.penguin.onGround{
        if ebiten.IsKeyPressed(ebiten.KeyA) {
            dx, dy := Normalize(g.ground.delta(g.penguin.Cx()))
            g.penguin.fx = append(g.penguin.fx, dx * PUSH)
            g.penguin.fy = append(g.penguin.fy, dy * PUSH)
        }

        if ebiten.IsKeyPressed(ebiten.KeyD) {
            dx, dy := Normalize(g.ground.delta(g.penguin.Cx()))
            g.penguin.fx = append(g.penguin.fx, dx * -PUSH)
            g.penguin.fy = append(g.penguin.fy, dy * -PUSH)
        }

        if ebiten.IsKeyPressed(ebiten.KeySpace) {
            g.penguin.onGround = false
            jx, jy := Normalize(g.ground.normal(g.penguin.Cx()))
            g.penguin.fx = append(g.penguin.fx, jx * JUMP_HEIGHT)
            g.penguin.fy = append(g.penguin.fy, jy * JUMP_HEIGHT)
        }
    }

    g.penguin.Update(g)

    // calculate scroll x and y

    if debug{
        sx = -(g.penguin.Cx() - screen_w/2)
        sy = -(g.penguin.Cy() - screen_h/2)
    } else {
        sx = (CAMERA_DELAY * sx - (1-CAMERA_DELAY) * (g.penguin.Cx() - screen_w/2))
        sy = (CAMERA_DELAY * sy - (1-CAMERA_DELAY) * (g.penguin.Cy() - screen_h/2))
    }

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Fill(BLUE)

    g.ground.draw(screen, BLACK, STROKE_WIDTH, DRAW_RESOLUTION, true);
    g.penguin.Draw(screen, *g)
    if debug {
        g.penguin.DrawForces(screen, *g)
    }
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screen_w, screen_h
}

func main() {
	ebiten.SetWindowSize(screen_w, screen_h)
	ebiten.SetWindowTitle("penguin")
    ebiten.SetTPS(60)

    game := Game{}

    game.ground = func(x float64)(y float64) {
        //return 200-math.Pow(x/200, 3)+math.Pow((x-320), 2)/300
        if x < 0 {
            return 0
        }
        return 500+50*-math.Cos(x/400*(math.Cos(x/4870)+1))*math.Sin(x/845*(math.Cos(x/547)+1))*math.Cos((x-400)/400) + -500*math.Cos(x/500)
    }

    game.init()
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
