package research100
import("time")
type NearExpiryWitness struct{
 Evaluated bool `json:"evaluated"`
 WithinExitBuffer bool `json:"within_exit_buffer"`
 SecondsToExpiry float64 `json:"seconds_to_expiry"`
 Status string `json:"status"`
}
func EvaluateNearExpiryWitness(pair ReplayPair)NearExpiryWitness{
 out:=NearExpiryWitness{Status:"INVALID_OR_MISSING_EXIT_WITNESS"}
 if !ValidateExitWitness(pair.Entry,pair.Exit,pair.Exit.QuoteAt).Valid{return out}
 remaining:=pair.Entry.ExpiresAt.Sub(pair.Exit.QuoteAt)
 out.Evaluated=true;out.SecondsToExpiry=remaining.Seconds()
 out.WithinExitBuffer=remaining<=32*time.Second
 out.Status="OBSERVED_TIME_ONLY_NOT_EXIT_FILL"
 return out
}
