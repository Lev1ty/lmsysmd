package sample

import (
	"context"
	"fmt"
	"os"
	"sync"

	"connectrpc.com/connect"
	samplev1 "github.com/Lev1ty/lmsysmd/pbi/lmsysmd/sample/v1"
	"github.com/Lev1ty/lmsysmd/pbi/lmsysmd/sample/v1/samplev1connect"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SampleService struct {
	samplev1connect.UnimplementedSampleServiceHandler
	dbOnce sync.Once
	db     *pgxpool.Pool
}

func (ss *SampleService) GetSample(
	ctx context.Context,
	req *connect.Request[samplev1.GetSampleRequest],
) (*connect.Response[samplev1.GetSampleResponse], error) {
	ss.dbOnce.Do(func() {
		var err error
		if ss.db, err = pgxpool.New(ctx, os.Getenv("POSTGRES_DSN")); err != nil {
			panic(err)
		}
	})
	sam := &samplev1.Sample{}
	if err := ss.db.QueryRow(ctx, "SELECT id, content, truth FROM samples WHERE id = $1", req.Msg.GetSampleId()).Scan(&sam.SampleId, &sam.Content, &sam.Truth); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get sample %d: %w", req.Msg.GetSampleId(), err))
	}
	rs, err := ss.db.Query(ctx, "SELECT id, content FROM sample_choices WHERE sample_id = $1", req.Msg.GetSampleId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list sample choices %d: %w", req.Msg.GetSampleId(), err))
	}
	defer rs.Close()
	for rs.Next() {
		var id uint32
		var content string
		if err := rs.Scan(&id, &content); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("scan sample choice %d: %w", req.Msg.GetSampleId(), err))
		}
		c := &samplev1.Sample_Choice{ChoiceId: id}
		common, ok := samplev1.Sample_Choice_ContentCommon_value[content]
		if ok {
			c.Content = &samplev1.Sample_Choice_Common{Common: samplev1.Sample_Choice_ContentCommon(common)}
		} else {
			c.Content = &samplev1.Sample_Choice_Specific{Specific: content}
		}
		sam.Choices = append(sam.Choices, c)
	}
	return connect.NewResponse(&samplev1.GetSampleResponse{Sample: sam}), nil
}
