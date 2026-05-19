# Markdown to DOCX Converter

A Python script that converts Markdown files to Microsoft Word DOCX format.

## Features

- Converts Markdown headers (H1-H6) to Word heading styles
- Supports **bold** and *italic* text formatting
- Handles unordered and ordered lists
- Converts blockquotes to italic text
- Processes inline code with monospace font
- Converts markdown links to underlined text
- Preserves paragraph structure

## Installation

1. Install the required dependency:
```bash
pip install -r requirements.txt
```

## Usage

### Basic Usage
```bash
python md_to_docx_converter.py input.md
```

This will create `input.docx` in the same directory.

### Specify Output File
```bash
python md_to_docx_converter.py input.md -o output.docx
```

### Verbose Output
```bash
python md_to_docx_converter.py input.md -v
```

## Example

Convert the included example file:
```bash
python md_to_docx_converter.py example.md -v
```

This will create `example.docx` with properly formatted content.

## Supported Markdown Syntax

- Headers: `#`, `##`, `###`, etc.
- Bold: `**text**` or `__text__`
- Italic: `*text*` or `_text_`
- Inline code: `` `code` ``
- Links: `[text](url)`
- Unordered lists: `- item` or `* item`
- Ordered lists: `1. item`
- Blockquotes: `> quote`

## Requirements

- Python 3.6+
- python-docx library

## License

This script is provided as-is for educational and personal use.
