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
