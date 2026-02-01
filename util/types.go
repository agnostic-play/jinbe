package util

import `context`

type LoggerFn func(ctx context.Context, identifier, msg string, object ...any)
