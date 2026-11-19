package research100
import("encoding/json";"errors";"strings")
func RejectSecretShapedJSONKeys(raw []byte)error{
 var value any
 if e:=json.Unmarshal(raw,&value);e!=nil{return e}
 suspicious:=[]string{"bot_token","telegram_token","password","api_key","private_key","secret_key","wallet_seed","mnemonic"}
 var walk func(any)error
 walk=func(v any)error{
  switch t:=v.(type){
  case map[string]any:
   for k,child:=range t{
    name:=strings.ToLower(k)
    for _,s:=range suspicious{if strings.Contains(name,s){return errors.New("sensitive key detected in report")}}
    if e:=walk(child);e!=nil{return e}
   }
  case []any:
   for _,child:=range t{if e:=walk(child);e!=nil{return e}}
  }
  return nil
 }
 return walk(value)
}
