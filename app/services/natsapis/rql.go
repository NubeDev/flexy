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

func rqlHandler(body string) func() ([]byte, error) {
	return func() ([]byte, error) {
		resp, err := rqlservice.RQL().RunAndDestroy(helpers.UUID(), body, nil)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, errors.New("rql response was empty")
		}
		typ := resp.ExportType()
		fmt.Printf("Type is %s\n", typ.Kind())
		switch typ.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue := resp.ToInteger()
			return []byte(strconv.Itoa(int(intValue))), nil
		case reflect.String:
			fmt.Println("Type is string")
		case reflect.Float64:
			fmt.Println("Type is float64")
		case reflect.Bool:
			fmt.Println("Type is bool")
		case reflect.Slice:
			return json.Marshal(resp)
		case reflect.Map:
			return json.Marshal(resp)
		case reflect.Struct:
			return json.Marshal(resp)
		case reflect.Ptr:
			return json.Marshal(resp)
		default:
			fmt.Printf("Type is %s\n", typ.Kind())
		}
		return nil, errors.New("rql response was empty")
	}

}
