package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

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

// CompressToolName 压缩超长的工具名称
// 如果名称长度 <= 64，直接返回原名称
// 如果名称长度 > 64，使用截断 + hash 后缀策略确保唯一性
func CompressToolName(name string) string {
	if len(name) <= MaxToolNameLength {
		return name
	}

	// 计算原始名称的 hash 作为唯一标识
	hash := sha256.Sum256([]byte(name))
	// 生成 7 字符 hash + "#" = 8 字符后缀
	hashSuffix := "#" + hex.EncodeToString(hash[:])[:ToolNameHashSuffixLength-1]

	// 截断名称保留有意义的前缀部分
	// 截断长度 = 64 - hash后缀长度
	truncatedLength := MaxToolNameLength - ToolNameHashSuffixLength
	truncatedName := name[:truncatedLength]

	return truncatedName + hashSuffix
}

// CompressToolNames 批量压缩工具名称，并返回映射表
func CompressToolNames(names []string) (compressedNames []string, mapping ToolNameMapping) {
	mapping = make(ToolNameMapping)
	compressedNames = make([]string, len(names))

	for i, name := range names {
		compressed := CompressToolName(name)
		compressedNames[i] = compressed
		if compressed != name {
			mapping[name] = compressed
		}
	}

	return compressedNames, mapping
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
func IsToolNameCompressed(name string) bool {
	return len(name) <= MaxToolNameLength && strings.Contains(name, "#")
}