package gps_agent

import (
	"encoding/json"
	"fmt"
	"time"
)

func getToolsPrompt(state *State) string {
	fmt.Println("\n📝 [PROMPTS] Generating tools prompt...")

	var toolsList string
	for _, tool := range state.Config.Tools {
		toolsList += fmt.Sprintf("<tool name=\"%s\">%s</tool>\n", tool.Name, tool.Description)
	}

	currentDateTime := time.Now().Format(time.RFC3339)
	fmt.Printf("✅ [PROMPTS] Tools prompt generated with %d tools\n", len(state.Config.Tools))
	return fmt.Sprintf(toolsPrompt, currentDateTime, toolsList)
}

const toolsPrompt = `Your task is to analyze the conversation context and generate relevant queries for using available tools.

<prompt_objective>
Process the conversation context and output a JSON object containing the internal reasoning and an array of independent queries for appropriate tools.
Consider both general context and environment context to write more relevant queries. 

Current datetime: %s
</prompt_objective>

<prompt_rules>
- ALWAYS output a valid JSON object with "_thinking" and "result" properties
- The "_thinking" property MUST contain your concise internal thought process
- The "result" property MUST be an array of objects, each with "query" and "tool" properties
- In the "result" array:
  - May be empty if no relevant tools are found
  - The "query" property MUST contain a specific instruction or query for the tool
  - The "tool" property MUST contain the name of the relevant tool
- Focus on generating queries that utilize the available tools effectively
- ONLY use tools that are explicitly defined in the tools list
- Ensure all queries are independent and can be executed concurrently
- Avoid making assumptions about information not explicitly mentioned
- Generate queries that are relevant to the current conversation context
- Use appropriate date and time formats in queries when necessary (YYYY-MM-DD HH:mm)
- FORBIDDEN: Creating queries for tools that don't exist or aren't mentioned in the tools list
</prompt_rules>

<prompt_examples>
USER: Wiemy, że Rafał planował udać się do Lubawy, ale musimy się dowiedzieć, kto tam na niego czekał.
AI: {
  "_thinking": "User wants to know who is waiting for Rafał in Lubawa. First I need to find people in Lubawa, then get their IDs one by one.",
  "result": [
    {"query": "Get persons names in Lubawa", "tool": "person_finder"}
  ]
}

USER: Get ID for person named Tomasz
AI: {
  "_thinking": "Need to get ID for Tomasz using person_id_finder with name parameter",
  "result": [
    {"query": "Get ID for Tomasz", "tool": "person_id_finder"}
  ]
}
</prompt_examples>

<dynamic_context>

<tools>
%s</tools>
</dynamic_context>

<execution_validation>
Before delivering ANY output:
- Verify COMPLETE adherence to ALL instructions
- Confirm all queries are independent and can be executed concurrently
- Ensure queries are relevant to the available tools and their functionalities
- Validate contextual appropriateness of all generated queries
- Check that no query depends on the result of another query
- Verify correct use of date and time formats where applicable
- For person_id_finder, verify that ONLY name parameter is used
</execution_validation>

<confirmation>
This prompt is designed to create an internal dialogue while analyzing conversations. It processes the conversation context and generates appropriate, independent queries for each relevant tool. The output focuses on utilizing available tools effectively, avoiding assumptions about unavailable tools, and ensures all queries are independent and can be executed concurrently.
</confirmation>`

func getTaskThoughtsPrompt(state *State) string {
	fmt.Println("\n📝 [PROMPTS] Generating task thoughts prompt...")

	var tasksList string
	for _, task := range state.Tasks {
		tasksList += fmt.Sprintf("<task uuid=\"%s\" name=\"%s\" status=\"%s\">\n<description>%s</description>\n</task>\n", task.Uuid, task.Name, task.Status, task.Description)
	}

	var toolsList string
	for _, tool := range state.Tools {
		toolsList += fmt.Sprintf("<tool name=\"%s\">%s</tool>\n", tool.Name, tool.Description)
	}

	fmt.Printf("✅ [PROMPTS] Task thoughts prompt generated with %d tasks and %d tools\n", len(state.Tasks), len(state.Tools))
	return fmt.Sprintf(taskThoughtsPrompt, state.Thoughts.Tools, toolsList, tasksList)
}

