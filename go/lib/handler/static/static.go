package static

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/Lev1ty/lmsysmd/lib/middleware/static"
	"github.com/klauspost/compress/gzhttp"
)

func Handler(efs embed.FS, pre string) http.Handler {
	ffs, err := fs.Sub(efs, pre)
	if err != nil {
		log.Fatal(err)
	}
	return static.Middleware(gzhttp.GzipHandler(http.FileServerFS(ffs)))
}
