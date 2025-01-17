package api

import (
	"github.com/suborbital/e2core/sat/engine2/runtime/instance"
	"github.com/suborbital/systemspec/capabilities"
)

func (d *defaultAPI) RequestGetFieldHandler() HostFn {
	fn := func(args ...interface{}) (interface{}, error) {
		fieldType := args[0].(int32)
		keyPointer := args[1].(int32)
		keySize := args[2].(int32)
		ident := args[3].(int32)

		ret := d.requestGetField(fieldType, keyPointer, keySize, ident)

		return ret, nil
	}

	return NewHostFn("request_get_field", 4, true, fn)
}

func (d *defaultAPI) requestGetField(fieldType int32, keyPointer int32, keySize int32, identifier int32) int32 {
	ll := d.logger.With().Str("method", "requestGetField").Logger()

	inst, err := instance.ForIdentifier(identifier, true)
	if err != nil {
		ll.Err(err).Msg("instance.ForIdentifier")
		return -1
	}

	keyBytes := inst.ReadMemory(keyPointer, keySize)
	key := string(keyBytes)

	req := RequestFromContext(inst.Ctx().Context)

	if req == nil {
		ll.Error().Msg("request is not set")
	}

	handler := capabilities.NewRequestHandler(*d.capabilities.RequestConfig, req)

	// err gets used in SetFFIResult below rather than returned
	val, err := handler.GetField(fieldType, key)
	if err != nil {
		if err == capabilities.ErrKeyNotFound {
			// treat this as an empty value rather than an actual error
			val = []byte{}
			err = nil
		} else {
			ll.Err(err).Msg("handler.GetField")
			return -1
		}
	}

	result, err := inst.Ctx().SetFFIResult(val, err)
	if err != nil {
		ll.Err(err).Msg("inst.Ctx().SetFFIResult")
		return -1
	}

	return result.FFISize()
}

func (d *defaultAPI) RequestSetFieldHandler() HostFn {
	fn := func(args ...interface{}) (interface{}, error) {
		fieldType := args[0].(int32)
		keyPointer := args[1].(int32)
		keySize := args[2].(int32)
		valPointer := args[3].(int32)
		valSize := args[4].(int32)
		ident := args[5].(int32)

		ret := d.requestSetField(fieldType, keyPointer, keySize, valPointer, valSize, ident)

		return ret, nil
	}

	return NewHostFn("request_set_field", 6, true, fn)
}

func (d *defaultAPI) requestSetField(fieldType int32, keyPointer int32, keySize int32, valPointer int32, valSize int32, identifier int32) int32 {
	ll := d.logger.With().Str("method", "requestSetField").Logger()

	inst, err := instance.ForIdentifier(identifier, true)
	if err != nil {
		ll.Err(err).Msg("instance.ForIdentifier")
		return -1
	}

	keyBytes := inst.ReadMemory(keyPointer, keySize)
	key := string(keyBytes)

	valBytes := inst.ReadMemory(valPointer, valSize)
	val := string(valBytes)

	req := RequestFromContext(inst.Ctx().Context)

	if req == nil {
		ll.Error().Msg("request is not set")
	}

	handler := capabilities.NewRequestHandler(*d.capabilities.RequestConfig, req)

	if err := handler.SetField(fieldType, key, val); err != nil {
		ll.Err(err).Msg("handler.SetField")
	}

	result, err := inst.Ctx().SetFFIResult(nil, err)
	if err != nil {
		ll.Err(err).Msg("inst.Ctx().SetFFIResult")
		return -1
	}

	return result.FFISize()
}
