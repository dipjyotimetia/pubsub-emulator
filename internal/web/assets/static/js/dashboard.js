// Dashboard JavaScript

let messages = [];
let currentMessageId = null;
let topics = [];
let subscriptions = [];

// Initialize dashboard on page load
document.addEventListener('DOMContentLoaded', function() {
    loadStats();
    loadMessages();
    setupSearchHandlers();

    // Refresh stats every 5 seconds
    setInterval(loadStats, 5000);

    // Refresh messages every 10 seconds
    setInterval(loadMessages, 10000);
});

// Load Statistics
async function loadStats() {
    try {
        const response = await fetch('/api/stats');
        const stats = await response.json();
        updateStats(stats);
    } catch (error) {
        console.error('Error loading stats:', error);
    }
}

function updateStats(stats) {
    document.getElementById('topicCount').textContent = stats.topics || 0;
    document.getElementById('subscriptionCount').textContent = stats.subscriptions || 0;
    document.getElementById('messageCount').textContent = stats.total_messages || 0;
    
    if (stats.last_message_time) {
        const time = new Date(stats.last_message_time);
        document.getElementById('lastMessage').textContent = formatTimeAgo(time);
    }
    
    // Update topics list
    if (stats.topic_list) {
        topics = stats.topic_list;
        updateTopicSelects();
        updateTopicFilter();
    }
    
    if (stats.subscription_list) {
        subscriptions = stats.subscription_list;
    }
}

// Load Messages
async function loadMessages() {
    try {
        const response = await fetch('/api/messages');
        messages = await response.json();
        renderMessages(messages);
        updateMessageBadge(messages.length);
    } catch (error) {
        console.error('Error loading messages:', error);
    }
}

