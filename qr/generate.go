package qr

import (
	"fmt"
	"strings"

	qr "github.com/skip2/go-qrcode"
)

func GenQRCode(text string, cellSize int) (string, error) {
	qrc, err := qr.New(text, qr.Medium)
	if err != nil {
		return "", err
	}

	bitmap := qrc.Bitmap() // [][]bool, true = czarny moduł
	size := len(bitmap) * cellSize
	var sb strings.Builder

	fmt.Fprintf(&sb,
		`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		size, size, size, size,
	)
	sb.WriteString(`<rect width="100%" height="100%" fill="white"/>`)

	for y, row := range bitmap {
		for x, filled := range row {
			if filled {
				fmt.Fprintf(&sb,
					`<rect x="%d" y="%d" width="%d" height="%d" fill="black"/>`,
					x*cellSize, y*cellSize, cellSize, cellSize,
				)
			}
		}
	}

	sb.WriteString(`</svg>`)
	return sb.String(), nil
}
