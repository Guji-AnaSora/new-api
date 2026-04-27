package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	// MaxToolNameLength 上游 API 对 tool name 的最大长度限制
	// Anthropic、OpenAI、Azure、BigModel 均为 64 字符
	MaxToolNameLength = 64

	// ToolNameHashSuffixLength 用于生成唯一后缀的 hash 长度
	// 格式: truncated_name + "#" + hash_suffix (共 64 字符)
	ToolNameHashSuffixLength = 8

	// ToolNameMappingKey 用于在 gin.Context 中存储名称映射表
	ToolNameMappingKey = "tool_name_mapping"
)

// ToolNameMapping 存储原始名称到截断名称的映射
type ToolNameMapping map[string]string // originalName -> compressedName

// truncateAtUTF8Boundary 截断字符串到最大字节数，确保不在多字节 UTF-8 字符中间切割
func truncateAtUTF8Boundary(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	for maxBytes > 0 && !utf8.RuneStart(s[maxBytes]) {
		maxBytes--
	}
	return s[:maxBytes]
}

// CompressToolName 压缩超长的工具名称
// 如果名称长度 <= 64，直接返回原名称
// 如果名称长度 > 64，使用截断 + hash 后缀策略确保唯一性
func CompressToolName(name string) string {
	if len(name) <= MaxToolNameLength {
		return name
	}

	hash := sha256.Sum256([]byte(name))
	// 生成 7 字符 hash + "#" = 8 字符后缀
	hashSuffix := "#" + hex.EncodeToString(hash[:])[:ToolNameHashSuffixLength-1]

	truncatedLength := MaxToolNameLength - ToolNameHashSuffixLength
	truncatedName := truncateAtUTF8Boundary(name, truncatedLength)

	return truncatedName + hashSuffix
}

// CompressToolNames 批量压缩工具名称，并返回映射表
// 包含碰撞检测：当两个名称产生相同压缩结果时，自动扩展 hash 后缀
func CompressToolNames(names []string) (compressedNames []string, mapping ToolNameMapping) {
	mapping = make(ToolNameMapping)
	compressedNames = make([]string, len(names))
	seen := make(map[string]string) // compressedName -> originalName

	for i, name := range names {
		compressed := CompressToolName(name)
		if compressed != name {
			if orig, exists := seen[compressed]; exists && orig != name {
				compressed = resolveCollision(name, seen)
			}
			seen[compressed] = name
			mapping[name] = compressed
		}
		compressedNames[i] = compressed
	}

	return compressedNames, mapping
}

// resolveCollision 碰撞时扩展 hash 后缀直到找到唯一名称
func resolveCollision(name string, seen map[string]string) string {
	hash := sha256.Sum256([]byte(name))
	hashHex := hex.EncodeToString(hash[:])

	for hashLen := ToolNameHashSuffixLength + 1; hashLen < MaxToolNameLength; hashLen++ {
		hashSuffix := "#" + hashHex[:hashLen-1]
		truncLen := MaxToolNameLength - hashLen
		if truncLen <= 0 {
			break
		}
		truncated := truncateAtUTF8Boundary(name, truncLen)
		candidate := truncated + hashSuffix
		if _, exists := seen[candidate]; !exists {
			return candidate
		}
	}
	// Fallback: 使用完整 hash（实际不会触发）
	return hashHex[:MaxToolNameLength]
}

// StoreToolNameMapping 将名称映射表存储到 gin.Context 中
func StoreToolNameMapping(c *gin.Context, mapping ToolNameMapping) {
	if len(mapping) == 0 {
		return
	}

	// 合并已有的映射（处理多次调用的场景）
	existing := GetToolNameMapping(c)
	if existing != nil {
		for k, v := range mapping {
			existing[k] = v
		}
		c.Set(ToolNameMappingKey, existing)
	} else {
		c.Set(ToolNameMappingKey, mapping)
	}
}

// GetToolNameMapping 从 gin.Context 获取名称映射表
func GetToolNameMapping(c *gin.Context) ToolNameMapping {
	if c == nil {
		return nil
	}
	val, exists := c.Get(ToolNameMappingKey)
	if !exists {
		return nil
	}
	mapping, ok := val.(ToolNameMapping)
	if !ok {
		return nil
	}
	return mapping
}

// GetOriginalToolName 根据截断后的名称获取原始名称（用于日志/调试）
func GetOriginalToolName(c *gin.Context, compressedName string) string {
	mapping := GetToolNameMapping(c)
	if mapping == nil {
		return compressedName
	}

	// 反向查找：遍历映射表找到原始名称
	for original, compressed := range mapping {
		if compressed == compressedName {
			return original
		}
	}

	return compressedName
}

// IsToolNameCompressed 判断工具名称是否已被压缩
// 检查末尾是否为 "# + 7位hex" 的压缩后缀格式
func IsToolNameCompressed(name string) bool {
	if len(name) > MaxToolNameLength || len(name) < ToolNameHashSuffixLength {
		return false
	}
	suffix := name[len(name)-ToolNameHashSuffixLength:]
	if suffix[0] != '#' {
		return false
	}
	for _, c := range suffix[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}