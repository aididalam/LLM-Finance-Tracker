package locale

// BDT returns the Bangladesh locale.
func BDT() Locale { return bdtLocale{} }

type bdtLocale struct{}

func (bdtLocale) Currency() string { return "BDT" }

func (bdtLocale) CurrencyAliases() []string {
	return []string{"taka", "tk", "৳"}
}

func (bdtLocale) PaymentAliases() map[string]string {
	return map[string]string{
		"bkash":  "mobile_wallet",
		"bikash": "mobile_wallet",
		"nagad":  "mobile_wallet",
		"rocket": "mobile_wallet",
	}
}
