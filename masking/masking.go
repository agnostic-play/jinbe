package masking

import (
	"github.com/showa-93/go-mask"
)

var masker *mask.Masker = mask.NewMasker()

func NewMasker(maskingChar string, setterFn ...MaskerOptionFn) {
	masker = mask.NewMasker()
	masker.SetMaskChar(maskingChar)
	for _, fn := range setterFn {
		fn(masker)
	}

	masker.RegisterMaskStringFunc(mask.MaskTypeFilled, masker.MaskFilledString)
	masker.RegisterMaskStringFunc(mask.MaskTypeFixed, masker.MaskFixedString)
	masker.RegisterMaskStringFunc(mask.MaskTypeHash, masker.MaskHashString)
	masker.RegisterMaskAnyFunc(mask.MaskTypeZero, masker.MaskZero)
	masker.RegisterMaskIntFunc(mask.MaskTypeRandom, masker.MaskRandomInt)
	masker.RegisterMaskFloat64Func(mask.MaskTypeRandom, masker.MaskRandomFloat64)
	masker.RegisterMaskStringFunc("frontBack", maskFrontBackHandler)
}
