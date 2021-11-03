package main

import (
	"bufio"
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

type DocStruct struct {
	filename      string
	text          LineSlice
	screen        ScreenStruct
	absolutCursor CursorStruct
	viewport      TopLeftStruct
}

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

func (line *LineStruct) concatenate(lines ...LineStruct) {
	for _, l := range lines {
		line.line = line.line + l.line
		line.len = line.len + l.len
	}
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

func (doc *DocStruct) load() error {
	// open file and read line by line
	f, err := os.Open(doc.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	doc.text = make([]LineStruct, 0, 256)
	scanner := bufio.NewScanner(f) // default delimiter is new line
	line := LineStruct{}
	for scanner.Scan() {
		line.setFromString(scanner.Text())
		doc.text = append(doc.text, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (doc *DocStruct) save() error {
	f, err := os.Create("data.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	for _, line := range doc.text {
		_, err = f.WriteString(line.line)
		if err != nil {
			return err
		}
	}

	return nil
}

func (doc *DocStruct) alignCursorX() {
	doc.absolutCursor.x = doc.absolutCursor.wantX
	len := doc.text[doc.absolutCursor.y].len
	if doc.absolutCursor.x > len {
		doc.absolutCursor.x = len
	}
	if doc.absolutCursor.x < 0 {
		doc.absolutCursor.x = 0
	}
}

func (doc *DocStruct) adjustViewport() {
	screenMaxX, screenMaxY := doc.screen.Size()
	if doc.absolutCursor.y-doc.viewport.y >= (screenMaxY - 1) {
		doc.viewport.y = doc.absolutCursor.y - (screenMaxY - 1)
		doc.renderScreen()
	}
	if doc.absolutCursor.y-doc.viewport.y < 0 {
		doc.viewport.y = doc.absolutCursor.y
		doc.renderScreen()
	}
	if doc.absolutCursor.x-doc.viewport.x >= (screenMaxX - 1) {
		doc.viewport.x = doc.absolutCursor.x - (screenMaxX - 1)
		doc.renderScreen()
	}
	if doc.absolutCursor.x-doc.viewport.x < 0 {
		doc.viewport.x = doc.absolutCursor.x
		doc.renderScreen()
	}
}

func (doc *DocStruct) showCursor() {
	doc.screen.ShowCursor(
		doc.absolutCursor.x-doc.viewport.x,
		doc.absolutCursor.y-doc.viewport.y,
	)
}

func (doc *DocStruct) renderScreen() {
	doc.screen.Clear()
	_, maxy := doc.screen.Size()
	for y := 0; y < maxy; y++ {
		if len(doc.text) <= doc.viewport.y+y {
			break
		}
		doc.screen.renderLine(doc.viewport.x, y, doc.text[doc.viewport.y+y].line)
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
		filename: args[0],
		text:     []LineStruct{},
		screen:   ScreenStruct{},
		absolutCursor: CursorStruct{
			x:     0,
			y:     0,
			wantX: 0,
		},
		viewport: TopLeftStruct{
			x: 0,
			y: 0,
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
			} else if event.Key() == tcell.KeyDown {
				// cursor down
				doc.absolutCursor.y++
				if doc.absolutCursor.y >= len(doc.text) {
					doc.absolutCursor.y = len(doc.text) - 1
				}
				doc.alignCursorX()
				doc.adjustViewport()
			} else if event.Key() == tcell.KeyUp {
				// cursor up
				doc.absolutCursor.y--
				if doc.absolutCursor.y < 0 {
					doc.absolutCursor.y = 0
				}
				doc.alignCursorX()
				doc.adjustViewport()
			} else if event.Key() == tcell.KeyRight {
				// cursor right
				l := doc.text[doc.absolutCursor.y].len
				if doc.absolutCursor.x < l {
					doc.absolutCursor.x++
				} else {
					// cursor right when curson is on last postion of line
					if doc.absolutCursor.y < len(doc.text)-1 {
						doc.absolutCursor.y++
						doc.absolutCursor.x = 0
					}
				}
				doc.absolutCursor.wantX = doc.absolutCursor.x
				doc.alignCursorX()
				doc.adjustViewport()
			} else if event.Key() == tcell.KeyLeft {
				// cursor left
				if doc.absolutCursor.x > 0 {
					doc.absolutCursor.x--
				} else {
					// cursor left when cursor is on first position of line
					if doc.absolutCursor.y > 0 {
						doc.absolutCursor.y--
						doc.absolutCursor.x = doc.text[doc.absolutCursor.y].len
					}
				}
				doc.absolutCursor.wantX = doc.absolutCursor.x
				doc.alignCursorX()
				doc.adjustViewport()
			} else if event.Key() == 268 {
				// go to begin of line (pos1)
				doc.absolutCursor.x = 0
				doc.absolutCursor.wantX = doc.absolutCursor.x
				doc.alignCursorX()
				doc.adjustViewport()
			} else if event.Key() == 269 {
				// go to end of line (end)
				doc.absolutCursor.x = doc.text[doc.absolutCursor.y].len
				doc.absolutCursor.wantX = doc.absolutCursor.x
				doc.adjustViewport()
			} else if event.Key() == tcell.KeyPgDn {
				// page down
				_, maxy := doc.screen.Size()
				if maxy > 1 {
					maxy--
				}
				doc.absolutCursor.y = doc.absolutCursor.y + maxy
				if doc.absolutCursor.y >= len(doc.text) {
					doc.absolutCursor.y = len(doc.text) - 1
				}
				doc.alignCursorX()
				doc.adjustViewport()
			} else if event.Key() == tcell.KeyPgUp {
				// page up
				_, maxy := doc.screen.Size()
				if maxy > 1 {
					maxy--
				}
				doc.absolutCursor.y = doc.absolutCursor.y - maxy
				if doc.absolutCursor.y < 0 {
					doc.absolutCursor.y = 0
				}
				doc.alignCursorX()
				doc.adjustViewport()
			} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
				// backspace
				x := doc.absolutCursor.x
				y := doc.absolutCursor.y
				if x <= 0 {
					// backspace when cursor is on first position of line
					if y > 0 {
						doc.absolutCursor.x = doc.text[y-1].len
						doc.absolutCursor.wantX = doc.absolutCursor.x
						doc.text[y-1].concatenate(doc.text[y])
						doc.text = append(doc.text[:y], doc.text[y+1:]...)
						doc.absolutCursor.y--
						doc.adjustViewport()
						doc.renderScreen()
					}
				} else {
					runes := []rune(doc.text[y].line)
					beforeCursor := string(runes[:x-1])
					afterCursor := string(runes[x:])
					doc.text[y].setFromString(beforeCursor, afterCursor)
					doc.screen.renderLine(doc.viewport.x, y-doc.viewport.y, doc.text[y].line+" ")
					doc.absolutCursor.x--
					doc.absolutCursor.wantX = doc.absolutCursor.x
					doc.alignCursorX()
				}
			} else if event.Key() == 271 {
				// delete
				x := doc.absolutCursor.x
				y := doc.absolutCursor.y
				l := len(doc.text)
				if x == doc.text[y].len {
					// pressing delete when cursor is at end of line
					if y+1 < l {
						doc.text[y].concatenate(doc.text[y+1])
						if y+2 < l {
							doc.text = append(doc.text[:y+1], doc.text[y+2:]...)
						} else {
							doc.text = doc.text[:y+1]
						}
						doc.adjustViewport()
						doc.renderScreen()
					}
				} else {
					// pressing delete somewhere in the line
					runes := []rune(doc.text[y].line)
					beforeCursor := string(runes[:x])
					afterCursor := string(runes[x+1:])
					doc.text[y].setFromString(beforeCursor, afterCursor)
					doc.screen.renderLine(doc.viewport.x, y-doc.viewport.y, doc.text[y].line+" ")
					doc.alignCursorX()
				}
			} else if event.Key() == tcell.KeyEnter {
				// enter
				x := doc.absolutCursor.x
				y := doc.absolutCursor.y
				if len(doc.text) == y+1 { // nil or empty slice or after last element
					doc.text = append(doc.text, LineStruct{})
				}
				doc.text = append(doc.text[:y+2], doc.text[y+1:]...) // index < len(text)
				if x == doc.text[y].len {
					// pressing enter when cursor is at end of line
					doc.text[y+1] = LineStruct{}
				} else {
					// pressing enter somewhere in the line
					runes := []rune(doc.text[y].line)
					beforeCursor := string(runes[:x])
					afterCursor := string(runes[x:])
					doc.text[y].setFromString(beforeCursor)
					doc.text[y+1].setFromString(afterCursor)
				}
				doc.absolutCursor.x = 0
				doc.absolutCursor.wantX = 0
				doc.absolutCursor.y++
				doc.adjustViewport()
				doc.renderScreen()
			} else if event.Key() == tcell.KeyCtrlS {
				// Ctrl+S save
				os.Rename(doc.filename, doc.filename+".bak")
				doc.save()
			} else if event.Key() == tcell.KeyRune {
				// insert character (rune)
				x := doc.absolutCursor.x
				y := doc.absolutCursor.y
				runes := []rune(doc.text[y].line)
				beforeCursor := string(runes[:x])
				afterCursor := string(runes[x:])
				doc.text[y].setFromString(beforeCursor, string(event.Rune()), afterCursor)
				doc.screen.renderLine(doc.viewport.x, y-doc.viewport.y, doc.text[y].line)
				doc.absolutCursor.x++
				doc.absolutCursor.wantX = doc.absolutCursor.x
				doc.adjustViewport()
			}

			doc.screen.renderInfoLine(
				fmt.Sprintf("Lines: %d Curs(x:%.2d y:%.2d)  Viewp(x:%d y:%d) #%04d  #%x ",
					len(doc.text),
					doc.absolutCursor.x,
					doc.absolutCursor.y,
					doc.viewport.x,
					doc.viewport.y,
					event.Key(),
					event.Modifiers(),
				))
		}
		doc.showCursor()
	}
}
