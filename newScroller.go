package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "embed"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/nfnt/resize"
	"github.com/nsf/termbox-go"
)

//go:embed "data/small8.64c"
var data []byte

//go:embed "data/tekscicho.txt"
var info string

//go:embed "data/music.mp3"
var music []byte

type ByteReadCloser struct {
	*bytes.Reader
}

func (b *ByteReadCloser) Close() error {
	return nil
}

// renderChar func input Ascii capital letter byte code and returns 8x8 font consist of 0 and 1 - 8 strings by 8x Zeros or Ones
func renderChar(b byte) *[]string {

	var items = []rune{
		'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S',
		'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '~', ']', '|', '\\', ' ', '!', '"', '#', '$', '%', '&',
		'\'', '(', ')', '*', '+', ',', '-', '.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		':', ';', '<', '=', '>', '?',
	}

	//this func maps code of letters with indices of pixels data regarding shape of certain semigraphics image of the letter
	translator := make(map[byte]int)
	for i, x := range items {
		translator[byte(x)] = i * 8

	}
	//charset data starts from the 3rd byte
	charset := data[2:]
	var rendered = make([]string, 0, 8) //create space for semigraphics image consist of Zeros and Ones!

	for y := 0; y < 8; y++ {

		t := translator[b]
		z := t + y
		x := charset[z]
		struna := fmt.Sprintf("%08b", x)
		rendered = append(rendered, struna)
	}
	return &rendered //return address of semigraphics "Big" image 8x8 cursors size.
}

// this func shifts left 0,1 make them 0,1,R or L font metacode for further exchanging by semigraphics
func createAlternativeString(message string) string {
	length := len(message)
	alternative := make([]byte, length)

	for i := 0; i < length-1; i++ {
		switch {
		case message[i] == '0' && message[i+1] == '0':
			alternative[i] = '0'
		case message[i] == '0' && message[i+1] == '1':
			alternative[i] = 'R'
		case message[i] == '1' && message[i+1] == '0':
			alternative[i] = 'L'
		case message[i] == '1' && message[i+1] == '1':
			alternative[i] = '1'
		}
	}

	return string(alternative)
}

// this func exchange 0,1,L,R by " "█"▐"▌" - what delivers posibility to scroll by half a cursor
func makeSemigraphic(subString string) string {
	subString = strings.ReplaceAll(subString, "0", ` `)
	subString = strings.ReplaceAll(subString, "1", `█`)
	subString = strings.ReplaceAll(subString, "R", `▐`)
	subString = strings.ReplaceAll(subString, "L", `▌`)
	return subString
}

type scroller struct {
	messageString       string              //text to scroll
	colorMessage        []termbox.Attribute //text colorization but of termbox.Attribute type
	messageStringMatrix [][]string          //two buffers for cursor fonts consist of 0,1 or in second buf 0,1,L,R
	messageBuferCells   [][][]termbox.Cell  //two buffers consist of termbox cells
	xMax                int
	yMax                int
	index               int
	interval            int
	lcx                 int
	lcy                 int
	rcx                 int
	frame               int
	speed               int
}

