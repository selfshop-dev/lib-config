package config

import "testing"

func FuzzParseTagOptions(f *testing.F) {
	f.Add("-")                     // skip sentinel — must return skip=true regardless of anything else.
	f.Add("")                      // empty tag — no name, no options.
	f.Add(",squash")               // squash with no name — valid combination.
	f.Add("name,squash")           // name with squash option.
	f.Add("name,squash,omitempty") // multiple options — omitempty must be silently ignored.
	f.Add("field,omitempty")       // unknown option only — must not set squash.

	f.Fuzz(func(_ *testing.T, raw string) {
		opts, skip := parseTagOptions(raw)
		if skip {
			return
		}
		_ = opts.name
		_ = opts.squash
	})
}
