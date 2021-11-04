package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

type CursorStruct struct {
	x, y, wantX int
}

type TopLeftStruct struct {
	x, y int
}

type ScreenStruct struct {
	tcell.Screen
	defaultStyle tcell.Style
	infoStyle    tcell.Style
}

type LineStruct struct {
	line string
	len  int
}

type LineSlice []LineStruct

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (line *LineStruct) setFromString(strings ...string) {
	line.line = ""
	for _, s := range strings {
		line.line = line.line + s
	}
	line.len = utf8.RuneCountInString(line.line)
}

func (scr ScreenStruct) renderLine(xoffset, screenLine int, line string) {
	maxx, maxy := scr.Size()
	y := screenLine
	if y < 0 {
		y = 0
	}
	if y >= maxy {
		y = maxy - 1
	}
	x := 0
	for _, r := range line {
		x = x - xoffset
		if x < 0 {
			continue
		}
		if x >= maxx {
			break
		}
		var comb []rune
		w := runewidth.RuneWidth(r)
		if w == 0 {
			comb = []rune{r}
			r = ' '
			w = 1
		}
		scr.SetContent(x, y, r, comb, scr.defaultStyle)
		x += w
	}
}

func (scr ScreenStruct) renderInfoLine(line string) {
	/*
		maxx, _ := scr.Size()
		x := maxx - utf8.RuneCountInString(line)
		ry := 0
		for lx, r := range line {
			rx := x + lx
			if rx >= maxx {
				break
			}
			scr.SetContent(rx, ry, r, nil, scr.infoStyle)
		}
	*/
	maxx, _ := scr.Size()
	x := maxx - utf8.RuneCountInString(line) - 1
	for _, c := range line {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		scr.SetContent(x, 0, c, comb, scr.infoStyle)
		x += w
	}
}

func main() {
	// handle parameters
	/*
		flag.Parse()
		if flag.NArg() != 1 {
			fmt.Printf("Missing parameter <filename>\n")
			fmt.Printf("use: edit <filename>\n")
			os.Exit(1)
		}
		args := flag.Args()
	*/
	args := []string{"test.txt"}
	// init doc object
	doc := DocStruct{
		filename:      args[0],
		text:          []LineStruct{},
		screen:        ScreenStruct{},
		absolutCursor: CursorStruct{x: 0, y: 0, wantX: 0},
		viewport:      TopLeftStruct{x: 0, y: 0},
		undoStack: UndoStackStruct{
			undoSlice: []UndoItemStruct{},
			top:       0,
		},
	}

	// Initialize tcell
	encoding.Register()
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Println("Not supported terminal")
		log.Fatalf("%+v\n", err)
	}
	doc.screen = ScreenStruct{
		Screen:       screen,
		defaultStyle: tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset),
		infoStyle:    tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
	}
	err = doc.screen.Init()
	if err != nil {
		fmt.Println("Not supported terminal")
		log.Fatalf("%+v\n", err)
	}

	// load document
	err = doc.load()
	if err == nil || errors.Is(err, os.ErrNotExist) {
		doc.text = append(doc.text, LineStruct{})
	} else {
		log.Fatalf("%+v\n", err)
	}

	// init screen
	doc.screen.SetStyle(doc.screen.defaultStyle)
	doc.renderScreen()
	doc.showCursor()

	// Event loop
	for {
		// Update screen
		doc.screen.Show()

		// handle event
		event := doc.screen.PollEvent()
		switch event := event.(type) {
		case *tcell.EventResize:
			doc.renderScreen()
			doc.screen.Sync()
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
				// exit
				doc.screen.Fini()
				os.Exit(0)
			} else if event.Key() == tcell.KeyCtrlL {
				// sync
				doc.screen.Sync()
			} else {
				// handle key events
				doc.handleEvent(event)
			}
			doc.showCursor()
		}
	}
}
