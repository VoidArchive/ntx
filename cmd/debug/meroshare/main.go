package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/voidarchive/ntx/internal/meroshare"
)

func main() {
	transactions, err := meroshare.ParseTransactions("data/transaction.csv")
	if err != nil {
		log.Fatal(err)
	}

	printSummary(transactions)
	fmt.Println()
	printTransactions(transactions)
}

func printSummary(txs []meroshare.Transaction) {
	typeCounts := make(map[meroshare.TransactionType]int)
	for _, t := range txs {
		typeCounts[t.HistoryDescription.Type]++
	}

	fmt.Println("=== Transaction Summary ===")
	fmt.Printf("Total: %d transactions\n\n", len(txs))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Type\tCount")
	fmt.Fprintln(w, "----\t-----")
	for typ, count := range typeCounts {
		fmt.Fprintf(w, "%s\t%d\n", typ, count)
	}
	w.Flush()
}

func printTransactions(txs []meroshare.Transaction) {
	fmt.Println("=== Recent Transactions ===")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Date\tScrip\tType\tQty\tBalance\tDetails")
	fmt.Fprintln(w, "----\t-----\t----\t---\t-------\t-------")

	limit := min(20, len(txs))
	for _, t := range txs[:limit] {
		details := formatDetails(t.HistoryDescription)
		qty := formatQty(t.CreditQuantity, t.DebitQuantity)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.0f\t%s\n",
			t.TransactionDate.Format("2006-01-02"),
			t.Scrip,
			shortType(t.HistoryDescription.Type),
			qty,
			t.BalanceAfterTransaction,
			details,
		)
	}
	w.Flush()

	if len(txs) > limit {
		fmt.Printf("\n... and %d more transactions\n", len(txs)-limit)
	}
}

func shortType(t meroshare.TransactionType) string {
	switch t {
	case meroshare.TypeBonus:
		return "Bonus"
	case meroshare.TypeMerger:
		return "Merger"
	case meroshare.TypeRights:
		return "Rights"
	case meroshare.TypeRearrangement:
		return "Rearr"
	case meroshare.TypeBuy:
		return "Buy"
	case meroshare.TypeSell:
		return "Sell"
	case meroshare.TypeIPO:
		return "IPO"
	case meroshare.TypeDemat:
		return "Demat"
	default:
		return string(t)
	}
}

func formatQty(credit, debit float64) string {
	if credit > 0 {
		return fmt.Sprintf("+%.0f", credit)
	}
	if debit > 0 {
		return fmt.Sprintf("-%.0f", debit)
	}
	return "0"
}

func formatDetails(h meroshare.HistoryDetails) string {
	switch h.Type {
	case meroshare.TypeBuy, meroshare.TypeSell:
		return fmt.Sprintf("SET:%s", h.SettlementCode)
	case meroshare.TypeBonus:
		return extractRate(h.BonusRate)
	case meroshare.TypeRights:
		return extractRate(h.RightsRate)
	case meroshare.TypeRearrangement:
		if h.PurchaseDate != "" {
			return "bought " + h.PurchaseDate
		}
		return ""
	case meroshare.TypeMerger:
		return ""
	case meroshare.TypeIPO, meroshare.TypeDemat:
		return ""
	default:
		return ""
	}
}

// extractRate pulls percentage from "B-6.5%-2023-24" -> "6.5%"
func extractRate(rate string) string {
	if rate == "" {
		return ""
	}
	// Find the percentage part
	for i := 0; i < len(rate); i++ {
		if rate[i] == '%' {
			// Walk back to find start of number
			start := i - 1
			for start >= 0 && (rate[start] >= '0' && rate[start] <= '9' || rate[start] == '.') {
				start--
			}
			return rate[start+1 : i+1]
		}
	}
	return rate
}
