package masking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/agnostic-play/jinbe/logger"
)

func MaskStruct(target any) (any, error) {
	if target == nil {
		return nil, nil
	}
	maskedStruct, err := masker.Mask(target)
	if err != nil {
		return target, err
	}

	return maskedStruct, nil
}

func MaskJSON(target []byte) (any, error) {
	var targetJson any
	if err := json.Unmarshal(target, &targetJson); err != nil {
		return target, err
	}

	return masker.Mask(targetJson)
}

func ShouldMaskStruct(target any) any {
	if target == nil {
		return nil
	}
	maskedStruct, err := masker.Mask(target)
	if err != nil {
		return target
	}

	return maskedStruct
}

func ShouldMaskJSON(target []byte) any {
	var targetJson any
	if err := json.Unmarshal(target, &targetJson); err != nil {
		return target
	}

	maskedStruct, err := masker.Mask(targetJson)
	if err != nil {
		return target
	}

	return maskedStruct
}

func ShouldMaskStructWithLogger(ctx context.Context, log logger.PublicLoggerWithoutParamsFn, target any) any {
	if target == nil {
		return nil
	}
	maskedStruct, err := masker.Mask(target)
	if err != nil {
		log(ctx, fmt.Sprintf("failed to mask: %s", err.Error()))
		return target
	}

	return maskedStruct
}

func ShouldMaskJSONWithLogger(ctx context.Context, log logger.PublicLoggerWithoutParamsFn, target []byte) any {
	var targetJson any
	if err := json.Unmarshal(target, &targetJson); err != nil {
		log(ctx, fmt.Sprintf("failed to mask: %s", err.Error()))
		return target
	}

	maskedStruct, err := masker.Mask(targetJson)
	if err != nil {
		log(ctx, fmt.Sprintf("failed to mask: %s", err.Error()))
		return targetJson // return original parsed value, not nil
	}

	return maskedStruct
}
