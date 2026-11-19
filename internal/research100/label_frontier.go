package research100
import("time")
type LabelFrontier struct{
 AsOf time.Time `json:"as_of"`
 Published int `json:"published"`
 NotYetPublished int `json:"not_yet_published"`
 FutureCaptures int `json:"future_captures"`
 LatestKnownPublication *time.Time `json:"latest_known_publication,omitempty"`
 SourceTrust string `json:"source_trust"`
}
func InspectLabelFrontier(d Dataset,asOf time.Time)LabelFrontier{
 out:=LabelFrontier{AsOf:asOf,SourceTrust:"Archive records only; independently authenticated Chainlink proof NOT supplied"}
 for _,o:=range d.Items{
  if o.CapturedAt.After(asOf){out.FutureCaptures++;continue}
  if o.Winner==""||o.ResolvedAt.After(asOf){out.NotYetPublished++;continue}
  out.Published++
  if out.LatestKnownPublication==nil||o.ResolvedAt.After(*out.LatestKnownPublication){v:=o.ResolvedAt;out.LatestKnownPublication=&v}
 }
 return out
}
