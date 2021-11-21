package main

import "errors"

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
			if !doc.mustAdjustViewport() {
				doc.renderLine(action.row)
			} else {
				doc.adjustViewport()
			}
		}
	}
}
