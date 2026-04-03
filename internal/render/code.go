package render

import (
	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
)

func beautifyJS(code string) string {
	opts := jsbeautifier.DefaultOptions()
	result, err := jsbeautifier.Beautify(&code, opts)
	if err != nil {
		return code
	}
	return result
}
