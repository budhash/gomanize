#!/usr/bin/env python
# Requires Python 3.8+
"""
Repo Tasks: one-file, hierarchy-aware task tracking (TASKS.md) with zero deps.

TASKS.md Document Structure
---------------------------
The file uses H1 headers to organize three main sections:

  # Meta           Schema documentation and metadata
  # Tasks          All features and tasks (the actual work items)
  # Notes          Free-form notes organized by task/feature ID

Under # Tasks, use H2 headers for workflow sections:
  ## Now           Active work (sorted by priority)
  ## Backlog       Future work
  ## Skipped       Deprioritized items

Item Schema
-----------
Items (Features or Tasks) MUST follow this exact single-line format:
  - [ ] (F-0001) [P1] [todo] Feature: Notifications MVP @deps=T-0001 @rel=F-0002
  - [x] (T-0007) [P0] [done] Implement token validation @done=YYYY-MM-DD

Required fields:
  (F-####) or (T-####)                Unique immutable ID (F=feature, T=task)
  [P0|P1|P2|P3]                       Priority (P0=critical, P3=low)
  [todo|doing|done|skipped|deferred]  Status

Checkbox rule:
  done  => [x]
  else  => [ ]

Hierarchy:
  Indent by 2 spaces per level to create parent/child relationships.
  Features are usually top-level; tasks are children; tasks can have subtasks.

Details (optional):
  Under any item, you may add indented bullets WITHOUT an (F-####)/(T-####) id.
  These are "detail lines" and are preserved during subtree moves.

Tags (optional, machine-friendly):
  Machine tags live in a trailing run of @… tokens AFTER the title. A bare
  @key=value written inside the descriptive title/prose is NOT parsed as a tag
  (so a task titled "convert @tags=x to @milestone=x" carries no such tags).
  @deps=F-0002,T-0011   Blocked by / depends on
  @rel=T-0100,F-0003    Related / associated
  @branch=feat/foo      Implementing branch
  @pr=123               Pull request number
  @issue=42             GitHub issue number
  @tags=security,fix    Custom tags (comma-separated)
  @effort=4h            Estimated effort in hours
  @system=sprint-1      System tag (sprint-*, milestone-*, phase-*, epic-*)
  @milestone=m1         Milestone id or alias (first-class; see Milestones)
  @done=YYYY-MM-DD      Completion date (auto-added)

Notes Section
-------------
The # Notes section supports structured notes organized by task/feature ID:

  # Notes

  ## F-0007
  Feature-level notes, design decisions, context...

  ## T-0003
  Task-specific implementation details, gotchas...

Use `show ID --full` to display notes alongside task details.
Task IDs are unique across the file, so notes follow the ID when tasks move.

Shadow Features
---------------
When a Task moves to a different section than its parent Feature (via skip,
start, backlog, now), a lightweight "shadow" copy of the Feature is auto-created
in the target section (marked with @shadow tag). This keeps tasks visible in
their section while preserving Feature grouping.

  - `tree` and `show F-N` display a merged view with [Section] labels
  - `backlog F-N` / `now F-N` collapses all shadows back into one section
  - Empty shadows are auto-cleaned after moves
  - `validate` warns about empty shadows and errors on orphaned shadows

Milestones (optional, opt-in)
-----------------------------
A task may carry ONE milestone via an @milestone=<id-or-alias> tag. It is
strictly optional and orthogonal to @system=/@tags=: a file with no milestone
data behaves exactly as before, and a task without the tag resolves to an
implicit sentinel bucket named "default" (configurable via the
TASKS_MILESTONE_SENTINEL env var; it must not look like an M<n> id). The
sentinel is never written into task lines.

An optional "# Milestones" registry (its own H1 section) maps ids to aliases,
statuses, and titles. When present it powers alias resolution (@milestone=alpha
and @milestone=m1 are the same milestone) and rollups; when absent, milestones
are freeform (any string accepted, grouped by raw value). Example:

  # Milestones

  - M1  alias=alpha  status=active   Complete federal estimate (North-Star)
  - M2  alias=beta   status=planned  Surfaces / API

  new task "x" --milestone m1     Assign at creation
  set T-7 --milestone alpha       Assign/change (alias ok); "", "clear", or the
                                    sentinel name clears back to the sentinel
  list --milestone m1             Filter (alias-resolved; 'default' = unassigned)
  milestone                       Per-milestone rollup (counts, %done, sentinel)
  milestone M1                    Detail: tasks under M1, grouped by status
  milestone --table               milestone x features x status table (features
                                    are attributed directly or via their tasks)
  next --milestone m1             Next actionable task within M1
  migrate-tags-to-milestone m1    Rewrite interim @tags=m1 -> @milestone=m1

Short References
----------------
CLI accepts: F-1, F-001, F-0001; T-7, T-007, T-0007
In text, you can write: F-001/T-0007 etc.

Commands
--------
  init [--force]                     Create TASKS.md template
  help                               Print this help
  version                            Print tool version + canonical source
  selfupdate [--check] [--source X] [--allow-untrusted-source]
                                     Sync this copy to canonical (https-only; a
                                     non-default --source needs the allow flag)
  validate                           Validate TASKS.md (CI-friendly; exits non-zero on failure)
  list [filters]                     List items (--prio --status --tag --section --issue --effort
                                       --system --milestone)
  tree                               Pretty hierarchy view
  show ID [--full]                   Show item with children (--full includes notes from # Notes)
  next [--milestone ID]              Show next actionable task (highest prio, unblocked, in Now)
  new (feature|task) "Title" [...]   Create item (--prio|--priority, --status, --under ID,
                                       --section NAME, --effort HOURS, --tags TAG1,TAG2, --deps ID1,ID2,
                                       --milestone ID)
  mv ID (--section NAME|--under ID)  Move item subtree (auto-sorts Now by priority)
  start|done|skip|defer ID           Status transitions (start enforces single WIP)
  reopen ID                          Reset terminal status (done/skipped/deferred) back to todo
  prio ID P0..P3                     Update priority
  link ID [--deps OP LIST] [--rel OP LIST]
                                     Edit @deps/@rel (OP: add|rm|set|clear)
  set ID --branch NAME               Set @branch= tag
  set ID --pr NUMBER                 Set @pr= tag
  set ID --issue NUMBER              Set @issue= tag
  set ID --tags TAG1,TAG2            Set @tags= (custom labels)
  set ID --effort HOURS              Set @effort= (e.g., 4h, 8h)
  set ID --system TAG                Set @system= (sprint-*, milestone-*, phase-*, epic-*)
  set ID --milestone ID              Set @milestone= (id/alias; "", "clear", or sentinel name clears)
  milestone [ID] [--table]           Milestone rollup; `milestone ID` for detail; `--table` for a
                                       milestone x features x status table
  migrate-tags-to-milestone TAG      Rewrite interim @tags=TAG into @milestone=TAG
  current                            Show current task(s) in [doing] status
  sync ID                            Sync task with linked GitHub issue
  backlog [ID]                       List Backlog items, or move ID to Backlog
  now [ID]                           List Now items, or move ID to Now (promote)
  nextid                             Show next available Feature and Task IDs (for manual editing)

Examples
--------
  ./tools/tasks.py init
  ./tools/tasks.py new feature "Auth MVP" --prio P0 --section Now
  ./tools/tasks.py new task "Token validation" --prio P0 --under F-1 --effort 4h --deps T-3
  ./tools/tasks.py start T-7
  ./tools/tasks.py skip T-8                    # Moves to ## Skipped (creates shadow if needed)
  ./tools/tasks.py reopen T-8                  # Reset status to todo (stays in current section)
  ./tools/tasks.py backlog F-1                 # Collapse all shadows, move Feature to Backlog
  ./tools/tasks.py set T-7 --branch feat/token-validation
  ./tools/tasks.py set T-7 --effort 4h
  ./tools/tasks.py set T-7 --system sprint-1
  ./tools/tasks.py show F-1 --full             # Merged view across sections
  ./tools/tasks.py tree                        # Merged Feature display with section labels
  ./tools/tasks.py list --system sprint-1 --status todo
  ./tools/tasks.py new task "Estimator" --under F-1 --milestone m1
  ./tools/tasks.py set T-7 --milestone alpha    # stored as-is; shows as M1 on read
  ./tools/tasks.py list --milestone m1          # alias-resolved; 'default' = unassigned
  ./tools/tasks.py milestone                    # per-milestone rollup
  ./tools/tasks.py milestone M1                 # one milestone's detail
  ./tools/tasks.py next
  ./tools/tasks.py done T-7
"""

import json
import os
import re
import subprocess
import sys
from datetime import date
from pathlib import Path
from typing import List, Tuple, Optional, Dict, Set, Iterator

__version__ = "1.1.1"

# Canonical source for `selfupdate` — the published raw URL. `selfupdate`
# overwrites this very script with the fetched content, so a NON-default source
# (via --source or the TASKS_CANONICAL_SOURCE env var) is treated as untrusted
# and requires explicit --allow-untrusted-source confirmation.
DEFAULT_CANONICAL_SOURCE = "https://raw.githubusercontent.com/budhash/tasks/main/tasks.py"
CANONICAL_SOURCE = os.environ.get("TASKS_CANONICAL_SOURCE", DEFAULT_CANONICAL_SOURCE)
_MAX_FETCH_BYTES = 5 * 1024 * 1024  # sanity cap; the tool is ~100 KB

# Allow override via environment variable
TASK_FILE = os.environ.get("TASKS_FILE", "TASKS.md")

# Valid values for status and priority
ALLOWED_STATUS = {"todo", "doing", "done", "skipped", "deferred"}
ALLOWED_PRIO = {"P0", "P1", "P2", "P3"}
PRIO_ORDER = {"P0": 0, "P1": 1, "P2": 2, "P3": 3}

# Magic number for sorting non-item lines (keeps them in place)
SORT_ORDER_NON_ITEM = 999

# Item line: (F-0001) or (T-0007)
ITEM_RE = re.compile(
    r'^(\s*)- \[( |x)\] '
    r'\(((?:F|T)-\d{4})\) '
    r'\[(P[0-3])\] '
    r'\[(todo|doing|done|skipped|deferred)\] '
    r'(.*)$'
)

SECTION_RE = re.compile(r'^##\s+(.+?)\s*$')

# Tag patterns - single source of truth for tag names and regexes
TAG_PATTERNS = {
    "deps": re.compile(r'@deps=([^\s]+)'),
    "rel": re.compile(r'@rel=([^\s]+)'),
    "branch": re.compile(r'@branch=([^\s]+)'),
    "pr": re.compile(r'@pr=(\d+)'),
    "issue": re.compile(r'@issue=(\d+)'),
    "tags": re.compile(r'@tags=([^\s]+)'),
    "effort": re.compile(r'@effort=(\d+h?)'),
    "system": re.compile(r'@system=([^\s]+)'),
    "milestone": re.compile(r'@milestone=([^\s]+)'),
    "shadow": re.compile(r'@shadow\b'),
}

# Valid @system= prefixes (system-controlled tags). NOTE: @system=milestone-* is
# a distinct, older namespace; the first-class milestone dimension uses the
# separate @milestone= tag (see the # Milestones registry). They do not overlap.
SYSTEM_PREFIXES = ("sprint-", "milestone-", "phase-", "epic-")

# Milestone support (opt-in; fully backward-compatible).
# A task optionally carries one @milestone=<id-or-alias>; absent => sentinel.
MILESTONE_HEADER = "# Milestones"
# Canonical milestone IDs look like M1, M2, ... (case-insensitive on read).
MILESTONE_ID_RE = re.compile(r'^M\d+$', re.IGNORECASE)

# Top-level document parts (H1 headers: # Meta, # Tasks, # Notes)
META_HEADER = "# Meta"
TASKS_HEADER = "# Tasks"
NOTES_HEADER = "# Notes"

# Task sections - subsections under # Tasks (H2 headers)
TASK_SECTIONS = {"Now", "Backlog", "Skipped"}

# Legacy aliases for backwards compatibility (TODO: migrate usages)
TAG_DEPS_RE = TAG_PATTERNS["deps"]
TAG_REL_RE = TAG_PATTERNS["rel"]
TAG_BRANCH_RE = TAG_PATTERNS["branch"]
TAG_PR_RE = TAG_PATTERNS["pr"]
TAG_ISSUE_RE = TAG_PATTERNS["issue"]
TAG_TAGS_RE = TAG_PATTERNS["tags"]
TAG_EFFORT_RE = TAG_PATTERNS["effort"]
TAG_SYSTEM_RE = TAG_PATTERNS["system"]
TAG_MILESTONE_RE = TAG_PATTERNS["milestone"]
TAG_SHADOW_RE = TAG_PATTERNS["shadow"]


# -------------------------
# Small utilities
# -------------------------
def eprint(*a):
    print(*a, file=sys.stderr)


def iter_content_lines(lines: List[str], skip_non_task_sections: bool = True) -> Iterator[Tuple[int, str]]:
    """
    Yield (index, line) pairs, skipping lines inside fenced code blocks
    and optionally content outside the # Tasks section.
    This is the single source of truth for content iteration.

    Document structure:
      # TASKS.md (title)
      ...intro...
      # Tasks
        ## Now
        ## Backlog
        ## Skipped
      # Notes
        (free format, referential integrity checked separately)
    """
    in_code_block = False
    in_tasks_section = False
    for i, line in enumerate(lines):
        # Track code blocks
        if line.strip().startswith("```"):
            in_code_block = not in_code_block
            continue
        if in_code_block:
            continue

        # Track H1 section transitions (# Tasks, # Notes)
        if skip_non_task_sections:
            if line.startswith("# ") and not line.startswith("# TASKS"):
                stripped = line.rstrip()
                if stripped == TASKS_HEADER:
                    in_tasks_section = True
                elif stripped.startswith("# "):
                    # Any other H1 header ends the Tasks section
                    in_tasks_section = False

            # Skip content outside Tasks section (but yield everything when not skipping)
            if not in_tasks_section:
                # Still yield H2 section headers for find_section_bounds compatibility
                if line.startswith("## "):
                    yield i, line
                continue

        yield i, line


