package db

import (
	"fmt"
)

/**
*   Budowanie identyfikatora autokaru
*   Format: T16, D11, W23, M102, itp.
 */
func CreateShortBusID(sra *SRA, sectorName string, sectorOrder int) string {
	if sra.StaticIdentifier != nil {
		return *sra.StaticIdentifier
	} else {
		return fmt.Sprintf("%s%d", sectorName, sectorOrder)
	}
}
