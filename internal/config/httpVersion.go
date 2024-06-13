package config

import (
	"encoding/json"
	"errors"
)

type HTTPVersion string

const (
	HTTP1_0 HTTPVersion = "HTTP/1.0"
	HTTP1_1 HTTPVersion = "HTTP/1.1"
	HTTP2_0 HTTPVersion = "HTTP/2.0"
	HTTP3_0 HTTPVersion = "HTTP/3.0"
)

var allHTTPVersions = []HTTPVersion{HTTP1_0, HTTP1_1, HTTP2_0, HTTP3_0}

func (hv *HTTPVersion) UnmarshalJSON(data []byte) error {
	var version string
	if err := json.Unmarshal(data, &version); err != nil {
		return err
	}

	for _, validVersion := range allHTTPVersions {
		if HTTPVersion(version) == validVersion {
			*hv = HTTPVersion(version)
			return nil
		}
	}
	return errors.New("invalid HTTP version")
}

func (hv HTTPVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(hv))
}
