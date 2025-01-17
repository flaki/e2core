package api

import (
	"github.com/suborbital/e2core/sat/engine2/runtime/instance"
)

func (d *defaultAPI) GetFFIResultHandler() HostFn {
	fn := func(args ...interface{}) (interface{}, error) {
		pointer := args[0].(int32)
		ident := args[1].(int32)

		ret := d.getFfiResult(pointer, ident)

		return ret, nil
	}

	return NewHostFn("get_ffi_result", 2, true, fn)
}

func (d *defaultAPI) getFfiResult(pointer int32, identifier int32) int32 {
	ll := d.logger.With().Str("method", "getFfiResult").Logger()

	inst, err := instance.ForIdentifier(identifier, false)
	if err != nil {
		ll.Err(err).Msg("instance.ForIdentifier")
		return -1
	}

	result, err := inst.Ctx().UseFFIResult()
	if err != nil {
		ll.Err(err).Msg("inst.Ctx().UseFFIResult")
		return -1
	}

	data := result.Result
	if result.Err != nil {
		data = []byte(result.Err.Error())
	}

	inst.WriteMemoryAtLocation(pointer, data)

	return 0
}
