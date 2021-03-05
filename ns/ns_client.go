package ns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/nullstone-io/module/config"
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

type RunConfig struct {
	WorkspaceUid  uuid.UUID `json:"workspaceUid"`
	Source        string    `json:"source"`
	SourceVersion string    `json:"sourceVersion"`
	//Variables     Variables   `json:"variables"`
	Connections Connections `json:"connections"`
	//Providers     Providers   `json:"providers"`
}

type Connections map[string]Connection

type Connection struct {
	config.Connection
	Target string `json:"target"`
	Unused bool   `json:"unused"`
}

func GetWorkspaceConfig(client *Client, workspaceLocation WorkspaceLocation) (*RunConfig, error) {
	workspace, err := client.GetWorkspace(workspaceLocation)
	if err != nil {
		return nil, err
	} else if workspace == nil {
		return nil, fmt.Errorf(`no nullstone workspace (stack=%s, env=%s, block=%s)`, workspaceLocation.Stack, workspaceLocation.Env, workspaceLocation.Block)
	}
	return client.GetLatestConfig(workspace.StackName, workspace.Uid)
}

type Client struct {
	Config Config
	Org    string
}

func (c *Client) GetWorkspace(workspaceLocation WorkspaceLocation) (*Workspace, error) {
	client := &http.Client{
		Transport: c.Config.CreateTransport(http.DefaultTransport),
	}

	u, err := c.Config.ConstructUrl(path.Join("orgs", c.Org, "stacks", workspaceLocation.Stack, "workspaces"))
	if err != nil {
		return nil, err
	}
	u.RawQuery = url.Values{
		"envName":   []string{workspaceLocation.Env},
		"blockName": []string{workspaceLocation.Block},
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

func (c *Client) GetLatestConfig(stackName string, workspaceUid uuid.UUID) (*RunConfig, error) {
	client := &http.Client{
		Transport: c.Config.CreateTransport(http.DefaultTransport),
	}

	u, err := c.Config.ConstructUrl(path.Join("orgs", c.Org, "stacks", stackName, "workspaces", workspaceUid.String(), "run-configs", "latest"))
	if err != nil {
		return nil, err
	}

	res, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting workspace run config (%d): %s", res.StatusCode, string(raw))
	}

	var runConfig RunConfig
	if err := json.Unmarshal(raw, &runConfig); err != nil {
		return nil, fmt.Errorf("invalid response getting workspace run config: %w", err)
	}
	return &runConfig, nil
}

// GET /orgs/autogen_subdomains/:subdomainName
func (c *Client) GetAutogenSubdomain(subdomainName string) (*AutogenSubdomain, error) {
	client := &http.Client{
		Transport: c.Config.CreateTransport(http.DefaultTransport),
	}

	u, err := c.Config.ConstructUrl(path.Join("orgs", c.Org, "autogen_subdomains", subdomainName))
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("error getting autogen subdomain (%d): %s", res.StatusCode, string(raw))
	}

	var subdomain AutogenSubdomain
	if err := json.Unmarshal(raw, &subdomain); err != nil {
		return nil, fmt.Errorf("invalid response getting autogen subdomain: %w", err)
	}
	return &subdomain, nil
}

// GET /orgs/autogen_subdomains/:subdomainName/delegation
func (c *Client) GetAutogenSubdomainDelegation(subdomainName string) (*AutogenSubdomainDelegation, error) {
	client := &http.Client{
		Transport: c.Config.CreateTransport(http.DefaultTransport),
	}

	u, err := c.Config.ConstructUrl(path.Join("orgs", c.Org, "autogen_subdomains", subdomainName, "delegation"))
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("error getting autogen subdomain delegation (%d): %s", res.StatusCode, string(raw))
	}

	var delegation AutogenSubdomainDelegation
	if err := json.Unmarshal(raw, &delegation); err != nil {
		return nil, fmt.Errorf("invalid response getting autogen subdomain delegation: %w", err)
	}
	return &delegation, nil
}

// PUT /orgs/autogen_subdomains/:subdomainId/delegation ...
func (c *Client) UpdateAutogenSubdomainDelegation(subdomainName string, delegation *AutogenSubdomainDelegation) (*AutogenSubdomainDelegation, error) {
	client := &http.Client{
		Transport: c.Config.CreateTransport(http.DefaultTransport),
	}

	u, err := c.Config.ConstructUrl(path.Join("orgs", c.Org, "autogen_subdomains", subdomainName, "delegation"))
	if err != nil {
		return nil, err
	}

	rawPayload, _ := json.Marshal(delegation)
	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewReader(rawPayload))
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	raw, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error updating autogen subdomain delegation (%d): %s", res.StatusCode, string(raw))
	}

	var updatedDelegation AutogenSubdomainDelegation
	if err := json.Unmarshal(raw, &updatedDelegation); err != nil {
		return nil, fmt.Errorf("invalid response updating autogen subdomain delegation: %w", err)
	}
	return &updatedDelegation, nil
}
