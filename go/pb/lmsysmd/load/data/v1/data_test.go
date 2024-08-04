package data

import (
	"testing"
)

func TestParseSpreadsheetId_ValidSheetsUrl_Success(t *testing.T) {
	url := "https://docs.google.com/spreadsheets/d/1MWVXNHXgLdTHHTVmpi9HmQ3DRV6uU3CWfHdxshUeEUU/edit?gid=0#gid=0"
	res, err := parseSpreadsheetId(url)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if res != "1MWVXNHXgLdTHHTVmpi9HmQ3DRV6uU3CWfHdxshUeEUU" {
		t.Errorf("Expected file id 1MWVXNHXgLdTHHTVmpi9HmQ3DRV6uU3CWfHdxshUeEUU, got: %v", res)
	}
}
func TestParseSpreadsheetId_Google_Docs_Url_Fail(t *testing.T) {
	url := "https://docs.google.com/document/d/1FZp1kPiByO7GlG77mpPDw_aCxp71_SJSPE2W7d1kQ6c/edit?pli=1#heading=h.v7p86b6itged"
	_, err := parseSpreadsheetId(url)
	if err.Error() != "could not parse spreadsheet id from url" {
		t.Errorf("Expected error: could not parse spreadsheet id from url, got: %v", err)
	}
}
