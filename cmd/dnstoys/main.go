package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/knadh/dns.toys/internal/fx"
	"github.com/knadh/dns.toys/internal/geo"
	"github.com/knadh/dns.toys/internal/timezones"
	"github.com/knadh/dns.toys/internal/weather"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/miekg/dns"
	flag "github.com/spf13/pflag"
)

var (
	lo = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	ko = koanf.New(".")

	// Version of the build injected at build time.
	buildString = "unknown"
)

func initConfig() {
	// Register --help handler.
	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.StringSlice("config", []string{"config.toml"}, "path to one or more TOML config files to load in order")
	f.Bool("version", false, "show build version")
	f.Parse(os.Args[1:])

	// Display version.
	if ok, _ := f.GetBool("version"); ok {
		fmt.Println(buildString)
		os.Exit(0)
	}

	// Read the config files.
	cFiles, _ := f.GetStringSlice("config")
	for _, f := range cFiles {
		lo.Printf("reading config: %s", f)
		if err := ko.Load(file.Provider(f), toml.Parser()); err != nil {
			lo.Printf("error reading config: %v", err)
		}
	}

	ko.Load(posflag.Provider(f, ".", ko), nil)
}

func main() {
	initConfig()

	var (
		h = &handlers{
			help:   makeHelp(),
			domain: ko.MustString("server.domain"),
		}
		ge *geo.Geo
	)

	// Timezone service.
	if ko.Bool("timezones.enabled") || ko.Bool("weather.enabled") {
		fPath := ko.MustString("timezones.geo_filepath")
		lo.Printf("reading geo locations from %s", fPath)

		g, err := geo.New(fPath)
		if err != nil {
			lo.Fatalf("error loading geo locations: %v", err)
		}
		ge = g

		log.Printf("%d geo location names loaded", g.Count())

	}

	// Timezone service.
	if ko.Bool("timezones.enabled") {
		h.tz = timezones.New(timezones.Opt{}, ge)
		dns.HandleFunc("time.", handle(h.handleTime))
	}

	// FX currency conversion.
	if ko.Bool("fx.enabled") {
		h.fx = fx.New(fx.Opt{
			APIkey:          ko.MustString("fx.api_key"),
			RefreshInterval: ko.MustDuration("fx.refresh_interval"),
		})
		dns.HandleFunc("fx.", handle(h.handleFX))
	}

	// IP echo.
	if ko.Bool("myip.enabled") {
		dns.HandleFunc("myip.", handle(h.handleMyIP))
	}

	// Weather.
	if ko.Bool("weather.enabled") {
		h.weather = weather.New(weather.Opt{
			MaxEntries: ko.MustInt("weather.max_entries"),
			CacheTTL:   ko.MustDuration("weather.cache_ttl"),
			ReqTimeout: time.Second * 3,
			UserAgent:  ko.MustString("server.domain"),
		}, ge)

		dns.HandleFunc("weather.", handle(h.handleWeather))
	}

	// Help service.
	dns.HandleFunc("help.", handle(h.handleHelp))
	dns.HandleFunc(".", handle(h.handleDefault))

	// Start the server.
	server := &dns.Server{Addr: ko.MustString("server.address"), Net: "udp"}
	lo.Println("listening on ", ko.String("server.address"))
	if err := server.ListenAndServe(); err != nil {
		lo.Fatalf("error starting server: %v", err)
	}
	defer server.Shutdown()
}
