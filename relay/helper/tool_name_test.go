package helper

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCompressToolName_ShortName(t *testing.T) {
	name := "short_tool_name"
	result := CompressToolName(name)
	assert.Equal(t, name, result, "short name should not be compressed")
}

func TestCompressToolName_Exactly64(t *testing.T) {
	// 构造精确 64 字符的名称
	name := ""
	for i := 0; i < 64; i++ {
		name += "a"
	}
	assert.Equal(t, 64, len(name))

	result := CompressToolName(name)
	assert.Equal(t, name, result, "64 character name should not be compressed")
}

func TestCompressToolName_Exactly64_RealExample(t *testing.T) {
	// 使用真实场景的名称结构，精确 64 字符
	name := "mcp__plugin_test_tool__function_name_with_exact_length_64_chars"
	// 调整长度到精确 64
	for len(name) < 64 {
		name += "x"
	}
	for len(name) > 64 {
		name = name[:len(name)-1]
	}
	assert.Equal(t, 64, len(name))

	result := CompressToolName(name)
	assert.Equal(t, name, result, "exactly 64 character name should not be compressed")
}

func TestCompressToolName_LongName(t *testing.T) {
	// MCP 工具名称示例 (70 字符)
	name := "mcp__plugin_everything-claude-code_sequential-thinking__sequentialthinking"
	assert.Greater(t, len(name), MaxToolNameLength, "test name should exceed 64 chars")

	result := CompressToolName(name)

	assert.LessOrEqual(t, len(result), MaxToolNameLength, "compressed name should not exceed 64 chars")
	assert.Contains(t, result, "#", "compressed name should contain hash separator")
	assert.NotEqual(t, name, result, "long name should be compressed")
}

func TestCompressToolName_VeryLongName(t *testing.T) {
	// 构造超长名称 (100 字符)
	name := ""
	for i := 0; i < 100; i++ {
		name += "a"
	}

	result := CompressToolName(name)

	assert.Equal(t, MaxToolNameLength, len(result), "compressed name should be exactly 64 chars")
	assert.Contains(t, result, "#", "compressed name should contain hash separator")
}

func TestCompressToolName_Uniqueness(t *testing.T) {
	// 两个不同但前缀相同的名称应该产生不同的压缩结果
	basePrefix := "mcp__plugin_one_tool__function_"

	// 延长到超过 64 字符
	name1 := basePrefix + "a" + "_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	name2 := basePrefix + "b" + "_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	// 确保超长
	for len(name1) <= MaxToolNameLength {
		name1 += "x"
		name2 += "x"
	}

	result1 := CompressToolName(name1)
	result2 := CompressToolName(name2)

	assert.NotEqual(t, result1, result2, "different names should produce different compressed names")
	assert.LessOrEqual(t, len(result1), MaxToolNameLength)
	assert.LessOrEqual(t, len(result2), MaxToolNameLength)
}

func TestCompressToolName_Deterministic(t *testing.T) {
	name := "mcp__plugin_test_tool__function_with_long_name_xxxxxxxxxxxxxx"
	for len(name) <= MaxToolNameLength {
		name += "x"
	}

	result1 := CompressToolName(name)
	result2 := CompressToolName(name)

	assert.Equal(t, result1, result2, "same name should produce same result (deterministic)")
}

func TestCompressToolName_HashSuffixFormat(t *testing.T) {
	name := "mcp__plugin_test_tool__very_long_function_name_exceeding_limit_xxxxxx"
	for len(name) <= MaxToolNameLength {
		name += "x"
	}

	result := CompressToolName(name)

	// 验证 hash 后缀格式：# + 7字符 hex
	hashSuffix := result[len(result)-ToolNameHashSuffixLength:]
	assert.Equal(t, "#", hashSuffix[:1], "hash suffix should start with #")
	// 后面 7 字符应该是 hex 编码
	for i, c := range hashSuffix[1:] {
		assert.True(t, isHexChar(c), "character at position %d should be hex char: %c", i, c)
	}
}

func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func TestCompressToolNames(t *testing.T) {
	names := []string{
		"short_tool",
		"mcp__plugin_everything-claude-code_sequential-thinking__sequentialthinking",
		"another_long_tool_name_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}

	// 确保测试名称超长
	for i := 1; i < len(names); i++ {
		for len(names[i]) <= MaxToolNameLength {
			names[i] += "x"
		}
	}

	compressed, mapping := CompressToolNames(names)

	assert.Equal(t, len(names), len(compressed), "should have same number of names")
	assert.Equal(t, names[0], compressed[0], "short name should not be compressed")

	// 检查映射表
	for original, compressedName := range mapping {
		assert.LessOrEqual(t, len(compressedName), MaxToolNameLength)
		assert.NotEqual(t, original, compressedName)
	}
}

func TestCompressToolNames_Empty(t *testing.T) {
	names := []string{}
	compressed, mapping := CompressToolNames(names)

	assert.Equal(t, 0, len(compressed))
	assert.Equal(t, 0, len(mapping))
}

func TestCompressToolNames_AllShort(t *testing.T) {
	names := []string{"tool1", "tool2", "tool3"}
	compressed, mapping := CompressToolNames(names)

	assert.Equal(t, names, compressed)
	assert.Equal(t, 0, len(mapping), "no mapping needed for short names")
}

func TestToolNameMapping_Context(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	mapping := ToolNameMapping{
		"long_original_name": "short_compressed#abc1234",
	}

	StoreToolNameMapping(c, mapping)

	retrieved := GetToolNameMapping(c)
	assert.NotNil(t, retrieved)
	assert.Equal(t, mapping["long_original_name"], retrieved["long_original_name"])

	original := GetOriginalToolName(c, "short_compressed#abc1234")
	assert.Equal(t, "long_original_name", original)
}

func TestToolNameMapping_ContextMerge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	mapping1 := ToolNameMapping{
		"name1": "compressed1#abc",
	}
	mapping2 := ToolNameMapping{
		"name2": "compressed2#def",
	}

	StoreToolNameMapping(c, mapping1)
	StoreToolNameMapping(c, mapping2)

	retrieved := GetToolNameMapping(c)
	assert.NotNil(t, retrieved)
	assert.Equal(t, 2, len(retrieved))
	assert.Equal(t, "compressed1#abc", retrieved["name1"])
	assert.Equal(t, "compressed2#def", retrieved["name2"])
}

func TestToolNameMapping_NilContext(t *testing.T) {
	retrieved := GetToolNameMapping(nil)
	assert.Nil(t, retrieved)
}

func TestToolNameMapping_ContextNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	retrieved := GetToolNameMapping(c)
	assert.Nil(t, retrieved)
}

func TestGetOriginalToolName_NotInMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	mapping := ToolNameMapping{
		"other_name": "compressed#abc",
	}
	StoreToolNameMapping(c, mapping)

	// 查询不存在的名称
	result := GetOriginalToolName(c, "not_compressed_name")
	assert.Equal(t, "not_compressed_name", result, "should return input if not in mapping")
}

func TestIsToolNameCompressed(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"short", false},
		{"tool#abc", true},     // 短但有 #
		{"tool_without_hash", false},
	}

	for _, tt := range tests {
		result := IsToolNameCompressed(tt.name)
		assert.Equal(t, tt.expected, result, "test: %s", tt.name)
	}
}