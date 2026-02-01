package masking

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/showa-93/go-mask"
)

// maskFrontBackHandler is a mask.StringFunc handler for `mask:"fbX,Y"`.
// It reveals X characters at the front and Y at the back, masking the rest.
func maskFrontBackHandler(arg, value string) (string, error) {
	front, back, err := parseFrontBackSpec(arg)
	if err != nil {
		return value, err // fallback to raw value if parsing fails
	}

	// use current global mask char from go-mask
	maskChar := '*'
	if chars := []rune(mask.MaskChar()); len(chars) > 0 {
		maskChar = chars[0]
	}

	return applyFrontBackMask(value, front, back, maskChar), nil
}

// applyFrontBackMask returns a masked string showing front and back characters.
func applyFrontBackMask(s string, front, back int, maskChar rune) string {
	runes := []rune(s)
	n := len(runes)

	if n <= front+back {
		return s // too short to mask
	}

	masked := make([]rune, n)
	copy(masked, runes[:front]) // front
	for i := front; i < n-back; i++ {
		masked[i] = maskChar
	}
	copy(masked[n-back:], runes[n-back:]) // back

	return string(masked)
}

// parseFrontBackSpec parses "X,Y" into front and back reveal lengths.
func parseFrontBackSpec(arg string) (front, back int, err error) {
	parts := strings.Split(arg, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("fb: invalid format %q, expected 'front,back'", arg)
	}

	front, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("fb: invalid front value %q: %w", parts[0], err)
	}

	back, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("fb: invalid back value %q: %w", parts[1], err)
	}

	if front < 0 || back < 0 {
		return 0, 0, fmt.Errorf("fb: front and back must be >= 0")
	}

	return front, back, nil
}
