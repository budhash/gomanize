#!/bin/bash
#
# Extract Aksharantar dataset for a specific language
# Usage: ./generate_dataset.sh <lang_code> [min_score]
#
# Examples:
#   ./generate_dataset.sh hi           # Hindi, human-verified only (default)
#   ./generate_dataset.sh hi -0.15     # Hindi, good quality balance
#   ./generate_dataset.sh hi -999      # Hindi, all entries
#
# Available languages (2-letter codes):
#   as - Assamese      ml - Malayalam     sa - Sanskrit
#   bn - Bengali       mr - Marathi       sd - Sindhi
#   bx - Bodo          mn - Manipuri      ta - Tamil
#   gu - Gujarati      ne - Nepali        te - Telugu
#   hi - Hindi         or - Oriya         ur - Urdu
#   kn - Kannada       pa - Punjabi
#   ks - Kashmiri      kk - Konkani
#   mt - Maithili
#
# Output: aksharantar_<lang>.csv (e.g., aksharantar_hi.csv)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Map 2-letter codes to 3-letter Aksharantar codes
map_lang() {
    case "$1" in
        as) echo "asm" ;; bn) echo "ben" ;; bx) echo "brx" ;; gu) echo "guj" ;;
        hi) echo "hin" ;; kn) echo "kan" ;; ks) echo "kas" ;; kk) echo "kok" ;;
        mt) echo "mai" ;; ml) echo "mal" ;; mn) echo "mni" ;; mr) echo "mar" ;;
        ne) echo "nep" ;; or) echo "ori" ;; pa) echo "pan" ;; sa) echo "san" ;;
        sd) echo "sid" ;; ta) echo "tam" ;; te) echo "tel" ;; ur) echo "urd" ;;
        *) echo "" ;;
    esac
}

# Check arguments
if [ -z "$1" ]; then
    echo "Usage: $0 <lang_code> [min_score]"
    echo ""
    echo "Arguments:"
    echo "  lang_code   - Language code (see below)"
    echo "  min_score   - Minimum confidence score (default: 0, human-verified only)"
    echo "                Scores: 0=human-verified, -0.15=good balance, -999=all"
    echo ""
    echo "Available languages (2-letter codes):"
    echo "  as - Assamese      ml - Malayalam     sa - Sanskrit"
    echo "  bn - Bengali       mr - Marathi       sd - Sindhi"
    echo "  bx - Bodo          mn - Manipuri      ta - Tamil"
    echo "  gu - Gujarati      ne - Nepali        te - Telugu"
    echo "  hi - Hindi         or - Oriya         ur - Urdu"
    echo "  kn - Kannada       pa - Punjabi"
    echo "  ks - Kashmiri      kk - Konkani"
    echo "  mt - Maithili"
    echo ""
    echo "Examples:"
    echo "  $0 hi"
    echo "  $0 mr -0.15"
    echo ""
    echo "Output: aksharantar_<lang>.csv"
    exit 1
fi

INPUT_LANG="$1"
# Default to 0 (only human-verified entries from Dakshina)
# Scores are log probabilities: null=human-verified, negative=model-scored
# Use -999 to accept all entries, -0.15 for good balance
MIN_SCORE="${2:-0}"

# Map to 3-letter code
LANG=$(map_lang "$INPUT_LANG")
if [ -z "$LANG" ]; then
    echo "Error: Unknown language code '$INPUT_LANG'"
    echo "Run '$0' without arguments to see available languages."
    exit 1
fi

ZIP_FILE="$SCRIPT_DIR/$LANG.zip"
OUTPUT_FILE="$SCRIPT_DIR/aksharantar_${INPUT_LANG}.csv"

# Check zip file exists
if [ ! -f "$ZIP_FILE" ]; then
    echo "Error: $LANG.zip not found in $SCRIPT_DIR"
    echo ""
    echo "Download Aksharantar dataset first:"
    echo "  curl -sL https://huggingface.co/datasets/ai4bharat/Aksharantar/resolve/main/$LANG.zip -o $ZIP_FILE"
    exit 1
fi

echo "Extracting Aksharantar $INPUT_LANG ($LANG) dataset (min score: $MIN_SCORE)..."

# Create temp dir for extraction
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Extract zip to temp
unzip -q "$ZIP_FILE" -d "$TEMP_DIR"

# Combine all splits (train + valid + test) into one file
# Format: native,roman,notes
echo "native,roman,notes" > "$OUTPUT_FILE"

TOTAL=0
for split in train valid test; do
    # Files are at root of zip (e.g., hin_train.json)
    JSON_FILE=$(find "$TEMP_DIR" -name "${LANG}_${split}.json" 2>/dev/null | head -1)

    if [ -z "$JSON_FILE" ]; then
        JSON_FILE="$TEMP_DIR/${LANG}_${split}.json"
    fi

    if [ ! -f "$JSON_FILE" ]; then
        echo "  $split: not found, skipping"
        continue
    fi

    # Process JSON lines, filter by score, output CSV
    # JSON format: {"unique_identifier":"hin1","native word":"...","english word":"...","source":"...","score":0.0 or null}
    COUNT=$(cat "$JSON_FILE" | \
        python3 -c "
import sys
import json

min_score = float('$MIN_SCORE')
split = '$split'

for line in sys.stdin:
    try:
        entry = json.loads(line.strip())
        score = entry.get('score')
        if score is None:
            score = 0.0
        if score >= min_score:
            native = entry.get('native word', '').replace(',', '\\\\,')
            roman = entry.get('english word', '').replace(',', '\\\\,')
            source = entry.get('source', '')
            print(f'{native},{roman},source={source} split={split}')
    except:
        pass
" | tee -a "$OUTPUT_FILE" | wc -l | tr -d ' ')

    echo "  $split: $COUNT entries"
    TOTAL=$((TOTAL + COUNT))
done

echo ""
echo "Done! Created $OUTPUT_FILE ($TOTAL entries)"
