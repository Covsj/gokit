package icaptcha

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Covsj/gokit/ihttp"
)

type CapSolver struct {
	ApiKey string
}

type CapSolution struct {
	Object             []bool    `json:"objects,omitempty"`
	Box                []float32 `json:"box,omitempty"`
	ImageSizes         []int32   `json:"imageSize,omitempty"`
	Text               string    `json:"text,omitempty"`
	UserAgent          string    `json:"userAgent,omitempty"`
	ExpireTime         int64     `json:"expireTime,omitempty"`
	GRecaptchaResponse string    `json:"gRecaptchaResponse,omitempty"`
	Challenge          string    `json:"challenge,omitempty"`
	Validate           string    `json:"validate,omitempty"`
	CaptchaId          string    `json:"captcha-id,omitempty"`
	CaptchaOutput      string    `json:"captcha-output,omitempty"`
	GenTime            string    `json:"gen_time,omitempty"`
	LogNumber          string    `json:"log_number,omitempty"`
	PassToken          string    `json:"pass_token,omitempty"`
	RiskType           string    `json:"risk_Type,omitempty"`
	Token              string    `json:"token,omitempty"`
	Cookie             string    `json:"cookie,omitempty"`
	Type               string    `json:"type,omitempty"`
}

type CapResponse struct {
	ErrorId          int32        `json:"errorId"`
	ErrorCode        string       `json:"errorCode"`
	ErrorDescription string       `json:"errorDescription,omitempty"`
	Status           string       `json:"status,omitempty"`
	Solution         *CapSolution `json:"solution,omitempty"`
	TaskId           string       `json:"taskId,omitempty"`
	Balance          float32      `json:"balance,omitempty"`
	Packages         []string     `json:"packages,omitempty"`
}

func (c *CapSolver) request(uri string, jsonBody map[string]any) (*CapResponse, error) {
	capResponse := &CapResponse{}
	_, err := ihttp.Do(&ihttp.Opt{
		URL:     fmt.Sprintf("%s%s", "https://api.capsolver.com", uri),
		Json:    jsonBody,
		RespOut: capResponse,
	})

	if err != nil {
		return nil, err
	}
	return capResponse, nil
}

func (c *CapSolver) Solve(task map[string]any) (*CapResponse, error) {
	capRes, err := c.request("/createTask", map[string]any{
		"clientKey": c.getApiKey(),
		"task":      task,
	})
	if err != nil {
		return nil, err
	}
	if capRes.ErrorId == 1 {
		return nil, errors.New(capRes.ErrorDescription)
	}
	if capRes.Status == "ready" {
		return capRes, nil
	}
	for i := 0; i < 50; i++ {
		capRes, err = c.request("/getTaskResult", map[string]any{
			"clientKey": c.getApiKey(),
			"taskId":    capRes.TaskId,
		})
		if err != nil {
			return nil, err
		}
		if capRes.ErrorId == 1 {
			return nil, errors.New(capRes.ErrorDescription)
		}
		if capRes.Status == "ready" {
			break
		}
		time.Sleep(time.Second * 1)
	}
	return capRes, err
}

func (c *CapSolver) Balance() (*CapResponse, error) {
	capRes, err := c.request("/getBalance", map[string]any{
		"clientKey": c.getApiKey(),
	})
	if err != nil {
		return nil, err
	}
	return capRes, nil
}

func (c *CapSolver) getApiKey() string {
	if c.ApiKey != "" {
		return c.ApiKey
	}
	return os.Getenv("CAPSOLVER_API_KEY")
}
