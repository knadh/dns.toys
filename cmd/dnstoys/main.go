package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/knadh/dns.toys/internal/geo"
	"github.com/knadh/dns.toys/internal/ifsc"
	"github.com/knadh/dns.toys/internal/services/aerial"
	"github.com/knadh/dns.toys/internal/services/base"
	"github.com/knadh/dns.toys/internal/services/cidr"
	"github.com/knadh/dns.toys/internal/services/coin"
	"github.com/knadh/dns.toys/internal/services/dice"
	"github.com/knadh/dns.toys/internal/services/dict"
	"github.com/knadh/dns.toys/internal/services/epoch"
	"github.com/knadh/dns.toys/internal/services/excuse"
	"github.com/knadh/dns.toys/internal/services/fx"
	"github.com/knadh/dns.toys/internal/services/nanoid"
	"github.com/knadh/dns.toys/internal/services/num2words"
	"github.com/knadh/dns.toys/internal/services/random"
	"github.com/knadh/dns.toys/internal/services/sudoku"
	"github.com/knadh/dns.toys/internal/services/timezones"
	"github.com/knadh/dns.toys/internal/services/units"
	"github.com/knadh/dns.toys/internal/services/uuid"
	"github.com/knadh/dns.toys/internal/services/weather"
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

// TTL is set to 1 day (60*60*24 = 86,400).
const HELP_TTL = 86400

// Not all platforms have syscall.SIGUNUSED so use Golang's default definition here
const SIGUNUSED = syscall.Signal(0x1f)

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

func saveSnapshot(h *handlers) {
	interruptSignal := make(chan os.Signal)
	signal.Notify(interruptSignal,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGINT,
		SIGUNUSED, // SIGUNUSED, can be used to avoid shutting down the app.
	)

	// On receiving an OS signal, iterate through services and
	// dump their snapshots to the disk if available.
	for {
		select {
		case i := <-interruptSignal:
			lo.Printf("received SIGNAL: `%s`", i.String())

			for name, s := range h.services {
				if !ko.Bool(name+".enabled") || !ko.Bool(name+".snapshot_enabled") {
					continue
				}

				b, err := s.Dump()
				if err != nil {
					lo.Printf("error generating %s snapshot: %v", name, err)
				}

				if b == nil {
					continue
				}

				filePath := ko.MustString(name + ".snapshot_file")
				lo.Printf("saving %s snapshot to %s", name, filePath)
				if err := os.WriteFile(filePath, b, 0644); err != nil {
					lo.Printf("error writing %s snapshot: %v", name, err)
				}
			}

			if i != SIGUNUSED {
				os.Exit(0)
			}
		}
	}
}

func loadSnapshot(service string) []byte {
	if !ko.Bool(service + ".snapshot_enabled") {
		return nil
	}

	filePath := ko.MustString(service + ".snapshot_file")

	b, err := os.ReadFile(filePath)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return nil
		}
		lo.Printf("error reading snapshot file %s: %v", filePath, err)
		return nil
	}

	return b
}