def parse_item(line: str) -> Optional[Tuple[str, ...]]:
    """Parse an item line into its components.

    Returns: (indent, box, id, prio, status, rest) or None if not an item line.
    """
    m = ITEM_RE.match(line)
    return m.groups() if m else None


def build_item_line(indent: str, item_id: str, prio: str, status: str, rest: str) -> str:
    """Build a properly formatted item line from components."""
    box = "x" if status == "done" else " "
    return f"{indent}- [{box}] ({item_id}) [{prio}] [{status}] {rest}\n"


def is_child_content(line: str, parent_indent: int) -> bool:
    """Check if a line is child content (blank or more deeply indented)."""
    if line.strip() == "":
        return True
    # Check if it's an indented detail bullet
    if line.lstrip().startswith("- "):
        cur_indent = len(line) - len(line.lstrip(" "))
        return cur_indent > parent_indent
    return False


def load() -> List[str]:
    path = Path(TASK_FILE)
    if not path.exists():
        raise SystemExit(f"{TASK_FILE} not found. Run: ./tools/tasks.py init")
    return path.read_text(encoding="utf-8").splitlines(keepends=True)


def save(lines: List[str]) -> None:
    Path(TASK_FILE).write_text("".join(lines), encoding="utf-8")


def normalize_id(s: str) -> str:
    """
    Normalize IDs:
      T-7, T-07, T-007, T-0007 -> T-0007
      F-1, F-001 -> F-0001
    Also accepts input wrapped in text like "(T-7)".
    """
    s = s.strip()
    m = re.search(r'([FT])-(\d{1,4})', s)
    if not m:
        raise SystemExit(f"Invalid id '{s}'. Expected like T-7 or F-001.")
    kind, num = m.group(1), m.group(2)
    return f"{kind}-{int(num):04d}"


def parse_id_list(s: str) -> List[str]:
    """
    Comma-separated list like: "F-1,T-7,T-0009"
    Returns canonical IDs, de-duped, sorted (F then T, numeric).
    """
    s = (s or "").strip()
    if not s:
        return []
    parts = [p.strip() for p in s.split(",") if p.strip()]
    ids = [normalize_id(p) for p in parts]
    return sort_ids(unique_preserve(ids))


def unique_preserve(ids: List[str]) -> List[str]:
    seen: Set[str] = set()
    out = []
    for x in ids:
        if x not in seen:
            seen.add(x)
            out.append(x)
    return out


def sort_ids(ids: List[str]) -> List[str]:
    """Sort IDs by type (F before T) then by number."""
    def key(x: str) -> Tuple[int, int]:
        parts = x.split("-")
        if len(parts) != 2:
            raise SystemExit(f"Malformed ID: {x}")
        kind, num_str = parts
        try:
            num = int(num_str)
        except ValueError:
            raise SystemExit(f"Malformed ID number: {x}")
        return (0 if kind == "F" else 1, num)
    return sorted(ids, key=key)


def is_section_header(line: str) -> bool:
    return line.startswith("## ")


def section_name(line: str) -> Optional[str]:
    m = SECTION_RE.match(line)
    return m.group(1).strip() if m else None


def find_section_bounds(lines: List[str], section: str) -> Tuple[Optional[int], int]:
    """Return (start, end) line indices for a section. On miss, start is None
    and end is -1 (callers guard on `start is None` before using end)."""
    target = section.lower()
    start = None
    for i, line in enumerate(lines):
        if is_section_header(line):
            name = section_name(line)
            if name and name.lower() == target:
                start = i
                break
    if start is None:
        return None, -1
    end = len(lines)
    for j in range(start + 1, len(lines)):
        if is_section_header(lines[j]):
            end = j
            break
    return start, end


def find_section_insert_pos(lines: List[str], section: str) -> int:
    start, end = find_section_bounds(lines, section)
    if start is None:
        raise SystemExit(f"Section '{section}' not found (need a '## {section}' heading)")

    last_item = None
    for i in range(start + 1, end):
        if ITEM_RE.match(lines[i]):
            last_item = i

    if last_item is not None:
        pos = last_item + 1
        while pos < end and lines[pos].strip() == "":
            pos += 1
        return pos

    pos = start + 1
    while pos < end and lines[pos].strip() == "":
        pos += 1
    return pos


def ensure_section(lines: List[str], section: str) -> List[str]:
    """Ensure a `## <section>` header exists under `# Tasks`.

    Auto-creates a missing STANDARD section (Now/Backlog/Skipped) in canonical
    order; non-standard names are left alone so typos still error. Idempotent.
    """
    if section not in TASK_SECTIONS:
        return lines
    start, _ = find_section_bounds(lines, section)
    if start is not None:
        return lines

    order = ["Now", "Backlog", "Skipped"]
    later = order[order.index(section) + 1:] if section in order else []

    tasks_idx = next((i for i, ln in enumerate(lines) if ln.rstrip() == TASKS_HEADER), None)
    if tasks_idx is None:
        prefix = [] if (lines and lines[-1].strip() == "") else [""]
        return lines + prefix + [TASKS_HEADER, "", f"## {section}", ""]

    end = next((j for j in range(tasks_idx + 1, len(lines)) if lines[j].startswith("# ")), len(lines))
    insert_at = end
    for j in range(tasks_idx + 1, end):
        if is_section_header(lines[j]) and section_name(lines[j]) in later:
            insert_at = j
            break
    lines[insert_at:insert_at] = [f"## {section}", ""]
    return lines


def find_item_line_index(lines: List[str], item_id: str) -> Tuple[Optional[int], int]:
    """Find an item by ID, skipping content inside fenced code blocks.

    Returns (index, indent_width). On miss, index is None and indent is -1
    (callers guard on `index is None` before using the indent).
    """
    for i, line in iter_content_lines(lines):
        m = ITEM_RE.match(line)
        if m and m.group(3) == item_id:
            return i, len(m.group(1))
    return None, -1


def next_id(lines: List[str], kind: str) -> str:
    """Generate next ID, skipping content inside fenced code blocks."""
    max_n = 0
    for _, line in iter_content_lines(lines):
        m = re.search(r'\(([FT])-(\d{4})\)', line)
        if m and m.group(1) == kind:
            max_n = max(max_n, int(m.group(2)))
    return f"{kind}-{max_n+1:04d}"


def normalize_item_line(line: str, *, status: Optional[str] = None, prio: Optional[str] = None) -> str:
    m = ITEM_RE.match(line)
    if not m:
        return line
    indent, box, iid, cur_prio, cur_status, rest = m.groups()
    new_status = status or cur_status
    new_prio = prio or cur_prio

    new_box = "x" if new_status == "done" else " "
    if new_status == "done" and "@done=" not in rest:
        rest += f" @done={date.today()}"

    return f"{indent}- [{new_box}] ({iid}) [{new_prio}] [{new_status}] {rest}\n"


def find_insert_under(lines: List[str], parent_id: str) -> Tuple[int, int]:
    idx, parent_indent = find_item_line_index(lines, parent_id)
    if idx is None:
        raise SystemExit(f"Parent item {parent_id} not found")

    insert_at = idx + 1
    while insert_at < len(lines):
        if is_section_header(lines[insert_at]):
            break

        parsed = parse_item(lines[insert_at])
        if parsed:
            indent_len = len(parsed[0])
            if indent_len <= parent_indent:
                break
            insert_at += 1
            continue

        # detail lines/blanks: keep scanning if indented deeper than parent
        if is_child_content(lines[insert_at], parent_indent):
            insert_at += 1
            continue

        break

    return insert_at, parent_indent


def extract_block(lines: List[str], item_id: str) -> Tuple[List[str], List[str], int]:
    idx, base_indent = find_item_line_index(lines, item_id)
    if idx is None:
        raise SystemExit(f"Item {item_id} not found")

    block = [lines[idx]]
    j = idx + 1
    while j < len(lines):
        if is_section_header(lines[j]):
            break

        parsed = parse_item(lines[j])
        if parsed:
            indent_len = len(parsed[0])
            if indent_len <= base_indent:
                break
            block.append(lines[j])
            j += 1
            continue

        # keep detail bullets / blanks if indented deeper than base indent
        if is_child_content(lines[j], base_indent):
            block.append(lines[j])
            j += 1
            continue

        break

    new_lines = lines[:idx] + lines[idx + len(block):]
    return new_lines, block, base_indent


def reindent_block(block: List[str], old_base: int, new_base: int) -> List[str]:
    delta = new_base - old_base
    out = []
    for line in block:
        parsed = parse_item(line)
        if parsed:
            cur = len(parsed[0])
            out.append((" " * max(0, cur + delta)) + line.lstrip(" "))
        else:
            if line.strip() == "":
                out.append(line)
            else:
                cur = len(line) - len(line.lstrip(" "))
                if cur > 0:
                    out.append((" " * max(0, cur + delta)) + line.lstrip(" "))
                else:
                    out.append(line)
    return out


def sort_section_top_level(lines: List[str], section: str) -> List[str]:
    start, end = find_section_bounds(lines, section)
    if start is None:
        raise SystemExit(f"Section '{section}' not found (need a '## {section}' heading)")

    region = lines[start + 1:end]
    blocks = []
    i = 0
    while i < len(region):
        parsed = parse_item(region[i])
        if parsed and len(parsed[0]) == 0:
            # top-level block (item + everything until next top-level item)
            block = [region[i]]
            j = i + 1
            while j < len(region):
                parsed2 = parse_item(region[j])
                if parsed2 and len(parsed2[0]) == 0:
                    break
                block.append(region[j])
                j += 1
            prio = parsed[3]
            blocks.append((PRIO_ORDER.get(prio, 99), i, block))
            i = j
        else:
            blocks.append((SORT_ORDER_NON_ITEM, i, [region[i]]))
            i += 1

    top = [b for b in blocks if b[0] != SORT_ORDER_NON_ITEM]
    top_sorted = sorted(top, key=lambda x: (x[0], x[1]))
    top_iter = iter([b[2] for b in top_sorted])

    out = []
    for prio_order, _, blk in blocks:
        if prio_order == SORT_ORDER_NON_ITEM:
            out.extend(blk)
        else:
            out.extend(next(top_iter))

    lines[start + 1:end] = out
    return lines


def get_all_items(lines: List[str]) -> Dict[str, dict]:
    """Parse all items into a dict keyed by ID, skipping code blocks."""
    items: Dict[str, dict] = {}
    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if parsed:
            indent, box, iid, prio, status, rest = parsed
            shadow = is_shadow(line)
            # Skip shadow entries — they duplicate primary IDs
            if shadow:
                continue
            # Extract tags from the trailing tag region only (never from prose).
            items[iid] = {
                "line_num": i,
                "indent": len(indent),
                "prio": prio,
                "status": status,
                "rest": rest,
                "deps": parse_id_list(_get_tag_value(rest, TAG_DEPS_RE) or ""),
                "milestone": _get_tag_value(rest, TAG_MILESTONE_RE),
                "is_shadow": shadow,
            }
    return items


def get_section_for_item(lines: List[str], item_line_num: int) -> Optional[str]:
    """Find which section an item belongs to."""
    current_section = None
    for i, line in enumerate(lines):
        if is_section_header(line):
            current_section = section_name(line)
        if i == item_line_num:
            return current_section
    return None


# -------------------------
# Shadow feature support
# -------------------------
def is_shadow(line: str) -> bool:
    """Check if an item line is a shadow feature (contains @shadow tag)."""
    return bool(TAG_SHADOW_RE.search(line))


def find_parent_feature(lines: List[str], item_line_idx: int) -> Optional[str]:
    """Scan upward from item_line_idx to find the nearest parent Feature (F-####).

    Returns the Feature ID, or None if the item is top-level.
    """
    parsed = parse_item(lines[item_line_idx])
    if not parsed:
        return None
    item_indent = len(parsed[0])
    if item_indent == 0:
        return None  # top-level item, no parent

    for i in range(item_line_idx - 1, -1, -1):
        if is_section_header(lines[i]):
            return None  # hit section boundary without finding parent
        p = parse_item(lines[i])
        if p:
            p_indent = len(p[0])
            if p_indent < item_indent and p[2].startswith("F-"):
                return p[2]
            if p_indent < item_indent and p[2].startswith("T-"):
                # parent is a Task, keep scanning for Feature ancestor
                item_indent = p_indent
    return None


def find_primary_line(lines: List[str], feature_id: str) -> Optional[int]:
    """Find the line index of the primary (non-shadow) Feature.

    Returns None if no primary exists (only shadows or not found).
    """
    feature_id = normalize_id(feature_id)
    for i, line in iter_content_lines(lines):
        p = parse_item(line)
        if p and p[2] == feature_id and not is_shadow(line):
            return i
    return None


def find_shadow_line(lines: List[str], feature_id: str, target_section: str) -> Optional[int]:
    """Find the line index of a shadow Feature in a specific section.

    Returns None if no shadow exists in that section.
    """
    feature_id = normalize_id(feature_id)
    sec_start, sec_end = find_section_bounds(lines, target_section)
    if sec_start is None:
        return None
    for i in range(sec_start + 1, sec_end or len(lines)):
        p = parse_item(lines[i])
        if p and p[2] == feature_id and is_shadow(lines[i]):
            return i
    return None


