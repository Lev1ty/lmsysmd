package rating

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/Lev1ty/lmsysmd/lib/context/value"
	ratingv1 "github.com/Lev1ty/lmsysmd/pbi/lmsysmd/rating/v1"
	"github.com/Lev1ty/lmsysmd/pbi/lmsysmd/rating/v1/ratingv1connect"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RatingService struct {
	ratingv1connect.UnimplementedRatingServiceHandler
	dbOnce sync.Once
	db     *pgxpool.Pool
}

func (rs *RatingService) CreateRating(
	ctx context.Context,
	req *connect.Request[ratingv1.CreateRatingRequest],
) (*connect.Response[ratingv1.CreateRatingResponse], error) {
	sc, ok := ctx.Value(value.ClerkSessionClaims).(*clerk.SessionClaims)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("clerk session claims not found"))
	}
	rs.dbOnce.Do(func() {
		var err error
		if rs.db, err = pgxpool.New(ctx, os.Getenv("POSTGRES_DSN")); err != nil {
			panic(err)
		}
	})
	t := time.Now()
	tx, err := rs.db.Begin(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("begin tx: %w", err))
	}
	defer tx.Rollback(ctx)
	var rid uint32
	if req.Msg.GetRating().GetRatingId() != 0 {
		if err := tx.QueryRow(ctx, "INSERT INTO ratings (user_id, id, sample_id, choice_id, create_time) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, id) DO NOTHING RETURNING id", sc.Subject, req.Msg.GetRating().GetRatingId(), req.Msg.GetRating().GetSampleId(), req.Msg.GetRating().GetChoiceId(), t).Scan(&rid); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				rid = req.Msg.GetRating().GetRatingId()
			} else {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create or get rating %d: %w", req.Msg.GetRating().GetRatingId(), err))
			}
		}
	} else if err := tx.QueryRow(ctx, "INSERT INTO ratings (user_id, sample_id, choice_id, create_time) VALUES ($1, $2, $3, $4) RETURNING id", sc.Subject, req.Msg.GetRating().GetSampleId(), req.Msg.GetRating().GetChoiceId(), t).Scan(&rid); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create rating for sample %d: %w", req.Msg.GetRating().GetSampleId(), err))
	}
	if _, err := tx.Exec(ctx, "INSERT INTO rating_states (user_id, rating_id, state, create_time) VALUES ($1, $2, $3, $4)", sc.Subject, rid, req.Msg.GetState().GetState(), t); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create rating state for sample %d: %w", req.Msg.GetRating().GetSampleId(), err))
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("commit tx: %w", err))
	}
	return connect.NewResponse(&ratingv1.CreateRatingResponse{RatingId: rid}), nil
}
