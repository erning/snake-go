package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const GRID_WIDTH = 32
const GRID_HEIGHT = 24
const GRID_PIXEL = 10

const VIEW_WIDTH = GRID_PIXEL * GRID_WIDTH
const VIEW_HEIGHT = GRID_PIXEL * GRID_HEIGHT

const (
	DIR_UP    = 0
	DIR_DOWN  = 1
	DIR_LEFT  = 2
	DIR_RIGHT = 3
)

const (
	STATUS_PAUSED   = 0
	STATUS_RUNNING  = 1
	STATUS_GAMEOVER = 2
)

//

type Snake struct {
	head *SnakeBody
	tail *SnakeBody
	mdir int
}

type SnakeBody struct {
	x    int
	y    int
	prev *SnakeBody
}

//

var snake Snake
var movingDirection int
var eggX int
var eggY int
var score int
var highest int

var tickMS int
var lastTick int

var status int
var lastScore int
var lastTimer int

var redraw bool

//

func createSnake() {
	snake = Snake{}
	body := &SnakeBody{x: GRID_WIDTH / 2, y: GRID_HEIGHT / 2, prev: nil}
	snake.head = body
	snake.tail = body
	snake.mdir = 0
}

func growSnakeAt(x, y int) {
	newHead := &SnakeBody{x: x, y: y, prev: nil}
	oldHead := snake.head
	oldHead.prev = newHead
	snake.head = newHead
}

func appendSnakeTail() {
	snake.tail = &SnakeBody{
		x:    snake.tail.x,
		y:    snake.tail.y,
		prev: snake.tail,
	}
}

func moveSnakeTo(x, y int) {
	newHead := &SnakeBody{x: x, y: y, prev: nil}
	snake.head.prev = newHead
	snake.head = newHead
	snake.tail = snake.tail.prev
}

func moveSnake(dir int) (int, int) {
	x := snake.head.x
	y := snake.head.y
	if (dir == DIR_UP && snake.mdir != DIR_DOWN) ||
		(dir == DIR_DOWN && snake.mdir != DIR_UP) ||
		(dir == DIR_LEFT && snake.mdir != DIR_RIGHT) ||
		(dir == DIR_RIGHT && snake.mdir != DIR_LEFT) {
		snake.mdir = dir
	}
	switch snake.mdir {
	case DIR_UP:
		y--
	case DIR_DOWN:
		y++
	case DIR_LEFT:
		x--
	case DIR_RIGHT:
		x++
	}
	x = (x + GRID_WIDTH) % GRID_WIDTH
	y = (y + GRID_HEIGHT) % GRID_HEIGHT
	return x, y
}

func isSnakeBody(x, y int) bool {
	for body := snake.tail; body != nil; body = body.prev {
		if x == body.x && y == body.y {
			return true
		}
	}
	return false
}

//

func resetGame() {
	rand.Seed(time.Now().Unix())
	createSnake()
	score = 0
	randomEgg()
	status = STATUS_PAUSED

	tickMS = 150
	lastTick = int(time.Now().UnixNano() / 1000000)
	lastTimer = lastTick
	lastScore = score

	redraw = true
}

func fasterTicker(t int) {
	if tickMS < 10 {
		return
	}
	tickMS -= t
}

func randomEgg() {
	for {
		eggX = rand.Intn(GRID_WIDTH)
		eggY = rand.Intn(GRID_HEIGHT)
		if !isSnakeBody(eggX, eggY) {
			break
		}
	}
}

func drawSnake(image *ebiten.Image) {
	for body := snake.tail; body != nil; body = body.prev {
		x := body.x
		y := body.y
		im := ebiten.NewImage(GRID_PIXEL, GRID_PIXEL)
		im.Fill(color.White)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(GRID_PIXEL*x), float64(GRID_PIXEL*y))
		image.DrawImage(im, op)
	}
}

func drawEgg(image *ebiten.Image) {
	x := eggX
	y := eggY
	im := ebiten.NewImage(GRID_PIXEL, GRID_PIXEL)
	im.Fill(color.RGBA{0xf0, 0x00, 0x00, 0xff})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(GRID_PIXEL*x), float64(GRID_PIXEL*y))
	image.DrawImage(im, op)
}

func handleGameOver() {
	// press space key to restart
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		resetGame()
	}
}

func handlePaused() {
	// press any key to continue
	if len(inpututil.PressedKeys()) > 0 {
		status = STATUS_RUNNING
	}
}

func handleKeyboard() {
	keys := inpututil.PressedKeys()
	if len(keys) == 1 {
		switch keys[0] {
		case ebiten.KeyI, ebiten.KeyArrowUp, ebiten.KeyA:
			movingDirection = DIR_UP
		case ebiten.KeyK, ebiten.KeyArrowDown, ebiten.KeyZ:
			movingDirection = DIR_DOWN
		case ebiten.KeyJ, ebiten.KeyArrowLeft, ebiten.KeyBracketLeft:
			movingDirection = DIR_LEFT
		case ebiten.KeyL, ebiten.KeyArrowRight, ebiten.KeyBracketRight:
			movingDirection = DIR_RIGHT
		}
	}
}

func handleMove() {
	// move every tick ms
	t := int(time.Now().UnixNano() / 1000000)
	if t-lastTick > tickMS {
		lastTick = t
		x, y := moveSnake(movingDirection)
		if isSnakeBody(x, y) {
			status = STATUS_GAMEOVER
			return
		}
		if x == eggX && y == eggY {
			growSnakeAt(x, y)
			randomEgg()
			fasterTicker(1)
			score++
			if score > highest {
				highest = score
			}
			lastTimer = t
		} else {
			moveSnakeTo(x, y)
		}
		redraw = true
	}
}

func handleTimeout() {
	// move faster and increase snake length when timeout
	t := int(time.Now().UnixNano() / 1000000)
	if t-lastTimer > 30*1000 {
		lastTimer = t
		if score > lastScore {
			lastScore = score
		} else {
			fasterTicker(5)
			appendSnakeTail()
		}
	}
}

//

type Game struct {
}

func (g *Game) Update() error {
	switch status {
	case STATUS_PAUSED:
		handlePaused()
	case STATUS_RUNNING:
		handleKeyboard()
		handleMove()
		handleTimeout()
	case STATUS_GAMEOVER:
		handleGameOver()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !redraw {
		return
	}
	redraw = false
	screen.Clear()
	screen.Fill(color.RGBA{0xff, 0xff, 0xff, 0x40})
	drawEgg(screen)
	drawSnake(screen)

	msg := fmt.Sprintf("SCORE: %d, HIGHEST: %d", score, highest)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return VIEW_WIDTH, VIEW_HEIGHT
}

func main() {
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(VIEW_WIDTH*2, VIEW_HEIGHT*2)
	ebiten.SetWindowTitle("Snake")
	ebiten.SetWindowResizable(true)
	ebiten.SetScreenClearedEveryFrame(false)

	// Call ebiten.RunGame to start your game loop.
	resetGame()
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
