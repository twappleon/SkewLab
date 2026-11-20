// bdd-gate validates documented Given/When/Then scenarios against Go tests.
// It does not count CI or historical engineering acceptance as BDD acceptance.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Entry struct {
	Round          int    `json:"round"`
	Status         string `json:"status"`
	Feature        string `json:"feature,omitempty"`
	TestFile       string `json:"test_file,omitempty"`
	TestFunction   string `json:"test_function,omitempty"`
	SourceFile     string `json:"source_file,omitempty"`
 SourceFiles []string `json:"source_files,omitempty"`
 TestFiles []string `json:"test_files,omitempty"`
	Revision       int    `json:"revision,omitempty"`
	RevisionReason string `json:"revision_reason,omitempty"`
}
type Manifest struct {
	SchemaVersion int     `json:"schema_version"`
	Rounds        []Entry `json:"rounds"`
}
type Summary struct {
	Total            int  `json:"total"`
	Verified         int  `json:"verified"`
	Ready            int  `json:"ready_for_ci"`
	Pending          int  `json:"pending_bdd"`
	EngineeringCount int  `json:"engineering_accepted"`
	BDDComplete      bool `json:"bdd_complete"`
}

var validTag = regexp.MustCompile(`(?m)^\s*@round-(\d{3})(?:\s|$)`)

