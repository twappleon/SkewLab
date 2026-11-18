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
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/twappleon/SkewLab/internal/analytics"
	"github.com/twappleon/SkewLab/internal/fees"
	"github.com/twappleon/SkewLab/internal/maker"
	"github.com/twappleon/SkewLab/internal/notify"
	"github.com/twappleon/SkewLab/internal/oracle"
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/storage"
	"github.com/twappleon/SkewLab/internal/strategy"
	"github.com/twappleon/SkewLab/internal/stream"
)

//go:embed dashboard.html performance.html
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
	alerts              *notify.Client
	lastSignalAt        time.Time
	lastSignalDirection string
	feedBadSince        time.Time
	feedDegraded        bool
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
	// Mark imminent expiries before Evaluate transitions them to PENDING_RESOLUTION.
	if !a.demo {
		for _, position := range a.ledger.Positions {
			if position.Status == "OPEN" && !now.Before(position.MarketEnd) {
				a.alerts.Publish("pending:"+position.ID, fmt.Sprintf("🧪 SkewLab 待結算\n市場：%s\n方向：%s\n尚未取得官方結果，不預估收益。", position.Market, position.Side))
			}
		}
	}
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
	a.alertDirection(now)
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
					a.alertPaperAction(now, action)
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
	a.alertHealth(now, errors)
	if a.persistError != "" {
		a.alerts.Publish("risk:disk:"+now.Format("2006-01-02"), "🚨 SkewLab 模擬交易已停止：持久化資料錯誤，須人工排除。")
	}
	a.alertRisk(now)
	a.alertDailySummary(now)
	if t.Price > 0 {
		a.saveObservation(Observation{Time: now, BTC: t.Price, Source: t.Source, Signal: a.sig, Markets: markets})
	}
	a.saveLedger()
	a.mu.Unlock()
	if !a.demo {
		a.reconcile(ctx, now)
	}
}

