fg() {
    local argument dispatch executable expect_value target status

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
            --config|--state|--cache|--icons|--theme)
                expect_value=1
                continue
                ;;
            --config=*|--state=*|--cache=*|--icons=*|--theme=*)
                continue
                ;;
            -h|--help|--version)
                dispatch=1
                break
                ;;
            -*)
                continue
                ;;
            list|pick|scan|open|reveal|run|config|workspace|favorite|recent|clone|completion|help)
                dispatch=1
                ;;
        esac
        break
    done
    # fg is also a Bash builtin, and `command fg` would still run the builtin,
    # so resolve the ForgePath executable from PATH explicitly.
    # Every failure is captured with `||` so the wrapper also works under set -e.
    executable="$(type -P fg)" || executable=""
    if [ -z "$executable" ]; then
        printf '%s\n' "ForgePath executable 'fg' was not found in PATH" >&2
        return 127
    fi

    if [ "$dispatch" -eq 1 ]; then
        "$executable" "$@" || return $?
        return 0
    fi

    status=0
    target="$("$executable" pick --print-path "$@")" || status=$?

    if [ "$status" -ne 0 ]; then
        return "$status"
    fi

    if [ -n "$target" ]; then
        pushd "$target" >/dev/null
    fi
}
