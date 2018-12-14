package plugins

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/rs/xstats"
	"github.com/rs/xstats/dogstatsd"
)

// DefaultStatMiddleware injects the default stats client on each incoming request's context
func DefaultStatMiddleware(tags ...string) func(http.Handler) http.Handler {
	var statsdWriter io.Writer
	var errWriter error
	statsdWriter, errWriter = net.Dial("udp", "127.0.0.1:8126")
	if errWriter != nil {
		log.Println(errWriter.Error())
		log.Println("stats disabled")
		statsdWriter = ioutil.Discard
	}
	stats := xstats.New(dogstatsd.New(statsdWriter, 10*time.Second))
	return CustomStatMiddleware(stats, tags...)
}

// NopStatMiddleware injects a nop stats client in each incoming request's context
func NopStatMiddleware(tags ...string) func(http.Handler) http.Handler {
	stats := xstats.New(dogstatsd.New(ioutil.Discard, 10*time.Second))
	return CustomStatMiddleware(stats, tags...)
}

// CustomStatMiddleware injects the provided stats client in each incoming request's context
func CustomStatMiddleware(stats xstats.XStater, tags ...string) func(http.Handler) http.Handler {
	return xstats.NewHandler(stats, tags)
}
