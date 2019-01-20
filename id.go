package groundcontrol

import (
	"encoding/base64"
	"strings"
)

func EncodeID(identifiers ...string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(identifiers, ":")))
}

func DecodeID(id string) ([]string, error) {
	bytes, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(bytes), ":"), nil
}