def get_all_shadows(lines: List[str], feature_id: str) -> List[Tuple[int, str]]:
    """Return list of (line_idx, section_name) for all shadows of a Feature."""
    feature_id = normalize_id(feature_id)
    shadows = []
    current_section = None
    for i, line in iter_content_lines(lines):
        if is_section_header(line):
            current_section = section_name(line)
            continue
        p = parse_item(line)
        if p and p[2] == feature_id and is_shadow(line):
            shadows.append((i, current_section))
    return shadows


def get_or_create_shadow(lines: List[str], feature_id: str, target_section: str) -> Tuple[List[str], int]:
    """Ensure a shadow of feature_id exists in target_section.

    If shadow already exists, returns (lines, shadow_line_idx).
    If not, copies the primary Feature line with @shadow appended, inserts at top of section.
    Raises SystemExit if no primary Feature found.
    """
    feature_id = normalize_id(feature_id)

    # Check if shadow already exists
    existing = find_shadow_line(lines, feature_id, target_section)
    if existing is not None:
        return lines, existing

    # Find primary to copy from
    primary_idx = find_primary_line(lines, feature_id)
    if primary_idx is None:
        raise SystemExit(f"{feature_id} has no primary definition. Run 'validate' to diagnose.")

    # Build shadow line from primary
    p = parse_item(lines[primary_idx])
    assert p is not None  # primary_idx is a known item line
    indent, _, iid, prio, status, rest = p
    # Strip any existing @shadow (shouldn't be there, but defensive)
    rest_clean = re.sub(r'\s*@shadow\b', '', rest).strip()
    shadow_rest = f"{rest_clean} @shadow" if rest_clean else "@shadow"
    shadow_line = build_item_line("", iid, prio, status, shadow_rest)

    # Insert right after section header, skipping blank lines but stopping at boundaries
    lines = ensure_section(lines, target_section)  # auto-create a missing standard section
    sec_start, _ = find_section_bounds(lines, target_section)
    assert sec_start is not None  # ensure_section guarantees a standard section exists
    insert_at = sec_start + 1
    while insert_at < len(lines):
        ln = lines[insert_at]
        if ln.strip() == "":
            insert_at += 1
            continue
        # Stop at boundaries — don't go past --- or # headers
        if ln.strip() == "---" or ln.startswith("# "):
            break
        break  # found non-blank content, insert before it

    lines.insert(insert_at, shadow_line)
    return lines, insert_at


def cleanup_empty_shadows(lines: List[str], feature_id: str) -> List[str]:
    """Remove shadow Features that have no item children.

    Scans all shadows for the given Feature. If a shadow has no child items
    (items indented deeper directly below it), removes the shadow line.
    """
    if not feature_id:
        return lines

    feature_id = normalize_id(feature_id)
    to_remove = []

    for shadow_idx, _ in get_all_shadows(lines, feature_id):
        # Check if shadow has any item children
        has_children = False
        for j in range(shadow_idx + 1, len(lines)):
            if is_section_header(lines[j]):
                break
            p = parse_item(lines[j])
            if p:
                if len(p[0]) > 0:  # indented = child
                    has_children = True
                    break
                else:  # same level = sibling, not child
                    break
            # blank or detail lines — keep scanning
            if lines[j].strip() != "" and not is_child_content(lines[j], 0):
                break
        if not has_children:
            to_remove.append(shadow_idx)

    # Remove in reverse order to preserve indices
    for idx in sorted(to_remove, reverse=True):
        lines.pop(idx)

    return lines


def collapse_feature(lines: List[str], feature_id: str, target_section: str) -> List[str]:
    """Unify a Feature and all its scattered children into target_section.

    Collects children from the primary Feature and all shadows, removes shadows,
    and inserts the merged Feature block into the target section.
    """
    feature_id = normalize_id(feature_id)

    primary_idx = find_primary_line(lines, feature_id)
    if primary_idx is None:
        raise SystemExit(f"{feature_id} has no primary definition. Run 'validate' to diagnose.")

    # Collect children from all shadows (extract children, discard shadow lines)
    shadow_children: List[str] = []
    shadows = get_all_shadows(lines, feature_id)
    # Process in reverse order to preserve indices during removal
    for shadow_idx, _ in sorted(shadows, key=lambda x: x[0], reverse=True):
        # Extract children of this shadow
        j = shadow_idx + 1
        children_block = []
        while j < len(lines):
            if is_section_header(lines[j]):
                break
            p = parse_item(lines[j])
            if p and len(p[0]) == 0:  # top-level item = sibling, stop
                break
            if p and len(p[0]) > 0:  # indented child
                children_block.append(lines[j])
                j += 1
                continue
            if is_child_content(lines[j], 0):
                children_block.append(lines[j])
                j += 1
                continue
            break
        shadow_children = children_block + shadow_children  # preserve order
        # Remove shadow + its children
        del lines[shadow_idx:shadow_idx + 1 + len(children_block)]

    # Now extract the primary block
    lines, primary_block, old_base = extract_block(lines, feature_id)

    # Append shadow children to primary block
    primary_block.extend(shadow_children)

    # Insert merged block into target section
    pos = find_section_insert_pos(lines, target_section)
    if pos > 0 and lines[pos - 1].strip() != "":
        lines.insert(pos, "\n")
        pos += 1
    # Reindent to top-level
    adjusted = reindent_block(primary_block, old_base, 0)
    for k, ln in enumerate(adjusted):
        lines.insert(pos + k, ln)

    if target_section.lower() == "now":
        lines = sort_section_top_level(lines, "Now")

    return lines


def move_item_to_section(lines: List[str], item_id: str, target_section: str,
                         new_status: Optional[str] = None,
                         enforce_single_wip: bool = False) -> List[str]:
    """Move an item (and its subtree) to a different section.

    Handles shadow creation/cleanup for Tasks under Features.
    For Features, collapses all shadows into the target section.

    Args:
        lines: Current file lines
        item_id: The item to move
        target_section: Target section name (Now, Backlog, Skipped)
        new_status: Optional new status to apply
        enforce_single_wip: If True and new_status is 'doing', reset other WIP items
    """
    item_id = normalize_id(item_id)

    # Auto-create the target section if it's a missing standard section, so the
    # move (and any shadow handling) never targets a non-existent section.
    lines = ensure_section(lines, target_section)

    # Find current position and section
    idx, _ = find_item_line_index(lines, item_id)
    if idx is None:
        raise SystemExit(f"Item {item_id} not found")
    current_section = get_section_for_item(lines, idx)

    is_feature = item_id.startswith("F-")

    # Feature-level move: collapse all shadows
    if is_feature:
        lines = collapse_feature(lines, item_id, target_section)
        # Apply status change if requested
        if new_status:
            new_idx, _ = find_item_line_index(lines, item_id)
            if new_idx is not None:
                lines[new_idx] = normalize_item_line(lines[new_idx], status=new_status)
        save(lines)
        print(f"Moved {item_id} to section '{target_section}'")
        return lines

    # Task-level move
    # Find parent Feature before extracting (position changes after extract)
    parent_feature_id = find_parent_feature(lines, idx)

    # Extract the task block
    lines, block, old_base = extract_block(lines, item_id)

    # Apply status change to block header
    if new_status:
        block[0] = normalize_item_line(block[0], status=new_status)

    # Determine where to insert
    if parent_feature_id:
        # Check if parent Feature (primary or shadow) exists in target section
        primary_in_target = False
        p_primary = find_primary_line(lines, parent_feature_id)
        if p_primary is not None:
            p_section = get_section_for_item(lines, p_primary)
            if p_section and p_section.lower() == target_section.lower():
                primary_in_target = True

        if primary_in_target:
            # Insert under primary Feature in target
            insert_at, parent_indent = find_insert_under(lines, parent_feature_id)
            new_base = parent_indent + 2
        else:
            # Need a shadow in target section
            lines, shadow_idx = get_or_create_shadow(lines, parent_feature_id, target_section)
            insert_at, parent_indent = find_insert_under(lines, parent_feature_id)
            # find_insert_under may find primary; we need shadow's children position
            # Recalculate under the shadow specifically
            shadow_line = find_shadow_line(lines, parent_feature_id, target_section)
            if shadow_line is not None:
                j = shadow_line + 1
                while j < len(lines):
                    ln = lines[j]
                    # Stop at section boundaries
                    if is_section_header(ln) or ln.startswith("# ") or ln.strip() == "---":
                        break
                    p = parse_item(ln)
                    if p and len(p[0]) == 0:  # top-level sibling
                        break
                    if p and len(p[0]) > 0:  # child item — keep scanning past it
                        j += 1
                        continue
                    if is_child_content(ln, 0):  # detail/blank line
                        j += 1
                        continue
                    break
                insert_at = j
                parent_indent = 0  # shadow is top-level
            new_base = parent_indent + 2
    else:
        # Top-level task (no parent Feature)
        pos = find_section_insert_pos(lines, target_section)
        insert_at = pos
        new_base = 0

    # Reindent and insert
    adjusted = reindent_block(block, old_base, new_base)
    for k, ln in enumerate(adjusted):
        lines.insert(insert_at + k, ln)

    # Enforce single WIP if needed
    if enforce_single_wip and new_status == "doing":
        for i, line in iter_content_lines(lines):
            p = parse_item(line)
            if p and p[2] != item_id and p[4] == "doing":
                lines[i] = normalize_item_line(line, status="todo")

    # Cleanup empty shadows on the old parent
    if parent_feature_id:
        lines = cleanup_empty_shadows(lines, parent_feature_id)

    # Sort Now section if that's the target
    if target_section.lower() == "now":
        lines = sort_section_top_level(lines, "Now")

    save(lines)
    print(f"Moved {item_id} to section '{target_section}'")
    return lines


# -------------------------
# Notes extraction
# -------------------------
def get_notes_for_id(lines: List[str], item_id: str) -> Optional[str]:
    """Extract notes for a specific task/feature ID from the # Notes section.

    Notes are organized with H2 headers like:
      ## F-0007
      Feature-level notes here...

      ## T-0003
      Task-specific notes here...

    Returns the notes content (without the header) or None if not found.
    """
    item_id = normalize_id(item_id)

    # Find the Notes section boundaries
    notes_start = None
    notes_end = len(lines)
    for i, line in enumerate(lines):
        if line.rstrip() == NOTES_HEADER:
            notes_start = i
        elif notes_start is not None and line.startswith("# "):
            notes_end = i
            break

    if notes_start is None:
        return None

    # Look for ## {item_id} header within Notes section
    id_header = f"## {item_id}"
    header_idx = None
    for i in range(notes_start + 1, notes_end):
        if lines[i].rstrip() == id_header:
            header_idx = i
            break

    if header_idx is None:
        return None

    # Collect content until next H2 header or end of Notes section
    content_lines = []
    for i in range(header_idx + 1, notes_end):
        line = lines[i]
        # Stop at next H2 header
        if line.startswith("## "):
            break
        content_lines.append(line.rstrip())

    # Strip leading/trailing empty lines
    while content_lines and content_lines[0] == "":
        content_lines.pop(0)
    while content_lines and content_lines[-1] == "":
        content_lines.pop()

    return "\n".join(content_lines) if content_lines else None


# -------------------------
# Milestone support
# -------------------------
def milestone_sentinel() -> str:
    """The implicit milestone bucket for tasks with no @milestone= tag.

    Defaults to 'default' (configurable via TASKS_MILESTONE_SENTINEL to e.g.
    'none'/'backlog'). It must NOT match the M<n> ID pattern, otherwise
    `set T-N --milestone m1` would be ambiguous between "assign to M1" and
    "clear to the sentinel".
    """
    s = (os.environ.get("TASKS_MILESTONE_SENTINEL", "default") or "").strip() or "default"
    if MILESTONE_ID_RE.match(s):
        raise SystemExit(
            f"Invalid milestone sentinel '{s}': must not match milestone ID pattern M<n> "
            f"(set TASKS_MILESTONE_SENTINEL to a non-'M<n>' value like 'default').")
    return s


def _parse_milestone_line(text: str) -> Optional[dict]:
    """Parse one registry entry body (without the leading '- ').

    Format: `M1  alias=alpha  status=active   Free-text title`
    alias= and status= are optional; the remaining words form the title.
    Returns None if the first token is not an M<n> ID.
    """
    parts = text.split()
    if not parts or not MILESTONE_ID_RE.match(parts[0]):
        return None
    alias = None
    status = None
    title_tokens: List[str] = []
    for tok in parts[1:]:
        low = tok.lower()
        if low.startswith("alias="):
            alias = tok.split("=", 1)[1] or None
        elif low.startswith("status="):
            status = tok.split("=", 1)[1] or None
        else:
            title_tokens.append(tok)
    return {
        "id": parts[0].upper(),
        "alias": alias,
        "status": status,
        "title": " ".join(title_tokens),
    }


def parse_milestone_registry(lines: List[str]) -> Tuple[Dict[str, dict], Dict[str, str]]:
    """Parse the optional `# Milestones` registry section.

    Returns (registry, alias_map):
      registry:  canonical-ID (upper, e.g. 'M1') -> {id, alias, status, title}
      alias_map: lowercased ID and alias -> canonical-ID
    Both are empty when there is no registry (freeform mode) — this is what keeps
    milestones backward-compatible and opt-in.
    """
    registry: Dict[str, dict] = {}
    alias_map: Dict[str, str] = {}

    start = None
    for i, line in enumerate(lines):
        if line.rstrip() == MILESTONE_HEADER:
            start = i
            break
    if start is None:
        return registry, alias_map

    end = len(lines)
    for j in range(start + 1, len(lines)):
        if lines[j].startswith("# ") and lines[j].rstrip() != MILESTONE_HEADER:
            end = j
            break

    in_code = False
    for j in range(start + 1, end):
        s = lines[j].strip()
        if s.startswith("```"):
            in_code = not in_code
            continue
        if in_code or not s.startswith("- "):
            continue
        entry = _parse_milestone_line(s[2:].strip())
        if not entry:
            continue
        cid = entry["id"]
        registry[cid] = entry
        alias_map[cid.lower()] = cid
        if entry["alias"]:
            alias_map[entry["alias"].lower()] = cid
    return registry, alias_map


