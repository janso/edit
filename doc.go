package main

import (
	"github.com/gdamore/tcell/v2"
)

type CursorStruct struct {
	x, y, wantX int
}

type ScreenStruct struct {
	tcell.Screen
	defaultStyle   tcell.Style
	selectionStyle tcell.Style
	infoStyle      tcell.Style
}

type LineType []rune

func concatenateLines(lines ...LineType) LineType {
	line := LineType{}
	for _, l := range lines {
		line = append(line, l...)
	}
	return line
}

type LineSlice []LineType

type selectionStruct struct {
	begin, end xyStruct
}

var emptySelection selectionStruct

type DocStruct struct {
	filename       string
	text           LineSlice
	screen         ScreenStruct
	absolutCursor  CursorStruct
	previousCursor CursorStruct
	viewport       xyStruct
	undoStack      UndoStackStruct
	selection      selectionStruct
}

func (doc *DocStruct) updateLine(ui *UndoItemStruct, row int, line LineType) {
	if ui != nil {
		// create action for undo
		action := ActionStruct{
			row:     row,
			delete:  false,
			insert:  false,
			update:  true,
			line:    doc.text[row],
			cursorX: doc.absolutCursor.x,
		}
		ui.actionSlice = append(ui.actionSlice, action)
	}
	// do actual update
	doc.text[row] = line
}

func (doc *DocStruct) insertLine(ui *UndoItemStruct, row int, line LineType) {
	// insert line
	if row >= len(doc.text) {
		doc.text = append(doc.text, LineType{})
	} else {
		doc.text = append(doc.text[:row+1], doc.text[row:]...)
	}
	if ui != nil {
		// create actionItem for Undo
		action := ActionStruct{
			row:     row,
			delete:  false,
			insert:  true,
			update:  true,
			line:    doc.text[row],
			cursorX: doc.absolutCursor.x,
		}
		ui.actionSlice = append(ui.actionSlice, action)
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
			row:     row,
			delete:  true,
			insert:  false,
			update:  true,
			line:    doc.text[row],
			cursorX: doc.absolutCursor.x,
		}
		ui.actionSlice = append(ui.actionSlice, action)
	}
	if row+1 < len(doc.text) {
		doc.text = append(doc.text[:row], doc.text[row+1:]...)
	} else {
		doc.text = doc.text[:row]
	}
}

func (doc *DocStruct) updateSelection(set bool) {
	if !set {
		// reset selection
		if doc.selection == emptySelection {
			return
		}
		doc.selection = emptySelection

		if doc.selection.begin.y == doc.selection.end.y {
			// doc.renderLine(doc.selection.end.y)
		} else {
			doc.renderScreen()
		}
		return
	}
	xyAbsolute := xyStruct{
		x: doc.absolutCursor.x,
		y: doc.absolutCursor.y,
	}

	xyPrevious := xyStruct{
		x: doc.previousCursor.x,
		y: doc.previousCursor.y,
	}
	moveRight := xyAbsolute.greater(xyPrevious)

	if doc.selection == emptySelection {
		if moveRight {
			xyAbsolute.x--
		}
		doc.selection.begin = xyAbsolute
		doc.selection.end = xyAbsolute
	} else {
		if xyAbsolute.greaterOrEqual(doc.selection.end) {
			xyAbsolute.x--
			doc.selection.end = xyAbsolute
		} else {
			doc.selection.begin = xyAbsolute
		}
	}
	if doc.selection.begin.x-1 == doc.selection.end.x {
		doc.selection = emptySelection
	}
	doc.renderScreen()
}

func (doc *DocStruct) deleteSelection(ui *UndoItemStruct) {
	if doc.selection == emptySelection {
		return
	}
	if doc.selection.begin.y == doc.selection.end.y {
		// selection in single line
		y := doc.selection.begin.y
		leftPartOfLine := doc.text[y][:doc.selection.begin.x]
		rightPartOfLine := doc.text[y][doc.selection.end.x:]
		doc.updateLine(ui, y, concatenateLines(leftPartOfLine, rightPartOfLine))
	}

	doc.absolutCursor.x = doc.selection.begin.x
	doc.absolutCursor.wantX = doc.absolutCursor.x
	doc.absolutCursor.y = doc.selection.begin.y
	doc.alignCursorX()

	doc.selection = emptySelection
	// doc.updateLine(ui, y, concatenateLines(doc.text[y], doc.text[y+1]))
	// doc.deleteLine(ui, y+1)
}

