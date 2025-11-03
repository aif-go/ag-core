package aghertzclient

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type (
	// RequestParam 定义请求配置
	RequestParam struct {
		Method      string
		Path        string
		PathVars    map[string]string
		QueryParams map[string]string
		// Body        []byte     // DISABLED: 现在通过 req 参数传递
		ContentType string     // 新增：内容类型配置
		Serializer  Serializer // 新增：序列化器配置
	}

	// HertzBaseClient is the base client for hertz.
	HertzBaseClient struct {
		endpoint string
		isSD     bool
		cli      *client.Client
	}

	// AgHertzClientOption is the option for hertz base client.
	AgHertzClientOption func(c *HertzBaseClient)
)

// NewHertzBaseClient is the constructor for HertzBaseClient.
func NewHertzBaseClient(cli *client.Client, opts ...AgHertzClientOption) *HertzBaseClient {
	client := &HertzBaseClient{
		cli: cli,
	}

	// Apply all options
	for _, opt := range opts {
		opt(client)
	}

	return client
}
func WithSD(isSD bool) AgHertzClientOption {
	return func(c *HertzBaseClient) {
		c.isSD = isSD
	}
}
func WithEndpoint(endpoint string) AgHertzClientOption {
	return func(c *HertzBaseClient) {
		c.endpoint = endpoint
	}
}

// WithSDEndpoint sets the endpoint for the client with service discovery.
func WithSDEndpoint(endpoint string) AgHertzClientOption {
	return func(c *HertzBaseClient) {
		c.endpoint = endpoint
		c.isSD = true
	}
}

// WithDirectEndpoint sets the endpoint for the client without service discovery.
func WithDirectEndpoint(endpoint string) AgHertzClientOption {
	return func(c *HertzBaseClient) {
		c.endpoint = endpoint
		c.isSD = false
	}
}

func (c *HertzBaseClient) DoRequest(ctx context.Context, reqParam *RequestParam, req, resp interface{}, opts ...config.RequestOption) error {

	// 构建请求URL和选项
	rurl, requestOpts, err := c.buildRequestURLAndOptions(reqParam, opts)
	if err != nil {
		return err
	}

	// 创建请求
	preq := &protocol.Request{}
	preq.SetMethod(reqParam.Method)
	preq.SetRequestURI(rurl)
	preq.SetOptions(requestOpts...)

	/*
		preq.Header.SetContentTypeBytes([]byte("application/json"))
		// 设置请求体（如果有）
		if len(reqConfig.Body) > 0 {
			preq.SetBody(reqConfig.Body)
		}
	*/

	// 序列化请求体
	if req != nil {
		body, contentType, err := c.serializeRequestBody(reqParam, req)
		if err != nil {
			return err
		}
		preq.SetBody(body)
		preq.Header.SetContentTypeBytes([]byte(contentType))
	}

	// 设置查询参数（如果有）
	if len(reqParam.QueryParams) > 0 {
		for key, value := range reqParam.QueryParams {
			preq.URI().QueryArgs().Add(key, value)
		}
	}

	// 执行请求
	res := &protocol.Response{}
	if err := c.cli.Do(ctx, preq, res); err != nil {
		slog.Error("hertz client request failed", "method", reqParam.Method, "url", rurl, "err", err)
		return fmt.Errorf("hertz client request failed: %w", err)
	}

	// 检查响应状态
	if err := c.checkResponseStatus(res, rurl); err != nil {
		return err
	}

	// 解析响应体
	if err := c.parseResponseBody(res, resp); err != nil {
		return err
	}

	// 记录成功日志
	slog.Debug("hertz client request succeeded",
		"method", reqParam.Method,
		"url", rurl,
		"status", res.StatusCode())

	return nil
}

func (c *HertzBaseClient) checkResponseStatus(res *protocol.Response, rurl string) error {
	if res.StatusCode() != consts.StatusOK {
		slog.Error("hertz client request returned non-200 status",
			"url", rurl,
			"status", res.StatusCode(),
			"body", string(res.Body()))
		return fmt.Errorf("request failed with status %d: %s", res.StatusCode(), string(res.Body()))
	}
	return nil
}

// buildRequestURLAndOptions 构建请求URL和选项
func (c *HertzBaseClient) buildRequestURLAndOptions(reqConfig *RequestParam, opts []config.RequestOption) (string, []config.RequestOption, error) {
	// 替换路径参数
	rpath := reqConfig.Path
	for key, value := range reqConfig.PathVars {
		rpath = strings.Replace(rpath, ":"+key, value, 1)
	}

	// 确定基础URL和服务发现选项
	var baseURL string
	// var requestOpts []config.RequestOption
	requestOpts := append([]config.RequestOption{}, opts...)

	// 使用服务发现
	// endpoint 可以是服务名，但最终URL需要包含 http://
	endpoint := c.endpoint
	if endpoint == "" {
		// endpoint = "aghzwdemo_hertz"
		return "", nil, fmt.Errorf("endpoint is empty")
	}

	// 确保 baseURL 包含 http:// 前缀
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		// servername 已经是完整URL，直接使用
		baseURL = endpoint
	} else {
		// servername 只是服务名，添加 http:// 前缀
		baseURL = "http://" + endpoint
	}

	if c.isSD {
		// 启用服务发现
		requestOpts = append([]config.RequestOption{config.WithSD(true)}, opts...)
	}

	// 构建完整URL
	// 智能路径拼接，避免双斜杠
	var rurl string
	if rpath == "" {
		// 空路径：直接返回 baseURL，移除结尾斜杠
		rurl = strings.TrimSuffix(baseURL, "/")
	} else if strings.HasSuffix(baseURL, "/") && strings.HasPrefix(rpath, "/") {
		rurl = baseURL + rpath[1:]
	} else if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(rpath, "/") {
		rurl = baseURL + "/" + rpath
	} else {
		rurl = baseURL + rpath
	}

	return rurl, requestOpts, nil
}

func (c *HertzBaseClient) serializeRequestBody(reqConfig *RequestParam, req interface{}) ([]byte, string, error) {
	// 使用配置的内容类型
	serializer := c.getSerializer(reqConfig, req)
	body, err := serializer.Marshal(req)
	return body, serializer.ContentType(), err
}

// parseResponseBody 解析响应体
func (c *HertzBaseClient) parseResponseBody(res *protocol.Response, resp interface{}) error {
	body := res.Body()

	if len(body) == 0 {
		return nil
	}

	// 这里假设响应是JSON格式，需要根据实际情况调整解析逻辑
	// 如果服务端返回的是protobuf格式，需要使用proto.Unmarshal
	if err := json.Unmarshal(body, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *HertzBaseClient) getSerializer(reqConfig *RequestParam, req interface{}) Serializer {
	// 优先使用配置的序列化器
	if reqConfig.Serializer != nil {
		return reqConfig.Serializer
	}

	if reqConfig.ContentType != "" {
		serializer := GetSerializer(reqConfig.ContentType)
		if serializer != nil {
			return serializer
		}
	}
	/*
		if req != nil {
			// 智能类型推断
			switch req.(type) {
			case []byte:
				// "application/octet-stream"
				return nil
			case string:
				// "text/plain"
				return nil
			case url.Values:
				// "application/x-www-form-urlencoded"
				return nil
			default:
			}
		}
	*/
	// 默认 JSON 序列化
	return GetSerializer(MIMEApplicationJSON)
}
