package ns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	Uid       uuid.UUID `json:"uid"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy string    `json:"createdBy"`
	OrgName   string    `json:"orgName"`
	StackName string    `json:"stackName"`
	BlockName string    `json:"blockName"`
	EnvName   string    `json:"envName"`
	Status    string    `json:"status"`
	StatusAt  time.Time `json:"statusAt"`
}

type Client struct {
	Config Config
	Org    string
}

func (c *Client) GetWorkspace(stackName, envName, blockName string) (*Workspace, error) {
	client := &http.Client{
		Transport: c.Config.CreateTransport(http.DefaultTransport),
	}

	u, err := c.Config.ConstructUrl(path.Join("orgs", c.Org, "stacks", stackName, "workspaces"))
	if err != nil {
		return nil, err
	}
	u.RawQuery = url.Values{
		"envName":   []string{envName},
		"blockName": []string{blockName},
	}.Encode()

	res, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	raw, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting workspace (%d): %s", res.StatusCode, string(raw))
	}

	workspaces := make([]*Workspace, 0)
	if err := json.Unmarshal(raw, &workspaces); err != nil {
		return nil, fmt.Errorf("invalid response getting workspaces: %w", err)
	}

	if len(workspaces) < 1 {
		return nil, nil
	}
	return workspaces[0], nil
}
