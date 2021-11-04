package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
)

type ActionStruct struct {
	row    int
	delete bool
	insert bool
	update bool
	line   LineStruct
}

type UndoItemStruct struct {
	actionSlice []ActionStruct
}

func (ui *UndoItemStruct) appendAction(action ActionStruct) {
	ui.actionSlice = append(ui.actionSlice, action)
}

func newUndoItem() UndoItemStruct {
	ui := UndoItemStruct{
		actionSlice: []ActionStruct{},
	}
	return ui
}

type UndoStackStruct struct {
	undoSlice []UndoItemStruct
	top       int
}

func (us *UndoStackStruct) push(ui UndoItemStruct) {
	us.undoSlice = append(us.undoSlice, ui)
	us.top++
}

func (us *UndoStackStruct) pop() UndoItemStruct {
	us.top--
	return us.undoSlice[us.top]
}

type DocStruct struct {
	filename      string
	text          LineSlice
	screen        ScreenStruct
	absolutCursor CursorStruct
	viewport      TopLeftStruct
	undoStack     UndoStackStruct
}

func (doc *DocStruct) updateLine(ui *UndoItemStruct, row int, line string) {
	if ui != nil {
		// create actionItem for Undo
		action := ActionStruct{
			row:    row,
			delete: false,
			insert: false,
			update: true,
			line:   doc.text[row],
		}
		ui.appendAction(action)
	}
	doc.text[row].setFromString(line)
}

func (doc *DocStruct) insertLine(ui *UndoItemStruct, row int, line string) {
	if row+1 > len(doc.text) {
		return
	}
	// insert line
	if len(doc.text) == row+1 {
		doc.text = append(doc.text, LineStruct{})
	} else {
		doc.text = append(doc.text[:row+1], doc.text[row:]...)
	}
	if ui != nil {
		// create actionItem for Undo
		action := ActionStruct{
			row:    row,
			delete: false,
			insert: true,
			update: true,
			line:   doc.text[row],
		}
		ui.appendAction(action)
	}
	// update line
	doc.updateLine(ui, row, line)
}