function renderMessages(messagesToRender) {
    const container = document.getElementById('messagesContainer');
    
    if (!messagesToRender || messagesToRender.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">📭</div>
                <p>No messages yet. Publish a message to get started!</p>
            </div>
        `;
        return;
    }
    
    // Sort messages by received time (newest first)
    messagesToRender.sort((a, b) => new Date(b.received) - new Date(a.received));
    
    container.innerHTML = messagesToRender.map(msg => createMessageCard(msg)).join('');
}

function createMessageCard(msg) {
    const receivedTime = new Date(msg.received);
    const publishTime = new Date(msg.publish_time);
    
    return `
        <div class="message-card" data-message-id="${msg.id}">
            <div class="message-header">
                <span class="message-id">ID: ${msg.id}</span>
                <span class="message-topic">${msg.topic}</span>
            </div>
            <div class="message-data">${escapeHtml(msg.data)}</div>
            <div class="message-footer">
                <div class="message-time">
                    <span>📤 Published: ${formatTime(publishTime)}</span>
                    <span>📥 Received: ${formatTime(receivedTime)}</span>
                </div>
                <div class="message-actions">
                    <button class="btn btn-info" onclick="showMessageDetails('${msg.id}')">View</button>
                    <button class="btn btn-primary" onclick="replayMessage('${msg.id}')">🔄 Replay</button>
                </div>
            </div>
        </div>
    `;
}

// Search and Filter
function setupSearchHandlers() {
    const searchInput = document.getElementById('searchInput');
    const topicFilter = document.getElementById('topicFilter');
    
    searchInput.addEventListener('input', debounce(performSearch, 300));
    topicFilter.addEventListener('change', performSearch);
}

function performSearch() {
    const searchTerm = document.getElementById('searchInput').value.toLowerCase();
    const topicFilter = document.getElementById('topicFilter').value;
    
    let filtered = messages;
    
    // Apply topic filter
    if (topicFilter) {
        filtered = filtered.filter(msg => msg.topic === topicFilter);
    }
    
    // Apply search term
    if (searchTerm) {
        filtered = filtered.filter(msg => 
            msg.data.toLowerCase().includes(searchTerm) || 
            msg.id.toLowerCase().includes(searchTerm)
        );
    }
    
    renderMessages(filtered);
    updateMessageBadge(filtered.length);
}

function updateTopicFilter() {
    const select = document.getElementById('topicFilter');
    const currentValue = select.value;
    
    select.innerHTML = '<option value="">All Topics</option>';
    topics.forEach(topic => {
        const option = document.createElement('option');
        option.value = topic;
        option.textContent = topic;
        select.appendChild(option);
    });
    
    select.value = currentValue;
}

function updateTopicSelects() {
    updateSelect('publishTopic', topics);
    updateSelect('subscriptionTopic', topics);
}

function updateSelect(selectId, options) {
    const select = document.getElementById(selectId);
    if (!select) return;
    
    const currentValue = select.value;
    select.innerHTML = '';
    
    options.forEach(option => {
        const opt = document.createElement('option');
        opt.value = option;
        opt.textContent = option;
        select.appendChild(opt);
    });
    
    if (options.includes(currentValue)) {
        select.value = currentValue;
    }
}

// Modal Functions
function showPublishModal() {
    if (topics.length === 0) {
        showToast('No topics available. Create a topic first.', 'error');
        return;
    }
    updateTopicSelects();
    openModal('publishModal');
}

function showCreateTopicModal() {
    openModal('createTopicModal');
}

function showCreateSubscriptionModal() {
    if (topics.length === 0) {
        showToast('No topics available. Create a topic first.', 'error');
        return;
    }
    updateTopicSelects();
    openModal('createSubscriptionModal');
}

function openModal(modalId) {
    document.getElementById(modalId).classList.add('show');
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.remove('show');
}

// Publish Message
async function publishMessage() {
    const topic = document.getElementById('publishTopic').value;
    const data = document.getElementById('publishData').value;
    const attributesText = document.getElementById('publishAttributes').value;
    
    if (!data) {
        showToast('Message data is required', 'error');
        return;
    }
    
    let attributes = {};
    if (attributesText) {
        try {
            attributes = JSON.parse(attributesText);
        } catch (e) {
            showToast('Invalid JSON in attributes', 'error');
            return;
        }
    }
    
    try {
        const response = await fetch('/api/publish', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ topic_id: topic, data: data, attributes: attributes })
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showToast('Message published successfully!', 'success');
            closeModal('publishModal');
            document.getElementById('publishData').value = '';
            document.getElementById('publishAttributes').value = '';

            // Reload messages immediately
            loadMessages();
        } else {
            showToast('Failed to publish message: ' + result.error, 'error');
        }
    } catch (error) {
        console.error('Error publishing message:', error);
        showToast('Error publishing message', 'error');
    }
}

// Create Topic
async function createTopic() {
    const topicId = document.getElementById('topicId').value.trim();
    
    if (!topicId) {
        showToast('Topic ID is required', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/topics', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ topic_id: topicId })
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showToast('Topic created successfully!', 'success');
            closeModal('createTopicModal');
            document.getElementById('topicId').value = '';
            loadStats();
        } else {
            showToast('Failed to create topic', 'error');
        }
    } catch (error) {
        console.error('Error creating topic:', error);
        showToast('Error creating topic', 'error');
    }
}

// Create Subscription
async function createSubscription() {
    const subscriptionId = document.getElementById('subscriptionId').value.trim();
    const topicId = document.getElementById('subscriptionTopic').value;
    const ackDeadline = parseInt(document.getElementById('ackDeadline').value);
    
    if (!subscriptionId) {
        showToast('Subscription ID is required', 'error');
        return;
    }
    
    try {
        const response = await fetch('/api/subscriptions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                subscription_id: subscriptionId, 
                topic_id: topicId,
                ack_deadline_seconds: ackDeadline
            })
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showToast('Subscription created successfully!', 'success');
            closeModal('createSubscriptionModal');
            document.getElementById('subscriptionId').value = '';
            loadStats();
        } else {
            showToast('Failed to create subscription', 'error');
        }
    } catch (error) {
        console.error('Error creating subscription:', error);
        showToast('Error creating subscription', 'error');
    }
}

// Message Details
function showMessageDetails(messageId) {
    const msg = messages.find(m => m.id === messageId);
    if (!msg) return;
    
    currentMessageId = messageId;
    
    const modalBody = document.getElementById('messageModalBody');
    modalBody.innerHTML = `
        <div class="detail-row">
            <div class="detail-label">Message ID</div>
            <div class="detail-value">${msg.id}</div>
        </div>
        <div class="detail-row">
            <div class="detail-label">Topic</div>
            <div class="detail-value">${msg.topic}</div>
        </div>
        <div class="detail-row">
            <div class="detail-label">Data</div>
            <div class="detail-value">${escapeHtml(msg.data)}</div>
        </div>
        <div class="detail-row">
            <div class="detail-label">Publish Time</div>
            <div class="detail-value">${new Date(msg.publish_time).toLocaleString()}</div>
        </div>
        <div class="detail-row">
            <div class="detail-label">Received Time</div>
            <div class="detail-value">${new Date(msg.received).toLocaleString()}</div>
        </div>
        <div class="detail-row">
            <div class="detail-label">Attributes</div>
            <div class="attributes-list">
                ${Object.entries(msg.attributes || {}).map(([key, value]) => `
                    <div class="attribute-item">
                        <span class="attribute-key">${escapeHtml(key)}:</span>
                        <span class="attribute-value">${escapeHtml(value)}</span>
                    </div>
                `).join('') || '<div class="detail-value">No attributes</div>'}
            </div>
        </div>
    `;
    
    openModal('messageModal');
}

// Replay Message
async function replayMessage(messageId) {
    try {
        const response = await fetch(`/api/replay?id=${messageId}`, {
            method: 'POST'
        });
        
        const result = await response.json();
        
        if (response.ok) {
            showToast('Message replayed successfully!', 'success');
            loadMessages();
        } else {
            showToast('Failed to replay message', 'error');
        }
    } catch (error) {
        console.error('Error replaying message:', error);
        showToast('Error replaying message', 'error');
    }
}

function replayCurrentMessage() {
    if (currentMessageId) {
        replayMessage(currentMessageId);
        closeModal('messageModal');
    }
}

// Export Functions
async function exportJSON() {
    window.location.href = '/api/messages/export/json';
    showToast('Exporting messages as JSON...', 'info');
}

async function exportCSV() {
    window.location.href = '/api/messages/export/csv';
    showToast('Exporting messages as CSV...', 'info');
}

// Clear Messages
function clearMessages() {
    if (confirm('Are you sure you want to clear all messages from the display? This will not delete messages from Pub/Sub.')) {
        messages = [];
        renderMessages(messages);
        updateMessageBadge(0);
        showToast('Messages cleared from display', 'info');
    }
}

// Utility Functions
function updateMessageBadge(count) {
    document.getElementById('messageBadge').textContent = count;
}

function formatTime(date) {
    return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

function formatTimeAgo(date) {
    const seconds = Math.floor((new Date() - date) / 1000);
    
    if (seconds < 60) return 'Just now';
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icons = {
        success: '✅',
        error: '❌',
        info: 'ℹ️'
    };
    
    toast.innerHTML = `
        <span class="toast-icon">${icons[type] || icons.info}</span>
        <span class="toast-message">${message}</span>
    `;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideInRight 0.3s reverse';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// Close modal when clicking outside
window.onclick = function(event) {
    if (event.target.classList.contains('modal')) {
        event.target.classList.remove('show');
    }
};
