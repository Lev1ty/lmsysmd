package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dotenv-org/godotenvvault"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	if err := godotenvvault.Load(os.Getenv("ENV")); err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(
		os.Getenv("EXTERNAL_API_ADDR"),
		h2c.NewHandler(external(http.NewServeMux(), memory.NewStore()), &http2.Server{}),
	))
}
