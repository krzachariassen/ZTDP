package ai

import (
	"context"

	"github.com/krzachariassen/ZTDP/internal/graph"
)

// PlanningRequest contains all context needed for AI planning
type PlanningRequest struct {
	Intent        string                 `json:"intent"`         // Human-readable deployment intent
	ApplicationID string                 `json:"application_id"` // Target application
	EdgeTypes     []string               `json:"edge_types"`     // Edge types to consider (deploy, create, owns, etc.)
	Context       *PlanningContext       `json:"context"`        // Complete graph context
	Metadata      map[string]interface{} `json:"metadata"`       // Additional metadata
}

// PlanningContext provides complete graph state for AI reasoning
type PlanningContext struct {
	TargetNodes   []*graph.Node `json:"target_nodes"`   // Nodes in deployment scope
	RelatedNodes  []*graph.Node `json:"related_nodes"`  // Dependencies and related nodes
	Edges         []*graph.Edge `json:"edges"`          // All relevant edges
	PolicyContext interface{}   `json:"policy_context"` // Policy constraints (flexible type)
	EnvironmentID string        `json:"environment_id"` // Target environment
}

// PlanningResponse contains AI-generated deployment plan with reasoning
type PlanningResponse struct {
	Plan       *DeploymentPlan        `json:"plan"`       // Generated deployment plan
	Reasoning  string                 `json:"reasoning"`  // AI reasoning explanation
	Confidence float64                `json:"confidence"` // AI confidence score (0-1)
	Metadata   map[string]interface{} `json:"metadata"`   // Additional response metadata
}

// DeploymentPlan represents an AI-generated deployment plan
type DeploymentPlan struct {
	Steps      []*DeploymentStep      `json:"steps"`      // Ordered deployment steps
	Strategy   string                 `json:"strategy"`   // Deployment strategy (rolling, blue-green, etc.)
	Validation []string               `json:"validation"` // Validation checks to perform
	Rollback   *RollbackPlan          `json:"rollback"`   // Rollback plan if needed
	Metadata   map[string]interface{} `json:"metadata"`   // Additional plan metadata
}

// DeploymentStep represents a single step in the deployment plan
type DeploymentStep struct {
	ID           string                 `json:"id"`           // Unique step identifier
	Action       string                 `json:"action"`       // Action to perform (deploy, create, configure, etc.)
	Target       string                 `json:"target"`       // Target node/resource ID
	Dependencies []string               `json:"dependencies"` // Step dependencies
	Metadata     map[string]interface{} `json:"metadata"`     // Step-specific metadata
	Reasoning    string                 `json:"reasoning"`    // Why this step is needed
}

// RollbackPlan contains instructions for rolling back a deployment
type RollbackPlan struct {
	Steps    []*DeploymentStep      `json:"steps"`    // Rollback steps
	Triggers []string               `json:"triggers"` // Conditions that trigger rollback
	Metadata map[string]interface{} `json:"metadata"` // Rollback metadata
}

// AIProvider defines the interface for AI reasoning providers
// This follows the same pattern as GraphBackend for clean abstraction
type AIProvider interface {
	// GeneratePlan creates an intelligent deployment plan using AI reasoning
	GeneratePlan(ctx context.Context, request *PlanningRequest) (*PlanningResponse, error)

	// EvaluatePolicy uses AI to evaluate policy compliance and suggest actions
	EvaluatePolicy(ctx context.Context, policyContext interface{}) (*PolicyEvaluation, error)

	// OptimizePlan refines an existing plan based on additional context
	OptimizePlan(ctx context.Context, plan *DeploymentPlan, context *PlanningContext) (*PlanningResponse, error)

	// *** REVOLUTIONARY AI CAPABILITIES - IMPOSSIBLE WITH TRADITIONAL IDPS ***

	// ChatWithPlatform enables natural language interaction with the platform graph
	// This allows developers to ask complex questions about their infrastructure
	ChatWithPlatform(ctx context.Context, query *ConversationalQuery) (*ConversationalResponse, error)

	// PredictImpact analyzes potential impact of changes before they happen
	// Uses AI to simulate and predict deployment consequences
	PredictImpact(ctx context.Context, request *ImpactAnalysisRequest) (*ImpactPrediction, error)

	// IntelligentTroubleshooting provides AI-driven root cause analysis
	// Analyzes failures and suggests intelligent fixes
	IntelligentTroubleshooting(ctx context.Context, incident *IncidentContext) (*TroubleshootingResponse, error)

	// ProactiveOptimization continuously analyzes platform for improvements
	// Identifies patterns and suggests architectural optimizations
	ProactiveOptimization(ctx context.Context, scope *OptimizationScope) (*OptimizationRecommendations, error)

	// LearningFromFailures learns from deployment patterns and failures
	// Builds institutional knowledge that improves over time
	LearningFromFailures(ctx context.Context, outcome *DeploymentOutcome) (*LearningInsights, error)

	// GetProviderInfo returns information about the AI provider
	GetProviderInfo() *ProviderInfo

	// Close cleans up provider resources
	Close() error
}

