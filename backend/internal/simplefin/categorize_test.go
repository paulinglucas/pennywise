package simplefin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeTransaction(t *testing.T) {
	tests := []struct {
		description string
		expected    string
	}{
		{"WALMART SUPERCENTER #1234", "shopping"},
		{"AMZN Mktp US*AB1CD2EF3", "shopping"},
		{"TARGET 00012345", "shopping"},
		{"STARBUCKS STORE 12345", "dining"},
		{"MCDONALD'S F12345", "dining"},
		{"CHIPOTLE 1234", "dining"},
		{"DOORDASH*ORDER", "dining"},
		{"KROGER #12345", "groceries"},
		{"TRADER JOE'S #123", "groceries"},
		{"WHOLE FOODS MKT", "groceries"},
		{"INSTACART", "groceries"},
		{"SHELL OIL 57442364500", "gas"},
		{"CHEVRON 0012345", "gas"},
		{"UBER TRIP", "transportation"},
		{"LYFT *RIDE", "transportation"},
		{"NETFLIX.COM", "subscriptions"},
		{"SPOTIFY USA", "subscriptions"},
		{"HULU 123456789", "subscriptions"},
		{"VERIZON WIRELESS", "phone_internet"},
		{"COMCAST CABLE", "phone_internet"},
		{"DUKE ENERGY", "utilities"},
		{"CVS/PHARMACY #12345", "healthcare"},
		{"WALGREENS #1234", "healthcare"},
		{"PLANET FITNESS", "fitness"},
		{"GEICO AUTO", "insurance"},
		{"STATE FARM INSURANCE", "insurance"},
		{"RENT PAYMENT", "housing"},
		{"MORTGAGE PMT", "housing"},
		{"PAYROLL DIRECT DEP", "income"},
		{"ACH DEPOSIT EMPLOYER INC", "income"},
		{"INTEREST PAYMENT", "investment_income"},
		{"DIVIDEND", "investment_income"},
		{"ZELLE PAYMENT", "transfer"},
		{"VENMO CASHOUT", "transfer"},
		{"ATM WITHDRAWAL", "cash"},
		{"RANDOM UNKNOWN MERCHANT", "uncategorized"},
		{"", "uncategorized"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := categorizeTransaction(tt.description)
			assert.Equal(t, tt.expected, result, "description: %s", tt.description)
		})
	}
}
