"""Conservative Google-phoneme alignment for Bengali inherent vowel labels.

Class 0 = absent, 1 = Google O (/ɔ/), 2 = Google o (/o/).
No fuzzy edit alignment, Latin spelling targets, or guessed labels.
Eligibility comes from lang/bengali/vowel_inventory.json, shared with the Go runtime.
"""
from collections import Counter
from functools import lru_cache
import json
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import ROOT, native_key

INVENTORY_PATH = ROOT/'lang/bengali/vowel_inventory.json'
INVENTORY = json.loads(INVENTORY_PATH.read_text())
SCOPE, LEGACY_SCOPE = INVENTORY['scope'], INVENTORY['legacy_scope']
HALANT, MAX_CLUSTER = INVENTORY['halant'], INVENTORY['max_cluster']
CONS = dict(zip('কখগঘঙচছজঝটঠডঢণতথদধনপফবভমযরলশষসহ',
                'k kh g gh N c ch j jh T Th D Dh n t th d dh n p f b bh m j r l sh sh sh h'.split()))
CONS['ড়'] = 'r'
VOWELS = {'অ': ('O', 'o'), 'আ': ('a',), 'ই': ('i', 'i^'), 'ঈ': ('i', 'i^'),
          'উ': ('u', 'u^'), 'ঊ': ('u', 'u^'), 'এ': ('e', 'E', 'e^'), 'ও': ('o', 'o^')}
MATRAS = {'া': ('a',), 'ি': ('i',), 'ী': ('i',), 'ু': ('u',), 'ূ': ('u',),
          'ে': ('e', 'E'), 'ো': ('o',)}
if (list(CONS) != INVENTORY['consonants'] or list(MATRAS) != INVENTORY['matras'] or
        list(VOWELS) != INVENTORY['independent_vowels']):
    raise ValueError('aligner phones do not match the shared vowel inventory')
FEATURES = ('cons', 'prev', 'next', 'next2', 'first', 'last', 'length', 'position')
# Consonant-only phone alternatives (never vowel labels).
OWN = {'শ': ('sh', 's'), 'ষ': ('sh', 's'), 'স': ('sh', 's'), 'ফ': ('f', 'p')}
# Extra phones for a non-first cluster member, keyed by (previous member, member).
# ঞ is not in the shared inventory, so জ্ঞ words stay ineligible and the জ/ঞ
# entry never fires (recorded as a deviation in the 2026-10-21 record).
CLUSTER_EXTRA = {('ক', 'ষ'): ('kh',), ('জ', 'ঞ'): ('g',)}
# Added from Google train alignment failures only (2026-10-21 record): the first
# member of these clusters may be silent (ক্ষ read as a single kh).
SILENT_FIRST = {('ক', 'ষ')}


def units(native, max_cluster=MAX_CLUSTER):
    """Return (source rune index, spelling, has-inherent-slot), or None.

    Units are consonants, halants, matras and independent vowels from the shared
    inventory. A consonant before a halant has no inherent slot; clusters have at
    most max_cluster members (1 = the legacy simple-word scope). Modifiers, silent
    glides, diphthongs, uncommon letters, malformed attachments and longer or
    unterminated clusters fall back to B1 at inference too.
    """
    word = native_key(native)
    out = []
    i = 0
    members = 0  # consonants in the current halant-joined run
    while i < len(word):
        ch = word[i]
        if word[i:i+2] == 'ড়':
            ch = 'ড়'
        if out and out[-1][1] == HALANT and ch not in CONS:
            return None
        if ch in CONS:
            end = i + len(ch)
            members = members + 1 if out and out[-1][1] == HALANT else 1
            if members > max_cluster:
                return None
            inherent = end == len(word) or word[end] not in MATRAS and word[end] != HALANT
            out.append((i, ch, inherent))
        elif ch == HALANT:
            if max_cluster < 2 or not out or out[-1][1] not in CONS or i+1 == len(word):
                return None
            out.append((i, ch, False))
        elif ch in VOWELS:
            out.append((i, ch, False))
        elif ch in MATRAS:
            if not out or out[-1][1] not in CONS or out[-1][2]:
                return None
            out.append((i, ch, False))
        else:
            return None
        i += len(ch)
    if any(slot and n+1 < len(out) and out[n+1][1] in VOWELS
           for n, (_, _, slot) in enumerate(out)):
        # The renderer owns pre-independent-vowel behavior; keep B1 there.
        return None
    return out or None


