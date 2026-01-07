# Datasets

This directory contains downloaded transliteration datasets. These files are **gitignored** due to their size.

## Available Datasets

| Dataset | Size | Languages | Total Pairs | License |
|---------|------|-----------|-------------|---------|
| [Dakshina](https://github.com/google-research-datasets/dakshina) | ~2GB | 12 | 250K+ | CC BY-SA 4.0 |
| [Aksharantar](https://huggingface.co/datasets/ai4bharat/Aksharantar) | ~6.5GB | 20 | 26M | CC0/CC-BY |

## Test Data vs Raw Downloads

| Location | Purpose | Checked into Git |
|----------|---------|------------------|
| `testbed/dakshina/` | Curated test files for accuracy testing | ✓ Yes |
| `datasets/` | Raw downloaded datasets | ✗ No (gitignored) |

The `testbed/` files are **curated subsets** with:
- High-confidence entries only (4+ human attestations)
- Native words separated from English loanwords
- One romanization per word (highest voted)

See `datasets/dakshina/README.md` for the full curation process.

## Directory Structure

```
datasets/
├── README.md                # This file
├── dakshina.tar             # Dakshina original archive
├── dakshina/                # Extracted Dakshina CSVs
│   ├── hi_train.csv
│   ├── hi_dev.csv
│   └── hi_test.csv
└── aksharantar/             # Aksharantar dataset
    ├── README.md            # Aksharantar documentation
    ├── {lang}.zip           # Original archives (20 languages)
    ├── {lang}_train.json    # Training data
    ├── {lang}_valid.json    # Validation data
    ├── {lang}_test.json     # Test data
    ├── {lang}_train.csv     # CSV format (converted)
    ├── {lang}_valid.csv
    └── {lang}_test.csv
```

## Setup

### Download Dakshina

```bash
cd datasets/dakshina

# Download (~2GB)
curl -L https://storage.googleapis.com/gresearch/dakshina/dakshina_dataset_v1.0.tar \
  -o dakshina.tar

# Extract Hindi (4+ attestations by default)
./generate_dataset.sh hi
# Output: dakshina_hi.csv

# Extract with different threshold
./generate_dataset.sh hi 1      # All attestations
```

### Download Aksharantar

```bash
cd datasets/aksharantar

# Download a language
curl -sL "https://huggingface.co/datasets/ai4bharat/Aksharantar/resolve/main/hin.zip" -o "hin.zip"

# Extract Hindi (human-verified only by default)
./generate_dataset.sh hi
# Output: aksharantar_hi.csv

# Extract with different threshold
./generate_dataset.sh hi -0.15  # Good quality balance
./generate_dataset.sh hi -999   # All entries
```

## CSV Format

Both datasets use the same output format:

```csv
native,roman,notes
अंकुर,ankur,attestations=4 split=train
जन्मदिवस,janamdivas,source=Dakshina split=train
```

The `notes` column contains space-separated key=value metadata:
- **Dakshina**: `attestations=N split=<train|dev|test>`
- **Aksharantar**: `source=<source> split=<train|valid|test>`

## Language Codes

### Dakshina (12 languages)
| Code | Language | Script |
|------|----------|--------|
| bn | Bengali | Bengali |
| gu | Gujarati | Gujarati |
| hi | Hindi | Devanagari |
| kn | Kannada | Kannada |
| ml | Malayalam | Malayalam |
| mr | Marathi | Devanagari |
| pa | Punjabi | Gurmukhi |
| sd | Sindhi | Arabic |
| si | Sinhala | Sinhala |
| ta | Tamil | Tamil |
| te | Telugu | Telugu |
| ur | Urdu | Arabic |

### Aksharantar (20 languages)
| Code | Language | Script |
|------|----------|--------|
| asm | Assamese | Bengali |
| ben | Bengali | Bengali |
| brx | Bodo | Devanagari |
| guj | Gujarati | Gujarati |
| hin | Hindi | Devanagari |
| kan | Kannada | Kannada |
| kas | Kashmiri | Arabic/Devanagari |
| kok | Konkani | Devanagari |
| mai | Maithili | Devanagari |
| mal | Malayalam | Malayalam |
| mar | Marathi | Devanagari |
| mni | Manipuri | Bengali |
| nep | Nepali | Devanagari |
| ori | Oriya | Odia |
| pan | Punjabi | Gurmukhi |
| san | Sanskrit | Devanagari |
| sid | Sindhi | Arabic |
| tam | Tamil | Tamil |
| tel | Telugu | Telugu |
| urd | Urdu | Arabic |

## References

- [Dakshina Paper](https://arxiv.org/abs/2007.01176) - Google Research
- [Aksharantar Paper](https://arxiv.org/abs/2205.03018) - AI4Bharat
- [IndicXlit](https://github.com/AI4Bharat/IndicXlit) - Neural transliteration model

## Useful Resources

### Word Frequency Data

- [wordfreq](https://pypi.org/project/wordfreq/) - Python library for word frequency lookups
  - Supports 40+ languages including Hindi
  - Can help identify common vs rare words
  - Useful for filtering English loanwords vs native Hindi

```python
# Example: Check if a word is common in Hindi
from wordfreq import word_frequency, zipf_frequency

# Higher frequency = more common
word_frequency('नमस्ते', 'hi')  # Returns frequency
zipf_frequency('नमस्ते', 'hi')  # Returns Zipf scale (1-7, higher = more common)
```

### Hindi Word Lists

- [Hindi Learner Dictionary](https://github.com/vbvss199/Language-Learning-decks) - 14K common Hindi words with translations

### Local Dictionaries (datasets/dictionary/)

| File | Description | Entries |
|------|-------------|---------|
| `English-Hindi Dictionary.csv` | English→Hindi word mappings | ~198K |
| `hindi.json` | Hindi words with romanizations & translations | ~14K |
| `vardha hindi dictionary.pdf` | Comprehensive Hindi dictionary (PDF) | - |