// PolicyViolation represents a specific policy violation
type PolicyViolation struct {
	PolicyID    string                 `json:"policy_id"`   // ID of violated policy
	Severity    string                 `json:"severity"`    // Violation severity (low, medium, high, critical)
	Description string                 `json:"description"` // Violation description
	Suggestion  string                 `json:"suggestion"`  // Suggested fix
	Metadata    map[string]interface{} `json:"metadata"`    // Additional violation metadata
}

// PolicyEvaluation contains AI-driven policy evaluation results
type PolicyEvaluation struct {
	Compliant       bool                   `json:"compliant"`       // Whether policies are satisfied
	Violations      []PolicyViolation      `json:"violations"`      // Policy violations found
	Recommendations []string               `json:"recommendations"` // AI recommendations for compliance
	Reasoning       string                 `json:"reasoning"`       // AI reasoning for evaluation
	Confidence      float64                `json:"confidence"`      // Confidence in evaluation
	Metadata        map[string]interface{} `json:"metadata"`        // Additional evaluation metadata
}

// ProviderInfo contains metadata about an AI provider
type ProviderInfo struct {
	Name         string                 `json:"name"`         // Provider name (e.g., "openai-gpt4")
	Version      string                 `json:"version"`      // Provider version
	Capabilities []string               `json:"capabilities"` // Supported capabilities
	Metadata     map[string]interface{} `json:"metadata"`     // Provider-specific metadata
}

// *** REVOLUTIONARY AI DATA STRUCTURES - GROUNDBREAKING CAPABILITIES ***

// ConversationalQuery represents a natural language query to the platform
type ConversationalQuery struct {
	Query    string                 `json:"query"`    // Natural language question
	Context  string                 `json:"context"`  // Current context (app, env, etc.)
	Intent   string                 `json:"intent"`   // Detected intent (question, command, analysis)
	Scope    []string               `json:"scope"`    // Scope of query (applications, services, etc.)
	Metadata map[string]interface{} `json:"metadata"` // Query metadata
}

// ConversationalResponse contains the AI's response to conversational queries
type ConversationalResponse struct {
	Answer     string                 `json:"answer"`     // Natural language response
	Insights   []string               `json:"insights"`   // Key insights discovered
	Actions    []SuggestedAction      `json:"actions"`    // Suggested actions
	Confidence float64                `json:"confidence"` // Confidence in response
	FollowUp   []string               `json:"follow_up"`  // Follow-up questions
	Reasoning  string                 `json:"reasoning"`  // AI reasoning process
	GraphData  interface{}            `json:"graph_data"` // Relevant graph visualization data
	Metadata   map[string]interface{} `json:"metadata"`   // Response metadata
}

// SuggestedAction represents an action the AI recommends
type SuggestedAction struct {
	ID          string                 `json:"id"`          // Unique action ID
	Title       string                 `json:"title"`       // Human-readable title
	Description string                 `json:"description"` // Detailed description
	Type        string                 `json:"type"`        // Action type (deploy, scale, fix, etc.)
	Urgency     string                 `json:"urgency"`     // low, medium, high, critical
	Impact      string                 `json:"impact"`      // Predicted impact
	Command     string                 `json:"command"`     // Executable command (if applicable)
	Metadata    map[string]interface{} `json:"metadata"`    // Action metadata
}

