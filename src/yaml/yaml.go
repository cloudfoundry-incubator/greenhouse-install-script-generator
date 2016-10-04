package yaml

import "gopkg.in/yaml.v2"

func Unmarshal(ba []byte, result interface{}) error {
	return yaml.Unmarshal(ba, result)
	// buf := bytes.NewBuffer(ba)
	// decoder := candiedyaml.NewDecoder(buf)
	// return decoder.Decode(result)
}
