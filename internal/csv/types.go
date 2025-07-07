package csv

// CSV field indices using iota for better maintainability
const (
	CSVFieldSerial = iota
	CSVFieldScrip
	CSVFieldDate
	CSVFieldCredit
	CSVFieldDebit
	CSVFieldBalance
	CSVFieldDescription
	CSVMinFields
)

// TransactionType represents the type of transaction using iota for better type safety
type TransactionType int

const (
	TransactionTypeIPO TransactionType = iota
	TransactionTypeBonus
	TransactionTypeRights
	TransactionTypeMerger
	TransactionTypeRearrangement
	TransactionTypeRegular
)

// String returns the string representation of TransactionType
func (tt TransactionType) String() string {
	switch tt {
	case TransactionTypeIPO:
		return "IPO"
	case TransactionTypeBonus:
		return "BONUS"
	case TransactionTypeRights:
		return "RIGHTS"
	case TransactionTypeMerger:
		return "MERGER"
	case TransactionTypeRearrangement:
		return "REARRANGEMENT"
	case TransactionTypeRegular:
		return "REGULAR"
	default:
		return "UNKNOWN"
	}
}