package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

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
		text:          []LineType{},
		screen:        ScreenStruct{},
		absolutCursor: CursorStruct{x: 0, y: 0, wantX: 0},
		viewport:      xyStruct{x: 0, y: 0},
		undoStack: UndoStackStruct{
			undoSlice: []UndoItemStruct{},
			top:       0,
		},
		selection: selectionStruct{
			begin: xyStruct{
				x: 0,
				y: 0,
			},
			end: xyStruct{
				x: 10,
				y: 0,
			},
		},
	}

	// Initialize tcell
	encoding.Register()
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Println("error creating screen")
		log.Fatalf("%+v\n", err)
	}
	doc.screen = ScreenStruct{
		Screen:       screen,
		defaultStyle: tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset),
		infoStyle:    tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorRed),
	}
	doc.screen.selectionStyle = doc.screen.defaultStyle.Reverse(true)

	err = doc.screen.Init()
	if err != nil {
		fmt.Println("error initializing screen")
		log.Fatalf("%+v\n", err)
	}

	// load document
	err = doc.handleEventLoad()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			doc.text = append(doc.text, LineType{})
		} else {
			log.Fatalf("%+v\n", err)
		}
	}

	// init screen
	doc.screen.SetStyle(doc.screen.defaultStyle)
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
			// doc.screen.renderLine(1, 0, "ABC")
			doc.screen.Sync()
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
				// exit
				doc.screen.Fini()
				os.Exit(0)
			} else if event.Key() == tcell.KeyCtrlL {
				// sync
				doc.screen.Sync()
			} else if event.Key() == tcell.KeyCtrlA {
				doc.renderScreen()
				doc.screen.Beep()
			} else {
				// handle key events
				doc.handleKeyEvent(event)
			}
			doc.showCursor()
		}
	}
}
