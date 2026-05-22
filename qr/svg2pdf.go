package qr

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
)

/**
 * SVGToPDF konwertuje dane SVG (w formie stringa) na PDF (w formie []byte).
 * Używa biblioteki oksvg do renderowania SVG do obrazu RGBA, a następnie fpdf do umieszczenia tego obrazu na stronie PDF.
 * Strona PDF jest ustawiona na format A4 w orientacji poziomej, a obraz SVG jest skalowany, aby wypełnić całą stronę.
 */
func SVGToPDF(svgData string) ([]byte, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("./rsvg-convert.exe", "-f", "pdf")

	case "linux":
		cmd = exec.Command("rsvg-convert", "-f", "pdf")
	}

	// cmd.Env = append(cmd.Env,
	// 	"FONTCONFIG_PATH=./fonts",
	// 	"FONTCONFIG_FILE=./fonts/fonts.conf",
	// )

	cmd.Stdin = bytes.NewReader([]byte(svgData))

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("rsvg-convert failed: %w: %s", err, stderr.String())
	}

	return out.Bytes(), nil
}
