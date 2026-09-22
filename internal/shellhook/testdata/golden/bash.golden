# Snapback shell hook for bash: tells the daemon when the shell enters a new
# physical directory. Load it with: eval "$(snapback shell-hook bash)"

__snapback_hook() {
  local rc=${1:-$?} dir
  # The trailing x keeps a directory name that ends in a newline intact.
  dir=$(builtin pwd -P 2>/dev/null && printf x) || return "$rc"
  dir=${dir%$'\n'x}
  if [[ $dir != "${__snapback_last_dir-}" ]]; then
    __snapback_last_dir=$dir
    (snapback notify --timeout 200ms --session "$$" -- "$dir" </dev/null >/dev/null 2>&1 &)
  fi
  return "$rc"
}

if [[ $(declare -p PROMPT_COMMAND 2>/dev/null) == "declare -a"* ]]; then
  PROMPT_COMMAND+=('__snapback_hook')
else
  # Save $? before the user's command runs so the prompt still sees it.
  printf -v PROMPT_COMMAND "__snapback_rc=\$?\n%s\n__snapback_hook \"\$__snapback_rc\"" "${PROMPT_COMMAND-}"
fi
