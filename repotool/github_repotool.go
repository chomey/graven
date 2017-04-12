package repotool

import (
	"github.com/cbegin/graven/config"
	"fmt"
	"os"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"github.com/cbegin/graven/domain"
	"github.com/cbegin/graven/vcstool"
	"context"
)

type GithubRepoTool struct {}

func (g *GithubRepoTool) Login() error {
	config := config.NewConfig()
	if err := config.Read(); err != nil {
		// ignore
	}
	err := config.SetSecret("github", "token", "Please type or paste a github token (will not echo): ")
	err = config.Write()
	if err != nil {
		return fmt.Errorf("Error writing configuration file. %v", err)
	}
	return nil
}

func (g *GithubRepoTool) Release(project *domain.Project) error {

	gh, ctx, err := authenticate()
	if err != nil {
		return err
	}

	repo, ok := project.Repositories["github"]
	if !ok {
		return fmt.Errorf("Sorry, could not find gihub repo configuration")
	}

	ownerName := repo["owner"]
	repoName := repo["repo"]

	tagName := fmt.Sprintf("v%s", project.Version)
	releaseName := tagName
	release := &github.RepositoryRelease{
		TagName: &tagName,
		Name: &releaseName,
	}

	// TODO: Make this configurable
	var vcsTool vcstool.VCSTool = &vcstool.GitVCSTool{}
	if err := vcsTool.Tag(project, tagName); err != nil {
		return err
	}

	release, _, err = gh.Repositories.CreateRelease(ctx, ownerName, repoName, release)
	if err != nil {
		return err
	}
	fmt.Printf("Created release %v/%v:%v\n", ownerName, repoName, *release.Name)

	for _, a := range project.Artifacts {
		filename := a.ArtifactFile(project)
		sourceFile, err := os.Open(project.TargetPath(filename))
		if err != nil {
			return err
		}
		opts := &github.UploadOptions{
			Name: filename,
		}
		_, _, err = gh.Repositories.UploadReleaseAsset(ctx, ownerName, repoName, *release.ID, opts, sourceFile)
		if err != nil {
			return err
		}
		fmt.Printf("Uploaded %v/%v/%v\n", ownerName, repoName, filename)
	}

	return err
}

func authenticate() (*github.Client, context.Context, error) {
	config := config.NewConfig()
	if err := config.Read(); err != nil {
		return nil, nil, fmt.Errorf("Error reading configuration (try: release --login): %v", err)
	}

	token := config.Get("github", "token")
	if token == "" {
		return nil, nil, fmt.Errorf("Configuration missing token (try: release --login).")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client, ctx, nil
}