package serce_agent

import (
	"fmt"
	"strings"
	"time"
)

func getToolsPrompt(state *State) string {
	var toolsList strings.Builder
	for _, tool := range state.Tools {
		toolsList.WriteString(fmt.Sprintf("- %s: %s\n  Instruction: %s\n", tool.Name, tool.Description, tool.Instruction))
	}

	return fmt.Sprintf(`Your task is to analyze the conversation context and select the most appropriate tool with its payload.

<prompt_objective>
Process the conversation context and output a JSON object containing:
1. Your internal reasoning (_thinking)
2. A single tool selection with its payload (result)
Consider both general context and environment context to make the best tool selection.

Current datetime: %s
</prompt_objective>

<prompt_rules>
- ALWAYS output a valid JSON object with "_thinking" and "result" properties
- The "_thinking" property MUST contain your concise internal thought process
- The "result" property MUST be an object with "tool" and "payload" properties
- The "tool" property MUST contain the name of the selected tool
- The "payload" property MUST be an object containing the required parameters for the tool
- ONLY use tools that are explicitly defined in the tools list
- Ensure the payload matches the tool's instruction format exactly
- Avoid making assumptions about information not explicitly mentioned
- Generate payloads that are relevant to the current conversation context
- Use appropriate date and time formats in payloads when necessary (YYYY-MM-DD HH:mm)
- FORBIDDEN: Creating payloads for tools that don't exist or aren't mentioned in the tools list
- NEVER return "none" as a tool - always select an appropriate tool
- For memory-related queries, use the answer_question tool
- When memory contains key-value pairs (e.g., "klucz=value, data=value"), ALWAYS use the value after "klucz="
</prompt_rules>

<prompt_examples>
USER: Analyze this image: https://example.com/image.jpg
AI: {
  "_thinking": "User wants to analyze an image. The image_analyzer tool is perfect for this task. The payload needs an 'image' field with the URL.",
  "result": {
    "tool": "image_analyzer",
    "payload": {
      "image": "https://example.com/image.jpg"
    }
  }
}

USER: What is the capital of France?
AI: {
  "_thinking": "User is asking a general knowledge question. The answer_question tool can handle this. The payload needs a 'question' field.",
  "result": {
    "tool": "answer_question",
    "payload": {
      "question": "What is the capital of France?"
    }
  }
}

USER: Save this information: The meeting is scheduled for tomorrow at 2 PM
AI: {
  "_thinking": "User wants to save information. The data_memory tool is appropriate. The payload needs a 'data' field with the information.",
  "result": {
    "tool": "data_memory",
    "payload": {
      "data": "The meeting is scheduled for tomorrow at 2 PM"
    }
  }
}

USER: What was the value of the 'key' variable?
AI: {
  "_thinking": "User is asking about a previously stored value. The answer_question tool can access memory and answer this question. When memory contains key-value pairs, I should use the value after 'klucz='.",
  "result": {
    "tool": "answer_question",
    "payload": {
      "question": "What was the value of the 'key' variable? Remember to use the value after 'klucz=' in the memory."
    }
  }
}
</prompt_examples>

<dynamic_context>
<tools>
%s
</tools>
</dynamic_context>

<execution_validation>
Before delivering ANY output:
- Verify COMPLETE adherence to ALL instructions
- Confirm the selected tool exists in the tools list
- Ensure the payload matches the tool's instruction format exactly
- Validate contextual appropriateness of the tool selection
- Verify correct use of date and time formats where applicable
- Check that all required payload fields are present
- NEVER return "none" as a tool - always select an appropriate tool
- For memory-related queries, use the answer_question tool
- When memory contains key-value pairs, verify using the value after "klucz="
</execution_validation>

<confirmation>
This prompt is designed to create an internal dialogue while analyzing conversations. It processes the conversation context and selects the most appropriate tool with its payload. The output focuses on utilizing available tools effectively, avoiding assumptions about unavailable tools, and ensures the payload matches the tool's requirements exactly. For any memory-related queries, the answer_question tool should be used as it has access to the saved memory. When memory contains key-value pairs, the value after "klucz=" should be used as the answer.
</confirmation>`, time.Now().Format("2006-01-02 15:04:05"), toolsList.String())
}