def resolve_milestone_key(raw: Optional[str], alias_map: Dict[str, str], sentinel: str) -> str:
    """Resolve a raw @milestone= value (or filter argument) to its canonical key.

    - None / empty / the sentinel name  -> the sentinel bucket.
    - With a registry: an ID or alias   -> canonical uppercase ID (e.g. 'M1').
    - Otherwise (freeform / unknown)     -> the raw value, unchanged.
    """
    if raw is None:
        return sentinel
    raw = raw.strip()
    if raw == "" or raw.lower() == sentinel.lower():
        return sentinel
    if alias_map:
        canon = alias_map.get(raw.lower())
        if canon:
            return canon
    return raw


def milestone_label(key: str, registry: Dict[str, dict], sentinel: str) -> str:
    """Human-readable label for a resolved milestone key, e.g. 'M1 (alpha)'."""
    entry = registry.get(key)
    if entry and entry.get("alias"):
        return f"{key} ({entry['alias']})"
    return key


def milestone_registry_issues(lines: List[str]) -> List[str]:
    """Warnings about the `# Milestones` registry's own health (not task values).

    parse_milestone_registry is last-wins and silently drops unparseable lines,
    so a duplicate id, a colliding alias, or a typo'd id would otherwise resolve
    tasks to the wrong milestone with no signal. `validate` surfaces these.
    """
    issues: List[str] = []
    start = None
    for i, line in enumerate(lines):
        if line.rstrip() == MILESTONE_HEADER:
            start = i
            break
    if start is None:
        return issues
    end = len(lines)
    for j in range(start + 1, len(lines)):
        if lines[j].startswith("# ") and lines[j].rstrip() != MILESTONE_HEADER:
            end = j
            break

    seen_ids: Dict[str, int] = {}          # canonical id -> first line
    alias_owner: Dict[str, Tuple[str, int]] = {}  # alias.lower() -> (id, line)
    in_code = False
    for j in range(start + 1, end):
        s = lines[j].strip()
        if s.startswith("```"):
            in_code = not in_code
            continue
        if in_code or not s.startswith("- "):
            continue
        body = s[2:].strip()
        entry = _parse_milestone_line(body)
        if not entry:
            # Looks intended (carries alias=/status=) but failed to parse.
            if "alias=" in body.lower() or "status=" in body.lower():
                issues.append(f"{TASK_FILE}:{j + 1}: unparseable milestone registry line (id must match M<n>): {body!r}")
            continue
        cid = entry["id"]
        if cid in seen_ids:
            issues.append(f"{TASK_FILE}:{j + 1}: duplicate milestone id {cid} (first at line {seen_ids[cid]})")
        else:
            seen_ids[cid] = j + 1
        if entry["alias"]:
            al = entry["alias"].lower()
            if al in alias_owner and alias_owner[al][0] != cid:
                issues.append(f"{TASK_FILE}:{j + 1}: alias '{entry['alias']}' already used by {alias_owner[al][0]} (line {alias_owner[al][1]})")
            else:
                alias_owner.setdefault(al, (cid, j + 1))
            if MILESTONE_ID_RE.match(entry["alias"]) and entry["alias"].upper() != cid:
                issues.append(f"{TASK_FILE}:{j + 1}: alias '{entry['alias']}' looks like a different milestone id")
    return issues


def _warn_unknown_milestone(value: str, lines: Optional[List[str]] = None) -> None:
    """Non-fatal nudge when assigning a milestone absent from an existing registry.

    Assignment still succeeds (freeform is allowed by design); this just prevents
    a typo like `--milestone m11` from silently dropping out of `--milestone m1`.
    """
    registry, alias_map = parse_milestone_registry(lines if lines is not None else load())
    if registry and value and value.lower() != milestone_sentinel().lower() and value.lower() not in alias_map:
        eprint(f"warning: '{value}' is not in the # Milestones registry (assigned as freeform).")


# -------------------------
# Tag editing (@deps / @rel / @branch / @pr)
# -------------------------
# Machine tags live in a trailing run of `@…` tokens after the title. A bare
# `@key=value` mentioned in the descriptive title/prose is NOT a tag — so tags
# are parsed from (and edited in) only that trailing region, never free text.
_KNOWN_MACHINE_TAGS = ("deps", "rel", "branch", "pr", "issue", "tags",
                       "effort", "system", "milestone", "done")
_TAG_TOKEN_RE = re.compile(r'^@(?:' + "|".join(_KNOWN_MACHINE_TAGS) + r')=\S+$')


def _is_tag_token(tok: str) -> bool:
    """True if `tok` is a recognized machine-tag token (`@shadow` or `@name=value`)."""
    return tok == "@shadow" or bool(_TAG_TOKEN_RE.match(tok))


def split_item_tags(rest: str) -> Tuple[str, List[str]]:
    """Split an item's `rest` into (title, trailing_tag_tokens).

    Only the trailing contiguous run of recognized `@…` tokens counts as tags;
    a `@key=value` that appears earlier (in the descriptive title) stays in the
    title. Whitespace is normalized to single spaces, matching how the tool
    already writes tags — so tool-managed lines are unaffected.
    """
    tokens = rest.split()
    i = len(tokens)
    while i > 0 and _is_tag_token(tokens[i - 1]):
        i -= 1
    return " ".join(tokens[:i]), tokens[i:]


def _tag_token_is(tok: str, tag: str) -> bool:
    return tok == f"@{tag}" or tok.startswith(f"@{tag}=")


def _join_title_tags(title: str, tag_tokens: List[str]) -> str:
    return " ".join(([title] if title else []) + tag_tokens)


def _get_tag_value(rest: str, tag_re: re.Pattern) -> Optional[str]:
    # Search only the trailing tag region, so prose mentions aren't read as tags.
    _title, tag_tokens = split_item_tags(rest)
    m = tag_re.search(" ".join(tag_tokens))
    return m.group(1) if m else None


def _set_or_remove_tag(rest: str, tag: str, values: List[str]) -> str:
    title, tag_tokens = split_item_tags(rest)
    tag_tokens = [t for t in tag_tokens if not _tag_token_is(t, tag)]
    if values:
        tag_tokens.append(f"@{tag}=" + ",".join(values))
    return _join_title_tags(title, tag_tokens)


def _set_single_tag(rest: str, tag: str, value: Optional[str]) -> str:
    """Set a single-value tag like @branch or @pr (only within the tag region)."""
    title, tag_tokens = split_item_tags(rest)
    tag_tokens = [t for t in tag_tokens if not _tag_token_is(t, tag)]
    if value:
        tag_tokens.append(f"@{tag}={value}")
    return _join_title_tags(title, tag_tokens)


def edit_links(item_id: str,
               deps_op: Optional[str] = None, deps_list: Optional[str] = None,
               rel_op: Optional[str] = None, rel_list: Optional[str] = None) -> None:
    item_id = normalize_id(item_id)
    lines = load()

    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if not parsed:
            continue
        indent, box, iid, prio, status, rest = parsed
        if iid != item_id:
            continue

        # current
        cur_deps = parse_id_list(_get_tag_value(rest, TAG_DEPS_RE) or "")
        cur_rel = parse_id_list(_get_tag_value(rest, TAG_REL_RE) or "")

        def apply(op: Optional[str], cur: List[str], arg: Optional[str]) -> List[str]:
            if not op:
                return cur
            op = op.lower()
            if op == "clear":
                return []
            if op == "set":
                return parse_id_list(arg or "")
            if op == "add":
                return sort_ids(unique_preserve(cur + parse_id_list(arg or "")))
            if op == "rm":
                rm = set(parse_id_list(arg or ""))
                return [x for x in cur if x not in rm]
            raise SystemExit(f"Unknown link op '{op}' (use add|rm|set|clear)")

        new_deps = apply(deps_op, cur_deps, deps_list)
        new_rel = apply(rel_op, cur_rel, rel_list)

        new_rest = rest
        if deps_op:
            new_rest = _set_or_remove_tag(new_rest, "deps", new_deps)
        if rel_op:
            new_rest = _set_or_remove_tag(new_rest, "rel", new_rel)

        lines[i] = build_item_line(indent, iid, prio, status, new_rest)
        save(lines)
        print(f"Linked {item_id}")
        return

    raise SystemExit(f"Item {item_id} not found")


def set_tag(item_id: str, tag: str, value: str) -> None:
    """Set a single-value tag on an item."""
    item_id = normalize_id(item_id)
    lines = load()

    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if not parsed:
            continue
        indent, box, iid, prio, status, rest = parsed
        if iid != item_id:
            continue

        new_rest = _set_single_tag(rest, tag, value)
        lines[i] = build_item_line(indent, iid, prio, status, new_rest)
        save(lines)
        print(f"Set @{tag}={value} on {item_id}")
        return

    raise SystemExit(f"Item {item_id} not found")


# -------------------------
# Commands
# -------------------------
def cmd_help():
    print(__doc__.strip())


def cmd_init(force: bool = False):
    path = Path(TASK_FILE)
    if path.exists():
        if not force:
            raise SystemExit(f"{TASK_FILE} already exists. Use: init --force")
        # SAFETY: Create backup before overwriting
        backup_path = Path(f"{TASK_FILE}.backup")
        backup_path.write_text(path.read_text(encoding="utf-8"), encoding="utf-8")
        eprint(f"⚠️  Backed up existing file to {TASK_FILE}.backup")

    template = f"""# TASKS.md

Single source of truth for features + tasks in this repo.

Use `./tools/tasks.py help` for CLI commands.

---

# Meta

## Info

This file is managed by `./tools/tasks.py` CLI tool.

- Run `./tools/tasks.py help` for full documentation
- Manual edits may break parsing - use CLI commands instead
- Task IDs (F-####, T-####) are auto-generated and must be unique
- Checkbox state must match status: `[x]` for done, `[ ]` otherwise

## Schema

Format: `- [ ] (ID) [PRIO] [STATUS] Title @tags...`

Example:
```
- [ ] (F-0001) [P0] [todo] Feature title @issue=42 @tags=security,mvp
  - [ ] (T-0001) [P0] [todo] Subtask @deps=T-0002 @effort=4h
  - [x] (T-0002) [P0] [done] Another subtask @done=2025-01-01
```

Tags: `@deps=` `@rel=` `@branch=` `@pr=` `@issue=` `@tags=` `@effort=` `@system=` `@done=`

---

# Tasks

## Now

## Backlog

## Skipped

---

# Notes

Use `## <ID>` headers (e.g., `## F-0001`) for structured notes per task.
Use `./tools/tasks.py show <id> --full` to display notes with task details.

- {date.today()}: Initialized TASKS.md
"""
    path.write_text(template, encoding="utf-8")
    print(f"Initialized {TASK_FILE}")


def cmd_validate():
    lines = load()
    errors = []
    warnings = []
    ids: Set[str] = set()
    # Track primary and shadow locations for Feature IDs
    primary_features: Dict[str, int] = {}  # F-ID -> line number
    shadow_features: Dict[str, List[int]] = {}  # F-ID -> [line numbers]
    doing_count = 0

    # Milestone registry (optional): powers the "unknown milestone" WARNING.
    registry, alias_map = parse_milestone_registry(lines)
    sentinel = milestone_sentinel()
    # Registry-internal health (duplicate ids, alias collisions) — warnings only.
    warnings.extend(milestone_registry_issues(lines))

    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if not parsed:
            continue
        indent, box, iid, prio, status, rest = parsed
        lineno = i + 1

        # Warn (never error) on an @milestone= value not in the registry.
        if registry:
            mval = _get_tag_value(rest, TAG_MILESTONE_RE)
            if mval and mval.lower() != sentinel.lower() and mval.lower() not in alias_map:
                warnings.append(
                    f"{TASK_FILE}:{lineno}: unknown milestone '{mval}' (not in # Milestones registry)")

        # Shadow-aware duplicate check for Features
        if iid.startswith("F-") and is_shadow(line):
            shadow_features.setdefault(iid, []).append(lineno)
            # Don't add to ids set — shadows are expected duplicates
        elif iid.startswith("F-") and iid in primary_features:
            errors.append(f"{TASK_FILE}:{lineno}: multiple primaries for {iid} (first at line {primary_features[iid]})")
        else:
            if iid in ids and not iid.startswith("F-"):
                errors.append(f"{TASK_FILE}:{lineno}: duplicate id {iid}")
            if iid.startswith("F-"):
                primary_features[iid] = lineno
        ids.add(iid)

        if prio not in ALLOWED_PRIO:
            errors.append(f"{TASK_FILE}:{lineno}: invalid priority {prio}")
        if status not in ALLOWED_STATUS:
            errors.append(f"{TASK_FILE}:{lineno}: invalid status {status}")

        if len(indent) % 2 != 0:
            errors.append(f"{TASK_FILE}:{lineno}: indentation must be multiple of 2 spaces")

        if status == "done" and box != "x":
            errors.append(f"{TASK_FILE}:{lineno}: status done must have [x]")
        if status != "done" and box == "x":
            errors.append(f"{TASK_FILE}:{lineno}: [x] checkbox only allowed when status is done")

        if status == "doing":
            doing_count += 1

        if "\r" in line:
            errors.append(f"{TASK_FILE}:{lineno}: contains CR characters (Windows line endings?)")

    for sec in ("Now", "Backlog"):
        start, _ = find_section_bounds(lines, sec)
        if start is None:
            errors.append(f"{TASK_FILE}: missing recommended section '## {sec}'")

    if doing_count > 1:
        errors.append(f"{TASK_FILE}: more than one item is [doing] ({doing_count} found). Use 'start' to enforce single WIP.")

    # Shadow-specific validations
    for fid, shadow_lines in shadow_features.items():
        if fid not in primary_features:
            for sl in shadow_lines:
                errors.append(f"{TASK_FILE}:{sl}: shadow {fid} has no primary definition")
        else:
            # Check for empty shadows
            for shadow_ln in shadow_lines:
                # Find by 0-indexed line
                shadow_idx = shadow_ln - 1
                has_children = False
                for j in range(shadow_idx + 1, len(lines)):
                    if is_section_header(lines[j]) or lines[j].strip() == "---" or lines[j].startswith("# "):
                        break
                    p = parse_item(lines[j])
                    if p and len(p[0]) > 0:
                        has_children = True
                        break
                    if p and len(p[0]) == 0:
                        break
                if not has_children:
                    warnings.append(f"{TASK_FILE}:{shadow_ln}: empty shadow for {fid} (no children)")

    # Validate task ID references in Notes section (# Notes H1 header)
    notes_start = None
    notes_end = len(lines)
    for i, line in enumerate(lines):
        if line.rstrip() == NOTES_HEADER:
            notes_start = i
        elif notes_start is not None and line.startswith("# "):
            notes_end = i
            break

    if notes_start is not None:
        in_code_block = False
        for i in range(notes_start + 1, notes_end):
            line = lines[i]
            # Skip code blocks
            if line.strip().startswith("```"):
                in_code_block = not in_code_block
                continue
            if in_code_block:
                continue
            # Find all task/feature ID references in the line
            # Skip IDs that are inside backticks (documentation examples)
            for match in re.finditer(r'\b([FT]-\d{1,4})\b', line):
                # Check if this match is inside backticks
                start_pos = match.start()
                before = line[:start_pos]
                # Count backticks before this position - odd count means we're inside backticks
                if before.count('`') % 2 == 1:
                    continue
                ref_id = normalize_id(match.group(1))
                if ref_id not in ids:
                    errors.append(f"{TASK_FILE}:{i + 1}: reference to unknown id {ref_id} in Notes section")

    if warnings:
        print("WARNINGS:")
        for w in warnings:
            print(" - " + w)
        print()

    if errors:
        print("VALIDATION FAILED:")
        for e in errors:
            print(" - " + e)
        print()
        print("Hint: Use './tools/tasks.py nextid' to find next available IDs")
        sys.exit(2)
    print("OK: TASKS.md validation passed")
    # Show next available IDs as helpful hint
    next_f = next_id(lines, "F")
    next_t = next_id(lines, "T")
    print(f"     Next IDs: {next_f}, {next_t}")


