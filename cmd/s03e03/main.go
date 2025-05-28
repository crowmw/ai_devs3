package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/sashabaranov/go-openai"
)

func main() {
	envSvc, err := env.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	aiSvc, err := ai.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	model := "gpt-4o"

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "Let's begin investigating the database. First, let's see what tables are available."},
	}

	var finalQuery string
	var result string

	for {
		response, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
			Model:    model,
			Messages: messages,
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		content := response.Choices[0].Message.Content

		fmt.Println("🤖 AI:", content)

		if strings.HasPrefix(content, "FINAL:") {
			finalQuery = strings.TrimPrefix(content, "FINAL:")
			result, err = http.PostSQLQueryToAPIDB(envSvc, finalQuery)
			if err != nil {
				fmt.Println(err)
				return
			}
			break
		}

		queryResult, err := http.PostSQLQueryToAPIDB(envSvc, content)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("📊 DB:", content)

		messages = append(messages,
			openai.ChatCompletionMessage{Role: "assistant", Content: content},
			openai.ChatCompletionMessage{Role: "user", Content: queryResult},
		)
	}

	fmt.Println("--------------------------------")
	fmt.Println("🎯 Final query:", finalQuery)
	fmt.Println("🎯 Result:", result)

	// Extract dc_id values from the result
	var response struct {
		Reply []struct {
			DCID string `json:"dc_id"`
		} `json:"reply"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal([]byte(result), &response); err != nil {
		fmt.Println("Error parsing result:", err)
		return
	}

	// Create array of dc_id values
	dcIDs := make([]string, len(response.Reply))
	for i, item := range response.Reply {
		dcIDs[i] = item.DCID
	}

	reportResult, err := http.SendC3ntralaReport("database", dcIDs)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("❇️ Report result:", reportResult)
}

const systemPrompt = `
You're a MySQL database expert, specialized in database design, back-engineering.

<context>
- We gained access to BanAN company's database
- HQ has provided you with a special API that allows you to execute almost any data retrieval queries from the mentioned database
- We know there are tables named users, datacenters and connections
- You may not need all tables right now
</context>

<objective>
Your task is to return the ID numbers of active datacenters that are managed by managers who are currently on vacation (inactive). This will help us better identify data centers that are more vulnerable to attack. You can only use MySQL queries for this purpose.
</objective>

<rules>
- ALWAYS response with raw sql query string
- Do not use any other language than sql
- No additional descriptions, explanations or Markdown formatting
- To see the list of available tables execute the query SHOW TABLES;
- For tables of interest, get their structure by executing SHOW CREATE TABLE table_name; (e.g. SHOW CREATE TABLE users;)
- Pay attention to columns that are primary keys (PRIMARY KEY) or contain indexes (INDEX)
- Pay attention to columns that are related to datacenters (e.g. datacenter_id)
- Pay attention to columns that are related to managers (e.g. manager_id)
- Pay attention to columns that are related to actions (e.g. action_id)
- Pay attention to columns that are related to time (e.g. created_at, updated_at)
- Pay attention to columns that are related to status (e.g. status)
- Pay attention to columns that are related to comments (e.g. comment)
- You can only perform operations like: select, show tables, desc table, show create table
- As a final result, write an SQL query that will return DC_ID of active datacenters whose managers (from users table) are inactive. You should deduce this yourself based on the results of subsequent queries.
- As final result, return only the SQL query string with FINAL: prefix, nothing else.
</rules>

<examples>
USER: [Typical input example]
AI: [SHOW TABLES;]

USER: [Value returned from previous query]
AI: [SQL query that created based on previous query response and rules]

USER: [Value returned from previous query]
AI: [SQL query that created based on previous query response and rules]

USER: [Value returned from previous query]
AI: [SQL query that created based on previous query response and rules]

USER: [Value returned from previous query]
AI: [SQL query that created based on previous query response and rules]

USER: [Value returned from previous query]
AI: [Final SQL query to run to get the desired result (DC_ID of active datacenters, where manager is not active)]
</examples>

After each command, analyze the results and decide what to investigate next.
Let's begin crafting the next snippet prompt.`
