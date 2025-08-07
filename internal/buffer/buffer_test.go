package buffer

import "testing"

func TestInsert(t *testing.T) {
	b := New(10)
	b.Insert('a')
	expected := "a"
	if b.String() != expected {
		t.Errorf("expected %q, got %q", expected, b.String())
	}
}

func TestMoveCursor(t *testing.T) {
	b := New(10)
	b.Insert('a')
	b.Insert('c') // Buffer is "ac", cursor at pos 2
	b.MoveCursor(-1) // Move to pos 1
	b.Insert('b')
	expected := "abc"
	if b.String() != expected {
		t.Errorf("expected %q after move and insert, got %q", expected, b.String())
	}
	if b.Cursor() != 2 {
		t.Errorf("expected cursor at pos 2, got %d", b.Cursor())
	}
}

func TestDelete(t *testing.T) {
	b := New(10)
	b.Insert('a')
	b.Insert('b')
	b.Insert('c')
	b.Delete() // Delete 'c'
	if b.String() != "ab" {
		t.Errorf("expected %q after delete, got %q", "ab", b.String())
	}
	b.MoveCursor(-1) // Cursor between 'a' and 'b'
	b.Delete()       // Delete 'a'
	if b.String() != "b" {
		t.Errorf("expected %q after move and delete, got %q", "b", b.String())
	}
}

func TestBufferGrowth(t *testing.T) {
	b := New(3)
	b.Insert('a')
	b.Insert('b')
	b.Insert('c')
	b.Insert('d') // Triggers growth
	expected := "abcd"
	if b.String() != expected {
		t.Errorf("expected %q after growth, got %q", expected, b.String())
	}
	if len(b.data) <= 3 {
		t.Errorf("buffer capacity should be > 3 after growth, got %d", len(b.data))
	}
}

func TestIntegrationScenario(t *testing.T) {
	b := New(5)
	b.Insert('h')
	b.Insert('e')
	b.Insert('l')
	b.Insert('o') // "helo"
	b.MoveCursor(-1) // "hel|o"
	b.Insert('l') // "hell|o"
	b.MoveCursor(1) // "hello|"
	b.Insert(' ')
	b.Insert('w')
	b.Insert('o')
	b.Insert('r')
	b.Insert('l')
	b.Insert('d') // "hello world"

	if b.String() != "hello world" {
		t.Errorf("expected 'hello world', got %q", b.String())
	}

	b.MoveCursor(-5) // "hello |world"
	b.Delete() // "hello|world"
	b.Delete() // "hell|world"
	b.Delete() // "hel|world"

	if b.String() != "helworld" {
		t.Errorf("expected 'helworld', got %q", b.String())
	}
}
