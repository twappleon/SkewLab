package research100
import("math")
type ProbabilityClip struct{
 Input float64 `json:"input"`
 Clipped float64 `json:"clipped"`
 Accepted bool `json:"accepted"`
 Status string `json:"status"`
}
// ClipProbability avoids log(0) under pre-specified epsilon and rejects
// nonfinite forecasts. Clipping cannot manufacture prediction skill.
func ClipProbability(input,epsilon float64)ProbabilityClip{
 out:=ProbabilityClip{Input:input,Status:"INVALID_PROBABILITY"}
 if !finite(input)||input<0||input>1||!finite(epsilon)||epsilon<=0||epsilon>=.5{return out}
 out.Clipped=math.Max(epsilon,math.Min(1-epsilon,input))
 out.Accepted=true;out.Status="VALID_PROBABILITY_WITH_PREDECLARED_CLIPPING"
 return out
}
