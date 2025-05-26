package factory

const analyzeReportsPrompt = `
<context>%s</context>
You are a keyword quality judge specializing in intelligence document analysis.

<objective>
Review the generated keywords for a report and determine if any additional keywords should be added.
Your role is to catch missing keywords that the first analyzer might have overlooked.
You must cross-reference any people mentioned in the report with the Context information.
If a person appears in either the report OR the Context, include their full details in the keywords.
You must identify and include specific locations where evidence (like fingerprints) was found.
</objective>

<rules>
- Review the original report content and the generated keywords
- Check if all important elements are covered by keywords
- Look for missing: technical terms, equipment, methods, emotions, time indicators, weather conditions, specific actions
- Consider synonyms and related terms that might be useful for searching
- Pay attention to implicit information that should be made explicit through keywords
- Cross-reference ALL names in both report AND Context - if a match is found in either, include complete details
- For ANY person found in report OR Context, MUST include their current/previous profession/role and name in keywords
- For developers/programmers in report OR Context, MUST include their specific technology stack and Polish role name
- If you find any person in Context who matches report criteria (e.g. JavaScript developer), MUST include their details even if not directly mentioned
- If the report contains any information about animals, MUST include the animal name and species in the keywords
- For any evidence (fingerprints, DNA, etc.), MUST include the specific location/sector where it was found
- If you find missing keywords, add them to the existing list
- Keywords should be in Polish
- Return keywords as comma-separated string
- Add _thinking to show your analysis process including Context cross-referencing
</rules>

<response_example>
Return the analysis in this exact format:
	{
		"_thinking": "Analysis of existing keywords and identification of missing ones. Cross-referenced all names with Context information. Identified specific locations of evidence.",
		"finalResult": {
			"filename.txt": "Mateusz,Kowalski,programista,frontend,React,JavaScript,ruch,bezpieczeństwo,sektorA,odciskipalców",
			"filename2.txt": "monitoring,peryferia,aktywność,cisza",
			"filename3.txt": "strażnik,czujnik,system,raport"
		}
	}
}
</response_example>
`

const extractKeyInformationPrompt = `
<context>%s</context>
You are an AI assistant specialized in extracting key information from text.
Please analyze the provided text and extract:
For the report:
- Generate initial keywords based on its content and file name.
- Identify the people/places mentioned in the report.
- Select matching 'facts' (e.g. relating to the same people).
- Combine keywords from the report with keywords derived from related facts.
Format the output as a clear, structured list in Polish.`

const extractFactsFilesKeywordsPrompt = `You are an AI assistant specialized in extracting key information from text.
Please analyze the provided text and extract key information such as:
- Names of individuals mentioned
- Their professions or roles
- Any special skills or abilities
Format the output as a clear, structured list in Polish.`
