package commands

import (
	"fmt"
	"context"
	"os"
	"io/ioutil"
	"os/user"
	"path"

	"github.com/urfave/cli"
	"github.com/cbegin/graven/domain"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"github.com/bgentry/speakeasy"
	"github.com/cbegin/graven/util"
	"strings"
)

type ConfigMap map[string]map[string]string
type Validator func(stdout, stderr string) error

var DeployCommand = cli.Command{
	Name: "deploy",
	Usage:       "Deploys artifacts to a repository",
	Action: deploy,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "login",
			Usage: "Prompts for repo login credentials.",
		},
	},
}

func deploy(c *cli.Context) error {
	project, err := domain.FindProject()
	if err != nil {
		return err
	}

	if err := verifyRepoState(project); err != nil {
		return err
	}

	if c.Bool("login") {
		return loginToGithub()
	}

	if err := pkg(c); err != nil {
		return err
	}

	return deployToGithub(project)
}

func loginToGithub() error {
	token, err := readSecret("Please type or paste a github token (will not echo): ")
	config, err := readConfig()
	if err != nil {
		config = ConfigMap{}
	}
	config["github"] = map[string]string{}
	config["github"]["token"] = token
	err = writeConfig(config)
	if err != nil {
		return fmt.Errorf("Error writing configuration file. %v", err)
	}
	return nil
}

func deployToGithub(project *domain.Project) error {

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
	config, err := readConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading configuration (try: deploy --login): %v", err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config["github"]["token"]},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client, ctx, nil
}

func readSecret(prompt string) (string, error) {
	password, err := speakeasy.Ask(prompt)
	if err != nil {
		return "", fmt.Errorf("Error reading secret from terminal: %v", err)
	}
	return password, nil
}

func readConfig() (ConfigMap, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path.Join(usr.HomeDir, ".graven.yaml"))
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config := ConfigMap{}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func writeConfig(config ConfigMap) (error) {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(usr.HomeDir, ".graven.yaml"), bytes, 0600)
	if err != nil {
		return err
	}

	return nil
}

func verifyRepoState(project *domain.Project) error {
	remoteName := "origin"
	branchName := "master"

	// Check if on expected branch (e.g. master)
	if err := verifyGitState(func(stdout, stderr string) error {
		actualBranch := strings.TrimSpace(stdout)
		if actualBranch != branchName {
			return fmt.Errorf("Expected to be on branch %v but found branch %v", branchName, actualBranch)
		}
		return nil
	}, project, "rev-parse", "--abbrev-ref", "HEAD"); err != nil {
		return err
	}

	// Ensure no uncommitted changes
	if err := verifyGitState(func(stdout, stderr string) error {
		if strings.TrimSpace(stdout) != "" || strings.TrimSpace(stderr) != "" {
			return fmt.Errorf("Cannot deploy with uncommitted changes.")
		}

		fmt.Println(stdout)
		fmt.Println(stderr)
		return nil
	}, project, "status", "--porcelain"); err != nil {
		return err
	}

	// Check if changes exist on server
	if err := verifyGitState(func(stdout, stderr string) error {
		fmt.Println(stdout)
		return nil
	}, project, "fetch", "--dry-run", remoteName, branchName); err != nil {
		return err
	}

	if err := verifyGitState(func(stdout, stderr string) error {
		return nil
	}, project, "rev-parse", branchName, fmt.Sprintf("%v/%v", remoteName, branchName)); err != nil {
		return err
	}

	return fmt.Errorf("Testing...")
}

func verifyGitState(validator Validator, project *domain.Project, args... string) error {
	sout, serr, err := util.RunCommand(project.ProjectPath(), nil, "git", args...)
	if err != nil {
		return fmt.Errorf("ERROR running Git command: %v\n", err)
	}
	return validator(sout, serr)
}