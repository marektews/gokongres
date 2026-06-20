package czw

import (
	"strconv"
	"strings"
)

/**
*	flexInt akceptuje numer identyfikatora przesyłany przez frontend zarówno jako liczba (123),
*	jak i jako tekst ("123") - pola <input type="number"> w Vue (bez modyfikatora .number)
*	zwracają wartość jako string.
 */
type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(strings.TrimSpace(string(b)), "\"")
	if s == "" || s == "null" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = flexInt(v)
	return nil
}
