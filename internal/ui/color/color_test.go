package color

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHexToTCell_Black(t *testing.T) {
	c := HexToTCell("#000000")
	if c == tcell.ColorDefault {
		t.Fatal("expected a valid color")
	}
}

func TestHexToTCell_White(t *testing.T) {
	c := HexToTCell("#ffffff")
	if c == tcell.ColorDefault {
		t.Fatal("expected a valid color")
	}
}

func TestHexToTCell_Red(t *testing.T) {
	c := HexToTCell("#ff0000")
	if c == tcell.ColorDefault {
		t.Fatal("expected a valid color")
	}
}

func TestHexToTCell_Empty(t *testing.T) {
	c := HexToTCell("")
	if c != tcell.ColorDefault {
		t.Fatal("expected default for empty string")
	}
}

func TestHexToTCell_Short(t *testing.T) {
	c := HexToTCell("#fff")
	if c != tcell.ColorDefault {
		t.Fatal("expected default for short string")
	}
}
