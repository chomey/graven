package commands

import (
	"github.com/urfave/cli"
	"github.com/cbegin/graven/domain"
	"gopkg.in/yaml.v2"
	"os"
	"io/ioutil"
)

var ReleaseCommand = cli.Command{
	Name: "release",
	Usage:       "project release",
	UsageText:   "release - release project",
	Description: "increments the revision and packages the release",
	Action: release,
}

func release(c *cli.Context) error {
	project := c.App.Metadata["project"].(*domain.Project)

	version := domain.Version{}

	err := version.Parse(project.Version)
	if err != nil {
		return err
	}

	arg := c.Args().First()
	switch arg {
	case "major":
		version.Major++
		version.Minor = 0
		version.Patch = 0
		version.Qualifier = ""
	case "minor":
		version.Minor++
		version.Patch = 0
		version.Qualifier = ""
	case "patch":
		version.Patch++
		version.Qualifier = ""
	default:
		version.Qualifier = arg
	}

	project.Version = version.ToString()

	bytes, err := yaml.Marshal(project)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(project.FilePath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(project.FilePath, bytes, fileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}


