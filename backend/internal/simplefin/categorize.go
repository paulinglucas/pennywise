package simplefin

import "strings"

type categoryRule struct {
	keywords []string
	category string
}

var categoryRules = []categoryRule{
	{keywords: []string{
		"payment thank you", "mobile payment", "online payment", "autopay", "auto pay",
		"cc payment", "minimum payment due", "card payment", "bill pay",
		"payment to chase", "payment to citi", "payment to capital one", "payment to amex",
		"payment to discover", "payment to wells fargo", "payment to bank of america",
		"payment to barclays", "payment to synchrony", "payment to navy federal",
		"transfer", "zelle", "venmo", "cashapp", "cash app", "paypal",
		"credit card payment", "ach payment", "wire transfer", "bank transfer",
		"direct payment", "electronic payment", "epayment",
	}, category: "transfer"},
	{keywords: []string{
		"payroll", "direct dep", "salary", "wage ", "paycheck", "ach deposit", "employer",
		"gusto", "adp payroll", "paychex", "irs treas", "tax refund", "soc sec",
		"social security", "pension", "retirement dist",
	}, category: "income"},
	{keywords: []string{
		"reinvestment", "interest earned", "dividend", "capital gain", "realized gain",
		"dist reinvest", "interest payment",
	}, category: "investment_income"},
	{keywords: []string{
		"atm ", "withdrawal", "cash back", "cash advance",
	}, category: "cash"},
	{keywords: []string{
		"rent ", "mortgage", "hoa ", "property tax", "homeowner", "escrow",
		"property mgmt", "property management", "lease payment", "apt ",
	}, category: "housing"},
	{keywords: []string{
		"kroger", "safeway", "publix", "aldi", "trader joe", "whole foods", "sprouts",
		"h-e-b", "heb ", "wegmans", "food lion", "harris teeter", "piggly wiggly",
		"winn-dixie", "albertson", "vons", "ralphs", "meijer", "market basket",
		"stop & shop", "giant eagle", "shoprite", "instacart", "fresh market",
		"grocery", "food mart", "food city", "save-a-lot", "iga ", "winco",
		"food4less", "stater bros", "jewel-osco", "acme market", "hannaford",
		"bi-lo", "lucky supermarket", "key food", "foodtown", "fairway market",
		"cub foods",
	}, category: "groceries"},
	{keywords: []string{
		"starbucks", "mcdonald", "chipotle", "chick-fil-a", "subway", "taco bell",
		"wendy", "burger king", "panera", "domino", "pizza hut", "panda express",
		"dunkin", "popeyes", "arby", "kfc ", "sonic drive", "waffle house", "ihop",
		"denny", "applebee", "olive garden", "chili's", "outback", "red lobster",
		"buffalo wild", "five guys", "in-n-out", "jack in the box", "whataburger",
		"raising cane", "wingstop", "zaxby", "grubhub", "doordash", "uber eats",
		"postmates", "restaurant", "cafe ", "bistro", "grill ", "sushi",
		"noodle", "ramen", "diner", "eatery", "tavern", "brew pub",
		"bbq", "steakhouse", "taqueria", "bakery", "bagel", "deli ",
		"smoothie king", "jamba", "tropical smoothie", "jersey mike",
		"jimmy john", "firehouse sub", "potbelly", "noodles & co",
		"sweetgreen", "cava ", "shake shack", "culver", "cook out",
		"del taco", "el pollo", "checkers", "rally's",
		"caribou coffe", "coffee", "papa murphy", "papa john", "pizza luc",
		"pizza ", "tst*", "brewin", "liquor", "krispy kreme",
		"pancheros", "texas roadho", "white castle", "biscuit",
		"dd *doordash", "broken clock",
	}, category: "dining"},
	{keywords: []string{
		"shell oil", "exxon", "chevron", "bp#", "bp gas", " bp ", "mobil gas", "exxonmobil",
		"texaco", "marathon pet", "marathon gas", "sunoco", "speedway", "wawa",
		"quiktrip", "racetrac", "circle k", "pilot flying", "loves travel",
		"murphy usa", "sheetz", "valero", "citgo", "phillips 66", "conoco",
		"sinclair", "casey's", "kwik trip", "kum & go", "getgo",
		"gas station", "petroleum", "gasoline", "holiday station",
	}, category: "gas"},
	{keywords: []string{
		"uber trip", "uber ride", "lyft", "parking", "transit", "metro ",
		"bus fare", "amtrak", "toll ", "ez pass", "ezpass", "ipass",
		"sunpass", "fastrak", "taxi", "cab ", "lime scooter", "bird scooter",
		"car wash", "jiffy lube", "valvoline", "autozone", "advance auto",
		"o'reilly", "napa auto", "pep boys", "tire", "midas", "maaco",
		"car rental", "hertz", "enterprise rent", "avis", "budget rent",
		"onstreet", "auto repair", "auto world",
	}, category: "transportation"},
	{keywords: []string{
		"walmart", "target", "costco", "sam's club", "dollar tree", "dollar general",
		"five below", "big lots", "amazon", "amzn", "bj's wholesale",
		"marshalls", "tj maxx", "ross dress", "burlington", "home depot",
		"homedepot", "lowe's", "lowes", "menards", "ace hardware", "bed bath",
		"ikea", "wayfair", "overstock", "etsy", "ebay", "best buy",
		"apple store", "apple.com", "microsoft store", "staples",
		"office depot", "michaels", "hobby lobby", "joann", "container store",
		"bath & body", "sephora", "ulta", "nordstrom", "macy's", "macys",
		"kohl's", "kohls", "jcpenney", "gap ", "old navy", "banana republic",
		"h&m ", "zara", "uniqlo", "nike", "adidas", "foot locker",
		"dick's sport", "rei ", "kindle", "tobasi", "tobacco",
	}, category: "shopping"},
	{keywords: []string{
		"netflix", "spotify", "hulu", "disney+", "disney plus", "hbo", "apple tv",
		"youtube premium", "paramount", "peacock", "crunchyroll", "audible",
		"apple music", "tidal", "pandora", "amazon prime", "prime video",
		"xbox game pass", "playstation", "nintendo", "adobe", "dropbox",
		"icloud", "google storage", "google one", "microsoft 365", "office 365",
		"nordvpn", "expressvpn", "1password", "lastpass", "bitwarden",
		"duolingo", "headspace", "calm app", "strava", "zwift",
		"patreon", "openai", "chatgpt", "claude.ai", "msbill.info",
		"adtsecurity", "adt security",
	}, category: "subscriptions"},
	{keywords: []string{
		"comcast", "xfinity", "verizon", "at&t", "t-mobile", "sprint", "spectrum",
		"cox comm", "optimum", "frontier comm", "centurylink", "mint mobile",
		"visible wireless", "google fi", "cricket wireless", "boost mobile",
		"metro by t", "straight talk", "us cellular", "consumer cellular",
		"starlink", "hughesnet", "att ", "tmobile",
	}, category: "phone_internet"},
	{keywords: []string{
		"electric", "water bill", "gas bill", "sewer", "utility", "power co",
		"energy co", "national grid", "duke energy", "dominion energy", "southern co",
		"pge", "pg&e", "con edison", "conedison", "eversource", "entergy",
		"exelon", "centerpoint", "consumers energy", "ameren",
		"dte energy", "alliant energy", "xcel energy", "puget sound energy",
		"waste management", "republic services", "trash", "garbage", "recycling",
		"water dept", "sewage", "stormwater", "curbside waste",
		"ecoshield", "pest contr",
	}, category: "utilities"},
	{keywords: []string{
		"cvs", "walgreens", "rite aid", "pharmacy", "doctor", "hospital", "medical",
		"dental", "optom", "urgent care", "labcorp", "quest diag", "kaiser",
		"aetna", "cigna", "united health", "blue cross", "anthem", "copay",
		"clinic", "dermatolog", "orthodont", "chiropract", "physical therap",
		"mental health", "therapist", "counseling", "prescription", "rx ",
		"lenscrafters", "pearle vision", "warby parker", "1-800 contacts",
		"zocdoc", "teladoc", "mdlive", "allinaheal", "allina h",
		"ognomy", "sleep apn",
	}, category: "healthcare"},
	{keywords: []string{
		"planet fitness", "la fitness", "anytime fitness", "orangetheory", "crossfit",
		"ymca", "equinox", "gold's gym", "crunch fitness", "lifetime fitness",
		"peloton", "24 hour fitness", "snap fitness", "blink fitness",
		"f45 training", "pure barre", "corepower", "soulcycle", "barry's",
	}, category: "fitness"},
	{keywords: []string{
		"petsmart", "petco", "chewy", "pet food", "veterinar", "vet clinic",
		"banfield", "metlife pet",
	}, category: "pets"},
	{keywords: []string{
		"geico", "state farm", "progressive", "allstate", "liberty mutual",
		"farmers ins", "usaa", "nationwide", "travelers ins", "erie insurance",
		"american family", "insurance prem", "hartford", "safeco",
		"metlife", "prudential", "mutual of omaha",
	}, category: "insurance"},
	{keywords: []string{
		"tuition", "student loan", "coursera", "udemy", "skillshare",
		"linkedin learning", "college", "university", "school district",
		"navient", "nelnet", "great lakes", "fedloan", "mohela",
		"sallie mae", "sofi student",
	}, category: "education"},
	{keywords: []string{
		"amc theatre", "regal cinema", "cinemark", "fandango",
		"ticketmaster", "stubhub", "seatgeek", "live nation",
		"concert", "theater", "museum", "zoo ", "aquarium",
		"bowling", "topgolf", "theme park", "amusement",
		"dave & buster", "escape room", "trampoline",
		"arcade", "mini golf", "laser tag",
		"comedy", "axs.com", "recreation.gov",
	}, category: "entertainment"},
	{keywords: []string{
		"airbnb", "vrbo", "marriott", "hilton", "hyatt", "ihg",
		"holiday inn", "hampton inn", "best western", "motel",
		"hotel", "resort", "expedia", "booking.com", "priceline",
		"travelocity", "kayak", "hotwire", "trivago",
		"southwest air", "delta air", "american air", "united air",
		"jetblue", "frontier air", "spirit air", "alaska air",
		"allegiant", "airline", "tsa precheck",
		"mount bohemia", "ski ",
	}, category: "travel"},
	{keywords: []string{
		"donation", "charity", "goodwill", "salvation army", "red cross",
		"united way", "habitat for", "march of dimes", "st jude",
		"gofundme", "tithe", "offering", "church", "temple", "mosque",
		"synagogue", "nonprofit",
	}, category: "charity"},
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