// Telegram only delivers significant transitions, never tick-by-tick price data.
func (a *App) alertDirection(now time.Time) {
	if a.demo || a.alerts == nil {
		return
	}
	direction := a.sig.Direction
	if direction == "WAIT" {
		a.lastSignalDirection = "WAIT"
		return
	}
	if !a.sig.Ready || (direction == a.lastSignalDirection) || now.Sub(a.lastSignalAt) < 5*time.Minute {
		return
	}
	a.lastSignalDirection = direction
	a.lastSignalAt = now
	bucket := now.Truncate(5 * time.Minute).Format(time.RFC3339)
	a.alerts.Publish("signal:"+bucket+":"+direction, fmt.Sprintf("📊 SkewLab BTC 方向觀察訊號\n方向：%s\n30秒動量：%+.4f%%\n半方差偏度：%+.3f\n來源：%s\n僅供模擬研究，並非已成交。", direction, a.sig.MomentumPct, a.sig.Skew, a.lastBTC.Source))
}
func (a *App) alertPaperAction(now time.Time, action string) {
	if a.demo || a.alerts == nil {
		return
	}
	for i := len(a.ledger.Positions) - 1; i >= 0; i-- {
		p := a.ledger.Positions[i]
		if strings.HasPrefix(action, "PAPER BUY") && p.Status == "OPEN" && p.EntryTime.Equal(now) {
			a.alerts.Publish("buy:"+p.ID, fmt.Sprintf("🧪 SkewLab 模擬開倉\n市場：%s\n方向：%s\n成本：%.2f USD\n模擬買價：%.4f\n非真實下單。", p.Market, p.Side, p.EntryCost, p.EntryPrice))
			return
		}
		if strings.HasPrefix(action, "PAPER SELL") && p.Status == "CLOSED" && p.ExitTime != nil && p.ExitTime.Equal(now) {
			headline := "🧪 SkewLab 模擬平倉"
			if p.Note == "net take profit" {
				headline = "✅ SkewLab 模擬止盈（淨收益已達目標）"
			}
			if p.Note == "net stop loss" {
				headline = "🛑 SkewLab 模擬止損"
			}
			a.alerts.Publish("sell:"+p.ID, fmt.Sprintf("%s\n市場：%s\n方向：%s\n成本：%.2f USD\n模擬賣價：%.4f\n淨損益：%+.2f USD（%+.2f%%）\n原因：%s\n非真實下單。", headline, p.Market, p.Side, p.EntryCost, p.ExitPrice, p.RealizedPnL, 100*p.RealizedPnL/p.EntryCost, p.Note))
			return
		}
	}
}
func (a *App) alertHealth(now time.Time, errors []string) {
	if a.demo || a.alerts == nil {
		return
	}
	if len(errors) > 0 {
		if a.feedBadSince.IsZero() {
			a.feedBadSince = now
		}
		if !a.feedDegraded && now.Sub(a.feedBadSince) >= time.Minute {
			a.feedDegraded = true
			a.alerts.Publish("outage:"+now.Truncate(20*time.Minute).Format(time.RFC3339), "⚠️ SkewLab 行情連線異常超過 60 秒\n暫停依缺失行情開新模擬倉。請查看健康狀態與 Railway 日誌。")
		}
		return
	}
	a.feedBadSince = time.Time{}
	if a.feedDegraded {
		a.feedDegraded = false
		a.alerts.Publish("recovery:"+now.Truncate(20*time.Minute).Format(time.RFC3339), "✅ SkewLab 行情資料已恢復。模擬運行持續，非實盤交易。")
	}
}
func (a *App) alertRisk(now time.Time) {
	if a.demo || a.alerts == nil {
		return
	}
	if a.oracleConflict != "" {
		a.alerts.Publish("risk:oracle:"+now.Format("2006-01-02"), "🚨 SkewLab 甲骨文結果衝突：已停止模擬新開倉，請檢查 Chainlink 與 Gamma 對帳。")
	}
	pnl := 0.0
	for _, p := range a.ledger.Positions {
		if p.ExitTime != nil && p.ExitTime.UTC().Format("2006-01-02") == now.UTC().Format("2006-01-02") {
			pnl += p.RealizedPnL
		}
		if p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION" {
			if p.MarkPnL < 0 {
				pnl += p.MarkPnL
			}
		}
	}
	if pnl <= -a.ledger.Config.MaxDailyLossUSD {
		a.alerts.Publish("risk:loss:"+now.UTC().Format("2006-01-02"), fmt.Sprintf("🚨 SkewLab 模擬每日虧損觸及限制\n估算當日損益：%.2f USD\n限制：%.2f USD\n模擬新開倉已由風控限制。", pnl, a.ledger.Config.MaxDailyLossUSD))
	}
}
func (a *App) alertDailySummary(now time.Time) {
	if a.demo || a.alerts == nil {
		return
	}
	local := now.In(time.FixedZone("Phnom Penh", 7*3600))
	if local.Hour() < 9 {
		return
	}
	day := local.Format("2006-01-02")
	stats := analytics.Calculate(a.ledger, now)
	pf := "不適用（沒有虧損樣本）"
	if stats.Last14Days.ProfitFactor != nil {
		pf = fmt.Sprintf("%.2f", *stats.Last14Days.ProfitFactor)
	}
	average := "不適用"
	if stats.Last14Days.AverageNetUSD != nil {
		average = fmt.Sprintf("%+.4f USD", *stats.Last14Days.AverageNetUSD)
	}
	evidence := "樣本不足，仍需驗證"
	if !stats.DataComplete {
		evidence = "歷史交易資料不完整，需查核"
	} else if stats.SampleLevel == "research_sample" {
		evidence = "研究樣本，仍需樣本外驗證"
	}
	msg := fmt.Sprintf("📋 SkewLab 每日模擬報告（%s，柬埔寨時間）\n模擬淨值：%.2f USD\n累計已實現淨損益：%+.2f USD\n完成平倉：%d，獲勝：%d\n最大回撤：%.2f%%\n近 7 天已實現損益：%+.2f USD（%d 筆）\n近 14 天平均每筆：%s\n近 14 天 Profit Factor：%s\n樣本狀態：%s\n未完成持倉：%d\n僅模擬，非可提領收益。",
		day, a.ledger.CurrentEquity(), a.ledger.RealizedPnL, a.ledger.TotalClosed, a.ledger.Wins, a.ledger.MaxDrawdownPct,
		stats.Last7Days.NetPnLUSD, stats.Last7Days.Closed, average, pf, evidence, a.ledger.OpenPositions())
	for _, group := range analytics.CompareMarkets(a.ledger, now) {
		if group.Market == "other" {
			continue
		}
		msg += fmt.Sprintf("\n%s 市場（14日）：%d 筆，淨損益 %+.2f USD，止盈 %d、止損 %d、官方結算 %d",
			group.Market, group.Period.Closed, group.Period.NetPnLUSD, group.Exits.TakeProfit, group.Exits.StopLoss, group.Exits.Settled)
	}
	causes := analytics.SummarizeCauses(a.ledger, now)
	msg += fmt.Sprintf("\n14日止損超過設定5%%幅度：%d 筆，超出金額 %.2f USD；成本分解 %d/%d 筆可計算（僅模型估算，非真實滑價）。", causes.StopThresholdBreaches, causes.ExcessStopLossUSD, causes.AnalyzedTrades, causes.CompletedTrades)
	a.alerts.Publish("daily:"+day, msg)
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
			for _, p := range a.ledger.Positions {
				if p.Status == "SETTLED" && p.ExitTime != nil && p.ExitTime.Equal(now) && p.Market == slug {
					a.alerts.Publish("settled:"+p.ID, fmt.Sprintf("🧪 SkewLab 官方結果結算\n市場：%s\n持倉：%s\n官方結果：%s\n模擬淨損益：%+.2f USD\n注意：Gamma 結算，非獨立 Chainlink 驗證。", slug, p.Side, settlement.Winner, p.RealizedPnL))
				}
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
		Telegram            notify.Status       `json:"telegram_notifications"`
	}{mode(a.demo), "0.4.0", a.lastBTC, a.sig, ordered, a.makerQuotes, fees.CryptoCoefficient(time.Now().UTC()), "Gamma-published outcome, not independent Chainlink", a.ledger, a.errorMessage, a.events, a.ledger.LastEvaluation, a.clob.GetStatus(), a.chainlink.Status(), a.queue.Snapshot(), a.queue.Candidates(), a.alerts.Status()})
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

