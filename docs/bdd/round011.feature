@round-011 @research @read-only
Feature: Round 011 - historical quote age never uses future observations
  As a SkewLab research operator
  I want quoted freshness calculated using only data available at the report time
  So that simulated evidence cannot be made better by future information

  Scenario: Exclude future observations from quote-age statistics
    Given two valid historical market observations and one future observation
    When the quote-age audit runs at the historical report time
    Then only the two historical observations are included
    And the P95 quote age is computed from those two observations
    And the analysis never authorizes paper or real trading

  Scenario: No eligible observation returns unknown, not fabricated latency
    Given observations whose captures are all in the future relative to the report time
    When the quote-age audit runs
    Then the number of usable observations is zero
    And P95 quote age is absent rather than fabricated
