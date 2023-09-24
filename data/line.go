package main

import (
	"github.com/nsf/termbox-go"
)

const (
	width  = 80
	height = 24
)

var screenChars [][]rune

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	drawLine(10, 10, 30, 45)
	putPoint(40, 12)

	znak := getCharAtPosition(10, 10)
	info := "to jest znak z pozucji 10,10:" + string(znak)
	for i, v := range info {
		termbox.SetCell(0+i, 0, v, 3, 1)

	}

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return
			}
		}
	}
}

func drawLine(x1, y1, x2, y2 int) {
	dx := x2 - x1
	dy := y2 - y1

	x := x1
	y := y1

	sx := 1
	if dx < 0 {
		sx = -1
		dx = -dx
	}

	sy := 1
	if dy < 0 {
		sy = -1
		dy = -dy
	}

	err := dx - dy

	for {
		termbox.SetCell(x, y, '█', termbox.ColorDefault, termbox.ColorDefault)

		if x == x2 && y == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}

	termbox.Flush()
}

func putPoint(x, y int) {
	termbox.SetCell(x, y, '█', termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
}

func getCursorPoint() (x, y int) {
	return termbox.Size()
}

func getCharAtPosition(x, y int) rune {
	if x >= 0 && x < width && y >= 0 && y < height {
		return screenChars[x][y]
	}
	return ' '
}
