package main

import (
	"context"
	"crypto/subtle"
	"embed"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/twappleon/SkewLab/internal/fees"
	"github.com/twappleon/SkewLab/internal/maker"
	"github.com/twappleon/SkewLab/internal/oracle"
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/storage"
	"github.com/twappleon/SkewLab/internal/strategy"
	"github.com/twappleon/SkewLab/internal/stream"
)

//go:embed dashboard.html
var dashboard embed.FS

type Observation struct {
	Time    time.Time           `json:"time"`
	BTC     float64             `json:"btc"`
	Source  string              `json:"source"`
	Signal  strategy.Result     `json:"signal"`
	Markets []polymarket.Market `json:"markets"`
}
type App struct {
	mu                  sync.RWMutex
	api                 *polymarket.Client
	clob                *stream.Feed
	chainlink           *oracle.PolyBolt
	queue               *maker.QueueMonitor
	persistError        string
	oracleConflict      string
	ticks               []strategy.Tick
	lastBTC             strategy.Tick
	sig                 strategy.Result
	markets             map[int]polymarket.Market
	makerQuotes         []maker.Proposal
	lastSettlementCheck time.Time
	ledger              *paper.Ledger
	demo                bool
	errorMessage        string
	events              []string
	dataDir             string
	sampleNumber        int
	minMomentum         float64
	minSkew             float64
	poll                time.Duration
}

