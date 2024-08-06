package data

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	datav1 "github.com/Lev1ty/lmsysmd/pbi/lmsysmd/load/data/v1"
	"github.com/Lev1ty/lmsysmd/pbi/lmsysmd/load/data/v1/datav1connect"
	modelv1 "github.com/Lev1ty/lmsysmd/pbi/lmsysmd/load/model/v1"
	"github.com/bufbuild/protovalidate-go"
	"github.com/jackc/pgx/v5/pgxpool"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type DataService struct {
	datav1connect.UnimplementedDataServiceHandler
	dbOnce         sync.Once
	db             *pgxpool.Pool
	gOnce          sync.Once
	gsSrv          *sheets.Service
	gdSrv          *drive.Service
	gConf          *jwt.Config
	pvOnce         sync.Once
	protoValidator *protovalidate.Validator
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
	"DD": datav1.Data_CASE_INSTRUCTION_DIFFERENTIAL_DIAGNOSIS,
}

var modelIdEnumMap = map[string]modelv1.ModelId{
	"GPT_3_5_TURBO_0125":         modelv1.ModelId_MODEL_ID_GPT_3_5_TURBO_0125,
	"GPT_4_TURBO_2024_04_09":     modelv1.ModelId_MODEL_ID_GPT_4_TURBO_2024_04_09,
	"GPT_4O_2024_05_13":          modelv1.ModelId_MODEL_ID_GPT_4O_2024_05_13,
	"CLAUDE_3_OPUS_20240229":     modelv1.ModelId_MODEL_ID_CLAUDE_3_OPUS_20240229,
	"GEMINI_1_5_PRO":             modelv1.ModelId_MODEL_ID_GEMINI_1_5_PRO,
	"LLAMA_3_70B":                modelv1.ModelId_MODEL_ID_LLAMA_3_70B,
	"GEMMA_1_2B":                 modelv1.ModelId_MODEL_ID_GEMMA_1_2B,
	"PHI_3_MINI":                 modelv1.ModelId_MODEL_ID_PHI_3_MINI,
	"GEMMA_1_7B":                 modelv1.ModelId_MODEL_ID_GEMMA_1_7B,
	"PHI_3_SMALL":                modelv1.ModelId_MODEL_ID_PHI_3_SMALL,
	"LLAMA_3_8B":                 modelv1.ModelId_MODEL_ID_LLAMA_3_8B,
	"PHI_3_MEDIUM":               modelv1.ModelId_MODEL_ID_PHI_3_MEDIUM,
	"CLAUDE_3_HAIKU_20240307":    modelv1.ModelId_MODEL_ID_CLAUDE_3_HAIKU_20240307,
	"CLAUDE_3_SONNET_20240229":   modelv1.ModelId_MODEL_ID_CLAUDE_3_SONNET_20240229,
	"CLAUDE_3_5_SONNET_20240620": modelv1.ModelId_MODEL_ID_CLAUDE_3_5_SONNET_20240620,
	"YI_LARGE":                   modelv1.ModelId_MODEL_ID_YI_LARGE,
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
	data, err := ds.gsSrv.Spreadsheets.Values.Get(spreadsheetId, spreadsheetRange).Do()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get google sheet data: %w", err))
	}
	// TODO(sunb26): Read data labels once labels have been finalized.

	ds.dbOnce.Do(func() {
		if ds.db, err = pgxpool.New(ctx, os.Getenv("POSTGRES_DSN")); err != nil {
			panic(err)
		}
	})
	ds.pvOnce.Do(func() {
		ds.protoValidator, err = protovalidate.New()
		if err != nil {
			panic(err)
		}
	})
	t := time.Now()
	tx, err := ds.db.Begin(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback(ctx)
	for i, row := range data.Values {
		// Type assert the values from the google sheet
		exp, ok := row[0].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid exp id: id not a string on row %d", i+2))
		}
		expId, err := strconv.ParseUint(exp, 10, 32)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("convert exp id string to uint: %w", err))
		}
		givenModelId, ok := row[1].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid model name: name not a string on row %d", i+2))
		}
		givenCategory, ok := row[2].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid case category: category not a string on row %d", i+2))
		}
		strCaseId, ok := row[3].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid case id: id not a string on row %d", i+2))
		}
		caseId, err := strconv.ParseUint(strCaseId, 10, 32)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("convert case id string to uint: %w", err))
		}
		truth, ok := row[4].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid truth: truth not a string on row %d", i+2))
		}
		prompt, ok := row[5].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid prompt: prompt not a string on row %d", i+2))
		}
		givenCaseInput, ok := row[6].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid case input: input not a string on row %d", i+2))
		}
		inputContent, ok := row[7].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid input content: content not a string on row %d", i+2))
		}
		givenInstruction, ok := row[8].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid instruction: instruction not a string on row %d", i+2))
		}
		results, ok := row[9].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid result: result not a string on row %d", i+2))
		}
		sampleChoices := strings.Split(results, ",")
		modelId, ok := modelIdEnumMap[strings.TrimSpace(givenModelId)]
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid model id: unkown model id on row %d", i+2))
		}
		category, ok := caseCategoryEnumMap[strings.TrimSpace(givenCategory)]
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("mapping to case category enum: unknown category on row %d", i+2))
		}
		caseInputType, ok := caseInputTypeEnumMap[strings.TrimSpace(givenCaseInput)]
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("mapping to case input type enum: unknown input type on row %d", i+2))
		}
		instruction, ok := caseInstructionEnumMap[strings.TrimSpace(givenInstruction)]
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("mapping to case instruction enum: unknown instruction on row %d", i+2))
		}
		// Insert into proto message for validation
		dataMsg := &datav1.Data{
			ExperimentId:     uint32(expId),
			ModelId:          modelId,
			CaseCategory:     category,
			CaseId:           uint32(caseId),
			GroundTruth:      truth,
			Prompt:           prompt,
			CaseInputType:    caseInputType,
			CaseInputContent: inputContent,
			CaseInstruction:  instruction,
			Results:          sampleChoices,
		}
		log.Printf("Row Proto Message: %v", dataMsg)
		if err := ds.protoValidator.Validate(dataMsg); err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("row %d: validate data row proto message: %w", i+2, err))
		}

		// Insert into database

		// 1. Create Caseset - return id
		if _, err := tx.Exec(ctx, "INSERT INTO casesets (id, create_time) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING", dataMsg.CaseId, t); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create caseset: %w", err))
		}

		// 2. Create Case
		var cid uint32
		content := map[string]interface{}{"messages": []map[string]string{{"role": "user", "content": dataMsg.CaseInputContent}}}
		if err := tx.QueryRow(ctx, "INSERT INTO cases (caseset_id, content, create_time, truth) VALUES ($1, $2, $3, $4) RETURNING id", dataMsg.CaseId, content, t, dataMsg.GroundTruth).Scan(&cid); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create case for caseset %d: %w", dataMsg.CaseId, err))
		}
	}

	return connect.NewResponse(&datav1.BatchCreateDataResponse{}), nil
}

func parseSpreadsheetId(url string) (string, error) {
	match := re.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("could not parse spreadsheet id from url")
}
