package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v43/github"
	"gitlab.com/golang-commonmark/markdown"
)

func main() {
	client := github.NewClient(nil)

	ctx := context.Background()

	latest := getLatestReleasedVersion(ctx, client)
	fmt.Println("Last release: ", latest)

	notes := getAllReleaseNotes(ctx, client)

	fmt.Println("features:")
	for _, feature := range notes.features {
		fmt.Println("\t", feature)
	}
	fmt.Println("bugs:")
	for _, bug := range notes.bugs {
		fmt.Println("\t", bug)
	}
}

func getLatestReleasedVersion(ctx context.Context, client *github.Client) string {
	rels, resp, err := client.Repositories.ListReleases(
		ctx,
		"replicatedhq",
		"kots",
		&github.ListOptions{},
	)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic("resp not code 200")
	}

	latest := ""
	for _, rel := range rels {
		if !*rel.Prerelease {
			latest = rel.GetTagName()
			break
		}
	}
	return latest
}

type releaseNotes struct {
	features []string
	bugs     []string
}

func getAllReleaseNotes(ctx context.Context, client *github.Client) releaseNotes {
	rls, resp, err := client.Repositories.ListReleases(
		ctx,
		"replicatedhq",
		"kots",
		&github.ListOptions{PerPage: 1},
	)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic("resp not code 200")
	}

	releaseNotes := releaseNotes{
		features: []string{},
		bugs:     []string{},
	}

	r := regexp.MustCompile(`#(\d{1,5})`)
	prsToCheck := r.FindAllStringSubmatch(*rls[0].Body, -1)
	for _, prToCheck := range prsToCheck {
		prNumber, err := strconv.Atoi(prToCheck[1])
		if err != nil {
			panic(err)
		}
		pr, resp, err := client.PullRequests.Get(
			ctx,
			"replicatedhq",
			"kots",
			prNumber,
		)
		if err != nil {
			panic(err)
		}

		if resp.StatusCode != 200 {
			panic("resp not code 200")
		}

		notes := getReleaseNotes(*pr.Body)

		if !strings.EqualFold(notes, "NONE") {
			for _, lbl := range pr.Labels {
				switch {
				case strings.EqualFold(*lbl.Name, "type::feature"):
					releaseNotes.features = append(releaseNotes.features, notes)
				case strings.EqualFold(*lbl.Name, "type::bug"):
					releaseNotes.bugs = append(releaseNotes.bugs, notes)
				}
				break
			}
		}
	}

	return releaseNotes
}

func getReleaseNotes(raw string) string {
	md := markdown.New()
	tokens := md.Parse([]byte(raw))

	for _, t := range tokens {
		snippet := getSnippet(t)
		if snippet.content != "" && snippet.lang == "release-note" {
			return strings.TrimSpace(snippet.content)
		}
	}
	return "NONE"
}

//snippet represents the snippet we will output.
type snippet struct {
	content string
	lang    string
}

//getSnippet extract only code snippet from markdown object.
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
