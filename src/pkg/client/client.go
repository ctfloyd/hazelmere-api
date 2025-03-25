package client

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ctfloyd/hazelmere-api/src/pkg/api"
	jsoniter "github.com/json-iterator/go"
	"io"
	"math"
	"net/http"
	"time"
)

var ErrHazelmereClient = errors.New("hazelmere client error")

type HazelmereClientConfig struct {
	Host           string
	TimeoutMs      int
	Retries        int
	RetryWaitMs    int
	RetryMaxWaitMs int
}
type HazelmereClient struct {
	config      HazelmereClientConfig
	client      *http.Client
	errorMap    map[string]error
	errorLogger func(string)
}

func NewHazelmereClient(config HazelmereClientConfig, errorLogger func(string)) *HazelmereClient {
	httpClient := http.Client{
		Timeout: time.Duration(config.TimeoutMs) * time.Millisecond,
	}

	return &HazelmereClient{
		config:      config,
		client:      &httpClient,
		errorMap:    make(map[string]error),
		errorLogger: errorLogger,
	}
}

func (hc *HazelmereClient) AddErrorMappings(errors map[string]error) {
	for k, v := range errors {
		hc.errorMap[k] = v
	}
}

func (hc *HazelmereClient) GetHost() string {
	return hc.config.Host
}

func (hc *HazelmereClient) GetV1Url() string {
	return fmt.Sprintf("%s/%s", hc.config.Host, "v1")
}

func (hc *HazelmereClient) Get(url string, response any) error {
	return hc.GetWithHeaders(url, nil, response)
}

func (hc *HazelmereClient) GetWithHeaders(url string, headers map[string]string, response any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Join(err, ErrHazelmereClient)
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return hc.doRequest(req, response)
}

func (hc *HazelmereClient) Post(url string, body any, response any) error {
	return hc.PostWithHeaders(url, nil, body, response)
}

func (hc *HazelmereClient) PostWithHeaders(url string, headers map[string]string, body any, response any) error {
	bodyBytes, err := jsoniter.Marshal(body)
	if err != nil {
		return errors.Join(ErrHazelmereClient, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.Join(err, ErrHazelmereClient)
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return hc.doRequest(req, response)
}

func (hc *HazelmereClient) Patch(url string, body any, response any) error {
	return hc.PatchWithHeaders(url, nil, body, response)
}

func (hc *HazelmereClient) PatchWithHeaders(url string, headers map[string]string, body any, response any) error {
	bodyBytes, err := jsoniter.Marshal(body)
	if err != nil {
		return errors.Join(ErrHazelmereClient, err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.Join(err, ErrHazelmereClient)
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	return hc.doRequest(req, response)
}

func (hc *HazelmereClient) doRequest(request *http.Request, response any) error {
	attempt := 0
	for attempt < hc.config.Retries+1 {
		res, err := hc.client.Do(request)
		if err != nil {
			return errors.Join(err, ErrHazelmereClient)
		}

		if res.StatusCode >= 200 && res.StatusCode <= 299 {
			return hc.parseSuccessResponse(res, response)
		}

		if res.StatusCode <= 499 {
			return hc.handleNonRetryableErrorResponse(res)
		}

		time.Sleep(time.Duration(hc.computeWaitMs(attempt)) * time.Millisecond)
		attempt += 1
	}
	return errors.New("maximum retry attempts exceeded")
}

func (hc *HazelmereClient) parseSuccessResponse(res *http.Response, response any) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			hc.errorLogger(err.Error() + "\n")
		}
	}(res.Body)

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.Join(err, ErrHazelmereClient)
	}

	err = jsoniter.Unmarshal(responseBytes, &response)
	if err != nil {
		return errors.Join(err, ErrHazelmereClient)
	}

	return nil
}

func (hc *HazelmereClient) handleNonRetryableErrorResponse(res *http.Response) error {
	errorResponse, err := hc.parseErrorResponse(res)
	if err != nil {
		return errors.Join(ErrHazelmereClient, err)
	}

	if value, ok := hc.errorMap[errorResponse.Code]; ok {
		return errors.Join(value, errors.New(errorResponse.Message))
	} else {
		return errors.Join(ErrHazelmereClient, errors.New(fmt.Sprintf("[%s] - %s", errorResponse.Code, errorResponse.Message)))
	}
}

func (hc *HazelmereClient) parseErrorResponse(res *http.Response) (api.ErrorResponse, error) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			hc.errorLogger(err.Error())
		}
	}(res.Body)

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return api.ErrorResponse{}, errors.Join(err, ErrHazelmereClient)
	}

	var errorResponse api.ErrorResponse
	err = jsoniter.Unmarshal(responseBytes, &errorResponse)
	if err != nil {
		return api.ErrorResponse{}, errors.Join(err, ErrHazelmereClient)
	}

	return errorResponse, nil
}

func (hc *HazelmereClient) computeWaitMs(attempt int) int {
	return int(math.Min(float64(attempt*hc.config.RetryWaitMs), float64(hc.config.RetryMaxWaitMs)))
}
