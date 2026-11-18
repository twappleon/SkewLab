package polymarketv2
import("testing";"time")
func TestPaperAccountLimitsAndSettlement(t *testing.T){
 now:=time.Date(2026,11,18,0,0,0,0,time.UTC)
 q:=Quote{TokenID:"token",BuyVWAP:0.5,SellVWAP:0.49,Executable:true}
 a:=NewPaperAccount()
 x:=Assessment{Decision:"PAPER_CANDIDATE_ONLY",Shares:9.5,ExpectedCostsUSD:0.1}
 if err:=a.Open(now,"p1","eventA",q,x);err!=nil{t.Fatal(err)}
 if err:=a.Open(now,"p2","eventA",q,x);err==nil{t.Fatal("allowed correlated position")}
 if err:=a.Open(now,"p1","eventB",q,x);err==nil{t.Fatal("allowed duplicate id")}
 if err:=a.Resolve("p1",false,now.Add(time.Hour));err!=nil{t.Fatal(err)}
 if err:=a.Resolve("p1",true,now.Add(2*time.Hour));err!=nil{t.Fatal(err)}
 if a.Positions[0].Won||a.RealizedPnLUSD()!=-4.85||a.CashUSD!=95.15{t.Fatalf("inconsistent accounting %+v",a)}
 if err:=a.Open(now.Add(3*time.Hour),"p3","eventB",q,x);err==nil{t.Fatal("daily loss cutoff did not fire")}
 if err:=a.Resolve("unknown",true,now);err==nil{t.Fatal("unknown settlement accepted")}
}