func localPath(root, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || strings.Contains(name, "\\") {
		return "", errors.New("empty or absolute/invalid path")
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("path traversal forbidden")
	}
	return filepath.Join(root, clean), nil
}
func parseEntry(root string, e Entry) error {
	feature, err := localPath(root, e.Feature)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(e.Feature, "docs/bdd/") || !strings.HasSuffix(e.Feature, ".feature") {
		return errors.New("feature path must be docs/bdd/*.feature")
	}
	content, err := os.ReadFile(feature)
	if err != nil {
		return fmt.Errorf("feature unreadable: %w", err)
	}
	s := string(content)
	tag := fmt.Sprintf("%03d", e.Round)
	foundTag := false
	for _, m := range validTag.FindAllStringSubmatch(s, -1) {
		if m[1] == tag {
			foundTag = true
		}
	}
	if !foundTag {
		return fmt.Errorf("missing @round-%s", tag)
	}
	if !strings.Contains(s, "Feature:") {
		return errors.New("no Feature: declaration")
	}
	// Check each Scenario (or Scenario Outline) independently, rather than
	// allowing Given from one scenario and Then from another to satisfy the gate.
	lines := strings.Split(s, "\n")
	scenarios := 0
	inside := false
	given, when, then := false, false, false
	flush := func() error {
		if inside && (!given || !when || !then) {
			return errors.New("scenario missing Given, When or Then")
		}
		return nil
	}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "Scenario:") || strings.HasPrefix(line, "Scenario Outline:") {
			if err := flush(); err != nil {
				return err
			}
			scenarios++
			inside = true
			given, when, then = false, false, false
			continue
		}
		if !inside {
			continue
		}
		switch {
		case strings.HasPrefix(line, "Given "):
			given = true
		case strings.HasPrefix(line, "When "):
			when = true
		case strings.HasPrefix(line, "Then "):
			then = true
		}
	}
	if err := flush(); err != nil {
		return err
	}
	if scenarios == 0 {
		return errors.New("feature contains no scenario")
	}
	if e.Revision < 1 || strings.TrimSpace(e.RevisionReason) == "" {
		return errors.New("missing BDD revision and change rationale")
	}
	source, err := localPath(root, e.SourceFile)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(e.SourceFile, ".go") {
		return errors.New("source_file must be Go implementation")
	}
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source missing: %w", err)
	}
    for _,extra := range append(append([]string{}, e.SourceFiles...), e.TestFiles...){
      full,err:=localPath(root,extra); if err!=nil{return err}
      if !isBehavioralFile(extra){return fmt.Errorf("extra mapping is not behavior source: %s",extra)}
      if _,err:=os.Stat(full);err!=nil{return fmt.Errorf("extra mapping missing %s: %w",extra,err)}
    }
	tf, err := localPath(root, e.TestFile)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(e.TestFile, "_test.go") || !strings.HasPrefix(e.TestFunction, "Test") || e.TestFunction == "Test" {
		return errors.New("invalid Go test mapping")
	}
	file, err := parser.ParseFile(token.NewFileSet(), tf, nil, 0)
	if err != nil {
		return fmt.Errorf("test parser error: %w", err)
	}
	found := false
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Recv == nil && f.Name.Name == e.TestFunction {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Go test function %s not found", e.TestFunction)
	}
	return nil
}
func validate(root string, m Manifest) (Summary, error) {
	out := Summary{Total: 100, EngineeringCount: 100}
	if m.SchemaVersion != 1 {
		return out, errors.New("unsupported BDD manifest schema")
	}
	if len(m.Rounds) != 100 {
		return out, fmt.Errorf("BDD manifest must index exactly 100 historical rounds; found %d", len(m.Rounds))
	}
	seen := map[int]bool{}
	for _, e := range m.Rounds {
		if e.Round < 1 || e.Round > 100 || seen[e.Round] {
			return out, fmt.Errorf("invalid or duplicate round %d", e.Round)
		}
		seen[e.Round] = true
		switch e.Status {
		case "PENDING_BDD":
			out.Pending++
		case "READY_FOR_CI":
			out.Ready++
			if err := parseEntry(root, e); err != nil {
				return out, fmt.Errorf("round %03d: %w", e.Round, err)
			}
		case "VERIFIED":
			out.Verified++
			if err := parseEntry(root, e); err != nil {
				return out, fmt.Errorf("round %03d: %w", e.Round, err)
			}
		default:
			return out, fmt.Errorf("round %03d: unsupported status %q", e.Round, e.Status)
		}
	}
	out.BDDComplete = out.Verified == 100
	return out, nil
}
// isBehavioralFile covers executable code, UI and deployment behavior, not
// only Go. Test-only changes also need BDD scenario review.
func isBehavioralFile(path string) bool {
 if strings.HasPrefix(path,"cmd/bdd-gate/"){return false}
 if path=="Dockerfile" || path=="railway.toml"{return true}
 if !(strings.HasPrefix(path,"internal/") || strings.HasPrefix(path,"cmd/") || strings.HasPrefix(path,"web/") || strings.HasPrefix(path,"frontend/") || strings.HasPrefix(path,"scripts/") || strings.HasPrefix(path,"docker/")){return false}
 ext:=strings.ToLower(filepath.Ext(path))
 switch ext {case ".go",".js",".jsx",".ts",".tsx",".vue",".html",".css",".scss",".py",".sql",".sh",".dart",".gd":return true}
 return false
}
// validateChangedCode requires a matching revised Gherkin scenario and a
// revised Go regression test for EVERY behavior-affecting file in a push.
// SourceFiles and TestFiles allow a single reviewed scenario to span modules.
func validateChangedCode(m Manifest, changed []string) error {
 names:=map[string]bool{}
 for _,c:=range changed{names[c]=true}
 required:=[]string{}
 for _,c:=range changed{if isBehavioralFile(c){required=append(required,c)}}
 if len(required)==0{return nil}
 if !names["docs/bdd/manifest.json"]{return errors.New("behavior-affecting code changed without updating docs/bdd/manifest.json")}
 for _,source:=range required{
  found:=false
  for _,e:=range m.Rounds{
   match:=e.SourceFile==source || e.TestFile==source
   for _,x:=range e.SourceFiles{if x==source{match=true}}
   for _,x:=range e.TestFiles{if x==source{match=true}}
   if (e.Status=="READY_FOR_CI"||e.Status=="VERIFIED")&&match&&names[e.Feature]&&names[e.TestFile]{found=true;break}
  }
  if !found{return fmt.Errorf("behavior-affecting file %s changed without revised Given/When/Then and corresponding Go test",source)}
 }
 return nil
}
func changedFiles(base string) ([]string, error) {
	if base == "" {
		return nil, nil
	}
	// GitHub push.before can be all zeros when a new branch is created.
	if strings.Trim(base, "0") == "" {
		return nil, errors.New("new branch has no diff baseline; review manually")
	}
	cmd := exec.Command("git", "diff", "--name-only", "-z", base, "HEAD")
	raw, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot inspect code changes from %s: %w", base, err)
	}
	parts := strings.Split(string(raw), "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, filepath.ToSlash(p))
		}
	}
	sort.Strings(out)
	return out, nil
}
func run(root, base string, requireAll bool) (Summary, error) {
	path := filepath.Join(root, "docs", "bdd", "manifest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Summary{}, fmt.Errorf("BDD manifest unavailable: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Summary{}, err
	}
	s, err := validate(root, m)
	if err != nil {
		return s, err
	}
	if base != "" {
		changes, e := changedFiles(base)
		if e != nil {
			return s, e
		}
		if e = validateChangedCode(m, changes); e != nil {
			return s, e
		}
	}
	if requireAll && !s.BDDComplete {
		return s, fmt.Errorf("BDD acceptance incomplete: %d/100 verified; %d pending; %d ready", s.Verified, s.Pending, s.Ready)
	}
	return s, nil
}
func main() {
	root := flag.String("root", ".", "repository root")
	base := flag.String("base", "", "git commit before the pushed changes, to verify changed Go code has matching BDD revision")
	requireAll := flag.Bool("require-all", false, "fail while any historical round BDD remains unverified")
	flag.Parse()
	s, err := run(*root, *base, *requireAll)
	b, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(b))
	if err != nil {
		fmt.Fprintln(os.Stderr, "BDD gate FAILED:", err)
		os.Exit(1)
	}
	fmt.Println("BDD scenario integrity and source/test traceability PASS; BDD complete is true only when all 100 scenarios are VERIFIED")
}
