<script>
    import { onMount } from "svelte";
    import { API_BASE } from "../api.js";
    import { user } from "../stores.js";
    import Header from "../components/Header.svelte";
    import "../styles/dashboard.css";
    import "../styles/inbox.css";

    let messages = [];
    let loading = true;
    let error = "";

    // UI State
    let selectedMessage = null;
    let showCompose = false;

    // Compose State
    let recipients = []; // list of employees to choose from
    let composeForm = {
        receiver_id: "",
        subject: "",
        body: "",
    };
    let sending = false;

    onMount(async () => {
        await loadMessages();
        loadRecipients();
    });

    async function loadMessages() {
        loading = true;
        try {
            const res = await fetch(`${API_BASE}/api/inbox`, {
                credentials: "include",
            });
            if (res.ok) {
                const data = await res.json();
                messages = data || [];
            }
        } catch (e) {
            console.error(e);
            error = "Failed to load messages";
        } finally {
            loading = false;
        }
    }

    async function loadRecipients() {
        try {
            // Using existing employees endpoint
            const res = await fetch(`${API_BASE}/api/employees`, {
                credentials: "include",
            });
            if (res.ok) {
                const data = await res.json();
                recipients = data.map((e) => ({
                    id: e.id,
                    name: e.name,
                    email: e.email,
                }));
            }
        } catch (e) {
            console.error("Failed to load recipients", e);
        }
    }

    function selectMessage(msg) {
        selectedMessage = msg;
        if (!msg.is_read) {
            markAsRead(msg.id);
            // Optimistic update
            messages = messages.map((m) =>
                m.id === msg.id ? { ...m, is_read: true } : m,
            );
        }
    }

    async function markAsRead(id) {
        try {
            await fetch(`${API_BASE}/api/inbox/${id}/read`, {
                method: "PUT",
                credentials: "include",
            });
        } catch (e) {
            console.error(e);
        }
    }

    function closeDetail() {
        selectedMessage = null;
    }

    function openCompose() {
        composeForm = { receiver_id: "", subject: "", body: "" };
        showCompose = true;
    }

    function closeCompose() {
        showCompose = false;
    }

    async function sendArgs() {
        if (
            !composeForm.receiver_id ||
            !composeForm.subject ||
            !composeForm.body
        )
            return;

        sending = true;
        try {
            const res = await fetch(`${API_BASE}/api/inbox`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                credentials: "include",
                body: JSON.stringify({
                    receiver_id: parseInt(composeForm.receiver_id),
                    subject: composeForm.subject,
                    body: composeForm.body,
                }),
            });

            if (res.ok) {
                closeCompose();
                alert("Message sent!");
                // No need to reload inbox as sent messages don't appear in inbox usually
                // But if we had "Sent" folder, we would reload.
            } else {
                alert("Failed to send message");
            }
        } catch (e) {
            console.error(e);
            alert("Error sending message");
        } finally {
            sending = false;
        }
    }

    function formatDate(dateStr) {
        return new Date(dateStr).toLocaleString();
    }

    // Delete Confirmation State
    let showDeleteModal = false;
    let messageToDeleteId = null;

    function promptDelete(id) {
        messageToDeleteId = id;
        showDeleteModal = true;
    }

    function cancelDelete() {
        showDeleteModal = false;
        messageToDeleteId = null;
    }

    async function confirmDelete() {
        if (!messageToDeleteId) return;

        try {
            const res = await fetch(
                `${API_BASE}/api/inbox/${messageToDeleteId}`,
                {
                    method: "DELETE",
                    credentials: "include",
                },
            );

            if (res.ok) {
                // remove from list
                messages = messages.filter((m) => m.id !== messageToDeleteId);
                selectedMessage = null; // deselect
                showDeleteModal = false; // close modal
                messageToDeleteId = null;
            } else {
                console.error("Failed to delete message");
                // Optional: show a toast or error state here
            }
        } catch (e) {
            console.error(e);
        }
    }
</script>

