fg() {
    local argument dispatch expect_value target status

    if [ "$#" -eq 1 ] && { [ "$1" = "back" ] || [ "$1" = "-" ]; }; then
        if ! popd >/dev/null 2>&1; then
            printf '%s\n' "ForgePath has no previous directory" >&2
            return 1
        fi
        return 0
    fi

    dispatch=0
    expect_value=0
    for argument in "$@"; do
        if [ "$expect_value" -eq 1 ]; then
            expect_value=0
            continue
        fi
        case "$argument" in
            --config|--state|--cache|--icons)
                expect_value=1
                continue
                ;;
            --config=*|--state=*|--cache=*|--icons=*)
                continue
                ;;
            -h|--help|--version)
                dispatch=1
                break
                ;;
            -*)
                continue
                ;;
            list|pick|scan|open|reveal|run|config|workspace|favorite|recent|completion|help)
                dispatch=1
                ;;
        esac
        break
    done
    if [ "$dispatch" -eq 1 ]; then
        command fg "$@"
        return $?
    fi

    target="$(command fg pick --print-path "$@")"
    status=$?

    if [ "$status" -ne 0 ]; then
        return "$status"
    fi

    if [ -n "$target" ]; then
        pushd "$target" >/dev/null
    fi
}
