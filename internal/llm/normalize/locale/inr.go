package locale

// INR returns the India locale.
func INR() Locale { return inrLocale{} }

type inrLocale struct{}

func (inrLocale) Currency() string { return "INR" }

func (inrLocale) CurrencyAliases() []string {
	return []string{"₹", "rs", "rs.", "rupee", "rupees", "indian rupee"}
}

func (inrLocale) PaymentAliases() map[string]string {
	return map[string]string{
		"phonepay":   "mobile_wallet",
		"phonepe":    "mobile_wallet",
		"gpay":       "mobile_wallet",
		"google pay": "mobile_wallet",
		"paytm":      "mobile_wallet",
		"upi":        "mobile_wallet",
		"bhim":       "mobile_wallet",
	}
}
