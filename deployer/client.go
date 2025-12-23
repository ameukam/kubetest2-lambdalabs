package deployer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://cloud.lambdalabs.com/api/v1"

type Instance struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	IP           string `json:"ip"`
	Name         string `json:"name"`
	InstanceType struct {
		Name string `json:"name"`
	} `json:"instance_type"`
}

type instanceListResponse struct {
	Data []Instance `json:"data"`
}

type launchRequest struct {
	RegionName       string   `json:"region_name"`
	InstanceTypeName string   `json:"instance_type_name"`
	SSHKeyNames      []string `json:"ssh_key_names"`
	Quantity         int      `json:"quantity"`
	Name             string   `json:"name,omitempty"`
	ImageName        string   `json:"image_name,omitempty"`
}

type launchResponse struct {
	Data struct {
		InstanceIDs []string `json:"instance_ids"`
	} `json:"data"`
}

type terminateRequest struct {
	InstanceIDs []string `json:"instance_ids"`
}

type SSHKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type sshKeyListResponse struct {
	Data []SSHKey `json:"data"`
}

type addSSHKeyRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func (d *deployer) getClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

func (d *deployer) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+d.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.getClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed: %s %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

func (d *deployer) listInstances() ([]Instance, error) {
	resp, err := d.doRequest("GET", "/instances", nil)
	if err != nil {
		return nil, err
	}

	var listResp instanceListResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, err
	}

	return listResp.Data, nil
}

func (d *deployer) launchInstance(name string) (string, error) {
	req := launchRequest{
		RegionName:       d.region,
		InstanceTypeName: d.instanceType,
		SSHKeyNames:      []string{d.sshKeyName},
		Quantity:         1,
		Name:             name,
		ImageName:        d.image,
	}

	resp, err := d.doRequest("POST", "/instance-operations/launch", req)
	if err != nil {
		return "", err
	}

	var launchResp launchResponse
	if err := json.Unmarshal(resp, &launchResp); err != nil {
		return "", err
	}

	if len(launchResp.Data.InstanceIDs) == 0 {
		return "", fmt.Errorf("no instance IDs returned")
	}

	return launchResp.Data.InstanceIDs[0], nil
}

func (d *deployer) terminateInstance(id string) error {
	req := terminateRequest{
		InstanceIDs: []string{id},
	}
	_, err := d.doRequest("POST", "/instance-operations/terminate", req)
	return err
}

func (d *deployer) listSSHKeys() ([]SSHKey, error) {
	resp, err := d.doRequest("GET", "/ssh-keys", nil)
	if err != nil {
		return nil, err
	}

	var listResp sshKeyListResponse
	if err := json.Unmarshal(resp, &listResp); err != nil {
		return nil, err
	}

	return listResp.Data, nil
}

func (d *deployer) addSSHKey(name, pubKey string) error {
	req := addSSHKeyRequest{
		Name:      name,
		PublicKey: pubKey,
	}
	_, err := d.doRequest("POST", "/ssh-keys", req)
	return err
}

func (d *deployer) deleteSSHKey(id string) error {
	_, err := d.doRequest("DELETE", "/ssh-keys/"+id, nil)
	return err
}
