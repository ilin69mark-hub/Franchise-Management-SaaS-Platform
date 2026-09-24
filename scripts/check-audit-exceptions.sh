#!/usr/bin/env bash
# Проверка парности маркеров AUDIT-EXCEPTION и реестра .audit-exceptions.yml.
# Fail, если:
#   1) в коде есть маркер с ID, которого нет в реестре (тихий хардкод исключения);
#   2) запись реестра не имеет маркера в указанном файле (протухшая запись);
#   3) pattern записи не найден рядом с маркером (±3 строки) — маркер перетащили
#      на чужой код (строку можно двигать, но вместе с реестром в том же коммите);
#   4) нарушена строгая схема записи.
# Запуск: bash scripts/check-audit-exceptions.sh (из корня репо) или make audit-exceptions.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

REG=".audit-exceptions.yml"
FAIL=0

if [[ ! -f "$REG" ]]; then
  echo "FAIL: реестр $REG не найден" >&2
  exit 1
fi

if ! grep -qE '^version: 1$' "$REG"; then
  echo "FAIL: $REG: нет 'version: 1'" >&2
  exit 1
fi

# --- ID из кода (все стили комментариев: //, #, <!-- -->) ---
CODE_IDS="$(grep -rhoE 'AUDIT-EXCEPTION\(E[0-9]+\)' \
  --exclude-dir=.git --exclude-dir=node_modules . 2>/dev/null \
  | grep -oE 'E[0-9]+' | sort -u || true)"

# --- ID из реестра ---
REG_IDS="$(grep -oE '^[[:space:]]*- id: "E[0-9]+"' "$REG" | grep -oE 'E[0-9]+' | sort -u)"

if [[ -z "$REG_IDS" ]]; then
  echo "FAIL: в $REG нет записей" >&2
  exit 1
fi

# --- 1) каждый маркер обязан иметь запись ---
while read -r id; do
  [[ -z "$id" ]] && continue
  if ! echo "$REG_IDS" | grep -qx "$id"; then
    echo "FAIL: маркер $id в коде без записи в $REG — тихое исключение запрещено" >&2
    grep -rn "AUDIT-EXCEPTION($id)" --exclude-dir=.git --exclude-dir=node_modules . 2>/dev/null | head -n 5 >&2 || true
    FAIL=1
  fi
done <<< "$CODE_IDS"

# --- разбор записей реестра в массивы ---
IDS=()
FILES=()
LINES=()
PATTERNS=()
cur=""
idx=-1
while IFS= read -r line; do
  if [[ "$line" =~ ^[[:space:]]*-[[:space:]]+id:[[:space:]]+\"(E[0-9]+)\" ]]; then
    cur="${BASH_REMATCH[1]}"
    idx=$((idx + 1))
    IDS+=("$cur")
    FILES+=("")
    LINES+=("")
    PATTERNS+=("")
  elif [[ -n "$cur" && "$line" =~ ^[[:space:]]+file:[[:space:]]+\"(.*)\" ]]; then
    FILES[$idx]="${BASH_REMATCH[1]}"
  elif [[ -n "$cur" && "$line" =~ ^[[:space:]]+line:[[:space:]]+([0-9]+) ]]; then
    LINES[$idx]="${BASH_REMATCH[1]}"
  elif [[ -n "$cur" && "$line" =~ ^[[:space:]]+pattern:[[:space:]]+\"(.*)\" ]]; then
    PATTERNS[$idx]="${BASH_REMATCH[1]}"
  fi
done < "$REG"

# --- обязательные ключи каждой записи ---
REQUIRED_KEYS=(title risk reason accepted_by accepted_at remove_when no_expiry)
for id in "${IDS[@]}"; do
  block="$(awk "/- id: \"$id\"/{f=1} f{print} f&&/no_expiry:/{exit}" "$REG")"
  for key in "${REQUIRED_KEYS[@]}"; do
    if ! echo "$block" | grep -qE "^[[:space:]]+$key:"; then
      echo "FAIL: запись $id: нет обязательного ключа '$key'" >&2
      FAIL=1
    fi
  done
done

# --- 2)+3) маркер в файле + pattern рядом ---
for i in "${!IDS[@]}"; do
  id="${IDS[$i]}"
  file="${FILES[$i]}"
  exp_line="${LINES[$i]}"
  pattern="${PATTERNS[$i]}"
  if [[ -z "$file" ]]; then
    echo "FAIL: запись $id: пустой file" >&2
    FAIL=1
    continue
  fi
  if [[ ! -f "$file" ]]; then
    echo "FAIL: запись $id: файл $file не существует" >&2
    FAIL=1
    continue
  fi
  marker_lines="$(grep -n -F "AUDIT-EXCEPTION($id)" "$file" | cut -d: -f1 || true)"
  if [[ -z "$marker_lines" ]]; then
    echo "FAIL: запись $id: маркер не найден в $file" >&2
    FAIL=1
    continue
  fi
  ok_pattern=0
  while read -r n; do
    [[ -z "$n" ]] && continue
    from=$(( n > 3 ? n - 3 : 1 ))
    to=$(( n + 3 ))
    if sed -n "${from},${to}p" "$file" | grep -F -q "$pattern"; then
      ok_pattern=1
      break
    fi
  done <<< "$marker_lines"
  if [[ "$ok_pattern" -ne 1 ]]; then
    echo "FAIL: запись $id: pattern '$pattern' не найден рядом (±3 строки) с маркером в $file" >&2
    FAIL=1
    continue
  fi
  # строка в реестре — для людей/аудита: предупреждаем о дрейфе, но не валим
  # (якорь строгости — pattern, а не номер строки).
  first_marker="$(echo "$marker_lines" | head -n 1)"
  if [[ "$first_marker" != "$exp_line" ]]; then
    echo "WARN: запись $id: маркер сейчас на строке $first_marker, в реестре $exp_line — обновите line в $REG этим же коммитом" >&2
  fi
done

if [[ "$FAIL" -ne 0 ]]; then
  echo "audit-exceptions: FAILED" >&2
  exit 1
fi

echo "audit-exceptions: OK ($(echo "$REG_IDS" | wc -l | tr -d ' ') записей, маркеры парные)"