def cmd_list(args: List[str] = None):
    """List items with optional filtering."""
    args = args or []
    filters = {
        "prio": None,      # P0, P1, P2, P3
        "status": None,    # todo, doing, done, skipped, deferred
        "tag": None,       # filter by @tags containing value
        "section": None,   # Now, Backlog, Skipped
        "issue": None,     # has @issue= tag (any or specific number)
        "effort": None,    # has @effort= tag (any or specific value like 4h)
        "system": None,    # has @system= tag (any or exact match like sprint-1)
        "milestone": None, # milestone id/alias (or 'default' for the sentinel)
    }

    # Parse filter arguments
    i = 0
    while i < len(args):
        if args[i] in ("--prio", "--priority") and i + 1 < len(args):
            filters["prio"] = args[i + 1].upper()
            i += 2
        elif args[i] == "--status" and i + 1 < len(args):
            filters["status"] = args[i + 1].lower()
            i += 2
        elif args[i] == "--tag" and i + 1 < len(args):
            filters["tag"] = args[i + 1].lower()
            i += 2
        elif args[i] == "--section" and i + 1 < len(args):
            filters["section"] = args[i + 1]
            i += 2
        elif args[i] == "--issue":
            if i + 1 < len(args) and args[i + 1].isdigit():
                filters["issue"] = args[i + 1]
                i += 2
            else:
                filters["issue"] = "*"  # Any issue
                i += 1
        elif args[i] == "--effort":
            if i + 1 < len(args) and re.match(r'^\d+h?$', args[i + 1]):
                filters["effort"] = args[i + 1]
                i += 2
            else:
                filters["effort"] = "*"  # Any effort
                i += 1
        elif args[i] == "--system":
            if i + 1 < len(args) and not args[i + 1].startswith("--"):
                filters["system"] = args[i + 1]
                i += 2
            else:
                filters["system"] = "*"  # Any system tag
                i += 1
        elif args[i] == "--milestone":
            if i + 1 >= len(args):
                raise SystemExit("--milestone requires a value (id/alias, or 'default' for unassigned)")
            filters["milestone"] = args[i + 1]
            i += 2
        else:
            raise SystemExit(f"Unknown filter: {args[i]}. Use --prio, --status, --tag, --section, --issue, --effort, --system, --milestone")

    lines = load()
    current_section = None

    # Resolve the milestone filter up front (registry powers alias resolution).
    milestone_filter_key = None
    alias_map: Dict[str, str] = {}
    sentinel = ""
    if filters["milestone"]:
        _registry, alias_map = parse_milestone_registry(lines)
        sentinel = milestone_sentinel()
        milestone_filter_key = resolve_milestone_key(filters["milestone"], alias_map, sentinel)

    for idx, line in iter_content_lines(lines):
        if is_section_header(line):
            current_section = section_name(line)
            continue

        parsed = parse_item(line)
        if not parsed:
            continue

        indent, box, iid, prio, status, rest = parsed

        # Apply filters
        if filters["prio"] and prio != filters["prio"]:
            continue
        if filters["status"] and status != filters["status"]:
            continue
        if filters["section"] and (not current_section or current_section.lower() != filters["section"].lower()):
            continue
        if filters["tag"]:
            tags_value = _get_tag_value(rest, TAG_TAGS_RE) or ""
            if filters["tag"] not in tags_value.lower():
                continue
        if filters["issue"]:
            issue_value = _get_tag_value(rest, TAG_ISSUE_RE)
            if filters["issue"] == "*":
                if not issue_value:
                    continue
            else:
                if issue_value != filters["issue"]:
                    continue
        if filters["effort"]:
            effort_value = _get_tag_value(rest, TAG_EFFORT_RE)
            if filters["effort"] == "*":
                if not effort_value:
                    continue
            else:
                # Normalize both to compare (add 'h' if missing)
                filter_normalized = filters["effort"] if filters["effort"].endswith("h") else filters["effort"] + "h"
                effort_normalized = effort_value if effort_value and effort_value.endswith("h") else (effort_value + "h" if effort_value else "")
                if effort_normalized != filter_normalized:
                    continue
        if filters["system"]:
            system_value = _get_tag_value(rest, TAG_SYSTEM_RE)
            if filters["system"] == "*":
                if not system_value:
                    continue
            else:
                # Exact match for system tags
                if system_value != filters["system"]:
                    continue
        if milestone_filter_key is not None:
            task_key = resolve_milestone_key(
                _get_tag_value(rest, TAG_MILESTONE_RE), alias_map, sentinel)
            if task_key.lower() != milestone_filter_key.lower():
                continue

        print(line.rstrip())


def _collect_children(lines: List[str], parent_idx: int, parent_indent: int) -> List[str]:
    """Collect child item lines (and detail lines) under a parent."""
    children = []
    for j in range(parent_idx + 1, len(lines)):
        ln = lines[j]
        if is_section_header(ln) or ln.strip() == "---" or ln.startswith("# "):
            break
        p = parse_item(ln)
        if p:
            if len(p[0]) <= parent_indent:
                break
            children.append(ln)
        elif is_child_content(ln, parent_indent) and ln.strip():
            children.append(ln)
        elif not ln.strip():
            continue
        else:
            break
    return children


def cmd_tree():
    lines = load()

    # Pass 1: identify Features with shadows
    shadow_feature_ids: Set[str] = set()
    shadow_line_indices: Set[int] = set()
    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if parsed and parsed[2].startswith("F-") and is_shadow(line):
            shadow_feature_ids.add(parsed[2])
            shadow_line_indices.add(i)

    # Mark children of shadow features AND primary features (that have shadows)
    # so we skip them in main pass — they'll be shown in the merged view instead
    skip_indices: Set[int] = set(shadow_line_indices)

    # Also find primary lines for features with shadows
    for fid in shadow_feature_ids:
        primary_idx = find_primary_line(lines, fid)
        if primary_idx is not None:
            # Mark primary's children (not the primary itself — we render it in merged view)
            for j in range(primary_idx + 1, len(lines)):
                ln = lines[j]
                if is_section_header(ln) or ln.strip() == "---" or ln.startswith("# "):
                    break
                p = parse_item(ln)
                if p:
                    if len(p[0]) == 0:
                        break
                    skip_indices.add(j)
                elif not ln.strip():
                    continue
                elif is_child_content(ln, 0):
                    skip_indices.add(j)
                else:
                    break

    for i in sorted(shadow_line_indices):
        parsed = parse_item(lines[i])
        if not parsed:
            continue
        parent_indent = len(parsed[0])
        for j in range(i + 1, len(lines)):
            ln = lines[j]
            if is_section_header(ln) or ln.strip() == "---" or ln.startswith("# "):
                break
            p = parse_item(ln)
            if p:
                if len(p[0]) <= parent_indent:
                    break
                skip_indices.add(j)
            elif not ln.strip():
                continue
            elif is_child_content(ln, parent_indent):
                skip_indices.add(j)
            else:
                break

    # Pass 2: render
    for i, line in iter_content_lines(lines):
        if i in skip_indices:
            continue

        parsed = parse_item(line)
        if not parsed:
            continue
        indent, box, iid, prio, status, rest = parsed
        level = len(indent) // 2
        sym = "✓" if status == "done" else ("▶" if status == "doing" else "•")

        if iid in shadow_feature_ids and level == 0:
            # Merged view for this Feature
            print(f"{sym} {iid} {prio} {status} — {rest}")

            # Collect children grouped by section
            # Primary children first
            primary_idx = find_primary_line(lines, iid)
            sections_children: List[Tuple[str, List[str]]] = []
            if primary_idx is not None:
                sec = get_section_for_item(lines, primary_idx) or "Now"
                children = _collect_children(lines, primary_idx, 0)
                if children:
                    sections_children.append((sec, children))

            # Shadow children
            for shadow_idx, shadow_section in get_all_shadows(lines, iid):
                children = _collect_children(lines, shadow_idx, 0)
                if children:
                    sections_children.append((shadow_section, children))

            for sec_name, children in sections_children:
                print(f"  [{sec_name}]")
                for child_line in children:
                    cp = parse_item(child_line)
                    if cp:
                        clevel = len(cp[0]) // 2
                        csym = "✓" if cp[4] == "done" else ("▶" if cp[4] == "doing" else "•")
                        print(f"{'  ' * (clevel + 1)}{csym} {cp[2]} {cp[3]} {cp[4]} — {cp[5]}")
                    else:
                        print(f"    {child_line.strip()}")
        else:
            print(f"{'  ' * level}{sym} {iid} {prio} {status} — {rest}")


def _print_show_metadata(lines: List[str], idx: int, item_id: str, full: bool):
    """Print metadata block for cmd_show."""
    parsed = parse_item(lines[idx])
    if parsed:
        rest = parsed[5]
        print()
        section = get_section_for_item(lines, idx)
        if section:
            print(f"  Section: {section}")

        for label, tag_re in [
            ("Branch", TAG_BRANCH_RE), ("PR", TAG_PR_RE),
            ("Depends on", TAG_DEPS_RE), ("Related", TAG_REL_RE),
            ("Issue", TAG_ISSUE_RE), ("Tags", TAG_TAGS_RE),
            ("Effort", TAG_EFFORT_RE), ("System", TAG_SYSTEM_RE),
        ]:
            val = _get_tag_value(rest, tag_re)
            if val:
                prefix = "#" if label in ("PR", "Issue") else ""
                print(f"  {label}: {prefix}{val}")

        # Milestone (resolved through the registry when one exists).
        mval = _get_tag_value(rest, TAG_MILESTONE_RE)
        if mval:
            registry, alias_map = parse_milestone_registry(lines)
            key = resolve_milestone_key(mval, alias_map, milestone_sentinel())
            print(f"  Milestone: {milestone_label(key, registry, milestone_sentinel())}")

    if full:
        notes = get_notes_for_id(lines, item_id)
        if notes:
            print("  Notes:")
            print("  " + "-" * 40)
            for line in notes.split("\n"):
                print(f"  {line}")
            print("  " + "-" * 40)
        else:
            print("  Notes: (none)")

    print()


