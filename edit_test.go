package main

import "testing"

func TestXyGreaterLess(t *testing.T) {
	xy1 := xyStruct{x: 10, y: 2}
	if !xy1.lessOrEqual(xyStruct{x: 10, y: 2}) {
		t.Fatalf("Error")
	}
	if !xy1.lessOrEqual(xyStruct{x: 11, y: 2}) {
		t.Fatalf("Error")
	}
	if !xy1.lessOrEqual(xyStruct{x: 100, y: 2}) {
		t.Fatalf("Error")
	}
	if !xy1.lessOrEqual(xyStruct{x: 1, y: 3}) {
		t.Fatalf("Error")
	}
	if !xy1.lessOrEqual(xyStruct{x: 0, y: 3}) {
		t.Fatalf("Error")
	}
	if xy1.lessOrEqual(xyStruct{x: 5, y: 2}) {
		t.Fatalf("Error")
	}
	if xy1.lessOrEqual(xyStruct{x: 20, y: 1}) {
		t.Fatalf("Error")
	}
	if xy1.lessOrEqual(xyStruct{x: 0, y: 0}) {
		t.Fatalf("Error")
	}
	if xy1.lessOrEqual(xyStruct{x: -1, y: -1}) {
		t.Fatalf("Error")
	}

	if !xy1.greaterOrEqual(xyStruct{x: 10, y: 2}) {
		t.Fatalf("Error")
	}
	if xy1.greaterOrEqual(xyStruct{x: 11, y: 2}) {
		t.Fatalf("Error")
	}
	if xy1.greaterOrEqual(xyStruct{x: 100, y: 2}) {
		t.Fatalf("Error")
	}
	if xy1.greaterOrEqual(xyStruct{x: 1, y: 3}) {
		t.Fatalf("Error")
	}
	if xy1.greaterOrEqual(xyStruct{x: 0, y: 3}) {
		t.Fatalf("Error")
	}
	if !xy1.greaterOrEqual(xyStruct{x: 5, y: 2}) {
		t.Fatalf("Error")
	}
	if !xy1.greaterOrEqual(xyStruct{x: 20, y: 1}) {
		t.Fatalf("Error")
	}
	if !xy1.greaterOrEqual(xyStruct{x: 0, y: 0}) {
		t.Fatalf("Error")
	}
	if !xy1.greaterOrEqual(xyStruct{x: -1, y: -1}) {
		t.Fatalf("Error")
	}
}

func TestSelectionIn(t *testing.T) {
	// single line
	sel := selectionStruct{
		begin: xyStruct{x: 2, y: 2},
		end:   xyStruct{x: 4, y: 2},
	}
	if !(xyStruct{x: 3, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 2, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 4, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 1, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 5, y: 2}).in(sel) {
		t.Fatalf("Error")
	}

	if (xyStruct{x: 3, y: 3}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 2, y: 3}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 4, y: 3}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 1, y: 3}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 5, y: 3}).in(sel) {
		t.Fatalf("Error")
	}

	if (xyStruct{x: 3, y: 1}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 2, y: 1}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 4, y: 1}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 1, y: 1}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 5, y: 1}).in(sel) {
		t.Fatalf("Error")
	}

	// multi line
	sel = selectionStruct{
		begin: xyStruct{x: 10, y: 2},
		end:   xyStruct{x: 5, y: 4},
	}
	if !(xyStruct{x: 11, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 10, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 4, y: 4}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 0, y: 4}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 0, y: 3}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 1000, y: 3}).in(sel) {
		t.Fatalf("Error")
	}
	if !(xyStruct{x: 1000, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: -1, y: -1}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 1000, y: 1000}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 9, y: 2}).in(sel) {
		t.Fatalf("Error")
	}
	if (xyStruct{x: 6, y: 4}).in(sel) {
		t.Fatalf("Error")
	}
}
