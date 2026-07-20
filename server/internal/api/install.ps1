$ErrorActionPreference = "Stop"

$repo = "krey-yon/Bitrok"
$version = if ($env:BITROK_VERSION) { $env:BITROK_VERSION } else { "latest" }
$arch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
switch ($arch.ToUpperInvariant()) {
  "AMD64" { $arch = "amd64" }
  "ARM64" { $arch = "arm64" }
  default { throw "bitrok: unsupported architecture: $arch" }
}

if ($version -eq "latest") {
  $url = "https://github.com/$repo/releases/latest/download/bitrok_windows_$arch.zip"
} else {
  $url = "https://github.com/$repo/releases/download/$version/bitrok_windows_$arch.zip"
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("bitrok-" + [guid]::NewGuid())
$archive = Join-Path $tmp "bitrok.zip"
$installDir = if ($env:BITROK_INSTALL_DIR) { $env:BITROK_INSTALL_DIR } else { Join-Path $HOME ".local\bin" }
try {
  New-Item -ItemType Directory -Force -Path $tmp | Out-Null
  Write-Host "bitrok: downloading $version (windows/$arch)..."
  Invoke-WebRequest -Uri $url -OutFile $archive -UseBasicParsing
  Expand-Archive -Path $archive -DestinationPath $tmp -Force
  New-Item -ItemType Directory -Force -Path $installDir | Out-Null
  Copy-Item (Join-Path $tmp "bitrok.exe") (Join-Path $installDir "bitrok.exe") -Force

  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $paths = @($userPath -split ";" | Where-Object { $_ })
  if ($paths -notcontains $installDir) {
    [Environment]::SetEnvironmentVariable("Path", (($paths + $installDir) -join ";"), "User")
    Write-Host "bitrok: added $installDir to your user PATH."
  }
  $env:Path = "$installDir;$env:Path"
  Write-Host "bitrok: installed to $(Join-Path $installDir 'bitrok.exe')"
  Write-Host "bitrok: open a new PowerShell window, then run 'bitrok --help'."
} finally {
  Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}
