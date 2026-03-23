#!/bin/bash
# SessionStart hook: main ブランチで TODO.md の全条件完了タスクを CHANGELOG.md に自動移動する

branch=$(git branch --show-current 2>/dev/null)
if [ "$branch" != "main" ]; then
  exit 0
fi

todo_file="$PWD/TODO.md"
changelog_file="$PWD/CHANGELOG.md"

if [ ! -f "$todo_file" ] || ! grep -q '\[x\]' "$todo_file"; then
  exit 0
fi

tmp_todo=$(mktemp)
tmp_names=$(mktemp)
trap 'rm -f "$tmp_todo" "$tmp_names"' EXIT

# TODO.md をパースし、全条件が [x] のタスクブロックを除去、タスク名を tmp_names に出力
awk -v names_file="$tmp_names" '
function flush() {
  if (block == "") return
  if (criteria > 0 && unchecked == 0) {
    print name > names_file
  } else {
    printf "%s", block
  }
  block = ""; name = ""; criteria = 0; unchecked = 0
}

/^## 未着手/ { in_sec = 1; print; next }

/^## / {
  if (in_sec) { flush(); in_sec = 0 }
  print; next
}

!in_sec { print; next }

/^- / {
  flush()
  block = $0 "\n"
  name = substr($0, 3)
  next
}

/^[[:space:]]+-/ {
  if (block != "") {
    block = block $0 "\n"
    if ($0 ~ /\[x\]/) criteria++
    if ($0 ~ /\[ \]/) { criteria++; unchecked++ }
  } else {
    print
  }
  next
}

/^[[:space:]]*$/ {
  if (block != "") flush()
  print
  next
}

{ print }

END { if (in_sec) flush() }
' "$todo_file" > "$tmp_todo"

if [ ! -s "$tmp_names" ]; then
  exit 0
fi

# CHANGELOG.md を更新（今日のセクションがあれば追記、なければ新規作成）
today=$(date +%Y-%m-%d)
if [ -f "$changelog_file" ]; then
  tmp_cl=$(mktemp)
  awk -v today="## $today" -v names_file="$tmp_names" '
  BEGIN {
    done = 0
    while ((getline line < names_file) > 0) {
      entries = entries "- " line "\n"
    }
    close(names_file)
  }
  {
    if (!done && $0 == today) {
      print
      if ((getline next_line) > 0) {
        if (next_line == "") {
          print ""
          printf "%s", entries
        } else {
          print ""
          printf "%s", entries
          print next_line
        }
      }
      done = 1
      next
    }
    if (!done && /^## [0-9]/) {
      print today
      print ""
      printf "%s", entries
      print ""
      done = 1
    }
    print
  }
  ' "$changelog_file" > "$tmp_cl"
  mv "$tmp_cl" "$changelog_file"
fi

mv "$tmp_todo" "$todo_file"

count=$(wc -l < "$tmp_names" | tr -d ' ')
echo "完了済みタスク ${count} 件を CHANGELOG.md に移動しました。"
