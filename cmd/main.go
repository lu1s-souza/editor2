package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/gdamore/tcell/v2"
	"github.com/lu1s-souza/editor2/internal/buffer"
	"github.com/lu1s-souza/editor2/internal/editor"
)

type Editor struct {
	screen           tcell.Screen
	buffer           *buffer.GapBuffer
	cursorX, cursorY int
	offsetX, offsetY int
	width, height    int

	running   bool
	lines     int
	filename  string
	dirty     bool
	status    string
	undoStack []editor.Action
	redoStack []editor.Action
}

func main() {
	editor := &Editor{
		buffer:  buffer.New(1024),
		running: true,
	}

	if len(os.Args) >= 2 {
		editor.OpenFile(os.Args[1])
	} else {
		editor.status = "New file"
	}
	if err := editor.Init(); err != nil {
		log.Fatalf("Failed to initialize editr: %v", err)
	}

	editor.Run()
}

func (e *Editor) Do(action editor.Action) {
	action.Do(e.buffer)
	e.undoStack = append(e.undoStack, action)
	e.redoStack = nil
	e.dirty = true
}

func (e *Editor) Undo() {
	if len(e.undoStack) == 0 {
		return
	}
	lastAction := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]

	lastAction.Undo(e.buffer)

	e.redoStack = append(e.redoStack, lastAction)
	e.dirty = true
}

func (e *Editor) Redo() {
	if len(e.redoStack) == 0 {
		return
	}

	lastAction := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]

	lastAction.Do(e.buffer)

	e.undoStack = append(e.undoStack, lastAction)
	e.dirty = true
}

func (e *Editor) OpenFile(filename string) {

	e.filename = filename
	content, err := os.ReadFile(filename)

	if err != nil {
		if os.IsNotExist(err) {
			e.status = fmt.Sprintf("New file: %s", filename)
			return
		}
		log.Fatalf("Failed to open file: %v", err)
	}

	for _, r := range string(content) {
		e.buffer.Insert(r)
	}
	e.buffer.MoveCursor(-e.buffer.Cursor())
	e.dirty = false
	e.status = fmt.Sprintf("Opened %s", filename)
}

func (e *Editor) SaveFile() {
	if e.filename == "" {
		e.drawSavePopup()
		e.status = "Cannot save: No filename specified."
		return
	}

	content := e.buffer.String()
	err := os.WriteFile(e.filename, []byte(content), 0644)
	if err != nil {
		e.status = fmt.Sprintf("Error saving file: %v", err)
		return
	}

	e.dirty = false
	e.status = fmt.Sprintf("Saved %s (%d bytes)", e.filename, len(content))
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

		event := e.screen.PollEvent()

		e.handleEvent(event)
		e.Draw()
	}
}

func (e *Editor) handleEvent(event tcell.Event) {
	switch ev := event.(type) {
	case *tcell.EventKey:
		e.status = ""

		if ev.Key() == tcell.KeyCtrlZ {
			e.Undo()
			return
		}

		if ev.Key() == tcell.KeyCtrlY {
			e.Redo()
			return
		}

		if ev.Key() == tcell.KeyCtrlS {
			e.SaveFile()
			return
		}
		switch ev.Key() {

		case tcell.KeyEnter:
			action := &editor.InsertAction{Pos: e.buffer.Cursor(), Rune: '\n'}
			e.Do(action)

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if e.buffer.Cursor() > 0 {
				deletedRune := []rune(e.buffer.String())[e.buffer.Cursor()-1]
				action := &editor.DeleteAction{Pos: e.buffer.Cursor(), Rune: deletedRune}
				e.Do(action)
			}
		case tcell.KeyCtrlC, tcell.KeyEscape:
			e.running = false
		case tcell.KeyRune:
			action := &editor.InsertAction{Pos: e.buffer.Cursor(), Rune: ev.Rune()}
			e.Do(action)
		case tcell.KeyLeft:
			e.buffer.MoveCursor(-1)
		case tcell.KeyRight:
			e.buffer.MoveCursor(1)
		case tcell.KeyUp:
			e.moveCursorLine(-1)
		case tcell.KeyDown:
			e.moveCursorLine(1)
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

	textAreaHeight := e.height - 1

	content := e.buffer.String()

	x, y := 0, 0

	for _, r := range content {
		if r == '\n' {
			y++
			x = 0
			continue
		}

		if y < textAreaHeight && x < e.width {
			e.screen.SetContent(x, y, r, nil, tcell.StyleDefault)
		}

		x++
	}

	e.updateCursorPos()
	e.screen.ShowCursor(e.cursorX, e.cursorY)
	e.drawStatusBar()
	e.screen.Show()
}

func (e *Editor) drawStatusBar() {
	statusStyle := tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorBlack)

	filename := e.filename
	if filename == "" {
		filename = "[No Name]"
	}

	dirtyMarker := ""
	if e.dirty {
		dirtyMarker = "*"
	}

	leftStatus := fmt.Sprintf(" %s%s  ", filename, dirtyMarker)
	rightStatus := fmt.Sprintf(" %d:%d ", e.cursorY+1, e.cursorX+1)

	for i := 0; i < e.width; i++ {
		e.screen.SetContent(i, e.height-1, ' ', nil, statusStyle)
	}

	col := 0
	for _, r := range leftStatus {
		e.screen.SetContent(col, e.height-1, r, nil, statusStyle)
		col++
	}

	col = e.width - len(rightStatus)

	for _, r := range rightStatus {
		e.screen.SetContent(col, e.height-1, r, nil, statusStyle)
		col++
	}

	if e.status != "" {
		statusCol := (e.width - len(e.status)) / 2
		for i, r := range e.status {
			e.screen.SetContent(statusCol+i, e.height-1, r, nil, statusStyle)
		}
	}

}

func (e *Editor) drawSavePopup() {
	saving := true
	popupStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	savingBuffer := buffer.New(1024)
	for saving {
		cwd, err := os.Getwd()

		if err != nil {
			return
		}
		saveMsg := "Saving file to directory: "
		startCol := (e.width / 2) - len(saveMsg)

		for _, r := range saveMsg {
			e.screen.SetContent(startCol, e.height/2, r, nil, popupStyle)
			startCol++
		}

		for i, r := range cwd {
			e.screen.SetContent(startCol+i, e.height/2, r, nil, popupStyle)
		}
		e.screen.SetContent(startCol, e.height/2, '\n', nil, popupStyle)

		fileNameMessage := "Enter file name: "
		startCol = (e.width / 2) - len(saveMsg)
		cursorPos := 0
		for _, r := range fileNameMessage {
			e.screen.SetContent(startCol, (e.height/2 + 1), r, nil, popupStyle)
			startCol++
		}
		e.screen.Show()
		event := e.screen.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEsc {
				saving = false
			}
			switch ev.Key() {
			case tcell.KeyRune:
				savingBuffer.Insert(ev.Rune())
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				savingBuffer.Delete()
			case tcell.KeyLeft:
				savingBuffer.MoveCursor(-1)
			case tcell.KeyRight:
				savingBuffer.MoveCursor(1)
			}
		}

		for _, r := range savingBuffer.String() {
			e.screen.SetContent(startCol, e.height/2+1, r, nil, popupStyle)
			startCol++
		}
		cursorPos = startCol
		e.screen.ShowCursor(cursorPos+savingBuffer.Cursor(), e.height/2+1)
		e.screen.Show()
	}
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
