#!/bin/bash
#
# Extract Dakshina dataset for a specific language
# Usage: ./generate_dataset.sh <lang_code> [min_attestations]
#
# Examples:
#   ./generate_dataset.sh hi           # Hindi, 4+ attestations (default)
#   ./generate_dataset.sh hi 1         # Hindi, all attestations
#   ./generate_dataset.sh mr 4         # Marathi, 4+ attestations
#
# Available languages: bn, gu, hi, kn, ml, mr, pa, sd, si, ta, te, ur
#
# Output: dakshina_<lang>.csv (e.g., dakshina_hi.csv)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TAR_FILE="$SCRIPT_DIR/dakshina.tar"

# Check arguments
if [ -z "$1" ]; then
    echo "Usage: $0 <lang_code> [min_attestations]"
    echo ""
    echo "Arguments:"
    echo "  lang_code        - Language code (see below)"
    echo "  min_attestations - Minimum human votes (default: 4)"
    echo ""
    echo "Available languages:"
    echo "  bn - Bengali      kn - Kannada       sd - Sindhi"
    echo "  gu - Gujarati     ml - Malayalam     si - Sinhala"
    echo "  hi - Hindi        mr - Marathi       ta - Tamil"
    echo "  pa - Punjabi      te - Telugu        ur - Urdu"
    echo ""
    echo "Examples:"
    echo "  $0 hi"
    echo "  $0 mr 4"
    echo ""
    echo "Output: dakshina_<lang>.csv"
    exit 1
fi

LANG="$1"
MIN_ATTEST="${2:-4}"
OUTPUT_FILE="$SCRIPT_DIR/dakshina_${LANG}.csv"

# Check tar file exists
if [ ! -f "$TAR_FILE" ]; then
    echo "Error: dakshina.tar not found in $SCRIPT_DIR"
    echo ""
    echo "Download Dakshina dataset first:"
    echo "  curl -L https://storage.googleapis.com/gresearch/dakshina/dakshina_dataset_v1.0.tar -o $TAR_FILE"
    exit 1
fi

echo "Extracting Dakshina $LANG dataset (min attestations: $MIN_ATTEST)..."

# Combine all splits (train + dev + test) into one file
# Format: native,roman,notes
echo "native,roman,notes" > "$OUTPUT_FILE"

TOTAL=0
for split in train dev test; do
    TSV_PATH="dakshina_dataset_v1.0/$LANG/lexicons/$LANG.translit.sampled.$split.tsv"

    COUNT=$(tar -xOf "$TAR_FILE" "$TSV_PATH" 2>/dev/null | \
        awk -F'\t' -v min="$MIN_ATTEST" -v sp="$split" '
            $3 >= min {
                # Escape any commas in the fields
                gsub(/,/, "\\,", $1)
                gsub(/,/, "\\,", $2)
                print $1 "," $2 ",attestations=" $3 " split=" sp
            }
        ' | tee -a "$OUTPUT_FILE" | wc -l | tr -d ' ')

    echo "  $split: $COUNT entries"
    TOTAL=$((TOTAL + COUNT))
done

echo ""
echo "Done! Created $OUTPUT_FILE ($TOTAL entries)"
