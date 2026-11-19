package research100
import("time")
type ExternalOutcomeClaim struct{
 ID string `json:"id"`
 Winner string `json:"winner"`
 Source string `json:"source"`
 ObservedAt time.Time `json:"observed_at"`
}
type ClaimComparison struct{
 Comparable int `json:"comparable"`
 Matches int `json:"matches"`
 Disagreements int `json:"disagreements"`
 UnknownClaims int `json:"unknown_claims"`
 Status string `json:"status"`
 IndependentProofVerified bool `json:"independent_proof_verified"`
}
func CompareOutcomeClaims(d Dataset,claims []ExternalOutcomeClaim,asOf time.Time)ClaimComparison{
 out:=ClaimComparison{Status:"CLAIM_MATCHES_ARE_NOT_CRYPTOGRAPHIC_ORACLE_PROOF",IndependentProofVerified:false}
 known:=map[string]string{}
 for _,o:=range d.Items{if o.Winner!=""&&!o.ResolvedAt.After(asOf)&&!o.CapturedAt.After(asOf){known[o.ID]=o.Winner}}
 seen:=map[string]bool{}
 for _,c:=range claims{
  if c.ID==""||c.ObservedAt.IsZero()||c.ObservedAt.After(asOf)||seen[c.ID]{continue}
  seen[c.ID]=true
  if c.Winner!="UP"&&c.Winner!="DOWN"{out.UnknownClaims++;continue}
  expected,ok:=known[c.ID]
  if !ok{out.UnknownClaims++;continue}
  out.Comparable++
  if expected==c.Winner{out.Matches++}else{out.Disagreements++}
 }
 if out.Disagreements>0{out.Status="OUTCOME_CLAIM_CONFLICT_REQUIRES_REVIEW"}
 return out
}
