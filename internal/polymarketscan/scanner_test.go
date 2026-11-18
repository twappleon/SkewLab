package polymarketscan
import (
 "encoding/json"
 "testing"
 "time"
)
func fixture()Market{
 return Market{ID:"123",Question:"Test event?",Slug:"test-event",Active:true,AcceptingOrders:true,EndDate:"2027-12-30T00:00:00Z",LiquidityNum:2000,Volume24hr:500,BestBid:0.50,BestAsk:0.51,ClobTokenIDs:json.RawMessage(`["yes","no"]`)}
}
func TestFailClosed(t *testing.T){
 now:=time.Date(2026,11,18,0,0,0,0,time.UTC)
 r,err:=Scan(now,[]Market{fixture()},DefaultPolicy());if err!=nil{t.Fatal(err)}
 if r.LiveTradingEnabled||len(r.Candidates)!=1||r.Candidates[0].Decision!="NO_TRADE"{t.Fatalf("unsafe: %+v",r)}
}
func TestRejectUnsafeMarkets(t *testing.T){
 now:=time.Date(2026,11,18,0,0,0,0,time.UTC)
 edits:=[]func(*Market){
 func(m *Market){m.Active=false},
 func(m *Market){m.AcceptingOrders=false},
 func(m *Market){m.BestAsk=0.60},
 func(m *Market){m.BestBid=0},
 func(m *Market){m.LiquidityNum=0},
 func(m *Market){m.Volume24hr=0},
 func(m *Market){m.EndDate="2026-11-17T00:00:00Z"},
 func(m *Market){m.ClobTokenIDs=nil},
 }
 for i,edit:=range edits{m:=fixture();edit(&m);if RejectReason(now,m,DefaultPolicy())==""{t.Errorf("unsafe fixture %d accepted",i)}}
}
func TestBudgetCap(t *testing.T){
 p:=DefaultPolicy();p.MaxPositionUSD=6;if p.Validate()==nil{t.Fatal("position limit widened")}
 p=DefaultPolicy();p.BankrollUSD=101;if p.Validate()==nil{t.Fatal("bankroll widened")}
}
