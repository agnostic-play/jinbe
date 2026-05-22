package masking

import (
	"github.com/showa-93/go-mask"
)

type MaskerOptionFn func(*mask.Masker)

func RegisterCustomMasking(maskType string, maskFunc mask.MaskAnyFunc) MaskerOptionFn {
	return func(mask *mask.Masker) {
		mask.RegisterMaskAnyFunc(maskType, maskFunc)
	}
}

func RegisterPredefinedFields(fieldMask map[string]string) MaskerOptionFn {
	return func(m *mask.Masker) {
		for fieldName, maskType := range fieldMask {
			m.RegisterMaskField(fieldName, maskType)
		}
	}
}