func (doc *DocStruct) deleteLine(ui *UndoItemStruct, row int) {
	if row >= len(doc.text) {
		return
	}
	if ui != nil {
		// create actionItem for Undo
		action := ActionStruct{
			row:    row,
			delete: false,
			insert: true,
			update: true,
			line:   doc.text[row],
		}
		ui.appendAction(action)
	}
	if row+1 < len(doc.text) {
		doc.text = append(doc.text[:row], doc.text[row+1:]...)
	} else {
		doc.text = doc.text[:row]
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

func (doc *DocStruct) undo() {
	ui := doc.undoStack.pop()

	for i := len(ui.actionSlice) - 1; i > 0; i-- {
		action := ui.actionSlice[i]
		if action.update {
			doc.text[action.row] = action.line
			doc.renderLine(action.row)
		}
	}
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

func (doc *DocStruct) renderLine(row int) {
	doc.screen.renderLine(doc.viewport.x, row, doc.text[doc.viewport.y+row].line+" ")
}

func (doc *DocStruct) renderScreen() {
	doc.screen.Clear()
	_, maxy := doc.screen.Size()
	for y := 0; y < maxy; y++ {
		if len(doc.text) <= doc.viewport.y+y {
			break
		}
		doc.renderLine(y)
	}
}

func (doc *DocStruct) handleEventCursor(event *tcell.EventKey) (handled bool) {
	handled = true
	if event.Key() == tcell.KeyDown {
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
	} else {
		handled = false
	}
	return
}

func (doc *DocStruct) handleEvent(event *tcell.EventKey) {
	if doc.handleEventCursor(event) {
		return
	}
	if event.Key() == tcell.KeyTab {
		// tab
		undoItem := newUndoItem()
		const fakeTab string = "    "
		x := doc.absolutCursor.x
		y := doc.absolutCursor.y
		runes := []rune(doc.text[y].line)
		beforeCursor := string(runes[:x])
		afterCursor := string(runes[x:])
		doc.updateLine(&undoItem, y, beforeCursor+fakeTab+afterCursor)
		doc.undoStack.push(undoItem)

		doc.renderLine(y)
		doc.absolutCursor.x += len(fakeTab)
		doc.absolutCursor.wantX = doc.absolutCursor.x
		doc.adjustViewport()
	} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
		// backspace
		undoItem := newUndoItem()
		x := doc.absolutCursor.x
		y := doc.absolutCursor.y
		if x <= 0 {
			// backspace when cursor is on first position of line
			if y > 0 {
				doc.absolutCursor.x = doc.text[y-1].len
				doc.absolutCursor.wantX = doc.absolutCursor.x
				doc.updateLine(&undoItem, y-1, doc.text[y-1].line+doc.text[y].line)
				doc.deleteLine(&undoItem, y)
				doc.undoStack.push(undoItem)

				doc.absolutCursor.y--
				doc.adjustViewport()
				doc.renderScreen()
			}
		} else {
			runes := []rune(doc.text[y].line)
			beforeCursor := string(runes[:x-1])
			afterCursor := string(runes[x:])
			doc.updateLine(&undoItem, y, beforeCursor+afterCursor)
			doc.undoStack.push(undoItem)
			doc.renderLine(y)

			doc.absolutCursor.x--
			doc.absolutCursor.wantX = doc.absolutCursor.x
			doc.alignCursorX()
		}
	} else if event.Key() == 271 {
		// delete
		undoItem := newUndoItem()
		x := doc.absolutCursor.x
		y := doc.absolutCursor.y
		l := len(doc.text)
		if x == doc.text[y].len {
			// pressing delete when cursor is at end of line
			if y+1 < l {
				doc.updateLine(&undoItem, y, doc.text[y].line+doc.text[y+1].line)
				doc.deleteLine(&undoItem, y+1)
				doc.undoStack.push(undoItem)

				doc.adjustViewport()
				doc.renderScreen()
			}
		} else {
			// pressing delete somewhere in the line
			runes := []rune(doc.text[y].line)
			beforeCursor := string(runes[:x])
			afterCursor := string(runes[x+1:])
			doc.updateLine(&undoItem, y, beforeCursor+afterCursor)
			doc.undoStack.push(undoItem)

			doc.renderLine(y)
			doc.alignCursorX()
		}
	} else if event.Key() == tcell.KeyEnter {
		// enter
		undoItem := newUndoItem()
		x := doc.absolutCursor.x
		y := doc.absolutCursor.y
		newLine := ""
		if x != doc.text[y].len {
			// split line if cursor is not at the end
			runes := []rune(doc.text[y].line)
			beforeCursor := string(runes[:x])
			newLine = string(runes[x:])
			doc.updateLine(&undoItem, y, beforeCursor)
		}
		doc.insertLine(&undoItem, y+1, newLine)
		doc.absolutCursor.x = 0
		doc.absolutCursor.wantX = 0
		doc.absolutCursor.y++
		doc.adjustViewport()
		doc.renderScreen()
		doc.screen.Clear()

	} else if event.Key() == tcell.KeyCtrlZ {
		// Undo
		doc.undo()
	} else if event.Key() == tcell.KeyCtrlS {
		// save
		doc.screen.Clear()
		doc.save()
	} else if event.Key() == tcell.KeyRune {
		// insert character (rune)
		undoItem := newUndoItem()
		x := doc.absolutCursor.x
		y := doc.absolutCursor.y

		runes := []rune(doc.text[y].line)
		beforeCursor := string(runes[:x])
		afterCursor := string(runes[x:])
		doc.updateLine(&undoItem, y, beforeCursor+string(event.Rune())+afterCursor)
		doc.undoStack.push(undoItem)

		doc.absolutCursor.x++
		doc.absolutCursor.wantX = doc.absolutCursor.x
		doc.adjustViewport()
		doc.renderLine(y)
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
