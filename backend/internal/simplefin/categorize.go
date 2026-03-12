package simplefin

import "strings"

type categoryRule struct {
	keywords []string
	category string
}

var categoryRules = []categoryRule{
	{keywords: []string{"walmart", "target", "costco", "sam's club", "dollar tree", "dollar general", "five below", "big lots", "amazon", "amzn"}, category: "shopping"},
	{keywords: []string{"starbucks", "mcdonald", "chipotle", "chick-fil-a", "subway", "taco bell", "wendy", "burger king", "panera", "domino", "pizza hut", "panda express", "dunkin", "popeyes", "arby", "kfc", "sonic", "waffle house", "ihop", "denny", "applebee", "olive garden", "chili", "outback", "red lobster", "buffalo wild", "five guys", "in-n-out", "jack in the box", "whataburger", "raising cane", "wingstop", "zaxby", "grubhub", "doordash", "uber eats", "postmates"}, category: "dining"},
	{keywords: []string{"kroger", "safeway", "publix", "aldi", "trader joe", "whole foods", "sprouts", "h-e-b", "heb", "wegmans", "food lion", "harris teeter", "piggly wiggly", "winn-dixie", "albertson", "vons", "ralphs", "meijer", "market basket", "stop & shop", "giant eagle", "shoprite", "instacart", "fresh market"}, category: "groceries"},
	{keywords: []string{"shell", "exxon", "chevron", "bp ", "mobil", "texaco", "marathon", "sunoco", "speedway", "wawa", "quiktrip", "racetrac", "circle k", "pilot", "loves", "murphy usa", "sheetz"}, category: "gas"},
	{keywords: []string{"uber", "lyft", "parking", "transit", "metro", "bus ", "amtrak", "toll", "ez pass", "ezpass"}, category: "transportation"},
	{keywords: []string{"netflix", "spotify", "hulu", "disney+", "disney plus", "hbo", "apple tv", "youtube premium", "paramount", "peacock", "crunchyroll", "audible"}, category: "subscriptions"},
	{keywords: []string{"comcast", "xfinity", "verizon", "at&t", "t-mobile", "sprint", "spectrum", "cox ", "optimum", "frontier", "centurylink", "mint mobile", "visible", "google fi"}, category: "phone_internet"},
	{keywords: []string{"electric", "water bill", "gas bill", "sewer", "utility", "power co", "energy", "national grid", "duke energy", "dominion", "southern co", "pge", "pg&e", "con edison"}, category: "utilities"},
	{keywords: []string{"cvs", "walgreens", "rite aid", "pharmacy", "doctor", "hospital", "medical", "dental", "optom", "urgent care", "labcorp", "quest diag", "kaiser", "aetna", "cigna", "united health", "blue cross", "anthem", "copay"}, category: "healthcare"},
	{keywords: []string{"planet fitness", "la fitness", "anytime fitness", "orangetheory", "crossfit", "ymca", "equinox", "gold's gym", "crunch fitness", "lifetime fitness", "peloton"}, category: "fitness"},
	{keywords: []string{"geico", "state farm", "progressive", "allstate", "liberty mutual", "farmers", "usaa", "nationwide", "travelers", "erie insurance", "american family", "insurance"}, category: "insurance"},
	{keywords: []string{"rent", "mortgage", "hoa", "property tax", "homeowner"}, category: "housing"},
	{keywords: []string{"tuition", "student loan", "coursera", "udemy", "skillshare", "linkedin learning", "college", "university"}, category: "education"},
	{keywords: []string{"payroll", "direct dep", "salary", "wage", "paycheck", "ach deposit", "employer"}, category: "income"},
	{keywords: []string{"interest", "dividend", "capital gain"}, category: "investment_income"},
	{keywords: []string{"transfer", "zelle", "venmo", "cashapp", "cash app", "paypal"}, category: "transfer"},
	{keywords: []string{"atm", "withdrawal", "cash back"}, category: "cash"},
}

func categorizeTransaction(description string) string {
	lower := strings.ToLower(description)

	for _, rule := range categoryRules {
		for _, keyword := range rule.keywords {
			if strings.Contains(lower, keyword) {
				return rule.category
			}
		}
	}

	return "uncategorized"
}
