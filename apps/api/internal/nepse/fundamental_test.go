package nepse

import "testing"

func TestFilingURL(t *testing.T) {
	tests := []struct{ path, want string }{
		{"", ""},
		{
			"nabil_user/2026-11-02/Nabil Bank Q1 report_1793628713961.pdf",
			"https://www.nepalstock.com.np/api/nots/security/fetchFiles?fileLocation=nabil_user%2F2026-11-02%2FNabil+Bank+Q1+report_1793628713961.pdf",
		},
	}
	for _, tt := range tests {
		if got := filingURL(tt.path); got != tt.want {
			t.Errorf("filingURL(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
