package main

import (
	"reflect"
	"strings"
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

func TestClassifyExtractedNotes(t *testing.T) {
	typeFeature := "type::feature"
	typeImprovement := "type::improvement"
	typeBug := "type::bug"

	typeLabelFlags := &TypeLabelFlags{
		FeatureTypeLabels:     []string{typeFeature},
		ImprovementTypeLabels: []string{typeImprovement},
		BugTypeLabels:         []string{typeBug},
	}

	htmlURL := "https://example.com/pr/1"

	tests := []struct {
		name           string
		notes          []ExtractedNote
		prLabels       []*github.Label
		wantFeatures   []string
		wantImprovements []string
		wantBugs       []string
	}{
		{
			name: "named fence overrides labels",
			notes: []ExtractedNote{
				{Text: "feat note", Category: CategoryFeature},
			},
			prLabels:       []*github.Label{{Name: &typeBug}},
			wantFeatures:   []string{"feat note"},
			wantImprovements: []string{},
			wantBugs:       []string{},
		},
		{
			name: "legacy note classified by feature label",
			notes: []ExtractedNote{
				{Text: "legacy feat", Category: CategoryUnspecified},
			},
			prLabels:       []*github.Label{{Name: &typeFeature}},
			wantFeatures:   []string{"legacy feat"},
			wantImprovements: []string{},
			wantBugs:       []string{},
		},
		{
			name: "legacy note classified by improvement label overriding feature",
			notes: []ExtractedNote{
				{Text: "legacy improvement", Category: CategoryUnspecified},
			},
			prLabels: []*github.Label{
				{Name: &typeFeature},
				{Name: &typeImprovement},
			},
			wantFeatures:     []string{},
			wantImprovements: []string{"legacy improvement"},
			wantBugs:         []string{},
		},
		{
			name: "legacy note with no matching label is dropped",
			notes: []ExtractedNote{
				{Text: "orphan note", Category: CategoryUnspecified},
			},
			prLabels:         []*github.Label{{Name: ptrString("kind::misc")}},
			wantFeatures:     []string{},
			wantImprovements: []string{},
			wantBugs:         []string{},
		},
		{
			name: "NONE skipped in legacy fence",
			notes: []ExtractedNote{
				{Text: "NONE", Category: CategoryUnspecified},
			},
			prLabels:         []*github.Label{{Name: &typeFeature}},
			wantFeatures:     []string{},
			wantImprovements: []string{},
			wantBugs:         []string{},
		},
		{
			name: "NONE skipped in named fence",
			notes: []ExtractedNote{
				{Text: "NONE", Category: CategoryFeature},
			},
			prLabels:         []*github.Label{{Name: &typeFeature}},
			wantFeatures:     []string{},
			wantImprovements: []string{},
			wantBugs:         []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseNotes := &ReleaseNotes{
				Features:     []string{},
				Improvements: []string{},
				Bugs:         []string{},
			}
			pr := &github.PullRequest{
				HTMLURL: &htmlURL,
				Labels:  tt.prLabels,
			}

			for _, en := range tt.notes {
				text := cleanReleaseNote(en.Text)
				if strings.EqualFold(text, "NONE") || text == "" {
					continue
				}
				switch en.Category {
				case CategoryFeature:
					releaseNotes.Features = append(releaseNotes.Features, text)
				case CategoryBug:
					releaseNotes.Bugs = append(releaseNotes.Bugs, text)
				case CategoryImprovement:
					releaseNotes.Improvements = append(releaseNotes.Improvements, text)
				case CategoryUnspecified:
					switch typeLabelFlags.GetNoteTypeFromLabels(pr.Labels) {
					case "feature":
						releaseNotes.Features = append(releaseNotes.Features, text)
					case "improvement":
						releaseNotes.Improvements = append(releaseNotes.Improvements, text)
					case "bug":
						releaseNotes.Bugs = append(releaseNotes.Bugs, text)
					}
				}
			}

			if !reflect.DeepEqual(releaseNotes.Features, tt.wantFeatures) {
				t.Errorf("Features = %#v, want %#v", releaseNotes.Features, tt.wantFeatures)
			}
			if !reflect.DeepEqual(releaseNotes.Improvements, tt.wantImprovements) {
				t.Errorf("Improvements = %#v, want %#v", releaseNotes.Improvements, tt.wantImprovements)
			}
			if !reflect.DeepEqual(releaseNotes.Bugs, tt.wantBugs) {
				t.Errorf("Bugs = %#v, want %#v", releaseNotes.Bugs, tt.wantBugs)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}