func NewApp(dir string, demo bool, minMomentum, minSkew float64, poll time.Duration) *App {
	l := paper.New(paper.DefaultConfig())
	persistentFailure := ""
	path := filepath.Join(dir, "paper-ledger.json")
	if data, err := os.ReadFile(path); err == nil {
		var saved paper.Ledger
		if e := json.Unmarshal(data, &saved); e != nil || saved.Config.StartingCash <= 0 || saved.Cash < 0 || math.IsNaN(saved.Cash) || math.IsInf(saved.Cash, 0) {
			persistentFailure = "LEDGER CORRUPTED: manual recovery required; all paper execution disabled"
		} else {
			l = &saved
			log.Printf("restored %d paper positions", len(saved.Positions))
		}
	} else if !os.IsNotExist(err) {
		persistentFailure = "LEDGER UNREADABLE: all paper execution disabled: " + err.Error()
	}
	a := &App{api: polymarket.New(), clob: stream.NewFeed(), queue: maker.NewQueueMonitor(), ledger: l, demo: demo, markets: map[int]polymarket.Market{}, dataDir: dir, minMomentum: minMomentum, minSkew: minSkew, poll: poll, persistError: persistentFailure}
	a.chainlink = oracle.NewPolyBolt(oracle.CredentialsFromEnv(), func(t oracle.Tick) { a.saveChainlink(t) })
	if !demo {
		records, e := storage.ReadRecentJSONL[oracle.Tick](filepath.Join(dir, "chainlink-twap-60s.jsonl"), 7200)
		if e != nil {
			log.Printf("Chainlink TWAP archive invalid, ignoring: %v", e)
		} else {
			a.chainlink.Seed(records)
		}
	}
	if demo {
		a.seedDemo()
	}
	return a
}
func (a *App) seedDemo() {
	now := time.Now().UTC()
	for i := 120; i >= 0; i-- {
		age := float64(120 - i)
		price := 98000 + age*1.9 + 22*math.Sin(age/8) + 9*math.Sin(age*1.7)
		a.ticks = append(a.ticks, strategy.Tick{Time: now.Add(-time.Duration(i) * time.Second), Price: price, Source: "SYNTHETIC DEMO (not live)"})
	}
	a.lastBTC = a.ticks[len(a.ticks)-1]
}
func (a *App) demoNext(now time.Time) (strategy.Tick, []polymarket.Market) {
	a.sampleNumber++
	k := float64(a.sampleNumber)
	price := a.lastBTC.Price + 1.8 + 7*math.Sin(k*0.32) + 3*math.Sin(k*1.4)
	tick := strategy.Tick{Time: now, Price: price, Source: "SYNTHETIC DEMO (not live)"}
	markets := make([]polymarket.Market, 0, 2)
	for _, seconds := range []int{300, 900} {
		end := time.Unix(now.Unix()/int64(seconds)*int64(seconds)+int64(seconds), 0)
		drift := 0.02 * math.Sin(k/20)
		m := polymarket.Market{Slug: polymarket.MarketSlug(now, seconds), Question: fmt.Sprintf("SIMULATED BTC up/down %dm", seconds/60), Seconds: seconds, End: end, Active: true, UpToken: "demo-up", DownToken: "demo-down"}
		m.Up = polymarket.Quote{Bid: .492 + drift, Ask: .507 + drift, BidDepth: 600, AskDepth: 600, TickSize: .001}
		m.Down = polymarket.Quote{Bid: .493 - drift, Ask: .508 - drift, BidDepth: 580, AskDepth: 580, TickSize: .001}
		markets = append(markets, m)
	}
	return tick, markets
}
func (a *App) cycle(ctx context.Context) {
	now := time.Now().UTC()
	var t strategy.Tick
	var markets []polymarket.Market
	var errors []string
	if a.demo {
		a.mu.Lock()
		t, markets = a.demoNext(now)
		a.mu.Unlock()
	} else {
		spot, err := a.api.BTC(ctx)
		if err != nil {
			errors = append(errors, "BTC feed unavailable: "+err.Error())
		} else {
			t = strategy.Tick{Time: now, Price: spot.Price, Source: spot.Source}
		}
		var assets []string
		for _, seconds := range []int{300, 900} {
			m, err := a.api.Discover(ctx, now, seconds)
			if err == nil {
				assets = append(assets, m.UpToken, m.DownToken)
				up, okUp := a.clob.Snapshot(m.UpToken, time.Now().UTC())
				down, okDown := a.clob.Snapshot(m.DownToken, time.Now().UTC())
				if okUp && okDown {
					m.Up, m.Down = up, down
				} else {
					m, err = a.api.RefreshQuotes(ctx, m) // REST fallback; timestamps still checked
				}
			}
			if err != nil {
				m.Active = false
				m.Error = err.Error()
				errors = append(errors, fmt.Sprintf("%dm: %s", seconds/60, err.Error()))
			}
			markets = append(markets, m)
		}
		if len(assets) > 0 {
			a.clob.SetAssets(assets)
		}
	}
	a.mu.Lock()
	// Use the completion time, not request start time, for freshness checks.
	now = time.Now().UTC()
	if t.Price > 0 {
		if a.lastBTC.Source != "" && a.lastBTC.Source != t.Source {
			a.ticks = nil
			errors = append(errors, "BTC source changed: 30-second warmup reset")
		}
		a.ticks = append(a.ticks, t)
		a.lastBTC = t
		first := 0
		for first < len(a.ticks) && now.Sub(a.ticks[first].Time) > 65*time.Second {
			first++
		}
		a.ticks = a.ticks[first:]
	}
	a.sig = strategy.Calculate(a.ticks, now, a.minMomentum, a.minSkew)
	a.makerQuotes = nil
	// Do not submit any paper orders when either BTC or the corresponding
	// CLOB book is missing; previous quotes are never reused for execution.
	if t.Price > 0 && now.Sub(t.Time) <= 5*time.Second {
		for _, m := range markets {
			a.markets[m.Seconds] = m
			if m.Active {
				proposal := maker.Plan(m, a.sig, now, maker.DefaultConfig())
				a.makerQuotes = append(a.makerQuotes, proposal)
				if !a.demo && m.Up.Source == "CLOB_WEBSOCKET" {
					a.queue.Update(m, proposal, now)
				}
				actions := []string{}
				if a.persistError == "" && a.oracleConflict == "" {
					actions = a.ledger.Evaluate(m, a.sig, now)
				}
				for _, action := range actions {
					a.events = append([]string{now.Format(time.RFC3339) + " " + action}, a.events...)
				}
			} else {
				if a.persistError == "" && a.oracleConflict == "" {
					a.ledger.Evaluate(m, strategy.Result{Direction: "WAIT"}, now)
				}
			}
		}
	} else {
		if t.Price > 0 {
			errors = append(errors, "market snapshot delayed; suppressed paper trading")
		}
		for _, m := range markets {
			a.markets[m.Seconds] = m
		}
	}
	if len(a.events) > 16 {
		a.events = a.events[:16]
	}
	a.ledger.ExpireUnresolved(now) // fail closed even during complete feed outage
	a.ledger.RecordEquity(now)
	if a.persistError != "" {
		errors = append(errors, a.persistError)
	}
	if a.oracleConflict != "" {
		errors = append(errors, a.oracleConflict)
	}
	a.errorMessage = strings.Join(errors, "; ")
	if t.Price > 0 {
		a.saveObservation(Observation{Time: now, BTC: t.Price, Source: t.Source, Signal: a.sig, Markets: markets})
	}
	a.saveLedger()
	a.mu.Unlock()
	if !a.demo {
		a.reconcile(ctx, now)
	}
}

