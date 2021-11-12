package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type ActionStruct struct {
	row     int
	delete  bool
	insert  bool
	update  bool
	line    LineType
	cursorX int
}

type UndoItemStruct struct {
	actionSlice []ActionStruct
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
	if !us.merge(ui) {
		us.undoSlice = append(us.undoSlice, ui)
		us.top++
	}
}

func (us *UndoStackStruct) merge(ui UndoItemStruct) bool {
	if len(ui.actionSlice) > 1 {
		return false // multiple actions... can't merge
	}
	action := ui.actionSlice[0]
	if action.insert || action.delete {
		return false // insert or delete... can't merge
	}
	// get previous undo item
	if us.top <= 0 {
		return false // no unto items available
	}
	prevUndoItem := us.undoSlice[us.top-1]

	if len(prevUndoItem.actionSlice) > 1 {
		return false // multiple actions... can't merge
	}
	prevAction := prevUndoItem.actionSlice[0]
	if prevAction.insert || prevAction.delete {
		return false // insert or delete... can't merge
	}
	if action.row != prevAction.row {
		return false // different rows affected... can't merge
	}
	// not required to save a new undo item (single updates on same row)
	return true
}

func (us *UndoStackStruct) pop() (UndoItemStruct, error) {
	if us.top <= 0 {
		return UndoItemStruct{}, errors.New("stack empty")
	}
	us.top--
	ui := us.undoSlice[us.top]
	us.undoSlice = us.undoSlice[:us.top]
	return ui, nil
}

type CursorStruct struct {
	x, y, wantX int
}

type xyStruct struct {
	x, y int
}

func (xy *xyStruct) lowerOrEqual(xy2 xyStruct) bool {
	if xy.y < xy2.y {
		return true
	} else if xy.y > xy2.y {
		return false
	} else {
		// equal row
		if xy.x <= xy2.x {
			return true
		} else {
			return false
		}
	}
}

func (xy *xyStruct) biggerOrEqual(xy2 xyStruct) bool {
	if xy.y > xy2.y {
		return true
	} else if xy.y < xy2.y {
		return false
	} else {
		// equal row
		if xy.x >= xy2.x {
			return true
		} else {
			return false
		}
	}
}

func (xy xyStruct) in(sel selectionStruct) bool {
	if sel.begin.lowerOrEqual(xy) && sel.end.biggerOrEqual(xy) {
		return true
	} else {
		return false
	}
}

type ScreenStruct struct {
	tcell.Screen
	defaultStyle   tcell.Style
	selectionStyle tcell.Style
	infoStyle      tcell.Style
}

func (scr ScreenStruct) renderInfoLine(line string) {
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

type DocStruct struct {
	filename      string
	text          LineSlice
	screen        ScreenStruct
	absolutCursor CursorStruct
	viewport      xyStruct
	undoStack     UndoStackStruct
	selection     selectionStruct
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

func (doc *DocStruct) alignCursorX() {
	doc.absolutCursor.x = doc.absolutCursor.wantX
	len := len(doc.text[doc.absolutCursor.y])
	if doc.absolutCursor.x > len {
		doc.absolutCursor.x = len
	}
	if doc.absolutCursor.x < 0 {
		doc.absolutCursor.x = 0
	}
}

func (doc *DocStruct) mustadjustViewport() bool {
	screenMaxX, screenMaxY := doc.screen.Size()
	if doc.absolutCursor.y-doc.viewport.y >= (screenMaxY - 1) {
		return true
	}
	if doc.absolutCursor.y-doc.viewport.y < 0 {
		return true
	}
	if doc.absolutCursor.x-doc.viewport.x >= (screenMaxX - 1) {
		return true
	}
	if doc.absolutCursor.x-doc.viewport.x < 0 {
		return true
	}
	return false
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
	// doc.screen.renderLine(doc.viewport.x, row, string(doc.text[doc.viewport.y+row]))
	style := doc.screen.defaultStyle
	maxx, maxy := doc.screen.Size()
	xy := xyStruct{
		x: -doc.viewport.x,
		y: row,
	}
	if xy.y < 0 {
		xy.y = 0
	}
	if xy.y >= maxy {
		xy.y = maxy - 1
	}
	for _, r := range doc.text[doc.viewport.y+row] {
		var comb []rune
		w := runewidth.RuneWidth(r)
		if w == 0 {
			comb = []rune{r}
			r = ' '
			w = 1
		}
		if xy.x >= 0 {
			if xy.in(doc.selection) {
				style = doc.screen.selectionStyle
			} else {
				style = doc.screen.defaultStyle
			}
			doc.screen.SetContent(xy.x, xy.y, r, comb, style)
		}
		xy.x += w
		if xy.x >= maxx {
			break
		}
	}
	// clear rest of line
	for ; xy.x < maxx-1; xy.x++ {
		doc.screen.SetContent(xy.x, xy.y, ' ', nil, doc.screen.defaultStyle)
	}
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

func (doc *DocStruct) handleEventCursorDown() {
	doc.absolutCursor.y++
	if doc.absolutCursor.y >= len(doc.text) {
		doc.absolutCursor.y = len(doc.text) - 1
	}
	doc.alignCursorX()
	doc.adjustViewport()
}

func (doc *DocStruct) handleEventCursorUp() {
	doc.absolutCursor.y--
	if doc.absolutCursor.y < 0 {
		doc.absolutCursor.y = 0
	}
	doc.alignCursorX()
	doc.adjustViewport()
}

func (doc *DocStruct) handleEventCursorRight(event *tcell.EventKey) {
	l := len(doc.text[doc.absolutCursor.y])
	// go the right
	if doc.absolutCursor.x < l {
		if event.Modifiers()&2 != 0 {
			// control is pressed - go one word to the right
			for ; doc.absolutCursor.x < l; doc.absolutCursor.x++ {
				r := doc.text[doc.absolutCursor.y][doc.absolutCursor.x]
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					break
				}
			}
			for ; doc.absolutCursor.x < l; doc.absolutCursor.x++ {
				r := doc.text[doc.absolutCursor.y][doc.absolutCursor.x]
				if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
					break
				}
			}
		} else {
			// go one character to the right
			doc.absolutCursor.x++
		}
	} else {
		// if cursor is on last position in line, go to beginning of next line
		if doc.absolutCursor.y < len(doc.text)-1 {
			doc.absolutCursor.y++
			doc.absolutCursor.x = 0
		}
	}

	doc.absolutCursor.wantX = doc.absolutCursor.x
	doc.alignCursorX()
	doc.adjustViewport()
}

func (doc *DocStruct) handleEventCursorLeft(event *tcell.EventKey) {
	if doc.absolutCursor.x > 0 {
		if event.Modifiers()&2 != 0 {
			doc.absolutCursor.x--
			// control is pressed - go one word left
			for ; doc.absolutCursor.x > 0; doc.absolutCursor.x-- {
				r := doc.text[doc.absolutCursor.y][doc.absolutCursor.x]
				if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
					break
				}
			}
			for ; doc.absolutCursor.x > 0; doc.absolutCursor.x-- {
				r := doc.text[doc.absolutCursor.y][doc.absolutCursor.x]
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					break
				}
			}
		} else {
			// go one character left
			doc.absolutCursor.x--
		}
	} else {
		// cursor left when cursor is on first position of line
		if doc.absolutCursor.y > 0 {
			doc.absolutCursor.y--
			doc.absolutCursor.x = len(doc.text[doc.absolutCursor.y])
		}
	}
	doc.absolutCursor.wantX = doc.absolutCursor.x
	doc.alignCursorX()
	doc.adjustViewport()
}

