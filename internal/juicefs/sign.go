package juicefs

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Client struct {
	Endpoint  string
	AccessKey string
	SecretKey string
}

// 参数说明：
//
//	timestamp 是整数时间戳，以秒为单位
//	method 是 HTTP 请求方法，例如 GET, POST, PUT, DELETE
//	path 是 HTTP 请求路径，注意不包括查询参数，例如 /api/v1/regions
//	headers 是 HTTP 请求头，例如 {'Host': 'juicefs.com'}
//	query_params 是 HTTP 请求的查询参数，如果同一个字段对应多个值，请使用列表来保存
//	例如 {'page': '1', 'per_page': '10', 'sort': ['name', 'created_at']}
//	body 是 HTTP 请求的原始 body 内容，对于大多数 HTTP 库来说，可能需要先创建一个 Request 对象，然后再从该对象中获取 body 内容
func (c *Client) sign(
	timestamp int64,
	method string,
	path string,
	headers http.Header,
	queryParams url.Values,
	body []byte,
) (string, error) {
	// 1. 按顺序处理请求头，字段名转为小写字符，然后用 `:` 连接字段名和值，最后用 `\n` 连接所有字段。
	// 得到的结果形如 `host:juicefs.com`
	var sortedHeaders []string
	for _, h := range []string{"Host"} {
		v := headers.Get(h)
		if v == "" {
			return "", fmt.Errorf("header %s is required", h)
		}
		sortedHeaders = append(sortedHeaders, fmt.Sprintf("%s:%s", strings.ToLower(h), v))
	}
	sortedHeadersString := strings.Join(sortedHeaders, "\n")

	// 2. 对查询参数进行排序和编码
	// 排序规则：先按照字段名排序，因为需要处理同一个字段对应多个值的情况，所以还需要对字段的值进行排序
	// 编码规则：对字段名和值进行 URL 编码，然后用 `=` 连接字段名和值，最后用 `&` 连接所有字段
	// 得到的结果形如 `a=1&a=2&b=3&c=4`
	sortedQueryString := ""
	if queryParams != nil {
		sortedKeys := make([]string, 0, len(queryParams))
		for k := range queryParams {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		sortedQueryParams := make([]string, 0, len(queryParams))
		for _, k := range sortedKeys {
			sort.Strings(queryParams[k])
			for _, value := range queryParams[k] {
				sortedQueryParams = append(sortedQueryParams, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(value)))
			}
		}
		sortedQueryString = strings.Join(sortedQueryParams, "&")
	}

	// 3. 对请求体进行 SHA256 哈希
	payloadHash := ""
	if body != nil {
		hash := sha256.New()
		hash.Write(body)
		payloadHash = hex.EncodeToString(hash.Sum(nil))
	}

	// 4. 用 `\n` 按顺序拼接上面的所有字符串
	parts := []string{
		fmt.Sprintf("%d", timestamp),
		method,
		path,
		sortedHeadersString,
		sortedQueryString,
		payloadHash,
	}
	data := strings.Join(parts, "\n")

	// 5. 对拼接后的字符串进行 HMAC-SHA256 签名
	hash := hmac.New(sha256.New, []byte(c.SecretKey))
	hash.Write([]byte(data))
	signature := hex.EncodeToString(hash.Sum(nil))

	return signature, nil
}

func (c *Client) request(
	method string,
	path string,
	queryParams url.Values,
	data interface{},
) (int, []byte, error) {
	var (
		body         []byte
		err          error
		requestBytes []byte
	)
	timestamp := time.Now().Unix()
	reqUrlWithParam := fmt.Sprintf("%s?%s", path, queryParams.Encode())
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return 0, nil, err
		}
	}
	req, err := http.NewRequest(method, reqUrlWithParam, bytes.NewBuffer(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", req.URL.Host)

	fmt.Printf("secret_key: %s...%s\n", c.SecretKey[:10], c.SecretKey[len(c.SecretKey)-10:])
	fmt.Printf("timestamp: %d\n", timestamp)
	fmt.Printf("method: %s\n", method)
	fmt.Printf("path: %s\n", req.URL.Path)
	requestBytes, err = json.Marshal(req.Header)
	if err != nil {
		return 0, nil, err
	}
	fmt.Printf("headers: %s\n", string(requestBytes))
	fmt.Printf("query_params: %v\n", req.URL.RawQuery)
	fmt.Printf("body: %s\n", body)

	signature, err := c.sign(timestamp, method, req.URL.Path, req.Header, queryParams, body)
	if err != nil {
		return 0, nil, err
	}
	fmt.Printf("signature: %s\n", signature)

	auth := map[string]interface{}{
		"access_key": c.AccessKey,
		"timestamp":  timestamp,
		"signature":  signature,
		"version":    1,
	}
	jsonString, err := json.Marshal(auth)
	if err != nil {
		return 0, nil, err
	}
	token := base64.StdEncoding.EncodeToString(jsonString)
	fmt.Printf("token: %s\n", token)

	req.Header.Set("Authorization", token)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	fmt.Printf("response: \n%s\n", string(respBody))
	return resp.StatusCode, respBody, nil
}

//func (c *Client) GetVolumeExports() error {
//	u := fmt.Sprintf("%s/volumes/1/exports", c.Endpoint)
//	return c.request("GET", u, nil, nil)
//}
//
//func (c *Client) CreateVolumeExport() error {
//	u := fmt.Sprintf("%s/volumes/1/exports", c.Endpoint)
//	return c.request(
//		"POST",
//		u,
//		nil,
//		map[string]interface{}{
//			"desc":       "for mount",
//			"iprange":    "192.168.0.1/24",
//			"apionly":    false,
//			"readonly":   false,
//			"appendonly": false,
//		},
//	)
//}
//
//func (c *Client) UpdateVolumeExport() error {
//	u := fmt.Sprintf("%s/volumes/1/exports/1", c.Endpoint)
//	return c.request("PUT", u, nil, map[string]interface{}{"desc": "abc", "iprange": "192.168.100.1/24"})
//}
//
//func (c *Client) GetVolumeQuotas() error {
//	u := fmt.Sprintf("%s/volumes/1/quotas", c.Endpoint)
//	return c.request("GET", u, nil, nil)
//}
//
//func (c *Client) CreateVolumeQuota() error {
//	u := fmt.Sprintf("%s/volumes/1/quotas", c.Endpoint)
//	return c.request(
//		"POST",
//		u,
//		nil,
//		map[string]interface{}{"path": "/path/to/subdir", "inodes": 1 << 20, "size": 1 << 30},
//	)
//}
//
//func (c *Client) UpdateVolumeQuota() error {
//	u := fmt.Sprintf("%s/volumes/1/quotas/1", c.Endpoint)
//	return c.request("PUT", u, nil, map[string]interface{}{"path": "/foo", "size": 10 << 30})
//}
