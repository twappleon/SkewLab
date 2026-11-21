package research100
type StakeEnvelope struct{
 Stakes int `json:"stakes"`
 TotalWorstCaseLossUSD float64 `json:"total_worst_case_loss_usd"`
 WithinDailyTwoDollarLoss bool `json:"within_daily_two_dollar_loss"`
 WithinThirtyDollarExposure bool `json:"within_thirty_dollar_exposure"`
 Status string `json:"status"`
 TradingAuthorized bool `json:"trading_authorized"`
}
func CheckStakeEnvelope(stakes []float64)StakeEnvelope{
 out:=StakeEnvelope{Stakes:len(stakes),Status:"INVALID_STAKE_ENVELOPE",TradingAuthorized:false}
 for _,stake:=range stakes{
  if !finite(stake)||stake<=0||stake>5{return out}
  out.TotalWorstCaseLossUSD+=stake
 }
 out.WithinDailyTwoDollarLoss=out.TotalWorstCaseLossUSD<=2
 out.WithinThirtyDollarExposure=out.TotalWorstCaseLossUSD<=30
 out.Status="RESEARCH_RISK_ENVELOPE_ONLY"
 if !out.WithinDailyTwoDollarLoss{out.Status="WORST_CASE_WOULD_BREACH_DAILY_STOP"}
 return out
}
