package research100
import("testing")
func TestRound095SensitiveReportKeyBlocked(t *testing.T){
 if e:=RejectSecretShapedJSONKeys([]byte(`{"performance":{"trading_authorized":false,"samples":0}}`));e!=nil{t.Fatal(e)}
 if e:=RejectSecretShapedJSONKeys([]byte(`{"nested":[{"telegram_bot_token":"not-a-real-token"}]}`));e==nil{t.Fatal("secret-shaped key accepted")}
}