def cmd_show(item_id: str, full: bool = False):
    """Show a single item with its details and children.

    Args:
        item_id: The task/feature ID to show
        full: If True, also display notes from the # Notes section
    """
    item_id = normalize_id(item_id)
    lines = load()

    # For Features, prefer the primary line (not shadow)
    if item_id.startswith("F-"):
        primary = find_primary_line(lines, item_id)
        if primary is not None:
            idx = primary
            base_indent = 0  # Features are always top-level
        else:
            idx, base_indent = find_item_line_index(lines, item_id)
    else:
        idx, base_indent = find_item_line_index(lines, item_id)

    if idx is None:
        raise SystemExit(f"Item {item_id} not found")

    # Check if this is a Feature with shadows — use merged view
    if item_id.startswith("F-") and not is_shadow(lines[idx]):
        shadows = get_all_shadows(lines, item_id)
        if shadows:
            # Merged Feature view
            print(f"\n{lines[idx].rstrip()}")

            # Primary children
            primary_section = get_section_for_item(lines, idx) or "Now"
            primary_children = _collect_children(lines, idx, base_indent)
            if primary_children:
                print(f"  [{primary_section}]")
                for cl in primary_children:
                    print(f"    {cl.strip()}")

            # Shadow children
            for shadow_idx, shadow_section in shadows:
                shadow_children = _collect_children(lines, shadow_idx, 0)
                if shadow_children:
                    print(f"  [{shadow_section}]")
                    for cl in shadow_children:
                        print(f"    {cl.strip()}")

            _print_show_metadata(lines, idx, item_id, full)
            return

    # Standard view (Tasks, or Features without shadows)
    print(f"\n{lines[idx].rstrip()}")

    j = idx + 1
    while j < len(lines):
        if is_section_header(lines[j]):
            break

        parsed = parse_item(lines[j])
        if parsed:
            indent_len = len(parsed[0])
            if indent_len <= base_indent:
                break
            print(lines[j].rstrip())
            j += 1
            continue

        # detail lines (blank or indented content)
        if is_child_content(lines[j], base_indent):
            if lines[j].strip():  # Only print non-blank lines
                print(lines[j].rstrip())
            j += 1
            continue

        break

    _print_show_metadata(lines, idx, item_id, full)


def cmd_next(args: List[str] = None):
    """Find the next actionable task: highest priority, unblocked, in 'Now' section, status=todo.

    With `--milestone <id|alias>`, restrict to the next actionable task within
    that milestone (alias-resolved). Default `next` is unchanged.
    """
    args = args or []
    milestone_filter = None
    i = 0
    while i < len(args):
        if args[i] == "--milestone":
            if i + 1 >= len(args):
                raise SystemExit("--milestone requires a value (id/alias, or 'default')")
            milestone_filter = args[i + 1]
            i += 2
        else:
            raise SystemExit(f"Unknown arg: {args[i]}. Usage: next [--milestone ID]")

    lines = load()
    items = get_all_items(lines)

    # Get done items for dependency checking
    done_ids = {iid for iid, item in items.items() if item["status"] == "done"}

    # Resolve the milestone filter (registry powers alias resolution).
    milestone_filter_key = None
    alias_map: Dict[str, str] = {}
    sentinel = ""
    if milestone_filter:
        _registry, alias_map = parse_milestone_registry(lines)
        sentinel = milestone_sentinel()
        milestone_filter_key = resolve_milestone_key(milestone_filter, alias_map, sentinel)

    # Find items in "Now" section
    now_start, now_end = find_section_bounds(lines, "Now")
    if now_start is None:
        raise SystemExit("No '## Now' section found")

    candidates = []
    for iid, item in items.items():
        # Only Tasks are actionable (not Features)
        if iid.startswith("F-"):
            continue

        # Must be in Now section
        if not (now_start < item["line_num"] < (now_end or len(lines))):
            continue

        # Must be todo status
        if item["status"] != "todo":
            continue

        # Optional milestone restriction
        if milestone_filter_key is not None:
            task_key = resolve_milestone_key(item["milestone"], alias_map, sentinel)
            if task_key.lower() != milestone_filter_key.lower():
                continue

        # Must not be blocked by unfinished deps
        blocked = False
        for dep in item["deps"]:
            if dep not in done_ids:
                blocked = True
                break
        if blocked:
            continue

        candidates.append((PRIO_ORDER.get(item["prio"], 99), item["line_num"], iid, item))

    if not candidates:
        print("No actionable tasks found in 'Now' section.")
        print("  - All tasks may be done, blocked, or in Backlog")
        return

    # Sort by priority, then by line number (earlier = higher)
    candidates.sort(key=lambda x: (x[0], x[1]))
    _, _, next_id, next_item = candidates[0]

    print(f"Next task: {next_id}")
    print(f"  Priority: {next_item['prio']}")
    print(f"  {next_item['rest']}")

    if len(candidates) > 1:
        print(f"\n  ({len(candidates) - 1} more actionable tasks in queue)")


def cmd_update(item_id: str, *, status: Optional[str] = None, prio: Optional[str] = None, enforce_single_wip: bool = False):
    item_id = normalize_id(item_id)
    lines = load()

    if enforce_single_wip and status == "doing":
        for i, line in iter_content_lines(lines):
            parsed = parse_item(line)
            if parsed and parsed[2] != item_id and parsed[4] == "doing":
                lines[i] = normalize_item_line(line, status="todo")

    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if not parsed:
            continue
        if parsed[2] != item_id:
            continue
        lines[i] = normalize_item_line(line, status=status, prio=prio)
        save(lines)
        print(f"Updated {item_id}")
        return

    raise SystemExit(f"Item {item_id} not found")


def cmd_new(args: List[str]):
    if len(args) < 2:
        raise SystemExit('usage: new (feature|task) "Title" [--prio P1] [--status todo] [--under ID | --section Backlog] [--effort 4h] [--tags x,y] [--deps T-1,T-2]')

    kind_word = args[0].lower()
    kind = "F" if kind_word == "feature" else ("T" if kind_word == "task" else None)
    if not kind:
        raise SystemExit("First argument must be 'feature' or 'task'")

    title = args[1]
    prio = "P2"
    status = "todo"
    under = None
    section = "Backlog"
    effort = None
    tags = None
    deps = None
    milestone = None

    i = 2
    while i < len(args):
        if args[i] in ("--prio", "--priority"):
            if i + 1 >= len(args):
                raise SystemExit("--prio requires a value (P0|P1|P2|P3)")
            prio = args[i + 1]
            i += 2
        elif args[i] == "--status":
            if i + 1 >= len(args):
                raise SystemExit("--status requires a value")
            status = args[i + 1]
            i += 2
        elif args[i] == "--under":
            if i + 1 >= len(args):
                raise SystemExit("--under requires an item ID")
            under = normalize_id(args[i + 1])
            i += 2
        elif args[i] == "--section":
            if i + 1 >= len(args):
                raise SystemExit("--section requires a section name")
            section = args[i + 1]
            i += 2
        elif args[i] == "--effort":
            if i + 1 >= len(args):
                raise SystemExit("--effort requires a value (e.g., 4h)")
            effort = args[i + 1]
            i += 2
        elif args[i] == "--tags":
            if i + 1 >= len(args):
                raise SystemExit("--tags requires comma-separated values")
            tags = args[i + 1]
            i += 2
        elif args[i] == "--deps":
            if i + 1 >= len(args):
                raise SystemExit("--deps requires comma-separated IDs (e.g., T-1,T-2)")
            deps = args[i + 1]
            i += 2
        elif args[i] == "--milestone":
            if i + 1 >= len(args):
                raise SystemExit("--milestone requires an id or alias (e.g., m1, alpha)")
            milestone = args[i + 1]
            i += 2
        else:
            raise SystemExit(f"Unknown arg: {args[i]}")

    if status not in ALLOWED_STATUS:
        raise SystemExit(f"Invalid status '{status}'")
    if prio not in ALLOWED_PRIO:
        raise SystemExit(f"Invalid priority '{prio}'")

    lines = load()
    iid = next_id(lines, kind)

    box = "x" if status == "done" else " "
    suffix = f" @done={date.today()}" if status == "done" else ""
    # Append inline tags
    tag_parts = []
    if effort:
        tag_parts.append(f"@effort={effort}")
    if tags:
        tag_parts.append(f"@tags={tags}")
    if deps:
        # Normalize dep IDs
        dep_ids = parse_id_list(deps)
        if dep_ids:
            tag_parts.append(f"@deps={','.join(dep_ids)}")
    if milestone is not None:
        mval = milestone.strip()
        # An explicit sentinel, "clear", or empty means "no milestone" — write no
        # tag (parity with `set --milestone`, which treats the same as a clear).
        if mval and mval.lower() not in ("clear", milestone_sentinel().lower()):
            if not re.match(r'^[a-zA-Z0-9_-]+$', mval):
                raise SystemExit(f"Invalid milestone '{mval}'. Use letters, digits, hyphens, underscores.")
            tag_parts.append(f"@milestone={mval}")
            _warn_unknown_milestone(mval, lines)
    tag_suffix = (" " + " ".join(tag_parts)) if tag_parts else ""
    item_line = f"- [{box}] ({iid}) [{prio}] [{status}] {title}{suffix}{tag_suffix}\n"

    if under:
        insert_at, parent_indent = find_insert_under(lines, under)
        child_indent = " " * (parent_indent + 2)
        lines.insert(insert_at, child_indent + item_line)
    else:
        pos = find_section_insert_pos(lines, section)
        if pos > 0 and lines[pos - 1].strip() != "":
            lines.insert(pos, "\n")
            pos += 1
        lines.insert(pos, item_line)
        if section.lower() == "now":
            lines = sort_section_top_level(lines, "Now")

    save(lines)
    print(f"Created {iid}")


def cmd_mv(args: List[str]):
    if len(args) < 3:
        raise SystemExit("usage: mv ID (--section NAME | --under ID)")

    item_id = normalize_id(args[0])
    flag = args[1]
    target = args[2]

    lines = load()
    # Capture old parent Feature before extraction (for shadow cleanup)
    old_idx, _ = find_item_line_index(lines, item_id)
    old_parent_feature = find_parent_feature(lines, old_idx) if old_idx is not None else None
    lines, block, old_base = extract_block(lines, item_id)

    if flag == "--section":
        new_base = 0
        adjusted = reindent_block(block, old_base, new_base)
        pos = find_section_insert_pos(lines, target)
        if pos > 0 and lines[pos - 1].strip() != "":
            lines.insert(pos, "\n")
            pos += 1
        for k, ln in enumerate(adjusted):
            lines.insert(pos + k, ln)
        # Cleanup empty shadows on old parent Feature
        if old_parent_feature:
            lines = cleanup_empty_shadows(lines, old_parent_feature)
        if target.lower() == "now":
            lines = sort_section_top_level(lines, "Now")
        save(lines)
        print(f"Moved {item_id} to section '{target}'")
        return

    if flag == "--under":
        target_id = normalize_id(target)
        insert_at, parent_indent = find_insert_under(lines, target_id)
        new_base = parent_indent + 2
        adjusted = reindent_block(block, old_base, new_base)
        for k, ln in enumerate(adjusted):
            lines.insert(insert_at + k, ln)
        # Cleanup empty shadows on old parent Feature
        if old_parent_feature:
            lines = cleanup_empty_shadows(lines, old_parent_feature)
        save(lines)
        print(f"Moved {item_id} under {target_id}")
        return

    raise SystemExit("mv requires --section or --under")


def cmd_link(args: List[str]):
    if len(args) < 1:
        raise SystemExit("usage: link ID [--deps add|rm|set|clear LIST] [--rel add|rm|set|clear LIST]")

    item_id = args[0]
    deps_op = deps_list = None
    rel_op = rel_list = None

    i = 1
    while i < len(args):
        if args[i] == "--deps":
            if i + 1 >= len(args):
                raise SystemExit("--deps requires an operation (add|rm|set|clear)")
            deps_op = args[i + 1]
            if deps_op.lower() != "clear":
                if i + 2 >= len(args):
                    raise SystemExit(f"--deps {deps_op} requires a comma-separated list of IDs")
                deps_list = args[i + 2]
                i += 3
            else:
                deps_list = ""
                i += 2
        elif args[i] == "--rel":
            if i + 1 >= len(args):
                raise SystemExit("--rel requires an operation (add|rm|set|clear)")
            rel_op = args[i + 1]
            if rel_op.lower() != "clear":
                if i + 2 >= len(args):
                    raise SystemExit(f"--rel {rel_op} requires a comma-separated list of IDs")
                rel_list = args[i + 2]
                i += 3
            else:
                rel_list = ""
                i += 2
        else:
            raise SystemExit(f"Unknown arg: {args[i]}")

    edit_links(item_id, deps_op=deps_op, deps_list=deps_list, rel_op=rel_op, rel_list=rel_list)