// ImpactAnalysisRequest requests prediction of deployment impact
type ImpactAnalysisRequest struct {
	Changes     []ProposedChange       `json:"changes"`     // Proposed changes
	Scope       string                 `json:"scope"`       // Analysis scope
	Environment string                 `json:"environment"` // Target environment
	Timeframe   string                 `json:"timeframe"`   // Analysis timeframe
	Metadata    map[string]interface{} `json:"metadata"`    // Request metadata
}

// ProposedChange represents a change to be analyzed
type ProposedChange struct {
	Type     string                 `json:"type"`     // Change type (deploy, delete, update)
	Target   string                 `json:"target"`   // Target resource
	Details  map[string]interface{} `json:"details"`  // Change details
	Metadata map[string]interface{} `json:"metadata"` // Change metadata
}

// ImpactPrediction contains AI-predicted impact of changes
type ImpactPrediction struct {
	OverallRisk     string                 `json:"overall_risk"`     // low, medium, high, critical
	AffectedSystems []string               `json:"affected_systems"` // Systems that will be affected
	Predictions     []ImpactForecast       `json:"predictions"`      // Specific predictions
	Recommendations []string               `json:"recommendations"`  // Risk mitigation recommendations
	Confidence      float64                `json:"confidence"`       // Confidence in predictions
	SimulationData  interface{}            `json:"simulation_data"`  // Simulation results
	Reasoning       string                 `json:"reasoning"`        // AI reasoning
	Metadata        map[string]interface{} `json:"metadata"`         // Prediction metadata
}

// ImpactForecast represents a specific impact prediction
type ImpactForecast struct {
	System      string                 `json:"system"`      // Affected system
	Probability float64                `json:"probability"` // Probability of impact (0.0-1.0)
	Severity    string                 `json:"severity"`    // Impact severity
	Description string                 `json:"description"` // Impact description
	Timeline    string                 `json:"timeline"`    // When impact occurs
	Metadata    map[string]interface{} `json:"metadata"`    // Forecast metadata
}

// IncidentContext provides context for troubleshooting
type IncidentContext struct {
	IncidentID  string                 `json:"incident_id"` // Unique incident ID
	Description string                 `json:"description"` // Problem description
	Symptoms    []string               `json:"symptoms"`    // Observed symptoms
	Environment string                 `json:"environment"` // Environment where incident occurred
	Timeline    []EventTimestamp       `json:"timeline"`    // Timeline of events
	Logs        []LogEntry             `json:"logs"`        // Relevant log entries
	Metrics     map[string]interface{} `json:"metrics"`     // Relevant metrics
	Context     map[string]interface{} `json:"context"`     // Additional context
	Metadata    map[string]interface{} `json:"metadata"`    // Incident metadata
}

// EventTimestamp represents a timestamped event
type EventTimestamp struct {
	Timestamp string                 `json:"timestamp"` // Event timestamp
	Event     string                 `json:"event"`     // Event description
	Source    string                 `json:"source"`    // Event source
	Severity  string                 `json:"severity"`  // Event severity
	Metadata  map[string]interface{} `json:"metadata"`  // Event metadata
}

// LogEntry represents a log entry for analysis
type LogEntry struct {
	Timestamp string                 `json:"timestamp"` // Log timestamp
	Level     string                 `json:"level"`     // Log level
	Message   string                 `json:"message"`   // Log message
	Source    string                 `json:"source"`    // Log source
	Metadata  map[string]interface{} `json:"metadata"`  // Log metadata
}

// TroubleshootingResponse contains AI-driven troubleshooting analysis
type TroubleshootingResponse struct {
	RootCause     string                 `json:"root_cause"`     // Identified root cause
	Diagnosis     string                 `json:"diagnosis"`      // AI diagnosis
	Solutions     []Solution             `json:"solutions"`      // Proposed solutions
	Investigation []InvestigationStep    `json:"investigation"`  // Investigation steps
	Prevention    []string               `json:"prevention"`     // Prevention recommendations
	Confidence    float64                `json:"confidence"`     // Confidence in diagnosis
	SimilarIssues []HistoricalIncident   `json:"similar_issues"` // Similar past issues
	Reasoning     string                 `json:"reasoning"`      // AI reasoning process
	Metadata      map[string]interface{} `json:"metadata"`       // Response metadata
}

