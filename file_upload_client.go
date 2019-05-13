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
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

type UploadClient struct {
	url    string
	client *http.Client
	Debug  bool
}

// New create new rpc client with given url
func NewUploadClient(url string) *UploadClient {
	rpc := &UploadClient{
		url:    url,
		client: http.DefaultClient,
	}
	return rpc
}

func (c *UploadClient) UploadFile(filename, token string) (string, error) {
	c.client = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		TokenType:   "Bearer",
		AccessToken: token,
	}))

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return "", err
	}
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return "", err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return "", err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	resp, err := c.client.Post(c.url, contentType, bodyBuf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}