// This func is for loading text to be scrolled and in the meantime changes all letters to be uppercase - embeded data requires it
func (s *scroller) loadMessage(ss string) {

	s.messageString = strings.ToUpper(ss)

}
func (s *scroller) scrollerInit() { //Initialize scroll tables;colors; ascii table; shift half cursor semigraphics; create buffers4termbox
	//colorize ascii text

	//var randomNumber int
	var randR, randG, randB uint8
	s.colorMessage = make([]termbox.Attribute, len(s.messageString))

	//allocate [1][7] matrix dimension (2x8 strings)
	s.messageStringMatrix = append(s.messageStringMatrix, []string{"", "", "", "", "", "", "", ""})
	s.messageStringMatrix = append(s.messageStringMatrix, []string{"", "", "", "", "", "", "", ""})
	//oldRandom := 0
	for i, v := range s.messageString {
		if v == ' ' {

			//	onceAgain:
			randR = uint8(rand.Intn(200) + 55)
			randG = uint8(rand.Intn(250) + 5)
			randB = uint8(rand.Intn(250) + 5)
			/*randomNumber = rand.Intn(250) + 5
			if (randomNumber == oldRandom) || (randomNumber == 9) {
				goto onceAgain
			}
			oldRandom = randomNumber
			*/
		}
		s.colorMessage[i] = termbox.RGBToAttribute(randR, randG, randB)
		rendered8x8 := renderChar(byte(v))
		for y, str := range *rendered8x8 {
			s.messageStringMatrix[0][y] = s.messageStringMatrix[0][y] + str
		}

	}

	for i, str := range s.messageStringMatrix[0] {
		s.messageStringMatrix[1][i] = createAlternativeString(str)
	}
	//fmt.Printf("długość Matrix:%d\n", len(s.messageStringMatrix[0][0]))
	s.messageBuferCells = make([][][]termbox.Cell, 2)

	for i := 0; i < 2; i++ {
		s.messageBuferCells[i] = make([][]termbox.Cell, 8)
		//var ixi *int
		for j := 0; j < 8; j++ {
			for _, xx := range s.messageStringMatrix[i][j] {

				var yy termbox.Cell
				vi := int(xx)

				yy.Ch = changeCharacter(vi)
				s.messageBuferCells[i][j] = append(s.messageBuferCells[i][j], yy)
				//		ixi = &ix
			}

		}
		//fmt.Printf("xx:% 3d\n", *ixi)

	}
	//colorization
	kolorIndex := 0
	for k := 0; k < len(s.messageBuferCells[0][0]); k = k + 8 {
		kolor := s.colorMessage[kolorIndex]
		for repeat := 0; repeat < 8; repeat++ {
			s.messageBuferCells[0][0][k+repeat].Fg = kolor
			s.messageBuferCells[0][1][k+repeat].Fg = kolor
			s.messageBuferCells[0][2][k+repeat].Fg = kolor
			s.messageBuferCells[0][3][k+repeat].Fg = kolor
			s.messageBuferCells[0][4][k+repeat].Fg = kolor
			s.messageBuferCells[0][5][k+repeat].Fg = kolor
			s.messageBuferCells[0][6][k+repeat].Fg = kolor
			s.messageBuferCells[0][7][k+repeat].Fg = kolor

		}
		kolorIndex++
		if kolorIndex >= len(s.colorMessage) {
			kolorIndex = kolorIndex - len(s.colorMessage)
		}

	}

	for r := 0; r <= 7; r++ {
		for i, object := range s.messageBuferCells[0][r][1:] {
			s.messageBuferCells[1][r][i].Fg = object.Fg

		}
	}
	s.xMax, s.yMax = termbox.Size()
	s.index = 0
}

func onExit() {
	exec.Command("/usr/bin/bash")
}

func changeCharacter(i int) rune {

	switch i {
	case 0:
		return ' '
	case 49:
		return '█'
	case 82:
		return '▐'
	case 76:
		return '▌'
	default:
		return ' '
	}

}

func printAt(x, y, c int, format string, args ...interface{}) error {

	//termbox.SetCursor(x, y)
	back := termbox.GetCell(x, y)
	termbox.SetCell(x, y, ' ', termbox.Attribute(c), back.Bg)

	// Format and print the string
	str := fmt.Sprintf(format, args...)
	for i, char := range str {
		//back:=termbox.GetCell(x)
		back = termbox.GetCell(x+i+1, y)
		termbox.SetCell(x+i+1, y, char, termbox.Attribute(c), back.Bg)
	}

	//termbox.Flush()
	return nil
}
func (s *scroller) scrolling() {
	//var frame int

	indexing := s.index
	frame := s.frame
	lenXbuffer := len(s.messageBuferCells[1][0])
	/*	termbox.Close()
		fmt.Println("len", lenXbuffer, "indexing", indexing, "s.frame", s.frame)
		os.Exit(128)
	*/
	//lenXbuffer = lenXbuffer

	for x := s.lcx; x <= s.rcx; x++ {
		znak := s.messageBuferCells[frame][0][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 0+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}
		znak = s.messageBuferCells[frame][1][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 1+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}
		znak = s.messageBuferCells[frame][2][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 2+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}
		znak = s.messageBuferCells[frame][3][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 3+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}
		znak = s.messageBuferCells[frame][4][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 4+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}
		znak = s.messageBuferCells[frame][5][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 5+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}
		znak = s.messageBuferCells[frame][6][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 6+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}
		znak = s.messageBuferCells[frame][7][indexing]
		if znak.Ch != ' ' {
			termbox.SetCell(x, 7+s.lcy, znak.Ch, znak.Fg, termbox.ColorDefault)
		}

		indexing++
		if indexing >= lenXbuffer {
			indexing = indexing - lenXbuffer
		}
	}

	/*switch frame {

	case 0:
		frame = 1
	case 1:
		frame = 0
	}*/
	//s.index++
	s.frame = s.frame + s.speed

looper:
	if s.frame >= 2 {
		s.frame = s.frame - 2
		s.index++
		if s.index >= len(s.messageBuferCells[0][0]) {
			s.index = s.index - len(s.messageBuferCells[0][0])
		}
	}
	if s.frame >= 2 {
		goto looper
	}
	/*
		for yyy := 0; yyy <= s.yMax; yyy++ {
			for xxx := 0; xxx <= s.xMax; xxx++ {
				termbox.SetBg(xxx, yyy, 0)
			}
		}*/

}