// Solution represents a proposed solution
type Solution struct {
	ID          string                 `json:"id"`          // Solution ID
	Title       string                 `json:"title"`       // Solution title
	Description string                 `json:"description"` // Solution description
	Steps       []string               `json:"steps"`       // Implementation steps
	Risk        string                 `json:"risk"`        // Risk level
	Effort      string                 `json:"effort"`      // Implementation effort
	Success     float64                `json:"success"`     // Success probability
	Metadata    map[string]interface{} `json:"metadata"`    // Solution metadata
}

// InvestigationStep represents a troubleshooting investigation step
type InvestigationStep struct {
	Step     string                 `json:"step"`     // Investigation step
	Command  string                 `json:"command"`  // Command to run (if applicable)
	Expected string                 `json:"expected"` // Expected result
	Purpose  string                 `json:"purpose"`  // Purpose of this step
	Metadata map[string]interface{} `json:"metadata"` // Step metadata
}

// HistoricalIncident represents a past incident for comparison
type HistoricalIncident struct {
	ID          string                 `json:"id"`          // Incident ID
	Similarity  float64                `json:"similarity"`  // Similarity score (0.0-1.0)
	Description string                 `json:"description"` // Incident description
	Resolution  string                 `json:"resolution"`  // How it was resolved
	Lessons     []string               `json:"lessons"`     // Lessons learned
	Metadata    map[string]interface{} `json:"metadata"`    // Incident metadata
}

// OptimizationScope defines the scope for proactive optimization
type OptimizationScope struct {
	Target      string                 `json:"target"`      // Optimization target (application, environment, etc.)
	Focus       []string               `json:"focus"`       // Focus areas (performance, cost, reliability)
	Timeframe   string                 `json:"timeframe"`   // Analysis timeframe
	Constraints []string               `json:"constraints"` // Optimization constraints
	Metadata    map[string]interface{} `json:"metadata"`    // Scope metadata
}

// OptimizationRecommendations contains AI-generated optimization suggestions
type OptimizationRecommendations struct {
	Summary         string                    `json:"summary"`         // Overall optimization summary
	Recommendations []Recommendation          `json:"recommendations"` // Specific recommendations
	Patterns        []DetectedPattern         `json:"patterns"`        // Detected patterns
	Opportunities   []OptimizationOpportunity `json:"opportunities"`   // Optimization opportunities
	Impact          ImpactAssessment          `json:"impact"`          // Expected impact
	Priority        []string                  `json:"priority"`        // Prioritized action list
	Reasoning       string                    `json:"reasoning"`       // AI reasoning
	Metadata        map[string]interface{}    `json:"metadata"`        // Recommendations metadata
}

// Recommendation represents a specific optimization recommendation
type Recommendation struct {
	ID          string                 `json:"id"`          // Recommendation ID
	Category    string                 `json:"category"`    // Category (performance, cost, etc.)
	Title       string                 `json:"title"`       // Recommendation title
	Description string                 `json:"description"` // Detailed description
	Benefits    []string               `json:"benefits"`    // Expected benefits
	Effort      string                 `json:"effort"`      // Implementation effort
	Priority    string                 `json:"priority"`    // Priority level
	Actions     []SuggestedAction      `json:"actions"`     // Required actions
	Metadata    map[string]interface{} `json:"metadata"`    // Recommendation metadata
}

// DetectedPattern represents a pattern detected by AI
type DetectedPattern struct {
	Pattern      string                 `json:"pattern"`      // Pattern description
	Frequency    int                    `json:"frequency"`    // How often pattern occurs
	Significance string                 `json:"significance"` // Pattern significance
	Examples     []string               `json:"examples"`     // Example occurrences
	Metadata     map[string]interface{} `json:"metadata"`     // Pattern metadata
}

// OptimizationOpportunity represents an opportunity for improvement
type OptimizationOpportunity struct {
	Area       string                 `json:"area"`       // Area of opportunity
	Potential  string                 `json:"potential"`  // Potential improvement
	Complexity string                 `json:"complexity"` // Implementation complexity
	ROI        string                 `json:"roi"`        // Return on investment
	Timeline   string                 `json:"timeline"`   // Implementation timeline
	Metadata   map[string]interface{} `json:"metadata"`   // Opportunity metadata
}

