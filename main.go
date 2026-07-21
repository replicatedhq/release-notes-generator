package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v43/github"
	"gitlab.com/golang-commonmark/markdown"
	"golang.org/x/oauth2"
)

var rnTemplate = `## {{.Title}}

Released on {{.DateString}}
{{""}}

{{- if .Description}}
{{.Description}}
{{end}}

{{- if .Features }}
### New Features {#new-features-{{.TitleDash}}}
{{- range .Features}}
* {{.}}.
{{- end}}
{{end}}

{{- if .Improvements }}
### Improvements {#improvements-{{.TitleDash}}}
{{- range .Improvements}}
* {{.}}.
{{- end}}
{{end}}

{{- if .Bugs }}
### Bug Fixes {#bug-fixes-{{.TitleDash}}}
{{- range .Bugs}}
* {{.}}.
{{- end}}
{{end}}`

const GithubAuthTokenEnvironmentVarName = "GITHUB_AUTH_TOKEN"

type ReleaseNotes struct {
	Title        string
	TitleDash    string
	DateString   string
	Description  string
	Features     []string
	Improvements []string
	Bugs         []string
}

func main() {
	var ownerRepo string
	var base string
	var head string
	var title string
	var description string
	var includePrLinks bool

	flag.StringVar(&ownerRepo, "owner-repo", "", "The owner/repo of the Github repository to be used (required)")
	flag.StringVar(&base, "base", "", "Base of release notes diff (defaults to the last non-prerelease tag)")
	flag.StringVar(&head, "head", "", "Head of release notes diff (required)")
	flag.StringVar(&title, "title", "", "The release notes title")
	flag.StringVar(&description, "description", "", "Description to be added to the release notes (optional)")
	flag.BoolVar(&includePrLinks, "include-pr-links", false, "Include links back to pull requests")
	typeLabelFlags := NewTypeLabelFlags(flag.CommandLine)
	flag.Parse()
	typeLabelFlags.Parse()

	var httpClient *http.Client
	// if we have a token, use it to authenticate to prevent rate limiting
	if token, ok := os.LookupEnv(GithubAuthTokenEnvironmentVarName); ok {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(context.Background(), ts)
	} else {
		fmt.Fprintf(os.Stderr, "WARNING: No %s environment variable found, rate limiting may occur\n", GithubAuthTokenEnvironmentVarName)
	}

	ownerRepoParts := strings.Split(ownerRepo, "/")
	if len(ownerRepoParts) != 2 {
		log.Fatalf("Wrong number of parts in %s", ownerRepo)
	}
	owner := ownerRepoParts[0]
	repo := ownerRepoParts[1]

	client := github.NewClient(httpClient)
	ctx := context.Background()

	// get the latest released version if we don't have a base
	if base == "" {
		latest, err := getLatestReleasedVersion(ctx, client, owner, repo)
		if err != nil {
			log.Fatalf("Failed to get latest released version: %v", err)
		}
		base = latest
	}

	notes, err := getAllReleaseNotes(ctx, client, owner, repo, base, head, includePrLinks, typeLabelFlags)
	if err != nil {
		log.Fatalf("Failed to get release notes: %v", err)
	}

	// if the release notes do not contain any features, improvements, or bugs, do not submit release notes
	if len(notes.Features) == 0 && len(notes.Improvements) == 0 && len(notes.Bugs) == 0 {
		return
	}

	notes.Title = title
	notes.TitleDash = strings.ReplaceAll(notes.Title, ".", "-")
	notes.DateString = time.Now().Format("January 2, 2006")
	notes.Description = strings.ReplaceAll(description, "`", "\\`") // escape backticks

	t := template.Must(template.New("template").Parse(rnTemplate))
	err = t.Execute(os.Stdout, notes)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func getLatestReleasedVersion(ctx context.Context, client *github.Client, owner string, repo string) (string, error) {
	var releases []*github.RepositoryRelease
	listOptions := github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	releases, response, err := client.Repositories.ListReleases(ctx, owner, repo, &listOptions)
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code when listing releases: %d", response.StatusCode)
	}

	latest := ""
	for _, rel := range releases {
		if !*rel.Prerelease {
			latest = rel.GetTagName()
			break
		}
	}
	return latest, nil
}