// publicPerformance serves a deliberately limited, read-only paper trading
// summary. It never exposes positions, token IDs, orderbooks, raw observation
// archives, strategy details, credentials or the internal event log.
func (a *App) publicPerformance(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	l := a.ledger
	equity := l.CurrentEquity()
	initial := l.Config.StartingCash
	roi := 0.0
	if initial > 0 {
		roi = 100 * (equity - initial) / initial
	}
	winRate := 0.0
	if l.TotalClosed > 0 {
		winRate = 100 * float64(l.Wins) / float64(l.TotalClosed)
	}
	now := time.Now().UTC()
	pending := 0
	for _, p := range l.Positions {
		if p.Status == "PENDING_RESOLUTION" {
			pending++
		}
	}
	fresh := a.lastBTC.Price > 0 && !a.lastBTC.Time.IsZero() && now.Sub(a.lastBTC.Time) >= 0 && now.Sub(a.lastBTC.Time) <= 15*time.Second
	_ = json.NewEncoder(w).Encode(struct {
		Mode              string                     `json:"mode"`
		Version           string                     `json:"version"`
		AsOf              time.Time                  `json:"as_of"`
		LastMarketTick    time.Time                  `json:"last_market_tick,omitempty"`
		FeedFresh         bool                       `json:"feed_fresh"`
		InitialCash       float64                    `json:"initial_cash_usd"`
		Equity            float64                    `json:"equity_usd"`
		ReturnPct         float64                    `json:"return_pct"`
		RealizedPnL       float64                    `json:"realized_pnl_usd"`
		ClosedTrades      int                        `json:"closed_trades"`
		Wins              int                        `json:"wins"`
		WinRatePct        float64                    `json:"win_rate_pct"`
		MaxDrawdownPct    float64                    `json:"max_drawdown_pct"`
		OpenPositions     int                        `json:"open_positions"`
		PendingSettlement int                        `json:"pending_settlement"`
		Analytics         analytics.Summary          `json:"analytics"`
		MarketWindows     []analytics.MarketWindow   `json:"market_windows"`
		RootCauses        analytics.RootCauseSummary `json:"root_causes"`
	}{
		mode(a.demo), "0.7.0", now, a.lastBTC.Time, fresh,
		initial, equity, roi, l.RealizedPnL, l.TotalClosed, l.Wins,
		winRate, l.MaxDrawdownPct, l.OpenPositions(), pending, analytics.Calculate(l, now), analytics.CompareMarkets(l, now), analytics.SummarizeCauses(l, now),
	})
}

