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
	match := re.FindStringSubmatch(req.Msg.GetGoogleSheetUrl())
	var spreadsheetId string
	if len(match) > 1 {
		spreadsheetId = match[1]
	} else {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get spreadsheet id: could not parse spreadsheet id from url"))
	}
	data, err := ds.gsSrv.Spreadsheets.Values.Get(spreadsheetId, "Sheet1!A2:K").Do()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get google sheet data: %s", err.Error()))
	}

	return connect.NewResponse(&datav1.BatchCreateDataResponse{}), nil
}

func mapEnumValues(reqValue string) string {
	switch reqValue {
	case "BR":
		return "BREAST"
	case "CH":
		return "CHEST"
	case "CV":
		return "CARDIOVASCULAR"
	case "GI":
		return "GASTROINTESTINAL"
	case "GU":
		return "GENITOURINARY"
	case "HN":
		return "HEAD_AND_NECK"
	case "MK":
		return "MUSCULOSKELETAL"
	case "NR":
		return "NEURORADIOLOGY"
	case "OB":
		return "OBSTETRIC"
	case "PD":
		return "PEDIATRIC"
	case "HIST":
		return "HISTORY"
	case "FIND":
		return "FINDINGS"
	case "HXF":
		return "HISTORY_AND_FINDINGS"
	default:
		return "UNSPECIFIED"
	}
}
