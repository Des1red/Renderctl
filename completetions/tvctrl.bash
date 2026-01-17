_renderctl() {
  local cur
  cur="${COMP_WORDS[COMP_CWORD]}"

  opts="--probe-only --mode --auto-cache --no-cache --list-cache \
        --forget-cache --select-cache --subnet --deep-search --ssdp \
        --Tip --Tport --Tpath --type --Lf --Lip --Ldir --LPort --version"

  COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
}

complete -F _renderctl renderctl
