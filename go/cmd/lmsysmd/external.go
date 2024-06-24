package main

import (
	"embed"
	"net/http"

	"connectrpc.com/connect"
	"github.com/Lev1ty/lmsysmd/cmd/lmsysmd/handler/rating"
	"github.com/Lev1ty/lmsysmd/lib/handler/static"
	"github.com/Lev1ty/lmsysmd/lib/middleware/buf/validate"
	"github.com/Lev1ty/lmsysmd/lib/middleware/clerk"
	pbrating "github.com/Lev1ty/lmsysmd/pb/lmsysmd/rating/v1"
	"github.com/Lev1ty/lmsysmd/pb/lmsysmd/sample/v1"
	"github.com/Lev1ty/lmsysmd/pbi/lmsysmd/rating/v1/ratingv1connect"
	"github.com/Lev1ty/lmsysmd/pbi/lmsysmd/sample/v1/samplev1connect"
	"github.com/ulule/limiter/v3"
)

//go:embed all:ts all:ts/_next
var ts embed.FS

func external(mux *http.ServeMux, rls limiter.Store) http.Handler {
	mux.Handle("/", clerk.Middleware{Configs: []clerk.Config{
		{Includes: []string{"/rating"}},
	}}.Handler(static.Handler(ts, "ts")))
	mux.Handle(rating.PatternAndHandler(func(h http.Handler) http.Handler {
		return clerk.Middleware{Configs: []clerk.Config{{Includes: []string{"/"}}}}.Handler(h)
	}))
	mux.Handle(clerk.WithHeaderAuthorization(ratingv1connect.NewRatingServiceHandler(&pbrating.RatingService{}, connect.WithInterceptors(
		clerk.Middleware{Configs: []clerk.Config{{Includes: []string{"/"}}}},
		validate.Middleware{},
	))))
	mux.Handle(clerk.WithHeaderAuthorization(samplev1connect.NewSampleServiceHandler(&sample.SampleService{}, connect.WithInterceptors(
		clerk.Middleware{Configs: []clerk.Config{{Includes: []string{"/"}}}},
		validate.Middleware{},
	))))
	return mux
}
