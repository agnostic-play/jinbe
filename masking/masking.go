package masking

import (
	"sync/atomic"

	"github.com/showa-93/go-mask"
)

// globalMasker holds the active masker instance. Using atomic.Pointer ensures
// concurrent reads and a NewMasker replacement are race-free without locks.
var globalMasker atomic.Pointer[mask.Masker]

func init() {
	m := mask.NewMasker()
	m.RegisterMaskStringFunc(mask.MaskTypeFilled, m.MaskFilledString)
	m.RegisterMaskStringFunc(mask.MaskTypeFixed, m.MaskFixedString)
	m.RegisterMaskStringFunc(mask.MaskTypeHash, m.MaskHashString)
	m.RegisterMaskAnyFunc(mask.MaskTypeZero, m.MaskZero)
	m.RegisterMaskIntFunc(mask.MaskTypeRandom, m.MaskRandomInt)
	m.RegisterMaskFloat64Func(mask.MaskTypeRandom, m.MaskRandomFloat64)
	m.RegisterMaskStringFunc("frontBack", maskFrontBackHandler)
	globalMasker.Store(m)
}

func getMasker() *mask.Masker {
	return globalMasker.Load()
}

// NewMasker replaces the global masker with a freshly configured instance.
// It is safe to call concurrently with any Mask* function: readers always
// observe a fully initialized masker via the atomic pointer swap.
// Call this once at startup before serving traffic.
func NewMasker(maskingChar string, setterFn ...MaskerOptionFn) {
	m := mask.NewMasker()
	m.SetMaskChar(maskingChar)
	for _, fn := range setterFn {
		fn(m)
	}

	m.RegisterMaskStringFunc(mask.MaskTypeFilled, m.MaskFilledString)
	m.RegisterMaskStringFunc(mask.MaskTypeFixed, m.MaskFixedString)
	m.RegisterMaskStringFunc(mask.MaskTypeHash, m.MaskHashString)
	m.RegisterMaskAnyFunc(mask.MaskTypeZero, m.MaskZero)
	m.RegisterMaskIntFunc(mask.MaskTypeRandom, m.MaskRandomInt)
	m.RegisterMaskFloat64Func(mask.MaskTypeRandom, m.MaskRandomFloat64)
	m.RegisterMaskStringFunc("frontBack", maskFrontBackHandler)

	globalMasker.Store(m)
}
