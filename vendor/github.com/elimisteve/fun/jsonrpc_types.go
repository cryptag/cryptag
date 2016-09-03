// Steve Phillips / elimisteve
// 2012.04.29
// Originally part of Decentra prototype

package fun

import (
	"encoding/json"
)

type JsonRpc1Request struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Id     interface{}   `json:"id"`
}

type JsonRpc1Response struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	Id     interface{} `json:"id"`
}

// From http://golang.org/src/pkg/net/rpc/jsonrpc/client.go

type ClientRequest struct {
	Method string         `json:"method"`
	Params [1]interface{} `json:"params"`
	Id     uint64         `json:"id"`
}

type ClientResponse struct {
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
	Id     uint64           `json:"id"`
}
