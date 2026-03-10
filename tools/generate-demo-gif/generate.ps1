$ErrorActionPreference = 'Stop'

$magickCandidates = @(
    'C:\Program Files\ImageMagick-7.1.2-Q16-HDRI\magick.exe',
    'C:\Program Files\ImageMagick-7.1.2-16-Q16-HDRI\magick.exe',
    'magick'
)

$magick = $null
foreach ($candidate in $magickCandidates) {
    if ($candidate -eq 'magick') {
        $command = Get-Command magick -ErrorAction SilentlyContinue
        if ($command) {
            $magick = $command.Source
            break
        }
        continue
    }

    if (Test-Path $candidate) {
        $magick = $candidate
        break
    }
}

if (-not $magick) {
    throw 'ImageMagick executable not found.'
}

$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$assetDir = Join-Path $repoRoot 'docs\assets'
$framesDir = Join-Path $assetDir 'tmp-demo-frames'
$outputPath = Join-Path $assetDir 'arpmap-demo.gif'

New-Item -ItemType Directory -Force -Path $assetDir | Out-Null
if (Test-Path $framesDir) {
    Remove-Item $framesDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $framesDir | Out-Null

$sceneData = @(
    @{
        Kind = 'terminal'
        Static = $true
        HoldDelay = 800
        Lines = @(
            'arpmap --help',
            '',
            'ARP-based local network scanner',
            '',
            'Commands:',
            '  scan    Discover active hosts and their MAC addresses',
            '  find    Report candidate free IP addresses'
        )
    },
    @{
        Kind = 'terminal'
        RevealDelay = 55
        HoldDelay = 400
        Lines = @(
            'sudo arpmap scan --interface eth0 --debug --output devices.json',
            '',
            '[INFO] Starting ARP scan on interface=eth0 subnets=192.168.1.0/24',
            '[DEBUG] Scan started | interface=eth0 mac=48:21:0b:22:7f:31 subnets=192.168.1.0/24',
            '[DEBUG] Dispatch settings | workers=256 source=auto',
            '[DEBUG] Scan parameters | targets=254 attempts=1',
            '[DEBUG] ARP response | sender_ip=192.168.1.1 sender_mac=dc:a6:32:00:11:02 frame_size=42'
        )
    },
    @{
        Kind = 'terminal'
        RevealDelay = 50
        HoldDelay = 400
        Lines = @(
            'sudo arpmap scan --interface eth0 --debug --output devices.json',
            '',
            '[INFO] Starting ARP scan on interface=eth0 subnets=192.168.1.0/24',
            '[DEBUG] ARP response | sender_ip=192.168.1.10 sender_mac=48:21:0b:22:7f:31 frame_size=42',
            '[DEBUG] ARP response | sender_ip=192.168.1.44 sender_mac=84:3a:4b:10:45:99 frame_size=42',
            '[DEBUG] Reader summary | reads=23 timeouts=11 unique_devices=18',
            '[DEBUG] Scan completed | duration=2.14s dispatched=254 total_targets=254',
            '[DEBUG] Response metrics | responded=18 dispatched=254 response_rate=7.1%',
            '[DEBUG] Sample IP addresses responding to ARP requests: [192.168.1.1 192.168.1.10 192.168.1.44]',
            '[DEBUG] Sample IP addresses with no ARP response: [192.168.1.120 192.168.1.121 192.168.1.122]',
            '[DEBUG] Scan finished'
        )
    },
    @{
        Kind = 'terminal'
        RevealDelay = 55
        HoldDelay = 400
        Lines = @(
            'sudo arpmap scan --interface eth0 --debug --output devices.json',
            '',
            '[INFO] Completed ARP scan on interface=eth0 discovered_devices=18',
            '[INFO] Scan results written to devices.json',
            '[INFO] Output is ready for automation'
        )
    },
    @{
        Kind = 'terminal'
        RevealDelay = 55
        HoldDelay = 400
        Lines = @(
            'sudo arpmap find --interface eth0 --count 10 --output free_ips.json',
            '',
            '[INFO] Starting free-IP discovery on interface=eth0 subnets=192.168.1.0/24',
            '[DEBUG] Scan started | interface=eth0 mac=48:21:0b:22:7f:31 subnets=192.168.1.0/24',
            '[DEBUG] Dispatch settings | workers=256 source=auto',
            '[DEBUG] Scan parameters | targets=254 attempts=1',
            '[DEBUG] Reader summary | reads=23 timeouts=11 unique_devices=18'
        )
    },
    @{
        Kind = 'terminal'
        RevealDelay = 55
        HoldDelay = 400
        Lines = @(
            'sudo arpmap find --interface eth0 --count 10 --output free_ips.json',
            '',
            '[DEBUG] Scan completed | duration=2.07s dispatched=254 total_targets=254',
            '[DEBUG] Response metrics | responded=18 dispatched=254 response_rate=7.1%',
            '[DEBUG] Sample IP addresses responding to ARP requests: [192.168.1.1 192.168.1.10 192.168.1.44]',
            '[DEBUG] Sample IP addresses with no ARP response: [192.168.1.120 192.168.1.121 192.168.1.122]',
            '[DEBUG] Scan finished',
            '[INFO] Completed free-IP discovery on interface=eth0 free_addresses=10',
            '[INFO] Free-IP results written to free_ips.json',
            '[INFO] Output is ready for automation'
        )
    }
)

function New-Frame {
    param(
        [string[]]$Lines,
	    [int]$Index,
	    [int]$VisibleThrough,
	    [bool]$ShowCursor
    )

    $framePath = Join-Path $framesDir ("frame-{0:D3}.png" -f $Index)
    $drawArgs = @(
        '-size', '1400x820',
        'xc:#07111b',
        '-fill', '#0d1724',
        '-draw', 'roundrectangle 30,30 1370,790 28,28',
        '-fill', '#132238',
        '-draw', 'roundrectangle 30,30 1370,92 28,28',
        '-fill', '#ff6b6b',
        '-draw', 'circle 72,61 72,52',
        '-fill', '#ffd166',
        '-draw', 'circle 102,61 102,52',
        '-fill', '#06d6a0',
        '-draw', 'circle 132,61 132,52',
        '-font', 'Consolas',
        '-fill', '#dbe7f5',
        '-pointsize', '24',
        '-annotate', '+180+68', 'arpmap demo',
        '-fill', '#9ef01a',
        '-pointsize', '30'
    )

    $y = 145
    for ($i = 0; $i -le $VisibleThrough; $i++) {
        $line = $Lines[$i]
        $fill = '#c8d6e5'
        $pointSize = '24'

        if ($line -eq '') {
            $y += 18
            continue
        }

        if ($i -eq 0) {
            $fill = '#9ef01a'
            $pointSize = '28'
        } elseif ($line.StartsWith('[INFO]')) {
            $fill = '#5eead4'
        } elseif ($line.StartsWith('[DEBUG]')) {
            $fill = '#facc15'
        } elseif ($line.StartsWith('{') -or $line.StartsWith('  "') -or $line.StartsWith('}')) {
            $fill = '#93c5fd'
        }

        if ($i -eq $VisibleThrough) {
            $highlightTop = [Math]::Max($y - 32, 108)
            $highlightBottom = $y + 16
            $drawArgs += @('-fill', '#17304d', '-draw', "roundrectangle 52,$highlightTop 1330,$highlightBottom 12,12")
        }

        $renderedLine = $line
        if ($ShowCursor -and $i -eq $VisibleThrough) {
            $renderedLine = "$line _"
        }

        $drawArgs += @('-fill', $fill, '-pointsize', $pointSize, '-annotate', "+70+$y", $renderedLine)
        $y += 48
    }

    & $magick @drawArgs $framePath
    return $framePath
}

$frameEntries = @()
$frameIndex = 1

foreach ($scene in $sceneData) {
    if ($scene.Static) {
        $framePath = New-Frame -Lines $scene.Lines -Index $frameIndex -VisibleThrough ($scene.Lines.Count - 1) -ShowCursor $false
        $frameEntries += @{
            Path = $framePath
            Delay = $scene.HoldDelay
        }
        $frameIndex++
        continue
    }

    $visibleLineIndexes = @()
    for ($lineIndex = 0; $lineIndex -lt $scene.Lines.Count; $lineIndex++) {
        $visibleLineIndexes += $lineIndex
    }

    for ($step = 0; $step -lt $visibleLineIndexes.Count; $step++) {
        $visibleThrough = $visibleLineIndexes[$step]
        $isLastStep = $step -eq ($visibleLineIndexes.Count - 1)
        $framePath = New-Frame -Lines $scene.Lines -Index $frameIndex -VisibleThrough $visibleThrough -ShowCursor (-not $isLastStep)
        $frameEntries += @{
            Path = $framePath
            Delay = if ($isLastStep) { $scene.HoldDelay } else { $scene.RevealDelay }
        }
        $frameIndex++
    }
}

$gifArgs = @()
foreach ($frame in $frameEntries) {
    $gifArgs += @('-delay', [string]$frame.Delay, $frame.Path)
}
$gifArgs += @('-loop', '0', '-layers', 'Optimize', $outputPath)
& $magick @gifArgs

Remove-Item $framesDir -Recurse -Force
Write-Host "Created $outputPath"