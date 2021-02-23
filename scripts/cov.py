import os
import re
import sys

# read output of (potentially many invocations of) "go test -cover -short",
# e.g., "coverage: 65.7% of statements"
cover_report_lines = sys.stdin.read()

if len(cover_report_lines) == 0:
    sys.exit(2)

# extract min coverage from GITHUB_WORKFLOW, e.g., "60.4%"
min_cover = float(re.findall(r'\d*\.\d+|\d+', os.environ['GITHUB_WORKFLOW'])[0])

for l in cover_report_lines.splitlines():
    cover = float(re.findall(r'\d*\.\d+|\d+', l)[0])
    if cover < min_cover:
        sys.exit(1)

sys.exit(0)
