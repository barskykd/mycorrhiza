# Format Conversion

Mycorrhiza Wiki supports two markup formats:
- **Mycomarkup** (`.myco` files) - the native markup format
- **Markdown** (`.md` files) - GitHub Flavored Markdown

## Converting Between Formats

You can convert all hyphae in your wiki from one format to another using the `-convert-format` command-line flag.

### Usage

```bash
mycorrhiza -convert-format <format> <wiki-directory>
```

Where `<format>` can be:
- `markdown` or `md` - Convert all hyphae to Markdown
- `mycomarkup` or `myco` - Convert all hyphae to Mycomarkup

### Examples

Convert all hyphae to Markdown:
```bash
mycorrhiza -convert-format markdown /path/to/wiki
```

Convert all hyphae to Mycomarkup:
```bash
mycorrhiza -convert-format mycomarkup /path/to/wiki
```

### Important Notes

⚠️ **BACKUP YOUR WIKI FIRST!**

This operation will:
- Modify hypha files in place
- Change file extensions (`.myco` ↔ `.md`)
- Convert markup syntax between formats

The conversion process will:
1. Index all hyphae in your wiki
2. Count how many need conversion
3. Ask for confirmation before proceeding
4. Convert each hypha using AST-based parsing
5. Rename files with the appropriate extension
6. Report progress and any warnings
7. **Commit all changes to the wiki's git repository**

### What Gets Converted

The converter handles:
- ✅ Headings (all levels)
- ✅ Bold and italic text
- ✅ Inline code and code blocks
- ✅ Links (internal wiki links and external URLs)
- ✅ Lists (ordered and unordered)
- ✅ Horizontal rules
- ✅ Blockquotes (Markdown to Mycomarkup)

### Format-Specific Features

Some features are specific to one format and cannot be converted:

**Mycomarkup-only features** (when converting to Markdown):
- Superscript (`^^text^^`)
- Subscript (`,,text,,`)
- Underline (`__text__`)
- Highlighting (`++text++`)
- Rocket links (`=> url`)
- Transclusions (`<= hypha`)
- Image blocks (`img { }`)
- Table blocks (`table { }`)

These will generate warnings during conversion and may need manual adjustment.

**Markdown-only features** (when converting to Mycomarkup):
- Strikethrough (`~~text~~`)
- Task lists
- Some advanced table features

### Example Output

```
INFO Indexing hyphae...
INFO Indexed hyphae n=142

WARNING: This will convert 89 hyphae to Markdown format.
This operation will modify files in your wiki directory.
It is STRONGLY recommended to backup your wiki before proceeding.

Continue? (yes/no): yes

INFO Starting format conversion targetFormat=Markdown
INFO Conversion progress converted=100 skipped=53 failed=0
INFO Conversion complete converted=89 skipped=53 failed=0 total=142
INFO Committing changes to git repository...
INFO Changes committed to git repository
```

### Git Integration

All conversions are automatically committed to your wiki's git repository with:
- **Commit message**: "Convert N hyphae to Format format"
- **Operation type**: Markup migration
- **Author**: wikimind (automated conversion)

The commit includes:
- File renames (when extension changes, e.g., `.myco` → `.md`)
- File modifications (when extension stays the same)

You can view the conversion in your git history:
```bash
cd /path/to/wiki
git log -1 --stat
```

### Error Handling

The converter will:
- Skip hyphae already in the target format
- Log warnings for features that can't be converted
- Continue processing if individual hyphae fail
- Report a summary of successes, skips, and failures

If the conversion fails for some hyphae, check the logs for details. The successfully converted hyphae will remain converted, while failed ones will keep their original format.

### Performance

The converter processes hyphae sequentially and logs progress every 100 hyphae. For large wikis (1000+ hyphae), the conversion may take a few minutes.
