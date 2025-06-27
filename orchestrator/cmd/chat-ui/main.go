package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

// ChatServer handles HTTP requests and makes API calls to WebBFF
type ChatServer struct {
	webBFFURL string
}

// ChatRequest represents the request to WebBFF API
type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// ChatResponse represents the response from WebBFF API
type ChatResponse struct {
	Success   bool   `json:"success"`
	Content   string `json:"content"`
	SessionID string `json:"session_id"`
	Intent    string `json:"intent,omitempty"`
	Error     string `json:"error,omitempty"`
}

func main() {
	// 🎯 REFACTORED: Chat UI as standalone service that calls WebBFF API
	chatServer := &ChatServer{
		webBFFURL: "http://localhost:8081", // WebBFF API URL
	}

	// Setup routes
	http.HandleFunc("/", chatServer.handleHome)
	http.HandleFunc("/conversation", chatServer.handleConversation)

	fmt.Println("🚀 AI Orchestrator Chat UI starting on http://localhost:8080")
	fmt.Println("🌐 Connecting to WebBFF API at http://localhost:8081")
	fmt.Println("💬 Open your browser to start chatting with the AI orchestrator!")
	fmt.Println("🔥 Now with REAL AI responses via WebBFF!")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleHome serves the chat HTML page
func (cs *ChatServer) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>AI Orchestrator Chat</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; border-radius: 10px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .header p { margin: 5px 0 0 0; opacity: 0.9; }
        .chat-container { min-height: 500px; max-height: 600px; overflow-y: auto; padding: 20px; }
        .message { margin: 15px 0; padding: 15px; border-radius: 8px; animation: fadeIn 0.3s ease-in; }
        .user-message { background: #e3f2fd; border-left: 4px solid #2563eb; margin-left: 20%; }
        .ai-message { background: #f3e5f5; border-left: 4px solid #9c27b0; margin-right: 20%; }
        .system-message { background: #f0f0f0; border-left: 4px solid #666; font-style: italic; text-align: center; }
        .message-header { font-size: 12px; color: #666; margin-bottom: 8px; font-weight: bold; }
        .message-content { line-height: 1.5; white-space: pre-wrap; }
        .typing { color: #2563eb; font-style: italic; }
        .input-container { padding: 20px; background: #f8f9fa; border-top: 1px solid #eee; }
        .input-group { display: flex; gap: 10px; }
        .message-input { flex: 1; padding: 12px; border: 1px solid #ddd; border-radius: 5px; font-size: 16px; }
        .send-button { padding: 12px 24px; background: #2563eb; color: white; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; }
        .send-button:hover { background: #1d4ed8; }
        .send-button:disabled { background: #9ca3af; cursor: not-allowed; }
        .examples { margin: 20px 0; }
        .example-btn { display: inline-block; margin: 5px; padding: 8px 16px; background: #f0f0f0; border: 1px solid #ddd; border-radius: 20px; cursor: pointer; font-size: 14px; transition: all 0.2s; }
        .example-btn:hover { background: #e0e0e0; transform: translateY(-1px); }
        .status { padding: 10px; text-align: center; color: #666; font-size: 14px; }
        .status.connected { color: #22c55e; }
        .status.thinking { color: #f59e0b; }
        .status.error { color: #ef4444; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
        .loading-dots { display: inline-block; }
        .loading-dots:after { content: '...'; animation: dots 1.5s steps(5, end) infinite; }
        @keyframes dots { 0%, 20% { color: rgba(0,0,0,0); text-shadow: .25em 0 0 rgba(0,0,0,0), .5em 0 0 rgba(0,0,0,0); }
          40% { color: black; text-shadow: .25em 0 0 rgba(0,0,0,0), .5em 0 0 rgba(0,0,0,0); }
          60% { text-shadow: .25em 0 0 black, .5em 0 0 rgba(0,0,0,0); }
          80%, 100% { text-shadow: .25em 0 0 black, .5em 0 0 black; } }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 AI Orchestrator Chat</h1>
            <p>Real-time conversation with the AI orchestrator</p>
        </div>
        
        <div class="chat-container" id="chatContainer">
            <div class="message ai-message">
                <div class="message-header">🤖 AI Orchestrator</div>
                <div class="message-content">Hello! I'm the AI orchestrator. I can help you with:

🚀 Application deployment guidance
🔧 Troubleshooting technical issues  
⚙️ Setting up CI/CD pipelines
🤝 Coordinating multiple agents
📊 Workflow orchestration

What would you like to do today?</div>
            </div>
            
            <div class="examples">
                <strong>💡 Try these examples:</strong><br>
                <span class="example-btn" onclick="setMessage('I need help deploying a Node.js application with Docker. What are the essential steps?')">Node.js Deployment</span>
                <span class="example-btn" onclick="setMessage('My deployment failed with database connection error ECONNREFUSED. How do I troubleshoot this?')">Troubleshoot Error</span>
                <span class="example-btn" onclick="setMessage('Set up a complete CI/CD pipeline for microservices with 3 services')">CI/CD Pipeline</span>
                <span class="example-btn" onclick="setMessage('What agents are available in the system and what can they do?')">List Agents</span>
            </div>
        </div>
        
        <div class="status connected" id="status">✅ Connected to AI orchestrator</div>
        
        <div class="input-container">
            <form id="chatForm" onsubmit="sendMessage(event)">
                <div class="input-group">
                    <input type="text" id="messageInput" name="message" class="message-input" placeholder="Ask the AI orchestrator anything..." required />
                    <button type="submit" id="sendButton" class="send-button">Send</button>
                </div>
            </form>
        </div>
    </div>

    <script>
        let conversationId = 'web-user-' + Date.now();
        
        function setMessage(text) {
            document.getElementById('messageInput').value = text;
        }

        function addMessage(type, content, sender = '') {
            const chatContainer = document.getElementById('chatContainer');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message ' + type + '-message';
            
            const headerDiv = document.createElement('div');
            headerDiv.className = 'message-header';
            headerDiv.textContent = sender || (type === 'user' ? '👤 You' : '🤖 AI Orchestrator');
            
            const contentDiv = document.createElement('div');
            contentDiv.className = 'message-content';
            contentDiv.textContent = content;
            
            messageDiv.appendChild(headerDiv);
            messageDiv.appendChild(contentDiv);
            chatContainer.appendChild(messageDiv);
            
            // Scroll to bottom
            chatContainer.scrollTop = chatContainer.scrollHeight;
            
            return messageDiv;
        }

        function setStatus(message, className = '') {
            const status = document.getElementById('status');
            status.textContent = message;
            status.className = 'status ' + className;
        }

        async function sendMessage(event) {
            event.preventDefault();
            
            const messageInput = document.getElementById('messageInput');
            const sendButton = document.getElementById('sendButton');
            const message = messageInput.value.trim();
            
            if (!message) return;

            // Disable input
            sendButton.disabled = true;
            messageInput.disabled = true;

            // Add user message
            addMessage('user', message);
            
            // Show AI is thinking
            const thinkingMsg = addMessage('ai', 'AI is thinking and processing your request', '🤖 AI Orchestrator');
            thinkingMsg.classList.add('typing');
            setStatus('🤔 AI orchestrator is thinking...', 'thinking');

            // Clear input
            messageInput.value = '';

            try {
                const response = await fetch('/conversation', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: 'message=' + encodeURIComponent(message) + '&conversation_id=' + encodeURIComponent(conversationId)
                });

                if (!response.ok) {
                    throw new Error('Failed to send message: ' + response.statusText);
                }

                const result = await response.text();
                
                // Remove thinking message
                thinkingMsg.remove();
                
                // Add AI response
                addMessage('ai', result);
                
                setStatus('✅ Connected to AI orchestrator', 'connected');
                
            } catch (error) {
                // Remove thinking message
                thinkingMsg.remove();
                
                // Add error message
                addMessage('system', 'Error: ' + error.message);
                setStatus('❌ Connection error', 'error');
            } finally {
                // Re-enable input
                sendButton.disabled = false;
                messageInput.disabled = false;
                messageInput.focus();
            }
        }
        
        // Focus input on load
        window.onload = function() {
            document.getElementById('messageInput').focus();
        };
    </script>
</body>
</html>`

	t, _ := template.New("chat").Parse(tmpl)
	t.Execute(w, nil)
}

// handleConversation handles real-time conversation via WebBFF API
func (cs *ChatServer) handleConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	message := r.FormValue("message")
	conversationID := r.FormValue("conversation_id")

	if message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	if conversationID == "" {
		conversationID = fmt.Sprintf("web-user-%d", time.Now().UnixNano())
	}

	log.Printf("🔄 Processing message via WebBFF API: %s (session: %s)", message, conversationID)

	// 🚀 REFACTORED: Make HTTP call to WebBFF API
	chatReq := ChatRequest{
		SessionID: conversationID,
		Message:   message,
	}

	// Marshal request
	jsonData, err := json.Marshal(chatReq)
	if err != nil {
		log.Printf("❌ Failed to marshal request: %v", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}

	// Make HTTP request to WebBFF
	resp, err := http.Post(cs.webBFFURL+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ WebBFF API call failed: %v", err)
		http.Error(w, "Failed to connect to AI service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read WebBFF response: %v", err)
		http.Error(w, "Failed to read AI response", http.StatusInternalServerError)
		return
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ WebBFF API returned status %d: %s", resp.StatusCode, string(body))
		http.Error(w, "AI service error", http.StatusInternalServerError)
		return
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		log.Printf("❌ Failed to parse WebBFF response: %v", err)
		http.Error(w, "Failed to parse AI response", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ WebBFF response: %s", chatResp.Content)

	// Return the AI response
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, chatResp.Content)
}