// Captured Chainlink values are stored separately from exchange spot ticks.
// The key and authentication frames are never written to disk.
func (a *App) saveChainlink(t oracle.Tick) {
	if a.demo {
		return
	}
	if err := storage.AppendJSONL(filepath.Join(a.dataDir, "chainlink-twap-60s.jsonl"), t, 25<<20); err != nil {
		log.Printf("chainlink audit record failed: %v", err)
	}
}

// Maker trade audit is non-executable research; it does not alter the paper ledger.
func (a *App) observeMaker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-a.clob.Trades():
			study, matched := a.queue.Observe(t)
			if !matched {
				continue
			}
			if err := storage.AppendJSONL(filepath.Join(a.dataDir, "maker-queue-audit.jsonl"), struct {
				Trade stream.Trade     `json:"public_trade"`
				Study maker.QueueStudy `json:"hypothetical_study"`
			}{t, study}, 25<<20); err != nil {
				log.Printf("queue audit write failed: %v", err)
			}
		}
	}
}

func (a *App) saveObservation(o Observation) {
	if err := storage.AppendJSONL(filepath.Join(a.dataDir, "observations.jsonl"), o, 50<<20); err != nil {
		log.Printf("observation record failed: %v", err)
	}
}
func (a *App) saveLedger() {
	if a.persistError != "" {
		return
	}
	data, err := json.MarshalIndent(a.ledger, "", "  ")
	if err != nil {
		a.persistError = "LEDGER SERIALIZATION FAILED; trading stopped: " + err.Error()
		return
	}
	if err = storage.Atomic(filepath.Join(a.dataDir, "paper-ledger.json"), data); err != nil {
		a.persistError = "LEDGER DISK WRITE FAILED; trading stopped: " + err.Error()
		log.Print(a.persistError)
	}
}

