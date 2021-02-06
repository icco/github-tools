package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	flag.Parse()

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	// Toggle a 👍 reaction on an issue.
	//
	// That involves first doing a query (and determining whether the reaction already exists),
	// then either adding or removing it.
	{
		var q struct {
			Repository struct {
				Issue struct {
					ID        githubv4.ID
					Reactions struct {
						ViewerHasReacted githubv4.Boolean
					} `graphql:"reactions(content:$reactionContent)"`
				} `graphql:"issue(number:$issueNumber)"`
			} `graphql:"repository(owner:$repositoryOwner,name:$repositoryName)"`
		}
		variables := map[string]interface{}{
			"repositoryOwner": githubv4.String("shurcooL-test"),
			"repositoryName":  githubv4.String("test-repo"),
			"issueNumber":     githubv4.Int(2),
			"reactionContent": githubv4.ReactionContentThumbsUp,
		}
		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			return err
		}
		fmt.Println("already reacted:", q.Repository.Issue.Reactions.ViewerHasReacted)

		if !q.Repository.Issue.Reactions.ViewerHasReacted {
			// Add reaction.
			var m struct {
				AddReaction struct {
					Subject struct {
						ReactionGroups []struct {
							Content githubv4.ReactionContent
							Users   struct {
								TotalCount githubv4.Int
							}
						}
					}
				} `graphql:"addReaction(input:$input)"`
			}
			input := githubv4.AddReactionInput{
				SubjectID: q.Repository.Issue.ID,
				Content:   githubv4.ReactionContentThumbsUp,
			}
			err := client.Mutate(context.Background(), &m, input, nil)
			if err != nil {
				return err
			}
			printJSON(m)
			fmt.Println("Successfully added reaction.")
		} else {
			// Remove reaction.
			var m struct {
				RemoveReaction struct {
					Subject struct {
						ReactionGroups []struct {
							Content githubv4.ReactionContent
							Users   struct {
								TotalCount githubv4.Int
							}
						}
					}
				} `graphql:"removeReaction(input:$input)"`
			}
			input := githubv4.RemoveReactionInput{
				SubjectID: q.Repository.Issue.ID,
				Content:   githubv4.ReactionContentThumbsUp,
			}
			err := client.Mutate(context.Background(), &m, input, nil)
			if err != nil {
				return err
			}
			printJSON(m)
			fmt.Println("Successfully removed reaction.")
		}
	}

	return nil
}
