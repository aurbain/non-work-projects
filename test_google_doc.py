import requests
from urllib.parse import urlparse


def extract_file_id(url: str) -> str:
    """Extract Google Doc file ID from the URL."""
    parsed = urlparse(url)
    
    # For published URLs: https://docs.google.com/document/d/e/TOKEN/pub
    if '/d/e/' in parsed.path and '/pub' in parsed.path:
        # Extract the token between '/d/e/' and '/pub'
        parts = parsed.path.strip('/').split('/')
        for i, part in enumerate(parts):
            if part == 'd' and i < len(parts) - 1 and parts[i + 1] == 'e':
                return parts[i + 2]
    
    # Check for standard Google Docs URL patterns
    for i, part in enumerate(parsed.path.strip('/').split('/')):
        if part == 'document' and i < len(parsed.path.strip('/').split('/')) - 1:
            return parsed.path.strip('/').split('/')[i + 1]
        elif part == 'd' and i < len(parsed.path.strip('/').split('/')) - 1:
            return parsed.path.strip('/').split('/')[i + 1]
    
    return None


def fetch_document_content(file_id: str) -> str:
    """
    Fetch document content from Google Drive API or export endpoint.
    Note: In production, you would use Google Drive API with proper authentication.
    This uses the public export endpoint if available.
    """
    # Try export endpoint first (for public docs)
    export_url = f"https://docs.google.com/document/d/{file_id}/export?format=txt"
    
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
    }
    
    try:
        response = requests.get(export_url, headers=headers, timeout=30)
        
        if response.status_code == 200:
            return response.text
        
        # Fallback: Try HTML export which might have the raw content
        html_export_url = f"https://docs.google.com/document/d/{file_id}/export?format=html"
        response = requests.get(html_export_url, headers=headers, timeout=30)
        
        if response.status_code == 200:
            return response.text
            
    except requests.exceptions.RequestException as e:
        print(f"Warning: Could not fetch document via export endpoints: {e}")
        print("In production, use Google Drive API v3 with OAuth2 authentication")
    
    return None


def parse_grid_data(content: str) -> list:
    """
    Parse the document content into coordinate-character pairs.
    Expected format:
    x-coordinate
    Character
    y-coordinate
    0
    █
    0
    ...
    """
    # Split content and remove empty lines
    lines = [line.strip() for line in content.split('\n') if line.strip()]
    
    # Remove headers
    lines = [line for line in lines if line.lower() not in 
             ['x-coordinate', 'character', 'y-coordinate']]
    
    data = []
    i = 0
    
    while i < len(lines):
        # Skip if we've run out of data
        if i >= len(lines):
            break
        
        x = lines[i].strip()
        if not x or not x.isdigit():
            i += 1
            continue
        
        # Check if we have enough lines for character and y-coordinate
        if i + 2 >= len(lines):
            break
            
        char = lines[i + 1].strip()
        y = lines[i + 2].strip()
        
        if char:  # Only add if character exists
            data.append({
                'x': int(x),
                'character': char,
                'y': int(y)
            })
        
        i += 3
    
    return data


def build_grid_from_data(data: list) -> list:
    """
    Build a 2D grid from the coordinate-character data.
    Empty positions are filled with spaces.
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
        x = item['x']
        y = item['y']
        char = item['character']
        grid[y][x] = char
    
    return grid


def print_grid(grid: list):
    """Print the grid in fixed-width format."""
    if not grid:
        print("Empty grid")
        return
    
    # Remove any trailing empty rows
    while grid and all(cell == ' ' for cell in grid[-1]):
        grid.pop()
    
    # Print each row
    for row in grid:
        # Join characters with no separator (fixed-width ensures proper alignment)
        print(''.join(row))


def print_unicode_grid_from_google_doc(doc_url: str):
    """
    Takes a Google Doc URL containing Unicode characters and their grid positions,
    retrieves the data, parses it, and prints the resulting grid.
    
    Args:
        doc_url: URL of the Google Doc (e.g., https://docs.google.com/document/d/FILE_ID/edit)
    """
    
    # Step 1: Extract document ID from URL
    file_id = extract_file_id(doc_url)
    
    if not file_id:
        raise ValueError("Could not extract document ID from URL")
    
    print(f"Extracted file ID: {file_id}")
    
    # Step 2: Fetch the document content
    doc_content = fetch_document_content(file_id)
    
    if not doc_content:
        raise ValueError("Could not fetch document content")
    
    print(f"Document content length: {len(doc_content)} characters")
    
    # Step 3: Parse the data (x-coordinate, Character, y-coordinate format)
    grid_data = parse_grid_data(doc_content)
    
    if not grid_data:
        print("No grid data found in document")
        return
    
    print(f"Parsed {len(grid_data)} coordinate-character pairs")
    
    # Step 4: Build the 2D grid
    grid = build_grid_from_data(grid_data)
    
    # Step 5: Print the grid with fixed-width formatting
    print_grid(grid)


# Test with the provided URL
if __name__ == "__main__":
    url = "https://docs.google.com/document/d/e/2PACX-1vTMOmshQe8YvaRXi6gEPKKlsC6UpFJSMAk4mQjLm_u1gmHdVVTaeh7nBNFBRlui0sTZ-snGwZM4DBCT/pub"
    
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
