package main
import("context";"encoding/json";"log";"net/http";"time")
func(a *App) comparisonRun(ctx context.Context){
 ticker:=time.NewTicker(time.Minute);defer ticker.Stop()
 for{
  select{case <-ctx.Done():return
  case <-ticker.C:
   if a.comparison==nil{continue}
   for _,s:=range a.comparison.Due(time.Now().UTC(),8){
    if ctx.Err()!=nil{return}
    outcome,err:=a.api.Settlement(ctx,s.Slug,s.ConditionID);if err!=nil{continue}
    if outcome.Winner!="UP"&&outcome.Winner!="DOWN"{continue}
    if err=a.comparison.Resolve(s.Slug,s.ConditionID,outcome.Winner,time.Now().UTC());err!=nil{
     log.Printf("comparison research outcome write failure: %v",err);break
    }
   }
  }
 }
}
func(a *App) comparisonPrivate(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("X-Content-Type-Options","nosniff")
 items:=a.comparison.Snapshot();if len(items)>200{items=items[len(items)-200:]}
 _=json.NewEncoder(w).Encode(map[string]any{
 "study_only":true,"real_money_trades":false,"paper_entries_authorized":false,
 "summary":a.comparison.Summary(time.Now().UTC()),
 "recent_private_observations":items,
 "warning":"Gamma outcomes and modeled costs are NOT independent Chainlink proof or executable real fills",
 })
}