func (doc *DocStruct) handleEventInsertCharacter(r rune) {
	undoItem := newUndoItem()
	x := doc.absolutCursor.x
	y := doc.absolutCursor.y

	newRune := LineType{r}
	doc.updateLine(&undoItem, y, concatenateLines(doc.text[y][:x], newRune, doc.text[y][x:]))
	doc.undoStack.push(undoItem)

	doc.absolutCursor.x++
	doc.absolutCursor.wantX = doc.absolutCursor.x
	doc.adjustViewport()
	doc.renderLine(y)
}

func (doc *DocStruct) handleEventBackspace() {
	undoItem := newUndoItem()
	x := doc.absolutCursor.x
	y := doc.absolutCursor.y
	if x <= 0 {
		// backspace when cursor is on first position of line
		if y > 0 {
			doc.absolutCursor.x = len(doc.text[y-1])
			doc.absolutCursor.wantX = doc.absolutCursor.x
			doc.updateLine(&undoItem, y-1, concatenateLines(doc.text[y-1], doc.text[y]))
			doc.deleteLine(&undoItem, y)
			doc.undoStack.push(undoItem)

			doc.absolutCursor.y--
			doc.adjustViewport()
			doc.renderScreen()
		}
	} else {
		doc.updateLine(&undoItem, y, concatenateLines(doc.text[y][:x-1], doc.text[y][x:]))
		doc.undoStack.push(undoItem)
		doc.renderLine(y)

		doc.absolutCursor.x--
		doc.absolutCursor.wantX = doc.absolutCursor.x
		doc.alignCursorX()
	}
}

func (doc *DocStruct) handleEventDelete() {
	undoItem := newUndoItem()

	if doc.selection != emptySelection {
		// dek whole selection
		doc.deleteSelection(&undoItem)
		doc.undoStack.push(undoItem)
		doc.renderScreen()
	}

	x := doc.absolutCursor.x
	y := doc.absolutCursor.y
	l := len(doc.text)
	if x == len(doc.text[y]) {
		// pressing delete when cursor is at end of line
		if y+1 < l {
			doc.updateLine(&undoItem, y, concatenateLines(doc.text[y], doc.text[y+1]))
			doc.deleteLine(&undoItem, y+1)
			doc.undoStack.push(undoItem)

			doc.adjustViewport()
			doc.renderScreen()
		}
	} else {
		// pressing delete somewhere in the line
		doc.updateLine(&undoItem, y, concatenateLines(doc.text[y][:x], doc.text[y][x+1:]))
		doc.undoStack.push(undoItem)

		doc.renderLine(y)
		doc.alignCursorX()
	}
}

func (doc *DocStruct) handleEventEnter() {
	undoItem := newUndoItem()
	x := doc.absolutCursor.x
	y := doc.absolutCursor.y
	newLine := LineType{}
	if x != len(doc.text[y]) {
		// split line if cursor is not at the end
		newLine = doc.text[y][x:]
		doc.updateLine(&undoItem, y, doc.text[y][:x])
	}
	doc.insertLine(&undoItem, y+1, newLine)
	doc.undoStack.push(undoItem)
	doc.absolutCursor.x = 0
	doc.absolutCursor.wantX = 0
	doc.absolutCursor.y++
	doc.adjustViewport()
	doc.renderScreen()
}

func (doc *DocStruct) handleEventInsertTab() {
	undoItem := newUndoItem()
	const fakeTab string = "    "
	x := doc.absolutCursor.x
	y := doc.absolutCursor.y
	doc.updateLine(&undoItem, y, concatenateLines(doc.text[y][:x], LineType(fakeTab), doc.text[y][x:]))
	doc.undoStack.push(undoItem)

	doc.renderLine(y)
	doc.absolutCursor.x += len(fakeTab)
	doc.absolutCursor.wantX = doc.absolutCursor.x
}

func (doc *DocStruct) handleKeyEvent(event *tcell.EventKey) {
	if doc.handleKeyEventCursor(event) {
		return
	} else if event.Key() == tcell.KeyRune {
		doc.handleEventInsertCharacter(event.Rune())
	} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
		doc.handleEventBackspace()
	} else if event.Key() == 271 {
		doc.handleEventDelete()
	} else if event.Key() == tcell.KeyEnter {
		doc.handleEventEnter()
	} else if event.Key() == tcell.KeyCtrlZ || event.Key() == tcell.KeyCtrlR {
		// Undo
		doc.handleEventUndo()
	} else if event.Key() == tcell.KeyCtrlS {
		// save
		doc.handleEventSave()
	} else if event.Key() == tcell.KeyTab {
		doc.handleEventInsertTab()
	}
}
