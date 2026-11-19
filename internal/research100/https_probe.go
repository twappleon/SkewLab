package research100

import (
 "context"
 "errors"
 "net/http"
 "net/url"
 "strings"
 "time"
)

type HTTPSProbe struct {
 Path string `json:"path"`
 HTTPStatus int `json:"http_status"`
 Passed bool `json:"passed"`
 Status string `json:"status"`
}

// ProbeHTTPSReadOnly performs only GET, without auth headers, cookies or
// credentials; it cannot claim private routes are authenticated for all users.
func ProbeHTTPSReadOnly(ctx context.Context,client *http.Client,rawURL string,expectedStatus int) HTTPSProbe {
 out:=HTTPSProbe{Status:"INVALID_HTTPS_PROBE"}
 target,e:=url.Parse(rawURL)
 if e!=nil||target.Scheme!="https"||target.Host==""||target.User!=nil||target.RawQuery!=""||target.Fragment!=""||strings.Contains(target.Host,"@")||expectedStatus<100||expectedStatus>599{return out}
 out.Path=target.EscapedPath()
 if client==nil{client=&http.Client{Timeout:5*time.Second,CheckRedirect:func(*http.Request,[]*http.Request)error{return http.ErrUseLastResponse}}}
 req,e:=http.NewRequestWithContext(ctx,http.MethodGet,rawURL,nil)
 if e!=nil{return out}
 req.Header.Set("Cache-Control","no-store")
 response,e:=client.Do(req)
 if e!=nil{out.Status="UNREACHABLE_OR_TLS_ERROR";return out}
 response.Body.Close()
 out.HTTPStatus=response.StatusCode
 out.Passed=response.StatusCode==expectedStatus
 out.Status="UNEXPECTED_HTTP_STATUS"
 if out.Passed{out.Status="HTTP_STATUS_MATCHED_READ_ONLY"}
 return out
}
