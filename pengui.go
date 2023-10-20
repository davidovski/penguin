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

var red = color.RGBA{0xcc, 0x66, 0x66, 0xff}
var green = color.RGBA{0xb5, 0xbd, 0x68, 0xff}
const screen_w, screen_h = 640, 480
const gravity, friction, air_resistance = 0.9, 0.01, 0.01

type Penguin struct {
    img *ebiten.Image

    x, y float64
    xv, yv float64
    onGround bool
    fx []float64
    fy []float64
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

    const s float64 = 200

    tx, ty := 0.0, 0.0
    for i := range penguin.fx{
        x, y := penguin.fx[i], penguin.fy[i]
        tx += x
        ty += y

        vector.StrokeLine(screen, 
            float32(penguin.Cx()), float32(penguin.y+penguin.Height()), 
            float32(penguin.Cx() + x*s), float32(penguin.y + penguin.Height() + y*s), 
            4, green, true)
    }
    vector.StrokeLine(screen, 
        float32(penguin.Cx()), float32(penguin.y+penguin.Height()), 
        float32(penguin.Cx() + tx*s), float32(penguin.y + penguin.Height() + ty*s), 
        4, red, true)

}

func (penguin *Penguin) update(game *Game) {
    gx, gy := 0.0, gravity
    penguin.fx = []float64{gx}
    penguin.fy = []float64{gy}

    if penguin.onGround{
        angle := game.ground.angle(penguin.Cx(), 1)
        dx, dy := Normalize(game.ground.normal(penguin.Cx(), 1))

        nx, ny := Multiply(dx, dy, gravity * math.Cos(angle))

        sd := 0.0
        if penguin.xv > 0 {
            sd = 1
        } else if penguin.xv < 0 {
            sd = -1
        }

        rx, ry := Multiply(-sd*dy, sd*dx, -friction)
        penguin.fx = append(penguin.fx, nx, rx)
        penguin.fy = append(penguin.fy, ny, ry)
    }
    ax, ay := Multiply(penguin.xv, penguin.yv, -air_resistance)
    penguin.fx = append(penguin.fx, ax)
    penguin.fy = append(penguin.fy, ay)

    for _, v := range penguin.fx{
        penguin.xv += v
    }
    for _, v := range penguin.fy{
        penguin.yv += v
    }
    penguin.x += penguin.xv
    penguin.y += penguin.yv

    if penguin.y+penguin.Height() > game.ground(penguin.Cx()){
        penguin.y = game.ground(penguin.Cx()) - penguin.Height()
        penguin.onGround = true
    } else {
        penguin.onGround = false
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

func Angle(x, y float64) (float64) {
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
    if ebiten.IsMouseButtonPressed(ebiten.MouseButton0){
        cx, cy := ebiten.CursorPosition()
        g.penguin.x = float64(cx) - g.penguin.Width()/2
        g.penguin.y = float64(cy) - g.penguin.Height()/2
        g.penguin.xv = 0
        g.penguin.yv = 0
    }

    if g.penguin.onGround{
        if ebiten.IsKeyPressed(ebiten.KeyA) {
            g.penguin.xv -= 0.1
        }

        if ebiten.IsKeyPressed(ebiten.KeyD) {
            g.penguin.xv += 0.1
        }

        if ebiten.IsKeyPressed(ebiten.KeySpace) {
            g.penguin.onGround = false
            fx, fy := Normalize(g.ground.normal(g.penguin.Cx(), 2))
            const jumpHeight = 8

            var ex, ey float64 = 0, -1
            switch {
            case ebiten.IsKeyPressed(ebiten.KeyA):
                ex -= 1
            case ebiten.IsKeyPressed(ebiten.KeyD):
                ex += 1
            case ebiten.IsKeyPressed(ebiten.KeyS):
                ey += 1 
            }

            g.penguin.xv += (ex + fx) * jumpHeight
            g.penguin.yv += (ey + fy) * jumpHeight
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
    //ebiten.SetTPS(10)

    game := Game{}

    game.ground = func(x float64)(y float64) {
        //return 200-math.Pow((x - 320)/320, 3)+math.Pow((x-320), 2)/300
        //return 400-math.Pow((x-320)/20, 2)
        //return 400
        return 240+50*math.Cos(x/30)
    }

    game.init()
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
