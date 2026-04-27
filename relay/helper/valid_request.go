package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"
	"github.com/samber/lo"

	"github.com/gin-gonic/gin"
)

func GetAndValidateRequest(c *gin.Context, format types.RelayFormat) (request dto.Request, err error) {
	relayMode := relayconstant.Path2RelayMode(c.Request.URL.Path)

	switch format {
	case types.RelayFormatOpenAI:
		request, err = GetAndValidateTextRequest(c, relayMode)
	case types.RelayFormatGemini:
		if strings.Contains(c.Request.URL.Path, ":embedContent") {
			request, err = GetAndValidateGeminiEmbeddingRequest(c)
		} else if strings.Contains(c.Request.URL.Path, ":batchEmbedContents") {
			request, err = GetAndValidateGeminiBatchEmbeddingRequest(c)
		} else {
			request, err = GetAndValidateGeminiRequest(c)
		}
	case types.RelayFormatClaude:
		request, err = GetAndValidateClaudeRequest(c)
	case types.RelayFormatOpenAIResponses:
		request, err = GetAndValidateResponsesRequest(c)
	case types.RelayFormatOpenAIResponsesCompaction:
		request, err = GetAndValidateResponsesCompactionRequest(c)

	case types.RelayFormatOpenAIImage:
		request, err = GetAndValidOpenAIImageRequest(c, relayMode)
	case types.RelayFormatEmbedding:
		request, err = GetAndValidateEmbeddingRequest(c, relayMode)
	case types.RelayFormatRerank:
		request, err = GetAndValidateRerankRequest(c)
	case types.RelayFormatOpenAIAudio:
		request, err = GetAndValidAudioRequest(c, relayMode)
	case types.RelayFormatOpenAIRealtime:
		request = &dto.BaseRequest{}
	default:
		return nil, fmt.Errorf("unsupported relay format: %s", format)
	}
	return request, err
}

func GetAndValidAudioRequest(c *gin.Context, relayMode int) (*dto.AudioRequest, error) {
	audioRequest := &dto.AudioRequest{}
	err := common.UnmarshalBodyReusable(c, audioRequest)
	if err != nil {
		return nil, err
	}
	switch relayMode {
	case relayconstant.RelayModeAudioSpeech:
		if audioRequest.Model == "" {
			return nil, errors.New("model is required")
		}
	default:
		if audioRequest.Model == "" {
			return nil, errors.New("model is required")
		}
		if audioRequest.ResponseFormat == "" {
			audioRequest.ResponseFormat = "json"
		}
	}
	return audioRequest, nil
}

func GetAndValidateRerankRequest(c *gin.Context) (*dto.RerankRequest, error) {
	var rerankRequest *dto.RerankRequest
	err := common.UnmarshalBodyReusable(c, &rerankRequest)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("getAndValidateTextRequest failed: %s", err.Error()))
		return nil, types.NewError(err, types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}

	if rerankRequest.Query == "" {
		return nil, types.NewError(fmt.Errorf("query is empty"), types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}
	if len(rerankRequest.Documents) == 0 {
		return nil, types.NewError(fmt.Errorf("documents is empty"), types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}
	return rerankRequest, nil
}

func GetAndValidateEmbeddingRequest(c *gin.Context, relayMode int) (*dto.EmbeddingRequest, error) {
	var embeddingRequest *dto.EmbeddingRequest
	err := common.UnmarshalBodyReusable(c, &embeddingRequest)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("getAndValidateTextRequest failed: %s", err.Error()))
		return nil, types.NewError(err, types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}

	if embeddingRequest.Input == nil {
		return nil, fmt.Errorf("input is empty")
	}
	if relayMode == relayconstant.RelayModeModerations && embeddingRequest.Model == "" {
		embeddingRequest.Model = "omni-moderation-latest"
	}
	if relayMode == relayconstant.RelayModeEmbeddings && embeddingRequest.Model == "" {
		embeddingRequest.Model = c.Param("model")
	}
	return embeddingRequest, nil
}

