package editor

import "github.com/lu1s-souza/editor2/internal/buffer"

type Action interface {
	Do(b *buffer.GapBuffer)
	Undo(b *buffer.GapBuffer)
}

type InsertAction struct {
	Pos  int
	Rune rune
}

func (a *InsertAction) Do(b *buffer.GapBuffer) {
	b.MoveCursor(a.Pos - b.Cursor())
	b.Insert(a.Rune)
}

func (a *InsertAction) Undo(b *buffer.GapBuffer) {
	b.MoveCursor(a.Pos + 1 - b.Cursor())
	b.Delete()
}

type DeleteAction struct {
	Pos  int
	Rune rune
}

func (d *DeleteAction) Do(b *buffer.GapBuffer) {
	b.MoveCursor(d.Pos - b.Cursor())
	b.Delete()
}

func (d *DeleteAction) Undo(b *buffer.GapBuffer) {
	b.MoveCursor(d.Pos - 1 - b.Cursor())
	b.Insert(d.Rune)
}
