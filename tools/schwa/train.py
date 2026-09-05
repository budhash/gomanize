#!/usr/bin/env python3
"""Train a schwa-deletion decision tree from Dakshina, eval on held-out test.

Pure stdlib. Emits lang/hindi/schwa_tree.json (loaded by the Go engine).
Usage: python3 tools/schwa/train.py [data.csv] [--depth N] [--min-leaf N]
"""
import csv, json, sys, os
sys.path.insert(0, os.path.dirname(__file__))
from features import features_at, FEATURE_NAMES

CONS = {
 'क':'k','ख':'kh','ग':'g','घ':'gh','ङ':'ng','च':'ch','छ':'chh','ज':'j','झ':'jh','ञ':'ny',
 'ट':'t','ठ':'th','ड':'d','ढ':'dh','ण':'n','त':'t','थ':'th','द':'d','ध':'dh','न':'n',
 'प':'p','फ':'ph','ब':'b','भ':'bh','म':'m','य':'y','र':'r','ल':'l','व':'v','श':'sh',
 'ष':'sh','स':'s','ह':'h','ळ':'l','क़':'q','ख़':'kh','ग़':'g','ज़':'z','ड़':'r','ढ़':'rh','फ़':'f',
}
MATRA = set('ािीुूृेैोौंःँॉ'); HALANT='्'; NUKTA='़'

def instances(native, roman):
    """Yield (features, deleted) for each inherent-schwa consonant, or None on misalign."""
    runes = list(native)
    # locate consonant rune indices + inherent flag
    cons_idx = []  # (rune_index, char, inherent)
    i = 0
    while i < len(runes):
        c = runes[i]; ci = i
        if i+1 < len(runes) and runes[i+1] == NUKTA:
            c = c + NUKTA
        if c in CONS:
            nxt = runes[ci+len(c)] if ci+len(c) < len(runes) else ''
            inherent = not (nxt == HALANT or nxt in MATRA)
            cons_idx.append((ci, c, inherent))
        i += len(c) if c in CONS else 1
    # align skeleton to roman
    out = []; pos = 0
    ncons = len(cons_idx)
    seen_cons = 0
    for k,(ri,c,inherent) in enumerate(cons_idx):
        cr = CONS[c]
        j = roman.find(cr, pos)
        if j < 0: return None
        after = j + len(cr)
        if inherent:
            deleted = not (after < len(roman) and roman[after] == 'a')
            is_first = (k == 0)
            is_last = all(not cc[2] and True for cc in [])  # placeholder
            # last = no inherent consonant after this one is a consonant at all
            is_last = (k == ncons-1)
            feats = features_at(runes, ri, is_first, is_last)
            out.append((feats, 1 if deleted else 0))
        pos = after
    return out

def load(path, split, min_att=2):
    data=[]
    with open(path, encoding='utf-8') as f:
        rd=csv.reader(f); next(rd,None)
        for row in rd:
            if len(row)<3 or f'split={split}' not in row[2]: continue
            att=next((int(t.split('=')[1]) for t in row[2].split() if t.startswith('attestations=')),1)
            if att<min_att: continue
            ins=instances(row[0],row[1])
            if ins: data.extend(ins)
    return data

# --- tiny CART (Gini), categorical equality splits ---
def gini(rows):
    n=len(rows)
    if n==0: return 0
    p=sum(r[1] for r in rows)/n
    return 1-p*p-(1-p)*(1-p)

def build(rows, depth, max_depth, min_leaf):
    n=len(rows); pos=sum(r[1] for r in rows)
    leaf={"leaf": 1 if pos*2>=n else 0, "p": round(pos/max(n,1),4), "n": n}
    if depth>=max_depth or n<2*min_leaf or pos==0 or pos==n:
        return leaf
    base=gini(rows); best=None
    for fi,fname in enumerate(FEATURE_NAMES):
        vals={}
        for r in rows: vals.setdefault(r[0][fname],0)
        for v in vals:
            L=[r for r in rows if r[0][fname]==v]; R=[r for r in rows if r[0][fname]!=v]
            if len(L)<min_leaf or len(R)<min_leaf: continue
            g=(len(L)*gini(L)+len(R)*gini(R))/n
            gain=base-g
            if best is None or gain>best[0]:
                best=(gain,fname,v,L,R)
    if best is None or best[0]<=1e-6: return leaf
    _,fname,v,L,R=best
    return {"f":fname,"v":v,"yes":build(L,depth+1,max_depth,min_leaf),"no":build(R,depth+1,max_depth,min_leaf)}

def predict(tree, feats):
    while "leaf" not in tree:
        tree = tree["yes"] if feats[tree["f"]]==tree["v"] else tree["no"]
    return tree["leaf"]

def acc(tree, data):
    if not data: return 0
    return sum(1 for f,y in data if predict(tree,f)==y)/len(data)

def main():
    path=next((a for a in sys.argv[1:] if a.endswith('.csv')), 'benchmark/data/dakshina_hi.csv')
    depth=int(_opt('--depth',12)); min_leaf=int(_opt('--min-leaf',15))
    tr=load(path,'train'); te=load(path,'test'); dv=load(path,'dev')
    print(f"train={len(tr)} dev={len(dv)} test={len(te)} instances")
    tree=build(tr,0,depth,min_leaf)
    # baselines
    del_rate=sum(y for _,y in tr)/len(tr)
    maj=1 if del_rate>=0.5 else 0
    base_acc=sum(1 for _,y in te if y==maj)/len(te)
    print(f"train deletion rate: {del_rate*100:.1f}%")
    print(f"TEST per-schwa accuracy:")
    print(f"  majority baseline (always {'delete' if maj else 'keep'}): {base_acc*100:.2f}%")
    print(f"  decision tree (depth={depth}, min_leaf={min_leaf}):        {acc(tree,te)*100:.2f}%")
    print(f"  (dev: {acc(tree,dv)*100:.2f}%, train: {acc(tree,tr)*100:.2f}%)")
    out='lang/hindi/schwa_tree.json'
    with open(out,'w',encoding='utf-8') as f:
        json.dump({"features":FEATURE_NAMES,"tree":tree}, f, ensure_ascii=False)
    print(f"exported {out} ({os.path.getsize(out)} bytes)")

def _opt(name,default):
    a=sys.argv
    return a[a.index(name)+1] if name in a else default

if __name__=='__main__': main()
