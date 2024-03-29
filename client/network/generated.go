// Package network provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime"
)

const (
	JwtBearerAuthScopes = "jwtBearerAuth.Scopes"
)

// Contact Describes the contact information of a node.
type Contact struct {
	// Address Address of the node.
	Address string `json:"address"`

	// Attempts Number of connection attempts since the node has (re-)started. It is reset to 0 when the connection succeeds.
	Attempts int `json:"attempts"`

	// Did DID of the node.
	Did *string `json:"did,omitempty"`

	// Error Error message of the last connection attempt.
	Error *string `json:"error,omitempty"`

	// LastAttempt Timestamp of the last attempt to contact the address.
	LastAttempt *time.Time `json:"lastAttempt,omitempty"`

	// NextAttempt Timestamp of the next attempt to contact the address.
	NextAttempt *time.Time `json:"nextAttempt,omitempty"`
}

// Event Non-completed event. An event represents a transaction that is of interest to a specific part of the Nuts node.
type Event struct {
	// Error Lists the last error if the event processing failed due to an error.
	Error *string `json:"error,omitempty"`

	// Hash Hash is the ID of the Event, usually the same as the transaction reference.
	Hash string `json:"hash"`

	// LatestNotificationAttempt Timestamp of the most recent notification attempt. Formatted according to RFC3339. Note: calculation of the next attempt does not use this timestamp.
	LatestNotificationAttempt *string `json:"latest_notification_attempt,omitempty"`

	// Retries Number of times the event has been retried.
	Retries int `json:"retries"`

	// Transaction The transaction reference
	Transaction string `json:"transaction"`

	// Type 'transaction' or 'payload'
	Type *string `json:"type,omitempty"`
}

// EventSubscriber Non-completed events for a subscriber
type EventSubscriber struct {
	Events []Event `json:"events"`

	// Name Name of the subscriber component
	Name string `json:"name"`
}

// PeerDiagnostics Diagnostic information of a peer.
type PeerDiagnostics struct {
	// Address Peer's address. This is an IP if it is an inbound connection.
	Address *string `json:"address,omitempty"`

	// Certificate Peer's Certificate
	Certificate *string `json:"certificate,omitempty"`

	// NodeDID Peer's NodeDID
	NodeDID *string `json:"nodeDID,omitempty"`

	// Peers IDs of the peer's peers.
	Peers *[]string `json:"peers,omitempty"`

	// SoftwareID Identification of the particular Nuts implementation of the node. For open source implementations it's recommended to specify URL to the public, open source repository. Proprietary implementations could specify the product or vendor's name.
	SoftwareID *string `json:"softwareID,omitempty"`

	// SoftwareVersion Indication of the software version of the node. It's recommended to use a (Git) commit ID that uniquely resolves to a code revision, alternatively a semantic version could be used (e.g. 1.2.5).
	SoftwareVersion *string `json:"softwareVersion,omitempty"`

	// TransactionNum Number of transactions on the peer's DAG.
	TransactionNum *float32 `json:"transactionNum,omitempty"`

	// Uptime Number of seconds since the node started.
	Uptime *float32 `json:"uptime,omitempty"`
}

// RenderGraphParams defines parameters for RenderGraph.
type RenderGraphParams struct {
	// Start Lamport Clock value from where to start rendering (inclusive). If omitted, rendering starts at the root.
	Start *int `form:"start,omitempty" json:"start,omitempty"`

	// End Lamport Clock value where to stop rendering (exclusive). If omitted, renders the remainder of the graph. Must be larger than the `start` parameter.
	End *int `form:"end,omitempty" json:"end,omitempty"`
}

