package locale

// USD returns the US Dollar locale.
func USD() Locale { return usdLocale{} }

type usdLocale struct{}

func (usdLocale) Currency() string { return "USD" }

func (usdLocale) CurrencyAliases() []string {
	return []string{"$", "dollar", "dollars", "usd"}
}

func (usdLocale) PaymentAliases() map[string]string {
	return map[string]string{
		"venmo":   "mobile_wallet",
		"cashapp": "mobile_wallet",
		"cash app": "mobile_wallet",
		"zelle":   "mobile_wallet",
		"paypal":  "mobile_wallet",
	}
}
