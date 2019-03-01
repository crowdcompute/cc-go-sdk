// Copyright 2019 The crowdcompute:cc-go-sdk Authors
// This file is part of the crowdcompute:cc-go-sdk library.
//
// The crowdcompute:cc-go-sdk library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The crowdcompute:cc-go-sdk library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the crowdcompute:cc-go-sdk library. If not, see <http://www.gnu.org/licenses/>.

package crowdcompute

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

type CCClient struct {
	url            string
	client         *http.Client
	versionJSONRPC string
	Debug          bool
}

// New create new rpc client with given url
func NewCCClient(url string) *CCClient {
	rpc := &CCClient{
		url:            url,
		client:         http.DefaultClient,
		versionJSONRPC: "2.0",
	}
	return rpc
}

func fatalIfErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s. ERROR: %v", message, err)
	}
}

// rpcError - ethereum error
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err rpcError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

type rpcResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// Call returns raw response of method call
func (rpc *CCClient) call(method string, params ...interface{}) (json.RawMessage, error) {
	request := rpcRequest{
		ID:      1,
		JSONRPC: rpc.versionJSONRPC,
		Method:  method,
		Params:  params,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	response, err := rpc.client.Post(rpc.url, "application/json", bytes.NewBuffer(body))
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if rpc.Debug {
		log.Println(fmt.Sprintf("%s\nRequest: %s, \nResponse: %s\n", method, body, data))
	}
	resp := new(rpcResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, *resp.Error
	}
	return resp.Result, nil
}

// ACCOUNTS
func (rpc *CCClient) CreateAccount(passphrase string) (string, error) {
	res, err := rpc.call("accounts_createAccount", passphrase)
	var account string
	unErr := json.Unmarshal(res, &account)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", account))
	return account, err
}

func (rpc *CCClient) UnlockAccount(acc, passphrase string) (string, error) {
	res, err := rpc.call("accounts_unlockAccount", acc, passphrase)
	var token string
	unErr := json.Unmarshal(res, &token)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", token))
	return token, err
}

func (rpc *CCClient) LockAccount(account, token string) error {
	rpc.client = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		TokenType:   "Bearer",
		AccessToken: token,
	}))
	_, err := rpc.call("accounts_lockAccount", account)
	return err
}

func (rpc *CCClient) DeleteAccount(acc, passphrase string) error {
	_, err := rpc.call("accounts_deleteAccount", acc, passphrase)
	return err
}

func (rpc *CCClient) ListAccounts() ([]string, error) {
	res, err := rpc.call("accounts_listAccounts")
	var accounts []string
	unErr := json.Unmarshal(res, &accounts)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", accounts))
	return accounts, err
}

// // BOOTNODES
func (rpc *CCClient) GetBootnodes() ([]string, error) {
	res, err := rpc.call("bootnodes_getBootnodes")
	var bootnodes []string
	unErr := json.Unmarshal(res, &bootnodes)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", bootnodes))
	return bootnodes, err
}

func (rpc *CCClient) SetBootnodes(nodes []string) error {
	_, err := rpc.call("bootnodes_setBootnodes", nodes)
	return err
}

// // SWARM SERVICE
func (rpc *CCClient) RunSwarmService(service string, nodes []string) error {
	_, err := rpc.call("service_run", service, nodes)
	return err
}

func (rpc *CCClient) StopSwarmService(nodes []string) error {
	_, err := rpc.call("service_stop", nodes)
	return err
}

// DISCOVER NODES
func (rpc *CCClient) DiscoverNodes(num int) (string, error) {
	res, err := rpc.call("discovery_discover", num)
	var msg string
	unErr := json.Unmarshal(res, &msg)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", msg))
	return msg, err
}

// DOCKER IMAGE MANAGER
func (rpc *CCClient) LoadImageToNode(nodeID, imageHash, token string) (string, error) {
	rpc.client = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		TokenType:   "Bearer",
		AccessToken: token,
	}))
	res, err := rpc.call("imagemanager_pushImage", nodeID, imageHash)
	var imgID string
	unErr := json.Unmarshal(res, &imgID)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", imgID))
	return imgID, err
}

func (rpc *CCClient) ExecuteImage(nodeID, dockImageID string) (string, error) {
	res, err := rpc.call("imagemanager_runImage", nodeID, dockImageID)
	var contID string
	unErr := json.Unmarshal(res, &contID)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", contID))
	return contID, err
}

func (rpc *CCClient) InspectContainer(nodeID, containerID string) (string, error) {
	res, err := rpc.call("imagemanager_inspectContainer", nodeID, containerID)
	var inspect string
	unErr := json.Unmarshal(res, &inspect)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", inspect))
	return inspect, err
}

func (rpc *CCClient) ListNodeImages(nodeID, token string) (string, error) {
	rpc.client = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		TokenType:   "Bearer",
		AccessToken: token,
	}))
	res, err := rpc.call("imagemanager_listImages", nodeID)
	var list string
	unErr := json.Unmarshal(res, &list)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", list))
	return list, err
}

func (rpc *CCClient) ListNodeContainers(nodeID, token string) (string, error) {
	rpc.client = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		TokenType:   "Bearer",
		AccessToken: token,
	}))
	res, err := rpc.call("imagemanager_listContainers", nodeID)
	var list string
	unErr := json.Unmarshal(res, &list)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", list))
	return list, err
}

// LEVEL DB
func (rpc *CCClient) LvlDBStats() (string, error) {
	res, err := rpc.call("lvldb_getDBStats")
	var stats string
	unErr := json.Unmarshal(res, &stats)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", stats))
	return stats, err
}

func (rpc *CCClient) LvlDBSelectImage(imageID string) (string, error) {
	res, err := rpc.call("lvldb_selectImage", imageID)
	var image string
	unErr := json.Unmarshal(res, &image)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", image))
	return image, err
}

func (rpc *CCClient) LvlDBSelectImageAccount(imageHash string) (string, error) {
	res, err := rpc.call("lvldb_selectImageAccount", imageHash)
	var image string
	unErr := json.Unmarshal(res, &image)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", image))
	return image, err
}

func (rpc *CCClient) LvlDBSelectType(typeName string) (string, error) {
	res, err := rpc.call("lvldb_selectType", typeName)
	var all string
	unErr := json.Unmarshal(res, &all)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", all))
	return all, err
}

func (rpc *CCClient) LvlDBSelectAll() (string, error) {
	res, err := rpc.call("lvldb_selectAll")
	var all string
	unErr := json.Unmarshal(res, &all)
	fatalIfErr(unErr, fmt.Sprintf("The result is not of type \"%T\" \n", all))
	return all, err
}
