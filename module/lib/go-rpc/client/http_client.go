// Copyright 2017 Baptist-Publication Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpcclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"go.uber.org/zap"

	. "github.com/Baptist-Publication/chorus/module/lib/go-common"
	"github.com/Baptist-Publication/chorus/module/lib/go-rpc/types"
	"github.com/Baptist-Publication/chorus/module/lib/go-wire"
)

// TODO: Deprecate support for IP:PORT or /path/to/socket
func makeHTTPDialer(logger *zap.Logger, remoteAddr string) (string, func(string, string) (net.Conn, error)) {

	parts := strings.SplitN(remoteAddr, "://", 2)
	var protocol, address string
	if len(parts) != 2 {
		logger.Warn("WARNING (go-rpc): Please use fully formed listening addresses, including the tcp:// or unix:// prefix")
		protocol = rpctypes.SocketType(remoteAddr)
		address = remoteAddr
	} else {
		protocol, address = parts[0], parts[1]
	}

	trimmedAddress := strings.Replace(address, "/", ".", -1) // replace / with . for http requests (dummy domain)
	return trimmedAddress, func(proto, addr string) (net.Conn, error) {
		return net.Dial(protocol, address)
	}
}

// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://)
func makeHTTPClient(logger *zap.Logger, remoteAddr string) (string, *http.Client) {
	address, dialer := makeHTTPDialer(logger, remoteAddr)
	return "http://" + address, &http.Client{
		Transport: &http.Transport{
			Dial: dialer,
		},
	}
}

//------------------------------------------------------------------------------------

type Client interface {
}

//------------------------------------------------------------------------------------

// JSON rpc takes params as a slice
type ClientJSONRPC struct {
	address string
	client  *http.Client
}

func NewClientJSONRPC(logger *zap.Logger, remote string) *ClientJSONRPC {
	address, client := makeHTTPClient(logger, remote)
	return &ClientJSONRPC{
		address: address,
		client:  client,
	}
}

func (c *ClientJSONRPC) Call(method string, params []interface{}, result interface{}) (interface{}, error) {
	return c.call(method, params, result)
}

func (c *ClientJSONRPC) call(method string, params []interface{}, result interface{}) (interface{}, error) {
	// Make request and get responseBytes
	request := rpctypes.RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      "",
	}
	requestBytes := wire.JSONBytes(request)
	requestBuf := bytes.NewBuffer(requestBytes)
	// log.Info(Fmt("RPC request to %v (%v): %v", c.remote, method, string(requestBytes)))
	httpResponse, err := c.client.Post(c.address, "text/json", requestBuf)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()
	responseBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	// 	log.Info(Fmt("RPC response: %v", string(responseBytes)))
	return unmarshalResponseBytes(responseBytes, result)
}

//-------------------------------------------------------------

// URI takes params as a map
type ClientURI struct {
	address string
	client  *http.Client
}

func NewClientURI(logger *zap.Logger, remote string) *ClientURI {
	address, client := makeHTTPClient(logger, remote)
	return &ClientURI{
		address: address,
		client:  client,
	}
}

func (c *ClientURI) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	return c.call(method, params, result)
}

func (c *ClientURI) call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	values, err := argsToURLValues(params)
	if err != nil {
		return nil, err
	}
	// log.Info(Fmt("URI request to %v (%v): %v", c.address, method, values))
	resp, err := c.client.PostForm(c.address+"/"+method, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return unmarshalResponseBytes(responseBytes, result)
}

//------------------------------------------------

func unmarshalResponseBytes(responseBytes []byte, result interface{}) (interface{}, error) {
	// read response
	// if rpc/core/types is imported, the result will unmarshal
	// into the correct type
	// log.Notice("response", "response", string(responseBytes))
	var err error
	response := &rpctypes.RPCResponse{}
	response.Result = result
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return nil, errors.New(Fmt("Error unmarshalling rpc response: %v", err))
	}
	errorStr := response.Error
	if errorStr != "" {
		return nil, errors.New(Fmt("Response error: %v", errorStr))
	}
	return response.Result, nil
}

func argsToURLValues(args map[string]interface{}) (url.Values, error) {
	values := make(url.Values)
	if len(args) == 0 {
		return values, nil
	}
	err := argsToJson(args)
	if err != nil {
		return nil, err
	}
	for key, val := range args {
		values.Set(key, val.(string))
	}
	return values, nil
}

func argsToJson(args map[string]interface{}) error {
	var n int
	var err error
	for k, v := range args {
		// Convert byte slices to "0x"-prefixed hex
		byteSlice, isByteSlice := reflect.ValueOf(v).Interface().([]byte)
		if isByteSlice {
			args[k] = fmt.Sprintf("0x%X", byteSlice)
			continue
		}

		// Pass everything else to go-wire
		buf := new(bytes.Buffer)
		wire.WriteJSON(v, buf, &n, &err)
		if err != nil {
			return err
		}
		args[k] = buf.String()
	}
	return nil
}