func GetAndValidateResponsesRequest(c *gin.Context) (*dto.OpenAIResponsesRequest, error) {
	request := &dto.OpenAIResponsesRequest{}
	err := common.UnmarshalBodyReusable(c, request)
	if err != nil {
		return nil, err
	}
	if request.Model == "" {
		return nil, errors.New("model is required")
	}
	if request.Input == nil {
		return nil, errors.New("input is required")
	}
	return request, nil
}

func GetAndValidateResponsesCompactionRequest(c *gin.Context) (*dto.OpenAIResponsesCompactionRequest, error) {
	request := &dto.OpenAIResponsesCompactionRequest{}
	if err := common.UnmarshalBodyReusable(c, request); err != nil {
		return nil, err
	}
	if request.Model == "" {
		return nil, errors.New("model is required")
	}
	return request, nil
}

func GetAndValidOpenAIImageRequest(c *gin.Context, relayMode int) (*dto.ImageRequest, error) {
	imageRequest := &dto.ImageRequest{}

	switch relayMode {
	case relayconstant.RelayModeImagesEdits:
		if strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
			_, err := c.MultipartForm()
			if err != nil {
				return nil, fmt.Errorf("failed to parse image edit form request: %w", err)
			}
			formData := c.Request.PostForm
			imageRequest.Prompt = formData.Get("prompt")
			imageRequest.Model = formData.Get("model")
			imageRequest.N = common.GetPointer(uint(common.String2Int(formData.Get("n"))))
			imageRequest.Quality = formData.Get("quality")
			imageRequest.Size = formData.Get("size")
			if imageValue := formData.Get("image"); imageValue != "" {
				imageRequest.Image, _ = json.Marshal(imageValue)
			}

			if imageRequest.Model == "gpt-image-1" {
				if imageRequest.Quality == "" {
					imageRequest.Quality = "standard"
				}
			}
			if imageRequest.N == nil || *imageRequest.N == 0 {
				imageRequest.N = common.GetPointer(uint(1))
			}

			hasWatermark := formData.Has("watermark")
			if hasWatermark {
				watermark := formData.Get("watermark") == "true"
				imageRequest.Watermark = &watermark
			}
			break
		}
		fallthrough
	default:
		err := common.UnmarshalBodyReusable(c, imageRequest)
		if err != nil {
			return nil, err
		}

		if imageRequest.Model == "" {
			//imageRequest.Model = "dall-e-3"
			return nil, errors.New("model is required")
		}

		if strings.Contains(imageRequest.Size, "×") {
			return nil, errors.New("size an unexpected error occurred in the parameter, please use 'x' instead of the multiplication sign '×'")
		}

		// Not "256x256", "512x512", or "1024x1024"
		if imageRequest.Model == "dall-e-2" || imageRequest.Model == "dall-e" {
			if imageRequest.Size != "" && imageRequest.Size != "256x256" && imageRequest.Size != "512x512" && imageRequest.Size != "1024x1024" {
				return nil, errors.New("size must be one of 256x256, 512x512, or 1024x1024 for dall-e-2 or dall-e")
			}
			if imageRequest.Size == "" {
				imageRequest.Size = "1024x1024"
			}
		} else if imageRequest.Model == "dall-e-3" {
			if imageRequest.Size != "" && imageRequest.Size != "1024x1024" && imageRequest.Size != "1024x1792" && imageRequest.Size != "1792x1024" {
				return nil, errors.New("size must be one of 1024x1024, 1024x1792 or 1792x1024 for dall-e-3")
			}
			if imageRequest.Quality == "" {
				imageRequest.Quality = "standard"
			}
			if imageRequest.Size == "" {
				imageRequest.Size = "1024x1024"
			}
		} else if imageRequest.Model == "gpt-image-1" {
			if imageRequest.Quality == "" {
				imageRequest.Quality = "auto"
			}
		}

		//if imageRequest.Prompt == "" {
		//	return nil, errors.New("prompt is required")
		//}

		if imageRequest.N == nil || *imageRequest.N == 0 {
			imageRequest.N = common.GetPointer(uint(1))
		}
	}

	return imageRequest, nil
}

