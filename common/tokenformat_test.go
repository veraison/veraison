package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TokenFormat_FromString(t *testing.T) {
	assert := assert.New(t)

	tf, err := TokenFormatFromString("psa")
	assert.Nil(err)
	assert.Equal(PsaIatToken, tf)

	tf, err = TokenFormatFromString("0")
	assert.Nil(err)
	assert.Equal(PsaIatToken, tf)

	tf, err = TokenFormatFromString("TokenFormat(123)")
	assert.Nil(err)
	assert.Equal(TokenFormat(123), tf)

	_, err = TokenFormatFromString("TokenFormat(abc)")
	assert.Contains(err.Error(), "invalid TokenFormat value")
}

func Test_TokenFormat_MarshalJSON(t *testing.T) {
	assert := assert.New(t)

	val, err := json.Marshal(PsaIatToken)
	assert.Nil(err)
	assert.Equal([]byte(`"psa"`), val)

	val, err = json.Marshal(TokenFormat(123))
	assert.Nil(err)
	assert.Equal([]byte(`"TokenFormat(123)"`), val)
}

func Test_TokenFormat_UnmarshalJSON(t *testing.T) {
	assert := assert.New(t)

	var tf TokenFormat

	err := json.Unmarshal([]byte(`"psa"`), &tf)
	assert.Nil(err)
	assert.Equal(PsaIatToken, tf)

	err = json.Unmarshal([]byte(`0`), &tf)
	assert.Nil(err)
	assert.Equal(PsaIatToken, tf)

	err = json.Unmarshal([]byte(`1.23`), &tf)
	assert.Contains(err.Error(), "non-integer numeric value")

	err = json.Unmarshal([]byte(`"hello"`), &tf)
	assert.Contains(err.Error(), "invalid TokenFormat value")

	err = json.Unmarshal([]byte(`true`), &tf)
	assert.Contains(err.Error(), "unexpected value for TokenFormat")
}
