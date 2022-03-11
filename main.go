package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v43/github"
	"gitlab.com/golang-commonmark/markdown"
)

func main() {
	client := github.NewClient(nil)

	ctx := context.Background()

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
	fmt.Println("Last release: ", latest)

	prs, resp, err := client.PullRequests.List(
		ctx,
		"replicatedhq",
		"kots",
		&github.PullRequestListOptions{
			State: "closed",
			ListOptions: github.ListOptions{
				PerPage: 50,
			},
		},
	)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic("resp not code 200")
	}

	var isMinor bool
	for _, pr := range prs {
		prType := []string{}
		if pr.MergedAt == nil {
			continue
		}
		for _, label := range pr.Labels {
			if strings.EqualFold(*label.Name, "type::feature") {
				isMinor = true
			}
			if strings.HasPrefix(*label.Name, "type::") {
				prType = append(prType, strings.TrimPrefix(*label.Name, "type::"))
			}
		}
		fmt.Println(*pr.MergeCommitSHA, prType, getReleaseNotes(pr.GetBody()))
		if strings.HasPrefix(*pr.MergeCommitSHA, os.Args[1]) {
			break
		}
	}
	fmt.Println("is patch release: ", !isMinor)
}

func getReleaseNotes(raw string) string {
	md := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	tokens := md.Parse([]byte(raw))

	//Print the result
	for _, t := range tokens {
		snippet := getSnippet(t)

		if snippet.content != "" && snippet.lang == "release-note" {
			return snippet.content
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
	case *markdown.CodeBlock:
		return snippet{
			tok.Content,
			"code",
		}
	case *markdown.CodeInline:
		return snippet{
			tok.Content,
			"code inline",
		}
	case *markdown.Fence:
		return snippet{
			tok.Content,
			tok.Params,
		}
	}
	return snippet{}
}
