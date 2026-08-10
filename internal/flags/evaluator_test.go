package flags

import "testing"

func TestEvaluateDisabled(t *testing.T) {
	f := &Flag{Name: "x", Enabled: false, Version: 3}
	got := Evaluate(f, EvaluationContext{SubjectID: "u"})
	if got.Enabled || got.Reason != "flag_disabled" || got.Version != 3 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestEvaluateRule(t *testing.T) {
	f := &Flag{Name: "x", Enabled: true, Version: 1, Rules: []Rule{{Attribute: "country", Operator: "equals", Value: "US"}}}
	got := Evaluate(f, EvaluationContext{SubjectID: "u", Attributes: map[string]string{"country": "US"}})
	if !got.Enabled {
		t.Fatalf("expected enabled: %+v", got)
	}
}

func TestPercentageDeterministic(t *testing.T) {
	f := &Flag{Name: "x", Enabled: true, Rules: []Rule{{Attribute: "country", Operator: "equals", Value: "US", Percentage: 50}}}
	input := EvaluationContext{SubjectID: "same-user", Attributes: map[string]string{"country": "US"}}
	a := Evaluate(f, input)
	b := Evaluate(f, input)
	if a.Enabled != b.Enabled || a.Reason != b.Reason {
		t.Fatalf("evaluation not deterministic: %+v %+v", a, b)
	}
}

func TestValidateRule(t *testing.T) {
	if err := ValidateRule(Rule{Attribute: "x", Operator: "bad"}); err == nil {
		t.Fatal("expected invalid operator")
	}
	if err := ValidateRule(Rule{Attribute: "x", Operator: "equals", Percentage: 101}); err == nil {
		t.Fatal("expected invalid percentage")
	}
}