func findAllOccurrences(input string, pattern string) []int {
	var occurrences []int
	startIndex := 0

	for {
		index := strings.Index(input[startIndex:], pattern)
		if index == -1 {
			break
		}

		// Adjust the index based on the startIndex
		index += startIndex

		occurrences = append(occurrences, index)

		// Update the startIndex for the next search
		startIndex = index + 1
	}

	return occurrences
}

func main() {

	go func() {
		for { /*
				f, err := os.Open("moon.mp3")
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
			*/
			mp3Reader := &ByteReadCloser{
				bytes.NewReader(music),
			}
			streamer, format, err := mp3.Decode(mp3Reader)
			if err != nil {
				log.Fatal(err)
			}
			defer streamer.Close()
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*100))
			done := make(chan struct{})
			speaker.Play(beep.Seq(streamer, beep.Callback(func() {
				close(done)
			})))
			<-done
		}
	}()

	//stopAllGorutines := false
	//stopAllGorutines = stopAllGorutines
	credits := "Original C64 Tune Composed by Jeroen Tel, Recorded and Rearanged by Matt Gray."
	credits3 := "Programmed by Andrzej Gajowniczek, 05-400 Otwock Poland"
	credits2 := "Contact: https://www.linkedin.com/in/andrzej-gajowniczek-5a6564b/"

	err := termbox.Init()
	if err != nil {
		log.Fatal("init() error", err)
	}
	termbox.SetOutputMode(termbox.OutputRGB)

	filePath := "data/zameczek.png"
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Cannot read file:", err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		log.Println("Cannot decode file:", err)
	}
	size := img.Bounds().Size()
	fmt.Printf("Ximg: %d x Yimg: %d \n", size.X, size.Y)

	xM, yM := termbox.Size()
	newImage := resize.Resize(uint(xM), uint(yM), img, resize.Bilinear)
	size = newImage.Bounds().Size()

	var sc scroller
	var sc2, sc3 scroller

	info2 := "                           " +
		"10 9 8 7 6 5 4 3 2 1 * :) ;) XD " +
		"This is termbox-go console demo program coded by Andy! Greetings to my wonderful colleagues, coworkers, supervisors, and all the amazing people I'm proud to know personally throughout the years, including:" +
		"Lukasz Sobczak, Mariusz Kaczorek, Pawel Osobinski, Michal Kossowski, Bartek Chodnicki, Krzysztof Lubanski, Krzysztof Jurkiewicz, Piotr Malendo, Jakub Olszewski, Jakub Boguslaw, Aman Kumar, David Matinyarare, David Smidke, Robert Petricek, Anna Makos, Milan Misovich, Pawel Sedrowski, Radoslaw Polubok, Anna Szade, Pawel Baran, Maciej Wroblewski, Katarzyna Siedlarek, Damian Galka, Tomasz Dratkiewicz, Olaf Siejka, Ewa Gorecka, Paulina Lapinska, Grzegorz Pawlik, Renata Zalecka, Arkadiusz Knopik, Wojciech Rozanski, Rafal Liberacki, Mateusz Baran, Magdalena Pabisiak, Marek Matys, Andrzej Ziebakowski, Lukasz Malinowski, Adam Wasowski, Anna Mikula, Lukasz Mazur, Pawel Kozlowski, Jakub Husak, Krzysztof Mazur, Robert Witan, Maciej Woznica, Piotr Michalczyk, Dariusz Przybysz, Lukasz Gawronski, Zbigniew Malec, Tomasz Kilinski, Ryszard Czarnecki, Andrzej Lisowski, John Waterhouse, Rafal Antas, Tadeusz Niebieszczanski, Krzysztof, Stolarek, Przemyslaw Niton, Kamil Krawitowski, Tomek Ziss,Marcin Zajac, Michal Prusek, Piotr Karpuk, Renata Zalecka, Radoslaw Bartosinski, Gosia Heba, Bartlomiej Kwiatkowski" +
		" Go with Andy :) Cheers !   Bye !  "
	sc.loadMessage(info)
	sc2.loadMessage(info2)
	sc3.loadMessage(info2)

	sc.scrollerInit()
	sc2.scrollerInit()
	sc3.scrollerInit()

	termbox.Flush()

	defer termbox.Close()

	sc.lcx = 0 //left corner x coordinate
	sc2.lcx = 0
	sc3.lcx = 0

	//func ifResolutonChanges(){}
	sc2.lcy = (sc.yMax-8)/2 - 7 //left corner y coordinate
	sc.lcy = (sc.yMax - 8) / 2
	sc3.lcy = (sc.yMax-8)/2 + 8

	sc.rcx = sc.xMax //right corner x max coordinate
	sc2.rcx = sc.xMax
	sc3.rcx = sc.xMax

	sc.frame = 0
	sc2.frame = 0
	sc3.frame = 0
	sc.interval = 16 + 17
	//	sc2.interval = 30
	//	sc3.interval = 30
	sc.speed = 1
	sc2.speed = 2
	sc3.speed = 2

	indicesof := findAllOccurrences(info2, "Andy")
	for _, index := range indicesof {

		positionStart := index * 8
		positionEnd := positionStart + 8*len("Andy")
		type rGb struct {
			r uint8
			g uint8
			b uint8
		}
		gradient := []rGb{{50, 50, 255}, {100, 0, 150}, {100, 0, 200}, {200, 0, 200}, {200, 0, 50}, {250, 0, 20}, {255, 200, 0}, {255, 200, 0}}
		for background := positionStart; background <= positionEnd; background++ {
			for y := 0; y <= 7; y++ {
				sc2.messageBuferCells[0][y][background].Fg = termbox.RGBToAttribute(gradient[y].r, gradient[y].g, gradient[y].b)
				//sc.messageBuferCells[1][y][background-1].Fg = termbox.RGBToAttribute(gradient[y].r, gradient[y].g, gradient[y].b)
				sc3.messageBuferCells[0][y][background].Fg = termbox.RGBToAttribute(gradient[y].r, gradient[y].g, gradient[y].b)
			}
		}

	}
	//	...
	eventCh := make(chan termbox.Event)
	errCh := make(chan error)
	go func() {
		for {
			event := termbox.PollEvent()
			eventCh <- event
			/*if stopAllGorutines == true {

				break
			}*/
		}
	}()
	// Render all scrollers in one goroutine because flush cannot be executed in separate routines and parallization doesn't improve the performance
	//direction := 1

	go func() {
		for {
			select {
			case event := <-eventCh:
				// Handle the event
				switch event.Type {
				case termbox.EventKey:
					// Handle key press event
					if event.Key == termbox.KeyEsc {
						// Exit the loop if the Escape key is pressed
						termbox.Close()
						os.Exit(0)
					}
					// Handle other key events as needed
				case termbox.EventMouse:
					// Handle mouse event
					// ...
				case termbox.EventResize:

					//stopAllGorutines = true
					break
				}
				// Perform other operations based on the event if any ...
			case err := <-errCh:
				// Handle errors, if any
				//stopAllGorutines = true
				termbox.Close()
				fmt.Println(err)
				os.Exit(127)
			default:
				// Perform other non-blocking operations
				// ...
			}
		}
	}()

	//go func() {
	for {
		start := time.Now()
		sc.scrolling()
		sc2.scrolling()
		sc3.scrolling()
		size = newImage.Bounds().Size()
		for i := 0; i < size.X; i++ {
			//var y []color.Color
			for j := 0; j < size.Y; j++ {
				//	y = append(y, newImage.At(i, j))
				r, g, b, _ := newImage.At(i, j).RGBA()
				//color := termbox.RGBToAttribute(uint8(r>>8), uint8(g>>16)
				r8 := uint8(r >> 8)
				g8 := uint8(g >> 8)
				b8 := uint8(b >> 8)
				termbox.SetBg(i, j, termbox.RGBToAttribute(r8, g8, b8))
			}
		}
		//chessy.drawChessBoard(0, 0, sc.xMax, sc.yMax)
		//termbox.Sync()
		printAt(0, yM-3, int(termbox.RGBToAttribute(100, 255, 200)), credits)
		printAt(0, yM-2, int(termbox.RGBToAttribute(0, 255, 227)), credits2)
		printAt(0, yM-1, int(termbox.RGBToAttribute(0, 255, 255)), credits3)
		termbox.Flush()
		termbox.Clear(0, 0)

		duration := time.Since(start)
		spent := duration.Microseconds()
		//spent = spent
		if (33333 - spent) > 33 {
			wait := time.Duration(33333-spent) * time.Microsecond
			time.Sleep(wait)
		}
		//time.Sleep(time.Millisecond * time.Duration(sc.interval))
	}
	//}()

	//the below code is for exit from this program by key press [ESC] only

}

/*
func decodeAACFile(filename string) (*audio.IntBuffer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := wav.NewDecoder(file)
	if !d.IsValidFile() || d.Format().SampleRate != 44100 || d.Format().SampleSize != 16 {
		return nil, fmt.Errorf("unsupported WAV format")
	}

	buf, err := d.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	return buf, nil
}
func playAudio(buf *audio.IntBuffer) error {
	player, err := oto.NewPlayer(buf.Format.SampleRate, buf.Format.Channels, 2, 8192)
	if err != nil {
		return err
	}
	defer player.Close()

	_, err = player.Write(buf.Data)
	if err != nil {
		return err
	}

	player.Close()

	return nil
}
*/