// reconcile checks expired inventory against Polymarket-published resolved
// market outcomes. Failing the API leaves positions pending and never guesses.
func (a *App) reconcile(ctx context.Context, now time.Time) {
	a.mu.Lock()
	if now.Sub(a.lastSettlementCheck) < 45*time.Second {
		a.mu.Unlock()
		return
	}
	if a.persistError != "" {
		a.mu.Unlock()
		return
	}
	a.lastSettlementCheck = now
	pending := map[string]string{}
	for _, p := range a.ledger.Positions {
		if (p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION") && !now.Before(p.MarketEnd) && p.ConditionID != "" {
			pending[p.Market] = p.ConditionID
		}
	}
	a.mu.Unlock()
	for slug, condition := range pending {
		settlement, err := a.api.Settlement(ctx, slug, condition)
		if err != nil {
			log.Printf("settlement still pending %s: %v", slug, err)
			continue
		}
		// Independently compare Chainlink if we possess the exact market-edge
		// TWAP observations. Missing oracle ticks remain Gamma-only (clearly
		// labeled), but a contradiction HALTS the paper strategy.
		duration := time.Duration(0)
		if strings.HasPrefix(slug, "btc-updown-5m-") {
			duration = 5 * time.Minute
		}
		if strings.HasPrefix(slug, "btc-updown-15m-") {
			duration = 15 * time.Minute
		}
		if duration > 0 {
			end := time.Time{}
			a.mu.RLock()
			for _, pos := range a.ledger.Positions {
				if pos.Market == slug {
					end = pos.MarketEnd
					break
				}
			}
			a.mu.RUnlock()
			if !end.IsZero() {
				audit, auditErr := a.chainlink.Audit(end.Add(-duration), duration)
				if auditErr == nil && audit.Comparable && audit.Direction != settlement.Winner {
					a.mu.Lock()
					a.oracleConflict = fmt.Sprintf("ORACLE CONFLICT %s: Chainlink=%s vs Gamma=%s; PAPER TRADING HALTED", slug, audit.Direction, settlement.Winner)
					a.errorMessage = a.oracleConflict
					a.mu.Unlock()
					log.Print("settlement BLOCKED: " + slug + " independent Chainlink disagreement")
					continue
				}
			}
		}
		a.mu.Lock()
		n := a.ledger.ApplySettlement(settlement, now)
		if n > 0 {
			a.events = append([]string{fmt.Sprintf("%s VERIFIED GAMMA OUTCOME %s %s (%d paper positions)", now.Format(time.RFC3339), slug, settlement.Winner, n)}, a.events...)
			if len(a.events) > 16 {
				a.events = a.events[:16]
			}
			a.ledger.RecordEquity(now)
			a.saveLedger()
		}
		a.mu.Unlock()
	}
}

