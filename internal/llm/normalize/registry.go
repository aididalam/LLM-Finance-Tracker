package normalize

import "github.com/aididalam/llmexpensetracker/internal/llm/normalize/locale"

// registry maps ISO 4217 currency codes to their locale.
// Add a new entry here when adding a new locale file.
var registry = map[string]locale.Locale{
	"BDT": locale.BDT(),
	"INR": locale.INR(),
	"USD": locale.USD(),
}
