package capsolver

import (
	"fmt"

	"github.com/Covsj/gokit/pkg/ihttp"
)

func request(uri string, solverRequest *Request) (*Response, error) {
	capResponse := &Response{}
	_, err := ihttp.DoRequest(&ihttp.RequestOptions{
		URL:     fmt.Sprintf("%s%s", ApiHost, uri),
		Body:    solverRequest,
		Headers: map[string]string{"Content-Type": "application/json"},
		ResponseOut: capResponse,
	})

	if err != nil {
		return nil, err
	}
	return capResponse, nil
}
