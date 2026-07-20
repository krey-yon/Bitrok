const script = `$ErrorActionPreference = "Stop"
$repo = "krey-yon/Bitrok"
$version = if ($env:BITROK_VERSION) { $env:BITROK_VERSION } else { "latest" }
if ($version -ne "latest" -and $version -notmatch '^v\\d+\\.\\d+\\.\\d+(?:[-+][0-9A-Za-z.-]+)?$') {
  throw "bitrok: invalid version: $version"
}
$arch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
switch ($arch.ToUpperInvariant()) {
  "AMD64" { $arch = "amd64" }
  "ARM64" { $arch = "arm64" }
  default { throw "bitrok: unsupported architecture: $arch" }
}
$archiveName = "bitrok_windows_$arch.zip"
$baseUrl = if ($version -eq "latest") {
  "https://github.com/$repo/releases/latest/download"
} else {
  "https://github.com/$repo/releases/download/$version"
}
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("bitrok-" + [guid]::NewGuid())
$archive = Join-Path $tmp "bitrok.zip"
$checksums = Join-Path $tmp "checksums.txt"
$installDir = if ($env:BITROK_INSTALL_DIR) { $env:BITROK_INSTALL_DIR } else { Join-Path $HOME ".local\bin" }
try {
  New-Item -ItemType Directory -Force -Path $tmp | Out-Null
  Write-Host "bitrok: downloading $version (windows/$arch)..."
  Invoke-WebRequest -Uri "$baseUrl/$archiveName" -OutFile $archive -UseBasicParsing
  Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile $checksums -UseBasicParsing
  $checksumLine = Get-Content $checksums | Where-Object { $_ -match "^[0-9a-fA-F]{64}\\s+$([regex]::Escape($archiveName))$" } | Select-Object -First 1
  if (-not $checksumLine) { throw "bitrok: release checksum for $archiveName was not found" }
  $expected = ($checksumLine -split '\\s+')[0].ToLowerInvariant()
  $actual = (Get-FileHash -Path $archive -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($actual -ne $expected) { throw "bitrok: checksum verification failed for $archiveName" }
  Expand-Archive -Path $archive -DestinationPath $tmp -Force
  if (-not (Test-Path (Join-Path $tmp "bitrok.exe") -PathType Leaf)) {
    throw "bitrok: release archive did not contain bitrok.exe"
  }
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
`;

export function GET() {
  return new Response(script, {
    headers: {
      "Cache-Control": "public, max-age=300",
      "Content-Disposition": 'inline; filename="install.ps1"',
      "Content-Type": "text/plain; charset=utf-8",
    },
  });
}
