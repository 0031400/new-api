package openaicompat

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

func ResponsesRequestToChatCompletionsRequest(req *dto.OpenAIResponsesRequest) (*dto.GeneralOpenAIRequest, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if strings.TrimSpace(req.Model) == "" {
		return nil, errors.New("model is required")
	}

	chatReq := &dto.GeneralOpenAIRequest{
		Model:                req.Model,
		Stream:               req.Stream,
		StreamOptions:        req.StreamOptions,
		Temperature:          req.Temperature,
		TopP:                 req.TopP,
		TopLogProbs:          req.TopLogProbs,
		ToolChoice:           req.ToolChoice,
		User:                 req.User,
		Metadata:             req.Metadata,
		Store:                req.Store,
		SafetyIdentifier:     req.SafetyIdentifier,
		PromptCacheRetention: req.PromptCacheRetention,
	}
	if strings.TrimSpace(req.ServiceTier) != "" {
		if raw, err := common.Marshal(req.ServiceTier); err == nil {
			chatReq.ServiceTier = raw
		}
	}
	if req.MaxOutputTokens != nil {
		chatReq.MaxCompletionTokens = req.MaxOutputTokens
	}

	if len(req.Instructions) > 0 {
		var instructions string
		if common.GetJsonType(req.Instructions) == "string" {
			_ = common.Unmarshal(req.Instructions, &instructions)
		} else {
			instructions = common.Interface2String(req.Instructions)
		}
		instructions = strings.TrimSpace(instructions)
		if instructions != "" {
			chatReq.Messages = append(chatReq.Messages, dto.Message{
				Role:    chatReq.GetSystemRoleName(),
				Content: instructions,
			})
		}
	}

	mediaInputs := req.ParseInput()
	if len(mediaInputs) == 0 {
		chatReq.Messages = append(chatReq.Messages, dto.Message{
			Role:    "user",
			Content: "",
		})
	} else {
		content := make([]dto.MediaContent, 0, len(mediaInputs))
		for _, input := range mediaInputs {
			switch input.Type {
			case "input_text", "output_text":
				content = append(content, dto.MediaContent{
					Type: dto.ContentTypeText,
					Text: input.Text,
				})
			case "input_image":
				content = append(content, dto.MediaContent{
					Type: dto.ContentTypeImageURL,
					ImageUrl: dto.MessageImageUrl{
						Url:    input.ImageUrl,
						Detail: input.Detail,
					},
				})
			case "input_file":
				content = append(content, dto.MediaContent{
					Type: dto.ContentTypeFile,
					File: dto.MessageFile{
						FileData: input.FileUrl,
					},
				})
			default:
				if strings.TrimSpace(input.Text) != "" {
					content = append(content, dto.MediaContent{
						Type: dto.ContentTypeText,
						Text: input.Text,
					})
				}
			}
		}
		if len(content) == 1 && content[0].Type == dto.ContentTypeText {
			chatReq.Messages = append(chatReq.Messages, dto.Message{
				Role:    "user",
				Content: content[0].Text,
			})
		} else {
			chatReq.Messages = append(chatReq.Messages, dto.Message{
				Role:    "user",
				Content: content,
			})
		}
	}

	if len(req.Tools) > 0 {
		var rawTools []map[string]any
		if err := common.Unmarshal(req.Tools, &rawTools); err == nil {
			tools := make([]dto.ToolCallRequest, 0, len(rawTools))
			for _, tool := range rawTools {
				if common.Interface2String(tool["type"]) != "function" {
					continue
				}
				functionData, ok := tool["function"].(map[string]any)
				if !ok {
					continue
				}
				name := strings.TrimSpace(common.Interface2String(functionData["name"]))
				if name == "" {
					continue
				}
				tools = append(tools, dto.ToolCallRequest{
					Type: "function",
					Function: dto.FunctionRequest{
						Name:        name,
						Description: common.Interface2String(functionData["description"]),
						Parameters:  functionData["parameters"],
					},
				})
			}
			if len(tools) > 0 {
				chatReq.Tools = tools
			}
		}
	}

	return chatReq, nil
}
