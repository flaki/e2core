package instance

import (
	"crypto/rand"
	"math"
	"math/big"
	"sync"

	"github.com/pkg/errors"
)

// the instance mapper is a global var that maps a random int32 to a wasm instance to make bi-directional FFI calls "easy"
var instanceMapper = sync.Map{}

func ForIdentifier(ident int32, needsFFIResult bool) (*Instance, error) {
	rawRef, exists := instanceMapper.Load(ident)
	if !exists {
		return nil, errors.New("instance does not exist")
	}

	inst := rawRef.(*Instance)

	if needsFFIResult && inst.Ctx().HasFFIResult() {
		return nil, errors.New("cannot use instance for host call with existing call in progress")
	}

	return inst, nil
}

func Store(inst *Instance) (int32, error) {
	for {
		ident, err := randomIdentifier()
		if err != nil {
			return -1, errors.Wrap(err, "failed to randomIdentifier")
		}

		// ensure we don't accidentally overwrite something else
		// (however unlikely that may be)
		if _, exists := instanceMapper.Load(ident); exists {
			continue
		}

		instanceMapper.Store(ident, inst)

		return ident, nil
	}
}

func Remove(ident int32) {
	instanceMapper.Delete(ident)
}

func randomIdentifier() (int32, error) {
	// generate a random number between 0 and the largest possible int32
	num, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		return -1, errors.Wrap(err, "failed to rand.Int")
	}

	return int32(num.Int64()), nil
}