func (a *App) state(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	var ordered []polymarket.Market
	for _, s := range []int{300, 900} {
		if m, ok := a.markets[s]; ok {
			ordered = append(ordered, m)
		}
	}
	_ = json.NewEncoder(w).Encode(struct {
		Mode                string              `json:"mode"`
		Version             string              `json:"version"`
		BTC                 strategy.Tick       `json:"btc"`
		Signal              strategy.Result     `json:"signal"`
		Markets             []polymarket.Market `json:"markets"`
		MakerQuotes         []maker.Proposal    `json:"maker_quotes"`
		FeeCoefficient      float64             `json:"taker_fee_coefficient"`
		SettlementSource    string              `json:"settlement_source"`
		Ledger              *paper.Ledger       `json:"ledger"`
		Error               string              `json:"error"`
		Events              []string            `json:"events"`
		Updated             time.Time           `json:"updated"`
		OrderbookFeed       stream.Status       `json:"orderbook_feed"`
		ChainlinkFeed       oracle.PolyStatus   `json:"chainlink_feed"`
		MakerQueue          []maker.QueueStudy  `json:"maker_queue_studies"`
		MakerCandidateCount uint64              `json:"maker_candidate_count"`
	}{mode(a.demo), "0.3.0", a.lastBTC, a.sig, ordered, a.makerQuotes, fees.CryptoCoefficient(time.Now().UTC()), "Gamma-published outcome, not independent Chainlink", a.ledger, a.errorMessage, a.events, a.ledger.LastEvaluation, a.clob.GetStatus(), a.chainlink.Status(), a.queue.Snapshot(), a.queue.Candidates()})
}
func mode(demo bool) string {
	if demo {
		return "SYNTHETIC DEMO"
	}
	return "LIVE DATA / PAPER TRADING"
}
func (a *App) export(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=paper-trades.csv")
	c := csv.NewWriter(w)
	defer c.Flush()
	_ = c.Write([]string{"id", "market", "side", "status", "entry_time", "entry_price", "qty", "entry_cost_usd", "exit_time", "exit_price", "realized_pnl_usd", "note"})
	for _, p := range a.ledger.Positions {
		et := ""
		if p.ExitTime != nil {
			et = p.ExitTime.Format(time.RFC3339)
		}
		row := []string{p.ID, p.Market, p.Side, p.Status, p.EntryTime.Format(time.RFC3339), fmt.Sprintf("%.6f", p.EntryPrice), fmt.Sprintf("%.6f", p.Qty), fmt.Sprintf("%.4f", p.EntryCost), et, fmt.Sprintf("%.6f", p.ExitPrice), fmt.Sprintf("%.4f", p.RealizedPnL), p.Note}
		_ = c.Write(row)
	}
}
func (a *App) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		data, _ := dashboard.ReadFile("dashboard.html")
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"execution":"paper_only"}`))
	})
	mux.HandleFunc("/api/state", a.state)
	mux.HandleFunc("/api/feeds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{"orderbook": a.clob.GetStatus(), "chainlink": a.chainlink.Status(), "maker_queue": a.queue.Snapshot()})
	})
	mux.HandleFunc("/api/chainlink-twap.jsonl", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		http.ServeFile(w, r, filepath.Join(a.dataDir, "chainlink-twap-60s.jsonl"))
	})
	mux.HandleFunc("/api/export.csv", a.export)
	mux.HandleFunc("/api/observations.jsonl", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		http.ServeFile(w, r, filepath.Join(a.dataDir, "observations.jsonl"))
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		password := os.Getenv("DASHBOARD_PASSWORD")
		if password != "" && r.URL.Path != "/health" {
			user := os.Getenv("DASHBOARD_USER")
			if user == "" {
				user = "admin"
			}
			gotUser, gotPass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(gotUser), []byte(user)) != 1 || subtle.ConstantTimeCompare([]byte(gotPass), []byte(password)) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="SkewLab", charset="UTF-8"`)
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}
func envFloat(key string, def float64) float64 {
	if n, e := strconv.ParseFloat(os.Getenv(key), 64); e == nil && n >= 0 {
		return n
	}
	return def
}
func main() {
	demo := flag.Bool("demo", false, "show clearly labeled synthetic BTC and order books; NEVER real money")
	dataDir := flag.String("data-dir", "./data", "directory for local ledger and observation logs")
	flag.Parse()
	if *demo {
		*dataDir = filepath.Join(*dataDir, "demo")
	}
	if err := os.MkdirAll(*dataDir, 0700); err != nil {
		log.Fatalf("mkdir data dir: %v", err)
	}
	poll := 3 * time.Second
	a := NewApp(*dataDir, *demo, envFloat("MIN_MOMENTUM_PCT", 0.02), envFloat("MIN_SKEW", 0.20), poll)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if !*demo {
		go a.clob.Run(ctx)
		go a.chainlink.Run(ctx)
		go a.observeMaker(ctx)
	}
	go func() {
		a.cycle(ctx)
		tick := time.NewTicker(poll)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				a.cycle(ctx)
			}
		}
	}()
	if !*demo && os.Getenv("RAILWAY_ENVIRONMENT") != "" && os.Getenv("DASHBOARD_PASSWORD") == "" {
		log.Fatal("Set DASHBOARD_PASSWORD before public Railway deployment; live paper dashboard is private")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Polymarket Skew Bot v0.3.0 (%s), http://localhost:%s", mode(*demo), port)
	srv := &http.Server{Addr: ":" + port, Handler: a.handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, done := context.WithTimeout(context.Background(), 6*time.Second)
		defer done()
		_ = srv.Shutdown(shutdownCtx)
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
