package main

import (
	"log"
	"regexp"

	"github.com/gdamore/tcell/v2"
	"github.com/lu1s-souza/editor2/internal/buffer"
)

type Editor struct {
	screen           tcell.Screen
	buffer           *buffer.GapBuffer
	cursorX, cursorY int
	offsetX, offsetY int
	width, height    int

	running bool
	lines   int
}

func main() {
	editor := &Editor{
		buffer:  buffer.New(1024),
		running: true,
	}

	if err := editor.Init(); err != nil {
		log.Fatalf("Failed to initialize editr: %v", err)
	}

	editor.Run()
}

func (e *Editor) Init() error {
	var err error
	e.screen, err = tcell.NewScreen()

	if err != nil {
		return err
	}

	if err := e.screen.Init(); err != nil {
		return err
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	e.screen.SetStyle(defStyle)

	e.width, e.height = e.screen.Size()

	return nil
}

func (e *Editor) Run() {
	defer e.screen.Fini()

	for e.running {
		e.Draw()

		event := e.screen.PollEvent()

		e.handleEvent(event)
	}
}

func (e *Editor) handleEvent(event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventKey:
		switch ev.Key() {

		case tcell.KeyCtrlC, tcell.KeyEscape:
			e.running = false

		case tcell.KeyLeft:
			e.buffer.MoveCursor(-1)
		case tcell.KeyRight:
			e.buffer.MoveCursor(1)
		case tcell.KeyUp:
			e.moveCursorLine(-1)
		case tcell.KeyDown:
			e.moveCursorLine(1)

		case tcell.KeyEnter:
			e.buffer.Insert('\n')
		case tcell.KeyBackspace2, tcell.KeyBackspace:
			e.buffer.Delete()

		case tcell.KeyRune:
			e.buffer.Insert(ev.Rune())

		case tcell.KeyEnd:
			e.handleEnd()
		case tcell.KeyHome:
			e.handleHome()
		}
	case *tcell.EventResize:
		e.width, e.height = e.screen.Size()
	}
}

func (e *Editor) handleEnd() {
	lineMap := e.buildLineMap()
	currentLine := e.getCurrentLine(lineMap)

	e.buffer.MoveCursor(lineMap[currentLine][1] - e.buffer.Cursor())
}

func (e *Editor) handleHome() {
	lineMap := e.buildLineMap()
	currentLine := e.getCurrentLine(lineMap)

	e.buffer.MoveCursor(lineMap[currentLine][0] - e.buffer.Cursor())
}

func (e *Editor) getCurrentLine(lineMap map[int][]int) int {
	currentLine := 0
	cursorPos := e.buffer.Cursor()
	for key, val := range lineMap {
		if val[1] >= cursorPos && cursorPos >= val[0] {
			currentLine = key
		}
	}

	return currentLine
}

func (e *Editor) buildLineMap() map[int][]int {

	re := regexp.MustCompile("\n")

	lines := re.Split(e.buffer.String(), -1)

	lineMap := make(map[int][]int)
	prevLineStart := 0
	prevLineSize := 0

	for i, r := range lines {

		if i == 0 {
			lineMap[i] = []int{0, len(r)}
			prevLineSize = len(r)
			continue
		}

		lineStartPos := prevLineStart + prevLineSize + 1
		prevLineStart = lineStartPos
		prevLineSize = len(r)
		lineMap[i] = []int{lineStartPos, lineStartPos + len(r)}
	}
	e.lines = len(lines) - 1
	return lineMap
}

func (e Editor) moveCursorLine(offset int) {
	cursorPos := e.buffer.Cursor()

	lineMap := e.buildLineMap()

	currentLine := e.getCurrentLine(lineMap)

	targetLine := currentLine + offset
	currentLineCursorOffset := cursorPos - lineMap[currentLine][0]

	if targetLine < 0 {
		targetLine = e.lines
	}

	if targetLine > e.lines {
		targetLine = 0
	}

	currentLineSize := lineMap[currentLine][1] - lineMap[currentLine][0]
	targetLineSize := lineMap[targetLine][1] - lineMap[targetLine][0]

	isTargetLineSmaller := currentLineSize > targetLineSize
	shouldRepositionCursor := isTargetLineSmaller && currentLineCursorOffset > targetLineSize

	newCursorPos := lineMap[targetLine][0] + (currentLineCursorOffset)

	if shouldRepositionCursor {
		newCursorPos = lineMap[targetLine][1]
	}

	e.buffer.MoveCursor(newCursorPos - cursorPos)
}

func (e *Editor) Draw() {
	e.screen.Clear()

	content := e.buffer.String()

	x, y := 0, 0

	for _, r := range content {
		if r == '\n' {
			y++
			x = 0
			continue
		}

		if y < e.height && x < e.width {
			e.screen.SetContent(x, y, r, nil, tcell.StyleDefault)
		}

		x++
	}

	e.updateCursorPos()
	e.screen.ShowCursor(e.cursorX, e.cursorY)

	e.screen.Show()
}

func (e *Editor) updateCursorPos() {
	cursorPos := e.buffer.Cursor()
	contentToCursor := e.buffer.String()[:cursorPos]

	x, y := 0, 0

	for _, r := range contentToCursor {
		if r == '\n' {
			y++
			x = 0
		} else {
			x++
		}
	}

	e.cursorX = x
	e.cursorY = y
}
