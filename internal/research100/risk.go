package research100

// RiskDecision distinguishes a stop-opening rule from an actual loss cap.
// A $5 prediction-market purchase can lose $5, even with a $2 daily stop.
type RiskDecision struct {
	Status                    string  `json:"status"`
	AccountUSD                float64 `json:"account_usd"`
	ProposedStakeUSD          float64 `json:"proposed_stake_usd"`
	WorstCaseOpenLossUSD      float64 `json:"worst_case_open_loss_usd"`
	RemainingDailyLossRoomUSD float64 `json:"remaining_daily_loss_room_usd"`
	TradingAuthorized         bool    `json:"trading_authorized"`
}

func AssessWorstCaseRisk(bankroll, stake, openExposure, realizedDailyLoss, dailyLimit float64) RiskDecision {
	d := RiskDecision{Status: "INVALID_RISK_BUDGET", AccountUSD: bankroll, ProposedStakeUSD: stake, TradingAuthorized: false}
	if !finite(bankroll) || !finite(stake) || !finite(openExposure) || !finite(realizedDailyLoss) || !finite(dailyLimit) || bankroll <= 0 || stake <= 0 || openExposure < 0 || realizedDailyLoss < 0 || dailyLimit <= 0 {
		return d
	}
	d.WorstCaseOpenLossUSD = stake + openExposure
	d.RemainingDailyLossRoomUSD = dailyLimit - realizedDailyLoss
	if stake > bankroll || d.WorstCaseOpenLossUSD > d.RemainingDailyLossRoomUSD {
		d.Status = "WORST_CASE_EXCEEDS_DAILY_STOP"
		return d
	}
	d.Status = "WITHIN_RESEARCH_BUDGET_NOT_TRADE_AUTHORIZATION"
	return d
}
