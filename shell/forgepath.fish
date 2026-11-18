function fg --description 'ForgePath: browse projects and cd into them'
    # The directory stack only uses builtins, so it works on any fish install
    # (pushd/popd are autoloaded functions that may be missing).
    if test (count $argv) -eq 1; and contains -- $argv[1] back -
        if test (count $__forgepath_dirs) -eq 0
            printf '%s\n' "ForgePath has no previous directory" >&2
            return 1
        end
        set -l previous $__forgepath_dirs[-1]
        set -e __forgepath_dirs[-1]
        cd $previous
        return $status
    end

    set -l dispatch 0
    set -l expect_value 0
    for argument in $argv
        if test $expect_value -eq 1
            set expect_value 0
            continue
        end
        switch $argument
            case --config --state --cache --icons --theme
                set expect_value 1
                continue
            case '--config=*' '--state=*' '--cache=*' '--icons=*' '--theme=*'
                continue
            case -h --help --version
                set dispatch 1
                break
            case '-*'
                continue
            case list pick scan open reveal run config workspace favorite recent clone completion help
                set dispatch 1
        end
        break
    end

    # fg is also a fish builtin (job control); `command` runs only the
    # ForgePath executable found in PATH.
    if not command -q fg
        printf '%s\n' "ForgePath executable 'fg' was not found in PATH" >&2
        return 127
    end

    if test $dispatch -eq 1
        command fg $argv
        return $status
    end

    set -l target (command fg pick --print-path $argv)
    set -l code $status
    if test $code -ne 0
        return $code
    end

    if test -n "$target"
        set -l previous $PWD
        cd $target; or return $status
        set -g __forgepath_dirs $__forgepath_dirs $previous
    end
end
