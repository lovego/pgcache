package logger

import (
	"encoding/json"
	"fmt"
)

var jsonFormatter jsonFmt
var readableFormatter readableFmt

type Formatter interface {
	Format(map[string]interface{}) ([]byte, error)
}

type jsonFmt struct {
}

func (jf jsonFmt) Format(fields map[string]interface{}) ([]byte, error) {
	return json.Marshal(fields)
}

type readableFmt struct {
}

func (jif readableFmt) Format(fields map[string]interface{}) ([]byte, error) {
	msg, stack := fields["msg"], fields["stack"]
	if msg != nil || stack != nil {
		delete(fields, "msg")
		delete(fields, "stack")
		buf, err := json.MarshalIndent(fields, "", "  ")
		fields["msg"], fields["stack"] = msg, stack
		if err != nil {
			return nil, err
		}
		return append([]byte(fmt.Sprintf("%v\n%v\n", msg, stack)), buf...), nil
	}
	return json.MarshalIndent(fields, "", "  ")
}