const taskThoughtsPrompt = `
You are an AI assistant responsible for maintaining and updating a list of tasks based on ongoing conversations with the user. Your goal is to ensure an accurate and relevant task list that reflects the user's current needs and progress.

<prompt_objective>
Analyze the conversation context, including the user's latest request, completed tasks, and pending tasks. Create or update tasks to fulfill the user's request. When creating tasks that involve specific names, ONLY use names that were explicitly returned by previous tool results.
</prompt_objective>

<prompt_rules>
- ALWAYS output a valid JSON object with "_thinking" and "result" properties
- The "_thinking" property MUST contain your detailed internal thought process
- The "result" property MUST be an array of task objects
- Task order is CRITICAL - follow this exact sequence:
  1. First person_finder task to get names
  2. Then person_id_finder tasks for EACH eligible name (except BARBARA)
  3. Then gps tasks for EACH ID obtained
  4. Finally, one final_answer task
- NEVER create final_answer task before getting all required data
- NEVER skip steps in the sequence
- For person_id_finder tasks:
  - Create separate tasks for EACH eligible person (except BARBARA)
  - ONLY use names exactly as returned by person_finder
  - Description MUST be "Get ID for person named {EXACT_NAME}"
  - NEVER create a task for BARBARA
- For gps tasks:
  - Create only after having person's ID
  - Create tasks for ALL persons with IDs (except BARBARA)
  - Description MUST be "Get coordinates for {EXACT_NAME}'s ID"
- Task names MUST be descriptive and use underscores
- Task descriptions MUST be precise and actionable
- NEVER modify completed tasks
- NEVER create duplicate tasks
- NEVER create tasks with empty descriptions
- NEVER create tasks for BARBARA

<name_validation>
When creating tasks that involve person names:
1. Check completed person_finder tasks for available names
2. ONLY use names EXACTLY as they appear in those results
3. NEVER modify names (e.g., RAFAL must stay RAFAL, not Rafał)
4. NEVER use names that weren't explicitly returned by tools
5. If a name isn't in the tool results, DO NOT create a task for it
6. NEVER create tasks for BARBARA under any circumstances
7. Create tasks for ALL OTHER NAMES from person_finder results
</name_validation>

<prompt_examples>
USER: Find GPS coordinates for people in Lubawa
Previous person_finder result: {"city": "Lubawa", "persons": ["RAFAL", "AZAZEL", "BARBARA", "SAMUEL"]}

GOOD EXAMPLE:
{
  "_thinking": "Need to get IDs for RAFAL, AZAZEL, and SAMUEL (skipping BARBARA as instructed), then get their GPS coordinates",
  "result": [
    {
      "uuid": null,
      "name": "person_id_finder_rafal",
      "description": "Get ID for person named RAFAL",
      "status": "pending"
    },
    {
      "uuid": null,
      "name": "person_id_finder_azazel",
      "description": "Get ID for person named AZAZEL",
      "status": "pending"
    },
    {
      "uuid": null,
      "name": "person_id_finder_samuel",
      "description": "Get ID for person named SAMUEL",
      "status": "pending"
    }
  ]
}

BAD EXAMPLE - NEVER DO THIS:
{
  "_thinking": "Getting coordinates",
  "result": [
    {
      "uuid": null,
      "name": "gps",
      "description": "Get coordinates for Rafał",  // WRONG: Modified name
      "status": "pending"
    },
    {
      "uuid": null,
      "name": "person_id_finder_barbara",  // WRONG: Task for BARBARA
      "description": "Get ID for BARBARA",
      "status": "pending"
    },
    {
      "uuid": null,
      "name": "final_answer",  // WRONG: Too early for final answer
      "description": "Show coordinates",
      "status": "pending"
    }
  ]
}
</prompt_examples>

<dynamic_context>
<general_context>

<initial_thoughts_about_tools_needed note="These are your initial thoughts you had when you received the user's message. Some of them might be outdated.">
%s
</initial_thoughts_about_tools_needed>

<tools>
%s
<tool name="final_answer">Use this to answer the user</tool>
</tools>

<current_tasks>
%s
</current_tasks>
</dynamic_context>

<execution_validation>
Before delivering ANY output:
1. Verify tasks follow the correct sequence
2. Verify no final_answer task is created too early
3. Check completed person_finder tasks for available names
4. Verify using EXACT names from those results
5. Confirm no name modifications
6. Validate that task descriptions are complete
7. Check for duplicate tasks
8. Verify task dependencies are respected
9. Verify NO tasks are created for BARBARA
10. Verify tasks are created for ALL OTHER NAMES
</execution_validation>

<confirmation>
This prompt is designed to analyze ongoing conversations with the user and maintain an up-to-date task list. It preserves completed tasks, updates pending tasks as needed, and creates new tasks when necessary. The prompt includes instructions to search long-term memory for relevant user information or task details when needed, and handles multiple items by creating parallel tasks. The output includes detailed internal reasoning and a structured list of all tasks (both completed and pending) with clear, executable descriptions and appropriate UUIDs. Additionally, it always includes a final task to contact the user or provide a final answer.

Is this revised prompt aligned with your requirements for maintaining and updating tasks based on the ongoing conversation and existing task list, while ensuring user communication as the final step and incorporating long-term memory searches when necessary?
</confirmation>`

