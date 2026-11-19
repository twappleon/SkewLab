package crossmarket
import (
 "testing"
 "time"
 "github.com/twappleon/SkewLab/internal/polymarketv2"
)
func fixture()(Input){
 now:=time.Date(2026,11,19,4,0,0,0,time.UTC)
 var samples []Spot
 for i:=0;i<=5;i++{samples=append(samples,Spot{Price:90000+float64(i)*20,Source:"Binance BTCUSDT",At:now.Add(time.Duration(i*6-30)*time.Second)})}
 return Input{Now:now,Samples:samples,Book:polymarketv2.Book{AssetID:"YES",Timestamp:now.Format(time.RFC3339),Asks:[]polymarketv2.Level{{Price:"0.40",Size:"100"}},Bids:[]polymarketv2.Level{{Price:"0.39",Size:"100"}}},Shares:4,UpMarket:true,Forecast:polymarketv2.ProbabilityEvidence{Probability:.8,Lower95:.7,SampleSize:130,ResolvedBefore:now.Add(-time.Hour),IndependentSource:"independent historical outcomes",OutOfSample:true},FeeCoefficient:.02}
}
func TestCandidateIsPaperOnly(t *testing.T){
 in:=fixture();r:=Evaluate(in)
 if r.Decision!="PAPER_CANDIDATE_ONLY"{t.Fatalf("%+v",r)}
 if r.Assessment.Shares*r.Quote.BuyVWAP+r.Assessment.ExpectedCostsUSD>2.00000001{t.Fatal("stake limit breached")}
}
func TestFailClosed(t *testing.T){
 tests:=[]struct{name string;change func(*Input)}{
  {"source switched",func(i *Input){i.Samples[3].Source="Coinbase"}},
  {"feed gap",func(i *Input){i.Samples[3].At=i.Samples[2].At.Add(9*time.Second)}},
  {"stale",func(i *Input){i.Now=i.Now.Add(6*time.Second)}},
  {"wrong direction",func(i *Input){i.UpMarket=false}},
  {"uncalibrated",func(i *Input){i.Forecast.OutOfSample=false}},
  {"risk",func(i *Input){i.RealizedDailyLossUSD=.01}},
  {"no bids",func(i *Input){i.Book.Bids=nil}},
  {"invalid fee",func(i *Input){i.FeeCoefficient=-1}},
 }
 for _,tt:=range tests{t.Run(tt.name,func(t *testing.T){i:=fixture();tt.change(&i);if v:=Evaluate(i);v.Decision!="NO_TRADE"{t.Fatalf("%+v",v)}})}
}
