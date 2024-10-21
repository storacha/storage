package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const pdpRoutePath = "/pdp"
const proofSetsPath = "/proof-sets"
const piecePath = "/piece"
const pingPath = "/ping"
const rootsPath = "/roots"

type ErrFailedResponse struct {
	StatusCode int
	Body       string
}

func errFromResponse(res *http.Response) ErrFailedResponse {
	err := ErrFailedResponse{StatusCode: res.StatusCode}

	message, merr := io.ReadAll(res.Body)
	if merr != nil {
		err.Body = merr.Error()
	} else {
		err.Body = string(message)
	}
	return err
}

func (e ErrFailedResponse) Error() string {
	return fmt.Sprintf("http request failed, status: %d %s, message: %s", e.StatusCode, http.StatusText(e.StatusCode), e.Body)
}

type (
	StatusRef struct {
		url string
	}

	CreateProofSet struct {
		RecordKeeper string `json:"recordKeeper"`
	}

	ProofSetStatus struct {
		CreateMessageHash string  `json:"createMessageHash"`
		ProofsetCreated   bool    `json:"proofsetCreated"`
		Service           string  `json:"service"`
		TxStatus          string  `json:"txStatus"`
		OK                *bool   `json:"ok"`
		ProofSetId        *uint64 `json:"proofSetId,omitempty"`
	}

	// RootEntry represents a root in the proof set for JSON serialization
	RootEntry struct {
		RootID        uint64 `json:"rootId"`
		RootCID       string `json:"rootCid"`
		SubrootCID    string `json:"subrootCid"`
		SubrootOffset int64  `json:"subrootOffset"`
	}

	ProofSet struct {
		ID                 uint64      `json:"id"`
		NextChallengeEpoch *int64      `json:"nextChallengeEpoch"`
		Roots              []RootEntry `json:"roots"`
	}

	SubrootEntry struct {
		SubrootCID string `json:"subrootCid"`
	}

	AddRoot struct {
		RootCID  string         `json:"rootCid"`
		Subroots []SubrootEntry `json:"subroots"`
	}

	UploadRef struct {
		url string
	}

	AddPiece struct {
		PieceCID string `json:"pieceCid"`
		Notify   string `json:"notify,omitempty"`
	}

	Client struct {
		authHeader string
		endpoint   *url.URL
		client     *http.Client
	}
)

func New(client *http.Client, endpoint *url.URL, authHeader string) *Client {
	return &Client{
		authHeader: authHeader,
		endpoint:   endpoint,
		client:     client,
	}
}

func (c *Client) Ping() error {
	url := c.endpoint.JoinPath(pdpRoutePath, pingPath).String()
	return c.verifySuccess(c.sendRequest(http.MethodGet, url, nil))
}

func (c *Client) CreateProofSet(request CreateProofSet) (StatusRef, error) {
	url := c.endpoint.JoinPath(pdpRoutePath, proofSetsPath).String()
	// send request
	res, err := c.postJson(url, request)
	if err != nil {
		return StatusRef{}, err
	}
	// all successful responses are 201
	if res.StatusCode != http.StatusCreated {
		return StatusRef{}, errFromResponse(res)
	}

	return StatusRef{url: res.Header.Get("Location")}, nil
}

func (c *Client) ProofSetCreationStatus(ref StatusRef) (ProofSetStatus, error) {
	// we could do this in a number of ways, including having StatusRef actually
	// just be the TXHash, extracted from the location header. But ultimately
	// it makes the most sense as an opaque reference from the standpoint of anyone
	// using the client
	// generate request
	var proofSetStatus ProofSetStatus
	err := c.getJsonResponse(ref.url, &proofSetStatus)
	return proofSetStatus, err
}

func (c *Client) GetProofSet(id uint64) (ProofSet, error) {
	url := c.endpoint.JoinPath(pdpRoutePath, proofSetsPath, "/", strconv.FormatUint(id, 10)).String()
	var proofSet ProofSet
	err := c.getJsonResponse(url, &proofSet)
	return proofSet, err
}

func (c *Client) DeleteProofSet(id uint64) error {
	url := c.endpoint.JoinPath(pdpRoutePath, proofSetsPath, strconv.FormatUint(id, 10)).String()
	return c.verifySuccess(c.sendRequest(http.MethodDelete, url, nil))
}

func (c *Client) AddRootToProofSet(id uint64, addRoot AddRoot) error {
	url := c.endpoint.JoinPath(pdpRoutePath, proofSetsPath, "/", strconv.FormatUint(id, 10), rootsPath).String()
	return c.verifySuccess(c.postJson(url, addRoot))
}

func (c *Client) AddPiece(addPiece AddPiece) (*UploadRef, error) {
	url := c.endpoint.JoinPath(pdpRoutePath, piecePath).String()
	res, err := c.postJson(url, addPiece)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if res.StatusCode == http.StatusCreated {
		return &UploadRef{
			url: res.Header.Get("Location"),
		}, nil
	}
	return nil, errFromResponse(res)
}

func (c *Client) UploadPiece(ref UploadRef, data io.Reader) error {
	return c.verifySuccess(c.sendRequest(http.MethodPut, ref.url, data))
}

func (c *Client) GetPiece(pieceCid string) (io.ReadCloser, error) {
	// piece gets are not at the pdp path but rather the raw /piece path
	url := c.endpoint.JoinPath(piecePath, "/", pieceCid).String()
	res, err := c.sendRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errFromResponse(res)
	}
	return res.Body, nil
}

func (c *Client) sendRequest(method string, url string, body io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("generating http request: %w", err)
	}
	// add authorization header
	req.Header.Add("Authorization", c.authHeader)
	// send request
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to curio: %w", err)
	}
	return res, nil
}

func (c *Client) postJson(url string, params interface{}) (*http.Response, error) {
	var body io.Reader
	if params != nil {
		asBytes, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("encoding request parameters: %w", err)
		}
		body = bytes.NewReader(asBytes)
	}
	return c.sendRequest(http.MethodPost, url, body)
}

func (c *Client) getJsonResponse(url string, target interface{}) error {
	res, err := c.sendRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errFromResponse(res)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("unmarshalling JSON response to target: %w", err)
	}
	return nil
}

func (c *Client) verifySuccess(res *http.Response, err error) error {
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errFromResponse(res)
	}
	return nil
}