<div class="page-wrapper">
    <Header>
        <div>
            <h1>Inbox</h1>
            <p>Internal messages and notifications.</p>
        </div>
    </Header>

    <div class="page-content inbox-layout">
        <!-- Message List -->
        <div class="message-list-pane {selectedMessage ? 'hidden-mobile' : ''}">
            <div class="inbox-toolbar">
                <button class="btn-primary full-width" on:click={openCompose}>
                    + New Message
                </button>
            </div>

            {#if loading}
                <div class="p-4 text-center muted">Loading...</div>
            {:else if messages.length === 0}
                <div class="empty-state">
                    <p>No messages yet.</p>
                </div>
            {:else}
                <ul class="message-list">
                    {#each messages as msg}
                        <li
                            class="message-item {msg.is_read
                                ? 'read'
                                : 'unread'} {selectedMessage?.id === msg.id
                                ? 'active'
                                : ''}"
                            on:click={() => selectMessage(msg)}
                            on:keydown={(e) =>
                                e.key === "Enter" && selectMessage(msg)}
                            role="button"
                            tabindex="0"
                        >
                            <div class="msg-header">
                                <span class="sender">{msg.sender_name}</span>
                                <span class="date"
                                    >{new Date(
                                        msg.created_at,
                                    ).toLocaleDateString()}</span
                                >
                            </div>
                            <div class="msg-subject">{msg.subject}</div>
                            <div class="msg-preview">
                                {msg.body.substring(0, 50)}...
                            </div>
                        </li>
                    {/each}
                </ul>
            {/if}
        </div>

        <!-- Message Detail -->
        <div
            class="message-detail-pane {selectedMessage
                ? 'visible-mobile'
                : ''}"
        >
            {#if selectedMessage}
                <div class="detail-header">
                    <div class="detail-top-row">
                        <button
                            class="back-btn mobile-only"
                            on:click={closeDetail}>&larr; Back</button
                        >
                        <button
                            class="btn-delete-msg"
                            on:click={() => promptDelete(selectedMessage.id)}
                            title="Delete Message"
                        >
                            🗑️ Delete
                        </button>
                    </div>
                    <h2>{selectedMessage.subject}</h2>
                    <div class="meta">
                        From <strong>{selectedMessage.sender_name}</strong> on {formatDate(
                            selectedMessage.created_at,
                        )}
                    </div>
                </div>
                <div class="detail-body">
                    {selectedMessage.body}
                </div>
            {:else}
                <div class="empty-detail">
                    <p>Select a message to read</p>
                </div>
            {/if}
        </div>
    </div>
</div>

{#if showCompose}
    <div class="modal-backdrop">
        <div class="modal-card">
            <div class="modal-header">
                <h2>New Message</h2>
                <button class="close-btn" on:click={closeCompose}
                    >&times;</button
                >
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label for="recipient">To</label>
                    <select id="recipient" bind:value={composeForm.receiver_id}>
                        <option value="">Select Recipient...</option>
                        {#each recipients as r}
                            <option value={r.id}>{r.name} ({r.email})</option>
                        {/each}
                    </select>
                </div>
                <div class="form-group">
                    <label for="subject">Subject</label>
                    <input
                        id="subject"
                        type="text"
                        bind:value={composeForm.subject}
                        placeholder="Subject..."
                    />
                </div>
                <div class="form-group">
                    <label for="message">Message</label>
                    <textarea
                        id="message"
                        rows="5"
                        bind:value={composeForm.body}
                        placeholder="Write your message..."
                    ></textarea>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-secondary" on:click={closeCompose}
                    >Cancel</button
                >
                <button
                    class="btn-primary"
                    on:click={sendArgs}
                    disabled={sending}
                >
                    {sending ? "Sending..." : "Send Message"}
                </button>
            </div>
        </div>
    </div>
{/if}

{#if showDeleteModal}
    <div class="modal-backdrop">
        <div class="modal-card delete-modal">
            <div class="modal-header">
                <h2>Confirm Delete</h2>
                <button class="close-btn" on:click={cancelDelete}
                    >&times;</button
                >
            </div>
            <div class="modal-body">
                <p class="delete-confirm-text">
                    Are you sure you want to permanently delete this message?
                </p>
            </div>
            <div class="modal-footer">
                <button class="btn-secondary" on:click={cancelDelete}
                    >Cancel</button
                >
                <button class="btn-delete-confirm" on:click={confirmDelete}
                    >Yes, Delete</button
                >
            </div>
        </div>
    </div>
{/if}