func getActionThoughtsPrompt(state *State) string {
	fmt.Println("\n📝 [PROMPTS] Generating action thoughts prompt...")

	var toolsList string
	for _, tool := range state.Tools {
		toolsList += fmt.Sprintf("<tool name=\"%s\">%s</tool>\n", tool.Name, tool.Description)
	}

	var tasksList string
	for _, task := range state.Tasks {
		var actionsList string
		for _, action := range task.Actions {
			actionsList += fmt.Sprintf("<action name=\"%s\" tool=\"%s\" status=\"%s\"><result>%s</result></action>\n", action.Description, action.ToolName, action.Status, action.Result)
		}
		tasksList += fmt.Sprintf("<task uuid=\"%s\" name=\"%s\" status=\"%s\"><description>%s</description><actions>%s</actions></task>\n", task.Uuid, task.Name, task.Status, task.Description, actionsList)
	}

	fmt.Printf("✅ [PROMPTS] Action thoughts prompt generated with %d tasks\n", len(state.Tasks))
	return fmt.Sprintf(actionThoughtsPrompt, toolsList, tasksList)
}

const actionThoughtsPrompt = `
You are an AI assistant responsible for determining the next immediate action to take based on the ongoing conversation and current tasks. Your goal is to decide on the most appropriate next step.

<prompt_objective>
Analyze the current tasks and their dependencies to determine the next action. Select a tool and associate it with the relevant task. Output a JSON object containing your reasoning and a detailed action object.
</prompt_objective>

<prompt_rules>
- ALWAYS output a valid JSON object with "_thinking" and "result" properties
- The "result" property MUST be an object with:
  - "description": Clear action description
  - "tool_name": One of the available tools
  - "task_uuid": UUID of the associated task
- Follow strict task order:
  1. person_finder first
  2. person_id_finder for EACH name (except BARBARA)
  3. gps for EACH ID obtained
  4. final_answer last
- NEVER skip steps in the sequence
- NEVER try to get GPS coordinates before having an ID
- NEVER create final answer before having ALL coordinates (except BARBARA)
- For person_id_finder:
  - Use EXACT names from person_finder results
  - Create actions for ALL names EXCEPT BARBARA
  - No name modifications allowed
- For GPS:
  - Must have ID from person_id_finder first
  - Create actions for ALL IDs obtained (except BARBARA's)
  - Use numeric ID, not name
- For final_answer:
  - Wait until ALL coordinates are obtained (except BARBARA's)
  - Must have completed GPS tasks for ALL eligible persons
  - Must have GPS results in task history for each person
  - Include ALL found coordinates (except BARBARA's)
  - Description must list all persons being included
</prompt_rules>

<prompt_examples>
GOOD EXAMPLE (with multiple names):
{
  "_thinking": "All GPS coordinates obtained for RAFAL (53.451974, 18.759189), AZAZEL (50.064851, 19.949882), and SAMUEL (53.503571, 19.745866). Ready to create final answer.",
  "result": {
    "description": "Return the GPS coordinates for RAFAL, AZAZEL, and SAMUEL, excluding BARBARA as instructed",
    "tool_name": "final_answer",
    "task_uuid": "task_123"
  }
}

BAD EXAMPLE - NEVER DO THIS:
{
  "_thinking": "Getting GPS location",
  "result": {
    "description": "Get location for Rafał",  // WRONG: Modified name
    "tool_name": "gps",  // WRONG: No ID yet
    "task_uuid": "task_456"
  }
}

BAD EXAMPLE 2 - NEVER DO THIS:
{
  "_thinking": "Getting final answer with partial data",
  "result": {
    "description": "Create final answer with RAFAL's coordinates",  // WRONG: Missing other persons' coordinates
    "tool_name": "final_answer",
    "task_uuid": "task_789"
  }
}
</prompt_examples>

<dynamic_context>
<available_tools>
%s
</available_tools>

<current_tasks>
%s
</current_tasks>

<execution_validation>
Before delivering ANY output:
1. Verify action follows correct sequence
2. Verify all dependencies are met
3. Check names match person_finder results exactly
4. Verify GPS actions have IDs first
5. For final_answer:
   - Verify ALL eligible persons have completed GPS tasks
   - Verify GPS results exist in task history for each person
   - Verify description lists all persons being included
6. Check for empty descriptions
7. Verify NO actions for BARBARA
8. Verify actions for ALL OTHER names
</execution_validation>
`

