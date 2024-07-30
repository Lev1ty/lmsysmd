package data

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"

	"connectrpc.com/connect"
	datav1 "github.com/Lev1ty/lmsysmd/pbi/lmsysmd/load/data/v1"
	"github.com/Lev1ty/lmsysmd/pbi/lmsysmd/load/data/v1/datav1connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type DataService struct {
	datav1connect.UnimplementedDataServiceHandler
	dbOnce sync.Once
	db     *pgxpool.Pool
	gOnce  sync.Once
	gsSrv  *sheets.Service
	gdSrv  *drive.Service
	gConf  *jwt.Config
}

var re = regexp.MustCompile(`https:\/\/docs\.google\.com\/spreadsheets\/d\/([a-zA-Z0-9_-]+)\/`)

var caseCategoryEnumMap = map[string]datav1.Data_CaseCategory{
	"BR": datav1.Data_CASE_CATEGORY_BREAST,
	"CH": datav1.Data_CASE_CATEGORY_CHEST,
	"CV": datav1.Data_CASE_CATEGORY_CARDIOVASCULAR,
	"GI": datav1.Data_CASE_CATEGORY_GASTROINTESTINAL,
	"GU": datav1.Data_CASE_CATEGORY_GENITOURINARY,
	"HN": datav1.Data_CASE_CATEGORY_HEAD_AND_NECK,
	"MK": datav1.Data_CASE_CATEGORY_MUSCULOSKELETAL,
	"NR": datav1.Data_CASE_CATEGORY_NEURORADIOLOGY,
	"OB": datav1.Data_CASE_CATEGORY_OBSTETRIC,
	"PD": datav1.Data_CASE_CATEGORY_PEDIATRIC,
}

var caseInputTypeEnumMap = map[string]datav1.Data_CaseInputType{
	"HIST": datav1.Data_CASE_INPUT_TYPE_HISTORY,
	"FIND": datav1.Data_CASE_INPUT_TYPE_FINDINGS,
	"HXF":  datav1.Data_CASE_INPUT_TYPE_HISTORY_AND_FINDINGS,
}

var caseInstructionEnumMap = map[string]datav1.Data_CaseInstruction{
	"DIFFERENTIAL_DIAGNOSIS": datav1.Data_CASE_INSTRUCTION_DIFFERENTIAL_DIAGNOSIS,
}

// This variable defines the range of the spreadsheet to be read. 'A2' assumes the csv has headers in the first row. 'K' is the last column
// and should be adjusted accordingly should a column be added or removed.
const spreadsheetRange = "Sheet1!A2:K"

func (ds *DataService) BatchCreateData(
	ctx context.Context,
	req *connect.Request[datav1.BatchCreateDataRequest],
) (*connect.Response[datav1.BatchCreateDataResponse], error) {
	ds.gOnce.Do(func() {
		ds.gConf = &jwt.Config{
			Email:      os.Getenv("GOOGLE_SERVICE_ACCOUNT"),
			PrivateKey: []byte(os.Getenv("GOOGLE_SERVICE_ACCOUNT_PRIVATE_KEY")),
			Scopes: []string{
				"https://www.googleapis.com/auth/spreadsheets.readonly",
				"https://www.googleapis.com/auth/drive.readonly",
			},
			TokenURL: google.JWTTokenURL,
		}
		var err error
		if ds.gsSrv, err = sheets.NewService(ctx, option.WithHTTPClient(ds.gConf.Client(ctx))); err != nil {
			panic(err)
		}
		if ds.gdSrv, err = drive.NewService(ctx, option.WithHTTPClient(ds.gConf.Client(ctx))); err != nil {
			panic(err)
		}
	})
	spreadsheetId, err := parseSpreadsheetId(req.Msg.GetGoogleSheetUrl())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get spreadsheet id: %w", err))
	}
	_, err = ds.gsSrv.Spreadsheets.Values.Get(spreadsheetId, spreadsheetRange).Do()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get google sheet data: %w", err))
	}

	// TODO(sunb26): Implement parse row data and insert into database accordingly

	return connect.NewResponse(&datav1.BatchCreateDataResponse{}), nil
}

func parseSpreadsheetId(url string) (string, error) {
	match := re.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("could not parse spreadsheet id from url")
}
