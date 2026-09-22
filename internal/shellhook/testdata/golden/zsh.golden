# Snapback shell hook for zsh: tells the daemon when the shell enters a new
# physical directory. Load it with: eval "$(snapback shell-hook zsh)"

__snapback_hook() {
  local rc=$?
  # Local to this function: the user's options are restored on return.
  emulate -L zsh
  local dir
  # The trailing x keeps a directory name that ends in a newline intact.
  dir=$(builtin pwd -P 2>/dev/null && print -n x) || return $rc
  dir=${dir%$'\n'x}
  if [[ $dir != ${__snapback_last_dir-} ]]; then
    typeset -g __snapback_last_dir=$dir
    snapback notify --timeout 200ms --session $$ -- $dir </dev/null >/dev/null 2>&1 &!
  fi
  return $rc
}

autoload -Uz add-zsh-hook
add-zsh-hook precmd __snapback_hook
