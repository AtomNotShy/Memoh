package chat

import "strings"

// ExtractAssistantOutputs collects assistant-role outputs from a slice of ModelMessages.
func ExtractAssistantOutputs(messages []ModelMessage) []AssistantOutput {
	if len(messages) == 0 {
		return nil
	}
	outputs := make([]AssistantOutput, 0, len(messages))
	for _, msg := range messages {
		if msg.Role != "assistant" {
			continue
		}
		content := strings.TrimSpace(msg.TextContent())
		parts := filterContentParts(msg.ContentParts())
		if content == "" && len(parts) == 0 {
			continue
		}
		outputs = append(outputs, AssistantOutput{Content: content, Parts: parts})
	}
	return outputs
}

func filterContentParts(parts []ContentPart) []ContentPart {
	if len(parts) == 0 {
		return nil
	}
	filtered := make([]ContentPart, 0, len(parts))
	for _, p := range parts {
		if p.HasValue() {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
