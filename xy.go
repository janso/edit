package main

type xyStruct struct {
	x, y int
}

func (xy *xyStruct) equals(xy2 xyStruct) bool {
	return xy.x == xy2.x && xy.y == xy2.y
}

func (xy *xyStruct) less(xy2 xyStruct) bool {
	if xy.y < xy2.y {
		return true
	} else if xy.y > xy2.y {
		return false
	} else {
		// for equal row compare columns
		if xy.x < xy2.x {
			return true
		} else {
			return false
		}
	}
}

func (xy *xyStruct) lessOrEqual(xy2 xyStruct) bool {
	return xy.less(xy2) || xy.equals(xy2)
}

func (xy *xyStruct) greater(xy2 xyStruct) bool {
	if xy.y > xy2.y {
		return true
	} else if xy.y < xy2.y {
		return false
	} else {
		// for equal row compare columns
		if xy.x > xy2.x {
			return true
		} else {
			return false
		}
	}
}

func (xy *xyStruct) greaterOrEqual(xy2 xyStruct) bool {
	return xy.greater(xy2) || xy.equals(xy2)
}

func (xy xyStruct) in(sel selectionStruct) bool {
	if sel.begin.lessOrEqual(xy) && sel.end.greaterOrEqual(xy) {
		return true
	} else {
		return false
	}
}
