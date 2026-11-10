package portfolio

import (
	"strings"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// kindOf is the stored form of a transaction kind: the enum name's last word, lower-cased.
func kindOf(k ntxv1.TransactionKind) string {
	return strings.ToLower(strings.TrimPrefix(k.String(), "TRANSACTION_KIND_"))
}

func kindFrom(s string) ntxv1.TransactionKind {
	return ntxv1.TransactionKind(ntxv1.TransactionKind_value["TRANSACTION_KIND_"+strings.ToUpper(s)])
}

// position derives quantity and weighted average cost as TMS does: every credit
// (buy, bonus, rights, IPO) adds its cost to the book, so a zero-price bonus lowers
// WAC; a sale removes shares at WAC and leaves it unchanged.
func position(txs []sqlc.Transaction) (qty int64, wac float64) {
	var cost float64
	for _, tx := range txs {
		if tx.Kind == kindOf(ntxv1.TransactionKind_TRANSACTION_KIND_SELL) {
			cost -= float64(tx.Quantity) * wac
			qty -= tx.Quantity
			continue
		}
		cost += float64(tx.Quantity) * tx.Price
		qty += tx.Quantity
		if qty > 0 {
			wac = cost / float64(qty)
		}
	}
	if qty <= 0 {
		return 0, 0
	}
	return qty, wac
}
