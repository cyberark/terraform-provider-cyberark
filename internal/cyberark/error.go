package cyberark

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func APIErrorFromResponse(code int, body io.ReadCloser) error {
	errorStr := fmt.Sprintf("HTTP status code %d", code)

	var jsonError interface{}
	err := json.NewDecoder(body).Decode(&jsonError)
	if err != nil {
		return errors.New(errorStr)
	}

	if jsonError != nil {
		if jsonErrorStr, ok := jsonError.(string); ok {
			// If the error is a string, just append it to the error message
			errorStr = fmt.Sprintf("%s\n\n%s", errorStr, jsonErrorStr)
		} else {
			prettyJSON, err := json.MarshalIndent(jsonError, "", "  ")
			if err == nil {
				errorStr = fmt.Sprintf("%s\n%s", errorStr, string(prettyJSON))
			}
		}
	}

	return errors.New(errorStr)
}