func getAllReleaseNotes(ctx context.Context, client *github.Client, owner, repo, base, head string, includePrLinks bool, typeLabelFlags *TypeLabelFlags) (*ReleaseNotes, error) {
	var commits []*github.RepositoryCommit
	listOptions := github.ListOptions{
		Page:    0,
		PerPage: 100,
	}
	for {
		cmp, response, err := client.Repositories.CompareCommits(
			ctx,
			owner,
			repo,
			base,
			head,
			&listOptions,
		)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status code when getting commits: %d", response.StatusCode)
		}
		if len(cmp.Commits) > 0 {
			commits = append(commits, cmp.Commits...)
		}
		if response.NextPage == 0 {
			break
		}
		listOptions.Page = response.NextPage
	}

	// Picks up merge commits and squash-merged PRs
	r := regexp.MustCompile(`#(\d{1,5})`)

	prsToCheck := []string{}
	for _, commit := range commits {
		matches := r.FindStringSubmatch(*commit.Commit.Message)
		if len(matches) > 1 {
			prsToCheck = append(prsToCheck, matches[1])
		}
	}

	releaseNotes := ReleaseNotes{
		Features:     []string{},
		Improvements: []string{},
		Bugs:         []string{},
	}

	for _, prToCheck := range prsToCheck {
		prNumber, err := strconv.Atoi(prToCheck)
		if err != nil {
			return nil, err
		}
		pr, resp, err := client.PullRequests.Get(
			ctx,
			owner,
			repo,
			prNumber,
		)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status code when getting PR: %d", resp.StatusCode)
		}

		if pr.Body == nil {
			continue
		}

		notes := getReleaseNotes(*pr.Body)
		if len(notes) == 0 {
			continue
		}

		for _, en := range notes {
			text := cleanReleaseNote(en.Text)

			if strings.EqualFold(text, "NONE") || text == "" {
				continue
			}

			if includePrLinks {
				text = fmt.Sprintf("[#%d](%s) %s", prNumber, *pr.HTMLURL, text)
			}

			switch classifyNote(en, pr.Labels, typeLabelFlags) {
			case CategoryFeature:
				releaseNotes.Features = append(releaseNotes.Features, text)
			case CategoryBug:
				releaseNotes.Bugs = append(releaseNotes.Bugs, text)
			case CategoryImprovement:
				releaseNotes.Improvements = append(releaseNotes.Improvements, text)
			}
		}
	}

	return &releaseNotes, nil
}

type NoteCategory string

const (
	CategoryFeature     NoteCategory = "feature"
	CategoryImprovement NoteCategory = "improvement"
	CategoryBug         NoteCategory = "bug"
	CategoryUnspecified NoteCategory = "unspecified"
)

type ExtractedNote struct {
	Text     string
	Category NoteCategory
}

func classifyNote(en ExtractedNote, prLabels []*github.Label, typeLabelFlags *TypeLabelFlags) NoteCategory {
	switch en.Category {
	case CategoryFeature, CategoryBug, CategoryImprovement:
		return en.Category
	case CategoryUnspecified:
		switch typeLabelFlags.GetNoteTypeFromLabels(prLabels) {
		case "feature":
			return CategoryFeature
		case "improvement":
			return CategoryImprovement
		case "bug":
			return CategoryBug
		}
	}
	return ""
}

func getReleaseNotes(raw string) []ExtractedNote {
	md := markdown.New()
	tokens := md.Parse([]byte(raw))

	var out []ExtractedNote
	for _, t := range tokens {
		snippet := getSnippet(t)
		snippet.content = strings.TrimSpace(snippet.content)
		if snippet.content == "" {
			continue
		}
		var cat NoteCategory
		switch snippet.lang {
		case "release-note":
			cat = CategoryUnspecified
		case "release-note-features":
			cat = CategoryFeature
		case "release-notes-fixes":
			cat = CategoryBug
		case "release-notes-improvements":
			cat = CategoryImprovement
		default:
			continue
		}
		for _, line := range strings.Split(snippet.content, "\n") {
			out = append(out, ExtractedNote{Text: line, Category: cat})
		}
	}
	return out
}

// snippet represents the snippet we will output.
type snippet struct {
	content string
	lang    string
}

// getSnippet extract only code snippet from markdown object.
func getSnippet(tok markdown.Token) snippet {
	switch tok := tok.(type) {
	case *markdown.Fence:
		return snippet{
			tok.Content,
			tok.Params,
		}
	}
	return snippet{}
}

func cleanReleaseNote(note string) string {
	note = strings.TrimSpace(note)
	note = strings.TrimPrefix(note, "-")
	note = strings.TrimPrefix(note, "*")
	note = strings.TrimSuffix(note, ".")
	note = strings.TrimSpace(note)
	note = strings.ReplaceAll(note, "`", "\\`") // escape backticks
	return note
}
