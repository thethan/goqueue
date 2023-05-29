package job

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestJsonJob_GetValue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			jsonSting := `{"name":"test","id": 1, "type":"json","payload":{"test":"test"}}`

			j, err := makeJsonJob(&Configuration{}, []byte(jsonSting))
			require.Nil(t, err)

			value, err := j.GetValue("name")
			_ = value
			require.Equal(t, reflect.String, value.Kind())
		})
		t.Run("nested object", func(t *testing.T) {
			jsonSting := `{"name":"test","id": 1, "type":"json","payload":{"test":"test"}}`

			j, err := makeJsonJob(&Configuration{}, []byte(jsonSting))
			require.Nil(t, err)

			value, err := j.GetValue("payload.test")
			_ = value
			require.Equal(t, reflect.String, value.Kind())
		})
	})
}

func TestJsonJob_unmarshal(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			jsonSting := `{"name":"test","id": 1, "type":"json","payload":{"test":"test"}}`

			j, err := makeJsonJob(&Configuration{}, []byte(jsonSting))
			require.Nil(t, err)

			val, ok := j.jsonMap["name"]
			assert.True(t, ok)
			assert.Equal(t, "test", val)
		})
		t.Run("string", func(t *testing.T) {
			jsonSting := `{"name":"test","id": 1, "type":"json","payload":{"test":"test"}}`

			j, err := makeJsonJob(&Configuration{}, []byte(jsonSting))
			require.Nil(t, err)

			val, ok := j.jsonMap["name"]
			assert.True(t, ok)
			assert.Equal(t, "test", val)
		})
	})
}

func TestJsonJob_validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			jsonSting := `{"name":"test","id": 1, "type":"json","payload":{"test":"test"}}`

			j, err := makeJsonJob(&Configuration{}, []byte(jsonSting))
			require.Nil(t, err)

			val, ok := j.jsonMap["name"]
			assert.True(t, ok)
			assert.Equal(t, "test", val)
		})
	})
	t.Run("failure", func(t *testing.T) {
		t.Run("string", func(t *testing.T) {
			jsonSting := `{"name":1,"id": 1, "type":"json","payload":{"test":"test"}}`

			j, err := makeJsonJob(&Configuration{}, []byte(jsonSting))
			require.Nil(t, err)

			val, ok := j.jsonMap["name"]
			assert.True(t, ok)
			assert.NotEqual(t, "test", val)
		})
	})
}
