package research100

import (
 "encoding/json"
 "errors"
 "strings"
)

// CheckPublicAggregateFields only checks JSON field names, not values.
// It is a defense-in-depth guard, not proof a payload contains no PII.
func CheckPublicAggregateFields(raw []byte) error {
 if err:=RejectSecretShapedJSONKeys(raw);err!=nil{return err}
 var root any
 if err:=json.Unmarshal(raw,&root);err!=nil{return err}
 forbidden:=map[string]bool{
  "condition_id":true,"token_id":true,"wallet_address":true,
  "market_slug":true,"trade_id":true,"chat_id":true,
  "raw_observations":true,"recent_private_observations":true,
  "private_positions":true,"api_secret":true,
 }
 var walk func(any) error
 walk=func(x any) error {
  switch v:=x.(type) {
  case map[string]any:
   for key,value:=range v {
    if forbidden[strings.ToLower(strings.TrimSpace(key))] {
     return errors.New("private identifier field detected in aggregate")
    }
    if err:=walk(value);err!=nil{return err}
   }
  case []any:
   for _,value:=range v{if err:=walk(value);err!=nil{return err}}
  }
  return nil
 }
 return walk(root)
}
