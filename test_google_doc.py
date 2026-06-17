import requests
from urllib.parse import urlparse
from typing import List, Dict, Any, Optional


def extract_file_id(url: str) -> Optional[str]:
    """Extract Google Doc file ID from the URL.

    Handles both standard URLs and published URLs with tokens.

    Args:
        url: Google Doc URL

    Returns:
        The file ID as a string, or None if not found
    """
    parsed = urlparse(url)

    # For published URLs: https://docs.google.com/document/d/e/TOKEN/pub
    if '/d/e/' in parsed.path and '/pub' in parsed.path:
        parts = parsed.path.strip('/').split('/')
        for i, part in enumerate(parts):
            if part == 'd' and i < len(parts) - 1 and parts[i + 1] == 'e':
                return parts[i + 2]

    # Check for standard Google Docs URL patterns
    path_parts = parsed.path.strip('/').split('/')
    for i, part in enumerate(path_parts):
        if part == 'document' and i < len(path_parts) - 1:
            return path_parts[i + 1]
        elif part == 'd' and i < len(path_parts) - 1:
            return path_parts[i + 1]

    return None


def fetch_document_content(file_id: str) -> Optional[str]:
    """Fetch document content from Google Docs.

    Tries the export endpoint first, then falls back to parsing the rendered page.

    Args:
        file_id: Google Doc file ID

    Returns:
        Document content as string, or None if fetch fails
    """
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
    }

    # Try export endpoint first (for public docs)
    export_url = f"https://docs.google.com/document/d/{file_id}/export?format=txt"

    try:
        response = requests.get(export_url, headers=headers, timeout=30)

        if response.status_code == 200:
            return response.text
    except requests.exceptions.RequestException:
        pass

    # Fallback: Try HTML export
    html_export_url = f"https://docs.google.com/document/d/{file_id}/export?format=html"

    try:
        response = requests.get(html_export_url, headers=headers, timeout=30)

        if response.status_code == 200:
            return response.text
    except requests.exceptions.RequestException:
        pass

    # Last resort: Try to parse the rendered page
    rendered_url = f"https://docs.google.com/document/d/{file_id}"

    try:
        response = requests.get(rendered_url, headers=headers, timeout=30)

        if response.status_code == 200:
            return response.text
    except requests.exceptions.RequestException:
        pass

    return None


def parse_grid_data(content: str) -> List[Dict[str, Any]]:
    """Parse the document content into coordinate-character pairs.

    Expected format: alternating x-coordinate, character, y-coordinate lines.
    Headers like "x-coordinate", "Character", "y-coordinate" are skipped.
    Also handles tab-separated values from table exports.

    Args:
        content: Raw document content

    Returns:
        List of dictionaries with 'x', 'character', and 'y' keys
    """
    # If it's HTML, try to extract the data table
    if '<table' in content.lower() or '<!--' in content:
        content = parse_html_table_content(content)

    lines = [line.strip() for line in content.split('\n') if line.strip()]

    # Remove header lines
    lines = [line for line in lines if line.lower() not in
             ['x-coordinate', 'character', 'y-coordinate', 'grid data']]

    data = []
    i = 0

    while i < len(lines):
        x = lines[i].strip()

        # Skip if empty or not a digit
        if not x or not x.isdigit():
            i += 1
            continue

        # Need character and y-coordinate
        if i + 2 >= len(lines):
            break

        char = lines[i + 1].strip()
        y = lines[i + 2].strip()

        if char:
            data.append({
                'x': int(x),
                'character': char,
                'y': int(y)
            })

        i += 3

    return data


def parse_html_table_content(html: str) -> str:
    """Parse table content from HTML to extract plain text data.

    Args:
        html: HTML content containing the data table

    Returns:
        Extracted plain text data
    """
    # Find the table content between common markers
    import re

    # Try to find text content related to grid data
    # Look for patterns like "x-coordinate", "Character", "y-coordinate"
    lines = html.split('\n')
    in_table = False
    extracted_lines = []

    for line in lines:
        line = line.strip()

        # Skip HTML tags and scripts
        if line.startswith('<') or '</' in line:
            continue

        # Check for data indicators
        if 'x-coordinate' in line.lower() or 'Character' in line or 'y-coordinate' in line.lower():
            continue
        elif line.isdigit():
            extracted_lines.append(line)
        elif len(line) == 1 and line.isascii():
            # Could be a character
            extracted_lines.append(line)
        elif len(line) <= 3 and line and line[0].isalpha():
            # Might be a character or partial data
            extracted_lines.append(line)

    return '\n'.join(extracted_lines)


def build_grid_from_data(data: List[Dict[str, Any]]) -> List[str]:
    """Build a 2D grid (as list of strings) from coordinate-character data.

    Args:
        data: List of dictionaries with 'x', 'character', and 'y' keys

    Returns:
        List of strings representing each row of the grid
    """
    if not data:
        return []

    # Find maximum coordinates
    max_x = max(item['x'] for item in data)
    max_y = max(item['y'] for item in data)

    # Create grid initialized with spaces
    grid = [[' ' for _ in range(max_x + 1)] for _ in range(max_y + 1)]

    # Fill in the characters
    for item in data:
        grid[item['y']][item['x']] = item['character']

    # Convert to list of strings
    return [''.join(row) for row in grid]


def print_unicode_grid_from_google_doc(doc_url: str) -> None:
    """Fetch a Google Doc and print the Unicode grid it contains.

    This function takes a URL to a Google Doc that contains coordinate-data
    representing a grid of Unicode characters. It retrieves the data, parses
    it, and prints the resulting visual grid of uppercase letters.

    Args:
        doc_url: URL of the Google Doc (e.g.,
                 https://docs.google.com/document/d/FILE_ID/edit or
                 https://docs.google.com/document/d/e/TOKEN/pub)

    Example:
        >>> print_unicode_grid_from_google_doc("https://docs.google.com/document/d/e/TOKEN/pub")
        ABCDE
        FGHIJ
        KLMNO
    """
    # Step 1: Extract document ID from URL
    file_id = extract_file_id(doc_url)

    if not file_id:
        raise ValueError(f"Could not extract document ID from URL: {doc_url}")

    print(f"Extracted file ID: {file_id}")

    # Step 2: Fetch the document content
    doc_content = fetch_document_content(file_id)

    if not doc_content:
        raise ValueError("Could not fetch document content. The document may not be public or accessible.")

    print(f"Document content length: {len(doc_content)} characters")

    # Step 3: Parse the data (x-coordinate, character, y-coordinate format)
    grid_data = parse_grid_data(doc_content)

    if not grid_data:
        print("No grid data found in document")
        return

    print(f"Parsed {len(grid_data)} coordinate-character pairs")

    # Step 4: Build the 2D grid as a list of strings
    grid = build_grid_from_data(grid_data)

    # Step 5: Print the grid
    for row in grid:
        print(row)


if __name__ == "__main__":
    # Example usage
    url = "https://docs.google.com/document/d/e/2PACX-1vSvM5gDlNvt7npYHhp_XfsJvuntUhq184By5xO_pA4b_gCWeXb6dM6ZxwN8rE6S4ghUsCj2VKR21oEP/pub"

    try:
        print("=" * 60)
        print("Testing Google Doc Unicode Grid Parser")
        print("=" * 60)
        print()

        print_unicode_grid_from_google_doc(url)

        print()
        print("=" * 60)
        print("Test completed successfully!")
        print("=" * 60)
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()
