package main

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v43/github"
)

func TestGetReleaseNotes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []ExtractedNote
	}{
		{
			name: "legacy single release-note block",
			raw: "Some text\n" +
				"```release-note\n" +
				"Added a new thing\n" +
				"Removed an old thing\n" +
				"```\n",
			want: []ExtractedNote{
				{Text: "Added a new thing", Category: CategoryUnspecified},
				{Text: "Removed an old thing", Category: CategoryUnspecified},
			},
		},
		{
			name: "new three-block template",
			raw: "```release-note-features\n" +
				"New feature A\n" +
				"```\n" +
				"```release-notes-fixes\n" +
				"Fixed bug B\n" +
				"```\n" +
				"```release-notes-improvements\n" +
				"Improved thing C\n" +
				"```\n",
			want: []ExtractedNote{
				{Text: "New feature A", Category: CategoryFeature},
				{Text: "Fixed bug B", Category: CategoryBug},
				{Text: "Improved thing C", Category: CategoryImprovement},
			},
		},
		{
			name: "mixed legacy and named fences",
			raw: "```release-note\n" +
				"Legacy note\n" +
				"```\n" +
				"```release-note-features\n" +
				"Named feature note\n" +
				"```\n",
			want: []ExtractedNote{
				{Text: "Legacy note", Category: CategoryUnspecified},
				{Text: "Named feature note", Category: CategoryFeature},
			},
		},
		{
			name: "release-note block with only NONE",
			raw: "```release-note\n" +
				"NONE\n" +
				"```\n",
			want: []ExtractedNote{
				{Text: "NONE", Category: CategoryUnspecified},
			},
		},
		{
			name: "no release-note fences at all",
			raw:  "Just some prose, no fences here.",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getReleaseNotes(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getReleaseNotes() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestClassifyNote(t *testing.T) {
	typeFeature := "type::feature"
	typeImprovement := "type::improvement"
	typeBug := "type::bug"

	typeLabelFlags := &TypeLabelFlags{
		FeatureTypeLabels:     []string{typeFeature},
		ImprovementTypeLabels: []string{typeImprovement},
		BugTypeLabels:         []string{typeBug},
	}

	tests := []struct {
		name       string
		note       ExtractedNote
		prLabels   []*github.Label
		wantBucket NoteCategory
	}{
		{
			name:       "named feature fence overrides bug label",
			note:       ExtractedNote{Text: "feat note", Category: CategoryFeature},
			prLabels:   []*github.Label{{Name: &typeBug}},
			wantBucket: CategoryFeature,
		},
		{
			name:       "named bug fence overrides feature label",
			note:       ExtractedNote{Text: "bug note", Category: CategoryBug},
			prLabels:   []*github.Label{{Name: &typeFeature}},
			wantBucket: CategoryBug,
		},
		{
			name:       "named improvement fence overrides feature label",
			note:       ExtractedNote{Text: "impr note", Category: CategoryImprovement},
			prLabels:   []*github.Label{{Name: &typeFeature}},
			wantBucket: CategoryImprovement,
		},
		{
			name:       "legacy note classified by feature label",
			note:       ExtractedNote{Text: "legacy feat", Category: CategoryUnspecified},
			prLabels:   []*github.Label{{Name: &typeFeature}},
			wantBucket: CategoryFeature,
		},
		{
			name:       "legacy note classified by bug label",
			note:       ExtractedNote{Text: "legacy bug", Category: CategoryUnspecified},
			prLabels:   []*github.Label{{Name: &typeBug}},
			wantBucket: CategoryBug,
		},
		{
			name:       "legacy note improvement label overrides feature label",
			note:       ExtractedNote{Text: "legacy improvement", Category: CategoryUnspecified},
			prLabels: []*github.Label{
				{Name: &typeFeature},
				{Name: &typeImprovement},
			},
			wantBucket: CategoryImprovement,
		},
		{
			name:       "legacy note with no matching label is dropped",
			note:       ExtractedNote{Text: "orphan note", Category: CategoryUnspecified},
			prLabels:   []*github.Label{{Name: ptrString("kind::misc")}},
			wantBucket: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyNote(tt.note, tt.prLabels, typeLabelFlags); got != tt.wantBucket {
				t.Errorf("classifyNote() = %v, want %v", got, tt.wantBucket)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}
