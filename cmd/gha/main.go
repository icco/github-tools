package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/manifoldco/promptui"
	"golang.org/x/oauth2"
)

func main() {
	flag.Parse()

	token := os.Getenv("GITHUB_TOKEN")

	if err := run(context.Background(), token); err != nil {
		log.Fatal(err)
	}
}

func StrToBool(input string) (bool, error) {
	pieces := strings.Split(strings.TrimSpace(strings.ToLower(input)), "")
	if len(pieces) == 0 {
		return false, fmt.Errorf("empty string not allowed")
	}

	if pieces[0] != "y" && pieces[0] != "n" {
		return false, fmt.Errorf("answer must be yes or no")
	}

	return pieces[0] == "y", nil
}

func run(ctx context.Context, token string) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{
		Sort:        "pushed",
		Visibility:  "public",
		Affiliation: "owner",
		ListOptions: github.ListOptions{PerPage: 10},
	}

	prompt := promptui.Prompt{
		Label: "Archive",
		Validate: func(input string) error {
			_, err := StrToBool(input)
			return err
		},
	}

	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return err
		}

		for _, r := range repos {
			if !r.GetFork() && !r.GetArchived() {
				fmt.Printf("%s\n\tFork: %v\n\tArchived: %v\n\tPushedAt: %v\n", r.GetFullName(), r.GetFork(), r.GetArchived(), r.GetPushedAt())

				result, err := prompt.Run()
				if err != nil {
					return err
				}

				archive, err := StrToBool(result)
				if err != nil {
					return err
				}

				if archive {
					log.Printf("will archive")
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return nil
}
