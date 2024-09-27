package juicefs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type VolumeAccessRules struct {
	IpRange    string `json:"iprange"`
	Token      string `json:"token"`
	ReadOnly   bool   `json:"readonly"`
	AppendOnly bool   `json:"appendonly"`
}

type Volume struct {
	Id          int64               `json:"id"`
	AccessRules []VolumeAccessRules `json:"access_rules"`
	Owner       int64               `json:"owner"`
	Size        *int64              `json:"size"`
	Inodes      *int64              `json:"inodes"`
	Created     time.Time           `json:"created"`
	Uuid        string              `json:"uuid"`
	Name        string              `json:"name"`
	Region      int64               `json:"region"`
	Bucket      string              `json:"bucket"`
	TrashTime   int64               `json:"trashtime"`
	BlockSize   int64               `json:"blockSize"`
	Compress    string              `json:"compress"`
	Compatible  bool                `json:"compatible"`
	Extend      *string             `json:"extend"`
	Storage     *string             `json:"storage"`
}

func (c *Client) GetVolumes() ([]Volume, error) {
	u := fmt.Sprintf("%s/volumes", c.Endpoint)
	statusCode, body, err := c.request("GET", u, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		type CreateVolumeErrorResp struct {
			Name []string `json:"name"`
		}
		errResp := CreateVolumeErrorResp{}
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return nil, fmt.Errorf("failed to create volume, status code %d, error %s", statusCode, body)
		}
		return nil, fmt.Errorf("failed to create volume %s", strings.Join(errResp.Name, "\n"))
	}
	var successRet []Volume
	err = json.Unmarshal(body, &successRet)
	return successRet, err
}

type CreateVolumeRequest struct {
	Name       string  `json:"name"`
	Region     int64   `json:"region"`
	Bucket     *string `json:"bucket,omitempty"`
	TrashTime  *int64  `json:"trash_time,omitempty"`
	BlockSize  *int64  `json:"block_size,omitempty"`
	Compress   *string `json:"compress,omitempty"`
	Compatible *bool   `json:"compatible,omitempty"`
	Extend     *string `json:"extend,omitempty"`
	Storage    *string `json:"storage,omitempty"`
}

func (c *Client) CreateVolume(req CreateVolumeRequest) (*Volume, error) {
	u := fmt.Sprintf("%s/volumes", c.Endpoint)
	statusCode, body, err := c.request("POST", u, nil, &req)
	if err != nil {
		return nil, err
	}
	if statusCode != 201 {
		type CreateVolumeErrorResp struct {
			Name []string `json:"name"`
		}
		errResp := CreateVolumeErrorResp{}
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return nil, fmt.Errorf("failed to create volume, status code %d, error %s", statusCode, body)
		}
		return nil, fmt.Errorf("failed to create volume %s", strings.Join(errResp.Name, "\n"))
	}
	successRet := &Volume{}
	err = json.Unmarshal(body, successRet)
	return successRet, err
}

func (c *Client) GetVolume(volumeID int64) (*Volume, error) {
	u := fmt.Sprintf("%s/volumes/%d", c.Endpoint, volumeID)
	statusCode, body, err := c.request("GET", u, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode == 404 {
		return nil, fmt.Errorf("volume %d not found", volumeID)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("failed to get volume %d, statusCode: %d, error: %s", volumeID, statusCode, string(body))
	}
	successRet := &Volume{}
	err = json.Unmarshal(body, successRet)
	return successRet, err
}

func (c *Client) DeleteVolume(volumeID int64) error {
	u := fmt.Sprintf("%s/volumes/%d", c.Endpoint, volumeID)
	statusCode, body, err := c.request("DELETE", u, nil, nil)
	if err != nil {
		return err
	}
	if statusCode != 204 {
		return fmt.Errorf("failed to delete volume, status code %d, error %s", statusCode, body)
	}
	return nil
}

func (c *Client) IsVolumeReady(volumeID int64) (bool, error) {
	u := fmt.Sprintf("%s/volumes/%d/is_ready", c.Endpoint, volumeID)
	statusCode, body, err := c.request("GET", u, nil, nil)
	if err != nil {
		return false, err
	}
	if statusCode != 200 {
		return false, fmt.Errorf("failed to check volume ready, status code %d, error %s", statusCode, body)
	}

	var successRet struct {
		IsReady bool `json:"is_ready"`
	}

	err = json.Unmarshal(body, &successRet)
	return successRet.IsReady, err
}
