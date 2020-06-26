package main

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/vorbis"
	"github.com/hajimehoshi/ebiten/audio/wav"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth         = 640
	screenHeight        = 480
	sampleRate          = 44100
	introLengthInSecond = 5
	loopLengthInSecond  = 4
)

const (
	bump   = 0
	crunch = 1
)

var (
	gophersImage *ebiten.Image
)
var (
	emptyImage *ebiten.Image
)

// Turn ...
type Turn struct {
	X          int64
	Y          int64
	W          int64
	H          int64
	isItTurned bool
	direction  int8
}

// Scale ...
type Scale struct {
	X int64
	Y int64
}

// Game ...
type Game struct {
	count                  int
	verticle               int64
	horizontal             int64
	width                  int64
	height                 int64
	moveDirection          int8
	AppleX                 int64
	AppleY                 int64
	countEat               int64
	turn                   Turn
	scalesLoc              map[int64]*Scale
	timer                  uint64
	speed                  int8
	turn2                  map[int64]*Turn
	turnKey                int64
	Score                  int64
	Level                  int8
	nearapple              bool
	soundEnable            bool
	audioContext           *audio.Context
	audioPlayer            *audio.Player
	audioPlayerCrunch      *audio.Player
	audioContextBackground *audio.Context
	audioPlayerBackground  *audio.Player
	snakeImage             *ebiten.Image
	snakeSkin              *ebiten.Image
	snakeHead              *ebiten.Image
	snakeHeadDown          *ebiten.Image
	snakeHeadUp            *ebiten.Image
	snakeHeadRight         *ebiten.Image
	snakeMouth             *ebiten.Image
	apple                  *ebiten.Image
}

func init() {
	emptyImage, _ = ebiten.NewImage(1, 1, ebiten.FilterDefault)
	_ = emptyImage.Fill(color.White)
}

func (g *Game) collision() bool {
	if g.scalesLoc[0].X < g.AppleX+10 &&
		g.scalesLoc[0].X+10 > g.AppleX &&
		g.scalesLoc[0].Y < g.AppleY+10 &&
		g.scalesLoc[0].Y+10 > g.AppleY {
		// collision detected!
		return true
	}
	return false
}

func (g *Game) nearApple() bool {
	if g.scalesLoc[0].X < (g.AppleX-30)+60 &&
		g.scalesLoc[0].X+10 > g.AppleX-30 &&
		g.scalesLoc[0].Y < (g.AppleY-30)+50 &&
		g.scalesLoc[0].Y+10 > g.AppleY-30 {
		// collision detected!
		return true
	}
	return false
}

func (g *Game) selfCollision() bool {
	for i, v := range g.scalesLoc {
		if i > 0 {
			if g.scalesLoc[0].X < v.X+10 &&
				g.scalesLoc[0].X+10 > v.X &&
				g.scalesLoc[0].Y < v.Y+10 &&
				g.scalesLoc[0].Y+10 > v.Y {
				// collision detected!
				return true
			}
		}
	}
	return false
}

func (g *Game) wallCollision() bool {
	if g.scalesLoc[0].X > screenWidth/2 ||
		g.scalesLoc[0].X < -(screenWidth/2) ||
		g.scalesLoc[0].Y > screenHeight/2 ||
		g.scalesLoc[0].Y < -(screenHeight/2) {
		return true
	}
	return false
}

func (g *Game) comeOutOtherEnd() {
	for _, v := range g.scalesLoc {
		if v.X > (screenWidth / 2) {
			v.X = -(screenWidth / 2)
		} else if v.X < -(screenWidth / 2) {
			v.X = screenWidth / 2
		}
		if v.Y > screenHeight/2 {
			v.Y = -screenHeight / 2
		} else if v.Y < -(screenHeight / 2) {
			v.Y = screenHeight / 2
		}
	}
}

