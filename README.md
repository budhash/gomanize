# Gomanize

A Go library and CLI tool for transliterating Devanagari script (Hindi) to Latin/Roman characters. Designed for practical use cases like song lyrics romanization.

## Features

- Romanize Hindi text from Devanagari script into readable Latin characters
- CLI tool for quick transliteration
- Library API for integration into Go projects
- Follows colloquial/phonetic Hindi romanization (not scholarly IAST)

## Installation

```bash
# As a library
go get github.com/budhash/gomanize

# Build CLI from source
git clone https://github.com/budhash/gomanize
cd gomanize
make build
```

## Usage

### CLI

```bash
# Direct argument
./gomanize "नमस्ते भारत"
# Output: namste bharat

# Pipe input
echo "हिंदी गाना" | ./gomanize
# Output: hindi gana

# Version
./gomanize --version
```

### Library

```go
package main

import (
    "fmt"
    gomanize "github.com/budhash/gomanize"
)

func main() {
    g, err := gomanize.New("hindi")
    if err != nil {
        panic(err)
    }

    output := g.Translit("नमस्ते दुनिया")
    fmt.Println(output)  // "namste duniya"
}
```

## Development

```bash
# First-time setup
make init

# Development workflow
make dev      # Format, vet, test
make ci       # Full CI pipeline

# Testing
make test           # Run all tests
make test-unit      # Unit tests only (fast)
make test-cover     # Tests with coverage
make test-dakshina  # Dakshina accuracy tests

# See all commands
make help
```

## Transliteration Standards

This project follows **colloquial/phonetic Hindi romanization**:
- Aggressive schwa deletion to match spoken Hindi
- No diacritics (aa not ā, i not ī)
- Phonetic spelling (kh not ḵẖ)

### References
- [Hunterian Transliteration](https://en.wikipedia.org/wiki/Hunterian_transliteration) - Official India standard
- [Schwa Deletion Rules](https://en.wikipedia.org/wiki/Schwa_deletion_in_Indo-Aryan_languages)
- [Google Dakshina Dataset](https://github.com/google-research-datasets/dakshina) - Test data source

## Current Status

| Dataset | Accuracy |
|---------|----------|
| Original (hindi-common.txt) | 56.9% |
| Dakshina (native Hindi) | 48.8% |
| Target | 80%+ |

See [Claude.md](Claude.md) for detailed failure analysis and roadmap.

## License

MIT License - Copyright (c) 2023 Budhaditya

## Author

Budhaditya (budhash@gmail.com)
