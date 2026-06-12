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
    setupMessageActions();
    setupModalKeyboard();

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

let messagesLoaded = false;

// Load Messages with loading state
async function loadMessages() {
    if (state.isLoading) return; // Prevent concurrent loads

    state.isLoading = true;

    // Show a spinner only before the first load completes, so polling never
    // flashes it.
    if (!messagesLoaded) showMessagesLoading();

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
        messagesLoaded = true;
    }
}

// showMessagesLoading renders a centered spinner inside the message list.
function showMessagesLoading() {
    const container = document.getElementById('messagesContainer');
    if (!container) return;
    container.innerHTML = `
        <div class="loading-state" role="status">
            <div class="loading-spinner" aria-hidden="true"></div>
            <p>Loading messages…</p>
        </div>
    `;
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
    const escapedId = escapeHtml(msg.id);

    return `
        <div class="message-card" data-message-id="${escapedId}">
            <div class="message-header">
                <span class="message-id">ID: ${escapedId}</span>
                <span class="message-topic">${escapeHtml(msg.topic)}</span>
            </div>
            <div class="message-data">${escapeHtml(formatPayload(msg.data))}</div>
            <div class="message-footer">
                <div class="message-time">
                    <span>📤 Published: ${formatTime(publishTime)}</span>
                    <span>📥 Received: ${formatTime(receivedTime)}</span>
                </div>
                <div class="message-actions">
                    <button class="btn btn-info" data-action="view">View</button>
                    <button class="btn btn-secondary" data-action="copy">📋 Copy</button>
                    <button class="btn btn-primary" data-action="replay">🔄 Replay</button>
                </div>
            </div>
        </div>
    `;
}

// setupMessageActions wires a single delegated click handler for the message
// list. IDs are read from the card's data attribute rather than interpolated
// into inline handlers, removing any script-injection surface.
function setupMessageActions() {
    const container = document.getElementById('messagesContainer');
    if (!container) return;

    container.addEventListener('click', (e) => {
        const button = e.target.closest('button[data-action]');
        if (!button) return;

        const card = button.closest('.message-card');
        if (!card) return;

        const id = card.dataset.messageId;
        switch (button.dataset.action) {
            case 'view':
                showMessageDetails(id);
                break;
            case 'replay':
                replayMessage(id);
                break;
            case 'copy':
                copyMessageData(id);
                break;
        }
    });
}

// copyMessageData copies the raw (unformatted) payload to the clipboard.
function copyMessageData(messageId) {
    const msg = state.messages.find(m => m.id === messageId);
    if (!msg) return;

    navigator.clipboard.writeText(msg.data)
        .then(() => showToast('Message data copied to clipboard', 'success'))
        .catch(() => showToast('Failed to copy message data', 'error'));
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

let lastFocusedElement = null;

function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;

    lastFocusedElement = document.activeElement;
    modal.classList.add('show');

    // Move focus into the dialog, preferring the first form field over the
    // close button, for keyboard and screen-reader users.
    const focusable = modal.querySelector('input, textarea, select') || modal.querySelector('button');
    if (focusable) focusable.focus();
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return;

    modal.classList.remove('show');

    // Return focus to whatever opened the dialog.
    if (lastFocusedElement && typeof lastFocusedElement.focus === 'function') {
        lastFocusedElement.focus();
        lastFocusedElement = null;
    }
}

// getOpenModal returns the currently visible modal element, if any.
function getOpenModal() {
    return document.querySelector('.modal.show');
}

// setupModalKeyboard adds Escape-to-close and a basic Tab focus trap for the
// open modal.
function setupModalKeyboard() {
    document.addEventListener('keydown', (e) => {
        const modal = getOpenModal();
        if (!modal) return;

        if (e.key === 'Escape') {
            closeModal(modal.id);
            return;
        }

        if (e.key !== 'Tab') return;

        const focusable = modal.querySelectorAll(
            'a[href], button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled])'
        );
        if (focusable.length === 0) return;

        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (e.shiftKey && document.activeElement === first) {
            e.preventDefault();
            last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
            e.preventDefault();
            first.focus();
        }
    });
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
            <div class="detail-value">${escapeHtml(msg.id)}</div>
        </div>
        <div class="detail-row">
            <div class="detail-label">Topic</div>
            <div class="detail-value">${escapeHtml(msg.topic)}</div>
        </div>
        <div class="detail-row">
            <div class="detail-label">Data</div>
            <div class="detail-value">${escapeHtml(formatPayload(msg.data))}</div>
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
        const response = await fetch(`/api/replay?id=${encodeURIComponent(messageId)}`, {
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
    const badge = document.getElementById('messageBadge');
    if (!badge) return;

    const changed = badge.textContent !== String(count);
    badge.textContent = count;

    // Pulse once when the count actually changes.
    if (changed) {
        badge.classList.remove('pulse');
        void badge.offsetWidth; // restart the animation
        badge.classList.add('pulse');
    }
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

// escapeHtml encodes &, <, >, " and ' so a value is safe in both element-text
// and quoted-attribute contexts.
function escapeHtml(text) {
    return String(text ?? '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

// formatPayload pretty-prints JSON payloads; non-JSON data is returned as-is.
function formatPayload(data) {
    try {
        return JSON.stringify(JSON.parse(data), null, 2);
    } catch {
        return data;
    }
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

// getToastContainer returns the shared (screen-reader announced) toast stack,
// creating it on first use.
function getToastContainer() {
    let container = document.getElementById('toastContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toastContainer';
        container.className = 'toast-container';
        container.setAttribute('role', 'status');
        container.setAttribute('aria-live', 'polite');
        document.body.appendChild(container);
    }
    return container;
}

function showToast(message, type = 'info') {
    const container = getToastContainer();

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;

    const icons = {
        success: '✅',
        error: '❌',
        info: 'ℹ️'
    };

    const icon = document.createElement('span');
    icon.className = 'toast-icon';
    icon.setAttribute('aria-hidden', 'true');
    icon.textContent = icons[type] || icons.info;

    // textContent keeps server-supplied error text (topic names, etc.) inert.
    const msg = document.createElement('span');
    msg.className = 'toast-message';
    msg.textContent = message;

    toast.append(icon, msg);
    container.appendChild(toast);

    const dismiss = () => {
        toast.style.animation = 'slideIn 0.3s reverse ease-in';
        setTimeout(() => toast.remove(), 300);
    };

    let timer = setTimeout(dismiss, 4000);
    toast.addEventListener('click', dismiss);
    toast.addEventListener('mouseenter', () => clearTimeout(timer));
    toast.addEventListener('mouseleave', () => { timer = setTimeout(dismiss, 2000); });
}

// Close modal when clicking outside
window.onclick = function(event) {
    if (event.target.classList.contains('modal')) {
        closeModal(event.target.id);
    }
};