func (g *Game) updateTimer() {
	if g.timer > 0xFFFFFFFFFFFFFFFE {
		g.timer = 0
	}
	g.timer++
}

func (g *Game) moveSnakeTimer() bool {
	if g.timer%uint64(g.speed) == 0 {
		return true
	}
	return false
}

func (g *Game) turnFunction(dir1 int8, dir2 int8, moveDirection int8) {
	if len(g.scalesLoc) >= 2 {
		if g.moveDirection == dir1 || g.moveDirection == dir2 {
			g.turn2[g.turnKey] = &Turn{
				X:         g.scalesLoc[0].X,
				Y:         g.scalesLoc[0].Y,
				direction: moveDirection,
			}
			g.turnKey++
		}
	}
}

func (g *Game) reset() {
	g.AppleX = 30
	g.AppleY = 30
	g.speed = 5 // means 83.33 ms
	l := len(g.scalesLoc)
	for i := int64(1); i < int64(l); i++ {
		delete(g.scalesLoc, i)
	}
	g.scalesLoc[0].X = 0
	g.scalesLoc[0].Y = 0
	g.turnKey = 0
	g.Score = 0
	g.Level = 1
	g.moveDirection = 0
	g.audioPlayerBackground.Pause()
}

// Update ...
func (g *Game) Update(screen *ebiten.Image) error {

	if g.moveSnakeTimer() {
		if g.selfCollision() {
			if g.soundEnable {
				g.audioPlayer.Rewind()
				g.audioPlayer.Play()
			} else {
				g.audioPlayer.Pause()
			}
			g.reset()
		}

		g.nearapple = g.nearApple()

		if g.collision() {
			g.AppleX = rand.Int63n(screenWidth-20) + -(screenWidth / 2)
			g.AppleY = rand.Int63n(screenHeight-20) + -(screenHeight / 2)
			g.scalesLoc[int64(len(g.scalesLoc))] = &Scale{
				X: g.scalesLoc[int64(len(g.scalesLoc)-1)].X,
				Y: g.scalesLoc[int64(len(g.scalesLoc)-1)].Y,
			}
			if len(g.scalesLoc) > 7 && len(g.scalesLoc) < 15 {
				g.Score += 100
				g.Level = 2
				g.speed = 4
			} else if len(g.scalesLoc) > 14 && len(g.scalesLoc) < 21 {
				g.Score += 300
				g.Level = 3
				g.speed = 3
			} else if len(g.scalesLoc) > 20 && len(g.scalesLoc) < 30 {
				g.Score += 500
				g.Level = 4
				g.speed = 2
			} else if len(g.scalesLoc) > 29 {
				g.Score += 500
				g.Level = 5
				g.speed = 1
			} else {
				g.Score += 10
				g.Level = 1
			}
			if g.soundEnable {
				g.audioPlayerCrunch.Rewind()
				g.audioPlayerCrunch.Play()
			} else {
				g.audioPlayerCrunch.Pause()
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		if g.moveDirection != 2 {
			g.moveDirection = 1
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		if g.moveDirection != 1 {
			g.moveDirection = 2
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		if g.moveDirection != 4 {
			g.moveDirection = 3
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		if g.moveDirection != 3 {
			g.moveDirection = 4
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.reset()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.soundEnable = !g.soundEnable
	}

	switch g.moveDirection {
	case 1:
		if g.moveSnakeTimer() {
			for i := int64(len(g.scalesLoc)) - 1; i > 0; i-- {
				g.scalesLoc[i].X = g.scalesLoc[i-1].X
				g.scalesLoc[i].Y = g.scalesLoc[i-1].Y
			}
			g.scalesLoc[0].X = g.scalesLoc[0].X - 10
		}
	case 2:
		if g.moveSnakeTimer() {
			for i := int64(len(g.scalesLoc)) - 1; i > 0; i-- {
				g.scalesLoc[i].X = g.scalesLoc[i-1].X
				g.scalesLoc[i].Y = g.scalesLoc[i-1].Y
			}
			g.scalesLoc[0].X += 10
		}
	case 3:

		if g.moveSnakeTimer() {
			for i := int64(len(g.scalesLoc)) - 1; i > 0; i-- {
				g.scalesLoc[i].X = g.scalesLoc[i-1].X
				g.scalesLoc[i].Y = g.scalesLoc[i-1].Y
			}
			g.scalesLoc[0].Y += 10
		}
	case 4:
		if g.moveSnakeTimer() {
			for i := int64(len(g.scalesLoc)) - 1; i > 0; i-- {
				g.scalesLoc[i].X = g.scalesLoc[i-1].X
				g.scalesLoc[i].Y = g.scalesLoc[i-1].Y
			}
			g.scalesLoc[0].Y -= 10
		}
	}

	// g.comeOutOtherEnd()
	if g.wallCollision() {
		if g.soundEnable {
			g.audioPlayer.Rewind()
			g.audioPlayer.Play()
		}
		g.reset()
	}
	g.updateTimer()

	return nil
}

// Draw ...
func (g *Game) Draw(screen *ebiten.Image) {

	backgroundColor(screen)

	if g.moveDirection == 0 {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Press up/down/left/right to start, M to Enable/Disable Sound"))

		w, h := g.snakeImage.Size()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
		op.GeoM.Translate(screenWidth/2, screenHeight/2)
		screen.DrawImage(g.snakeImage, op)

	} else {
		if g.nearapple {
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Level: %d Score: %d Near", g.Level, g.Score))
		} else {
			ebitenutil.DebugPrint(screen, fmt.Sprintf("Level: %d Score: %d", g.Level, g.Score))
		}
		if g.soundEnable {
			g.audioPlayerBackground.Play()
		} else {
			g.audioPlayerBackground.Pause()
		}
	}

	for i, v := range g.scalesLoc {
		if i == 0 {
			if g.nearapple {
				w, h := g.snakeMouth.Size()
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(float64(10)/float64(w), float64(10)/float64(h))
				op.GeoM.Translate((screenWidth/2)+float64(v.X), (screenHeight/2)+float64(v.Y))
				screen.DrawImage(g.snakeMouth, op)
			} else {
				switch g.moveDirection {
				case 1:
					w, h := g.snakeHead.Size()
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(float64(10)/float64(w), float64(10)/float64(h))
					op.GeoM.Translate((screenWidth/2)+float64(v.X), (screenHeight/2)+float64(v.Y))
					screen.DrawImage(g.snakeHead, op)
				case 2:
					w, h := g.snakeHeadRight.Size()
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(float64(10)/float64(w), float64(10)/float64(h))
					op.GeoM.Translate((screenWidth/2)+float64(v.X), (screenHeight/2)+float64(v.Y))
					screen.DrawImage(g.snakeHeadRight, op)
				case 3:
					w, h := g.snakeHeadDown.Size()
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(float64(10)/float64(w), float64(10)/float64(h))
					op.GeoM.Translate((screenWidth/2)+float64(v.X), (screenHeight/2)+float64(v.Y))
					screen.DrawImage(g.snakeHeadDown, op)
				case 4:
					w, h := g.snakeHeadUp.Size()
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(float64(10)/float64(w), float64(10)/float64(h))
					op.GeoM.Translate((screenWidth/2)+float64(v.X), (screenHeight/2)+float64(v.Y))
					screen.DrawImage(g.snakeHeadUp, op)
				}

			}
		} else {
			w, h := g.snakeSkin.Size()
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(10)/float64(w), float64(10)/float64(h))
			op.GeoM.Translate((screenWidth/2)+float64(v.X), (screenHeight/2)+float64(v.Y))
			screen.DrawImage(g.snakeSkin, op)
		}
	}

	w, h := g.apple.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(10)/float64(w), float64(10)/float64(h))
	op.GeoM.Translate((screenWidth/2)+float64(g.AppleX), (screenHeight/2)+float64(g.AppleY))
	screen.DrawImage(g.apple, op)

}

func backgroundColor(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(screenWidth, screenHeight)
	op.GeoM.Translate(0, 0)
	op.ColorM.Scale(62, 66, 46, 0.1)
	_ = screen.DrawImage(emptyImage, op)
}

// Layout ...
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) setupAudio() error {
	var err error
	// Initialize audio context.
	g.audioContext, err = audio.NewContext(sampleRate)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("jab.wav")
	if err != nil {
		return err
	}

	f2, err := os.Open("crunchWav.wav")
	if err != nil {
		return err
	}

	// Decode wav-formatted data and retrieve decoded PCM stream.
	d, err := wav.Decode(g.audioContext, f)
	if err != nil {
		log.Fatal(err)
	}

	d2, err := wav.Decode(g.audioContext, f2)
	if err != nil {
		log.Fatal(err)
	}

	// Create an audio.Player that has one stream.
	g.audioPlayer, err = audio.NewPlayer(g.audioContext, d)
	if err != nil {
		log.Fatal(err)
	}
	g.audioPlayer.SetVolume(0.1)

	g.audioPlayerCrunch, err = audio.NewPlayer(g.audioContext, d2)
	if err != nil {
		log.Fatal(err)
	}

	// Infinite loop background music ============================
	f3, err := os.Open("ragtime.ogg")
	if err != nil {
		return err
	}
	// oggS, err := vorbis.Decode(g.audioContext, audio.BytesReadSeekCloser(audio.Ragtime_ogg))
	oggS, err := vorbis.Decode(g.audioContext, f3)
	if err != nil {
		return err
	}

	// s := audio.NewInfiniteLoopWithIntro(oggS, introLengthInSecond*4*sampleRate, loopLengthInSecond*4*sampleRate)
	s := audio.NewInfiniteLoopWithIntro(oggS, 0, oggS.Size())

	g.audioPlayerBackground, err = audio.NewPlayer(g.audioContext, s)
	if err != nil {
		return err
	}
	g.audioPlayerBackground.SetVolume(0.1)
	// Infinite loop background music ============================

	return nil
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hungry-Snake")
	// Game config ==========================
	g := &Game{width: 10, height: 10, horizontal: 10, verticle: 10, AppleX: 30, AppleY: 30}
	g.speed = 5 // means 100 ms
	g.scalesLoc = make(map[int64]*Scale)
	g.scalesLoc[0] = &Scale{X: 0, Y: 0}
	g.turn2 = make(map[int64]*Turn)
	g.turnKey = 0
	g.soundEnable = true
	g.setupAudio()
	// ===============================

	// Snake image ========================
	g.snakeImage, _, _ = ebitenutil.NewImageFromFile("snake2.jpg", ebiten.FilterDefault)
	g.snakeSkin, _, _ = ebitenutil.NewImageFromFile("skin.png", ebiten.FilterLinear)
	g.snakeHead, _, _ = ebitenutil.NewImageFromFile("snakeHead.png", ebiten.FilterLinear)
	g.snakeHeadDown, _, _ = ebitenutil.NewImageFromFile("snakeHeadDown.png", ebiten.FilterLinear)
	g.snakeHeadUp, _, _ = ebitenutil.NewImageFromFile("snakeHeadUp.png", ebiten.FilterLinear)
	g.snakeHeadRight, _, _ = ebitenutil.NewImageFromFile("snakeHeadRight.png", ebiten.FilterLinear)
	g.snakeMouth, _, _ = ebitenutil.NewImageFromFile("snakeMouth.png", ebiten.FilterLinear)
	g.apple, _, _ = ebitenutil.NewImageFromFile("rabbit.png", ebiten.FilterLinear)
	// ===========================

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}
