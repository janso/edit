package main

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

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

func (doc *DocStruct) handleKeyEventCursor(event *tcell.EventKey) (handled bool) {
	doc.renderKeyInfo(event)
	handled = true
	savedCursor := doc.absolutCursor
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
	if handled {
		doc.previousCursor = savedCursor
		setSelection := event.Modifiers()&1 != 0
		doc.updateSelection(setSelection) // false -- remove selection

		doc.renderInfoLine()
	}
	return
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
