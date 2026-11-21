package research100

import "testing"

func TestRound096PublicAggregateRejectsPrivateFields(t *testing.T) {
 if e:=CheckPublicAggregateFields([]byte(`{"status":"NO_TRADE","trading_authorized":false,"groups":[{"samples":0}]}`));e!=nil{t.Fatal(e)}
 for _,raw:=range []string{
  `{"groups":[{"condition_id":"private"}]}`,
  `{"records":[{"wallet_address":"private"}]}`,
  `{"telegram_bot_token":"private"}`,
 } {
  if e:=CheckPublicAggregateFields([]byte(raw));e==nil{t.Fatalf("private field leaked: %s",raw)}
 }
}