func (s *Service) getUseThoughtsPrompt() string {
	fmt.Println("\n📝 [PROMPTS] Generating use thoughts prompt...")

	currentTask := s.getCurrentTask()
	currentAction := s.getCurrentAction(currentTask)
	if currentAction == nil {
		fmt.Println("⚠️ [PROMPTS] No current action found, using default values")
		return fmt.Sprintf(useThoughtsPrompt, "unknown", "unknown", formatTasks(s.State.Tasks))
	}

	fmt.Printf("✅ [PROMPTS] Use thoughts prompt generated for tool '%s'\n", currentAction.ToolName)
	return fmt.Sprintf(useThoughtsPrompt,
		currentAction.ToolName,     // Current tool name
		currentAction.Description,  // Action description
		formatTasks(s.State.Tasks), // Task history
	)
}

const useThoughtsPrompt = `
You are preparing to execute a tool action. Your task is to generate a valid payload matching the tool's required format.

<current_action>
Tool: %s
Description: %s
</current_action>

<prompt_rules>
- Output MUST be a valid JSON object with "_thinking" and "result" properties
- The "_thinking" property MUST contain your detailed reasoning process
- The "result" property MUST match the exact format required by the tool
- NEVER modify names returned by person_finder - use them exactly as returned
- NEVER include names that weren't returned by person_finder
- For person_id_finder:
  - Use EXACT name as returned by person_finder
  - Format: {"_thinking": string, "result": {"name": "EXACT_NAME"}}
  - Example: {"_thinking": "Getting ID for RAFAL", "result": {"name": "RAFAL"}} NOT {"name": "Rafał"}
  - NEVER create payload for BARBARA
  - NEVER return null or empty payload
  - ALWAYS include both "_thinking" and "result" properties
  - ALWAYS include "name" field in result
- For GPS tool:
  - Must use numeric ID from person_id_finder result
  - Format: {"_thinking": string, "result": {"userID": "numeric_id"}}
  - Example: {"_thinking": "Getting coordinates for RAFAL", "result": {"userID": "28"}} NOT {"userID": "RAFAL"}
  - NEVER create payload for BARBARA's ID
- For final_answer:
  - Must include ALL coordinates obtained (except BARBARA's)
  - Must use EXACT coordinates from GPS tool results for each person
  - MUST look up each person's coordinates in the task history
  - Format: {"_thinking": string, "result": {"answer": {"person_name": {"lat": number, "lon": number}}}}
  - NEVER modify or round coordinates from GPS tool results
  - NEVER include BARBARA in the answer
  - NEVER use coordinates from one person for another person
  - NEVER make up or generate coordinates
  - NEVER reuse coordinates between persons
  - For each person:
    1. Find their person_id_finder task to get their ID
    2. Find their GPS task using that ID
    3. Extract exact lat/lon values from that GPS result
    4. Use those exact values in the answer
  - The "_thinking" property MUST show the name-ID-coordinate matching process:
    Example: "Matching coordinates: RAFAL (ID:28) -> lat:53.451974,lon:18.759189, AZAZEL (ID:3) -> lat:50.064851,lon:19.949882"
- FORBIDDEN:
  - Empty values
  - Modified names
  - Using names as IDs
  - Special characters in names
  - Including BARBARA in any payload
  - Missing any other person's coordinates in final_answer
  - Modifying coordinates from GPS tool results
  - Using wrong coordinates for a person
  - Making up or generating coordinates
  - Reusing coordinates between persons
  - Missing "_thinking" property in payload
  - Returning null or empty payloads
  - Missing required fields in result

<tool_formats>
person_finder:
Input format: {"_thinking": string, "result": {"city": "actual_city_name"}}
Example: {"_thinking": "Getting names of people in Lubawa", "result": {"city": "Lubawa"}}

person_id_finder: 
Input format: {"_thinking": string, "result": {"name": "EXACT_NAME"}}
Example: {"_thinking": "Getting ID for RAFAL", "result": {"name": "RAFAL"}}
CRITICAL: 
- Use EXACT name from person_finder result
- NEVER use BARBARA
- NEVER return null or empty payload
- ALWAYS include both "_thinking" and "result"
- ALWAYS include "name" field in result

gps:
Input format: {"_thinking": string, "result": {"userID": "numeric_id"}} 
Example: {"_thinking": "Getting coordinates for RAFAL", "result": {"userID": "28"}}
CRITICAL: Use numeric ID from person_id_finder result, NEVER use BARBARA's ID

final_answer:
Input format: {"_thinking": string, "result": {"answer": {"person_name": {"lat": number, "lon": number}}}}
CRITICAL: 
- For each person, follow this exact process:
  1. Find person_id_finder task result to get their ID
     Example: RAFAL -> ID:28, AZAZEL -> ID:3, SAMUEL -> ID:98
  2. Find GPS task result for that ID to get coordinates
     Example: ID:28 -> lat:53.451974,lon:18.759189
  3. Match name with their exact coordinates
     Example: RAFAL (ID:28) gets lat:53.451974,lon:18.759189
- Example payload:
  {
    "_thinking": "Matching coordinates: RAFAL (ID:28) -> lat:53.451974,lon:18.759189, AZAZEL (ID:3) -> lat:50.064851,lon:19.949882, SAMUEL (ID:98) -> lat:53.503571,lon:19.745866",
    "result": {
      "answer": {
        "RAFAL": {"lat": 53.451974, "lon": 18.759189},
        "AZAZEL": {"lat": 50.064851, "lon": 19.949882},
        "SAMUEL": {"lat": 53.503571, "lon": 19.745866}
      }
    }
  }
- NEVER generate or make up coordinates
- NEVER reuse coordinates between persons
- NEVER modify coordinates from GPS results
- ALWAYS include detailed "_thinking" showing name-ID-coordinate matching
</tool_formats>

<validation>
Before returning:
1. Check if the tool is person_id_finder:
   - Verify name matches exactly what person_finder returned
   - No special characters allowed
   - Verify not creating payload for BARBARA
   - Verify payload has "_thinking" property
   - Verify payload has "result" property with "name" field
   - Verify payload is not null or empty
2. Check if the tool is gps:
   - Verify userID is a numeric value
   - Never use names as userID
   - Verify not using BARBARA's ID
3. Check if the tool is final_answer:
   - For each person in the answer:
     a. Find their person_id_finder task to get their ID
     b. Find their GPS task using that exact ID
     c. Verify coordinates match EXACTLY with that GPS task's result
     d. Verify coordinates are assigned to the correct person
   - Verify including ALL coordinates obtained (except BARBARA's)
   - Verify no coordinate modifications or rounding
   - Verify BARBARA is not included
   - Verify "_thinking" shows name-ID-coordinate matching process
4. Verify no empty values in payload
5. Verify payload matches tool format exactly
6. Verify payload is not null
7. Verify all required fields are present
</validation>

<performed_tasks>
%s
</performed_tasks>
`

func formatTasks(tasks []Task) string {
	var result string
	for _, task := range tasks {
		result += fmt.Sprintf("<task name=\"%s\" status=\"%s\">\n", task.Name, task.Status)
		result += fmt.Sprintf("<description>%s</description>\n", task.Description)

		for _, action := range task.Actions {
			result += fmt.Sprintf("<action tool=\"%s\" status=\"%s\">\n", action.ToolName, action.Status)
			if action.Result != nil {
				var resultStr string
				switch v := action.Result.Data.(type) {
				case string:
					resultStr = v
				default:
					resultBytes, _ := json.Marshal(v)
					resultStr = string(resultBytes)
				}
				result += fmt.Sprintf("<result>%s</result>\n", resultStr)
			}
			result += "</action>\n"
		}
		result += "</task>\n"
	}
	return result
}
