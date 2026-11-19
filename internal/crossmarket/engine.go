package crossmarket

// Research-only BTC/CLOB gate. Never submits an order.
import (
 "errors"
 "math"
 "time"
 "github.com/twappleon/SkewLab/internal/polymarketv2"
)
type Spot struct { Price float64; Source string; At time.Time }
type Input struct {
 Now time.Time
 Samples []Spot
 Book polymarketv2.Book
 Shares float64
 UpMarket bool
 Forecast polymarketv2.ProbabilityEvidence
 FeeCoefficient float64
 RealizedDailyLossUSD float64
 OpenWorstCaseLossUSD float64
}
type Result struct {
 Decision string `json:"decision"`
 Reason string `json:"reason"`
 Momentum30s float64 `json:"momentum_30s"`
 Quote polymarketv2.Quote `json:"quote"`
 Assessment polymarketv2.Assessment `json:"assessment"`
}
func finite(v float64) bool {return !math.IsNaN(v)&&!math.IsInf(v,0)}
func ValidateFeed(a []Spot, now time.Time) error {
 if len(a)<2{return errors.New("insufficient samples")}
 source:=a[0].Source
 if source==""||now.IsZero(){return errors.New("missing source or time")}
 for i,v:=range a{
  if v.Source!=source||!finite(v.Price)||v.Price<=100||v.At.IsZero()||v.At.After(now){return errors.New("invalid or mixed spot feed")}
  if i>0&&(v.At.Sub(a[i-1].At)<=0||v.At.Sub(a[i-1].At)>8*time.Second){return errors.New("gap or out-of-order spot feed")}
 }
 if now.Sub(a[len(a)-1].At)>5*time.Second{return errors.New("stale spot feed")}
 d:=a[len(a)-1].At.Sub(a[0].At)
 if d<27*time.Second||d>33*time.Second{return errors.New("30-second window missing")}
 return nil
}
func Evaluate(in Input) Result{
 r:=Result{Decision:"NO_TRADE",Reason:"insufficient evidence"}
 if err:=ValidateFeed(in.Samples,in.Now);err!=nil{r.Reason=err.Error();return r}
 first,last:=in.Samples[0],in.Samples[len(in.Samples)-1]
 r.Momentum30s=(last.Price-first.Price)/first.Price
 if in.UpMarket&&r.Momentum30s<=0||!in.UpMarket&&r.Momentum30s>=0{r.Reason="BTC movement against contract direction";return r}
 if !finite(in.RealizedDailyLossUSD)||!finite(in.OpenWorstCaseLossUSD)||in.RealizedDailyLossUSD<0||in.OpenWorstCaseLossUSD<0||in.RealizedDailyLossUSD+in.OpenWorstCaseLossUSD+2>2+1e-9{r.Reason="worst-case daily loss budget exceeded";return r}
 if !finite(in.Shares)||in.Shares<=0{r.Reason="invalid quantity";return r}
 r.Quote=polymarketv2.Evaluate(in.Book,in.Now,in.Shares)
 if !r.Quote.Executable{r.Reason="invalid executable quote: "+r.Quote.Reason;return r}
 if !in.Forecast.Valid(in.Now){r.Reason="independent out-of-sample probability evidence missing";return r}
 if !finite(in.FeeCoefficient)||in.FeeCoefficient<0||in.FeeCoefficient>0.1{r.Reason="invalid fee coefficient";return r}
 r.Assessment=polymarketv2.Assess(r.Quote,in.Forecast,in.Now,in.FeeCoefficient,2)
 if r.Assessment.Decision!="PAPER_CANDIDATE_ONLY"{r.Reason=r.Assessment.Reason;return r}
 cost:=r.Assessment.Shares*r.Quote.BuyVWAP+r.Assessment.ExpectedCostsUSD
 if !finite(cost)||cost<=0||cost>2+1e-8{r.Reason="all-in cost exceeds 2 USDC";return r}
 r.Decision="PAPER_CANDIDATE_ONLY"
 r.Reason="read-only research, not a simulated fill or real order"
 return r
}
