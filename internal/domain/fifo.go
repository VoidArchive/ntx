package domain

import (
	"fmt"
	"time"
)

type FIFOQueue struct {
	StockSymbol string
	Lots        []Lot
}

func NewFIFOQueue(stockSymbol string) *FIFOQueue {
	return &FIFOQueue{
		stockSymbol,
		make([]Lot, 0),
	}
}

func (fq *FIFOQueue) Buy(quantity int, price Money, date time.Time) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive, got %d", quantity)
	}
	if price.IsNegative() {
		return fmt.Errorf("price cannot be negative, got %s", price.String())
	}
	lot := Lot{
		quantity,
		price,
		date,
	}
	fq.Lots = append(fq.Lots, lot)
	return nil
}

func (fq *FIFOQueue) Sell(quantity int, salePrice Money, saleDate time.Time) (*SaleResult, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("sell quantity must be positive, got %d", quantity)
	}
	if salePrice.IsNegative() {
		return nil, fmt.Errorf("sale price cannot be negative, got %s", salePrice.String())
	}
	totalShares := fq.TotalShares()
	if quantity > totalShares {
		return nil, fmt.Errorf("cannot sell %d shares, only %d available", quantity, totalShares)
	}

	result := &SaleResult{
		make([]RealizedGain, 0),
		quantity,
		Zero(),
		Zero(),
		Zero(),
	}
	remainingToSell := quantity
	lotIndex := 0

	for remainingToSell > 0 && lotIndex < len(fq.Lots) {
		lot := &fq.Lots[lotIndex]

		sharesToSellFromLot := min(remainingToSell, lot.Quantity)

		// NOTE: Holding period excludes purchase date (starts counting next day)
		holdingDays := int(saleDate.Sub(lot.Date)/(24*time.Hour)) - 1
		isLongTerm := holdingDays > 365

		costBasisPerShare := lot.Price
		totalCostBasis := costBasisPerShare.Multiply(sharesToSellFromLot)
		totalProceeds := salePrice.Multiply(sharesToSellFromLot)
		gainLoss := totalProceeds.Subtract(totalCostBasis)

		realizedGain := RealizedGain{
			fq.StockSymbol,
			saleDate,
			sharesToSellFromLot,
			salePrice,
			costBasisPerShare,
			gainLoss,
			holdingDays,
			isLongTerm,
		}

		result.RealizedGains = append(result.RealizedGains, realizedGain)
		result.TotalGainLoss = result.TotalGainLoss.Add(gainLoss)
		result.TotalProceeds = result.TotalProceeds.Add(totalProceeds)
		result.TotalCostBasis = result.TotalCostBasis.Add(totalCostBasis)

		lot.Quantity -= sharesToSellFromLot
		remainingToSell -= sharesToSellFromLot

		if lot.Quantity == 0 {
			lotIndex++
		}
	}
	fq.removeExhaustedLots()
	return result, nil
}

func (fq *FIFOQueue) removeExhaustedLots() {
	activeLots := make([]Lot, 0, len(fq.Lots))
	for _, lot := range fq.Lots {
		if lot.Quantity > 0 {
			activeLots = append(activeLots, lot)
		}
	}
	fq.Lots = activeLots
}

func (fq *FIFOQueue) TotalShares() int {
	total := 0
	for _, lot := range fq.Lots {
		total += lot.Quantity
	}
	return total
}

func (fq *FIFOQueue) TotalCost() Money {
	total := Zero()
	for _, lot := range fq.Lots {
		total = total.Add(lot.Price.Multiply(lot.Quantity))
	}
	return total
}

func (fq *FIFOQueue) WeightedAverageCost() Money {
	totalShares := fq.TotalShares()
	if totalShares == 0 {
		return Zero()
	}
	return fq.TotalCost().Divide(totalShares)
}

func (fq *FIFOQueue) IsEmpty() bool {
	return fq.TotalShares() == 0
}

func (fq *FIFOQueue) GetLots() []Lot {
	lotsCopy := make([]Lot, len(fq.Lots))
	copy(lotsCopy, fq.Lots)
	return lotsCopy
}

func (fq *FIFOQueue) GetHolding(currentPrice Money) Holding {
	totalShares := fq.TotalShares()
	totalCost := fq.TotalCost()
	marketValue := currentPrice.Multiply(totalShares)
	unrealizedGainLoss := marketValue.Subtract(totalCost)

	var unrealizedGainPct float64
	if !totalCost.IsZero() {
		unrealizedGainPct = totalCost.PercentageChange(marketValue)
	}

	return Holding{
		fq.StockSymbol,
		totalShares,
		fq.WeightedAverageCost(),
		totalCost,
		currentPrice,
		marketValue,
		unrealizedGainLoss,
		unrealizedGainPct,
		time.Now(),
	}
}
