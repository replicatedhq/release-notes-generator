package main

import (
	"flag"
	"strings"

	"github.com/google/go-github/v43/github"
)

type TypeLabelFlags struct {
	FeatureTypeLabels, ImprovementTypeLabels, BugTypeLabels                   []string
	featureTypeLabelsString, improvementTypeLabelsString, bugTypeLabelsString string
}

func NewTypeLabelFlags(flags *flag.FlagSet) *TypeLabelFlags {
	f := &TypeLabelFlags{}
	// defaults dont work here because its too difficult to omit args in action based on input
	flags.StringVar(&f.featureTypeLabelsString, "feature-type-labels", "", "A comma separated list of labels to consider as features")
	flags.StringVar(&f.improvementTypeLabelsString, "improvement-type-labels", "", "A comma separated list of labels to consider as improvements")
	flags.StringVar(&f.bugTypeLabelsString, "bug-type-labels", "", "A comma separated list of labels to consider as bugs")
	return f
}

func (f *TypeLabelFlags) Parse() {
	f.FeatureTypeLabels = parseLabelsStringFlag(f.featureTypeLabelsString)
	f.ImprovementTypeLabels = parseLabelsStringFlag(f.improvementTypeLabelsString)
	f.BugTypeLabels = parseLabelsStringFlag(f.bugTypeLabelsString)
}

func (f *TypeLabelFlags) GetNoteTypeFromLabels(prLabels []*github.Label) (noteType string) {
	for _, lbl := range prLabels {
		for _, label := range f.FeatureTypeLabels {
			if strings.EqualFold(*lbl.Name, label) {
				// type::improvement is a secondary label and co-exists with type::feature,
				// so we don't break when we find type::feature label.
				if label == "type::feature" {
					noteType = "feature"
				} else {
					return "feature"
				}
			}
		}
		for _, label := range f.ImprovementTypeLabels {
			if strings.EqualFold(*lbl.Name, label) {
				return "improvement"
			}
		}
		for _, label := range f.BugTypeLabels {
			if strings.EqualFold(*lbl.Name, label) {
				return "bug"
			}
		}
	}
	return
}

func parseLabelsStringFlag(str string) (labels []string) {
	for _, label := range strings.Split(str, ",") {
		clean := strings.TrimSpace(label)
		if clean != "" {
			labels = append(labels, clean)
		}
	}
	return
}
