# Code and embedded-data attribution

Gomanize implementation code is MIT licensed, Copyright (c) 2023–2026
Budhaditya (budhash@gmail.com). The WebAssembly binary also embeds learned data;
the code license does not replace the data licenses below.

- **Google Dakshina**, Google / Roark et al., 2020:
  [source](https://github.com/google-research-datasets/dakshina),
  [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
  Hindi learned components and the Bengali spelling lexicon use filtered
  training spellings. The Bengali lexicon normalizes native keys, excludes
  held-out vocabularies, selects unique high-attestation winners and serializes
  those spellings as TSV; its derived data remains CC BY-SA 4.0.
- **Google Bengali pronunciation lexicon**, Copyright Google 2015/2016:
  [source](https://github.com/googlei18n/language-resources/tree/master/bn),
  [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
  The Bengali vowel model is learned from aligned, isolated training entries.
  The embedded model is attributed to this CC BY 4.0 source.
- **Bengali native selector** combines training-derived features from both
  sources above through out-of-fold candidate generation and tree learning.
  This derived selector is CC BY-SA 4.0.

Source hashes, transformations, exclusions and reproduction tools are recorded
in the [training data record](https://github.com/budhash/gomanize/blob/main/training/data/bengali/README.md),
[Bengali lexicon record](https://github.com/budhash/gomanize/blob/main/lang/bengali/LEXICON.md),
[selector record](https://github.com/budhash/gomanize/blob/main/lang/bengali/RERANKER.md),
and [research record](https://github.com/budhash/gomanize/blob/main/docs/RESEARCH.md).
These notices identify adapted data and do not imply endorsement by its creators.

## MIT code license

MIT License

Copyright (c) 2023-2026 Budhaditya (budhash@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
