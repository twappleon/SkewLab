package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) (string, Manifest) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, s string) {
		p := filepath.Join(dir, name)
		if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(p, []byte(s), 0600); e != nil {
			t.Fatal(e)
		}
	}
	write("internal/research100/quote_age.go", "package research100")
	write("internal/research100/quote_age_test.go", "package research100\nimport \"testing\"\nfunc TestRound011QuoteAgeHonorsAsOf(t *testing.T){}")
	write("docs/bdd/round011.feature", "@round-011\nFeature: Quote freshness\n  Scenario: Historical quote age\n    Given a valid historical sample\n    When the quote age is checked\n    Then the future quote is excluded\n")
	m := Manifest{SchemaVersion: 1}
	for i := 1; i <= 100; i++ {
		m.Rounds = append(m.Rounds, Entry{Round: i, Status: "PENDING_BDD"})
	}
	m.Rounds[10] = Entry{Round: 11, Status: "READY_FOR_CI", Feature: "docs/bdd/round011.feature", TestFile: "internal/research100/quote_age_test.go", TestFunction: "TestRound011QuoteAgeHonorsAsOf", SourceFile: "internal/research100/quote_age.go", Revision: 1, RevisionReason: "Document existing future-time quote guard"}
	raw, _ := json.Marshal(m)
	write("docs/bdd/manifest.json", string(raw))
	return dir, m
}
func TestScenarioAuditMustNotAutoAcceptHistoricalRounds(t *testing.T) {
	dir, m := fixture(t)
	s, e := validate(dir, m)
	if e != nil || s.Verified != 0 || s.Ready != 1 || s.Pending != 99 || s.BDDComplete {
		t.Fatalf("invalid historical status %+v %v", s, e)
	}
}
func TestFailOnMissingThenOrTest(t *testing.T) {
	dir, m := fixture(t)
	path := filepath.Join(dir, "docs/bdd/round011.feature")
	if e := os.WriteFile(path, []byte("@round-011\nFeature: x\nScenario: invalid\nGiven x\nWhen y\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := validate(dir, m); e == nil || !strings.Contains(e.Error(), "Then") {
		t.Fatalf("missing Then not blocked: %v", e)
	}
}
func TestEveryChangedImplementationNeedsBDDRevision(t *testing.T) {
	_, m := fixture(t)
	change := []string{"internal/research100/quote_age.go"}
	if e := validateChangedCode(m, change); e == nil {
		t.Fatal("no BDD revision accepted")
	}
	change = []string{"internal/research100/quote_age.go", "docs/bdd/manifest.json", "docs/bdd/round011.feature", "internal/research100/quote_age_test.go"}
	if e := validateChangedCode(m, change); e != nil {
		t.Fatalf("matching revision rejected: %v", e)
	}
	change = append(change, "internal/research100/dataset.go")
	if e := validateChangedCode(m, change); e == nil {
		t.Fatal("untraced second implementation accepted")
	}
}
func TestNoPathTraversalInManifest(t *testing.T) {
	dir, m := fixture(t)
	m.Rounds[10].Feature = "../outside.feature"
	if _, e := validate(dir, m); e == nil {
		t.Fatal("path traversal allowed")
	}
}
func TestIncompleteBDDStrictModeFails(t *testing.T) {
	dir, _ := fixture(t)
	s, e := run(dir, "", true)
	if e == nil || s.Verified != 0 {
		t.Fatalf("strict completion must fail: %+v %v", s, e)
	}
}
