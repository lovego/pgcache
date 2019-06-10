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

	d.dataV = reflect.ValueOf(d.DataPtr)
	typ := d.dataV.Type()
	if typ.Kind() != reflect.Ptr || d.dataV.IsNil() ||
		(typ.Elem().Kind() != reflect.Map && typ.Elem().Kind() != reflect.Slice) {
		return errors.New("Data.DataPtr should be a non nil pointer to a map or slice.")
	}
	d.dataV = d.dataV.Elem()

	innerType, err := d.checkMapKeys(rowStruct)
	if err != nil {
		return err
	}
	if innerType.Kind() == reflect.Slice {
		innerType = innerType.Elem()
		d.isSortedSets = true
	}
	valueType, err := d.checkValue(rowStruct, innerType)
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
	typ := d.dataV.Type()
	if typ.Kind() == reflect.Slice {
		if len(d.MapKeys) > 0 {
			return nil, errors.New("DataPtr is a slice, so Data.MapKeys should be empty.")
		} else {
			return typ, nil
		}
	}

	layers := 0
	for ; typ.Kind() == reflect.Map; layers++ {
		if err := d.checkMapKey(layers, rowStruct, typ.Key()); err != nil {
			return nil, err
		}
		typ = typ.Elem()
	}

	if layers != len(d.MapKeys) {
		return nil, fmt.Errorf(
			"Data.DataPtr is a %d layers map, but Data.MapKeys has %d field.", layers, len(d.MapKeys),
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

func (d *Data) checkValue(rowStruct, realValueType reflect.Type) (reflect.Type, error) {
	valueType := rowStruct
	if d.Value != "" {
		if field, ok := rowStruct.FieldByName(d.Value); ok {
			valueType = field.Type
		} else {
			return nil, fmt.Errorf("Data.Value: %s, no such field in row struct.", d.Value)
		}
	}
	if !valueType.AssignableTo(realValueType) {
		if realValueType.Kind() == reflect.Ptr && valueType.AssignableTo(realValueType.Elem()) {
			d.realValueIsPointer = true
		} else {
			return nil, fmt.Errorf(
				"Data.Value: %s, type %v is not assignable to %v.", d.Value, valueType, realValueType,
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
