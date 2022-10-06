package count

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

func LocalizedCount() string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%v", number.Decimal(1234))
}
