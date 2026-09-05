"""Shared feature extraction — MUST stay identical to the Go inference side
(lang/hindi/schwa_model.go). Features are computed from the raw Devanagari rune
sequence so both sides agree without sharing the engine's Unit model.

For an inherent-schwa consonant at rune index i in `runes`:
  cons   : the consonant char (with nukta if present)
  prev   : rune before the consonant, or '^' at word start
  next   : rune after the consonant, or '$' at word end
  next2  : rune two after, or '$'
  first  : "1" if word-initial consonant position, else "0"
  last   : "1" if no consonant occurs after this one, else "0"
"""
NUKTA = '़'
FEATURE_NAMES = ["cons", "prev", "next", "next2", "first", "last"]

def features_at(runes, i, is_first, is_last):
    cons = runes[i]
    if i+1 < len(runes) and runes[i+1] == NUKTA:
        cons = cons + NUKTA
    prev = runes[i-1] if i > 0 else '^'
    nx = runes[i+1] if i+1 < len(runes) else '$'
    nx2 = runes[i+2] if i+2 < len(runes) else '$'
    return {"cons": cons, "prev": prev, "next": nx, "next2": nx2,
            "first": "1" if is_first else "0", "last": "1" if is_last else "0"}
