package agent

type SubmitAnalysisInput struct {
	Analysis AnalysisResult `json:"analysis" description:"The structured analysis of the uploaded documents"`
}

type AnalysisResult struct {
	Summary        string          `json:"summary" description:"Overall summary across all documents, this should be a summary of the contents and key insights, not a description of the analysis process"`
	Documents      []DocumentInfo  `json:"documents" description:"Per-document summaries"`
	Decisions      []Decision      `json:"decisions" description:"Key decisions found across documents"`
	Owners         []Ownership     `json:"owners" description:"Ownership assignments found"`
	Deadlines      []Deadline      `json:"deadlines" description:"Deadlines mentioned"`
	Contradictions []Contradiction `json:"contradictions" description:"Contradictions found across documents"`
}

type DocumentInfo struct {
	ID      string `json:"id" description:"Document ID"`
	Name    string `json:"name" description:"Document filename"`
	Summary string `json:"summary" description:"Brief summary of this document"`
}

type Decision struct {
	Description string    `json:"description" description:"What was decided"`
	Status      *string   `json:"status,omitempty" description:"Current status: decided, reversed, pending"`
	Source      *Citation `json:"source,omitempty" description:"Where this was found"`
	Reasoning   string    `json:"reasoning,omitempty" description:"Why this was identified, required when source is not provided"`
}

type Ownership struct {
	Item      string    `json:"item" description:"What is owned"`
	Owner     string    `json:"owner" description:"Who owns it"`
	Source    *Citation `json:"source,omitempty" description:"Where this was found"`
	Reasoning string    `json:"reasoning,omitempty" description:"Why this was identified, required when source is not provided"`
}

type Deadline struct {
	Item      string    `json:"item" description:"What has the deadline"`
	Date      string    `json:"date" description:"The deadline date or timeframe"`
	Source    *Citation `json:"source,omitempty" description:"Where this was found"`
	Reasoning string    `json:"reasoning,omitempty" description:"Why this was identified, required when source is not provided"`
}

type Contradiction struct {
	Description string  `json:"description" description:"Description of the contradiction"`
	Claims      []Claim `json:"claims" description:"The conflicting claims"`
}

type Claim struct {
	Statement string    `json:"statement" description:"What was claimed"`
	Source    *Citation `json:"source,omitempty" description:"Where this claim was found"`
	Reasoning string    `json:"reasoning,omitempty" description:"Why this was identified, required when source is not provided"`
}

type Citation struct {
	DocumentID   string `json:"document_id" description:"ID of the source document"`
	DocumentName string `json:"document_name" description:"Name of the source document"`
	Excerpt      string `json:"excerpt" description:"Relevant excerpt from the document"`
}
