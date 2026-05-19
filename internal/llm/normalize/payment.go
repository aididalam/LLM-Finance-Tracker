package normalize

import "strings"

// universalMethods covers payment method names that are not locale-specific.
var universalMethods = map[string]string{
	"credit_card":    "credit_card",
	"credit card":    "credit_card",
	"creditcard":     "credit_card",
	"credit-card":    "credit_card",
	"cc":             "credit_card",
	"debit_card":     "debit_card",
	"debit card":     "debit_card",
	"debitcard":      "debit_card",
	"debit-card":     "debit_card",
	"debit":          "debit_card",
	"card":           "card_unknown",
	"visa":           "card_unknown",
	"mastercard":     "card_unknown",
	"pos":            "card_unknown",
	"card_unknown":   "card_unknown",
	"unknown_card":   "card_unknown",
	"bank":           "bank",
	"bank_transfer":  "bank",
	"bank transfer":  "bank",
	"neft":           "bank",
	"rtgs":           "bank",
	"imps":           "bank",
	"net banking":    "bank",
	"mobile_wallet":  "mobile_wallet",
	"mobile wallet":  "mobile_wallet",
	"cash":           "cash",
	"":               "cash",
	"other":          "other",
}

// PaymentMethod normalizes a payment method string.
// Universal names are resolved first; locale-specific app names (bKash, UPI, etc.)
// are resolved by checking all registered locales.
func PaymentMethod(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if m, ok := universalMethods[v]; ok {
		return m
	}
	for _, loc := range registry {
		if m, ok := loc.PaymentAliases()[v]; ok {
			return m
		}
	}
	return "other"
}
