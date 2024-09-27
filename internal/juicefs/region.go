package juicefs

import (
	"encoding/json"
	"fmt"
)

type Region struct {
	Id        int64  `json:"id"`
	Cloud     int64  `json:"cloud"`
	Name      string `json:"name"`
	Desp      string `json:"desp"`
	Owner     int64  `json:"owner"`
	Token     string `json:"token"`
	TrashTime int64  `json:"trashtime"`
}

func (c *Client) GetRegions() ([]Region, error) {
	u := fmt.Sprintf("%s/regions", c.Endpoint)
	statusCode, body, err := c.request("GET", u, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("get regions failed, status code: %d, body: %s", statusCode, body)
	}
	var regions []Region
	if err := json.Unmarshal(body, &regions); err != nil {
		return nil, err
	}
	return regions, nil
}
