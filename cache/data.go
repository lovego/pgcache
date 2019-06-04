package cache

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/lovego/sorted_sets"
)

type Data struct {
	*sync.RWMutex
	// MapPtr is a pointer to a map to store data, required.
	MapPtr interface{}
	// MapKeys is the field names to get map keys from row struct, required.
	MapKeys []string
	// MapValue is the field name to get map value from row struct.
	// If it's empty, the whole row struct is used as map value.
	MapValue string

	// If the map value is a slice, it's used as sorted set. If it's a sorted set of struct,
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

	// the map value to store data
	mapV reflect.Value
	// map value is a sorted set
	isSortedSets bool
	// real map value is a pointer of the row struct or row struct's {MapValue} field.
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

	mapV := d.mapV
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
	value := row
	if d.MapValue != "" {
		value = row.FieldByName(d.MapValue)
	}
	if d.realValueIsPointer {
		value = value.Addr()
	}
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

	mapV := d.mapV
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
		value := row
		if d.MapValue != "" {
			value = row.FieldByName(d.MapValue)
		}
		slice = sorted_sets.RemoveValue(slice, value, d.SortedSetUniqueKey...)
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
	d.mapV.Set(reflect.MakeMap(d.mapV.Type()))
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
		d.manageKey = addKeyValueNames(d.mapV.Type().String(), d.MapKeys, d.MapValue)
	}
	return d.manageKey
}

func (d *Data) Size() int {
	return d.mapV.Len()
}

func (d *Data) Data() interface{} {
	return d.MapPtr
}

func addKeyValueNames(mapType string, keyNames []string, valueName string) string {
	i := 0
	mapType = regexp.MustCompile(`\[\w+\]`).ReplaceAllStringFunc(mapType, func(submatch string) string {
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
