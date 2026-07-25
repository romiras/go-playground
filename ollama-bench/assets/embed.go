package assets

import _ "embed"

// PrefillContext is the embedded high-entropy prose corpus for Test Case A.
//
//go:embed prefill_context.txt
var PrefillContext string
