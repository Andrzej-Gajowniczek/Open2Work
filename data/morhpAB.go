package main

import (
	"github.com/nsf/termbox-go"
	"math/rand"
	"time"
)

const (
	width  = 80
	height = 24
)

var (
	lettersA = []rune(`
█████
█   █
█████
█   █
█   █
`)

	lettersB = []rune(`
█████
█   █
█████
█   █
█████
`)

	currentLetter []rune
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	rand.Seed(time.Now().UnixNano())

	currentLetter = lettersA

	drawLetter()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

loop:
	for {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeySpace {
				morphToLetter(lettersB)
			} else if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
				break loop
			}

		default:
			// No events, continue animating
			animate()
			drawLetter()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func morphToLetter(targetLetter []rune) {
	if len(currentLetter) != len(targetLetter) {
		panic("Letters have different lengths")
	}

	for i := 0; i < len(currentLetter); i++ {
		if currentLetter[i] != targetLetter[i] {
			currentLetter[i] = targetLetter[i]
		}
	}
}

func animate() {
	for i := 0; i < len(currentLetter); i++ {
		if currentLetter[i] == ' ' {
			continue
		}

		randomOffset := rand.Intn(3) - 1
		newChar := rune(int(currentLetter[i]) + randomOffset)

		if newChar < 32 || newChar > 126 {
			newChar = currentLetter[i]
		}

		currentLetter[i] = newChar
	}
}

func drawLetter() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	x := (width - 8) / 2
	y := (height - 5) / 2

	for i, char := range currentLetter {
		if char == ' ' {
			continue
		}

		color := termbox.Attribute(i%256 + 1)

		termbox.SetCell(x+i%8, y+i/8, char, termbox.ColorDefault, color)
	}

	termbox.Flush()
}

