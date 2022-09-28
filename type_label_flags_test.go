package main

import (
	"flag"
	"testing"

	"github.com/google/go-github/v43/github"
)

func TestTypeLabelFlags_Parse(t *testing.T) {
	feature := "type::feature"
	improvement := "type::improvement,type::security"
	bug := "type::bug"
	empty := ""

	tests := []struct {
		name                        string
		featureTypeLabelsString     *string
		improvementTypeLabelsString *string
		bugTypeLabelsString         *string
		wantFeatureTypeLabels       []string
		wantImprovementTypeLabels   []string
		wantBugTypeLabels           []string
	}{
		{
			name:                        "basic",
			featureTypeLabelsString:     &feature,
			improvementTypeLabelsString: &improvement,
			bugTypeLabelsString:         &bug,
			wantFeatureTypeLabels:       []string{"type::feature"},
			wantImprovementTypeLabels:   []string{"type::improvement", "type::security"},
			wantBugTypeLabels:           []string{"type::bug"},
		},
		{
			name:                      "empty",
			featureTypeLabelsString:   &empty,
			bugTypeLabelsString:       nil,
			wantFeatureTypeLabels:     []string{},
			wantImprovementTypeLabels: []string{},
			wantBugTypeLabels:         []string{},
		},
	}

	assertLabels := func(t *testing.T, got, want []string, labelType string) {
		if len(got) != len(want) {
			t.Errorf("len(TypeLabelFlags.%s), got %q, want len %v", labelType, got, len(want))
		}
		for _, l := range want {
			found := false
			for _, ll := range got {
				if l == ll {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("TypeLabelFlags.%s, not found, want %v", labelType, l)
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := flag.NewFlagSet("test", flag.ExitOnError)
			f := NewTypeLabelFlags(flagSet)
			args := []string{}
			if tt.featureTypeLabelsString != nil {
				args = append(args, "--feature-type-labels", *tt.featureTypeLabelsString)
			}
			if tt.improvementTypeLabelsString != nil {
				args = append(args, "--improvement-type-labels", *tt.improvementTypeLabelsString)
			}
			if tt.bugTypeLabelsString != nil {
				args = append(args, "--bug-type-labels", *tt.bugTypeLabelsString)
			}
			flagSet.Parse(args)
			f.Parse()

			assertLabels(t, f.FeatureTypeLabels, tt.wantFeatureTypeLabels, "FeatureTypeLabels")
			assertLabels(t, f.ImprovementTypeLabels, tt.wantImprovementTypeLabels, "ImprovementTypeLabels")
			assertLabels(t, f.BugTypeLabels, tt.wantBugTypeLabels, "BugTypeLabels")
		})
	}
}

func TestTypeLabelFlags_GetNoteTypeFromLabels(t *testing.T) {
	typeFeature := "type::feature"
	typeImprovement := "type::improvement"
	typeSecurity := "type::security"
	typeBug := "type::bug"
	someOtherLabel := "blah::blah"

	typeLabelFlags := &TypeLabelFlags{
		FeatureTypeLabels:     []string{typeFeature},
		ImprovementTypeLabels: []string{typeImprovement, typeSecurity},
		BugTypeLabels:         []string{typeBug},
	}

	tests := []struct {
		name         string
		prLabels     []*github.Label
		wantNoteType string
	}{
		{
			name: "feature",
			prLabels: []*github.Label{
				{Name: &someOtherLabel},
				{Name: &typeFeature},
				{Name: &someOtherLabel},
			},
			wantNoteType: "feature",
		},
		{
			name: "improvement first",
			prLabels: []*github.Label{
				{Name: &typeImprovement},
				{Name: &typeFeature},
			},
			wantNoteType: "improvement",
		},
		{
			name: "improvement second",
			prLabels: []*github.Label{
				{Name: &typeFeature},
				{Name: &typeImprovement},
			},
			wantNoteType: "improvement",
		},
		{
			name: "improvement security",
			prLabels: []*github.Label{
				{Name: &typeSecurity},
			},
			wantNoteType: "improvement",
		},
		{
			name: "bug",
			prLabels: []*github.Label{
				{Name: &typeBug},
			},
			wantNoteType: "bug",
		},
		{
			name: "none",
			prLabels: []*github.Label{
				{Name: &someOtherLabel},
				{Name: &someOtherLabel},
			},
			wantNoteType: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNoteType := typeLabelFlags.GetNoteTypeFromLabels(tt.prLabels); gotNoteType != tt.wantNoteType {
				t.Errorf("TypeLabelFlags.GetNoteTypeFromLabels() = %v, want %v", gotNoteType, tt.wantNoteType)
			}
		})
	}
}
