# Snapback shell hook for fish: tells the daemon when the shell enters a new
# physical directory. Load it with: snapback shell-hook fish | source

function __snapback_hook --on-variable PWD
    set -l rc $status
    # read -z keeps every newline; the split then drops only the one pwd
    # adds, and as the last step of the substitution it keeps its fields whole.
    builtin pwd -P 2>/dev/null | read -lz dir
    or return $rc
    set dir (string split -r -m1 \n -- $dir)[1]
    if test "$dir" != "$__snapback_last_dir"
        set -g __snapback_last_dir $dir
        command snapback notify --timeout 200ms --session $fish_pid -- $dir </dev/null >/dev/null 2>&1 & disown $last_pid 2>/dev/null
    end
    return $rc
end

function __snapback_first_prompt --on-event fish_prompt
    functions -e __snapback_first_prompt
    __snapback_hook
end
