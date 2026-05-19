package locale

// Locale defines currency symbols/words and payment app aliases for a region.
// Adding a new country/currency is just adding a new file that implements this.
type Locale interface {
	// Currency returns the ISO 4217 code for this locale (e.g. "BDT").
	Currency() string

	// CurrencyAliases returns words and symbols that should map to this locale's currency.
	// Values are matched case-insensitively.
	// Example for BDT: ["taka", "tk", "৳"]
	CurrencyAliases() []string

	// PaymentAliases maps local payment app/service names to standard payment methods.
	// Keys are lowercase. Values are one of: "mobile_wallet", "bank", "cash", etc.
	// Example for BDT: {"bkash": "mobile_wallet", "nagad": "mobile_wallet"}
	PaymentAliases() map[string]string
}
