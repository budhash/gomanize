#!/usr/bin/env python3
"""Force-align Dakshina (native, roman) pairs to per-schwa delete/keep labels.

For each Devanagari consonant carrying an inherent schwa (i.e. NOT followed by a
dependent vowel matra, halant, or another combining sign), decide from the roman
whether that schwa was realized ('a' present) or deleted. Pure stdlib.
"""
import csv, sys

# Minimal Devanagari inventory (consonant -> base roman), matras, signs.
CONS = {
 'क':'k','ख':'kh','ग':'g','घ':'gh','ङ':'ng','च':'ch','छ':'chh','ज':'j','झ':'jh','ञ':'ny',
 'ट':'t','ठ':'th','ड':'d','ढ':'dh','ण':'n','त':'t','थ':'th','द':'d','ध':'dh','न':'n',
 'प':'p','फ':'ph','ब':'b','भ':'bh','म':'m','य':'y','र':'r','ल':'l','व':'v','श':'sh',
 'ष':'sh','स':'s','ह':'h','ळ':'l','क़':'q','ख़':'kh','ग़':'g','ज़':'z','ड़':'r','ढ़':'rh','फ़':'f',
}
MATRA = set('ािीुूृेैोौंःँॉ')   # dependent vowel signs + nasal/visarga (schwa-replacing or modifying)
HALANT = '्'
NUKTA = '़'

def units(native):
    """Yield consonants with a flag: has_inherent_schwa (no following matra/halant)."""
    chars = list(native)
    i = 0
    out = []
    while i < len(chars):
        c = chars[i]
        # combine nukta
        if i+1 < len(chars) and chars[i+1] == NUKTA:
            c = c + NUKTA
            i += 1
        if c in CONS:
            # look ahead for matra / halant
            nxt = chars[i+1] if i+1 < len(chars) else ''
            if nxt == HALANT:
                out.append((c, False, 'halant'))  # conjunct, schwa suppressed
            elif nxt in MATRA:
                out.append((c, False, 'matra'))    # vowel replaces schwa
            else:
                out.append((c, True, 'inherent'))  # inherent schwa candidate
        i += 1
    return out

def label(native, roman):
    """Return list of (cons, deleted_bool) for inherent-schwa consonants, or None if align fails."""
    us = units(native)
    # Build expected consonant skeleton (romans) in order.
    skel = [CONS[c] for c,_,_ in us]
    # Greedy left-to-right match of skeleton against roman; between consecutive
    # consonants, an 'a' (and only a) indicates a realized schwa.
    labels = []
    pos = 0
    r = roman
    for idx,(c,inherent,_) in enumerate(us):
        cr = CONS[c]
        # find this consonant's roman starting at pos
        j = r.find(cr, pos)
        if j < 0:
            return None  # skeleton mismatch -> give up (noisy pair)
        after = j + len(cr)
        if inherent:
            # realized if an 'a' immediately follows (before next consonant/vowel)
            deleted = not (after < len(r) and r[after] == 'a')
            labels.append((c, deleted))
        pos = after
    return labels

def main():
    path = sys.argv[1] if len(sys.argv)>1 else 'benchmark/data/dakshina_hi.csv'
    split = sys.argv[2] if len(sys.argv)>2 else 'train'
    tot_pairs=aligned=0
    inst=0; deleted=0
    per_cons={}
    with open(path, encoding='utf-8') as f:
        rd=csv.reader(f); next(rd,None)
        for row in rd:
            if len(row)<3 or f'split={split}' not in row[2]: continue
            # use only high-attestation pairs to reduce noise
            att = next((int(t.split('=')[1]) for t in row[2].split() if t.startswith('attestations=')),1)
            if att < 2: continue
            tot_pairs+=1
            lab = label(row[0], row[1])
            if lab is None: continue
            aligned+=1
            for c,d in lab:
                inst+=1; deleted+=d
                pc=per_cons.setdefault(c,[0,0]); pc[1]+=1; pc[0]+=d
    print(f"split={split} pairs(att>=2)={tot_pairs} aligned={aligned} ({100*aligned/max(tot_pairs,1):.1f}%)")
    print(f"schwa instances={inst} deleted={deleted} ({100*deleted/max(inst,1):.1f}% deletion rate)")
    print("sample per-consonant deletion rates (deleted/total):")
    for c in ['न','त','र','म','क','ल','स']:
        if c in per_cons:
            d,t=per_cons[c]; print(f"  {c}: {d}/{t} = {100*d/t:.0f}%")

if __name__=='__main__':
    main()