func main() {
	initConfig()

	var (
		h = &handlers{
			services: make(map[string]Service),
			domain:   ko.MustString("server.domain"),
		}
		ge  *geo.Geo
		mux = dns.NewServeMux()

		help = [][]string{}
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

		lo.Printf("%d geo location names loaded", g.Count())
	}

	// Timezone service.
	if ko.Bool("timezones.enabled") {
		tz := timezones.New(timezones.Opt{}, ge)
		h.register("time", tz, mux)

		help = append(help, []string{"get time for a city", "dig mumbai.time @%s"})
	}

	// FX currency conversion.
	if ko.Bool("fx.enabled") {
		f := fx.New(fx.Opt{
			RefreshInterval: ko.MustDuration("fx.refresh_interval"),
		})

		// Load snapshot?
		if b := loadSnapshot("fx"); b != nil {
			if err := f.Load(b); err != nil {
				lo.Printf("error reading fx snapshot: %v", err)
			}
		}

		h.register("fx", f, mux)

		help = append(help, []string{"convert currency rates", "dig 99USD-INR.fx @%s"})
	}

	// IP echo.
	if ko.Bool("ip.enabled") {
		mux.HandleFunc("ip.", h.handleEchoIP)

		help = append(help, []string{"get your host's requesting IP.", "dig ip @%s"})
	}

	// Weather.
	if ko.Bool("weather.enabled") {
		w := weather.New(weather.Opt{
			MaxEntries:       ko.MustInt("weather.max_entries"),
			ForecastInterval: ko.MustDuration("weather.forecast_interval"),
			CacheTTL:         ko.MustDuration("weather.cache_ttl"),
			ReqTimeout:       time.Second * 3,
			UserAgent:        ko.MustString("server.domain"),
		}, ge)

		// Load snapshot?
		if b := loadSnapshot("weather"); b != nil {
			if err := w.Load(b); err != nil {
				lo.Printf("error reading weather snapshot: %v", err)
			}
		}

		h.register("weather", w, mux)

		help = append(help, []string{"get weather forecast for a city.", "dig berlin.weather @%s"})
	}

	// Units.
	if ko.Bool("units.enabled") {
		u, err := units.New()
		if err != nil {
			lo.Fatalf("error initializing units service: %v", err)
		}
		h.register("unit", u, mux)

		help = append(help, []string{"convert between units.", "dig 42km-cm.unit @%s"})
	}

	// Numbers to words.
	if ko.Bool("num2words.enabled") {
		n := num2words.New()
		h.register("words", n, mux)

		help = append(help, []string{"convert numbers to words.", "dig 123456.words @%s"})
	}

	// CIDR.
	if ko.Bool("cidr.enabled") {
		n := cidr.New()
		h.register("cidr", n, mux)

		help = append(help, []string{"convert cidr to ip range.", "dig 10.100.0.0/24.cidr @%s"})
	}

	// PI.
	if ko.Bool("pi.enabled") {
		mux.HandleFunc("pi.", h.handlePi)

		help = append(help, []string{"return digits of Pi as TXT or A or AAAA record.", "dig pi @%s"})
	}

	// Base
	if ko.Bool("base.enabled") {
		n := base.New()
		h.register("base", n, mux)

		help = append(help, []string{"convert numbers from one base to another", "dig 100dec-hex.base @%s"})
	}

	// Dictionary.
	if ko.Bool("dict.enabled") {
		d := dict.New(dict.Opt{
			WordNetPath: ko.MustString("dict.wordnet_path"),
			MaxResults:  ko.MustInt("dict.max_results"),
		})
		h.register("dict", d, mux)

		help = append(help, []string{"get the definition of an English word, powered by WordNet(R).", "dig fun.dict @%s"})
	}

	// Rolling dice
	if ko.Bool("dice.enabled") {
		n := dice.New()
		h.register("dice", n, mux)

		help = append(help, []string{"roll dice", "dig 1d6.dice @%s"})
	}

	// Random number generator.
	if ko.Bool("rand.enabled") {
		// seed the RNG:
		rand.Seed(time.Now().Unix())

		n := random.New()
		h.register("rand", n, mux)

		help = append(help, []string{"generate random numbers", "dig 1-100.rand @%s"})
	}

	// Coin toss.
	if ko.Bool("coin.enabled") {
		n := coin.New()
		h.register("coin", n, mux)

		help = append(help, []string{"toss coin", "dig 2.coin @%s"})
	}

	// Epoch / Unix timestamp conversion.
	if ko.Bool("epoch.enabled") {
		n := epoch.New(ko.Bool("epoch.send_local_time"))
		h.register("epoch", n, mux)

		help = append(help, []string{"convert epoch / UNIX time to human readable time.", "dig 784783800.epoch @%s"})
	}

	// Aerial Distance between Lat,Lng
	if ko.Bool("aerial.enabled") {
		a := aerial.New()
		h.register("aerial", a, mux)

		help = append(help, []string{"get aerial distance between lat lng pair", "dig A12.9352,77.6245/12.9698,77.7500.aerial @%s"})
	}

	// Random UUID.
	if ko.Bool("uuid.enabled") {
		u := uuid.New(ko.Int("uuid.max_results"))
		h.register("uuid", u, mux)

		help = append(help, []string{"generate random UUID-v4s", "dig 2.uuid @%s"})
	}

	// Sudoku Solver
	if ko.Bool("sudoku.enabled") {
		ssolver := sudoku.New()
		h.register("sudoku", ssolver, mux)
		// enter the sudoku puzzle string in row major format, each row separated by a dot, empty cells should have value 0
		help = append(help, []string{"solve a sudoku puzzle", "dig 002840003.076000000.100006050.030080000.007503200.000020010.080100004.000000730.700064500.sudoku @%s"})
	}

	// Developer Excuse
	if ko.Bool("excuse.enabled") {
		e, err := excuse.New(ko.MustString("excuse.file"))
		if err != nil {
			lo.Fatalf("error initializing units service: %v", err)
		}
		h.register("excuse", e, mux)
		help = append(help, []string{"return a developer excuse", "dig excuse @%s"})
	}

	// NanoID Generator
	if ko.Bool("nanoid.enabled") {
		n := nanoid.New(ko.MustInt("nanoid.max_results"), ko.MustInt("nanoid.max_length"))
		h.register("nanoid", n, mux)

		help = append(help, []string{"generate random NanoIDs", "dig 2.10.nanoid @%s"})
	}

	// IFSC service
	if ko.Bool("ifsc.enabled") {
		e, err := ifsc.New(ko.MustString("ifsc.data_path"))
		if err != nil {
			lo.Fatalf("error initializing ifsc service: %v", err)
		}
		h.register("ifsc", e, mux)
		help = append(help, []string{"lookup (Indian) bank details by IFSC code", "dig ABNA0000001.ifsc @%s"})
	}

	// Prepare the static help response for the `help` query.
	for _, l := range help {
		r, err := dns.NewRR(fmt.Sprintf("help. %d TXT \"%s\" \"%s\"", HELP_TTL, l[0], fmt.Sprintf(l[1], h.domain)))
		if err != nil {
			lo.Fatalf("error preparing: %v", err)
		}

		h.help = append(h.help, r)
	}

	mux.HandleFunc("help.", h.handleHelp)
	mux.HandleFunc(".", (h.handleDefault))

	// Start the snapshot listener.
	go saveSnapshot(h)

	// Start the server.
	server := &dns.Server{
		Addr:    ko.MustString("server.address"),
		Net:     "udp",
		Handler: mux,
	}
	lo.Println("listening on ", ko.String("server.address"))
	if err := server.ListenAndServe(); err != nil {
		lo.Fatalf("error starting server: %v", err)
	}
	defer server.Shutdown()
}
