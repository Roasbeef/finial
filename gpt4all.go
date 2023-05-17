package main

import (
	"encoding/json"
	"io"

	"github.com/rakyll/openai-go/chat"
)

type GPT4AllRequest struct {
	*chat.CreateCompletionParams
}

func UmarshalJsonReq(jsonReader io.Reader) (*GPT4AllRequest, error) {
	req := &GPT4AllRequest{}

	decoder := json.NewDecoder(jsonReader)
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}
