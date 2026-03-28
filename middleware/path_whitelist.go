package middleware

import "strings"

// pathMatchesWhitelistEntry 路径与配置项匹配（session 白名单、JWT paths 等共用）：
// - 以 / 结尾的项（如 /openapi/）按前缀匹配其子路径；
// - 不以 / 结尾的项仅匹配自身或「自身 + / + 子路径」，不匹配仅追加字符的兄弟路径（如 /a/b 不匹配 /a/bc）。
func pathMatchesWhitelistEntry(path, entry string) bool {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return false
	}
	if strings.HasSuffix(entry, "/") {
		return path == entry || strings.HasPrefix(path, entry)
	}
	if path == entry {
		return true
	}
	if !strings.HasPrefix(path, entry) {
		return false
	}
	if len(path) > len(entry) {
		return path[len(entry)] == '/'
	}
	return true
}
