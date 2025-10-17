// Dashboard JavaScript
'use strict';

// State management
const state = {
    messages: [],
    currentMessageId: null,
    topics: [],
    subscriptions: [],
    isLoading: false,
    lastUpdate: null,
    theme: localStorage.getItem('theme') || 'light'
};

// Theme Management
function toggleTheme() {
    const html = document.documentElement;
    const currentTheme = html.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

    html.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    state.theme = newTheme;

    // Update theme icon with animation
    const themeIcon = document.querySelector('.theme-icon');
    themeIcon.style.transform = 'rotate(360deg)';
    setTimeout(() => {
        themeIcon.textContent = newTheme === 'dark' ? '☀️' : '🌙';
        themeIcon.style.transform = 'rotate(0deg)';
    }, 150);
}

// Initialize theme on load
function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'light';
    document.documentElement.setAttribute('data-theme', savedTheme);
    state.theme = savedTheme;

    // Set correct icon
    const themeIcon = document.querySelector('.theme-icon');
    if (themeIcon) {
        themeIcon.textContent = savedTheme === 'dark' ? '☀️' : '🌙';
    }
}

// Performance optimization: Request Animation Frame for smooth UI updates
const rafScheduler = (callback) => {
    requestAnimationFrame(callback);
};

// Initialize dashboard on page load
document.addEventListener('DOMContentLoaded', function() {
    console.log('🚀 Initializing Pub/Sub Dashboard...');

    // Initialize theme first
    initTheme();

    // Initial load
    loadStats();
    loadMessages();
    setupSearchHandlers();

    // Update connection status
    updateConnectionStatus(true);

    // Refresh stats every 5 seconds
    setInterval(() => rafScheduler(loadStats), 5000);

    // Refresh messages every 10 seconds
    setInterval(() => rafScheduler(loadMessages), 10000);

    // Performance monitoring
    if ('performance' in window && performance.getEntriesByType) {
        window.addEventListener('load', () => {
            const perfEntries = performance.getEntriesByType('navigation');
            if (perfEntries.length > 0) {
                const pageLoadTime = perfEntries[0].loadEventEnd - perfEntries[0].startTime;
                console.log(`⚡ Page loaded in ${Math.round(pageLoadTime)}ms`);
            }
        });
    }
});

// Connection status management
function updateConnectionStatus(connected) {
    const indicator = document.getElementById('wsStatus');
    const text = document.getElementById('wsStatusText');

    if (connected) {
        indicator.classList.add('connected');
        text.textContent = 'Connected';
    } else {
        indicator.classList.remove('connected');
        text.textContent = 'Disconnected';
    }
}

// Load Statistics with error handling
async function loadStats() {
    try {
        const response = await fetch('/api/stats', {
            method: 'GET',
            headers: { 'Accept': 'application/json' }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const stats = await response.json();
        updateStats(stats);
        updateConnectionStatus(true);
    } catch (error) {
        console.error('❌ Error loading stats:', error);
        updateConnectionStatus(false);
    }
}

function updateStats(stats) {
    // Use RAF for smooth updates
    rafScheduler(() => {
        const elements = {
            topicCount: stats.topics || 0,
            subscriptionCount: stats.subscriptions || 0,
            messageCount: stats.total_messages || 0
        };

        Object.entries(elements).forEach(([id, value]) => {
            const el = document.getElementById(id);
            if (el && el.textContent !== String(value)) {
                animateNumber(el, parseInt(el.textContent) || 0, value);
            }
        });

        if (stats.last_message_time) {
            const time = new Date(stats.last_message_time);
            document.getElementById('lastMessage').textContent = formatTimeAgo(time);
        }

        // Update topics list
        if (stats.topic_list) {
            state.topics = stats.topic_list;
            updateTopicSelects();
            updateTopicFilter();
        }

        if (stats.subscription_list) {
            state.subscriptions = stats.subscription_list;
        }

        state.lastUpdate = new Date();
    });
}

// Animate number changes for better UX
function animateNumber(element, start, end, duration = 500) {
    if (start === end) return;

    const startTime = performance.now();
    const difference = end - start;

    function update(currentTime) {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);

        const easeOutQuad = progress * (2 - progress);
        const current = Math.round(start + (difference * easeOutQuad));

        element.textContent = current;

        if (progress < 1) {
            requestAnimationFrame(update);
        }
    }

    requestAnimationFrame(update);
}

// Load Messages with loading state
async function loadMessages() {
    if (state.isLoading) return; // Prevent concurrent loads

    state.isLoading = true;

    try {
        const response = await fetch('/api/messages', {
            method: 'GET',
            headers: { 'Accept': 'application/json' }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const newMessages = await response.json();

        // Only update if messages changed (performance optimization)
        if (JSON.stringify(newMessages) !== JSON.stringify(state.messages)) {
            state.messages = newMessages;
            renderMessages(state.messages);
            updateMessageBadge(state.messages.length);
            console.log(`📬 Loaded ${state.messages.length} messages`);
        }
    } catch (error) {
        console.error('❌ Error loading messages:', error);
        showToast('Failed to load messages', 'error');
    } finally {
        state.isLoading = false;
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

    let filtered = state.messages;

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
    state.topics.forEach(topic => {
        const option = document.createElement('option');
        option.value = topic;
        option.textContent = topic;
        select.appendChild(option);
    });

    select.value = currentValue;
}

function updateTopicSelects() {
    updateSelect('publishTopic', state.topics);
    updateSelect('subscriptionTopic', state.topics);
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
    if (state.topics.length === 0) {
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
    if (state.topics.length === 0) {
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
    const msg = state.messages.find(m => m.id === messageId);
    if (!msg) return;

    state.currentMessageId = messageId;
    
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
    if (state.currentMessageId) {
        replayMessage(state.currentMessageId);
        closeModal('messageModal');
    }
}

// Clear Messages
function clearMessages() {
    if (confirm('Are you sure you want to clear all messages from the display? This will not delete messages from Pub/Sub.')) {
        state.messages = [];
        renderMessages(state.messages);
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
    // Remove existing toasts
    const existingToasts = document.querySelectorAll('.toast');
    existingToasts.forEach(t => t.remove());

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

    // Auto-dismiss after 4 seconds with smooth animation
    setTimeout(() => {
        toast.style.animation = 'slideInRight 0.4s reverse ease-in';
        setTimeout(() => toast.remove(), 400);
    }, 4000);

    // Allow click to dismiss
    toast.addEventListener('click', () => {
        toast.style.animation = 'slideInRight 0.3s reverse ease-in';
        setTimeout(() => toast.remove(), 300);
    });

    // Add hover effect to pause auto-dismiss
    let dismissTimeout;
    toast.addEventListener('mouseenter', () => {
        clearTimeout(dismissTimeout);
    });
}

// Close modal when clicking outside
window.onclick = function(event) {
    if (event.target.classList.contains('modal')) {
        event.target.classList.remove('show');
    }
};
