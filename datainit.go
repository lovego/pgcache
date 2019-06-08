package pgcache

import (
	"errors"
	"fmt"
	"reflect"
)

func (d *Data) init(rowStruct reflect.Type) error {
	if d.RWMutex == nil {
		return errors.New("Data.RWMutex is nil.")
	}

	d.mapV = reflect.ValueOf(d.MapPtr)
	typ := d.mapV.Type()
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Map || d.mapV.IsNil() {
		return errors.New("Data.Map should be a non nil pointer to a map.")
	}
	d.mapV = d.mapV.Elem()

	innerType, err := d.checkMapKeys(rowStruct)
	if err != nil {
		return err
	}
	if innerType.Kind() == reflect.Slice {
		innerType = innerType.Elem()
		d.isSortedSets = true
	}
	valueType, err := d.checkMapValue(rowStruct, innerType)
	if err != nil {
		return err
	}
	if err := d.checkSortedSetUniqueKey(valueType); err != nil {
		return err
	}
	if err := d.checkPreprocess(rowStruct); err != nil {
		return err
	}
	return d.checkPrecond(rowStruct)
}

func (d *Data) checkMapKeys(rowStruct reflect.Type) (reflect.Type, error) {
	depth := 0
	typ := d.mapV.Type()
	for ; typ.Kind() == reflect.Map; depth++ {
		if err := d.checkMapKey(depth, rowStruct, typ.Key()); err != nil {
			return nil, err
		}
		typ = typ.Elem()
	}

	if depth != len(d.MapKeys) {
		return nil, fmt.Errorf(
			"Data.Map has depth: %d, but Data.MapKeys has %d field.", depth, len(d.MapKeys),
		)
	}
	return typ, nil
}

func (d *Data) checkMapKey(i int, rowStruct, keyType reflect.Type) error {
	if i >= len(d.MapKeys) {
		return nil
	}
	name := d.MapKeys[i]
	field, ok := rowStruct.FieldByName(name)
	if !ok {
		return fmt.Errorf("Data.MapKeys[%d]: %s, no such field in row struct.", i, name)
	}
	if !field.Type.AssignableTo(keyType) {
		return fmt.Errorf(
			"Data.MapKeys[%d]: %s, type %v is not assignable to %v.", i, name, field.Type, keyType,
		)
	}
	return nil
}

func (d *Data) checkMapValue(rowStruct, realValueType reflect.Type) (reflect.Type, error) {
	valueType := rowStruct
	if d.MapValue != "" {
		if field, ok := rowStruct.FieldByName(d.MapValue); ok {
			valueType = field.Type
		} else {
			return nil, fmt.Errorf("Data.MapValue: %s, no such field in row struct.", d.MapValue)
		}
	}
	if !valueType.AssignableTo(realValueType) {
		if realValueType.Kind() == reflect.Ptr && valueType.AssignableTo(realValueType.Elem()) {
			d.realValueIsPointer = true
		} else {
			return nil, fmt.Errorf(
				"Data.MapValue: %s, type %v is not assignable to %v.", d.MapValue, valueType, realValueType,
			)
		}
	}
	return valueType, nil
}

func (d *Data) checkSortedSetUniqueKey(valueType reflect.Type) error {
	for valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}
	if !d.isSortedSets || valueType.Kind() != reflect.Struct {
		if len(d.SortedSetUniqueKey) > 0 {
			return errors.New("Data.SortedSetUniqueKey should be empty.")
		}
		return nil
	}
	if len(d.SortedSetUniqueKey) == 0 {
		return errors.New("Data.SortedSetUniqueKey should not be empty.")
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
				return fmt.Errorf(
					"Data.SortedSetUniqueKey[%d]: %s, should be a integer or string type.", i, name,
				)
			}
		} else {
			return fmt.Errorf("Data.SortedSetUniqueKey[%d]: %s, no such field in value struct.", i, name)
		}
	}
	return nil
}

func (d *Data) checkPreprocess(rowStruct reflect.Type) error {
	if d.Preprocess == "" {
		d.preprocessMethodIndex = -1
		return nil
	}
	method, ok := reflect.PtrTo(rowStruct).MethodByName(d.Preprocess)
	if !ok {
		return fmt.Errorf("Data.Preprocess: %s, no such method for the row struct.", d.Preprocess)
	}
	typ := method.Type
	if typ.NumIn() != 1 || typ.NumOut() != 0 {
		return fmt.Errorf(`Data.Preprocess: %s, should be of "func ()" form.`, d.Preprocess)
	}
	d.preprocessMethodIndex = method.Index
	return nil
}

func (d *Data) checkPrecond(rowStruct reflect.Type) error {
	if d.Precond == "" {
		d.precondMethodIndex = -1
		return nil
	}
	method, ok := reflect.PtrTo(rowStruct).MethodByName(d.Precond)
	if !ok {
		return fmt.Errorf("Data.Precond: %s, no such method for the row struct.", d.Precond)
	}
	typ := method.Type
	if typ.NumIn() != 1 || typ.NumOut() != 1 || typ.Out(0) != reflect.TypeOf(true) {
		return fmt.Errorf(`Data.Precond: %s, should be of "func () bool" form.`, d.Precond)
	}
	d.precondMethodIndex = method.Index
	return nil
}
