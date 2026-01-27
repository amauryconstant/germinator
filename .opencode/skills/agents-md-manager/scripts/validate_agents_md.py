#!/usr/bin/env python3
"""
AGENTS.md Validation Script

Validates AGENTS.md files against best practices for OpenCode.
"""

import re
import sys
from pathlib import Path
from typing import List, Tuple


class AGENTSValidator:
    def __init__(self, filepath: Path):
        self.filepath = filepath
        self.content = filepath.read_text()
        self.lines = self.content.split("\n")
        self.issues: List[Tuple[str, str, int]] = []

    def validate(self) -> bool:
        """Run all validation checks. Returns True if no errors."""
        self.check_length()
        self.estimate_tokens()
        self.check_anti_patterns()
        self.check_vague_language()
        self.check_structure()
        self.check_code_blocks()
        self.check_progressive_disclosure()

        return not any(severity == "ERROR" for severity, _, _ in self.issues)

    def check_length(self):
        """Check file length against recommendations."""
        line_count = len(self.lines)

        if line_count > 1000:
            self.issues.append(
                (
                    "ERROR",
                    f"File is {line_count} lines. Strongly consider splitting into modular .opencode/rules/ files (recommended max: 300 lines)",
                    0,
                )
            )
        elif line_count > 500:
            self.issues.append(
                (
                    "WARNING",
                    f"File is {line_count} lines. Consider splitting for better maintainability (recommended: 200-300 lines)",
                    0,
                )
            )
        elif line_count > 300:
            self.issues.append(
                (
                    "INFO",
                    f"File is {line_count} lines. Optimal range is 200-300 lines for best context efficiency",
                    0,
                )
            )

    def estimate_tokens(self):
        """Estimate token count (rough: 1 line ≈ 4 tokens)."""
        line_count = len(self.lines)
        estimated_tokens = line_count * 4

        if estimated_tokens > 4000:
            self.issues.append(
                (
                    "ERROR",
                    f"Estimated {estimated_tokens} tokens ({line_count} lines × 4). Critical: May cause opencode to ignore rules (recommended <1200 tokens)",
                    0,
                )
            )
        elif estimated_tokens > 2000:
            self.issues.append(
                (
                    "WARNING",
                    f"Estimated {estimated_tokens} tokens ({line_count} lines × 4). Consider splitting (recommended <1200 tokens)",
                    0,
                )
            )
        elif estimated_tokens > 1200:
            self.issues.append(
                (
                    "INFO",
                    f"Estimated {estimated_tokens} tokens ({line_count} lines × 4). Approaching recommended limit of 1200 tokens",
                    0,
                )
            )

    def check_anti_patterns(self):
        """Check for common anti-patterns that waste tokens."""
        anti_patterns = [
            (
                r"\b(clean code|good code|best practices|write good|code quality)\b",
                "Generic advice opencode already knows",
            ),
            (
                r"(is a|are a|was a|were a).+(created by|developed by|written by)",
                "Historical background opencode doesn't need",
            ),
            (
                r"^\s*#+\s*(introduction|overview|about this)",
                "Unnecessary introduction sections",
            ),
            (
                r"\b(step \d+|first,|second,|third,|finally,|let\'s)\b",
                "Tutorial language (be directive, not educational)",
            ),
            (
                r"(workspace \d+|command \d+|keybinding \d+)",
                "Exhaustive list detected (document patterns, not every item)",
            ),
        ]

        for line_num, line in enumerate(self.lines, 1):
            lower_line = line.lower()
            for pattern, message in anti_patterns:
                if re.search(pattern, lower_line):
                    self.issues.append(
                        ("WARNING", f"Possible anti-pattern: {message}", line_num)
                    )

    def check_vague_language(self):
        """Check for vague language that should be made specific."""
        vague_patterns = [
            r"\b(properly|correctly|appropriately|adequately)\b",
            r"\b(should be|needs to be|must be)\s+\w+ed\b",
            r"\b(as needed|when necessary|where appropriate)\b",
        ]

        for line_num, line in enumerate(self.lines, 1):
            if line.strip().startswith("```") or line.strip().startswith("    "):
                continue

            for pattern in vague_patterns:
                if re.search(pattern, line, re.IGNORECASE):
                    self.issues.append(
                        (
                            "INFO",
                            f'Vague language detected. Consider being more specific: "{line.strip()}"',
                            line_num,
                        )
                    )

    def check_structure(self):
        """Check markdown structure quality."""
        heading_pattern = r"^(#{1,6})\s+(.+)$"
        headings = []

        for line_num, line in enumerate(self.lines, 1):
            match = re.match(heading_pattern, line)
            if match:
                level = len(match.group(1))
                text = match.group(2)
                headings.append((level, text, line_num))

        prev_level = 0
        for level, text, line_num in headings:
            if level > prev_level + 1 and level > 3:
                # Only warn on large skips (H1 -> H4+), H1->H3 is acceptable
                self.issues.append(
                    (
                        "INFO",
                        f"Heading hierarchy skip: jumped from H{prev_level} to H{level}",
                        line_num,
                    )
                )
            prev_level = level

        if len(headings) < 3 and len(self.lines) > 100:
            self.issues.append(
                (
                    "WARNING",
                    "Few headings for file length. Consider adding more structure for scannability.",
                    0,
                )
            )

    def check_code_blocks(self):
        """Check for incomplete or improper code blocks."""
        in_code_block = False
        code_block_start = 0

        for line_num, line in enumerate(self.lines, 1):
            if line.strip().startswith("```"):
                if in_code_block:
                    in_code_block = False
                else:
                    in_code_block = True
                    code_block_start = line_num

        if in_code_block:
            self.issues.append(("ERROR", "Unclosed code block", code_block_start))

    def check_progressive_disclosure(self):
        """Check for progressive disclosure patterns."""
        has_imports = False
        has_cross_refs = False
        see_pattern = r"(see|refer to|details in|documented in)\s+[`\[]?[\w/.-]+\.md"

        for line in self.lines:
            # Check for @ imports
            if re.search(r"@[\w/.~-]+\.md", line):
                has_imports = True

            # Check for cross-references (See X.md pattern)
            if re.search(see_pattern, line, re.IGNORECASE):
                has_cross_refs = True

        # If file is large but has no progressive disclosure, suggest it
        if len(self.lines) > 300 and not has_imports and not has_cross_refs:
            self.issues.append(
                (
                    "INFO",
                    "Large file without progressive disclosure. Consider using @imports or cross-references to split content",
                    0,
                )
            )

    def report(self) -> str:
        """Generate validation report."""
        if not self.issues:
            return f"✅ {self.filepath.name}: All checks passed"

        report_lines = [f"\n📄 Validation Report: {self.filepath.name}"]
        report_lines.append("=" * 60)

        errors = [i for i in self.issues if i[0] == "ERROR"]
        warnings = [i for i in self.issues if i[0] == "WARNING"]
        info = [i for i in self.issues if i[0] == "INFO"]

        if errors:
            report_lines.append(f"\n❌ Errors ({len(errors)}):")
            for _, msg, line in errors:
                line_str = f" (line {line})" if line > 0 else ""
                report_lines.append(f"  • {msg}{line_str}")

        if warnings:
            report_lines.append(f"\n⚠️  Warnings ({len(warnings)}):")
            for _, msg, line in warnings:
                line_str = f" (line {line})" if line > 0 else ""
                report_lines.append(f"  • {msg}{line_str}")

        if info:
            report_lines.append(f"\nℹ️  Info ({len(info)}):")
            for _, msg, line in info[:5]:
                line_str = f" (line {line})" if line > 0 else ""
                report_lines.append(f"  • {msg}{line_str}")
            if len(info) > 5:
                report_lines.append(f"  ... and {len(info) - 5} more")

        report_lines.append("\n" + "=" * 60)
        return "\n".join(report_lines)


def find_agents_md_files(root_path: Path) -> List[Path]:
    """Find all AGENTS.md files in directory tree."""
    return sorted(root_path.rglob("AGENTS.md"))


def main():
    if len(sys.argv) < 2:
        print("Usage: python validate_agents_md.py <path>")
        print("  <path> can be a AGENTS.md file or directory to search")
        sys.exit(1)

    path = Path(sys.argv[1])

    if not path.exists():
        print(f"Error: {path} does not exist")
        sys.exit(1)

    files_to_validate = []
    if path.is_file():
        files_to_validate = [path]
    else:
        files_to_validate = find_agents_md_files(path)

    if not files_to_validate:
        print(f"No AGENTS.md files found in {path}")
        sys.exit(0)

    print(f"Found {len(files_to_validate)} AGENTS.md file(s) to validate\n")

    all_passed = True
    for filepath in files_to_validate:
        validator = AGENTSValidator(filepath)
        passed = validator.validate()
        print(validator.report())

        if not passed:
            all_passed = False

    print("\n" + "=" * 60)
    if all_passed:
        print("✅ All AGENTS.md files passed validation")
        sys.exit(0)
    else:
        print("❌ Some AGENTS.md files have errors")
        sys.exit(1)


if __name__ == "__main__":
    main()
