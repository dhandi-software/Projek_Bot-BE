package provider

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsProvider struct {
	Service *sheets.Service
}

func InitGoogleSheets(ctx context.Context, credentialsFile string) (*SheetsProvider, error) {
	clientOption := option.WithCredentialsFile(credentialsFile)
	srv, err := sheets.NewService(ctx, clientOption)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat koneksi ke Sheets API: %w", err)
	}

	log.Println("Berhasil terhubung ke Google Sheets API!")
	return &SheetsProvider{Service: srv}, nil
}

// AppendData menambahkan satu baris data (array interface{}) ke spreadsheet.
func (p *SheetsProvider) AppendData(spreadsheetId string, sheetRange string, data []interface{}) error {
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, data)

	_, err := p.Service.Spreadsheets.Values.Append(spreadsheetId, sheetRange, &vr).
		ValueInputOption("USER_ENTERED").
		Do()

	if err != nil {
		return fmt.Errorf("gagal insert data ke sheets: %w", err)
	}

	return nil
}