// ListTransactionsParams defines parameters for ListTransactions.
type ListTransactionsParams struct {
	// Start Inclusive start of range (in lamport clock); default=0
	Start *int `form:"start,omitempty" json:"start,omitempty"`

	// End Exclusive stop of range (in lamport clock); default=∞
	End *int `form:"end,omitempty" json:"end,omitempty"`
}

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetAddressBook request
	GetAddressBook(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RenderGraph request
	RenderGraph(ctx context.Context, params *RenderGraphParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetPeerDiagnostics request
	GetPeerDiagnostics(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListEvents request
	ListEvents(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListTransactions request
	ListTransactions(ctx context.Context, params *ListTransactionsParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetTransaction request
	GetTransaction(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetTransactionPayload request
	GetTransactionPayload(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetAddressBook(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetAddressBookRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RenderGraph(ctx context.Context, params *RenderGraphParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRenderGraphRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetPeerDiagnostics(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetPeerDiagnosticsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListEvents(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListEventsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListTransactions(ctx context.Context, params *ListTransactionsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListTransactionsRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetTransaction(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetTransactionRequest(c.Server, ref)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetTransactionPayload(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetTransactionPayloadRequest(c.Server, ref)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetAddressBookRequest generates requests for GetAddressBook
func NewGetAddressBookRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/network/v1/addressbook")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRenderGraphRequest generates requests for RenderGraph
func NewRenderGraphRequest(server string, params *RenderGraphParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/network/v1/diagnostics/graph")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if params.Start != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "start", runtime.ParamLocationQuery, *params.Start); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.End != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "end", runtime.ParamLocationQuery, *params.End); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetPeerDiagnosticsRequest generates requests for GetPeerDiagnostics
func NewGetPeerDiagnosticsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/network/v1/diagnostics/peers")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListEventsRequest generates requests for ListEvents
func NewListEventsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/network/v1/events")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListTransactionsRequest generates requests for ListTransactions
func NewListTransactionsRequest(server string, params *ListTransactionsParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/network/v1/transaction")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if params.Start != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "start", runtime.ParamLocationQuery, *params.Start); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		if params.End != nil {

			if queryFrag, err := runtime.StyleParamWithLocation("form", true, "end", runtime.ParamLocationQuery, *params.End); err != nil {
				return nil, err
			} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
				return nil, err
			} else {
				for k, v := range parsed {
					for _, v2 := range v {
						queryValues.Add(k, v2)
					}
				}
			}

		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetTransactionRequest generates requests for GetTransaction
func NewGetTransactionRequest(server string, ref string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "ref", runtime.ParamLocationPath, ref)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/network/v1/transaction/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetTransactionPayloadRequest generates requests for GetTransactionPayload
func NewGetTransactionPayloadRequest(server string, ref string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "ref", runtime.ParamLocationPath, ref)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/network/v1/transaction/%s/payload", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetAddressBookWithResponse request
	GetAddressBookWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetAddressBookResponse, error)

	// RenderGraphWithResponse request
	RenderGraphWithResponse(ctx context.Context, params *RenderGraphParams, reqEditors ...RequestEditorFn) (*RenderGraphResponse, error)

	// GetPeerDiagnosticsWithResponse request
	GetPeerDiagnosticsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetPeerDiagnosticsResponse, error)

	// ListEventsWithResponse request
	ListEventsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListEventsResponse, error)

	// ListTransactionsWithResponse request
	ListTransactionsWithResponse(ctx context.Context, params *ListTransactionsParams, reqEditors ...RequestEditorFn) (*ListTransactionsResponse, error)

	// GetTransactionWithResponse request
	GetTransactionWithResponse(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*GetTransactionResponse, error)

	// GetTransactionPayloadWithResponse request
	GetTransactionPayloadWithResponse(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*GetTransactionPayloadResponse, error)
}

type GetAddressBookResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Contact
}

// Status returns HTTPResponse.Status
func (r GetAddressBookResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetAddressBookResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RenderGraphResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	ApplicationproblemJSONDefault *struct {
		// Detail A human-readable explanation specific to this occurrence of the problem.
		Detail string `json:"detail"`

		// Status HTTP statuscode
		Status float32 `json:"status"`

		// Title A short, human-readable summary of the problem type.
		Title string `json:"title"`
	}
}

// Status returns HTTPResponse.Status
func (r RenderGraphResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RenderGraphResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetPeerDiagnosticsResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *map[string]PeerDiagnostics
	ApplicationproblemJSONDefault *struct {
		// Detail A human-readable explanation specific to this occurrence of the problem.
		Detail string `json:"detail"`

		// Status HTTP statuscode
		Status float32 `json:"status"`

		// Title A short, human-readable summary of the problem type.
		Title string `json:"title"`
	}
}

// Status returns HTTPResponse.Status
func (r GetPeerDiagnosticsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetPeerDiagnosticsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListEventsResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *[]EventSubscriber
	ApplicationproblemJSONDefault *struct {
		// Detail A human-readable explanation specific to this occurrence of the problem.
		Detail string `json:"detail"`

		// Status HTTP statuscode
		Status float32 `json:"status"`

		// Title A short, human-readable summary of the problem type.
		Title string `json:"title"`
	}
}

// Status returns HTTPResponse.Status
func (r ListEventsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListEventsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListTransactionsResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	JSON200                       *[]string
	ApplicationproblemJSONDefault *struct {
		// Detail A human-readable explanation specific to this occurrence of the problem.
		Detail string `json:"detail"`

		// Status HTTP statuscode
		Status float32 `json:"status"`

		// Title A short, human-readable summary of the problem type.
		Title string `json:"title"`
	}
}

// Status returns HTTPResponse.Status
func (r ListTransactionsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListTransactionsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetTransactionResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	ApplicationproblemJSONDefault *struct {
		// Detail A human-readable explanation specific to this occurrence of the problem.
		Detail string `json:"detail"`

		// Status HTTP statuscode
		Status float32 `json:"status"`

		// Title A short, human-readable summary of the problem type.
		Title string `json:"title"`
	}
}

// Status returns HTTPResponse.Status
func (r GetTransactionResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetTransactionResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetTransactionPayloadResponse struct {
	Body                          []byte
	HTTPResponse                  *http.Response
	ApplicationproblemJSONDefault *struct {
		// Detail A human-readable explanation specific to this occurrence of the problem.
		Detail string `json:"detail"`

		// Status HTTP statuscode
		Status float32 `json:"status"`

		// Title A short, human-readable summary of the problem type.
		Title string `json:"title"`
	}
}

// Status returns HTTPResponse.Status
func (r GetTransactionPayloadResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetTransactionPayloadResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetAddressBookWithResponse request returning *GetAddressBookResponse
func (c *ClientWithResponses) GetAddressBookWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetAddressBookResponse, error) {
	rsp, err := c.GetAddressBook(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetAddressBookResponse(rsp)
}

// RenderGraphWithResponse request returning *RenderGraphResponse
func (c *ClientWithResponses) RenderGraphWithResponse(ctx context.Context, params *RenderGraphParams, reqEditors ...RequestEditorFn) (*RenderGraphResponse, error) {
	rsp, err := c.RenderGraph(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRenderGraphResponse(rsp)
}

// GetPeerDiagnosticsWithResponse request returning *GetPeerDiagnosticsResponse
func (c *ClientWithResponses) GetPeerDiagnosticsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetPeerDiagnosticsResponse, error) {
	rsp, err := c.GetPeerDiagnostics(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetPeerDiagnosticsResponse(rsp)
}

// ListEventsWithResponse request returning *ListEventsResponse
func (c *ClientWithResponses) ListEventsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListEventsResponse, error) {
	rsp, err := c.ListEvents(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListEventsResponse(rsp)
}

// ListTransactionsWithResponse request returning *ListTransactionsResponse
func (c *ClientWithResponses) ListTransactionsWithResponse(ctx context.Context, params *ListTransactionsParams, reqEditors ...RequestEditorFn) (*ListTransactionsResponse, error) {
	rsp, err := c.ListTransactions(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListTransactionsResponse(rsp)
}

// GetTransactionWithResponse request returning *GetTransactionResponse
func (c *ClientWithResponses) GetTransactionWithResponse(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*GetTransactionResponse, error) {
	rsp, err := c.GetTransaction(ctx, ref, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetTransactionResponse(rsp)
}

// GetTransactionPayloadWithResponse request returning *GetTransactionPayloadResponse
func (c *ClientWithResponses) GetTransactionPayloadWithResponse(ctx context.Context, ref string, reqEditors ...RequestEditorFn) (*GetTransactionPayloadResponse, error) {
	rsp, err := c.GetTransactionPayload(ctx, ref, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetTransactionPayloadResponse(rsp)
}

// ParseGetAddressBookResponse parses an HTTP response from a GetAddressBookWithResponse call
func ParseGetAddressBookResponse(rsp *http.Response) (*GetAddressBookResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetAddressBookResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []Contact
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRenderGraphResponse parses an HTTP response from a RenderGraphWithResponse call
func ParseRenderGraphResponse(rsp *http.Response) (*RenderGraphResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RenderGraphResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest struct {
			// Detail A human-readable explanation specific to this occurrence of the problem.
			Detail string `json:"detail"`

			// Status HTTP statuscode
			Status float32 `json:"status"`

			// Title A short, human-readable summary of the problem type.
			Title string `json:"title"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseGetPeerDiagnosticsResponse parses an HTTP response from a GetPeerDiagnosticsWithResponse call
func ParseGetPeerDiagnosticsResponse(rsp *http.Response) (*GetPeerDiagnosticsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetPeerDiagnosticsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest map[string]PeerDiagnostics
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest struct {
			// Detail A human-readable explanation specific to this occurrence of the problem.
			Detail string `json:"detail"`

			// Status HTTP statuscode
			Status float32 `json:"status"`

			// Title A short, human-readable summary of the problem type.
			Title string `json:"title"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseListEventsResponse parses an HTTP response from a ListEventsWithResponse call
func ParseListEventsResponse(rsp *http.Response) (*ListEventsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListEventsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []EventSubscriber
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest struct {
			// Detail A human-readable explanation specific to this occurrence of the problem.
			Detail string `json:"detail"`

			// Status HTTP statuscode
			Status float32 `json:"status"`

			// Title A short, human-readable summary of the problem type.
			Title string `json:"title"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseListTransactionsResponse parses an HTTP response from a ListTransactionsWithResponse call
func ParseListTransactionsResponse(rsp *http.Response) (*ListTransactionsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListTransactionsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest struct {
			// Detail A human-readable explanation specific to this occurrence of the problem.
			Detail string `json:"detail"`

			// Status HTTP statuscode
			Status float32 `json:"status"`

			// Title A short, human-readable summary of the problem type.
			Title string `json:"title"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseGetTransactionResponse parses an HTTP response from a GetTransactionWithResponse call
func ParseGetTransactionResponse(rsp *http.Response) (*GetTransactionResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetTransactionResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest struct {
			// Detail A human-readable explanation specific to this occurrence of the problem.
			Detail string `json:"detail"`

			// Status HTTP statuscode
			Status float32 `json:"status"`

			// Title A short, human-readable summary of the problem type.
			Title string `json:"title"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}

// ParseGetTransactionPayloadResponse parses an HTTP response from a GetTransactionPayloadWithResponse call
func ParseGetTransactionPayloadResponse(rsp *http.Response) (*GetTransactionPayloadResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetTransactionPayloadResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest struct {
			// Detail A human-readable explanation specific to this occurrence of the problem.
			Detail string `json:"detail"`

			// Status HTTP statuscode
			Status float32 `json:"status"`

			// Title A short, human-readable summary of the problem type.
			Title string `json:"title"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.ApplicationproblemJSONDefault = &dest

	}

	return response, nil
}
