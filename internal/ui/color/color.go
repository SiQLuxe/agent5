package color

import "github.com/gdamore/tcell/v2"

func HexToTCell(hex string) tcell.Color {
	if len(hex) == 0 {
		return tcell.ColorDefault
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return tcell.ColorDefault
	}
	return tcell.NewHexColor(int32(
		(hexDigit(hex[0])<<20 |
			hexDigit(hex[1])<<16 |
			hexDigit(hex[2])<<12 |
			hexDigit(hex[3])<<8 |
			hexDigit(hex[4])<<4 |
			hexDigit(hex[5])),
	))
}

func hexDigit(c byte) int32 {
	switch {
	case c >= '0' && c <= '9':
		return int32(c - '0')
	case c >= 'a' && c <= 'f':
		return int32(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int32(c - 'A' + 10)
	default:
		return 0
	}
}
