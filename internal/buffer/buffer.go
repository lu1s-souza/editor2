package buffer

type GapBuffer struct {
	data     []rune
	gapStart int
	gapLen   int
}

func New(initialCapacity int) *GapBuffer {
	if initialCapacity <= 0 {
		initialCapacity = 1024
	}

	return &GapBuffer{
		data:     make([]rune, initialCapacity),
		gapStart: 0,
		gapLen:   initialCapacity,
	}
}

func (b *GapBuffer) String() string {
	gapEnd := b.gapStart + b.gapLen
	return string(b.data[:b.gapStart]) + string(b.data[gapEnd:])
}

func (b *GapBuffer) moveGap(pos int) {
	if pos == b.gapStart {
		return
	}

	gapEnd := b.gapStart + b.gapLen

	if pos < b.gapStart {
		// Move gap left: copy text from [pos, gapStart) to the right side of the gap.
		lenToMove := b.gapStart - pos
		dataA := b.data[gapEnd-lenToMove : gapEnd]
		dataB := b.data[pos:b.gapStart]
		copy(dataA, dataB)
	} else { // pos > b.gapStart
		// Move gap right: copy text from [gapEnd, ...) to the left side of the gap.
		lenToMove := pos - b.gapStart
		copy(b.data[b.gapStart:b.gapStart+lenToMove], b.data[gapEnd:gapEnd+lenToMove])
	}
	b.gapStart = pos
}

func (b *GapBuffer) Length() int {
	return len(b.data) - b.gapLen
}

func (b *GapBuffer) Cursor() int {
	return b.gapStart
}

func (b *GapBuffer) MoveCursor(offset int) {
	newPos := b.gapStart + offset

	if newPos < 0 {
		newPos = 0
	}

	if newPos > b.Length() {
		newPos = b.Length()
	}
	b.moveGap(newPos)
}

func (b *GapBuffer) grow(needed int) {
	newSize := len(b.data) * 2
	if newSize < len(b.data)+needed {
		newSize = len(b.data) + needed
	}

	newData := make([]rune, newSize)
	gapEnd := b.gapStart + b.gapLen

	copy(newData, b.data[:b.gapStart])
	copy(newData[newSize-(len(b.data)-gapEnd):], b.data[gapEnd:])

	b.gapLen = newSize - b.Length()
	b.data = newData
}

func (b *GapBuffer) Insert(r rune) {
	if b.gapLen < 1 {
		b.grow(1)
	}

	b.data[b.gapStart] = r
	b.gapStart++
	b.gapLen--
}

func (b *GapBuffer) Delete() {
	if b.gapStart > 0 {
		b.gapStart--
		b.gapLen++
	}
}
