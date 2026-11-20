# SkewLab — mandatory round-by-round BDD acceptance

**Historical engineering acceptance is not retroactive BDD acceptance.** The original 100 code/test rounds passed Go CI, but were not authored and mapped as 100 independently reviewed BDD scenarios. `manifest.json` records the status for **every** round. Do not automatically change `PENDING_BDD` to `VERIFIED`.

## Mandatory acceptance loop for *each* new or remediated round

1. **Given** the current requirement and prior BDD revision, define at least one behavior scenario in `docs/bdd/roundNNN.feature`, including positive, failure and boundary cases as applicable. Include `@round-NNN`.
2. **When** source or behavior changes, update the `.feature` scenario, mapped Go `_test.go` function, the `docs/bdd/manifest.json` entry, and its `revision_reason`. If the behavior was intentionally unchanged (refactor), document this in `revision_reason` and still rerun the mapped scenarios.
3. **Then** run the mapped Go tests plus full `go test -race ./...`, `go vet ./...`, and `go run ./cmd/bdd-gate`. The CI workflow enforces that every changed production Go file is linked to a **changed** feature and **changed** mapped test. Passing tests without Given/When/Then is not sufficient.
4. **Confirm** the GitHub CI run for the exact commit is PASS, inspect its assertions against the scenario, and only then mark that round `VERIFIED`. A failure, cancellation, skipped test or pending CI must never count as accepted BDD.
5. **Record** round, scenario revision, test function, acceptance evidence SHA/CI, risk notes, next action and any blocker in `docs/SKEWLAB_100_ROUND_PROGRESS.md` before replying to the user. Updating a BDD document does not count as a new engineering round.

## Machine-checkable statuses

- `PENDING_BDD`: historical engineering round **without** approved, mapped BDD acceptance. Does not count as BDD completed.
- `READY_FOR_CI`: Given/When/Then and mapped Go test exist; CI pass and scenario review are still outstanding. Does not count as BDD completed.
- `VERIFIED`: mapped behavior scenario is reviewed and has passed the exact-revision CI. The manifest itself cannot prove the human review or actual CI result; verify and cite both externally.

Execute locally (Go 1.23+):

```bash
go test -race ./...
go vet ./...
go run ./cmd/bdd-gate
# Explicitly fail if any of the 100 historical BDD rounds remains incomplete:
go run ./cmd/bdd-gate --require-all
```

The normal CI check validates declared scenarios and **rejects any new production Go changes** lacking a same-push revision of a `.feature`, the scenario manifest, and the mapped Go test. The optional `--require-all` is intentionally red while legacy BDD debt exists; its failure must be reported as BDD incomplete, not hidden or bypassed.

Existing example: `round011.feature` maps to `TestRound011QuoteAgeHonorsAsOf`; the Go assertions verify chronology, P95, empty sample and no authorization. The round remains unverified until the current BDD CI and review have succeeded.

No scenarios may claim live execution or real profitability. Maintain `PAPER_ENTRIES_ENABLED` off; protect raw observations, passwords, persistent ledgers and Telegram credentials.