def cmd_set(args: List[str]):
    if len(args) < 3:
        raise SystemExit("usage: set ID --branch NAME | --pr NUMBER | --issue NUMBER | --tags TAG1,TAG2 | --effort HOURS")

    item_id = args[0]
    flag = args[1]
    value = args[2]

    if flag == "--branch":
        # Allow clearing with empty string or 'clear'
        if value.lower() == "clear" or value == "":
            set_tag(item_id, "branch", "")
            print(f"Cleared @branch on {normalize_id(item_id)}")
            return
        set_tag(item_id, "branch", value)
    elif flag == "--pr":
        # Allow clearing with 'clear' or '0'
        if value.lower() == "clear" or value == "0":
            set_tag(item_id, "pr", "")
            print(f"Cleared @pr on {normalize_id(item_id)}")
            return
        if not value.isdigit() or int(value) <= 0:
            raise SystemExit(f"PR number must be a positive integer, got '{value}'")
        set_tag(item_id, "pr", value)
    elif flag == "--issue":
        # Allow clearing with 'clear' or '0'
        if value.lower() == "clear" or value == "0":
            set_tag(item_id, "issue", "")
            print(f"Cleared @issue on {normalize_id(item_id)}")
            return
        if not value.isdigit() or int(value) <= 0:
            raise SystemExit(f"Issue number must be a positive integer, got '{value}'")
        set_tag(item_id, "issue", value)
    elif flag == "--tags":
        # Allow clearing with 'clear' or empty
        if value.lower() == "clear" or value == "":
            set_tag(item_id, "tags", "")
            print(f"Cleared @tags on {normalize_id(item_id)}")
            return
        # Validate tags are alphanumeric with hyphens, comma-separated
        tags = [t.strip() for t in value.split(",") if t.strip()]
        for tag in tags:
            if not re.match(r'^[a-zA-Z0-9_-]+$', tag):
                raise SystemExit(f"Invalid tag '{tag}'. Tags must be alphanumeric with hyphens/underscores.")
        set_tag(item_id, "tags", ",".join(tags))
    elif flag == "--effort":
        # Allow clearing with 'clear' or '0' or '0h'
        if value.lower() == "clear" or value == "0" or value == "0h":
            set_tag(item_id, "effort", "")
            print(f"Cleared @effort on {normalize_id(item_id)}")
            return
        # Validate effort format: number with optional 'h' suffix (e.g., 4, 4h, 8h)
        effort_match = re.match(r'^(\d+)h?$', value)
        if not effort_match:
            raise SystemExit(f"Effort must be hours (e.g., 4h, 8h, 16h), got '{value}'")
        hours = int(effort_match.group(1))
        if hours <= 0:
            raise SystemExit(f"Effort must be positive, got '{value}'")
        set_tag(item_id, "effort", f"{hours}h")
    elif flag == "--system":
        # Allow clearing with 'clear' or empty
        if value.lower() == "clear" or value == "":
            set_tag(item_id, "system", "")
            print(f"Cleared @system on {normalize_id(item_id)}")
            return
        # Validate system tag has valid prefix
        if not any(value.startswith(prefix) for prefix in SYSTEM_PREFIXES):
            valid_prefixes = ", ".join(SYSTEM_PREFIXES)
            raise SystemExit(f"System tag must start with: {valid_prefixes}. Got '{value}'")
        # Validate format: prefix + alphanumeric/hyphen identifier
        if not re.match(r'^[a-z]+-[a-zA-Z0-9_-]+$', value):
            raise SystemExit(f"Invalid system tag '{value}'. Format: prefix-identifier (e.g., sprint-1, milestone-mvp)")
        set_tag(item_id, "system", value)
    elif flag == "--milestone":
        sentinel = milestone_sentinel()
        # Clear back to the sentinel with empty string, 'clear', or the sentinel name.
        if value == "" or value.lower() in ("clear", sentinel.lower()):
            set_tag(item_id, "milestone", "")
            print(f"Cleared @milestone on {normalize_id(item_id)} (back to '{sentinel}')")
            return
        if not re.match(r'^[a-zA-Z0-9_-]+$', value):
            raise SystemExit(f"Invalid milestone '{value}'. Use letters, digits, hyphens, underscores.")
        set_tag(item_id, "milestone", value)
        _warn_unknown_milestone(value)
    else:
        raise SystemExit(f"Unknown flag: {flag}. Use --branch, --pr, --issue, --tags, --effort, --system, or --milestone")


def cmd_current():
    """Show current task(s) with [doing] status."""
    lines = load()
    doing_items = []

    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if parsed and parsed[4] == "doing":
            indent, box, iid, prio, status, rest = parsed
            section = get_section_for_item(lines, i)
            doing_items.append({
                "id": iid,
                "prio": prio,
                "rest": rest,
                "section": section,
                "line_num": i,
            })

    if not doing_items:
        print("No task currently in progress.")
        print("  Use './tools/tasks.py start <ID>' to start a task.")
        return

    print(f"\n{'=' * 60}")
    print("CURRENT WORK IN PROGRESS")
    print(f"{'=' * 60}\n")

    for item in doing_items:
        print(f"▶ {item['id']} [{item['prio']}] {item['rest']}")
        print(f"  Section: {item['section'] or 'Unknown'}")

        # Show metadata
        rest = item['rest']
        branch = _get_tag_value(rest, TAG_BRANCH_RE)
        if branch:
            print(f"  Branch: {branch}")

        pr = _get_tag_value(rest, TAG_PR_RE)
        if pr:
            print(f"  PR: #{pr}")

        issue = _get_tag_value(rest, TAG_ISSUE_RE)
        if issue:
            print(f"  Issue: #{issue}")

        tags = _get_tag_value(rest, TAG_TAGS_RE)
        if tags:
            print(f"  Tags: {tags}")

        deps = _get_tag_value(rest, TAG_DEPS_RE)
        if deps:
            print(f"  Depends on: {deps}")

        print()

    if len(doing_items) > 1:
        eprint(f"⚠️  Warning: Multiple tasks in [doing] status ({len(doing_items)}). Use 'start' to enforce single WIP.")


