package portfolio

import (
	"math"
	"testing"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// MEN in the owner's note: 120 shares at WAC 556.54, then the 20% bonus credits
// 24 shares at no cost, so WAC falls to 66,785 / 144. A later sale leaves WAC alone.
func TestPositionBonusLowersWACAndSellKeepsIt(t *testing.T) {
	txs := []sqlc.Transaction{
		{Kind: kindOf(ntxv1.TransactionKind_TRANSACTION_KIND_BUY), Quantity: 120, Price: 556.54},
		{Kind: kindOf(ntxv1.TransactionKind_TRANSACTION_KIND_BONUS), Quantity: 24, Price: 0},
	}
	qty, wac := position(txs)
	if qty != 144 || math.Abs(wac-463.78) > 0.01 {
		t.Fatalf("after bonus: qty %d wac %.2f, want 144 and 463.78", qty, wac)
	}

	txs = append(
		txs,
		sqlc.Transaction{Kind: kindOf(ntxv1.TransactionKind_TRANSACTION_KIND_SELL), Quantity: 44, Price: 700},
	)
	qty, wac = position(txs)
	if qty != 100 || math.Abs(wac-463.78) > 0.01 {
		t.Fatalf("after sale: qty %d wac %.2f, want 100 and 463.78", qty, wac)
	}
}

func TestPositionWithoutTransactionsIsWatched(t *testing.T) {
	if qty, wac := position(nil); qty != 0 || wac != 0 {
		t.Fatalf("got %d %.2f, want 0 0", qty, wac)
	}
}
