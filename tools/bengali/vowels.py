"""Conservative Google-phoneme alignment for Bengali inherent vowel labels.

Class 0 = absent, 1 = Google O (/ɔ/), 2 = Google o (/o/).
No fuzzy edit alignment, Latin spelling targets, or guessed labels.
"""
from collections import Counter
from functools import lru_cache
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bengali_training_data import native_key

CONS = dict(zip('কখগঘঙচছজঝটঠডঢণতথদধনপফবভমযরলশষসহ',
                'k kh g gh N c ch j jh T Th D Dh n t th d dh n p f b bh m j r l sh sh sh h'.split()))
CONS['ড়'] = 'r'
VOWELS = {'অ': ('O', 'o'), 'আ': ('a',), 'ই': ('i', 'i^'), 'ঈ': ('i', 'i^'),
          'উ': ('u', 'u^'), 'ঊ': ('u', 'u^'), 'এ': ('e', 'E', 'e^'), 'ও': ('o', 'o^')}
MATRAS = {'া': ('a',), 'ি': ('i',), 'ী': ('i',), 'ু': ('u',), 'ূ': ('u',),
          'ে': ('e', 'E'), 'ো': ('o',)}
FEATURES = ('cons', 'prev', 'next', 'next2', 'first', 'last', 'length', 'position')


def units(native):
    """Return (source rune index, spelling, has-inherent-slot), or None.

    Whole words with conjuncts, modifiers, silent glides, diphthongs, uncommon
    letters, or malformed vowel attachments fall back to B1 at inference too.
    """
    word = native_key(native)
    out = []
    i = 0
    while i < len(word):
        ch = word[i]
        if word[i:i+2] == 'ড়':
            ch = 'ড়'
        if ch in CONS:
            end = i + len(ch)
            inherent = end == len(word) or word[end] not in MATRAS
            out.append((i, ch, inherent))
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

    @lru_cache(None)
    def visit(unit, pos):
        if unit == len(parsed):
            return {()} if pos == len(phones) else set()
        index, ch, inherent = parsed[unit]
        if ch in CONS:
            # These alternatives change consonants, never vowel labels.
            alternatives = (CONS[ch],)
            if ch in 'শষস':
                alternatives = ('sh', 's')
            if ch == 'ফ':
                alternatives = ('f', 'p')
        else:
            alternatives = VOWELS.get(ch, MATRAS.get(ch))
        answers = set()
        for phone in alternatives:
            if pos >= len(phones) or phones[pos] != phone:
                continue
            after = pos + 1
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
