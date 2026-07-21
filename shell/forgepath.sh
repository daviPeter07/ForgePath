fp() {
    local target status

    target="$(forgepath pick --print-path "$@")"
    status=$?

    if [ "$status" -ne 0 ]; then
        return "$status"
    fi

    if [ -n "$target" ]; then
        cd -- "$target"
    fi
}
