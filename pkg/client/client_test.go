package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalPostBody(t *testing.T) {
	postBody, err := unmarshalPostBody(nil)
	assert.NoError(t, err)
	assert.Nil(t, postBody)

	testBody := map[string]interface{}{
		"common_name": fmt.Sprintf("%s.%s.puppetdiscovery.com", "test", "DEV-000-0000-0000"),
		"ttl":         "8765h",
		"exclude_cn_from_sans": true,
		"alt_names":            fmt.Sprintf("%s,localhost", "test"),
	}

	bodyBytes, err := json.Marshal(testBody)
	require.NoError(t, err)
	require.NotNil(t, bodyBytes)

	postBody, err = unmarshalPostBody(bytes.NewReader(bodyBytes))
	assert.NoError(t, err)
	assert.NotNil(t, postBody)

	errBytes := []byte("blah")
	postBody, err = unmarshalPostBody(bytes.NewReader(errBytes))
	assert.Error(t, err)
	assert.Nil(t, postBody)
}
