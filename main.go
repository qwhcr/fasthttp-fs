// package main

// import (
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"log"

// 	"github.com/valyala/fasthttp"
// )

// var (
// 	addr     = flag.String("addr", ":8080", "TCP address to listen to")
// 	compress = flag.Bool("compress", false, "Whether to enable transparent response compression")
// )

// func main() {
// 	flag.Parse()

// 	h := requestHandler
// 	if *compress {
// 		h = fasthttp.CompressHandler(h)
// 	}

// 	if err := fasthttp.ListenAndServe(*addr, h); err != nil {
// 		log.Fatalf("Error in ListenAndServe: %s", err)
// 	}
// }

// func requestHandler(ctx *fasthttp.RequestCtx) {
// 	fmt.Fprintf(ctx, "Hello, world!\n\n")

// 	fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
// 	fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
// 	fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
// 	fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
// 	fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
// 	fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
// 	fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
// 	fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
// 	fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
// 	fmt.Fprintf(ctx, "Your ip is %q\n\n", ctx.RemoteIP())
// 	fmt.Fprintf(ctx, "Headers %q", ctx.Request.Header.Peek("Authorization"))
// 	fmt.Fprintf(ctx, "Post Args is: %q", ctx.PostBody())

// 	var form map[string]interface{}
// 	if err := json.Unmarshal(ctx.PostBody(), &form); err != nil {
// 		panic(err)
// 	}

// 	fmt.Fprintf(ctx, "Json request is %q", form)

// 	fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)
// 	ctx.SetContentType("text/plain; charset=utf8")

// 	// Set arbitrary headers
// 	ctx.Response.Header.Set("X-My-Header", "my-header-value")

// 	// Set cookies
// 	var c fasthttp.Cookie
// 	c.SetKey("cookie-name")
// 	c.SetValue("cookie-value")
// 	ctx.Response.Header.SetCookie(&c)
// }

// Example static file server.
//
// Serves static files from the given directory.
// Exports various stats at /stats .
package main

import (
	"expvar"
	"flag"
	"log"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/expvarhandler"
)

var (
	addr               = flag.String("addr", "localhost:8080", "TCP address to listen to")
	addrTLS            = flag.String("addrTLS", "", "TCP address to listen to TLS (aka SSL or HTTPS) requests. Leave empty for disabling TLS")
	byteRange          = flag.Bool("byteRange", false, "Enables byte range requests if set to true")
	certFile           = flag.String("certFile", "./ssl-cert-snakeoil.pem", "Path to TLS certificate file")
	compress           = flag.Bool("compress", false, "Enables transparent response compression if set to true")
	dir                = flag.String("dir", "./", "Directory to serve static files from")
	generateIndexPages = flag.Bool("generateIndexPages", true, "Whether to generate directory index pages")
	keyFile            = flag.String("keyFile", "./ssl-cert-snakeoil.key", "Path to TLS key file")
	vhost              = flag.Bool("vhost", false, "Enables virtual hosting by prepending the requested path with the requested hostname")
)

func main() {
	// Parse command-line flags.
	flag.Parse()

	// Setup FS handler
	fs := &fasthttp.FS{
		Root:               *dir,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: *generateIndexPages,
		Compress:           *compress,
		AcceptByteRange:    *byteRange,
	}
	if *vhost {
		fs.PathRewrite = fasthttp.NewVHostPathRewriter(0)
	}
	fsHandler := fs.NewRequestHandler()

	// Create RequestHandler serving server stats on /stats and files
	// on other requested paths.
	// /stats output may be filtered using regexps. For example:
	//
	//   * /stats?r=fs will show only stats (expvars) containing 'fs'
	//     in their names.
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/stats":
			expvarhandler.ExpvarHandler(ctx)
		case "/status-app":
			fsHandler(ctx)
		default:
			log.Fatal(string(ctx.Path()))
			log.Fatalf("Miss")
			// fsHandler(ctx)
			// updateFSCounters(ctx)
		}
	}

	// Start HTTP server.
	if len(*addr) > 0 {
		log.Printf("Starting HTTP server on %q", *addr)
		go func() {
			if err := fasthttp.ListenAndServe(*addr, requestHandler); err != nil {
				log.Fatalf("error in ListenAndServe: %s", err)
			}
		}()
	}

	// Start HTTPS server.
	if len(*addrTLS) > 0 {
		log.Printf("Starting HTTPS server on %q", *addrTLS)
		go func() {
			if err := fasthttp.ListenAndServeTLS(*addrTLS, *certFile, *keyFile, requestHandler); err != nil {
				log.Fatalf("error in ListenAndServeTLS: %s", err)
			}
		}()
	}

	log.Printf("Serving files from directory %q", *dir)
	log.Printf("See stats at http://%s/stats", *addr)

	// Wait forever.
	select {}
}

func updateFSCounters(ctx *fasthttp.RequestCtx) {
	// Increment the number of fsHandler calls.
	fsCalls.Add(1)

	// Update other stats counters
	resp := &ctx.Response
	switch resp.StatusCode() {
	case fasthttp.StatusOK:
		fsOKResponses.Add(1)
		fsResponseBodyBytes.Add(int64(resp.Header.ContentLength()))
	case fasthttp.StatusNotModified:
		fsNotModifiedResponses.Add(1)
	case fasthttp.StatusNotFound:
		fsNotFoundResponses.Add(1)
	default:
		fsOtherResponses.Add(1)
	}
}

// Various counters - see https://golang.org/pkg/expvar/ for details.
var (
	// Counter for total number of fs calls
	fsCalls = expvar.NewInt("fsCalls")

	// Counters for various response status codes
	fsOKResponses          = expvar.NewInt("fsOKResponses")
	fsNotModifiedResponses = expvar.NewInt("fsNotModifiedResponses")
	fsNotFoundResponses    = expvar.NewInt("fsNotFoundResponses")
	fsOtherResponses       = expvar.NewInt("fsOtherResponses")

	// Total size in bytes for OK response bodies served.
	fsResponseBodyBytes = expvar.NewInt("fsResponseBodyBytes")
)
