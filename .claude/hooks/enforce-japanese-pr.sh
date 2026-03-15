#!/usr/bin/env bash
# PR作成時にタイトル・本文が日本語で記述されているかチェックするフック
# Claude Code PreToolUse フックとして動作する

set -euo pipefail

# stdinからツール入力JSONを読み込む
INPUT=$(cat)

# Bashツールのcommandを取得
COMMAND=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('tool_input',{}).get('command',''))" 2>/dev/null || true)

# gh pr create コマンドでなければスキップ
if ! echo "$COMMAND" | grep -q "gh pr create"; then
  exit 0
fi

# タイトルを抽出（--title の値）
TITLE=$(echo "$COMMAND" | python3 -c "
import sys, re
cmd = sys.stdin.read()
m = re.search(r'--title\s+[\"\'](.*?)[\"\']', cmd, re.DOTALL)
if m:
    print(m.group(1))
" 2>/dev/null || true)

# ボディを抽出（--body の値）
BODY=$(echo "$COMMAND" | python3 -c "
import sys, re
cmd = sys.stdin.read()
m = re.search(r'--body\s+[\"\'](.*?)[\"\']', cmd, re.DOTALL)
if m:
    print(m.group(1))
" 2>/dev/null || true)

# 日本語文字（ひらがな・カタカナ・漢字）が含まれているか確認する関数
contains_japanese() {
  local text="$1"
  echo "$text" | python3 -c "
import sys, re
text = sys.stdin.read()
# ひらがな、カタカナ、CJK統一漢字をチェック
has_japanese = bool(re.search(r'[\u3040-\u309f\u30a0-\u30ff\u4e00-\u9fff]', text))
sys.exit(0 if has_japanese else 1)
"
}

ERRORS=()

# タイトルチェック
if [ -n "$TITLE" ]; then
  if ! contains_japanese "$TITLE"; then
    ERRORS+=("タイトルが日本語で記述されていません: \"$TITLE\"")
  fi
fi

# ボディチェック
if [ -n "$BODY" ]; then
  if ! contains_japanese "$BODY"; then
    ERRORS+=("本文が日本語で記述されていません")
  fi
fi

# エラーがあればブロック
if [ ${#ERRORS[@]} -gt 0 ]; then
  echo "❌ PR作成をブロックしました: PRのタイトル・本文は必ず日本語で記述してください。" >&2
  echo "" >&2
  for err in "${ERRORS[@]}"; do
    echo "  - $err" >&2
  done
  echo "" >&2
  echo "rules/language.md を参照し、日本語でPRを作成してください。" >&2
  exit 2
fi

exit 0
