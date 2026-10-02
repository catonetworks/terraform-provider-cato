#!/usr/bin/env python3
"""Convert the fixed Sections/Rules YAML format to Markdown.

Requires PyYAML (install with: python3 -m pip install PyYAML).
Usage:
    python3 guidelines-to-markdown.py coding-rules.yaml
    python3 guidelines-to-markdown.py coding-rules.yaml -o guidelines.md

Statement and Rationale retain their Markdown formatting.
"""

import argparse
from pathlib import Path
import sys

try:
    import yaml
except ModuleNotFoundError:
    sys.exit("PyYAML is required. Install it with: python3 -m pip install PyYAML")


RULE_FIELDS = ("ID", "Slug", "Level", "Statement", "Rationale")


def require_text(mapping, key, location, single_line=False):
    value = mapping.get(key)
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"{location}.{key} must be a non-empty string")
    value = value.strip()
    if single_line and ("\n" in value or "\r" in value):
        raise ValueError(f"{location}.{key} must be a single line")
    return value


def to_markdown(document):
    """Validate the fixed schema and render every section and rule in order."""
    if not isinstance(document, dict) or not isinstance(document.get("Sections"), list):
        raise ValueError("YAML must contain a Sections list")

    lines = ["# Coding Guidelines", ""]
    for section_index, section in enumerate(document["Sections"], start=1):
        location = f"Sections[{section_index}]"
        if not isinstance(section, dict):
            raise ValueError(f"{location} must be a mapping")
        name = require_text(section, "Name", location, single_line=True)
        rules = section.get("Rules")
        if not isinstance(rules, list):
            raise ValueError(f"{location}.Rules must be a list")
        lines.extend([f"## {name}", ""])

        for rule_index, rule in enumerate(rules, start=1):
            rule_location = f"{location}.Rules[{rule_index}]"
            if not isinstance(rule, dict):
                raise ValueError(f"{rule_location} must be a mapping")
            values = {
                field: require_text(
                    rule, field, rule_location,
                    single_line=field in ("ID", "Slug", "Level"),
                )
                for field in RULE_FIELDS
            }
            lines.extend([
                f"### {values['ID']}: {values['Slug']}",
                "",
                f"**Level:** {values['Level']}",
                "",
                values["Statement"],
                "",
                "**Rationale:**",
                "",
                values["Rationale"],
                "",
            ])

    return "\n".join(lines).rstrip() + "\n"


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("input", type=Path, help="YAML guidelines file")
    parser.add_argument("-o", "--output", type=Path, help="Markdown output file (default: stdout)")
    args = parser.parse_args()

    try:
        if args.output and args.input.resolve() == args.output.resolve():
            raise ValueError("Input and output paths must differ")
        if args.output and args.output.exists() and args.input.samefile(args.output):
            raise ValueError("Input and output must not refer to the same file")
        with args.input.open(encoding="utf-8") as source:
            # SafeLoader prevents YAML from constructing arbitrary Python objects.
            markdown = to_markdown(yaml.safe_load(source))
        if args.output:
            args.output.write_text(markdown, encoding="utf-8")
        else:
            sys.stdout.write(markdown)
    except (OSError, UnicodeError, ValueError, yaml.YAMLError) as error:
        parser.exit(1, f"error: {error}\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
