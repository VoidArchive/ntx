package sharesansar

// CSS selectors and column mappings based on test results
const (
	// Page URLs
	URLLiveTrading   = "https://www.sharesansar.com/live-trading"
	URLCompanyDetail = "https://www.sharesansar.com/company/%s" // %s = lowercase symbol
	URLFloorsheet    = "https://www.sharesansar.com/floorsheet"
	URLToday         = "https://www.sharesansar.com/today-share-price"

	// Table selectors - confirmed working from test
	SelectorLiveTable    = "table#headFixed tbody tr"
	SelectorCompanyTable = "div.company-details tr"
	SelectorFloorsheet   = "table.table tbody tr"

	// Alternative selectors as fallback
	SelectorTableGeneric = "table.table tbody tr"
	SelectorTableHover   = "table.table-hover tbody tr"
	SelectorTableStriped = "table.table-striped tbody tr"

	// Rate limiting
	RequestDelay      = 15 // seconds between requests
	MaxRetries        = 3
	CircuitBreakerMax = 5 // failures before circuit opens
)

// Column positions in live trading table (0-indexed)
// Headers: [S.No Symbol LTP Point Change % Change Open High Low Volume Prev. Close]
const (
	ColSNo       = 0  // Serial Number
	ColSymbol    = 1  // Symbol with link
	ColLTP       = 2  // Last Traded Price
	ColChange    = 3  // Point Change
	ColChangePct = 4  // Percentage Change
	ColOpen      = 5  // Opening Price
	ColHigh      = 6  // Day High
	ColLow       = 7  // Day Low
	ColVolume    = 8  // Volume
	ColPrevious  = 9  // Previous Close
	ColTurnover  = 10 // Turnover (might not be in header but in data)
)

// CompanyDetailFields maps the field names from company page
var CompanyDetailFields = map[string]string{
	"Symbol":                "symbol",
	"Name":                  "name",
	"Sector":                "sector",
	"Operation Date":        "operation_date",
	"Listed Shares":         "listed_shares",
	"Paid Up":               "paid_up_value",
	"Total Paid Up Value":   "total_paid_up",
	"Book Close Date":       "book_close_date",
	"Pivot Point (PP)":      "pivot_point",
	"Support level (S1)":    "support_1",
	"Support level (S2)":    "support_2",
	"Support level (S3)":    "support_3",
	"Resistance level (R1)": "resistance_1",
	"Resistance level (R2)": "resistance_2",
	"Resistance level (R3)": "resistance_3",
	"MA5":                   "ma_5",
	"MA20":                  "ma_20",
	"MA180":                 "ma_180",
	"Signal":                "signal",
	"Bonus Share":           "bonus_share",
	"Cash Dividend":         "cash_dividend",
	"Phone Number":          "phone",
	"Email":                 "email",
	"Address":               "address",
}

// User agents for rotation
var UserAgents = []string{
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
}

// Headers for all requests
var DefaultHeaders = map[string]string{
	"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"Accept-Language":           "en-US,en;q=0.9",
	"Accept-Encoding":           "gzip, deflate, br",
	"Connection":                "keep-alive",
	"Upgrade-Insecure-Requests": "1",
	"Cache-Control":             "max-age=0",
	"Sec-Fetch-Dest":            "document",
	"Sec-Fetch-Mode":            "navigate",
	"Sec-Fetch-Site":            "none",
}