// ImpactAssessment contains impact assessment for optimizations
type ImpactAssessment struct {
	Performance string                 `json:"performance"` // Performance impact
	Cost        string                 `json:"cost"`        // Cost impact
	Reliability string                 `json:"reliability"` // Reliability impact
	Security    string                 `json:"security"`    // Security impact
	Maintenance string                 `json:"maintenance"` // Maintenance impact
	Metadata    map[string]interface{} `json:"metadata"`    // Assessment metadata
}

// DeploymentOutcome represents the outcome of a deployment for learning
type DeploymentOutcome struct {
	DeploymentID string                 `json:"deployment_id"` // Deployment ID
	Success      bool                   `json:"success"`       // Whether deployment succeeded
	Duration     int64                  `json:"duration"`      // Deployment duration (seconds)
	Issues       []DeploymentIssue      `json:"issues"`        // Issues encountered
	Metrics      map[string]interface{} `json:"metrics"`       // Performance metrics
	Context      map[string]interface{} `json:"context"`       // Deployment context
	Metadata     map[string]interface{} `json:"metadata"`      // Outcome metadata
}

// DeploymentIssue represents an issue during deployment
type DeploymentIssue struct {
	Type        string                 `json:"type"`        // Issue type
	Description string                 `json:"description"` // Issue description
	Severity    string                 `json:"severity"`    // Issue severity
	Resolution  string                 `json:"resolution"`  // How it was resolved
	Timestamp   string                 `json:"timestamp"`   // When issue occurred
	Metadata    map[string]interface{} `json:"metadata"`    // Issue metadata
}

// LearningInsights contains insights learned from deployment outcomes
type LearningInsights struct {
	Insights      []Insight              `json:"insights"`      // Learned insights
	Patterns      []LearnedPattern       `json:"patterns"`      // Learned patterns
	Improvements  []ProcessImprovement   `json:"improvements"`  // Process improvements
	Predictions   []FuturePrediction     `json:"predictions"`   // Future predictions
	Confidence    float64                `json:"confidence"`    // Confidence in insights
	Applicability []string               `json:"applicability"` // Where insights apply
	Reasoning     string                 `json:"reasoning"`     // AI reasoning
	Metadata      map[string]interface{} `json:"metadata"`      // Insights metadata
}

// Insight represents a specific learned insight
type Insight struct {
	ID          string                 `json:"id"`          // Insight ID
	Type        string                 `json:"type"`        // Insight type
	Description string                 `json:"description"` // Insight description
	Evidence    []string               `json:"evidence"`    // Supporting evidence
	Confidence  float64                `json:"confidence"`  // Confidence in insight
	Impact      string                 `json:"impact"`      // Potential impact
	Metadata    map[string]interface{} `json:"metadata"`    // Insight metadata
}

// LearnedPattern represents a pattern learned from deployment data
type LearnedPattern struct {
	Pattern     string                 `json:"pattern"`     // Pattern description
	Conditions  []string               `json:"conditions"`  // Conditions where pattern applies
	Outcomes    []string               `json:"outcomes"`    // Typical outcomes
	Reliability float64                `json:"reliability"` // Pattern reliability
	Metadata    map[string]interface{} `json:"metadata"`    // Pattern metadata
}

// ProcessImprovement represents a suggested process improvement
type ProcessImprovement struct {
	Area     string                 `json:"area"`     // Area for improvement
	Current  string                 `json:"current"`  // Current process
	Proposed string                 `json:"proposed"` // Proposed improvement
	Benefits []string               `json:"benefits"` // Expected benefits
	Effort   string                 `json:"effort"`   // Implementation effort
	Metadata map[string]interface{} `json:"metadata"` // Improvement metadata
}

// FuturePrediction represents a prediction about future deployments
type FuturePrediction struct {
	Prediction string                 `json:"prediction"` // Prediction description
	Likelihood float64                `json:"likelihood"` // Likelihood (0.0-1.0)
	Timeframe  string                 `json:"timeframe"`  // When prediction applies
	Indicators []string               `json:"indicators"` // Leading indicators
	Actions    []string               `json:"actions"`    // Recommended actions
	Metadata   map[string]interface{} `json:"metadata"`   // Prediction metadata
}
