package helpers

import "time"

/**
* Zwraca:
* 	"d1" - każdego innego dnia niż sobota i niedziela
*   "d2" - w sobotę
*   "d3" - w niedzielę
 */
func GetActiveDay() time.Weekday {
	wd := time.Now().Weekday()
	if wd != time.Saturday && wd != time.Sunday {
		return time.Friday
	} else {
		return wd
	}
}

/**
* Zwraca nazwę pola dnia weryfikacji ("d1"/"d2"/"d3") dla aktywnego dnia kongresu:
*   "d1" - piątek (i każdy inny dzień poza sobotą/niedzielą)
*   "d2" - sobota
*   "d3" - niedziela
 */
func GetActiveDayField() string {
	switch GetActiveDay() {
	case time.Saturday:
		return "d2"
	case time.Sunday:
		return "d3"
	default:
		return "d1"
	}
}