func GetAndValidateClaudeRequest(c *gin.Context) (textRequest *dto.ClaudeRequest, err error) {
	textRequest = &dto.ClaudeRequest{}
	err = common.UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}
	if textRequest.Messages == nil || len(textRequest.Messages) == 0 {
		return nil, errors.New("field messages is required")
	}
	if textRequest.Model == "" {
		return nil, errors.New("field model is required")
	}

	//if textRequest.Stream {
	//	relayInfo.IsStream = true
	//}

	// 压缩超长的工具名称（上游 API 限制 64 字符）
	compressClaudeToolNames(c, textRequest)

	return textRequest, nil
}

func GetAndValidateTextRequest(c *gin.Context, relayMode int) (*dto.GeneralOpenAIRequest, error) {
	textRequest := &dto.GeneralOpenAIRequest{}
	err := common.UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}

	if relayMode == relayconstant.RelayModeModerations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == relayconstant.RelayModeEmbeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}

	if lo.FromPtrOr(textRequest.MaxTokens, uint(0)) > math.MaxInt32/2 {
		return nil, errors.New("max_tokens is invalid")
	}
	if textRequest.Model == "" {
		return nil, errors.New("model is required")
	}
	if textRequest.WebSearchOptions != nil {
		if textRequest.WebSearchOptions.SearchContextSize != "" {
			validSizes := map[string]bool{
				"high":   true,
				"medium": true,
				"low":    true,
			}
			if !validSizes[textRequest.WebSearchOptions.SearchContextSize] {
				return nil, errors.New("invalid search_context_size, must be one of: high, medium, low")
			}
		} else {
			textRequest.WebSearchOptions.SearchContextSize = "medium"
		}
	}
	switch relayMode {
	case relayconstant.RelayModeCompletions:
		if textRequest.Prompt == "" {
			return nil, errors.New("field prompt is required")
		}
	case relayconstant.RelayModeChatCompletions:
		// For FIM (Fill-in-the-middle) requests with prefix/suffix, messages is optional
		// It will be filled by provider-specific adaptors if needed (e.g., SiliconFlow)。Or it is allowed by model vendor(s) (e.g., DeepSeek)
		if len(textRequest.Messages) == 0 && textRequest.Prefix == nil && textRequest.Suffix == nil {
			return nil, errors.New("field messages is required")
		}
	case relayconstant.RelayModeEmbeddings:
	case relayconstant.RelayModeModerations:
		if textRequest.Input == nil || textRequest.Input == "" {
			return nil, errors.New("field input is required")
		}
	case relayconstant.RelayModeEdits:
		if textRequest.Instruction == "" {
			return nil, errors.New("field instruction is required")
		}
	}

	// 压缩超长的工具名称（上游 API 限制 64 字符）
	if len(textRequest.Tools) > 0 {
		compressOpenAIToolNames(c, textRequest)
	}

	return textRequest, nil
}

func GetAndValidateGeminiRequest(c *gin.Context) (*dto.GeminiChatRequest, error) {
	request := &dto.GeminiChatRequest{}
	err := common.UnmarshalBodyReusable(c, request)
	if err != nil {
		return nil, err
	}
	if len(request.Contents) == 0 && len(request.Requests) == 0 {
		return nil, errors.New("contents is required")
	}

	//if c.Query("alt") == "sse" {
	//	relayInfo.IsStream = true
	//}

	// 压缩超长的工具名称（上游 API 限制 64 字符）
	compressGeminiToolNames(c, request)

	return request, nil
}

