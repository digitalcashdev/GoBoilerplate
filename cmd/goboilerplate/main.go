package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/DigitalCashDev/goboilerplate"
	"github.com/DigitalCashDev/goboilerplate/internal"
	"github.com/DigitalCashDev/goboilerplate/static"

	"github.com/joho/godotenv"
)

var (
	name = "goboilerplate"
	// these will be replaced by goreleaser
	version = "0.0.0-dev"
	date    = "0001-01-01T00:00:00Z"
	commit  = "0000000"
)

var config goboilerplate.GoBoilerplateConfig

func printVersion() {
	// go run
	fmt.Printf("%s v%s %s (%s)\n", name, version, commit[:7], date)
	fmt.Printf("Copyright (C) 2025 AJ ONeal\n")
	fmt.Printf("Licensed under the MPL-2.0 license\n")
}

func main() {
	var subcmd string
	var envPath string
	var httpPort int

	nArgs := len(os.Args)
	if nArgs >= 2 {
		opt := os.Args[1]
		subcmd = strings.TrimPrefix(opt, "-")
		if opt == "-V" || subcmd == "version" {
			printVersion()
			os.Exit(0)
			return
		}
	}

	{
		envPath = peekOption(os.Args, []string{"--env", "-env"})
		if len(envPath) > 0 {
			fmt.Fprintf(os.Stderr, "reading ENVs from %s", envPath)
			if err := godotenv.Load(envPath); err != nil {
				fmt.Fprintf(os.Stderr, ": skipped (%s)", err.Error())
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	defaultHTTPPort := 8080
	httpPortStr := os.Getenv("PORT")
	if len(httpPortStr) > 0 {
		defaultHTTPPort, _ = strconv.Atoi(httpPortStr)
		if defaultHTTPPort == 0 {
			defaultHTTPPort = 8080
		}
	}

	defaultRPCProtocol := "http"
	rpcProtocol := os.Getenv("DASHD_RPC_PROTOCOL")
	if len(rpcProtocol) == 0 {
		rpcProtocol = defaultRPCProtocol
	}

	rpcHostname := os.Getenv("DASHD_RPC_HOSTNAME")
	if len(rpcHostname) == 0 {
		rpcHostname = "localhost"
	}

	defaultRPCPort := 9998
	rpcPortStr := os.Getenv("DASHD_RPC_PORT")
	if len(rpcPortStr) > 0 {
		defaultRPCPort, _ = strconv.Atoi(rpcPortStr)
		if defaultRPCPort == 0 {
			defaultRPCPort = 8080
		}
	}

	defaultConfigJSONPath := "public-config.json"

	configJSONPath := defaultConfigJSONPath
	overlayFS := &internal.OverlayFS{}
	proxyURL := fmt.Sprintf("%s://%s:%d", rpcProtocol, rpcHostname, defaultRPCPort)
	username := ""
	password := ""
	flag.StringVar(&configJSONPath, "config", defaultConfigJSONPath, "JSON config path, relative to ./static/, ex: ./config.json")
	flag.StringVar(&envPath, "env", "", "load ENVs from file, ex: ./.env")
	flag.StringVar(&proxyURL, "rpc-url", proxyURL, "dashd RPC base URL")
	flag.StringVar(&username, "rpc-username", "", "dashd RPC username")
	flag.StringVar(&password, "rpc-password", "", "dashd RPC password")
	flag.StringVar(&overlayFS.WebRoot, "web-root", "./public/", "serve from the given directory")
	flag.BoolVar(&overlayFS.WebRootOnly, "web-root-only", false, "do not serve the embedded web root")
	flag.IntVar(&httpPort, "port", defaultHTTPPort, "bind and listen for http on this port")
	flag.Parse()
	if subcmd == "help" {
		flag.Usage()
		os.Exit(0)
		return
	}

	overlayFS.LocalFS = http.Dir(overlayFS.WebRoot)
	overlayFS.EmbedFS = http.FS(static.FS)

	f, err := overlayFS.ForceLocalOrEmbedOpen(configJSONPath)
	if err != nil {
		log.Fatalf("loading RPC JSON description file '%s' failed: %v", configJSONPath, err)
	}

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("decoding %s failed: %v", configJSONPath, err)
		return
	}

	if len(username) == 0 {
		username = os.Getenv("DASHD_RPC_USERNAME")
	}
	if len(password) == 0 {
		password = os.Getenv("DASHD_RPC_PASSWORD")
	}

	srv := &goboilerplate.Server{
		RPCConfig: goboilerplate.RPCConfig{
			BaseURL:  proxyURL,
			Username: username,
			Password: password,
		},
		Config: config,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("OPTIONS /", goboilerplate.AddCORSHandler)

	fileServer := http.FileServer(overlayFS)
	mux.Handle("GET /", fileServer)

	mux.HandleFunc("GET /api/version", goboilerplate.CORSMiddleware(versionHandler))
	mux.HandleFunc("GET /api/hello", goboilerplate.CORSMiddleware(helloHandler))
	mux.HandleFunc("GET /api/hello/{name}", goboilerplate.CORSMiddleware(helloHandler))

	limitedGoBoilerplateHandler := goboilerplate.RateLimitMiddleware(srv.GoBoilerplateHandler)
	mux.HandleFunc("GET /api/goboilerplate/config", goboilerplate.CORSMiddleware(limitedGoBoilerplateHandler))
	mux.HandleFunc("GET /api/goboilerplate/config/{key}", goboilerplate.CORSMiddleware(limitedGoBoilerplateHandler))

	mux.HandleFunc("/", MethodNotAllowedHandler)

	// magic := certmagic.NewDefault()
	// myACME := certmagic.NewACMEIssuer(magic, certmagic.DefaultACME)
	// httpRouter := myACME.HTTPChallengeHandler(mux)
	httpRouter := mux

	fmt.Printf("Listening on :%d\n", httpPort)
	addr := fmt.Sprintf(":%d", httpPort)
	log.Fatal(http.ListenAndServe(addr, httpRouter))
}

func peekOption(args, aliases []string) string {
	var flagIndex int

	for _, alias := range aliases {
		flagIndex = slices.Index(args, alias)
		if flagIndex > -1 {
			break
		}
	}

	if flagIndex == -1 {
		return ""
	}

	argIndex := flagIndex + 1
	if len(args) <= argIndex {
		return ""
	}

	return args[argIndex]
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	result := struct {
		Version string `json:"version"`
	}{
		Version: version,
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(result)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if len(name) == 0 {
		name = "World"
	}
	result := struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf("Hello, %s!", name),
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(result)
}

func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
