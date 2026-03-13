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
		{"THE HOME DEPOT FRIDLEY MN", "shopping"},
		{"AplPay HOMEDEPOT.COM800-430-3376 GA", "shopping"},
		{"AplPay ETSY, INC. BROOKLYN NY", "shopping"},
		{"KINDLE SVCS*QL5XL259888-802-3080 WA", "shopping"},
		{"STARBUCKS STORE 12345", "dining"},
		{"MCDONALD'S F12345", "dining"},
		{"CHIPOTLE 1234", "dining"},
		{"DOORDASH*ORDER", "dining"},
		{"DD *DOORDASH RUSTYTASAN FRANCISCO CA", "dining"},
		{"KROGER #12345", "groceries"},
		{"TRADER JOE'S #123", "groceries"},
		{"WHOLE FOODS MKT", "groceries"},
		{"INSTACART", "groceries"},
		{"CUB FOODS #1630 0000FRIDLEY MN", "groceries"},
		{"SHELL OIL 57442364500", "gas"},
		{"CHEVRON 0012345", "gas"},
		{"BP#1826254EAST RIVERFRIDLEY MN", "gas"},
		{"KWIK TRIP APPLETON WI", "gas"},
		{"HOLIDAY STATIONS 046MINNEAPOLIS MN", "gas"},
		{"FRIDLEY BP 206464072FRIDLEY MN", "gas"},
		{"UBER TRIP help.uber.com CA", "transportation"},
		{"LYFT *RIDE", "transportation"},
		{"HERTZ CAR RENTAL OKLAHOMA CITY OK", "transportation"},
		{"NETFLIX.COM", "subscriptions"},
		{"SPOTIFY USA", "subscriptions"},
		{"HULU 123456789", "subscriptions"},
		{"GOOGLE *GOOGLE ONE G.CO/HELPPAY# CA", "subscriptions"},
		{"OPENAI *CHATGPT SUBSSAN FRANCISCO CA", "subscriptions"},
		{"CLAUDE.AI SUBSCRIPTISAN FRANCISCO CA", "subscriptions"},
		{"AplPay PATREON* MEMBSAN FRANCISCO CA", "subscriptions"},
		{"MICROSOFT MSBILL.INFO", "subscriptions"},
		{"ADTSECURITY MYADT.CO800-238-2727 FL", "subscriptions"},
		{"VERIZON WIRELESS", "phone_internet"},
		{"COMCAST CABLE", "phone_internet"},
		{"DUKE ENERGY", "utilities"},
		{"CURBSIDE WASTE 00-08BROOKLYN PARK MN", "utilities"},
		{"ECOSHIELD PEST CONTRCHANDLER AZ", "utilities"},
		{"CVS/PHARMACY #12345", "healthcare"},
		{"WALGREENS #1234", "healthcare"},
		{"ALLINAHEAL* ALLINA HMINNEAPOLIS MN", "healthcare"},
		{"PLANET FITNESS", "fitness"},
		{"GEICO AUTO", "insurance"},
		{"STATE FARM INSURANCE", "insurance"},
		{"RENT PAYMENT", "housing"},
		{"MORTGAGE PMT", "housing"},
		{"PAYROLL DIRECT DEP", "income"},
		{"ACH DEPOSIT EMPLOYER INC", "income"},
		{"REINVESTMENT FIDELITY GOVERNMENT MONEY MARKET (SPAXX) (Cash)", "investment_income"},
		{"DIVIDEND RECEIVED", "investment_income"},
		{"MOBILE PAYMENT - THANK YOU", "transfer"},
		{"ONLINE PAYMENT THANK YOU", "transfer"},
		{"CREDIT CARD PAYMENT", "transfer"},
		{"AUTOPAY PAYMENT", "transfer"},
		{"CC PAYMENT ACH", "transfer"},
		{"PAYMENT TO CHASE CARD", "transfer"},
		{"ZELLE PAYMENT", "transfer"},
		{"VENMO CASHOUT", "transfer"},
		{"ATM WITHDRAWAL", "cash"},
		{"PETSMART PHOENIX AZ", "pets"},
		{"METLIFE PET JEFFERSONVILLE IN", "pets"},
		{"DELTA AIR LINES ATLANTA", "travel"},
		{"AIRBNB * HMZ9FCW528 SAN FRANCISCO CA", "travel"},
		{"MOUNT BOHEMIA SKI REEagle Harbor MI", "travel"},
		{"EVO HOTEL SALT LAKE CIT UT", "travel"},
		{"ACME COMEDY COMPANY/MINNEAPOLIS MN", "entertainment"},
		{"RECREATION.GOV 00000ALBUQUERQUE NM", "entertainment"},
		{"AplPay AXS.COM*FIRSTLOS ANGELES CA", "entertainment"},
		{"CARIBOU COFFEE MINNEAPOLIS MN", "dining"},
		{"TST* WHITEYS OLD TOWMINNEAPOLIS MN", "dining"},
		{"SISYPHUS BREWING MINNEAPOLIS MN", "dining"},
		{"BEAVER ISLAND BREWINSt. Cloud MN", "dining"},
		{"FSP*BROKEN CLOCK BREMINNEAPOLIS MN", "dining"},
		{"PAPA MURPHY'S MN028 FRIDLEY MN", "dining"},
		{"CULVERS OF BROOKLYN BROOKLYN CENT MN", "dining"},
		{"FRIDLEY LIQUOR 00-08FRIDLEY MN", "dining"},
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
