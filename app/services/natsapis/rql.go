package natsapis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NubeDev/flexy/app/services/rqlservice"
	"github.com/NubeDev/flexy/utils/helpers"
	"reflect"
	"strconv"
)

func rqlHandler(body string) ([]byte, error) {
	resp, err := rqlservice.RQL().RunAndDestroy(helpers.UUID(), body, nil)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("rql response was empty")
	}
	typ := resp.ExportType()
	if typ == (reflect.TypeOf(nil)) {
		return nil, errors.New("rql response type was nil")
	}
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue := resp.ToInteger()
		return []byte(strconv.Itoa(int(intValue))), nil
	case reflect.String:
		return []byte(resp.String()), nil
	case reflect.Float64:
		return []byte(strconv.FormatFloat(resp.ToFloat(), 'f', -1, 64)), nil
	case reflect.Bool:
		return []byte(strconv.FormatBool(resp.ToBoolean())), nil
	case reflect.Slice, reflect.Map, reflect.Struct, reflect.Ptr:
		return json.Marshal(resp)
	default:
		fmt.Printf("Type is %s\n", typ.Kind())
	}

	return nil, errors.New("rql response was empty")
}
