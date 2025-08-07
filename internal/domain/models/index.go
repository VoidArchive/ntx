package models

type Index struct {
	Name        string
	Open        float64
	High        float64
	Low         float64
	Close       float64
	PointChange float64
	IsMain      bool
}

type MarketOverview struct {
	MainIndex   *Index
	SubIndices  []*Index
	LastUpdated string
}
