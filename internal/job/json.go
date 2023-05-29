package job

import (
	"encoding/json"
	"reflect"
	"strings"
)

type jsonJob struct {
	config  *Configuration
	bytes   []byte
	jsonMap map[string]interface{}
}

func (j *jsonJob) GetValue(key string) (reflect.Value, error) {
	keys := strings.Split(key, ".")
	jsonMap := j.jsonMap
	return getKindAndValueFromJsonMapFromKeys(keys, jsonMap)
}

func (j *jsonJob) Raw() []byte {
	return j.bytes
}

func getKindAndValueFromJsonMapFromKeys(keys []string, jsonMap map[string]interface{}) (reflect.Value, error) {
	val := reflect.Value{}
	var err error
	for idx := range keys {
		k := keys[idx]
		val, err = getKindAndValueFromJsonMapFromKey(k, jsonMap)
		if err != nil {
			return reflect.Value{}, err
		}

		if val.Kind() == reflect.Map {
			jsonMap = val.Interface().(map[string]interface{})
		}
	}

	return val, err
}

func getKindAndValueFromJsonMapFromKey(key string, jsonMap map[string]interface{}) (reflect.Value, error) {
	val, ok := jsonMap[key]
	if !ok {
		return reflect.Value{}, nil
	}

	reflectVal := reflect.ValueOf(val)

	return reflectVal, nil
}
func makeJsonJob(config *Configuration, bts []byte) (*jsonJob, error) {
	jJob := jsonJob{
		config:  config,
		bytes:   bts,
		jsonMap: map[string]interface{}{},
	}

	err := unmarshalJsonJob(&jJob)
	if err != nil {
		return nil, err
	}

	return &jJob, nil
}

func unmarshalJsonJob(j *jsonJob) error {
	err := json.Unmarshal(j.bytes, &j.jsonMap)
	if err != nil {
		return err
	}

	return nil
}

func (j *jsonJob) validate() bool {
	return false
}
