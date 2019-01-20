package groundcontrol

import (
	"container/list"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// LoadYAML loads the content of a YAML file into a User model.
func LoadYAML(file string, user *User) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if err := yaml.UnmarshalStrict(bytes, user); err != nil {
		return err
	}

	user.ID = EncodeID("User", "0")

	for i, workspace := range user.Workspaces {
		user.Workspaces[i].ID = EncodeID("Workspace", workspace.Slug)
		for j := range user.Workspaces[i].Projects {
			user.Workspaces[i].Projects[j].ID = EncodeID(
				"Project",
				workspace.Slug,
				workspace.Projects[j].Repository,
				workspace.Projects[j].Branch,
			)
			user.Workspaces[i].Projects[j].Workspace = &user.Workspaces[i]
			user.Workspaces[i].Projects[j].commitList = list.New()
		}
	}

	return nil
}
