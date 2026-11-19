// research100 produces a read-only, evidence-limited status report.
// It never signs or submits market orders and never modifies any archive.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/twappleon/SkewLab/internal/experiment"
	"github.com/twappleon/SkewLab/internal/research100"
	"io"
	"os"
	"time"
)

type Output struct {
	Mode              string                    `json:"mode"`
	TradingAuthorized bool                      `json:"trading_authorized"`
	DataSource        string                    `json:"data_source"`
	AsOf              time.Time                 `json:"as_of"`
	Status            string                    `json:"status"`
	Calibration       *research100.Calibration  `json:"quote_proxy_calibration,omitempty"`
	QuoteAge          *research100.QuoteAgeAudit `json:"quote_age,omitempty"`
	CaptureQuality    *research100.CaptureGate `json:"capture_quality,omitempty"`
	SettlementQuality *research100.SettlementReview `json:"settlement_quality,omitempty"`
	BookQuality *research100.BookReview `json:"book_quality,omitempty"`
	ProbabilityReview *research100.CalibrationReview `json:"probability_review,omitempty"`
	ForwardResearch *research100.ForwardResearchReview `json:"forward_research,omitempty"`
	CostResearch *research100.CostResearchReview `json:"cost_research,omitempty"`
	ExitResearch *research100.ExitResearchReview `json:"exit_research,omitempty"`
	LabelLag          []research100.LabelLag `json:"label_lag,omitempty"`
	DayDiversity      []research100.DayDiversity `json:"day_diversity,omitempty"`
	WalkForward       *research100.WalkForward  `json:"prospective_research,omitempty"`
	Coverage          *research100.Coverage     `json:"observed_coverage,omitempty"`
	Risk              *research100.RiskDecision `json:"conservative_100_usd_risk,omitempty"`
	Comparison        *experiment.Summary       `json:"fixed_time_comparison,omitempty"`
	Warning           string                    `json:"warning"`
}

func Execute(kind string, raw []byte, asOf time.Time) (Output, error) {
	out := Output{Mode: "READ_ONLY_RESEARCH_100_ROUNDS", TradingAuthorized: false, AsOf: asOf, DataSource: kind, Warning: "No live fills or profit proof. Gamma settlement is not independent oracle proof. No paper or real trades are authorized."}
	if len(raw) == 0 || len(raw) > 32<<20 {
		return out, errors.New("input absent or oversized")
	}
	switch kind {
	case "audit":
		var records []research100.Observation
		if e := json.Unmarshal(raw, &records); e != nil {
			return out, fmt.Errorf("invalid observation archive: %w", e)
		}
		d, e := research100.BuildDataset(records)
		if e != nil {
			return out, e
		}
		c := research100.ScoreCalibration(d, asOf)
		w := research100.RunWalkForward(d, asOf)
		cov := research100.AssessCoverage(d, 5*time.Minute, nil, nil)
		risk := research100.AssessWorstCaseRisk(100, 5, 0, 0, 2)
		qa := research100.AssessQuoteAge(d, asOf)
		gate := research100.CheckCaptureReadiness(d, asOf)
		out.QuoteAge = &qa
		out.CaptureQuality = &gate
		settlement := research100.ReviewSettlementData(d, asOf)
		out.SettlementQuality = &settlement
		book := research100.ReviewBookEvidence(d, asOf)
		out.BookQuality = &book
		probability := research100.ReviewCalibrationEvidence(d, asOf)
		out.ProbabilityReview = &probability
		forward := research100.ReviewForwardResearch(d, asOf)
		out.ForwardResearch = &forward
		costResearch := research100.ReviewCostResearch(d, asOf)
		out.CostResearch = &costResearch
		exitReview := research100.ReviewExitEvidence(d, nil, asOf)
		out.ExitResearch = &exitReview
		out.LabelLag = research100.AssessLabelLag(d, asOf)
		out.DayDiversity = research100.DayConcentration(d, asOf)
		out.Calibration = &c
		out.WalkForward = &w
		out.Coverage = &cov
		out.Risk = &risk
		out.Status = "INSUFFICIENT_EVIDENCE"
		if w.Status == "FORWARD_RESEARCH_SAMPLE_NOT_TRADE_APPROVAL" {
			out.Status = "RESEARCH_ONLY_NOT_TRADE_APPROVAL"
		}
	case "comparison":
		var records []experiment.Sample
		if e := json.Unmarshal(raw, &records); e != nil {
			return out, fmt.Errorf("invalid fixed-time archive: %w", e)
		}
		if len(records) > 50000 {
			return out, errors.New("fixed-time archive cap exceeded")
		}
		c := experiment.Compare(records, asOf)
		out.Comparison = &c
		out.Status = c.Status
	default:
		return out, errors.New("kind must be audit or comparison")
	}
	return out, nil
}
// ExecuteWithWitnesses only pairs explicitly supplied private post-entry quotes.
// Passing no witness bytes never creates synthetic fills.
func ExecuteWithWitnesses(kind string, raw, witnessRaw []byte, asOf time.Time) (Output,error){
 out,e:=Execute(kind,raw,asOf)
 if e!=nil{return out,e}
 if len(witnessRaw)==0{return out,nil}
 if kind!="audit"{return out,errors.New("exit witnesses require audit observation input")}
 if len(witnessRaw)>16<<20{return out,errors.New("witness input too large")}
 var entries []research100.Observation
 if e=json.Unmarshal(raw,&entries);e!=nil{return out,e}
 d,e:=research100.BuildDataset(entries);if e!=nil{return out,e}
 var witnesses []research100.ExitWitness
 if e=json.Unmarshal(witnessRaw,&witnesses);e!=nil{return out,fmt.Errorf("invalid exit witness archive: %w",e)}
 if len(witnesses)>50000{return out,errors.New("too many exit witnesses")}
 r:=research100.ReviewExitEvidence(d,witnesses,asOf)
 out.ExitResearch=&r
 return out,nil
}

func main() {
	file := flag.String("file", "", "existing private JSON research file; required")
	kind := flag.String("kind", "comparison", "comparison or audit")
	witnessPath := flag.String("exit-witness-file", "", "optional private post-entry quote archive; no synthetic fills")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file required; no synthetic archive fallback")
		os.Exit(2)
	}
	f, e := os.Open(*file)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	b, e := io.ReadAll(io.LimitReader(f, 32<<20+1))
	f.Close()
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	var witnessBytes []byte
	if *witnessPath!=""{
		wf,er:=os.Open(*witnessPath)
		if er!=nil{fmt.Fprintln(os.Stderr,"exit witness file unavailable:",er);os.Exit(1)}
		witnessBytes,er=io.ReadAll(io.LimitReader(wf,(16<<20)+1));wf.Close()
		if er!=nil||len(witnessBytes)>16<<20{fmt.Fprintln(os.Stderr,"exit witness archive unreadable or oversized");os.Exit(1)}
	}
	r, e := ExecuteWithWitnesses(*kind,b,witnessBytes,time.Now().UTC())
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if e = enc.Encode(r); e != nil {
		os.Exit(1)
	}
}
