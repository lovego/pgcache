package pgcache

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/lovego/sorted_sets"
)

type Data struct {
	*sync.RWMutex
	// DataPtr is a pointer to a map or slice to store data, required.
	DataPtr interface{}
	// MapKeys is the field names to get map keys from row struct, required if DataPtr is a map.
	MapKeys []string
	// Value is the field name to get map or slice value from row struct.
	// If it's empty, the whole row struct is used.
	Value string

	// If the DataPtr or map value is a slice, it's used as sorted set. If it's a sorted set of struct,
	// SortedSetUniqueKey is required, it specifies the fields used as unique key.
	SortedSetUniqueKey []string

	// Preprocess is optional. It's a method name of row struct. It should be of "func ()" form.
	// It is called before Precond method is called.
	Preprocess string
	// Precond is optional. It's a method name of row struct. It should be of "func () bool" form.
	// It is called before handling, if the return value is false, no handling(save or remove) is performed.
	Precond string

	// for cache manage
	manageKey  string
	manageType string

	// the data value to store data
	dataV reflect.Value
	// map value is a sorted set
	isSortedSets bool
	// real map value is a pointer of the row struct or row struct's {Value} field.
	realValueIsPointer bool
	// negative if no Preprocess present.
	preprocessMethodIndex int
	// negative if no Precond present.
	precondMethodIndex int
}

func (d *Data) save(row reflect.Value) {
	d.preprocess(row)
	if !d.precond(row) {
		return
	}
	d.Lock()
	defer d.Unlock()

	if d.dataV.Kind() == reflect.Slice {
		d.dataV.Set(sorted_sets.SaveValue(d.dataV, d.getValue(row), d.SortedSetUniqueKey...))
	} else {
		d.saveToMap(row)
	}
}

func (d *Data) saveToMap(row reflect.Value) {
	mapV := d.dataV
	if mapV.IsNil() {
		mapV.Set(reflect.MakeMap(mapV.Type()))
	}
	for i := 0; i < len(d.MapKeys)-1; i++ {
		key := row.FieldByName(d.MapKeys[i])
		value := mapV.MapIndex(key)
		if !value.IsValid() {
			value = reflect.MakeMap(mapV.Type().Elem())
			mapV.SetMapIndex(key, value)
		} else if value.IsNil() {
			value.Set(reflect.MakeMap(mapV.Type().Elem()))
		}
		mapV = value
	}

	key := row.FieldByName(d.MapKeys[len(d.MapKeys)-1])
	value := d.getValue(row)
	if d.isSortedSets {
		value = sorted_sets.SaveValue(mapV.MapIndex(key), value, d.SortedSetUniqueKey...)
	}
	mapV.SetMapIndex(key, value)
}

func (d *Data) remove(row reflect.Value) {
	d.preprocess(row)
	if !d.precond(row) {
		return
	}
	d.Lock()
	defer d.Unlock()

	if d.dataV.Kind() == reflect.Slice {
		d.dataV.Set(sorted_sets.RemoveValue(d.dataV, d.getValue(row), d.SortedSetUniqueKey...))
	} else {
		d.removeFromMap(row)
	}
}

func (d *Data) removeFromMap(row reflect.Value) {
	mapV := d.dataV
	for i := 0; i < len(d.MapKeys)-1; i++ {
		key := row.FieldByName(d.MapKeys[i])
		mapV = mapV.MapIndex(key)
		if !mapV.IsValid() || mapV.IsNil() {
			return
		}
	}
	key := row.FieldByName(d.MapKeys[len(d.MapKeys)-1])
	if d.isSortedSets {
		slice := mapV.MapIndex(key)
		if !slice.IsValid() {
			return
		}
		slice = sorted_sets.RemoveValue(slice, d.getValue(row), d.SortedSetUniqueKey...)
		if !slice.IsValid() || slice.Len() == 0 {
			mapV.SetMapIndex(key, reflect.Value{})
		} else {
			mapV.SetMapIndex(key, slice)
		}
	} else {
		mapV.SetMapIndex(key, reflect.Value{})
	}
}

func (d *Data) clear() {
	d.Lock()
	defer d.Unlock()
	if d.dataV.Kind() == reflect.Slice {
		d.dataV.Set(reflect.MakeSlice(d.dataV.Type(), 0, d.dataV.Cap()))
	} else {
		d.dataV.Set(reflect.MakeMap(d.dataV.Type()))
	}
}

func (d *Data) getValue(row reflect.Value) reflect.Value {
	value := row
	if d.Value != "" {
		value = row.FieldByName(d.Value)
	}
	if d.realValueIsPointer {
		value = value.Addr()
	}
	return value
}

func (d *Data) preprocess(row reflect.Value) {
	if d.preprocessMethodIndex < 0 {
		return
	}
	row.Addr().Method(d.preprocessMethodIndex).Call(nil)
}

func (d *Data) precond(row reflect.Value) bool {
	if d.precondMethodIndex < 0 {
		return true
	}
	out := row.Addr().Method(d.precondMethodIndex).Call(nil)
	return out[0].Bool()
}

func (d *Data) Key() string {
	if d.manageKey == `` {
		d.manageKey = addKeyValueNames(d.dataV.Type().String(), d.MapKeys, d.Value)
		if d.Precond != "" {
			d.manageKey += fmt.Sprintf("(%s)", d.Precond)
		}
	}
	return d.manageKey
}

func (d *Data) Size() int {
	return d.dataV.Len()
}

func (d *Data) Data(keys ...string) (interface{}, error) {
	if len(keys) == 0 {
		return d.DataPtr, nil
	}
	var data = d.dataV
	for _, str := range keys {
		switch data.Kind() {
		case reflect.Map:
			if key, err := convertStrToType(str, data.Type().Key()); err != nil {
				return nil, err
			} else {
				data = data.MapIndex(key)
			}
		case reflect.Slice, reflect.Array:
			if index, err := strconv.Atoi(str); err != nil {
				return nil, err
			} else if index >= data.Len() {
				return nil, fmt.Errorf("Index %d out of range: 0 ~ %d.", index, data.Len())
			} else {
				data = data.Index(index)
			}
		default:
			return nil, fmt.Errorf("Not map/slice/array for key: %s", str)
		}
	}
	if !data.IsValid() {
		return nil, errors.New("No such value found.")
	}
	return data.Interface(), nil
}

var mapKeyRegexp = regexp.MustCompile(`\[\w+\]`)

func addKeyValueNames(mapType string, keyNames []string, valueName string) string {
	i := 0
	mapType = mapKeyRegexp.ReplaceAllStringFunc(mapType, func(submatch string) string {
		if i >= len(keyNames) {
			return submatch
		}
		name := keyNames[i]
		i++
		return submatch[:1] + name + ":" + submatch[1:]
	})
	if valueName != `` {
		if index := strings.LastIndexByte(mapType, ']'); index > 0 {
			return mapType[:index+1] + valueName + ":" + mapType[index+1:]
		}
	}
	return mapType
}

func convertStrToType(str string, typ reflect.Type) (reflect.Value, error) {
	var value interface{}
	var err error
	switch typ.Kind() {
	case reflect.String:
		value = str
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err = strconv.ParseInt(str, 10, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err = strconv.ParseUint(str, 10, 64)
	case reflect.Bool:
		value, err = strconv.ParseBool(str)
	default:
		return reflect.Value{}, fmt.Errorf("don't know how to convert string to %v", typ)
	}
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(value).Convert(typ), nil
}
