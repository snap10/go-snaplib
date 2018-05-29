package codingutils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
)

func EncodeToBytes(data interface{}) []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(data)
	return buf.Bytes()
}

func DecodeFromBytes(data []byte, object interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(object)
	if err != nil {
		return err
	}
	return nil
}

func ParseToJson(object interface{}) ([]byte, error) {
	dataJson, err := json.Marshal(object)
	return dataJson, err
}

func ParseToJsonString(object interface{}) (string, error) {
	data, err := ParseToJson(object)
	return BytesToString(data), err
}
func ParseFromJsonReader(jsonString io.ReadCloser, v interface{}) error {
	err := json.NewDecoder(jsonString).Decode(v)
	return err
}
func BytesToString(data []byte) string {
	return string(data[:])
}
