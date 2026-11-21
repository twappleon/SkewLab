package polymarketv2

// Read-only CLOB orderbook, conservative fees, and evidence-gated paper candidates.
import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "math"
 "net/http"
 "net/url"
 "sort"
 "strconv"
 "time"
)
const CLOB="https://clob.polymarket.com"
type Level struct {Price string `json:"price"`;Size string `json:"size"`}
type Book struct {AssetID string `json:"asset_id"`;Timestamp string `json:"timestamp"`;Bids []Level `json:"bids"`;Asks []Level `json:"asks"`}
type Quote struct {TokenID string `json:"token_id"`;ObservedAt time.Time `json:"observed_at"`;BuyVWAP float64 `json:"buy_vwap"`;SellVWAP float64 `json:"sell_vwap"`;Executable bool `json:"executable"`;Reason string `json:"reason"`}
func FetchBook(ctx context.Context,client *http.Client,token string)(Book,error){
 if token==""{return Book{},errors.New("missing token")}
 if client==nil{client=&http.Client{Timeout:8*time.Second}}
 req,err:=http.NewRequestWithContext(ctx,"GET",CLOB+"/book?token_id="+url.QueryEscape(token),nil);if err!=nil{return Book{},err}
 r,err:=client.Do(req);if err!=nil{return Book{},err};defer r.Body.Close()
 if r.StatusCode!=200{return Book{},fmt.Errorf("CLOB HTTP %d",r.StatusCode)}
 raw,err:=io.ReadAll(io.LimitReader(r.Body,2<<20));if err!=nil{return Book{},err}
 var b Book;if err=json.Unmarshal(raw,&b);err!=nil{return Book{},err}
 if b.AssetID!=token{return Book{},errors.New("token mismatch")}
 return b,nil
}
func parseTime(raw string)(time.Time,error){
 if n,err:=strconv.ParseInt(raw,10,64);err==nil{if n>100000000000{return time.UnixMilli(n).UTC(),nil};return time.Unix(n,0).UTC(),nil}
 return time.Parse(time.RFC3339Nano,raw)
}
func finite(v float64)bool{return !math.IsNaN(v)&&!math.IsInf(v,0)}
func vwap(levels []Level,shares float64,ascending bool)(float64,error){
 if shares<=0||!finite(shares){return 0,errors.New("invalid shares")}
 type row struct{price,size float64};rows:=[]row{}
 for _,l:=range levels{p,e:=strconv.ParseFloat(l.Price,64);if e!=nil{return 0,e};s,e:=strconv.ParseFloat(l.Size,64);if e!=nil{return 0,e};if p<=0||p>=1||s<=0||!finite(p)||!finite(s){return 0,errors.New("invalid level")};rows=append(rows,row{p,s})}
 sort.Slice(rows,func(i,j int)bool{if ascending{return rows[i].price<rows[j].price};return rows[i].price>rows[j].price})
 left,total:=shares,0.0
 for _,l:=range rows{fill:=math.Min(left,l.size);total+=fill*l.price;left-=fill;if left<=1e-9{break}}
 if left>1e-9{return 0,errors.New("insufficient visible depth")}
 return total/shares,nil
}
func Evaluate(b Book,now time.Time,shares float64)Quote{
 q:=Quote{TokenID:b.AssetID,ObservedAt:now.UTC(),Reason:"NO_TRADE"}
 stamp,err:=parseTime(b.Timestamp);if err!=nil||stamp.After(now.Add(3*time.Second))||now.Sub(stamp)>10*time.Second{q.Reason="stale or invalid timestamp";return q}
 buy,err:=vwap(b.Asks,shares,true);if err!=nil{q.Reason="ask depth invalid";return q}
 sell,err:=vwap(b.Bids,shares,false);if err!=nil{q.Reason="bid depth invalid";return q}
 if sell>=buy{q.Reason="crossed or locked book";return q}
 q.BuyVWAP=buy;q.SellVWAP=sell;q.Executable=true;q.Reason="visible depth only, no actual fill";return q
}
type ProbabilityEvidence struct {Probability float64 `json:"probability"`;Lower95 float64 `json:"lower_95"`;SampleSize int `json:"sample_size"`;ResolvedBefore time.Time `json:"resolved_before"`;IndependentSource string `json:"independent_source"`;OutOfSample bool `json:"out_of_sample"`}
func(p ProbabilityEvidence)Valid(now time.Time)bool{return p.SampleSize>=100&&p.OutOfSample&&p.IndependentSource!=""&&!p.ResolvedBefore.IsZero()&&p.ResolvedBefore.Before(now)&&finite(p.Probability)&&finite(p.Lower95)&&p.Probability>=0&&p.Probability<=1&&p.Lower95>=0&&p.Lower95<=p.Probability}
func Fee(shares,price,coefficient float64)float64{return shares*coefficient*price*(1-price)}
type Assessment struct {Decision string `json:"decision"`;Reason string `json:"reason"`;NetEVUSD float64 `json:"net_ev_usd"`;ConservativeEdge float64 `json:"conservative_edge"`;ExpectedCostsUSD float64 `json:"expected_costs_usd"`;Shares float64 `json:"shares"`}
func Assess(q Quote,p ProbabilityEvidence,now time.Time,coefficient,stake float64)Assessment{
 a:=Assessment{Decision:"NO_TRADE",Reason:"independent calibrated forecast absent or book invalid"}
 if !q.Executable||!p.Valid(now)||stake<=0||stake>5||coefficient<0||coefficient>0.1||!finite(coefficient){return a}
 // Fee coefficient must be externally checked for the market and time.
 perShareFee:=coefficient*q.BuyVWAP*(1-q.BuyVWAP)
 shares:=stake/(q.BuyVWAP+perShareFee+0.002+0.0025*q.BuyVWAP)
 cost:=Fee(shares,q.BuyVWAP,coefficient)+shares*0.002+shares*q.BuyVWAP*0.0025
 a.Shares=shares;a.ExpectedCostsUSD=cost
 a.NetEVUSD=shares*(p.Probability-q.BuyVWAP)-cost
 a.ConservativeEdge=p.Lower95-q.BuyVWAP-cost/shares
 if a.NetEVUSD>0&&a.ConservativeEdge>=0.05{a.Decision="PAPER_CANDIDATE_ONLY";a.Reason="requires verified settlement and prospective trial"}else{a.Reason="no conservative 5 percentage point net edge"}
 return a
}
