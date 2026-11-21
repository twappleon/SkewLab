package polymarketv2
import (
 "errors"
 "math"
 "sort"
 "time"
)
// LabeledOutcome represents a real, externally resolved event, never a
// synthetic label or a post-decision correction to historical forecasts.
type LabeledOutcome struct {
 Cohort string `json:"cohort"`
 OccurredAt time.Time `json:"occurred_at"`
 ResolvedAt time.Time `json:"resolved_at"`
 Outcome bool `json:"outcome"`
 Source string `json:"source"`
}
type Calibration struct {
 Cohort string `json:"cohort"`
 TrainingCount int `json:"training_count"`
 ForwardCount int `json:"forward_count"`
 TrainingWinRate float64 `json:"training_win_rate"`
 ForwardBrier float64 `json:"forward_brier"`
 ForwardBaselineBrier float64 `json:"forward_baseline_brier"`
 Forecast ProbabilityEvidence `json:"forecast"`
 Status string `json:"status"`
}
// Calibrate applies a conservative cohort base-rate prior. This is a weak
// baseline, NOT a text/event model. Its probability cannot be accepted until
// 100 earlier resolutions and at least 30 later out-of-sample outcomes exist.
// No in-sample data may leak across the time boundary.
func Calibrate(cohort string,now time.Time,observations []LabeledOutcome)(Calibration,error){
 result:=Calibration{Cohort:cohort,Status:"INSUFFICIENT_EVIDENCE"}
 if cohort==""{return result,errors.New("missing cohort")}
 var records []LabeledOutcome
 for _,o:=range observations{
  if o.Cohort!=cohort {continue}
  if o.Source==""||o.OccurredAt.IsZero()||o.ResolvedAt.IsZero()||o.ResolvedAt.Before(o.OccurredAt)||!o.ResolvedAt.Before(now){return result,errors.New("invalid or future resolution")}
  records=append(records,o)
 }
 sort.Slice(records,func(i,j int)bool{return records[i].ResolvedAt.Before(records[j].ResolvedAt)})
 if len(records)<130{return result,nil}
 // The first 100 older resolved outcomes form the training set. All later
 // outcomes are prospective relative to the fixed initial training cutoff.
 train:=records[:100];future:=records[100:]
 cutoff:=train[len(train)-1].ResolvedAt
 for _,o:=range future {if !o.OccurredAt.After(cutoff){return result,errors.New("overlapping event chronology")}}
 wins:=0;for _,o:=range train{if o.Outcome{wins++}}
 // Wilson lower bound with 95% z-score for a binary rate.
 n:=float64(len(train));phat:=float64(wins)/n;z:=1.96
 lower:=(phat+z*z/(2*n)-z*math.Sqrt(phat*(1-phat)/n+z*z/(4*n*n)))/(1+z*z/n)
 // Compare frozen Laplace-smoothed base rate with naive 50/50 baseline.
 prob:=(float64(wins)+1)/(n+2)
 sum,baseline:=0.0,0.0
 for _,o:=range future{y:=0.0;if o.Outcome{y=1};sum+=(prob-y)*(prob-y);baseline+=(0.5-y)*(0.5-y)}
 result.TrainingCount=len(train);result.ForwardCount=len(future);result.TrainingWinRate=phat
 result.ForwardBrier=sum/float64(len(future));result.ForwardBaselineBrier=baseline/float64(len(future))
 result.Forecast=ProbabilityEvidence{Probability:prob,Lower95:math.Min(lower,prob),SampleSize:len(train),ResolvedBefore:cutoff,IndependentSource:"independent resolved cohort base-rate (weak baseline)",OutOfSample:true}
 if result.ForwardBrier>=result.ForwardBaselineBrier{result.Status="NO_EDGE_VS_50_50_BASELINE";return result,nil}
 result.Status="RESEARCH_BASELINE_ONLY"
 // Cohort priors have no event-specific information, and NEVER grant live
 // execution or authorization to trade a specific market.
 return result,nil
}