def eligible(native, max_cluster=MAX_CLUSTER):
    """A word the model decides: supported units with at least one inherent slot."""
    parsed = units(native, max_cluster)
    return parsed is not None and any(slot for _, _, slot in parsed)


def features(native, index):
    word = native_key(native)
    parsed = units(word)
    consonants = [i for i, ch, _ in parsed if ch in CONS]
    ch = next(ch for i, ch, _ in parsed if i == index)
    return dict(zip(FEATURES, (ch, word[index-1] if index else '^',
                              word[index+1] if index+1 < len(word) else '$',
                              word[index+2] if index+2 < len(word) else '$',
                              str(int(index == consonants[0])), str(int(index == consonants[-1])),
                              str(min(len(word), 12)), str(min(index, 6)))))


def align(native, transcription):
    parsed = units(native)
    if parsed is None:
        return 'unsupported', None
    phones = tuple(t for t in transcription.split() if t != '.')

    def consonant_phones(unit):
        ch = parsed[unit][1]
        own = [(phone,) for phone in OWN.get(ch, (CONS[ch],))]
        if unit < 2 or parsed[unit-1][1] != HALANT:
            if unit+2 < len(parsed) and parsed[unit+1][1] == HALANT and (ch, parsed[unit+2][1]) in SILENT_FIRST:
                return own + [()]
            return own
        previous = parsed[unit-2][1]
        # Non-first cluster member: own phone, gemination of the previous member,
        # silence, or a listed cluster reading. Vowel handling is unchanged.
        extra = [(phone,) for phone in OWN.get(previous, (CONS[previous],))]
        extra += [(phone,) for phone in CLUSTER_EXTRA.get((previous, ch), ())]
        return list(dict.fromkeys(own + extra + [()]))

    @lru_cache(None)
    def visit(unit, pos):
        if unit == len(parsed):
            return {()} if pos == len(phones) else set()
        index, ch, inherent = parsed[unit]
        if ch == HALANT:
            return visit(unit+1, pos)
        if ch in CONS:
            alternatives = consonant_phones(unit)
        else:
            alternatives = [(phone,) for phone in VOWELS.get(ch, MATRAS.get(ch))]
        answers = set()
        for match in alternatives:
            if phones[pos:pos+len(match)] != match:
                continue
            after = pos + len(match)
            choices = [(after, None)] if not inherent else [(after, 0)]
            if inherent and after < len(phones) and phones[after] in ('O', 'o'):
                choices.append((after+1, 1 if phones[after] == 'O' else 2))
            for end, label in choices:
                for tail in visit(unit+1, end):
                    answers.add(((index, label),) + tail if label is not None else tail)
                    if len(answers) > 1:
                        # Two distinct label paths suffice to reject this alignment.
                        return answers
        return answers

    answers = visit(0, 0)
    if not answers:
        return 'unmatched', None
    if len(answers) != 1:
        return 'ambiguous', None
    labels = next(iter(answers))
    return ('aligned' if labels else 'no_slots'), labels


def dataset(rows):
    """Reject entire types with any unaligned/conflicting pronunciation variant."""
    stats = Counter({name: 0 for name in (
        "aligned_rows", "unsupported_rows", "unmatched_rows", "ambiguous_rows",
        "no_slots_rows", "rejected_types", "conflicting_variant_types")})
    by_word = {}
    for row in rows:
        word, _, _, pronunciation, *_ = row
        status, labels = align(word, pronunciation)
        stats[status + '_rows'] += 1
        by_word.setdefault(word, []).append((status, labels))
    examples = []
    accepted = set()
    for word, variants in sorted(by_word.items()):
        if any(status != 'aligned' for status, _ in variants):
            stats['rejected_types'] += 1
            continue
        labels = {lab for _, lab in variants}
        if len(labels) != 1:
            stats['conflicting_variant_types'] += 1
            continue
        accepted.add(word)
        # Each word/slot appears once, not once per pronunciation or POS tag.
        for index, label in next(iter(labels)):
            examples.append((features(word, index), label, word, index))
            stats['class_' + str(label)] += 1
    stats.update({'source_rows': len(rows), 'source_types': len(by_word),
                  'accepted_types': len(accepted), 'instances': len(examples)})
    return examples, dict(sorted(stats.items()))
