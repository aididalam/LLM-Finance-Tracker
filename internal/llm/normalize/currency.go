package normalize

import "strings"

// Currency normalizes a currency value to an ISO 4217 code.
// It checks all registered locale aliases, then falls back to defaultCurrency
// for unknown or empty values.
func Currency(value, defaultCurrency string) string {
	v := strings.ToUpper(strings.TrimSpace(value))
	if v == "" {
		return defaultCurrency
	}
	// Check each registered locale's aliases
	for _, loc := range registry {
		for _, alias := range loc.CurrencyAliases() {
			if strings.ToUpper(alias) == v {
				return loc.Currency()
			}
		}
	}
	// 1–3 character values are treated as ISO codes and passed through as-is
	if len(v) <= 3 {
		return v
	}
	// Unknown long string — use the user's default
	return defaultCurrency
}
