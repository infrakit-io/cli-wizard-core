package wizard

import "testing"

func TestReconcileWithTemplate_AddMissingAndKeepExisting(t *testing.T) {
	current := map[string]any{
		"vm": map[string]any{
			"name": "node-01",
		},
	}
	template := map[string]any{
		"vm": map[string]any{
			"name":       "",
			"profile":    "talos",
			"ip_address": "",
		},
	}

	out, report, err := ReconcileWithTemplate(current, template, ReconcileOptions{})
	if err != nil {
		t.Fatalf("ReconcileWithTemplate error: %v", err)
	}
	vm, ok := out["vm"].(map[string]any)
	if !ok {
		t.Fatalf("vm should be object")
	}
	if got := vm["name"]; got != "node-01" {
		t.Fatalf("expected existing name to stay, got %v", got)
	}
	if got := vm["profile"]; got != "talos" {
		t.Fatalf("expected missing profile from template, got %v", got)
	}
	if len(report.Added) != 2 {
		t.Fatalf("expected 2 added paths, got %d: %v", len(report.Added), report.Added)
	}
}

func TestReconcileWithTemplate_DropUnknown(t *testing.T) {
	current := map[string]any{
		"vm": map[string]any{
			"name":     "node-01",
			"username": "",
		},
	}
	template := map[string]any{
		"vm": map[string]any{
			"name": "node-01",
		},
	}
	out, report, err := ReconcileWithTemplate(current, template, ReconcileOptions{DropUnknown: true})
	if err != nil {
		t.Fatalf("ReconcileWithTemplate error: %v", err)
	}
	vm := out["vm"].(map[string]any)
	if _, ok := vm["username"]; ok {
		t.Fatalf("expected username to be removed")
	}
	if len(report.Removed) != 1 || report.Removed[0] != "vm.username" {
		t.Fatalf("unexpected removed paths: %v", report.Removed)
	}
}

func TestReconcileWithTemplate_RequiredPaths(t *testing.T) {
	current := map[string]any{
		"vm": map[string]any{
			"name": "node-01",
		},
	}
	template := map[string]any{
		"vm": map[string]any{
			"name": "node-01",
		},
	}
	_, report, err := ReconcileWithTemplate(current, template, ReconcileOptions{
		RequiredPaths: []string{"vm.name", "vm.ip_address"},
	})
	if err != nil {
		t.Fatalf("ReconcileWithTemplate error: %v", err)
	}
	if len(report.MissingRequired) != 1 || report.MissingRequired[0] != "vm.ip_address" {
		t.Fatalf("unexpected missing required: %v", report.MissingRequired)
	}
}

func TestIsPlaceholderValue(t *testing.T) {
	cases := []struct {
		input any
		want  bool
	}{
		{"", true},
		{"   ", true},
		{"CHANGE_ME", true},
		{"change_me", true},
		{"CHANGE_ME_TO_SOMETHING", true},
		{nil, true},
		{"actual-value", false},
		{"192.168.1.1", false},
		{42, false},
		{false, false},
	}
	for _, c := range cases {
		got := IsPlaceholderValue(c.input)
		if got != c.want {
			t.Errorf("IsPlaceholderValue(%v) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestFindMissingRequired_EmptyPath(t *testing.T) {
	// Empty string in required paths should be skipped, not reported as missing
	data := map[string]any{"vm": map[string]any{"name": "node-01"}}
	missing := findMissingRequired(data, []string{"", "vm.name", ""})
	if len(missing) != 0 {
		t.Fatalf("expected no missing, got %v", missing)
	}
}

func TestFindPlaceholderValues_EmptyPath(t *testing.T) {
	// Empty string in required should be skipped
	data := map[string]any{"vm": map[string]any{"name": "CHANGE_ME"}}
	result := findPlaceholderValues(data, []string{"", "vm.name"}, nil)
	if len(result) != 1 || result[0] != "vm.name" {
		t.Fatalf("expected [vm.name], got %v", result)
	}
}

func TestGetPath_NonMapIntermediate(t *testing.T) {
	// Intermediate path element that is not a map should return nil
	data := map[string]any{"vm": "string-not-map"}
	if got := getPath(data, "vm.name"); got != nil {
		t.Fatalf("getPath with non-map intermediate should return nil, got %v", got)
	}
}

func TestHasPath_NonMapIntermediate(t *testing.T) {
	// Intermediate path element that is not a map should return false
	data := map[string]any{"vm": "string-not-map"}
	if hasPath(data, "vm.name") {
		t.Fatal("hasPath with non-map intermediate should return false")
	}
}

func TestCloneAnySlice(t *testing.T) {
	in := []any{"hello", 42, true}
	out := cloneAny(in).([]any)
	if len(out) != len(in) {
		t.Fatalf("cloneAny slice length: got %d, want %d", len(out), len(in))
	}
	for i, v := range in {
		if out[i] != v {
			t.Errorf("cloneAny slice[%d]: got %v, want %v", i, out[i], v)
		}
	}
	// Nested: slice containing a map
	nested := []any{map[string]any{"k": "v"}}
	outNested := cloneAny(nested).([]any)
	m := outNested[0].(map[string]any)
	if m["k"] != "v" {
		t.Fatalf("cloneAny nested slice map: got %v", m)
	}
}

func TestReconcileWithTemplate_CheckPlaceholders(t *testing.T) {
	current := map[string]any{
		"vm": map[string]any{
			"name":       "CHANGE_ME",
			"ip_address": "192.168.1.10",
			"netmask":    "",
		},
	}
	template := map[string]any{
		"vm": map[string]any{
			"name":       "",
			"ip_address": "",
			"netmask":    "",
		},
	}
	_, report, err := ReconcileWithTemplate(current, template, ReconcileOptions{
		RequiredPaths:     []string{"vm.name", "vm.ip_address", "vm.netmask"},
		CheckPlaceholders: true,
	})
	if err != nil {
		t.Fatalf("ReconcileWithTemplate error: %v", err)
	}
	// vm.netmask is empty — it's in required AND present but placeholder
	if len(report.MissingRequired) != 0 {
		t.Fatalf("expected no missing required (all paths exist), got: %v", report.MissingRequired)
	}
	// vm.name and vm.netmask are placeholders
	if len(report.PlaceholderValues) != 2 {
		t.Fatalf("expected 2 placeholder values, got %d: %v", len(report.PlaceholderValues), report.PlaceholderValues)
	}
	// ip_address should not be in placeholders
	for _, p := range report.PlaceholderValues {
		if p == "vm.ip_address" {
			t.Fatalf("vm.ip_address should not be a placeholder (has real value)")
		}
	}
}