def cmd_sync(item_id: str):
    """Sync task with linked GitHub issue (fetch title, update status)."""
    item_id = normalize_id(item_id)
    lines = load()
    idx, _ = find_item_line_index(lines, item_id)

    if idx is None:
        raise SystemExit(f"Item {item_id} not found")

    parsed = parse_item(lines[idx])
    if not parsed:
        raise SystemExit(f"Item {item_id} not found")

    rest = parsed[5]
    issue_num = _get_tag_value(rest, TAG_ISSUE_RE)

    if not issue_num:
        raise SystemExit(f"Item {item_id} has no @issue= tag. Set one with: ./tools/tasks.py set {item_id} --issue NUMBER")

    # Check if gh CLI is available
    try:
        subprocess.run(["gh", "--version"], capture_output=True, check=True)
    except (subprocess.CalledProcessError, FileNotFoundError):
        raise SystemExit("GitHub CLI (gh) not found. Install from https://cli.github.com/")

    # Fetch issue details
    print(f"Fetching GitHub issue #{issue_num}...")
    try:
        result = subprocess.run(
            ["gh", "issue", "view", issue_num, "--json", "title,state,labels,url"],
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise SystemExit(f"Failed to fetch issue #{issue_num}: {e.stderr}")

    issue_data = json.loads(result.stdout)

    print(f"\n  Title: {issue_data['title']}")
    print(f"  State: {issue_data['state']}")
    print(f"  URL: {issue_data['url']}")

    if issue_data.get('labels'):
        label_names = [l['name'] for l in issue_data['labels']]
        print(f"  Labels: {', '.join(label_names)}")

        # Offer to sync labels as tags
        current_tags = _get_tag_value(rest, TAG_TAGS_RE)
        if label_names and current_tags != ','.join(label_names):
            print(f"\n  Current @tags: {current_tags or '(none)'}")
            print(f"  GitHub labels: {', '.join(label_names)}")
            print(f"\n  To sync labels as tags, run:")
            print(f"    ./tools/tasks.py set {item_id} --tags {','.join(label_names)}")

    # Check if issue is closed but task is not done
    status = parsed[4]
    if issue_data['state'] == 'CLOSED' and status != 'done':
        print(f"\n  ⚠️  Issue is CLOSED but task is [{status}]")
        print(f"  To mark task done: ./tools/tasks.py done {item_id}")
    elif issue_data['state'] == 'OPEN' and status == 'done':
        print(f"\n  ⚠️  Issue is OPEN but task is [done]")
        print(f"  Consider reopening the issue or reviewing the task status.")

    print()


def cmd_backlog(args: List[str] = None):
    """List backlog items, or move an item to Backlog section.

    For Features: collapses all shadows, unifying children into Backlog.
    For Tasks: creates shadow of parent Feature in Backlog if needed.

    Usage:
        backlog           # List all items in Backlog section
        backlog T-7       # Move T-7 to Backlog section
        backlog F-1       # Move F-1 + all children to Backlog (collapses shadows)
    """
    args = args or []
    if not args:
        cmd_list(["--section", "Backlog"])
    else:
        item_id = normalize_id(args[0])
        lines = load()
        move_item_to_section(lines, item_id, "Backlog")


def cmd_now(args: List[str] = None):
    """List Now items, or move an item to Now section.

    For Features: collapses all shadows, unifying children into Now.
    For Tasks: creates shadow of parent Feature in Now if needed.

    Usage:
        now               # List all items in Now section
        now T-7           # Move T-7 to Now section (promotes from Backlog)
        now F-1           # Move F-1 + all children to Now (collapses shadows)
    """
    args = args or []
    if not args:
        cmd_list(["--section", "Now"])
    else:
        item_id = normalize_id(args[0])
        lines = load()
        move_item_to_section(lines, item_id, "Now")


def cmd_nextid():
    """Show the next available Feature and Task IDs.

    Useful when manually editing TASKS.md to avoid duplicate IDs.
    """
    lines = load()
    next_f = next_id(lines, "F")
    next_t = next_id(lines, "T")

    print("Next available IDs:")
    print(f"  Feature: {next_f}")
    print(f"  Task:    {next_t}")
    print()
    print("Use these when manually adding items to TASKS.md:")
    print(f"  - [ ] ({next_f}) [P2] [todo] New feature title @tags=...")
    print(f"    - [ ] ({next_t}) [P2] [todo] New task title @tags=...")


def _fmt_milestone_counts(counts: Dict[str, int]) -> str:
    """One-line status breakdown + total + %done for a milestone bucket."""
    total = sum(counts.values())
    done = counts.get("done", 0)
    pct = int(round(done * 100 / total)) if total else 0
    parts = "  ".join(f"{s}:{counts.get(s, 0)}"
                       for s in ("todo", "doing", "done", "skipped", "deferred"))
    noun = "task" if total == 1 else "tasks"
    return f"{parts}  |  {total} {noun}, {pct}% done"


def _milestone_header(key: str, registry: Dict[str, dict], sentinel: str) -> str:
    """Display header for a bucket: 'M1 (alpha) [active] — Title', 'default', or raw."""
    if key == sentinel:
        return sentinel
    entry = registry.get(key)
    if not entry:
        return key  # freeform (no registry)
    label = milestone_label(key, registry, sentinel)
    status = f" [{entry['status']}]" if entry.get("status") else ""
    title = f" — {entry['title']}" if entry.get("title") else ""
    return f"{label}{status}{title}"


def _print_milestone_table(rows: List[Tuple[str, str, str]]):
    """Render the milestone x features x status table with aligned columns."""
    headers = ("MILESTONE", "FEATURES", "STATUS")
    all_rows = [headers] + rows
    w0 = max(len(r[0]) for r in all_rows)
    w1 = max(len(r[1]) for r in all_rows)
    for label, feats, status in all_rows:
        print(f"{label.ljust(w0)}  {feats.ljust(w1)}  {status}")


def _ordered_milestone_keys(keys: Set[str], registry: Dict[str, dict], sentinel: str) -> List[str]:
    """Registry milestones first (by numeric ID), then freeform (sorted), sentinel last."""
    reg = sorted((k for k in registry if k in keys),
                 key=lambda k: int(re.sub(r'\D', '', k) or 0))
    freeform = sorted(k for k in keys if k not in registry and k != sentinel)
    ordered = reg + freeform
    if sentinel in keys:
        ordered.append(sentinel)
    return ordered


def cmd_milestone(args: List[str] = None):
    """Milestone rollup, or `milestone <id|alias>` detail.

    Counts Tasks (T-) grouped by their resolved milestone (ID+alias unified via
    the registry), including the implicit sentinel bucket. Read-only.
    """
    args = args or []
    detail_arg = None
    table_mode = False
    for a in args:
        if a == "--table":
            table_mode = True
        elif a.startswith("--"):
            raise SystemExit(f"Unknown arg: {a}. Usage: milestone [ID|alias] [--table]")
        elif detail_arg is None:
            detail_arg = a
        else:
            raise SystemExit(f"Unexpected extra argument '{a}'. Usage: milestone [ID|alias] [--table]")
    if table_mode and detail_arg is not None:
        raise SystemExit("milestone: give either an ID/alias (detail) or --table, not both.")

    lines = load()
    registry, alias_map = parse_milestone_registry(lines)
    sentinel = milestone_sentinel()
    items = get_all_items(lines)

    # Group case-insensitively (matching the list/next filters), but keep a stable
    # display: registry keys are already canonical (M1); for freeform values the
    # first spelling seen wins so `@milestone=Alpha` and `=alpha` share one bucket.
    canon: Dict[str, str] = {}

    def canon_key(item_milestone: Optional[str]) -> str:
        key = resolve_milestone_key(item_milestone, alias_map, sentinel)
        return canon.setdefault(key.lower(), key)

    buckets: Dict[str, Dict[str, int]] = {}
    bucket_tasks: Dict[str, List[dict]] = {}
    for iid, item in items.items():
        if not iid.startswith("T-"):
            continue
        key = canon_key(item["milestone"])
        counts = buckets.setdefault(key, {s: 0 for s in ALLOWED_STATUS})
        counts[item["status"]] = counts.get(item["status"], 0) + 1
        bucket_tasks.setdefault(key, []).append({**item, "id": iid})

    if table_mode:
        # Attribute features to milestones: a feature's own @milestone=, plus the
        # milestone of any task under it (a top-level tagged task stands in for
        # itself). Features with neither land in the sentinel row.
        feat_ms: Dict[str, Set[str]] = {}
        for iid, item in items.items():
            if iid.startswith("F-"):
                keys = feat_ms.setdefault(iid, set())
                if item["milestone"] is not None:
                    keys.add(canon_key(item["milestone"]))
        for iid, item in items.items():
            if iid.startswith("T-") and item["milestone"] is not None:
                key = canon_key(item["milestone"])
                parent = find_parent_feature(lines, item["line_num"])
                feat_ms.setdefault(parent or iid, set()).add(key)

        milestone_features: Dict[str, Set[str]] = {}
        default_features: Set[str] = set()
        for fid, keys in feat_ms.items():
            if keys:
                for k in keys:
                    milestone_features.setdefault(k, set()).add(fid)
            else:
                default_features.add(fid)

        display = set(milestone_features) | set(registry)
        ordered = [k for k in _ordered_milestone_keys(display, registry, sentinel) if k != sentinel]
        rows = []
        for key in ordered:
            feats = ", ".join(sort_ids(list(milestone_features.get(key, set())))) or "—"
            status = (registry.get(key) or {}).get("status") or "—"
            rows.append((milestone_label(key, registry, sentinel), feats, status))
        if default_features:
            rows.append((sentinel, ", ".join(sort_ids(list(default_features))), "—"))
        _print_milestone_table(rows)
        return

    if detail_arg is not None:
        key = canon_key(detail_arg)
        header = f"MILESTONE {_milestone_header(key, registry, sentinel)}"
        print(f"\n{header}")
        print("=" * max(16, len(header.splitlines()[0])))
        tasks = bucket_tasks.get(key, [])
        if not tasks:
            print("  (no tasks in this milestone)")
            print()
            return
        for st in ("doing", "todo", "deferred", "skipped", "done"):
            group = [t for t in tasks if t["status"] == st]
            if not group:
                continue
            print(f"  {st}:")
            for t in sorted(group, key=lambda x: (PRIO_ORDER.get(x["prio"], 9), x["id"])):
                print(f"    {t['id']} [{t['prio']}] {t['rest']}")
        print()
        print("  " + _fmt_milestone_counts(buckets.get(key, {})))
        print()
        return

    # Rollup — include registry milestones even when they carry no tasks yet.
    display_keys = set(buckets) | set(registry)
    ordered = _ordered_milestone_keys(display_keys, registry, sentinel)
    print("MILESTONE ROLLUP")
    print("=" * 16)
    if not ordered:
        print("  (no tasks)")
        return
    for key in ordered:
        print(_milestone_header(key, registry, sentinel))
        print("  " + _fmt_milestone_counts(buckets.get(key, {})))


def cmd_migrate_tags_to_milestone(args: List[str] = None):
    """Rewrite an interim `@tags=<id>` convention to first-class `@milestone=<id>`.

    Projects adopting milestones before this landed tagged work with a generic
    `@tags=m1`; the rollup won't recognize those. This one-shot helper moves that
    label out of `@tags=` and into `@milestone=`, so no task is silently orphaned
    from its milestone.

    A task that already carries an @milestone= is handled carefully: if it
    resolves to the SAME milestone, the redundant tag is just dropped; if it
    differs, the task is left completely untouched and reported as a conflict —
    never silently stripped (that would delete the very association we protect).
    """
    args = args or []
    if not args or not args[0].strip():
        raise SystemExit('usage: migrate-tags-to-milestone <tag>   (e.g. m1)')
    tag = args[0].strip()

    # Same guards as `new`/`set --milestone`: don't mint an invalid or sentinel
    # milestone (the sentinel must never be written into a task line).
    sentinel = milestone_sentinel()
    if tag.lower() == sentinel.lower():
        raise SystemExit(f"migrate-tags-to-milestone: '{tag}' is the sentinel; nothing to migrate.")
    if not re.match(r'^[a-zA-Z0-9_-]+$', tag):
        raise SystemExit(f"Invalid milestone '{tag}'. Use letters, digits, hyphens, underscores.")

    lines = load()
    _registry, alias_map = parse_milestone_registry(lines)
    tag_key = resolve_milestone_key(tag, alias_map, sentinel)

    changed = 0
    conflicts: List[Tuple[str, str]] = []  # (item_id, existing @milestone= value)
    for i, line in iter_content_lines(lines):
        parsed = parse_item(line)
        if not parsed:
            continue
        indent, box, iid, prio, status, rest = parsed
        tags_value = _get_tag_value(rest, TAG_TAGS_RE)
        if not tags_value:
            continue
        tag_list = [t for t in tags_value.split(",") if t]
        if not any(t.lower() == tag.lower() for t in tag_list):
            continue

        existing = _get_tag_value(rest, TAG_MILESTONE_RE)
        if existing is not None and resolve_milestone_key(existing, alias_map, sentinel).lower() != tag_key.lower():
            # Different milestone already assigned — leave the line alone so we
            # don't silently delete either association. Surface it instead.
            conflicts.append((iid, existing))
            continue

        remaining = [t for t in tag_list if t.lower() != tag.lower()]
        new_rest = _set_or_remove_tag(rest, "tags", remaining)
        if existing is None:
            new_rest = _set_single_tag(new_rest, "milestone", tag)
        lines[i] = build_item_line(indent, iid, prio, status, new_rest)
        changed += 1

    if changed:
        save(lines)

    if changed == 0 and not conflicts:
        print(f"No tasks had @tags={tag}; nothing migrated.")
    else:
        print(f"Migrated {changed} task(s): @tags={tag} -> @milestone={tag}")
    for iid, existing in conflicts:
        eprint(f"  ⚠️  skipped {iid}: keeps existing @milestone={existing} (conflicts with '{tag}'); resolve by hand.")


def _version_tuple(v: str) -> Tuple[int, ...]:
    parts = re.findall(r'\d+', v)
    if not parts:
        raise ValueError(f"unparseable version string: {v!r}")
    return tuple(int(x) for x in parts)


def _read_source(src: str) -> str:
    """Read canonical tasks.py from a local path or HTTPS URL (zero-dep).

    Plaintext http:// is refused: fetch-then-overwrite-then-execute over an
    unauthenticated channel is a MITM remote-code-execution vector.
    """
    if src.startswith("http://"):
        raise SystemExit("selfupdate: refusing insecure http:// source; use https://")
    if src.startswith("https://"):
        import urllib.request
        with urllib.request.urlopen(src, timeout=10) as resp:
            final = getattr(resp, "url", src) or src
            if not final.startswith("https://"):  # reject https->http downgrade redirects
                raise SystemExit(f"selfupdate: refusing non-https redirect to '{final}'")
            data = resp.read(_MAX_FETCH_BYTES + 1)
        if len(data) > _MAX_FETCH_BYTES:
            raise SystemExit(f"selfupdate: source exceeds {_MAX_FETCH_BYTES} bytes")
        return data.decode("utf-8")
    with open(src, "r", encoding="utf-8") as fh:
        return fh.read()


def _parse_version(content: str) -> Optional[str]:
    m = re.search(r'^__version__\s*=\s*["\']([^"\']+)["\']', content, re.M)
    return m.group(1) if m else None


def cmd_version():
    print(f"tasks {__version__}")
    print(f"canonical source: {CANONICAL_SOURCE}")


def cmd_selfupdate(args: List[str]):
    """Sync THIS vendored tasks.py to the canonical version.

    Governs drift across projects that vendor tasks.py: a stale copy pulls the
    canonical engine in place. Initial adoption is a copy (older copies lack
    this command); selfupdate keeps already-adopted copies current thereafter.
    """
    check_only = "--check" in args
    allow_untrusted = "--allow-untrusted-source" in args
    src = CANONICAL_SOURCE
    if "--source" in args:
        i = args.index("--source")
        if i + 1 >= len(args) or args[i + 1].startswith("-"):
            raise SystemExit(
                "usage: selfupdate [--check] [--source PATH_OR_URL] [--allow-untrusted-source]")
        src = args[i + 1]

    # selfupdate overwrites this script and runs the result next time, so a
    # non-default source is a trust decision the user must make explicitly.
    if src != DEFAULT_CANONICAL_SOURCE and not allow_untrusted:
        raise SystemExit(
            f"selfupdate: '{src}' is not the default canonical source.\n"
            f"It would overwrite {os.path.realpath(__file__)} with code from there.\n"
            f"Re-run with --allow-untrusted-source to confirm.")

    try:
        content = _read_source(src)
    except (OSError, ValueError) as e:
        raise SystemExit(f"selfupdate: cannot read canonical source '{src}': {e}")

    canon_ver = _parse_version(content)
    if not canon_ver:
        raise SystemExit(f"selfupdate: canonical source has no __version__ marker ({src})")

    # Validate the payload really is this tool before replacing ourselves.
    if "def cmd_selfupdate(" not in content or "def main(" not in content:
        raise SystemExit(f"selfupdate: fetched content does not look like tasks.py ({src})")
    try:
        compile(content, "<canonical-tasks.py>", "exec")
    except SyntaxError as e:
        raise SystemExit(f"selfupdate: fetched content is not valid Python: {e}")

    try:
        is_newer = _version_tuple(canon_ver) > _version_tuple(__version__)
    except ValueError as e:
        raise SystemExit(f"selfupdate: {e}")
    if not is_newer:
        print(f"already current: tasks {__version__} (canonical {canon_ver})")
        return

    print(f"update available: {__version__} -> {canon_ver} (source: {src})")
    if check_only:
        print("(--check: not writing)")
        return

    target = os.path.realpath(__file__)  # update the real file, not a symlink to it
    tmp = target + ".selfupdate.tmp"
    mode = os.stat(target).st_mode & 0o777
    try:
        # O_EXCL: never follow a pre-existing temp file / planted symlink.
        fd = os.open(tmp, os.O_CREAT | os.O_EXCL | os.O_WRONLY, mode)
        with os.fdopen(fd, "w", encoding="utf-8") as fh:
            fh.write(content)
        os.replace(tmp, target)  # atomic on the same filesystem
    except OSError as e:
        if os.path.exists(tmp):
            try:
                os.remove(tmp)
            except OSError:
                pass
        raise SystemExit(f"selfupdate: failed to write {target}: {e}")
    print(f"updated {target}: tasks {__version__} -> {canon_ver}. Run `tasks validate`.")


def main():
    if len(sys.argv) < 2:
        cmd_help()
        sys.exit(1)

    cmd = sys.argv[1].lower()

    if cmd in ("help", "-h", "--help"):
        cmd_help()
    elif cmd in ("version", "--version", "-v"):
        cmd_version()
    elif cmd == "selfupdate":
        cmd_selfupdate(sys.argv[2:])
    elif cmd == "init":
        force = ("--force" in sys.argv[2:])
        cmd_init(force=force)
    elif cmd == "validate":
        cmd_validate()
    elif cmd == "list":
        cmd_list(sys.argv[2:])
    elif cmd == "tree":
        cmd_tree()
    elif cmd == "show":
        if len(sys.argv) < 3:
            raise SystemExit("usage: show ID [--full]")
        item_id = sys.argv[2]
        full = "--full" in sys.argv[3:]
        cmd_show(item_id, full=full)
    elif cmd == "next":
        cmd_next(sys.argv[2:])
    elif cmd == "new":
        cmd_new(sys.argv[2:])
    elif cmd == "mv":
        cmd_mv(sys.argv[2:])
    elif cmd == "set":
        cmd_set(sys.argv[2:])
    elif cmd == "prio":
        if len(sys.argv) < 4:
            raise SystemExit("usage: prio ID P0|P1|P2|P3")
        cmd_update(sys.argv[2], prio=sys.argv[3])
    elif cmd in ("start", "done", "skip", "defer"):
        if len(sys.argv) < 3:
            raise SystemExit(f"usage: {cmd} ID")
        item_id_raw = sys.argv[2]
        item_id_norm = normalize_id(item_id_raw)
        is_task = item_id_norm.startswith("T-")

        if cmd == "skip" and is_task:
            # skip Task: move to ## Skipped (unless already there)
            _lines = load()
            _idx, _ = find_item_line_index(_lines, item_id_norm)
            _sec = get_section_for_item(_lines, _idx) if _idx is not None else None
            if _sec and _sec.lower() == "skipped":
                cmd_update(item_id_raw, status="skipped")
            else:
                move_item_to_section(_lines, item_id_norm, "Skipped", new_status="skipped")
        elif cmd == "start" and is_task:
            # start Task: move to ## Now (unless already there)
            _lines = load()
            _idx, _ = find_item_line_index(_lines, item_id_norm)
            _sec = get_section_for_item(_lines, _idx) if _idx is not None else None
            if _sec and _sec.lower() == "now":
                cmd_update(item_id_raw, status="doing", enforce_single_wip=True)
            else:
                move_item_to_section(_lines, item_id_norm, "Now", new_status="doing", enforce_single_wip=True)
        else:
            # done, defer (terminal, no move), or Feature status changes (no move)
            # Guard: reject if already in target status
            _lines = load()
            _idx, _ = find_item_line_index(_lines, item_id_norm)
            if _idx is not None:
                _parsed = parse_item(_lines[_idx])
                if _parsed:
                    target_status = {"start": "doing", "done": "done", "skip": "skipped", "defer": "deferred"}[cmd]
                    if _parsed[4] == target_status:
                        raise SystemExit(f"{item_id_norm} is already {target_status}")
            status = {"start": "doing", "done": "done", "skip": "skipped", "defer": "deferred"}[cmd]
            cmd_update(item_id_raw, status=status, enforce_single_wip=(cmd == "start"))
    elif cmd == "reopen":
        if len(sys.argv) < 3:
            raise SystemExit("usage: reopen ID")
        _reopen_id = normalize_id(sys.argv[2])
        _lines = load()
        _idx, _ = find_item_line_index(_lines, _reopen_id)
        if _idx is None:
            raise SystemExit(f"Item {_reopen_id} not found")
        _parsed = parse_item(_lines[_idx])
        if _parsed and _parsed[4] not in ("done", "skipped", "deferred"):
            raise SystemExit(f"{_reopen_id} is [{_parsed[4]}], not in a terminal state (done/skipped/deferred)")
        cmd_update(sys.argv[2], status="todo")
    elif cmd == "link":
        cmd_link(sys.argv[2:])
    elif cmd == "current":
        cmd_current()
    elif cmd == "sync":
        if len(sys.argv) < 3:
            raise SystemExit("usage: sync ID")
        cmd_sync(sys.argv[2])
    elif cmd == "backlog":
        cmd_backlog(sys.argv[2:])
    elif cmd == "now":
        cmd_now(sys.argv[2:])
    elif cmd == "nextid":
        cmd_nextid()
    elif cmd == "milestone":
        cmd_milestone(sys.argv[2:])
    elif cmd == "migrate-tags-to-milestone":
        cmd_migrate_tags_to_milestone(sys.argv[2:])
    else:
        raise SystemExit(f"Unknown command '{cmd}'. Try: ./tools/tasks.py help")


if __name__ == "__main__":
    main()
