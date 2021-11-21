package main

import (
	"fmt"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func (doc *DocStruct) showCursor() {
	doc.screen.ShowCursor(
		doc.absolutCursor.x-doc.viewport.x,
		doc.absolutCursor.y-doc.viewport.y,
	)
}

func (doc *DocStruct) renderLine(row int) {
	style := doc.screen.defaultStyle
	maxx, maxy := doc.screen.Size()
	xyRelative := xyStruct{x: -doc.viewport.x, y: row}
	if xyRelative.y < 0 {
		xyRelative.y = 0
	}
	if xyRelative.y >= maxy {
		xyRelative.y = maxy - 1
	}
	xyAbsolute := xyStruct{x: 0, y: doc.viewport.y + row}

	// iterate runes of line
	for _, r := range doc.text[xyAbsolute.y] {
		var comb []rune
		w := runewidth.RuneWidth(r)
		if w == 0 {
			comb = []rune{r}
			r = ' '
			w = 1
		}
		if xyRelative.x >= 0 {
			if xyAbsolute.in(doc.selection) {
				style = doc.screen.selectionStyle
			} else {
				style = doc.screen.defaultStyle
			}
			doc.screen.SetContent(xyRelative.x, xyRelative.y, r, comb, style)
		}
		xyRelative.x += w
		xyAbsolute.x++
		if xyRelative.x >= maxx {
			break
		}
	}

	// mark the first character if an empty line is part of selection
	if xyRelative.x == 0 && xyAbsolute.in(doc.selection) {
		doc.screen.SetContent(0, xyRelative.y, ' ', nil, doc.screen.selectionStyle)
		xyRelative.x++
	}

	// clear rest of line
	if xyRelative.x < 0 {
		xyRelative.x = 0
	}
	for ; xyRelative.x < maxx-1; xyRelative.x++ {
		doc.screen.SetContent(xyRelative.x, xyRelative.y, ' ', nil, doc.screen.defaultStyle)
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
	doc.renderInfoLine()
}

func (doc *DocStruct) renderInfoLine() {
	line := fmt.Sprintf("C:%d,%d | P:%d,%d | Ss:%d,%d | Se:%d,%d",
		doc.absolutCursor.x, doc.absolutCursor.y,
		doc.previousCursor.x, doc.previousCursor.y,
		doc.selection.begin.x, doc.selection.begin.y,
		doc.selection.end.x, doc.selection.end.y,
	)

	maxx, _ := doc.screen.Size()
	x := maxx - utf8.RuneCountInString(line) - 1
	doc.renderString(x, 0, line, doc.screen.infoStyle)
	/*
		for _, c := range line {
			var comb []rune
			w := runewidth.RuneWidth(c)
			if w == 0 {
				comb = []rune{c}
				c = ' '
				w = 1
			}
			doc.screen.SetContent(x, 0, c, comb, doc.screen.infoStyle)
			x += w
		}
	*/
}

func (doc *DocStruct) renderKeyInfo(event *tcell.EventKey) {

	keyInfo := fmt.Sprintf("'%c'(%d) Mod:%d",
		event.Rune(),
		event.Key(),
		event.Modifiers(),
	)
	doc.renderString(0, 0, keyInfo, doc.screen.infoStyle)

}

func (doc *DocStruct) renderString(x, y int, s string, style tcell.Style) {
	maxx, maxy := doc.screen.Size()
	if y >= maxy || x >= maxx {
		return
	}

	for _, c := range s {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		if x >= maxx {
			break
		}
		doc.screen.SetContent(x, y, c, comb, style)
		x += w
	}

}

func (doc *DocStruct) mustAdjustViewport() bool {
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
