function fg {
    if ($args.Count -eq 1 -and $args[0] -in @("back", "-")) {
        if ((Get-Location -Stack -ErrorAction SilentlyContinue).Count -eq 0) {
            Write-Error "ForgePath has no previous directory"
            return
        }
        try {
            Pop-Location -ErrorAction Stop
        }
        catch {
            Write-Error "ForgePath has no previous directory"
        }
        return
    }

    $commands = @("list", "pick", "scan", "open", "reveal", "run", "config", "workspace", "favorite", "recent", "completion", "help")
    $dispatch = $false
    $expectValue = $false
    foreach ($argument in $args) {
        if ($expectValue) {
            $expectValue = $false
            continue
        }
        if ($argument -in @("--config", "--state", "--cache", "--icons")) {
            $expectValue = $true
            continue
        }
        if ($argument -match '^--(config|state|cache|icons)=') {
            continue
        }
        if ($argument -in @("-h", "--help", "--version")) {
            $dispatch = $true
            break
        }
        if ($argument.StartsWith("-")) {
            continue
        }
        if ($commands -contains $argument) {
            $dispatch = $true
        }
        break
    }
    $applicationName = if ($env:OS -eq "Windows_NT") { "fg.exe" } else { "fg" }
    $executable = @(Get-Command $applicationName -CommandType Application -ErrorAction Stop)[0].Source
    if ($dispatch) {
        & $executable @args
        return
    }

    $previousOutputEncoding = [Console]::OutputEncoding

    try {
        [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
        $target = & $executable pick --print-path @args
        $exitCode = $LASTEXITCODE
    }
    finally {
        [Console]::OutputEncoding = $previousOutputEncoding
    }

    if ($exitCode -ne 0) {
        Write-Error "ForgePath failed with exit code $exitCode"
        return
    }

    if ($target) {
        Push-Location -LiteralPath $target
    }
}
