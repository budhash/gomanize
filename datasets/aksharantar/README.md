# Aksharantar Dataset

**Source**: [AI4Bharat Aksharantar](https://huggingface.co/datasets/ai4bharat/Aksharantar)
**Paper**: [Aksharantar: Towards Building Open Transliteration Tools for the Next Billion Users](https://arxiv.org/abs/2205.03018)
**License**: CC BY 4.0

## Overview

The Aksharantar dataset contains **26 million transliteration pairs** across 20 Indic languages. Each entry maps native script words to their romanized (Latin script) transliterations.

## Languages

| Code | Language | Script | Train | Valid | Test |
|------|----------|--------|-------|-------|------|
| asm | Assamese | Bengali | 178,630 | 3,788 | 5,506 |
| ben | Bengali | Bengali | 1,231,428 | 11,276 | 14,167 |
| brx | Bodo | Devanagari | 35,618 | 3,068 | 4,081 |
| guj | Gujarati | Gujarati | 1,143,212 | 12,419 | 18,077 |
| hin | Hindi | Devanagari | 1,299,155 | 6,357 | 10,112 |
| kan | Kannada | Kannada | 2,906,728 | 7,025 | 11,380 |
| kas | Kashmiri | Arabic/Devanagari | 46,635 | 4,456 | 6,908 |
| kok | Konkani | Devanagari | 612,525 | 3,502 | 5,042 |
| mai | Maithili | Devanagari | 282,639 | 3,790 | 5,449 |
| mal | Malayalam | Malayalam | 4,100,621 | 7,613 | 12,451 |
| mar | Marathi | Devanagari | 1,452,748 | 7,646 | 12,190 |
| mni | Manipuri | Bengali | 10,060 | 3,260 | 4,889 |
| nep | Nepali | Devanagari | 2,397,414 | 2,804 | 4,101 |
| ori | Oriya | Odia | 346,492 | 3,093 | 4,228 |
| pan | Punjabi | Gurmukhi | 514,724 | 8,880 | 11,237 |
| san | Sanskrit | Devanagari | 1,813,369 | 3,398 | 5,302 |
| sid | Sindhi | Arabic | 59,715 | 8,375 | 6,407 |
| tam | Tamil | Tamil | 3,230,902 | 8,824 | 11,499 |
| tel | Telugu | Telugu | 2,429,562 | 7,681 | 10,260 |
| urd | Urdu | Arabic | 699,024 | 12,419 | 14,878 |

## File Format

### CSV Format (Recommended)
Files: `{lang}_train.csv`, `{lang}_valid.csv`, `{lang}_test.csv`

```csv
id,native,english,source,score
hin1,मैट्रोलॉजिस्ट,maitrologist,AK-Freq,0.000000
hin2,पीएचडब्ल्यूसीएस,phwcs,AK-Freq,0.000000
```

### JSON Format (Original)
Files: `{lang}_train.json`, `{lang}_valid.json`, `{lang}_test.json`

```json
{"unique_identifier": "hin1", "native word": "मैट्रोलॉजिस्ट", "english word": "maitrologist", "source": "AK-Freq", "score": 0.0}
```

## Fields

| Field | Description |
|-------|-------------|
| id | Unique identifier (language code + number) |
| native | Word in native Indic script |
| english | Romanized transliteration |
| source | Data source (IndicCorp, Wikidata, Samanantar, AK-Freq, etc.) |
| score | Character-level confidence score (0 = not scored) |

## Devanagari Languages

For gomanize, the following languages use Devanagari script and can be tested:

- **hin** (Hindi) - Primary target
- **mar** (Marathi) - Similar phonology
- **nep** (Nepali) - Similar phonology
- **kok** (Konkani) - Similar phonology
- **mai** (Maithili) - Similar phonology
- **brx** (Bodo) - Different phonology
- **san** (Sanskrit) - Different conventions (scholarly IAST)

## Usage Notes

1. **Inconsistent Romanization**: Different annotators use different conventions (e.g., 'aa' vs 'a' for ा)
2. **Multiple Valid Spellings**: Some words have duplicate entries with different romanizations
3. **Score Field**: Lower scores indicate higher confidence for synthetic data

## Dataset Generation Methodology

### Data Sources

The dataset was built from **5 main sources**:

| Source | Code | Description |
|--------|------|-------------|
| **Existing datasets** | Exs | Pre-existing transliteration datasets (e.g., LDCIL) |
| **Wikidata** | Wik | Wikidata transliteration pairs |
| **Samanantar** | Sam | Mined from parallel corpus (Indic-English) |
| **IndicCorp** | Ind | Mined from monolingual Indic corpus |
| **Manual** | Man | Human-annotated transliterations |

### Mining Process

1. **Samanantar Mining**: Extracted transliteration pairs from parallel Indic-English sentence corpus by identifying matching word pairs

2. **IndicCorp Mining**: Extracted from monolingual Indic text using English dictionary matching and transliteration detection

Mining scripts available at:
- [Samanantar mining](https://github.com/AI4Bharat/IndicXlit/tree/master/data_mining/transliteration_mining_samanantar)
- [IndicCorp mining](https://github.com/AI4Bharat/IndicXlit/tree/master/data_mining/IndicCorp)

### Quality Control & Scoring

- **Score field**: Character-level log probability of Indic word given Roman word, computed by the IndicXlit model
- **Threshold**: Pairs with average score ≥ 0.35 were kept
- **Validation tools**: `Transliteration_Checker.java` and `Transliteration_Checker.py`

### Test Set Subsets

| Subset | Description |
|--------|-------------|
| AK-Freq | Most frequent words |
| AK-Uni | Uniformly sampled words |
| AK-NEF | Named entities (foreign) |
| AK-NEI | Named entities (Indian) |

### Licensing

- **Mined data** (Samanantar, IndicCorp): CC0 license
- **Manual annotations**: CC-BY license

## Extraction

```bash
cd datasets/aksharantar

# Download a language if not present (~varies by language)
curl -sL https://huggingface.co/datasets/ai4bharat/Aksharantar/resolve/main/hin.zip -o hin.zip

# Extract Hindi (human-verified only, default)
./generate_dataset.sh hi
# Output: aksharantar_hi.csv

# Extract with different score thresholds
./generate_dataset.sh hi -0.15   # Good quality balance
./generate_dataset.sh hi -999    # All entries

# Extract other languages
./generate_dataset.sh bn         # Bengali
./generate_dataset.sh mr         # Marathi
```

### Score Thresholds

| Threshold | Description | Hindi Entries |
|-----------|-------------|---------------|
| 0 (default) | Human-verified only (Dakshina source) | ~343K |
| -0.15 | Good quality balance | ~653K |
| -999 | All entries | ~1.3M |

Scores are log probabilities from the IndicXlit model. `null` scores (stored as 0) indicate human-verified entries from the Dakshina dataset.

## Output Format

```csv
native,roman,notes
जन्मदिवस,janamdivas,source=Dakshina,score=0.0000,split=train
मैट्रोलॉजिस्ट,maitrologist,source=AK-Freq,score=0.0000,split=test
```

The `notes` column contains:
- `source` - Data source (Dakshina, AK-Freq, AK-Uni, IndicCorp, etc.)
- `score` - Confidence score (0 = not scored, higher = better)
- `split` - Original split (train, valid, test)

## References

- [Aksharantar Paper](https://arxiv.org/abs/2205.03018) - "Aksharantar: Towards Building Open Transliteration Tools for the Next Billion Users"
- [IndicXlit GitHub](https://github.com/AI4Bharat/IndicXlit) - Neural transliteration model (11M params)
- [Aksharantar on HuggingFace](https://huggingface.co/datasets/ai4bharat/Aksharantar)
- [AI4Bharat Tools](https://ai4bharat.iitm.ac.in/tools) - Online transliteration demo
- [LDCIL](https://ldcil.org/) - Linguistic Data Consortium for Indian Languages
- [LDCIL Data Portal](https://data.ldcil.org/) - Indian language datasets and corpora
- [Anuvadika](https://anuvadika.ciil.org/index.php) - CIIL transliteration tool

## Citation

```bibtex
@article{madhani2022aksharantar,
  title={Aksharantar: Towards Building Open Transliteration Tools for the Next Billion Users},
  author={Madhani, Yash and Parthasarathy, Sriram and Bedekar, Aditya and others},
  journal={arXiv preprint arXiv:2205.03018},
  year={2022}
}
```

## Downloaded

- Date: 2025-01-06
- Version: v1.0 (June 2022)
