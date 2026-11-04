package nepse

import (
	"context"
	"fmt"
	"net/url"
)

// shortcut: the download endpoint is NEPSE's site convention, not exposed by go-nepse and not verified
// from this machine (the server drops non-browser clients); confirm by opening one link, fix here if it moved.
const filingBaseURL = "https://www.nepalstock.com.np/api/nots/security/fetchFiles?fileLocation="

type Fundamental struct {
	CompanyID     int64
	FiscalYear    string
	Quarter       string // empty for annual reports
	EPS           float64
	PERatio       float64
	BookValue     float64
	PaidUpCapital float64
	ProfitAmount  float64
	PublishedAt   string // submitted date of the filing, "" when NEPSE has no document
	SourceURL     string
	FiscalYearNp  string // FYNameNepali verbatim, e.g. "2080-2081"
}

// filingURL turns NEPSE's relative filePath into a link a reader can open.
func filingURL(filePath string) string {
	if filePath == "" {
		return ""
	}
	return filingBaseURL + url.QueryEscape(filePath)
}

func (c *Client) Fundamentals(ctx context.Context, securityID int32) ([]Fundamental, error) {
	reports, err := c.api.Reports(ctx, securityID)
	if err != nil {
		return nil, fmt.Errorf("fetch reports: %w", err)
	}

	var fundamentals []Fundamental
	for _, r := range reports {
		if r.FiscalReport == nil {
			continue
		}
		fr := r.FiscalReport

		var fiscalYear, fiscalYearNp string
		if fr.FinancialYear != nil {
			fiscalYear = fr.FinancialYear.FYName
			fiscalYearNp = fr.FinancialYear.FYNameNepali
		}

		// The first document is the filing the figures came from; later ones are rare re-uploads.
		var publishedAt, sourceURL string
		if len(r.ApplicationDocumentDetailsList) > 0 {
			doc := r.ApplicationDocumentDetailsList[0]
			publishedAt = doc.SubmittedDate
			sourceURL = filingURL(doc.FilePath)
		}

		fundamentals = append(fundamentals, Fundamental{
			CompanyID:     int64(r.ID),
			FiscalYear:    fiscalYear,
			Quarter:       r.QuarterName(),
			EPS:           fr.EPSValue,
			PERatio:       fr.PEValue,
			BookValue:     fr.NetWorthPerShare,
			PaidUpCapital: fr.PaidUpCapital,
			ProfitAmount:  fr.ProfitAmount,
			PublishedAt:   publishedAt,
			SourceURL:     sourceURL,
			FiscalYearNp:  fiscalYearNp,
		})
	}
	return fundamentals, nil
}
