package ui

import "fmt"

// FormatSpeed formats transfer speed for menus (B/s, KB/s, MB/s).
func FormatSpeed(bytesPerSec int64) string {
	if bytesPerSec == 0 {
		return Msg(MsgSpeedZero)
	}
	if bytesPerSec < 1024 {
		return Msgf(MsgSpeedBpsFmt, bytesPerSec)
	}
	if bytesPerSec < 1024*1024 {
		return Msgf(MsgSpeedKBpsFmt, float64(bytesPerSec)/1024)
	}

	return Msgf(MsgSpeedMBpsFmt, float64(bytesPerSec)/(1024*1024))
}

// FormatSpeedLimit returns formatted speed limit (MB/s) or empty string if no limit.
func FormatSpeedLimit(bytesPerSec int64) string {
	if bytesPerSec == 0 {
		return ""
	}

	mbPerSec := float64(bytesPerSec) / (1024 * 1024)

	return Msgf(MsgSpeedLimitMBpsFmt, mbPerSec)
}

// FormatBytes formats bytes using 1024 base and "KB/MB/..." suffixes.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSpeedBytes formats speed as human-readable bytes plus "/s".
func FormatSpeedBytes(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return Msgf(MsgSpeedBpsFmt, bytesPerSec)
	}

	return FormatBytes(bytesPerSec) + Msg(MsgTorrentProgressSpeedSuffixPerSec)
}
