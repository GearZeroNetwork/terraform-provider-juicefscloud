package juicefs

import (
	"encoding/json"
	"fmt"
)

type Cloud struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Storage string `json:"storage"`
}

func (c *Client) GetClouds() ([]Cloud, error) {
	u := fmt.Sprintf("%s/clouds", c.Endpoint)
	statusCode, body, err := c.request("GET", u, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("get regions failed, status code: %d, body: %s", statusCode, body)
	}
	var clouds []Cloud
	if err := json.Unmarshal(body, &clouds); err != nil {
		return nil, err
	}
	return clouds, nil
}
