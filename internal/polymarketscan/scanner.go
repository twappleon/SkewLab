// Package polymarketscan only reads public market data; no order signing or execution.
package polymarketscan

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "math"
 "net/http"
 "net/url"
 "strconv"
 "strings"
 "time"
)

const GammaURL = "https://gamma-api.polymarket.com"

type Market struct {
 ID string `json:"id"`
 Question string `json:"question"`
 Slug string `json:"slug"`
 Active bool `json:"active"`
 Closed bool `json:"closed"`
 AcceptingOrders bool `json:"acceptingOrders"`
 EndDate string `json:"endDate"`
 LiquidityNum float64 `json:"liquidityNum"`
 Volume24hr float64 `json:"volume24hr"`
 BestBid float64 `json:"bestBid"`
 BestAsk float64 `json:"bestAsk"`
 ClobTokenIDs json.RawMessage `json:"clobTokenIds"`
}
type Candidate struct {
 ID string `json:"market_id"`
 Question string `json:"question"`
 URL string `json:"url"`
 Bid float64 `json:"best_bid"`
 Ask float64 `json:"best_ask"`
 Spread float64 `json:"spread"`
 LiquidityUSD float64 `json:"liquidity_usd"`
 Volume24hUSD float64 `json:"volume_24h_usd"`
 Decision string `json:"decision"`
 Reason string `json:"reason"`
}
type Report struct {
 GeneratedAt time.Time `json:"generated_at"`
 Mode string `json:"mode"`
 BankrollUSD float64 `json:"bankroll_usd"`
 MaxPositionUSD float64 `json:"max_position_usd"`
 MaxOpenExposureUSD float64 `json:"max_open_exposure_usd"`
 DailyLossLimitUSD float64 `json:"daily_loss_limit_usd"`
 LiveTradingEnabled bool `json:"live_trading_enabled"`
 Candidates []Candidate `json:"candidates"`
 RejectedCount int `json:"rejected_count"`
 Note string `json:"note"`
}
type Policy struct {
 BankrollUSD float64
 MaxPositionUSD float64
 MaxOpenExposureUSD float64
 DailyLossLimitUSD float64
 MinLiquidityUSD float64
 MinVolume24hUSD float64
 MaxSpread float64
}
func DefaultPolicy() Policy {
 return Policy{BankrollUSD:100,MaxPositionUSD:5,MaxOpenExposureUSD:30,DailyLossLimitUSD:2,MinLiquidityUSD:1000,MinVolume24hUSD:100,MaxSpread:0.03}
}
func (p Policy) Validate() error {
 if p.BankrollUSD!=100||p.MaxPositionUSD<=0||p.MaxPositionUSD>5||p.MaxOpenExposureUSD<=0||p.MaxOpenExposureUSD>30||p.DailyLossLimitUSD<=0||p.DailyLossLimitUSD>2 {return errors.New("risk policy violates $100 research envelope")}
 if p.MinLiquidityUSD<0||p.MinVolume24hUSD<0||p.MaxSpread<=0||p.MaxSpread>0.05{return errors.New("invalid market filter")}
 return nil
}
func Fetch(ctx context.Context, client *http.Client, limit int)([]Market,error){
 if limit<=0||limit>100{return nil,errors.New("limit must be 1..100")}
 if client==nil{client=&http.Client{Timeout:10*time.Second}}
 q:=url.Values{}
 q.Set("active","true");q.Set("closed","false");q.Set("limit",strconv.Itoa(limit));q.Set("order","volume24hr");q.Set("ascending","false")
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,GammaURL+"/markets?"+q.Encode(),nil)
 if err!=nil{return nil,err}
 resp,err:=client.Do(req);if err!=nil{return nil,fmt.Errorf("Gamma fetch failed: %w",err)}
 defer resp.Body.Close()
 if resp.StatusCode!=http.StatusOK{return nil,fmt.Errorf("Gamma status %d",resp.StatusCode)}
 b,err:=io.ReadAll(io.LimitReader(resp.Body,3<<20));if err!=nil{return nil,err}
 var markets []Market
 if err:=json.Unmarshal(b,&markets);err!=nil{return nil,fmt.Errorf("Gamma schema: %w",err)}
 return markets,nil
}
func RejectReason(now time.Time,m Market,p Policy)string{
 if !m.Active||m.Closed||!m.AcceptingOrders{return "inactive or orders disabled"}
 if m.ID==""||strings.TrimSpace(m.Question)==""||strings.TrimSpace(m.Slug)==""{return "incomplete market identity"}
 expiry,err:=time.Parse(time.RFC3339,m.EndDate)
 if err!=nil{expiry,err=time.Parse("2006-01-02",m.EndDate)}
 if err!=nil||!expiry.After(now.Add(time.Hour)){return "expired or imminent"}
 if math.IsNaN(m.BestBid)||math.IsNaN(m.BestAsk)||m.BestBid<=0||m.BestAsk>=1||m.BestAsk<=m.BestBid{return "invalid quote"}
 if m.BestAsk-m.BestBid>p.MaxSpread{return "wide spread"}
 if m.LiquidityNum<p.MinLiquidityUSD||m.Volume24hr<p.MinVolume24hUSD{return "insufficient liquidity or volume"}
 var tokens []string
 if len(m.ClobTokenIDs)==0||json.Unmarshal(m.ClobTokenIDs,&tokens)!=nil||len(tokens)!=2||tokens[0]==""||tokens[1]==""{return "missing binary tokens"}
 return ""
}
func Scan(now time.Time,markets []Market,p Policy)(Report,error){
 if err:=p.Validate();err!=nil{return Report{},err}
 r:=Report{GeneratedAt:now.UTC(),Mode:"PUBLIC_READ_ONLY_NO_TRADES",BankrollUSD:p.BankrollUSD,MaxPositionUSD:p.MaxPositionUSD,MaxOpenExposureUSD:p.MaxOpenExposureUSD,DailyLossLimitUSD:p.DailyLossLimitUSD,LiveTradingEnabled:false,Candidates:make([]Candidate,0),Note:"Discovery only. Gamma quotes are not executable. Probability, CLOB depth, fees, settlement terms, legal and geoblock eligibility unverified. NO_TRADE."}
 for _,m:=range markets{
  if RejectReason(now,m,p)!=""{r.RejectedCount++;continue}
  r.Candidates=append(r.Candidates,Candidate{ID:m.ID,Question:m.Question,URL:"https://polymarket.com/event/"+url.PathEscape(m.Slug),Bid:m.BestBid,Ask:m.BestAsk,Spread:m.BestAsk-m.BestBid,LiquidityUSD:m.LiquidityNum,Volume24hUSD:m.Volume24hr,Decision:"NO_TRADE",Reason:"No independently calibrated probability or verified executable price/cost, legal eligibility or settlement audit"})
 }
 return r,nil
}