func (doc *DocStruct) handleEventCursorBeginOfLine() {
	doc.absolutCursor.x = 0
	doc.absolutCursor.wantX = doc.absolutCursor.x
	doc.alignCursorX()
	doc.adjustViewport()
}

func (doc *DocStruct) handleEventCursorEndOfLine() {
	doc.absolutCursor.x = len(doc.text[doc.absolutCursor.y])
	doc.absolutCursor.wantX = doc.absolutCursor.x
	doc.alignCursorX()
	doc.adjustViewport()
}

func (doc *DocStruct) handleEventPageDown() {
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
}

func (doc *DocStruct) handleEventPageUp() {
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

func (doc *DocStruct) handleEventLoad() error {
	// open file and read line by line
	f, err := os.Open(doc.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	doc.text = make([]LineType, 0, 256)
	scanner := bufio.NewScanner(f) // default delimiter is new line
	for scanner.Scan() {
		doc.text = append(doc.text, LineType(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (doc *DocStruct) handleEventSave() error {
	f, err := os.Create("data.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	for _, line := range doc.text {
		_, err = f.WriteString(string(line))
		if err != nil {
			return err
		}
	}
	return nil
}

func (doc *DocStruct) handleEventUndo() {
	ui, err := doc.undoStack.pop()
	if err != nil {
		return
	}

	for i := len(ui.actionSlice) - 1; i >= 0; i-- {
		action := ui.actionSlice[i]
		if action.delete {
			doc.insertLine(nil, action.row, action.line)
			doc.absolutCursor.x = action.cursorX
			doc.absolutCursor.wantX = action.cursorX
			doc.absolutCursor.y = action.row
			doc.adjustViewport()
		}
		if action.insert {
			doc.deleteLine(nil, action.row)
			doc.absolutCursor.x = action.cursorX
			doc.absolutCursor.wantX = action.cursorX
			doc.absolutCursor.y = action.row
			doc.adjustViewport()
		}
		if action.update {
			doc.text[action.row] = action.line
			doc.absolutCursor.x = action.cursorX
			doc.absolutCursor.wantX = action.cursorX
			doc.absolutCursor.y = action.row
			if !doc.mustadjustViewport() {
				doc.renderLine(action.row)
			} else {
				doc.adjustViewport()
			}
		}
	}
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

func (doc *DocStruct) handleKeyEventCursor(event *tcell.EventKey) (handled bool) {
	handled = true
	if event.Key() == tcell.KeyDown {
		doc.handleEventCursorDown()
	} else if event.Key() == tcell.KeyUp {
		doc.handleEventCursorUp()
	} else if event.Key() == tcell.KeyRight {
		doc.handleEventCursorRight(event)
	} else if event.Key() == tcell.KeyLeft {
		doc.handleEventCursorLeft(event)
	} else if event.Key() == 268 {
		doc.handleEventCursorBeginOfLine()
	} else if event.Key() == 269 {
		doc.handleEventCursorEndOfLine()
	} else if event.Key() == tcell.KeyPgDn {
		doc.handleEventPageDown()
	} else if event.Key() == tcell.KeyPgUp {
		doc.handleEventPageUp()
	} else {
		handled = false
	}
	return
}

func (doc *DocStruct) handleKeyEvent(event *tcell.EventKey) {
	doc.screen.renderInfoLine(
		fmt.Sprintf("C: %3d,%3d | %.3d  %c  %x",
			doc.absolutCursor.x, doc.absolutCursor.y,
			event.Key(),
			event.Rune(),
			event.Modifiers(),
		))

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