func GetAndValidateGeminiEmbeddingRequest(c *gin.Context) (*dto.GeminiEmbeddingRequest, error) {
	request := &dto.GeminiEmbeddingRequest{}
	err := common.UnmarshalBodyReusable(c, request)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func GetAndValidateGeminiBatchEmbeddingRequest(c *gin.Context) (*dto.GeminiBatchEmbeddingRequest, error) {
	request := &dto.GeminiBatchEmbeddingRequest{}
	err := common.UnmarshalBodyReusable(c, request)
	if err != nil {
		return nil, err
	}
	return request, nil
}

// compressOpenAIToolNames 压缩 OpenAI 格式请求中的超长工具名称
func compressOpenAIToolNames(c *gin.Context, request *dto.GeneralOpenAIRequest) {
	if len(request.Tools) == 0 {
		return
	}

	// 收集所有工具名称
	originalNames := make([]string, len(request.Tools))
	for i, tool := range request.Tools {
		originalNames[i] = tool.Function.Name
	}

	// 批量压缩并获取映射表
	compressedNames, mapping := CompressToolNames(originalNames)

	// 存储映射表到 Context（用于日志/调试）
	StoreToolNameMapping(c, mapping)

	// 更新工具名称
	for i := range request.Tools {
		request.Tools[i].Function.Name = compressedNames[i]
	}

	// 处理 messages 中的 tool_calls
	compressToolCallsInMessages(c, request.Messages, mapping)

	// 处理 tool_choice 中指定的函数名称
	compressToolChoice(request, mapping)
}

// compressToolCallsInMessages 压缩消息中的 tool_call 名称
func compressToolCallsInMessages(c *gin.Context, messages []dto.Message, mapping ToolNameMapping) {
	for i := range messages {
		toolCalls := messages[i].ParseToolCalls()
		if len(toolCalls) == 0 {
			continue
		}

		for j := range toolCalls {
			originalName := toolCalls[j].Function.Name
			if originalName == "" {
				continue
			}
			compressed := CompressToolName(originalName)
			if compressed != originalName {
				toolCalls[j].Function.Name = compressed
				if mapping != nil {
					mapping[originalName] = compressed
				}
			}
		}
		messages[i].SetToolCalls(toolCalls)
	}

	// 更新映射表
	if len(mapping) > 0 {
		StoreToolNameMapping(c, mapping)
	}
}

// compressToolChoice 压缩 tool_choice 中指定的函数名称
func compressToolChoice(request *dto.GeneralOpenAIRequest, mapping ToolNameMapping) {
	if request.ToolChoice == nil {
		return
	}

	// 处理字符串值（auto/none/required 不需要处理）
	if _, ok := request.ToolChoice.(string); ok {
		return // auto/none/required 不涉及具体名称
	}

	// 处理对象值 {"type": "function", "function": {"name": "xxx"}}
	choiceMap, ok := request.ToolChoice.(map[string]interface{})
	if !ok {
		return
	}

	if choiceMap["type"] != "function" {
		return
	}

	function, ok := choiceMap["function"].(map[string]interface{})
	if !ok {
		return
	}

	name, ok := function["name"].(string)
	if !ok || name == "" {
		return
	}

	compressed := CompressToolName(name)
	if compressed != name {
		function["name"] = compressed
		if mapping != nil {
			mapping[name] = compressed
		}
	}
}

// compressClaudeToolNames 压缩 Claude 格式请求中的超长工具名称
// 直接操作 map[string]any 避免深拷贝导致修改丢失
func compressClaudeToolNames(c *gin.Context, request *dto.ClaudeRequest) {
	mapping := make(ToolNameMapping)

	// 处理 Tools 字段（直接操作 map[string]any，避免 ProcessTools 类型断言失败）
	if request.Tools != nil {
		toolsSlice, ok := request.Tools.([]any)
		if ok && len(toolsSlice) > 0 {
			for _, toolAny := range toolsSlice {
				toolMap, ok := toolAny.(map[string]any)
				if !ok {
					continue
				}
				name, ok := toolMap["name"].(string)
				if !ok || name == "" || len(name) <= MaxToolNameLength {
					continue
				}
				compressed := CompressToolName(name)
				toolMap["name"] = compressed
				mapping[name] = compressed
			}
		}
	}

	// 处理 Messages 中的 tool_use（直接操作 map[string]any，避免 ParseContent 深拷贝丢失）
	for i := range request.Messages {
		contentSlice, ok := request.Messages[i].Content.([]any)
		if !ok || len(contentSlice) == 0 {
			continue
		}
		for j := range contentSlice {
			contentMap, ok := contentSlice[j].(map[string]any)
			if !ok {
				continue
			}
			if contentMap["type"] != "tool_use" {
				continue
			}
			name, ok := contentMap["name"].(string)
			if !ok || name == "" || len(name) <= MaxToolNameLength {
				continue
			}
			compressed := CompressToolName(name)
			contentMap["name"] = compressed
			mapping[name] = compressed
		}
	}

	// 处理 ToolChoice
	if request.ToolChoice != nil {
		switch choice := request.ToolChoice.(type) {
		case *dto.ClaudeToolChoice:
			if choice.Name != "" && len(choice.Name) > MaxToolNameLength {
				compressed := CompressToolName(choice.Name)
				mapping[choice.Name] = compressed
				choice.Name = compressed
			}
		case dto.ClaudeToolChoice:
			if choice.Name != "" && len(choice.Name) > MaxToolNameLength {
				compressed := CompressToolName(choice.Name)
				mapping[choice.Name] = compressed
				request.ToolChoice = &dto.ClaudeToolChoice{
					Type:                   choice.Type,
					Name:                   compressed,
					DisableParallelToolUse: choice.DisableParallelToolUse,
				}
			}
		case map[string]interface{}:
			if name, ok := choice["name"].(string); ok && name != "" && len(name) > MaxToolNameLength {
				compressed := CompressToolName(name)
				choice["name"] = compressed
				mapping[name] = compressed
			}
		}
	}

	if len(mapping) > 0 {
		StoreToolNameMapping(c, mapping)
	}
}

// compressGeminiToolNames 压缩 Gemini 格式请求中的超长工具名称
func compressGeminiToolNames(c *gin.Context, request *dto.GeminiChatRequest) {
	mapping := make(ToolNameMapping)

	// 处理 Tools 字段
	tools := request.GetTools()
	if len(tools) > 0 {
		for i := range tools {
			if tools[i].FunctionDeclarations == nil {
				continue
			}
			// FunctionDeclarations 可能是 []dto.FunctionRequest 或 []any
			switch funcs := tools[i].FunctionDeclarations.(type) {
			case []dto.FunctionRequest:
				for j := range funcs {
					name := funcs[j].Name
					if name != "" && len(name) > MaxToolNameLength {
						original := name
						compressed := CompressToolName(original)
						funcs[j].Name = compressed
						mapping[original] = compressed
					}
				}
				tools[i].FunctionDeclarations = funcs
			case []interface{}:
				for j, fn := range funcs {
					if fnMap, ok := fn.(map[string]interface{}); ok {
						if name, ok := fnMap["name"].(string); ok && name != "" && len(name) > MaxToolNameLength {
							original := name
							compressed := CompressToolName(original)
							fnMap["name"] = compressed
							funcs[j] = fnMap
							mapping[original] = compressed
						}
					}
				}
				tools[i].FunctionDeclarations = funcs
			}
		}
		// 更新 Tools 字段
		request.SetTools(tools)
	}

	// 处理 Contents 中的 functionCall
	for i := range request.Contents {
		for j := range request.Contents[i].Parts {
			part := &request.Contents[i].Parts[j]
			if part.FunctionCall != nil && part.FunctionCall.FunctionName != "" {
				original := part.FunctionCall.FunctionName
				if len(original) > MaxToolNameLength {
					compressed := CompressToolName(original)
					part.FunctionCall.FunctionName = compressed
					mapping[original] = compressed
				}
			}
		}
	}

	// 处理 ToolConfig
	if request.ToolConfig != nil && request.ToolConfig.FunctionCallingConfig != nil {
		config := request.ToolConfig.FunctionCallingConfig
		if len(config.AllowedFunctionNames) > 0 {
			for i, name := range config.AllowedFunctionNames {
				if len(name) > MaxToolNameLength {
					original := name
					compressed := CompressToolName(original)
					config.AllowedFunctionNames[i] = compressed
					mapping[original] = compressed
				}
			}
		}
	}

	if len(mapping) > 0 {
		StoreToolNameMapping(c, mapping)
	}
}
