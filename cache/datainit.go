package cache

import (
	"log"
	"reflect"
)

func (d *Data) init(rowStruct reflect.Type) {
	if d.RWMutex == nil {
		log.Panic("Data.RWMutex is nil.")
	}

	d.mapV = reflect.ValueOf(d.MapPtr)
	typ := d.mapV.Type()
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Map || d.mapV.IsNil() {
		log.Panic("Data.Map should be a non nil pointer to a map.")
	}
	d.mapV = d.mapV.Elem()

	innerType := d.checkMapKeys(rowStruct)
	if innerType.Kind() == reflect.Slice {
		innerType = innerType.Elem()
		d.isSortedSets = true
	}
	valueType := d.checkMapValue(rowStruct, innerType)
	d.checkSortedSetUniqueKey(valueType)
	d.checkPreprocess(rowStruct)
	d.checkPrecond(rowStruct)
}

func (d *Data) checkMapKeys(rowStruct reflect.Type) reflect.Type {
	depth := 0
	typ := d.mapV.Type()
	for ; typ.Kind() == reflect.Map; depth++ {
		d.checkMapKey(depth, rowStruct, typ.Key())
		typ = typ.Elem()
	}

	if depth != len(d.MapKeys) {
		log.Panicf("Data.Map has depth: %d, but Data.MapKeys has %d field.", depth, len(d.MapKeys))
	}
	return typ
}

func (d *Data) checkMapKey(i int, rowStruct, keyType reflect.Type) {
	if i >= len(d.MapKeys) {
		return
	}
	name := d.MapKeys[i]
	field, ok := rowStruct.FieldByName(name)
	if !ok {
		log.Panicf("Data.MapKeys[%d]: %s, no such field in row struct.", i, name)
	}
	if !field.Type.AssignableTo(keyType) {
		log.Panicf(
			"Data.MapKeys[%d]: %s, type %v is not assignable to %v.", i, name, field.Type, keyType,
		)
	}
}

func (d *Data) checkMapValue(rowStruct, realValueType reflect.Type) reflect.Type {
	valueType := rowStruct
	if d.MapValue != "" {
		if field, ok := rowStruct.FieldByName(d.MapValue); ok {
			valueType = field.Type
		} else {
			log.Panicf("Data.MapValue: %s, no such field in row struct.", d.MapValue)
		}
	}
	if !valueType.AssignableTo(realValueType) {
		if realValueType.Kind() == reflect.Ptr && valueType.AssignableTo(realValueType.Elem()) {
			d.realValueIsPointer = true
		} else {
			log.Panicf(
				"Data.MapValue: %s, type %v is not assignable to %v.", d.MapValue, valueType, realValueType,
			)
		}
	}
	return valueType
}

func (d *Data) checkSortedSetUniqueKey(valueType reflect.Type) {
	for valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}
	if !d.isSortedSets || valueType.Kind() != reflect.Struct {
		if len(d.SortedSetUniqueKey) > 0 {
			log.Panic("Data.SortedSetUniqueKey should be empty.")
		}
		return
	}
	if len(d.SortedSetUniqueKey) == 0 {
		log.Panic("Data.SortedSetUniqueKey should not be empty.")
	}
	for i, name := range d.SortedSetUniqueKey {
		if field, ok := valueType.FieldByName(name); ok {
			typ := field.Type
			for typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			switch typ.Kind() {
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
				reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8,
				reflect.String:
			default:
				log.Panicf("Data.SortedSetUniqueKey[%d]: %s, should be a integer or string type.", i, name)
			}
		} else {
			log.Panicf("Data.SortedSetUniqueKey[%d]: %s, no such field in value struct.", i, name)
		}
	}
}

func (d *Data) checkPreprocess(rowStruct reflect.Type) {
	if d.Preprocess == "" {
		d.preprocessMethodIndex = -1
		return
	}
	method, ok := reflect.PtrTo(rowStruct).MethodByName(d.Preprocess)
	if !ok {
		log.Panicf("Data.Preprocess: %s, no such method for the row struct.", d.Preprocess)
	}
	typ := method.Type
	if typ.NumIn() != 1 || typ.NumOut() != 0 {
		log.Panicf(`Data.Preprocess: %s, should be of "func ()" form.`, d.Preprocess)
	}
	d.preprocessMethodIndex = method.Index
}

func (d *Data) checkPrecond(rowStruct reflect.Type) {
	if d.Precond == "" {
		d.precondMethodIndex = -1
		return
	}
	method, ok := reflect.PtrTo(rowStruct).MethodByName(d.Precond)
	if !ok {
		log.Panicf("Data.Precond: %s, no such method for the row struct.", d.Precond)
	}
	typ := method.Type
	if typ.NumIn() != 1 || typ.NumOut() != 1 || typ.Out(0) != reflect.TypeOf(true) {
		log.Panicf(`Data.Precond: %s, should be of "func () bool" form.`, d.Precond)
	}
	d.precondMethodIndex = method.Index
}
