package handler

import (
	"reflect"
	"sync"

	"github.com/lovego/sorted_sets"
)

type Data struct {
	*sync.RWMutex
	// a pointer to a map to store data
	MapPtr interface{}
	// field name to get map keys from row struct
	MapKeys []string
	// field name to get map value from row struct, leave this empty to use whole row.
	MapValue string

	// if the map value is a slice, it's used as sorted set.
	// if it's a sorted set of struct, use the {SortedSetStructUniqueKey} fields as unique key.
	SortedSetUniqueKey []string

	// the map value to store data
	mapV reflect.Value
	// map value is a sorted set
	isSortedSets bool
	// real map value is a pointer of the row struct or row struct's {MapValue} field.
	realValueIsPointer bool
}

func (d *Data) save(row reflect.Value) {
	d.Lock()
	defer d.Unlock()

	mapV := d.mapV
	for i := 0; i < len(d.MapKeys)-1; i++ {
		key := row.FieldByName(d.MapKeys[i])
		value := mapV.MapIndex(key)
		if !value.IsValid() || value.IsNil() {
			value = reflect.MakeMap(mapV.Type().Elem())
			mapV.SetMapIndex(key, value)
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
		value = sorted_sets.Save(mapV.MapIndex(key), value, d.SortedSetUniqueKey...)
	}
	mapV.SetMapIndex(key, value)
}

func (d *Data) remove(row reflect.Value) {
	d.Lock()
	defer d.Unlock()

	mapV := d.mapV
	for i := 0; i < len(d.MapKeys)-1; i++ {
		key := row.FieldByName(d.MapKeys[i])
		mapV := mapV.MapIndex(key)
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
		slice = sorted_sets.Remove(slice, value, d.SortedSetUniqueKey...)
		if !slice.IsValid() || slice.Len() == 0 {
			mapV.SetMapIndex(key, reflect.Value{})
		} else {
			mapV.SetMapIndex(key, value)
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