func (a *App) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/performance" && r.URL.Path != "/admin" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		filename := "performance.html"
		if r.URL.Path == "/admin" {
			filename = "dashboard.html"
		}
		data, err := dashboard.ReadFile(filename)
		if err != nil {
			http.Error(w, "page unavailable", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(data)
	})
	mux.HandleFunc("/api/performance", a.publicPerformance)
	mux.HandleFunc("/api/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		a.mu.RLock()
		defer a.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_ = json.NewEncoder(w).Encode(map[string]any{"paper_only": true, "source": "modeled ledger, not real trades", "trades": analytics.PrivateDiagnostics(a.ledger, time.Now().UTC(), 100)})
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
		public := r.URL.Path == "/health" || r.URL.Path == "/" || r.URL.Path == "/performance" || r.URL.Path == "/api/performance"
		if !public && password != "" {
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

// localFirstListenAddr prevents unauthenticated portfolio and tick archives from
// being reachable over LAN/WAN merely because the bot was started on a Mac.
// Non-loopback serving is permitted only behind a trusted HTTPS reverse proxy
// (or Railway, where HTTPS terminates at the edge) with dashboard authentication.
func localFirstListenAddr(host, port string, railway, passwordSet, trustedProxy bool) (string, error) {
	if port == "" {
		port = "8080"
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("invalid PORT %q", port)
	}
	if host == "" {
		if railway {
			host = "0.0.0.0"
		} else {
			host = "127.0.0.1"
		}
	}
	if host != "127.0.0.1" && host != "::1" && host != "localhost" {
		if !passwordSet || !(railway || trustedProxy) {
			return "", fmt.Errorf("remote dashboard bind blocked: set DASHBOARD_PASSWORD and TRUSTED_REVERSE_PROXY=1 only behind HTTPS, or run with BIND_ADDR=127.0.0.1")
		}
	}
	return net.JoinHostPort(host, port), nil
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
	if !*demo {
		a.alerts = notify.FromEnv(*dataDir)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if a.alerts != nil {
		go a.alerts.Run(ctx)
	}
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
	listenAddr, err := localFirstListenAddr(os.Getenv("BIND_ADDR"), os.Getenv("PORT"), os.Getenv("RAILWAY_ENVIRONMENT") != "", os.Getenv("DASHBOARD_PASSWORD") != "", os.Getenv("TRUSTED_REVERSE_PROXY") == "1")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Polymarket Skew Bot v0.7.0 (%s), http://%s", mode(*demo), listenAddr)
	srv := &http.Server{Addr: listenAddr, Handler: a.handler(), ReadHeaderTimeout: 5 * time.Second}
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
