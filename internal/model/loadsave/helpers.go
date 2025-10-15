package loadsave

import (
	"fmt"
	"strings"
)

func nthPriorityColumnHeader(n int) string {
	return fmt.Sprintf("Priorität %d", n)
}

func invalidHeaderError(sheetName string, gotHeader, wantHeader []string) error {
	return fmt.Errorf(
		"Tabellenblatt: %s\nKopfzeile anders als erwartet. Gefunden: '%v', Erwartet: '%v'",
		sheetName,
		strings.Join(gotHeader, ", "),
		strings.Join(wantHeader, ", "),
	)
}
