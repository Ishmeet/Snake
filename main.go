package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 480
	gridSize     = 10
	xNumInScreen = screenWidth / gridSize
	yNumInScreen = screenHeight / gridSize
)

const (
	dirNone = iota
	dirLeft
	dirRight
	dirDown
	dirUp
)

type Position struct {
	X int
	Y int
}

type Game struct {
	moveDirection int
	snakeBody     []Position
	apple         Position
	timer         int
	moveTime      int
	score         int
	bestScore     int
	level         int
	gameMode      bool
	prevLength    int
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (g *Game) collidesWithApple() bool {
	return g.snakeBody[0].X == g.apple.X &&
		g.snakeBody[0].Y == g.apple.Y
}

func (g *Game) collidesWithSelf() bool {
	for _, v := range g.snakeBody[1:] {
		if g.snakeBody[0].X == v.X &&
			g.snakeBody[0].Y == v.Y {
			return true
		}
	}
	return false
}

func (g *Game) collidesWithWall() bool {
	return g.snakeBody[0].X < 0 ||
		g.snakeBody[0].Y < 0 ||
		g.snakeBody[0].X >= xNumInScreen ||
		g.snakeBody[0].Y >= yNumInScreen
}

func (g *Game) needsToMoveSnake() bool {
	return g.timer%g.moveTime == 0
}

func (g *Game) reset() {
	g.apple.X = 3 * gridSize
	g.apple.Y = 3 * gridSize
	g.moveTime = 4
	g.snakeBody = g.snakeBody[:1]
	g.snakeBody[0].X = xNumInScreen / 2
	g.snakeBody[0].Y = yNumInScreen / 2
	g.score = 0
	g.level = 1
	g.moveDirection = dirNone
}

func (g *Game) AIMovement() {
	// AI movement
	if g.gameMode {
		length := int(math.Hypot(float64(g.snakeBody[0].X*gridSize-g.apple.X*gridSize),
			float64(g.snakeBody[0].Y*gridSize-g.apple.Y*gridSize)))
		if g.prevLength == 0 {
			g.prevLength = length
		}
		if length < g.prevLength && length != gridSize {
			// Keep on moving in the same direction
		} else {
			// Find if we have to move up/down/left/right.
			switch g.moveDirection {
			case dirRight:
				if g.apple.Y > g.snakeBody[0].Y {
					g.moveDirection = dirDown
				} else {
					g.moveDirection = dirUp
				}
			case dirLeft:
				if g.apple.Y > g.snakeBody[0].Y {
					g.moveDirection = dirDown
				} else {
					g.moveDirection = dirUp
				}
			case dirDown:
				if g.apple.X > g.snakeBody[0].X {
					g.moveDirection = dirRight
				} else {
					g.moveDirection = dirLeft
				}
			case dirUp:
				if g.apple.X > g.snakeBody[0].X {
					g.moveDirection = dirRight
				} else {
					g.moveDirection = dirLeft
				}
			}
		}
		g.prevLength = length
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		if g.moveDirection != dirRight {
			g.moveDirection = dirLeft
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		if g.moveDirection != dirLeft {
			g.moveDirection = dirRight
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if g.moveDirection != dirUp {
			g.moveDirection = dirDown
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if g.moveDirection != dirDown {
			g.moveDirection = dirUp
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.reset()
	} else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.gameMode = !g.gameMode
	}

	if g.needsToMoveSnake() {
		g.AIMovement()

		if g.collidesWithWall() || g.collidesWithSelf() {
			g.reset()
		}

		if g.collidesWithApple() {
			g.apple.X = rand.Intn(xNumInScreen - 1)
			g.apple.Y = rand.Intn(yNumInScreen - 1)
			g.snakeBody = append(g.snakeBody, Position{
				X: g.snakeBody[len(g.snakeBody)-1].X,
				Y: g.snakeBody[len(g.snakeBody)-1].Y,
			})
			if len(g.snakeBody) > 10 && len(g.snakeBody) < 20 {
				g.level = 2
				g.moveTime = 3
			} else if len(g.snakeBody) > 20 {
				g.level = 3
				g.moveTime = 2
			} else {
				g.level = 1
			}
			g.score++
			if g.bestScore < g.score {
				g.bestScore = g.score
			}
		}

		for i := int64(len(g.snakeBody)) - 1; i > 0; i-- {
			g.snakeBody[i].X = g.snakeBody[i-1].X
			g.snakeBody[i].Y = g.snakeBody[i-1].Y
		}
		switch g.moveDirection {
		case dirLeft:
			g.snakeBody[0].X--
		case dirRight:
			g.snakeBody[0].X++
		case dirDown:
			g.snakeBody[0].Y++
		case dirUp:
			g.snakeBody[0].Y--
		}
	}

	g.timer++

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, v := range g.snakeBody {
		ebitenutil.DrawRect(screen, float64(v.X*gridSize), float64(v.Y*gridSize), gridSize, gridSize, color.RGBA{0x80, 0xa0, 0xc0, 0xff})
	}
	ebitenutil.DrawRect(screen, float64(g.apple.X*gridSize), float64(g.apple.Y*gridSize), gridSize, gridSize, color.RGBA{0xFF, 0x00, 0x00, 0xff})

	if g.moveDirection == dirNone {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Press up/down/left/right to start"))
	} else {
		msg := func() string {
			if g.gameMode {
				return "AI"
			}
			return "Manual"
		}()
		ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f Level: %d Score: %d Best Score: %d, %s, Len: %d", ebiten.CurrentFPS(), g.level, g.score, g.bestScore, msg, int(math.Hypot(float64(g.snakeBody[0].X*gridSize-g.apple.X*gridSize), float64(g.snakeBody[0].Y*gridSize-g.apple.Y*gridSize)))))
	}
	ebitenutil.DrawLine(screen, float64(g.snakeBody[0].X*gridSize), float64(g.snakeBody[0].Y*gridSize), float64(g.apple.X*gridSize), float64(g.apple.Y*gridSize), color.RGBA{0x00, 0x00, 0xFF, 0xFF})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func newGame() *Game {
	g := &Game{
		apple:     Position{X: 3 * gridSize, Y: 3 * gridSize},
		moveTime:  4,
		snakeBody: make([]Position, 1),
	}
	g.snakeBody[0].X = xNumInScreen / 2
	g.snakeBody[0].Y = yNumInScreen / 2
	return g
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake (Ebiten Demo)")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
