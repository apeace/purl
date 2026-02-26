<template>
  <div v-if="ticket" class="ticket-detail">
    <!-- Thread header -->
    <div class="thread-header">
      <div class="thread-meta">
        <div class="thread-customer-row">
          <div class="thread-avatar" :style="{ background: ticket.avatarColor }">
            {{ ticket.name[0] }}
          </div>
          <div>
            <div class="thread-name">{{ ticket.name }}
              <span class="thread-company">· {{ ticket.company }}</span>
            </div>
            <div class="thread-id">{{ ticket.ticketId }}</div>
          </div>
        </div>
        <div class="thread-subject">{{ ticket.subject }}</div>
      </div>
      <div class="thread-badges">
        <span class="badge" :class="`badge--${ticket.status}`">{{ ticket.status }}</span>
      </div>
    </div>

    <!-- Tab bar -->
    <div class="tab-bar">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="tab-btn"
        :class="{ 'tab-btn--active': activeTab === tab.id }"
        :title="tab.label"
        @click="activeTab = tab.id"
      >
        <component :is="tab.icon" :size="16" />
      </button>
    </div>

    <!-- Zendesk link -->
    <a
      v-if="zendeskUrl"
      :href="zendeskUrl"
      target="_blank"
      rel="noopener"
      class="zendesk-link"
    >
      <ExternalLink :size="12" />
      <span>View in Zendesk</span>
    </a>

    <!-- Tab: Communications -->
    <template v-if="activeTab === 'comms'">
      <div ref="messagesEl" class="thread-messages">
        <TransitionGroup name="msg">
          <div
            v-for="msg in ticket.messages"
            :key="msg.id"
            class="message"
            :class="{
              'message--agent': msg.from === 'agent',
              'message--customer': msg.from === 'customer',
              'message--system': msg.from === 'system',
              'message--internal': msg.channel === 'internal'
            }"
          >
            <div v-if="msg.from === 'customer'" class="msg-avatar" :style="{ background: ticket.avatarColor }">
              {{ ticket.name[0] }}
            </div>
            <div v-if="msg.type !== 'recording'" class="msg-bubble" :class="{ 'msg-bubble--system': msg.messageType === 'merge_notice' }">
              <div class="msg-header">
                <span class="msg-sender">{{ msg.authorName || (msg.from === 'agent' ? 'You' : ticket.name) }}</span>
                <span v-if="msg.automated" class="msg-automated">via automation</span>
                <span class="msg-channel" :class="`msg-channel--${commChannelCategory(msg)}`">
                  <Lock v-if="msg.commChannel === 'internal_note'" :size="10" />
                  <MessageCircle v-else-if="msg.commChannel === 'web_chat'" :size="10" />
                  <Mail v-else-if="msg.commChannel?.startsWith('email')" :size="10" />
                  <Send v-else-if="msg.commChannel === 'public_reply'" :size="10" />
                  <Globe v-else-if="msg.commChannel === 'web_form'" :size="10" />
                  <MessageSquare v-else-if="msg.commChannel === 'sms_inbound'" :size="10" />
                  <Phone v-else-if="msg.commChannel?.startsWith('call') || msg.commChannel === 'voicemail'" :size="10" />
                  <Globe v-else-if="msg.commChannel === 'ticket_merge'" :size="10" />
                  <Globe v-else :size="10" />
                  {{ commChannelLabel(msg) }}
                </span>
                <span class="msg-time">{{ msg.time }}</span>
              </div>

              <!-- Call record card -->
              <div v-if="msg.call" class="call-card">
                <div class="call-card-header">
                  <Phone :size="14" />
                  <span>{{ msg.messageType === 'call_summary' ? 'Call Log' : msg.call.direction === 'outbound' ? 'Outbound Call' : 'Inbound Call' }}</span>
                </div>
                <div class="call-card-fields">
                  <div v-if="msg.call.customerPhone" class="call-card-field">
                    <span class="call-card-label">Phone</span>
                    <span>{{ msg.call.customerPhone }}</span>
                  </div>
                  <div v-if="!msg.call.customerPhone && msg.call.callFrom" class="call-card-field">
                    <span class="call-card-label">From</span>
                    <span>{{ msg.call.callFrom }}</span>
                  </div>
                  <div v-if="!msg.call.customerPhone && msg.call.callTo" class="call-card-field">
                    <span class="call-card-label">To</span>
                    <span>{{ msg.call.callTo }}</span>
                  </div>
                  <div v-if="msg.call.agentName" class="call-card-field">
                    <span class="call-card-label">{{ msg.call.direction === 'outbound' ? 'Called by' : 'Answered by' }}</span>
                    <span>{{ msg.call.agentName }}</span>
                  </div>
                  <div v-if="msg.call.duration" class="call-card-field">
                    <span class="call-card-label">Duration</span>
                    <span>{{ msg.call.duration }}</span>
                  </div>
                  <div v-if="msg.call.timeOfCall" class="call-card-field">
                    <span class="call-card-label">Time</span>
                    <span>{{ msg.call.timeOfCall }}</span>
                  </div>
                </div>
                <div v-if="msg.hasRecording" class="call-audio-player">
                  <button class="call-audio-play-btn" @click="toggleAudioPlayback(msg)">
                    <Pause v-if="audioIsPlaying(msg)" :size="14" />
                    <Play v-else :size="14" />
                  </button>
                  <div class="call-audio-track" @click="seekAudio($event)">
                    <div class="call-audio-progress" :style="{ width: `${audioProgressFor(msg) * 100}%` }" />
                  </div>
                  <span class="call-audio-time">{{ audioTimeDisplay(msg) }} / {{ audioDurationDisplay(msg) }}</span>
                </div>
                <a v-else-if="msg.call.recordingUrl" :href="msg.call.recordingUrl" target="_blank" class="call-card-recording">
                  <Play :size="12" /> Listen to recording
                </a>
                <button v-if="msg.transcript" class="transcript-toggle" @click="toggleTranscript(msg.id)">
                  <Sparkles :size="13" />
                  <span>Call Transcript</span>
                  <ChevronDown :size="14" :class="{ 'chevron-flipped': expandedTranscriptId === msg.id }" />
                </button>
                <div v-if="msg.transcript && expandedTranscriptId === msg.id" class="call-card-transcript">
                  {{ msg.transcript }}
                </div>
              </div>

              <!-- Voicemail card -->
              <div v-else-if="msg.voicemail" class="call-card call-card--voicemail">
                <div class="call-card-header">
                  <Phone :size="14" />
                  <span>Voicemail</span>
                </div>
                <div class="call-card-fields">
                  <div v-if="msg.voicemail.customerPhone" class="call-card-field">
                    <span class="call-card-label">From</span>
                    <span>{{ msg.voicemail.customerPhone }}</span>
                  </div>
                  <div v-if="msg.voicemail.duration" class="call-card-field">
                    <span class="call-card-label">Duration</span>
                    <span>{{ msg.voicemail.duration }}</span>
                  </div>
                  <div v-if="msg.voicemail.location" class="call-card-field">
                    <span class="call-card-label">Location</span>
                    <span>{{ msg.voicemail.location }}</span>
                  </div>
                </div>
                <div v-if="msg.hasRecording" class="call-audio-player call-audio-player--voicemail">
                  <button class="call-audio-play-btn call-audio-play-btn--voicemail" @click="toggleAudioPlayback(msg)">
                    <Pause v-if="audioIsPlaying(msg)" :size="14" />
                    <Play v-else :size="14" />
                  </button>
                  <div class="call-audio-track call-audio-track--voicemail" @click="seekAudio($event)">
                    <div class="call-audio-progress call-audio-progress--voicemail" :style="{ width: `${audioProgressFor(msg) * 100}%` }" />
                  </div>
                  <span class="call-audio-time">{{ audioTimeDisplay(msg) }} / {{ audioDurationDisplay(msg) }}</span>
                </div>
                <a v-else-if="msg.voicemail.recordingUrl" :href="msg.voicemail.recordingUrl" target="_blank" class="call-card-recording">
                  <Play :size="12" /> Listen to voicemail
                </a>
                <button v-if="msg.transcript" class="transcript-toggle" @click="toggleTranscript(msg.id)">
                  <Sparkles :size="13" />
                  <span>Call Transcript</span>
                  <ChevronDown :size="14" :class="{ 'chevron-flipped': expandedTranscriptId === msg.id }" />
                </button>
                <div v-if="msg.transcript && expandedTranscriptId === msg.id" class="call-card-transcript">
                  {{ msg.transcript }}
                </div>
              </div>

              <!-- Web chat conversation -->
              <div v-else-if="msg.messageType === 'web_chat'" class="msg-body msg-body--webchat">
                <div class="webchat-thread">
                  <div
                    v-for="(line, i) in parseWebChatLines(msg.text)"
                    :key="i"
                    class="webchat-line"
                    :class="line.role === 'customer' ? 'webchat-line--user' : 'webchat-line--bot'"
                  >
                    <div class="webchat-avatar" :class="`webchat-avatar--${line.role}`">
                      <Bot v-if="line.role === 'bot'" :size="11" />
                      <User v-else :size="11" />
                    </div>
                    <div class="webchat-bubble" :class="line.role === 'customer' ? 'webchat-bubble--user' : 'webchat-bubble--bot'">
                      <div class="webchat-meta">
                        <span class="webchat-speaker">{{ line.role === 'customer' ? 'Customer' : line.speaker }}</span>
                        <span class="webchat-time">{{ line.time }}</span>
                      </div>
                      <div class="webchat-text">{{ line.text }}</div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Merge notice -->
              <div v-else-if="msg.messageType === 'merge_notice'" class="msg-body msg-body--merge">
                {{ msg.text }}
              </div>

              <!-- Rich HTML body (emails, internal notes — preserves tables, lists, formatting) -->
              <div v-else-if="msg.htmlBody" class="msg-body" :class="{ 'msg-body--email': isEmailMessage(msg), 'msg-body--note': msg.commChannel === 'internal_note' }">
                <div class="email-content email-content--html" v-html="msg.htmlBody" />
              </div>

              <!-- Email message body (plain text) -->
              <div v-else-if="isEmailMessage(msg)" class="msg-body msg-body--email">
                <div class="email-content">{{ splitEmailBody(msg.text).main }}</div>
                <div v-if="splitEmailBody(msg.text).quoted" class="email-quoted">
                  <button class="email-quoted-toggle" @click="toggleQuoted(msg.id)">
                    <ChevronDown :size="12" :class="{ 'chevron-flipped': expandedQuotedId === msg.id }" />
                    <span>{{ expandedQuotedId === msg.id ? 'Hide' : 'Show' }} quoted text</span>
                  </button>
                  <div v-if="expandedQuotedId === msg.id" class="email-quoted-text">{{ splitEmailBody(msg.text).quoted }}</div>
                </div>
              </div>

              <!-- Regular message body (auto-link URLs) -->
              <div v-else class="msg-body" :class="{ 'msg-body--note': msg.commChannel === 'internal_note' }" v-html="autoLinkUrls(msg.text)" />
            </div>
            <div v-else class="recording-card">
              <div class="recording-header">
                <div class="recording-icon">
                  <Phone :size="16" />
                </div>
                <div class="recording-title">Call Recording</div>
                <span class="recording-duration">{{ formatDuration(msg.recording!.duration) }}</span>
                <span class="msg-channel msg-channel--phone">
                  <Phone :size="10" />
                  phone
                </span>
                <span class="msg-time">{{ msg.time }}</span>
              </div>
              <div class="waveform-player">
                <button class="waveform-play-btn" @click="togglePlayback(msg)">
                  <Pause v-if="playingRecordingId === msg.id" :size="16" />
                  <Play v-else :size="16" />
                </button>
                <div class="waveform-bars">
                  <div
                    v-for="(bar, i) in msg.recording!.waveform"
                    :key="i"
                    class="waveform-bar"
                    :class="{ 'waveform-bar--played': playingRecordingId === msg.id && i / msg.recording!.waveform.length <= playbackProgress }"
                    :style="{ height: `${bar * 100}%` }"
                  />
                </div>
                <div class="waveform-time">
                  <span>{{ playbackElapsedFor(msg) }}</span>
                  <span>{{ formatDuration(msg.recording!.duration) }}</span>
                </div>
              </div>
              <button class="transcript-toggle" @click="toggleTranscript(msg.id)">
                <Sparkles :size="13" />
                <span>AI Transcript</span>
                <ChevronDown :size="14" :class="{ 'chevron-flipped': expandedTranscriptId === msg.id }" />
              </button>
              <div v-if="expandedTranscriptId === msg.id" class="transcript-body">
                <div
                  v-for="(line, i) in msg.recording!.transcript"
                  :key="i"
                  class="transcript-line"
                >
                  <span class="transcript-time">{{ line.time }}</span>
                  <span class="transcript-speaker">{{ line.speaker }}</span>
                  <span class="transcript-text">{{ line.text }}</span>
                </div>
              </div>
            </div>
            <div v-if="msg.from === 'agent'" class="msg-avatar msg-avatar--agent" :style="{ background: agentAvatarColor(msg) }">{{ agentInitial(msg) }}</div>
          </div>
        </TransitionGroup>
      </div>

    </template>

    <!-- Tab: Contact Info -->
    <div v-else-if="activeTab === 'contact'" class="tab-panel">
      <div class="contact-card">
        <div class="contact-avatar" :style="{ background: ticket.avatarColor }">
          {{ ticket.name[0] }}
        </div>
        <div class="contact-name">{{ ticket.name }}</div>
        <div class="contact-company">{{ ticket.company }}</div>
      </div>

      <div class="detail-grid">
        <div class="detail-row">
          <Mail :size="14" class="detail-icon" />
          <div class="detail-content">
            <div class="detail-label">Email</div>
            <div class="detail-value">{{ ticket.email }}</div>
          </div>
        </div>
        <div class="detail-row">
          <Phone :size="14" class="detail-icon" />
          <div class="detail-content">
            <div class="detail-label">Phone</div>
            <div class="detail-value">{{ ticket.phone }}</div>
          </div>
        </div>
        <div class="detail-row">
          <Zap :size="14" class="detail-icon" />
          <div class="detail-content">
            <div class="detail-label">Subscription</div>
            <div class="detail-value">
              {{ ticket.subscription.plan }}
              <span class="sub-status" :class="`sub-status--${ticket.subscription.status}`">{{ ticket.subscription.status }}</span>
            </div>
            <div class="detail-sub">{{ ticket.subscription.id }}</div>
          </div>
        </div>
      </div>

      <!-- Internal notes -->
      <div class="notes-section">
        <div class="notes-label">Internal Notes</div>
        <textarea
          class="notes-input"
          placeholder="Add internal notes about this subscriber…"
          rows="4"
          :value="ticket.notes"
          @input="updateNotes(ticketId, ($event.target as HTMLTextAreaElement).value)"
        />
      </div>
    </div>

    <!-- Tab: Ticket History -->
    <div v-else-if="activeTab === 'history'" class="tab-panel">
      <div class="panel-title">Activity on {{ ticket.ticketId }}</div>
      <div class="timeline">
        <div
          v-for="(entry, i) in ticket.ticketHistory"
          :key="i"
          class="timeline-item"
        >
          <div class="timeline-dot" />
          <div class="timeline-content">
            <div class="timeline-event">{{ entry.event }}</div>
            <div class="timeline-time">{{ entry.time }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab: Subscriber History -->
    <div v-else-if="activeTab === 'subscriber'" class="tab-panel">
      <div class="panel-title">{{ ticket.name }}'s ticket history</div>
      <div class="panel-subtitle">Customer since {{ ticket.subscription.id }}</div>

      <!-- Current ticket -->
      <div class="history-card history-card--current">
        <div class="history-card-top">
          <span class="history-tid">{{ ticket.ticketId }}</span>
          <span class="history-status" :class="`history-status--${ticket.status}`">{{ ticket.status }}</span>
        </div>
        <div class="history-subject">{{ ticket.subject }}</div>
        <div class="history-date">Current</div>
      </div>

      <!-- Past tickets -->
      <div
        v-for="(past, i) in ticket.subscriberHistory"
        :key="i"
        class="history-card"
      >
        <div class="history-card-top">
          <span class="history-tid">{{ past.ticketId }}</span>
          <span class="history-status" :class="`history-status--${past.status}`">{{ past.status }}</span>
        </div>
        <div class="history-subject">{{ past.subject }}</div>
        <div class="history-date">{{ past.date }}</div>
      </div>

      <div v-if="!ticket.subscriberHistory.length" class="panel-empty">
        No previous tickets on record.
      </div>
    </div>

    <!-- Tab: Settings -->
    <div v-else-if="activeTab === 'settings'" class="tab-panel">
      <!-- Tags -->
      <div class="settings-section">
        <div class="settings-label">Tags</div>
        <div class="tags-wrap">
          <span
            v-for="tag in ticket.tags"
            :key="tag"
            class="tag"
          >
            {{ tag }}
            <button class="tag-remove" @click="removeTag(ticketId, tag)">&times;</button>
          </span>
          <div class="tag-add">
            <input
              v-model="newTag"
              class="tag-input"
              placeholder="Add tag…"
              @keydown.enter="handleAddTag"
            />
          </div>
        </div>
      </div>

      <!-- Temperature -->
      <div class="settings-section">
        <div class="settings-label">Temperature</div>
        <div class="option-row">
          <button
            v-for="opt in tempOptions"
            :key="opt"
            class="option-btn"
            :class="{ 'option-btn--active': ticket.temperature === opt, [`option-btn--${opt}`]: true }"
            @click="setTemperature(ticketId, opt)"
          >{{ opt }}</button>
        </div>
      </div>

      <!-- Assignee -->
      <div class="settings-section">
        <div class="settings-label">Assignee</div>
        <select
          class="settings-select"
          :value="ticket.assignee"
          @change="setAssignee(ticketId, ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="a in assigneeOptions" :key="a" :value="a">{{ a }}</option>
        </select>
      </div>

      <!-- Status -->
      <div class="settings-section">
        <div class="settings-label">Status</div>
        <div class="option-row option-row--wrap">
          <button
            v-for="s in statusOptions"
            :key="s"
            class="option-btn"
            :class="{ 'option-btn--active': ticket.status === s }"
            @click="setStatus(ticketId, s)"
          >{{ s.replace(/_/g, " ") }}</button>
        </div>
      </div>
    </div>

    <!-- Tab: Actions -->
    <div v-else-if="activeTab === 'actions'" class="tab-panel">
      <div class="panel-title">Quick Actions</div>
      <div class="actions-grid">
        <button class="action-card" @click="handleAction('truck')">
          <Truck :size="20" class="action-icon action-icon--blue" />
          <span class="action-label">Roll a Truck</span>
          <span class="action-desc">Dispatch a field technician</span>
        </button>
        <button class="action-card" @click="handleAction('escalate')">
          <AlertTriangle :size="20" class="action-icon action-icon--orange" />
          <span class="action-label">Escalate</span>
          <span class="action-desc">Raise priority to Tier 2</span>
        </button>
        <button class="action-card" @click="handleAction('reboot')">
          <RotateCcw :size="20" class="action-icon action-icon--green" />
          <span class="action-label">Reboot ONT</span>
          <span class="action-desc">Remote restart optical terminal</span>
        </button>
        <button class="action-card" @click="handleAction('credit')">
          <DollarSign :size="20" class="action-icon action-icon--purple" />
          <span class="action-label">Apply Credit</span>
          <span class="action-desc">Issue account credit</span>
        </button>
        <button v-if="showAddToBoard" class="action-card" @click="emit('addToBoard', ticketId)">
          <Columns3 :size="20" class="action-icon action-icon--indigo" />
          <span class="action-label">Add to Board</span>
          <span class="action-desc">Place on a custom board</span>
        </button>
      </div>
    </div>

    <!-- Bottom dock: collapsible reply + AI panel -->
    <div class="bottom-dock">
      <div v-if="activeTab === 'comms'" class="compose-section">
        <button class="compose-toggle" @click="composeCollapsed = !composeCollapsed">
          <ChevronDown :size="14" :class="{ 'chevron-flipped': !composeCollapsed }" />
          <span>Reply</span>
        </button>
        <div v-show="!composeCollapsed" class="thread-compose">
          <div class="channel-bar">
            <button
              v-for="ch in channelOptions"
              :key="ch.id"
              class="channel-btn"
              :class="{ 'channel-btn--active': activeChannel === ch.id, [`channel-btn--${ch.id}`]: true }"
              @click="replyChannel = ch.id"
            >
              <component :is="ch.icon" :size="13" />
              <span class="channel-label">{{ ch.label }}</span>
            </button>
          </div>
          <!-- Standard compose (non-phone channels) -->
          <template v-if="!isPhoneChannel">
            <textarea
              v-model="replyText"
              class="compose-input"
              placeholder="Write a reply…"
              rows="3"
              @keydown.meta.enter="sendReply"
            />
            <div class="compose-actions">
              <button class="btn btn--ghost" @click="handleResolve">Resolve</button>
              <button class="btn btn--primary" :disabled="!replyText.trim()" @click="sendReply">
                <Send :size="14" /> Send Reply
              </button>
            </div>
          </template>

          <!-- Phone call UI -->
          <div v-else class="phone-ui">
            <!-- Status card (idle or non-merged calls) -->
            <div
              v-if="!isMerged"
              class="phone-status-card"
              :class="primaryCall ? `phone-status-card--${primaryCall.status}` : ''"
            >
              <div class="phone-status-icon" :class="{ 'phone-status-icon--ringing': primaryCall?.status === 'ringing' }">
                <Pause v-if="primaryCall?.status === 'on-hold'" :size="20" />
                <PhoneCall v-else :size="20" />
              </div>
              <div class="phone-status-info">
                <div class="phone-status-label">
                  {{ !primaryCall ? "Ready to call" : primaryCall.status === "ringing" ? "Ringing…" : primaryCall.status === "connected" ? "Connected" : "On Hold" }}
                </div>
                <div class="phone-status-number">
                  {{ !primaryCall ? customerPhoneDisplay : primaryCall.number }}
                </div>
              </div>
              <div v-if="primaryCall && (primaryCall.status === 'connected' || primaryCall.status === 'on-hold')" class="phone-timer">
                <Clock :size="12" />
                {{ formattedTime(primaryCall) }}
              </div>
            </div>

            <!-- Merged conference card -->
            <div
              v-if="isMerged"
              class="phone-status-card"
              :class="calls.every(c => c.status === 'on-hold') ? 'phone-status-card--on-hold' : 'phone-status-card--connected'"
            >
              <div class="phone-status-icon">
                <Users :size="20" />
              </div>
              <div class="phone-status-info">
                <div class="phone-status-label">
                  {{ calls.every(c => c.status === "on-hold") ? "Conference — On Hold" : "Conference" }}
                </div>
                <div class="phone-status-number">{{ calls.length }} participants</div>
              </div>
            </div>

            <!-- Held call card (two calls, not merged) -->
            <div
              v-if="heldCall && !isMerged && hasTwoCalls"
              class="phone-status-card phone-status-card--on-hold phone-status-card--held"
            >
              <div class="phone-status-icon">
                <Pause :size="16" />
              </div>
              <div class="phone-status-info">
                <div class="phone-status-label">On Hold</div>
                <div class="phone-status-number">{{ heldCall.name }} · {{ heldCall.number }}</div>
              </div>
              <div class="phone-timer">
                <Clock :size="12" />
                {{ formattedTime(heldCall) }}
              </div>
            </div>

            <!-- Merged participants list -->
            <div v-if="isMerged" class="phone-participants">
              <div v-for="call in calls" :key="call.id" class="phone-participant">
                <span class="phone-participant-dot" :class="call.status === 'connected' ? 'phone-participant-dot--connected' : 'phone-participant-dot--on-hold'" />
                <span class="phone-participant-name">{{ call.name }}</span>
                <span class="phone-participant-status">{{ formattedTime(call) }}</span>
                <button class="phone-participant-remove" @click="requestDrop(call)">
                  <X :size="12" />
                </button>
              </div>
            </div>

            <!-- New call picker (inline, when single call on hold) -->
            <div v-if="showNewCall" class="call-menu call-menu--inline">
              <button class="call-menu-item" @click="startCall">
                <Phone :size="14" />
                <span class="call-menu-text">
                  <span class="call-menu-label">Subscriber number</span>
                  <span class="call-menu-number">{{ customerPhoneDisplay }}</span>
                </span>
              </button>
              <div class="call-menu-divider" />
              <button v-if="!showCustomNumber" class="call-menu-item call-menu-item--other" @click="showCustomNumber = true">
                <PhoneCall :size="14" />
                <span class="call-menu-label">Other number…</span>
              </button>
              <div v-else class="call-menu-custom">
                <input
                  ref="customNumberInput"
                  v-model="callNumber"
                  class="call-menu-input"
                  placeholder="Enter number…"
                  @keydown.enter="callNumber.trim() && startCall()"
                />
                <button
                  class="call-menu-dial"
                  :disabled="!callNumber.trim()"
                  @click="startCall"
                >
                  <PhoneCall :size="13" />
                </button>
              </div>
            </div>

            <!-- Confirmation dialog -->
            <div v-if="confirmAction" class="phone-confirm">
              <div class="phone-confirm-message">{{ confirmAction.message }}</div>
              <div class="phone-confirm-actions">
                <button class="phone-btn" @click="cancelConfirm">Cancel</button>
                <button class="phone-btn phone-btn--end" @click="executeConfirm">
                  {{ confirmAction.label }}
                </button>
              </div>
            </div>

            <!-- Controls -->
            <div class="phone-controls">
              <!-- Idle -->
              <template v-if="isIdle">
                <div class="call-menu-wrap">
                  <button class="phone-btn phone-btn--call" @click="showCallMenu = !showCallMenu">
                    <PhoneCall :size="15" /> Call {{ ticket.name.split(" ")[0] }}
                    <ChevronDown :size="13" :class="{ 'chevron-flipped': showCallMenu }" />
                  </button>
                  <div v-if="showCallMenu" class="call-menu">
                    <button class="call-menu-item" @click="startCall">
                      <Phone :size="14" />
                      <span class="call-menu-text">
                        <span class="call-menu-label">Subscriber number</span>
                        <span class="call-menu-number">{{ customerPhoneDisplay }}</span>
                      </span>
                    </button>
                    <div class="call-menu-divider" />
                    <button v-if="!showCustomNumber" class="call-menu-item call-menu-item--other" @click="showCustomNumber = true">
                      <PhoneCall :size="14" />
                      <span class="call-menu-label">Other number…</span>
                    </button>
                    <div v-else class="call-menu-custom">
                      <input
                        ref="customNumberInput"
                        v-model="callNumber"
                        class="call-menu-input"
                        placeholder="Enter number…"
                        @keydown.enter="callNumber.trim() && startCall()"
                      />
                      <button
                        class="call-menu-dial"
                        :disabled="!callNumber.trim()"
                        @click="startCall"
                      >
                        <PhoneCall :size="13" />
                      </button>
                    </div>
                  </div>
                </div>
              </template>
              <!-- Two calls (one active, one held) -->
              <template v-else-if="hasTwoCalls && !isMerged">
                <button class="phone-btn phone-btn--mute" :class="{ 'phone-btn--active': isMuted }" @click="toggleMute">
                  <MicOff v-if="isMuted" :size="15" />
                  <Mic v-else :size="15" />
                  {{ isMuted ? "Unmute" : "Mute" }}
                </button>
                <button class="phone-btn phone-btn--hold" @click="swapCalls">
                  <Phone :size="15" />
                  Swap
                </button>
                <button class="phone-btn phone-btn--call" @click="requestMerge">
                  <Users :size="15" />
                  Merge
                </button>
                <button class="phone-btn phone-btn--end" @click="requestHangUp">
                  <PhoneOff :size="15" /> End
                </button>
              </template>
              <!-- Merged conference -->
              <template v-else-if="isMerged">
                <button class="phone-btn phone-btn--mute" :class="{ 'phone-btn--active': isMuted }" @click="toggleMute">
                  <MicOff v-if="isMuted" :size="15" />
                  <Mic v-else :size="15" />
                  {{ isMuted ? "Unmute" : "Mute" }}
                </button>
                <button
                  class="phone-btn phone-btn--hold"
                  :class="{ 'phone-btn--active': calls.every(c => c.status === 'on-hold') }"
                  @click="toggleHoldMerged"
                >
                  <Play v-if="calls.every(c => c.status === 'on-hold')" :size="15" />
                  <Pause v-else :size="15" />
                  {{ calls.every(c => c.status === "on-hold") ? "Resume" : "Hold" }}
                </button>
                <button class="phone-btn phone-btn--end" @click="requestHangUp">
                  <PhoneOff :size="15" /> End
                </button>
              </template>
              <!-- Single call (ringing, connected, or on hold) -->
              <template v-else>
                <button
                  class="phone-btn phone-btn--mute"
                  :class="{ 'phone-btn--active': isMuted }"
                  :disabled="primaryCall?.status === 'ringing'"
                  @click="toggleMute"
                >
                  <MicOff v-if="isMuted" :size="15" />
                  <Mic v-else :size="15" />
                  {{ isMuted ? "Unmute" : "Mute" }}
                </button>
                <button
                  class="phone-btn phone-btn--hold"
                  :class="{ 'phone-btn--active': primaryCall?.status === 'on-hold' }"
                  :disabled="primaryCall?.status === 'ringing'"
                  @click="toggleHold"
                >
                  <Play v-if="primaryCall?.status === 'on-hold'" :size="15" />
                  <Pause v-else :size="15" />
                  {{ primaryCall?.status === "on-hold" ? "Resume" : "Hold" }}
                </button>
                <button
                  v-if="canStartSecondCall"
                  class="phone-btn phone-btn--call"
                  :class="{ 'phone-btn--active': showNewCall }"
                  @click="showNewCall = !showNewCall"
                >
                  <PhoneCall :size="15" />
                  New Call
                </button>
                <button class="phone-btn phone-btn--end" @click="requestHangUp">
                  <PhoneOff :size="15" /> End
                </button>
              </template>
            </div>
          </div>
        </div>
      </div>

      <div v-if="currentAi" class="ai-panel ai-panel--overlay">
        <ComingSoon />
        <div class="ai-panel-header">
          <div class="ai-badge">
            <Sparkles :size="11" /> AI
          </div>
          <span class="ai-panel-headline">{{ currentAi.headline }}</span>
        </div>
        <p class="ai-panel-body">{{ currentAi.body }}</p>
        <button class="btn btn--ai" @click="followAi">
          {{ currentAi.action }} <ChevronRight :size="14" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AlertTriangle, Bot, ChevronDown, ChevronRight, Clock, Cog, Columns3, DollarSign, ExternalLink, Globe, History, Lock, Mail, MessageCircle, MessageSquare, Mic, MicOff, Pause, Phone, PhoneCall, PhoneOff, Play, RotateCcw, Send, Sparkles, Truck, User, Users, X, Zap } from "lucide-vue-next"
import { storeToRefs } from "pinia"
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue"
import { useAiStore } from "../stores/useAiStore"
import { API_KEY_STORAGE_KEY } from "../utils/api"
import type { Message } from "../stores/useTicketStore"
import { avatarColor, useTicketStore } from "../stores/useTicketStore"
import ComingSoon from "./ComingSoon.vue"

const props = defineProps<{
  ticketId: string
  showAddToBoard?: boolean
}>()

const emit = defineEmits<{
  resolve: []
  addToBoard: [ticketId: string]
}>()

const ticketStore = useTicketStore()
const { tickets, zendeskSubdomain } = storeToRefs(ticketStore)
const { addTag, loadComments, removeTag, resolveTicket, sendReply: sharedSendReply, setAssignee, setStatus, setTemperature, updateNotes } = ticketStore

const aiStore = useAiStore()
const { suggestions: aiSuggestions } = storeToRefs(aiStore)

const activeTab = ref("comms")
const composeCollapsed = ref(true)
const replyText = ref("")
const replyChannel = ref<string | null>(null)
const newTag = ref("")
const messagesEl = ref<HTMLElement | null>(null)

// Phone call state
interface Call {
  id: number
  name: string
  number: string
  status: string
  seconds: number
  timerHandle: ReturnType<typeof setInterval> | null
  ringHandle: ReturnType<typeof setTimeout> | null
}

interface ConfirmAction {
  type: string
  message: string
  label: string
  callId?: number
}

const calls = ref<Call[]>([])
const isMerged = ref(false)
const isMuted = ref(false)
const showCallMenu = ref(false)
const showCustomNumber = ref(false)
const showNewCall = ref(false)
const customNumberInput = ref<HTMLInputElement | null>(null)
const callNumber = ref("")
const confirmAction = ref<ConfirmAction | null>(null)
const sessionStartTime = ref<number | null>(null)
let callIdCounter = 0

// Recording playback state (demo waveform player)
const playingRecordingId = ref<number | null>(null)
const playbackProgress = ref(0)
const playbackTimerHandle = ref<ReturnType<typeof setInterval> | null>(null)
const expandedTranscriptId = ref<number | null>(null)
const expandedQuotedId = ref<number | null>(null)

// Real audio playback state (call/voicemail recordings from Zendesk)
const audioEl = ref<HTMLAudioElement | null>(null)
const audioPlayingMsgId = ref<number | null>(null)
const audioCurrentTime = ref(0)
const audioDuration = ref(0)
const audioLoading = ref(false)

const channelOptions = [
  { id: "chat", icon: MessageCircle, label: "Chat" },
  { id: "sms", icon: MessageSquare, label: "SMS" },
  { id: "email", icon: Mail, label: "Email" },
  { id: "phone", icon: Phone, label: "Phone" },
]

const tabs = [
  { id: "comms", icon: MessageSquare, label: "Communications" },
  { id: "contact", icon: User, label: "Contact" },
  { id: "history", icon: History, label: "Ticket History" },
  { id: "subscriber", icon: Users, label: "Subscriber History" },
  { id: "actions", icon: Zap, label: "Actions" },
  { id: "settings", icon: Cog, label: "Settings" },
]

const statusOptions = ["new", "open", "pending", "escalated", "solved", "closed"]
const tempOptions = ["hot", "warm", "cool"]
const assigneeOptions = ["Alex Chen", "Sarah Kim", "Jordan Lee", "Unassigned"]

const ticket = computed(() => tickets.value.find((t) => t.id === props.ticketId))
const currentAi = computed(() => aiSuggestions.value[props.ticketId] ?? null)

const zendeskUrl = computed(() => {
  const sub = zendeskSubdomain.value
  const zdId = ticket.value?.zendeskTicketId
  if (!sub || !zdId) return null
  return `https://${sub}.zendesk.com/agent/tickets/${zdId}`
})

function recordingUrl(msg: Message): string {
  const base = import.meta.env.VITE_API_URL ?? "http://localhost:9090"
  const key = localStorage.getItem(API_KEY_STORAGE_KEY) ?? ""
  return `${base}/tickets/${props.ticketId}/comments/${msg.commentId}/recording?api_key=${encodeURIComponent(key)}`
}

function commChannelCategory(msg: Message): string {
  const cc = msg.commChannel
  if (!cc) return msg.channel
  if (cc.startsWith("email")) return "email"
  if (cc.startsWith("call") || cc === "voicemail") return "phone"
  if (cc === "sms_inbound") return "sms"
  if (cc === "internal_note") return "internal"
  if (cc === "web_chat") return "chat"
  if (cc === "web_form") return "web"
  if (cc === "public_reply") return "reply"
  if (cc === "ticket_merge") return "web"
  return msg.channel
}

const COMM_CHANNEL_LABELS: Record<string, string> = {
  email_inbound: "received",
  email_outbound: "sent",
  sms_inbound: "sms",
  call_outbound: "call",
  call_inbound: "call",
  call_summary: "call",
  voicemail: "voicemail",
  web_chat: "chat",
  web_form: "web",
  public_reply: "reply",
  internal_note: "internal",
  ticket_merge: "system",
}

function commChannelLabel(msg: Message): string {
  if (msg.commChannel) return COMM_CHANNEL_LABELS[msg.commChannel] ?? msg.channel
  return msg.channel === "voice" ? "phone" : msg.channel
}

// Email body helpers
function isEmailMessage(msg: Message): boolean {
  if (msg.automated) return true
  const cc = msg.commChannel
  return cc === "email_inbound" || cc === "email_outbound" || cc === "public_reply" || cc === "web_form"
}

function splitEmailBody(text: string): { main: string, quoted: string | null } {
  // "On ... wrote:" reply header
  const wroteMatch = text.match(/\n\s*On .+wrote:\s*\n/)
  if (wroteMatch?.index !== undefined) {
    return { main: text.slice(0, wroteMatch.index).trim(), quoted: text.slice(wroteMatch.index).trim() }
  }
  // "--- Original Message ---" separator
  const separatorMatch = text.match(/\n\s*-{3,}\s*(Original Message|Forwarded message)/i)
  if (separatorMatch?.index !== undefined) {
    return { main: text.slice(0, separatorMatch.index).trim(), quoted: text.slice(separatorMatch.index).trim() }
  }
  // Block of `>` quoted lines at the end
  const lines = text.split("\n")
  for (let i = 1; i < lines.length; i++) {
    if (lines[i].trim().startsWith(">") && lines.slice(i).every((l) => l.trim().startsWith(">") || l.trim() === "")) {
      return { main: lines.slice(0, i).join("\n").trim(), quoted: lines.slice(i).join("\n").trim() }
    }
  }
  return { main: text, quoted: null }
}

function toggleQuoted(msgId: number) {
  expandedQuotedId.value = expandedQuotedId.value === msgId ? null : msgId
}

function agentInitial(msg: Message): string {
  const name = msg.authorName || "?"
  return name[0].toUpperCase()
}

function agentAvatarColor(msg: Message): string {
  return avatarColor(msg.authorName || "Agent")
}

// Auto-link URLs in plain text (escapes HTML first to prevent injection)
function autoLinkUrls(text: string): string {
  const escaped = text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
  return escaped.replace(
    /https?:\/\/[^\s<>"')\]]+/g,
    (url) => `<a href="${url}" target="_blank" rel="noopener" class="auto-link">${url}</a>`
  )
}

// Web chat body parser — splits "(HH:MM:SS) Speaker: text" lines into structured entries
type ChatRole = "customer" | "bot" | "agent"

interface ChatLine {
  speaker: string
  role: ChatRole
  time: string
  text: string
}

function chatSpeakerRole(speaker: string): ChatRole {
  if (/bot\b/i.test(speaker) || speaker.toLowerCase() === "system") return "bot"
  if (/^web user\b/i.test(speaker)) return "customer"
  return "agent"
}

function parseWebChatLines(body: string): ChatLine[] {
  const lines: ChatLine[] = []
  // Split on timestamp markers: (HH:MM:SS)
  const parts = body.split(/(?=\(\d{1,2}:\d{2}:\d{2}\)\s)/)
  for (const part of parts) {
    const m = part.match(/^\((\d{1,2}:\d{2}:\d{2})\)\s+(.+?):\s*([\s\S]*)$/)
    if (!m) continue
    const time = m[1]
    const speaker = m[2].trim()
    const text = m[3].trim()
    if (!text) continue
    lines.push({ speaker, role: chatSpeakerRole(speaker), time, text })
  }
  return lines
}

// Load comments whenever the displayed ticket changes
watch(() => props.ticketId, (id) => { loadComments(id) }, { immediate: true })

const activeChannel = computed(() => {
  if (replyChannel.value) return replyChannel.value
  const firstCustomerMsg = ticket.value?.messages.find((m) => m.from === "customer")
  return firstCustomerMsg?.channel ?? "chat"
})

const isPhoneChannel = computed(() => activeChannel.value === "phone")

const customerPhoneDisplay = computed(() => ticket.value?.phone || "(555) 867-5309")
const dialNumber = computed(() => callNumber.value.trim() || customerPhoneDisplay.value)

const isIdle = computed(() => calls.value.length === 0)
const activeCall = computed(() => calls.value.find((c) => c.status === "connected" || c.status === "ringing"))
const heldCall = computed(() => calls.value.find((c) => c.status === "on-hold"))
const hasTwoCalls = computed(() => calls.value.length === 2)
const canStartSecondCall = computed(() => calls.value.length === 1 && calls.value[0].status === "on-hold" && !isMerged.value)
const primaryCall = computed(() => activeCall.value || heldCall.value)

function formattedTime(call: Call) {
  const mins = Math.floor(call.seconds / 60)
  const secs = call.seconds % 60
  return `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`
}

// Phone call functions
function startCallTimer(call: Call) {
  stopCallTimer(call)
  call.timerHandle = setInterval(() => {
    call.seconds++
  }, 1000)
}

function stopCallTimer(call: Call) {
  if (call.timerHandle) {
    clearInterval(call.timerHandle)
    call.timerHandle = null
  }
}

function clearRingTimer(call: Call) {
  if (call.ringHandle) {
    clearTimeout(call.ringHandle)
    call.ringHandle = null
  }
}

function addRecording(duration: number) {
  if (duration > 0 && ticket.value) {
    ticket.value.messages.push({
      id: ticket.value.messages.length + 1,
      from: "system",
      channel: "phone",
      time: "just now",
      text: "",
      type: "recording",
      recording: {
        duration,
        waveform: generateFakeWaveform(),
        transcript: generateFakeTranscript(ticket.value.name, duration),
      },
    })
    scrollToBottom()
  }
}

function startCall() {
  showCallMenu.value = false
  showCustomNumber.value = false
  showNewCall.value = false
  const isSecond = calls.value.length === 1
  const name = isSecond ? "Call 2" : (ticket.value?.name ?? "Customer")
  const id = ++callIdCounter
  calls.value.push({
    id,
    name,
    number: dialNumber.value,
    status: "ringing",
    seconds: 0,
    timerHandle: null,
    ringHandle: null,
  })
  const call = calls.value[calls.value.length - 1]
  call.ringHandle = setTimeout(() => {
    call.status = "connected"
    call.ringHandle = null
    if (!sessionStartTime.value) sessionStartTime.value = Date.now()
    startCallTimer(call)
  }, 2000)
  callNumber.value = ""
}

function endSession() {
  if (sessionStartTime.value) {
    const duration = Math.round((Date.now() - sessionStartTime.value) / 1000)
    sessionStartTime.value = null
    addRecording(duration)
  }
}

function hangUp() {
  if (isMerged.value) {
    calls.value.forEach((c) => {
      stopCallTimer(c)
      clearRingTimer(c)
    })
    calls.value = []
    isMerged.value = false
    isMuted.value = false
    endSession()
    return
  }

  const active = activeCall.value
  if (!active) return

  stopCallTimer(active)
  clearRingTimer(active)

  const idx = calls.value.findIndex((c) => c.id === active.id)
  if (idx !== -1) calls.value.splice(idx, 1)

  // If a held call remains, resume it
  const held = calls.value.find((c) => c.status === "on-hold")
  if (held) {
    held.status = "connected"
    startCallTimer(held)
  }

  if (calls.value.length === 0) {
    isMuted.value = false
    endSession()
  }
}

function toggleHold() {
  if (calls.value.length !== 1) return
  const call = calls.value[0]
  if (call.status === "connected") {
    call.status = "on-hold"
    stopCallTimer(call)
  } else if (call.status === "on-hold") {
    call.status = "connected"
    showNewCall.value = false
    startCallTimer(call)
  }
}

function toggleMute() {
  isMuted.value = !isMuted.value
}

function swapCalls() {
  if (calls.value.length !== 2) return
  const active = activeCall.value
  const held = heldCall.value
  if (!active || !held) return
  stopCallTimer(active)
  active.status = "on-hold"
  held.status = "connected"
  startCallTimer(held)
}

function mergeCalls() {
  if (calls.value.length !== 2) return
  isMerged.value = true
  calls.value.forEach((c) => {
    if (c.status === "on-hold") {
      c.status = "connected"
      startCallTimer(c)
    }
  })
}

function toggleHoldMerged() {
  const allHeld = calls.value.every((c) => c.status === "on-hold")
  if (allHeld) {
    calls.value.forEach((c) => {
      c.status = "connected"
      startCallTimer(c)
    })
  } else {
    calls.value.forEach((c) => {
      c.status = "on-hold"
      stopCallTimer(c)
    })
  }
}

function dropCall(callId: number) {
  const idx = calls.value.findIndex((c) => c.id === callId)
  if (idx === -1) return
  const call = calls.value[idx]
  stopCallTimer(call)
  clearRingTimer(call)
  calls.value.splice(idx, 1)

  if (calls.value.length <= 1) {
    isMerged.value = false
    // Resume the remaining call if it was on hold
    const remaining = calls.value[0]
    if (remaining && remaining.status === "on-hold") {
      remaining.status = "connected"
      startCallTimer(remaining)
    }
  }

  if (calls.value.length === 0) {
    endSession()
  }
}

function requestHangUp() {
  if (!primaryCall.value) return
  // No confirmation for ringing calls
  if (primaryCall.value.status === "ringing") {
    hangUp()
    return
  }
  if (isMerged.value) {
    confirmAction.value = { type: "hangup", message: "End conference? All participants will be disconnected.", label: "End All" }
  } else if (hasTwoCalls.value) {
    confirmAction.value = { type: "hangup", message: "End this call? The held call will resume.", label: "End" }
  } else {
    confirmAction.value = { type: "hangup", message: "End this call?", label: "End" }
  }
}

function requestMerge() {
  confirmAction.value = { type: "merge", message: "Merge these calls into a conference?", label: "Merge" }
}

function requestDrop(call: Call) {
  confirmAction.value = { type: "drop", callId: call.id, message: `Remove ${call.name} from the conference?`, label: "Remove" }
}

function executeConfirm() {
  if (!confirmAction.value) return
  const action = confirmAction.value
  confirmAction.value = null
  if (action.type === "hangup") hangUp()
  else if (action.type === "merge") mergeCalls()
  else if (action.type === "drop") dropCall(action.callId!)
}

function cancelConfirm() {
  confirmAction.value = null
}

function hangUpAll() {
  calls.value.forEach((c) => {
    stopCallTimer(c)
    clearRingTimer(c)
  })
  const hadCalls = calls.value.length > 0
  calls.value = []
  isMerged.value = false
  isMuted.value = false
  confirmAction.value = null
  showNewCall.value = false
  showCallMenu.value = false
  showCustomNumber.value = false
  callNumber.value = ""
  if (hadCalls) endSession()
}

// Recording helpers
function formatDuration(seconds: number) {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${String(secs).padStart(2, "0")}`
}

function generateFakeWaveform() {
  const bars = 40
  const waveform = []
  for (let i = 0; i < bars; i++) {
    const position = i / bars
    const envelope = position < 0.1 ? position / 0.1
      : position > 0.85 ? (1 - position) / 0.15
      : 1
    const randomness = 0.3 + Math.random() * 0.7
    waveform.push(Math.max(0.15, Math.min(1, envelope * randomness)))
  }
  return waveform
}

function generateFakeTranscript(customerName: string, durationSeconds: number) {
  const firstName = customerName.split(" ")[0]
  const allLines = [
    { speaker: "Agent", offset: 0, text: `Hi ${firstName}, thanks for calling in. How can I help you today?` },
    { speaker: firstName, offset: 8, text: "Yeah, I've been having some issues with my internet connection dropping out." },
    { speaker: "Agent", offset: 18, text: "I'm sorry to hear that. Let me pull up your account and take a look." },
    { speaker: firstName, offset: 28, text: "It's been happening on and off for the past couple of days." },
    { speaker: "Agent", offset: 38, text: "I can see some signal fluctuations on your line. Let me run a quick diagnostic." },
    { speaker: "Agent", offset: 55, text: "Alright, I've refreshed your connection from our end. That should help stabilize things." },
    { speaker: firstName, offset: 68, text: "Okay great, I'll keep an eye on it. Anything else I should do?" },
    { speaker: "Agent", offset: 78, text: `If it happens again, don't hesitate to call back. Is there anything else I can help with, ${firstName}?` },
  ]
  return allLines
    .filter((l) => l.offset < durationSeconds)
    .map((l) => ({ speaker: l.speaker, time: formatDuration(l.offset), text: l.text }))
}

function playbackElapsedFor(msg: Message) {
  if (playingRecordingId.value !== msg.id) return "0:00"
  return formatDuration(Math.floor(playbackProgress.value * msg.recording!.duration))
}

// Recording playback
function togglePlayback(msg: Message) {
  if (playingRecordingId.value === msg.id) {
    stopPlayback()
  } else {
    stopPlayback()
    startPlayback(msg)
  }
}

function startPlayback(msg: Message) {
  playingRecordingId.value = msg.id
  playbackProgress.value = 0
  const duration = msg.recording!.duration
  const stepMs = 50
  const increment = stepMs / (duration * 1000)
  playbackTimerHandle.value = setInterval(() => {
    playbackProgress.value += increment
    if (playbackProgress.value >= 1) {
      playbackProgress.value = 1
      stopPlayback()
    }
  }, stepMs)
}

function stopPlayback() {
  if (playbackTimerHandle.value) {
    clearInterval(playbackTimerHandle.value)
    playbackTimerHandle.value = null
  }
  playingRecordingId.value = null
  playbackProgress.value = 0
}

// Real audio playback (call/voicemail recordings)
function toggleAudioPlayback(msg: Message) {
  if (audioPlayingMsgId.value === msg.id) {
    if (audioEl.value?.paused) {
      audioEl.value.play()
    } else {
      audioEl.value?.pause()
    }
    return
  }
  stopAudioPlayback()
  audioLoading.value = true
  audioPlayingMsgId.value = msg.id
  const audio = new Audio(recordingUrl(msg))
  audioEl.value = audio
  audio.addEventListener("loadedmetadata", () => {
    audioDuration.value = audio.duration
    audioLoading.value = false
  })
  audio.addEventListener("timeupdate", () => {
    audioCurrentTime.value = audio.currentTime
  })
  audio.addEventListener("ended", () => {
    stopAudioPlayback()
  })
  audio.addEventListener("error", () => {
    audioLoading.value = false
    stopAudioPlayback()
  })
  audio.play()
}

function stopAudioPlayback() {
  if (audioEl.value) {
    audioEl.value.pause()
    audioEl.value.src = ""
    audioEl.value = null
  }
  audioPlayingMsgId.value = null
  audioCurrentTime.value = 0
  audioDuration.value = 0
  audioLoading.value = false
}

function seekAudio(event: MouseEvent) {
  if (!audioEl.value || !audioDuration.value) return
  const bar = event.currentTarget as HTMLElement
  const rect = bar.getBoundingClientRect()
  const pct = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width))
  audioEl.value.currentTime = pct * audioDuration.value
}

function audioProgressFor(msg: Message): number {
  if (audioPlayingMsgId.value !== msg.id || !audioDuration.value) return 0
  return audioCurrentTime.value / audioDuration.value
}

function audioIsPlaying(msg: Message): boolean {
  return audioPlayingMsgId.value === msg.id && !!audioEl.value && !audioEl.value.paused
}

function audioTimeDisplay(msg: Message): string {
  if (audioPlayingMsgId.value !== msg.id) return "0:00"
  return formatDuration(Math.floor(audioCurrentTime.value))
}

function audioDurationDisplay(msg: Message): string {
  if (audioPlayingMsgId.value === msg.id && audioDuration.value) {
    return formatDuration(Math.floor(audioDuration.value))
  }
  // Fall back to call/voicemail metadata duration if available
  const raw = msg.call?.duration ?? msg.voicemail?.duration
  if (raw) return raw
  return "--:--"
}

function toggleTranscript(msgId: number) {
  expandedTranscriptId.value = expandedTranscriptId.value === msgId ? null : msgId
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) {
      messagesEl.value.scrollTop = messagesEl.value.scrollHeight
    }
  })
}

function sendReply() {
  const text = replyText.value.trim()
  if (!text || !props.ticketId) return
  sharedSendReply(props.ticketId, text, activeChannel.value)
  replyText.value = ""
  scrollToBottom()
}

function followAi() {
  const suggestion = currentAi.value
  if (!suggestion || !props.ticketId) return
  activeTab.value = "comms"
  composeCollapsed.value = false
  replyText.value = suggestion.replyText
  nextTick(() => sendReply())
}

function handleResolve() {
  resolveTicket(props.ticketId)
  emit("resolve")
}

function handleAddTag() {
  const tag = newTag.value.trim().toLowerCase()
  if (!tag || !props.ticketId) return
  addTag(props.ticketId, tag)
  newTag.value = ""
}

function handleAction(action: string) {
  // Placeholder — will wire up real actions later
  console.log(`Action triggered: ${action} for ticket ${props.ticketId}`)
}

watch(showCustomNumber, (val) => {
  if (val) nextTick(() => customNumberInput.value?.focus())
})

watch(() => props.ticketId, () => {
  replyText.value = ""
  replyChannel.value = null
  activeTab.value = "comms"
  hangUpAll()
  stopPlayback()
  scrollToBottom()
})

onBeforeUnmount(() => {
  calls.value.forEach((c) => {
    stopCallTimer(c)
    clearRingTimer(c)
  })
  stopPlayback()
  stopAudioPlayback()
})
</script>

<style scoped>
/* ── Layout ────────────────────────────────────────────── */

.ticket-detail {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

/* ── Bottom dock ───────────────────────────────────────── */

.bottom-dock {
  flex-shrink: 0;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.compose-section {
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.compose-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 24px;
  background: none;
  border: none;
  color: rgba(148, 163, 184, 0.5);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
}

.compose-toggle:hover {
  color: #94a3b8;
  background: rgba(255, 255, 255, 0.02);
}

.compose-toggle svg {
  transition: transform 0.2s ease;
}

.chevron-flipped {
  transform: rotate(180deg);
}

/* ── AI panel ──────────────────────────────────────────── */

.ai-panel {
  padding: 16px 24px;
  background: rgba(168, 85, 247, 0.04);
}

.ai-panel--overlay {
  position: relative;
  overflow: hidden;
}

.ai-panel-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.ai-panel-headline {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  line-height: 1.3;
}

.ai-panel-body {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.6);
  line-height: 1.5;
  margin: 0 0 4px;
}

/* ── Actions tab ───────────────────────────────────────── */

.actions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  padding-top: 8px;
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px 14px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  cursor: pointer;
  font-family: inherit;
  text-align: center;
  transition: all 0.15s;
}

.action-card:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.12);
  transform: translateY(-1px);
}

.action-card:active {
  transform: translateY(0);
}

.action-icon { opacity: 0.85; }
.action-icon--blue { color: #60a5fa; }
.action-icon--orange { color: #fb923c; }
.action-icon--green { color: #34d399; }
.action-icon--purple { color: #c084fc; }
.action-icon--indigo { color: #818cf8; }

.action-label {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
}

.action-desc {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.45);
  line-height: 1.3;
}

/* ── Thread header ─────────────────────────────────────── */

.thread-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.thread-customer-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.thread-avatar {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.thread-name {
  font-size: 17px;
  font-weight: 600;
  color: #e2e8f0;
}

.thread-company {
  font-weight: 400;
  color: rgba(148, 163, 184, 0.6);
}

.thread-id {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.4);
  margin-top: 2px;
}

.thread-subject {
  font-size: 18px;
  font-weight: 600;
  color: #f1f5f9;
  letter-spacing: -0.01em;
}

.thread-badges {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
  padding-right: 36px;
}

.badge {
  font-size: 13px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 7px;
  text-transform: capitalize;
}

.badge--new { background: rgba(56, 189, 248, 0.15); color: #7dd3fc; }
.badge--open { background: rgba(99, 102, 241, 0.15); color: #a5b4fc; }
.badge--pending { background: rgba(168, 85, 247, 0.15); color: #d8b4fe; }
.badge--escalated { background: rgba(249, 115, 22, 0.15); color: #fdba74; }
.badge--solved { background: rgba(52, 211, 153, 0.15); color: #6ee7b7; }
.badge--closed { background: rgba(148, 163, 184, 0.15); color: #94a3b8; }
.badge--high { background: rgba(239, 68, 68, 0.12); color: #fca5a5; }
.badge--medium { background: rgba(245, 158, 11, 0.12); color: #fcd34d; }
.badge--low { background: rgba(52, 211, 153, 0.12); color: #6ee7b7; }

/* ── Tab bar ────────────────────────────────────────────── */

.tab-bar {
  display: flex;
  gap: 2px;
  padding: 6px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  flex-shrink: 0;
}

.zendesk-link {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 24px;
  font-size: 11px;
  font-weight: 500;
  color: rgba(148, 163, 184, 0.45);
  text-decoration: none;
  transition: color 0.15s;
  flex-shrink: 0;
}

.zendesk-link:hover {
  color: #94a3b8;
}

.tab-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: rgba(148, 163, 184, 0.4);
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #94a3b8;
}

.tab-btn--active {
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
}

/* ── Channel badges ────────────────────────────────────── */

.msg-channel {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.msg-channel--chat { background: rgba(56, 189, 248, 0.1); color: #7dd3fc; }
.msg-channel--email { background: rgba(99, 102, 241, 0.1); color: #a5b4fc; }
.msg-channel--sms { background: rgba(52, 211, 153, 0.1); color: #6ee7b7; }
.msg-channel--phone,
.msg-channel--voice { background: rgba(245, 158, 11, 0.1); color: #fcd34d; }
.msg-channel--web { background: rgba(148, 163, 184, 0.1); color: #94a3b8; }
.msg-channel--reply { background: rgba(52, 211, 153, 0.1); color: #6ee7b7; }
.msg-channel--internal { background: rgba(251, 191, 36, 0.1); color: #fbbf24; }

/* ── Tab panels (shared) ───────────────────────────────── */

.tab-panel {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  animation: content-up 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes content-up {
  from { opacity: 0; transform: translateY(12px); }
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #e2e8f0;
  margin-bottom: 4px;
}

.panel-subtitle {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.45);
  margin-bottom: 20px;
}

.panel-empty {
  font-size: 14px;
  color: rgba(148, 163, 184, 0.35);
  padding: 24px 0;
  text-align: center;
}

/* ── Contact tab ───────────────────────────────────────── */

.contact-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0 28px;
}

.contact-avatar {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
  color: #fff;
  margin-bottom: 12px;
}

.contact-name {
  font-size: 18px;
  font-weight: 700;
  color: #f1f5f9;
}

.contact-company {
  font-size: 14px;
  color: rgba(148, 163, 184, 0.5);
  margin-top: 2px;
}

.detail-grid {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-row {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px 16px;
  border-radius: 10px;
  transition: background 0.15s;
}

.detail-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.detail-icon {
  color: rgba(148, 163, 184, 0.4);
  margin-top: 2px;
  flex-shrink: 0;
}

.detail-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(148, 163, 184, 0.35);
  margin-bottom: 2px;
}

.detail-value {
  font-size: 15px;
  font-weight: 500;
  color: #e2e8f0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-sub {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.4);
  margin-top: 2px;
}

.sub-status {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 5px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.sub-status--active { background: rgba(52, 211, 153, 0.1); color: #6ee7b7; }
.sub-status--trial { background: rgba(245, 158, 11, 0.1); color: #fcd34d; }
.sub-status--churned { background: rgba(239, 68, 68, 0.1); color: #fca5a5; }

/* ── Notes ──────────────────────────────────────────────── */

.notes-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.notes-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.4);
  margin-bottom: 10px;
}

.notes-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 12px 16px;
  color: #e2e8f0;
  font-size: 14px;
  font-family: inherit;
  line-height: 1.6;
  resize: vertical;
  outline: none;
  transition: border-color 0.15s;
  min-height: 80px;
}

.notes-input::placeholder { color: rgba(148, 163, 184, 0.3); }
.notes-input:focus { border-color: rgba(99, 102, 241, 0.4); }

/* ── Timeline (ticket history) ─────────────────────────── */

.timeline {
  display: flex;
  flex-direction: column;
  padding-top: 8px;
}

.timeline-item {
  display: flex;
  gap: 14px;
  padding: 10px 0;
  position: relative;
}

.timeline-item:not(:last-child)::before {
  content: "";
  position: absolute;
  left: 5px;
  top: 28px;
  bottom: -2px;
  width: 1px;
  background: rgba(255, 255, 255, 0.06);
}

.timeline-dot {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: rgba(99, 102, 241, 0.3);
  border: 2px solid rgba(99, 102, 241, 0.5);
  flex-shrink: 0;
  margin-top: 3px;
}

.timeline-event {
  font-size: 14px;
  color: #e2e8f0;
  line-height: 1.4;
}

.timeline-time {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.35);
  margin-top: 2px;
}

/* ── History cards (subscriber history) ────────────────── */

.history-card {
  padding: 14px 16px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  margin-bottom: 8px;
  transition: background 0.15s;
}

.history-card:hover { background: rgba(255, 255, 255, 0.04); }

.history-card--current {
  border-color: rgba(99, 102, 241, 0.2);
  background: rgba(99, 102, 241, 0.04);
}

.history-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.history-tid {
  font-size: 12px;
  font-weight: 600;
  color: rgba(148, 163, 184, 0.5);
}

.history-status {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 5px;
  text-transform: capitalize;
}

.history-status--open,
.history-status--new { background: rgba(99, 102, 241, 0.12); color: #a5b4fc; }
.history-status--pending { background: rgba(168, 85, 247, 0.12); color: #d8b4fe; }
.history-status--escalated { background: rgba(249, 115, 22, 0.12); color: #fdba74; }
.history-status--solved { background: rgba(52, 211, 153, 0.12); color: #6ee7b7; }
.history-status--closed { background: rgba(148, 163, 184, 0.12); color: #94a3b8; }

.history-subject {
  font-size: 14px;
  font-weight: 500;
  color: #e2e8f0;
  line-height: 1.35;
}

.history-date {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.35);
  margin-top: 4px;
}

/* ── Settings tab ──────────────────────────────────────── */

.settings-section { margin-bottom: 24px; }

.settings-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: rgba(148, 163, 184, 0.4);
  margin-bottom: 10px;
}

.tags-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
  color: #c7d2fe;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.15);
  border-radius: 6px;
  padding: 4px 8px;
}

.tag-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border: none;
  background: transparent;
  color: rgba(199, 210, 254, 0.5);
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
  border-radius: 3px;
  transition: color 0.15s, background 0.15s;
}

.tag-remove:hover {
  color: #fca5a5;
  background: rgba(239, 68, 68, 0.15);
}

.tag-add { display: inline-flex; }

.tag-input {
  width: 80px;
  padding: 4px 8px;
  font-size: 12px;
  font-family: inherit;
  background: rgba(255, 255, 255, 0.03);
  border: 1px dashed rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  color: #e2e8f0;
  outline: none;
  transition: border-color 0.15s;
}

.tag-input::placeholder { color: rgba(148, 163, 184, 0.3); }

.tag-input:focus {
  border-color: rgba(99, 102, 241, 0.4);
  border-style: solid;
}

.option-row {
  display: flex;
  gap: 6px;
}

.option-row--wrap { flex-wrap: wrap; }

.option-btn {
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  color: rgba(148, 163, 184, 0.6);
  cursor: pointer;
  text-transform: capitalize;
  transition: all 0.15s;
}

.option-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
}

.option-btn--active {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

.option-btn--hot.option-btn--active {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.25);
  color: #fca5a5;
}

.option-btn--warm.option-btn--active {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.25);
  color: #fcd34d;
}

.option-btn--cool.option-btn--active {
  background: rgba(52, 211, 153, 0.12);
  border-color: rgba(52, 211, 153, 0.25);
  color: #6ee7b7;
}

.settings-select {
  width: 100%;
  padding: 9px 14px;
  font-size: 14px;
  font-family: inherit;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 9px;
  color: #e2e8f0;
  outline: none;
  cursor: pointer;
  appearance: none;
  transition: border-color 0.15s;
}

.settings-select:focus { border-color: rgba(99, 102, 241, 0.4); }
.settings-select option { background: #0f172a; color: #e2e8f0; }

/* ── Messages ───────────────────────────────────────────── */

.thread-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.message--agent { flex-direction: row-reverse; }

.msg-avatar {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.msg-avatar--agent {
  /* background set dynamically via :style binding */
}

.msg-bubble {
  max-width: 72%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 14px;
  padding: 14px 18px;
}

.message--agent .msg-bubble {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.2);
}

/* Internal notes span the full width with an amber tint */
.message--internal {
  flex-direction: row;
  align-self: stretch;
}

.message--internal .msg-bubble {
  max-width: 100%;
  width: 100%;
  background: rgba(251, 191, 36, 0.05);
  border-color: rgba(251, 191, 36, 0.2);
  border-left: 3px solid rgba(251, 191, 36, 0.5);
}

.msg-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.msg-sender {
  font-size: 14px;
  font-weight: 600;
  color: #94a3b8;
}

.msg-automated {
  font-size: 10px;
  font-weight: 600;
  color: #a78bfa;
  background: rgba(167, 139, 250, 0.15);
  padding: 2px 7px;
  border-radius: 4px;
  letter-spacing: 0.03em;
}

.msg-time {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.4);
}

.msg-body {
  font-size: 16px;
  color: #e2e8f0;
  line-height: 1.6;
}

/* ── Email body ─────────────────────────────────────────── */

.msg-body--email {
  white-space: pre-wrap;
  word-break: break-word;
}

.email-content {
  font-size: 15px;
  line-height: 1.7;
}

.email-content--html {
  overflow-x: auto;
  line-height: 1.6;
  color: #e2e8f0;
}

/* Tables rendered from rich HTML email bodies */
.email-content--html :deep(table) {
  border-collapse: collapse;
  margin: 8px 0;
  font-size: 12.5px;
}

.email-content--html :deep(th),
.email-content--html :deep(td) {
  padding: 5px 10px;
  text-align: left !important;
  vertical-align: top !important;
  border: none;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

/* Collapse empty cells (email layout artifacts) */
.email-content--html :deep(td:empty),
.email-content--html :deep(th:empty) {
  padding: 0;
  width: 0;
  max-width: 0;
  overflow: hidden;
  border: none;
}

.email-content--html :deep(th) {
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: #64748b;
  padding-bottom: 6px;
}

.email-content--html :deep(td) {
  color: #cbd5e1;
}

.email-content--html :deep(td:first-child) {
  color: #94a3b8;
  font-weight: 500;
  white-space: nowrap;
  padding-right: 16px;
}

.email-content--html :deep(tr:last-child td) {
  border-bottom: none;
}

.email-content--html :deep(tr:hover td:not(:empty)) {
  background: rgba(255, 255, 255, 0.02);
}

/* Strip margins from all elements inside cells so rows align to the top */
.email-content--html :deep(td > *),
.email-content--html :deep(th > *) {
  margin: 0 !important;
  padding: 0 !important;
  vertical-align: top !important;
}

/* Lists in rich email bodies */
.email-content--html :deep(ul),
.email-content--html :deep(ol) {
  padding-left: 20px;
  margin: 8px 0;
}

.email-content--html :deep(li) {
  margin: 4px 0;
  color: #cbd5e1;
  font-size: 14px;
  line-height: 1.6;
}

.email-content--html :deep(blockquote) {
  margin: 8px 0;
  padding: 8px 12px;
  border-left: 2px solid rgba(99, 102, 241, 0.25);
  color: rgba(148, 163, 184, 0.7);
  font-size: 13px;
}

.email-content--html :deep(p) {
  margin: 6px 0;
}

.email-content--html :deep(strong),
.email-content--html :deep(b) {
  font-weight: 600;
  color: #e2e8f0;
}

.email-content--html :deep(a) {
  color: #818cf8;
  text-decoration: none;
}

.email-content--html :deep(a:hover) {
  text-decoration: underline;
}

.email-quoted {
  margin-top: 12px;
}

.email-quoted-toggle {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border: none;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.08);
  color: #818cf8;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s;
}

.email-quoted-toggle:hover {
  background: rgba(99, 102, 241, 0.15);
}

.email-quoted-toggle svg {
  transition: transform 0.2s;
}

.email-quoted-text {
  margin-top: 8px;
  padding: 10px 12px;
  border-left: 2px solid rgba(99, 102, 241, 0.25);
  border-radius: 0 6px 6px 0;
  background: rgba(0, 0, 0, 0.15);
  color: rgba(148, 163, 184, 0.7);
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

/* ── Channel bar ────────────────────────────────────────── */

.channel-bar {
  display: flex;
  gap: 4px;
  margin-bottom: 10px;
}

.channel-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  border-radius: 7px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  color: rgba(148, 163, 184, 0.5);
  font-size: 12px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.15s;
}

.channel-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.12);
  color: #94a3b8;
}

.channel-btn--active.channel-btn--chat {
  background: rgba(56, 189, 248, 0.1);
  border-color: rgba(56, 189, 248, 0.25);
  color: #7dd3fc;
}

.channel-btn--active.channel-btn--sms {
  background: rgba(52, 211, 153, 0.1);
  border-color: rgba(52, 211, 153, 0.25);
  color: #6ee7b7;
}

.channel-btn--active.channel-btn--email {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

.channel-btn--active.channel-btn--phone {
  background: rgba(245, 158, 11, 0.1);
  border-color: rgba(245, 158, 11, 0.25);
  color: #fcd34d;
}

.channel-label {
  line-height: 1;
}

/* ── Compose ────────────────────────────────────────────── */

.thread-compose {
  padding: 0 24px 16px;
  flex-shrink: 0;
}

.compose-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  padding: 14px 18px;
  color: #e2e8f0;
  font-size: 16px;
  font-family: inherit;
  resize: none;
  outline: none;
  transition: border-color 0.15s;
}

.compose-input::placeholder { color: rgba(148, 163, 184, 0.3); }
.compose-input:focus { border-color: rgba(99, 102, 241, 0.4); }

.compose-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

/* ── Buttons ────────────────────────────────────────────── */

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  border: none;
  transition: all 0.15s;
}

.btn--ghost {
  background: rgba(255, 255, 255, 0.04);
  color: #64748b;
}

.btn--ghost:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #94a3b8;
}

.btn--primary {
  background: linear-gradient(135deg, #6366f1, #a855f7);
  color: #fff;
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.25);
}

.btn--primary:hover:not(:disabled) {
  box-shadow: 0 4px 24px rgba(99, 102, 241, 0.4);
}

.btn--primary:disabled {
  opacity: 0.4;
  cursor: default;
}

.btn--ai {
  background: rgba(168, 85, 247, 0.12);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.2);
  width: 100%;
  justify-content: center;
  margin-top: 8px;
}

.btn--ai:hover {
  background: rgba(168, 85, 247, 0.2);
  border-color: rgba(168, 85, 247, 0.35);
}

.ai-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  font-weight: 700;
  color: #c084fc;
  background: rgba(168, 85, 247, 0.15);
  border-radius: 6px;
  padding: 3px 8px;
  letter-spacing: 0.04em;
  flex-shrink: 0;
}

/* ── Phone call UI ─────────────────────────────────────── */

.phone-ui {
  padding: 8px 0 0;
  animation: content-up 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.phone-status-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  margin-bottom: 10px;
  transition: all 0.25s;
}

.phone-status-card--ringing {
  background: rgba(245, 158, 11, 0.06);
  border-color: rgba(245, 158, 11, 0.2);
}

.phone-status-card--connected {
  background: rgba(52, 211, 153, 0.06);
  border-color: rgba(52, 211, 153, 0.2);
}

.phone-status-card--on-hold {
  background: rgba(99, 102, 241, 0.06);
  border-color: rgba(99, 102, 241, 0.2);
}

.phone-status-card--held {
  padding: 10px 14px;
  margin-bottom: 8px;
  opacity: 0.75;
}

.phone-status-card--held .phone-status-icon {
  width: 32px;
  height: 32px;
  border-radius: 9px;
}

.phone-status-card--held .phone-status-label {
  font-size: 12px;
}

.phone-status-card--held .phone-status-number {
  font-size: 12px;
}

.phone-status-card--held .phone-timer {
  font-size: 12px;
}

.phone-status-icon {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: rgba(245, 158, 11, 0.12);
  color: #fcd34d;
  transition: all 0.25s;
}

.phone-status-card--connected .phone-status-icon {
  background: rgba(52, 211, 153, 0.12);
  color: #6ee7b7;
}

.phone-status-card--on-hold .phone-status-icon {
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
}

.phone-status-icon--ringing {
  animation: pulse-ring 1.2s ease-in-out infinite;
}

@keyframes pulse-ring {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.1); opacity: 0.7; }
}

.phone-status-info {
  flex: 1;
  min-width: 0;
}

.phone-status-label {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
}

.phone-status-number {
  font-size: 13px;
  color: rgba(148, 163, 184, 0.5);
  margin-top: 1px;
}

/* Call menu */

.call-menu-wrap {
  position: relative;
  flex: 1;
}

.call-menu {
  position: absolute;
  bottom: calc(100% + 6px);
  left: 0;
  right: 0;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 4px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  animation: content-up 0.15s cubic-bezier(0.16, 1, 0.3, 1);
  z-index: 10;
}

.call-menu--inline {
  position: static;
  margin-bottom: 10px;
  animation: content-up 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.call-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  border-radius: 7px;
  background: none;
  color: #e2e8f0;
  font-family: inherit;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.12s;
}

.call-menu-item:hover {
  background: rgba(245, 158, 11, 0.1);
}

.call-menu-item--other {
  color: rgba(148, 163, 184, 0.6);
}

.call-menu-text {
  display: flex;
  flex-direction: column;
  text-align: left;
}

.call-menu-label {
  font-weight: 600;
  line-height: 1.3;
}

.call-menu-number {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.5);
  line-height: 1.3;
}

.call-menu-divider {
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  margin: 2px 8px;
}

.call-menu-custom {
  display: flex;
  gap: 4px;
  padding: 6px 6px 4px;
}

.call-menu-input {
  flex: 1;
  padding: 7px 10px;
  font-size: 13px;
  font-family: inherit;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 7px;
  color: #e2e8f0;
  outline: none;
  transition: border-color 0.15s;
}

.call-menu-input::placeholder {
  color: rgba(148, 163, 184, 0.3);
}

.call-menu-input:focus {
  border-color: rgba(245, 158, 11, 0.4);
}

.call-menu-dial {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  border: 1px solid rgba(245, 158, 11, 0.25);
  border-radius: 7px;
  background: rgba(245, 158, 11, 0.12);
  color: #fcd34d;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.call-menu-dial:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.2);
  border-color: rgba(245, 158, 11, 0.35);
}

.call-menu-dial:disabled {
  opacity: 0.35;
  cursor: default;
}

.phone-timer {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: #fcd34d;
  flex-shrink: 0;
}

.phone-status-card--on-hold .phone-timer {
  color: #a5b4fc;
}

/* Participants */

.phone-participants {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 10px;
}

.phone-participant {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 8px;
  transition: background 0.15s;
}

.phone-participant:hover {
  background: rgba(255, 255, 255, 0.03);
}

.phone-participant-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.phone-participant-dot--connected {
  background: #6ee7b7;
}

.phone-participant-dot--ringing {
  background: #fcd34d;
  animation: pulse-ring 1.2s ease-in-out infinite;
}

.phone-participant-dot--on-hold {
  background: #a5b4fc;
}

.phone-participant-name {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  flex: 1;
  min-width: 0;
}

.phone-participant-status {
  font-size: 12px;
  color: rgba(148, 163, 184, 0.45);
  text-transform: capitalize;
}

.phone-participant-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: rgba(148, 163, 184, 0.3);
  cursor: pointer;
  flex-shrink: 0;
  opacity: 0;
  transition: all 0.15s;
}

.phone-participant:hover .phone-participant-remove {
  opacity: 1;
}

.phone-participant-remove:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #fca5a5;
}

/* Phone buttons */

.phone-controls {
  display: flex;
  gap: 6px;
}

.phone-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  flex: 1;
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.04);
  color: rgba(148, 163, 184, 0.7);
  transition: all 0.15s;
}

.phone-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.12);
  color: #e2e8f0;
}

.phone-btn:disabled {
  opacity: 0.35;
  cursor: default;
}

.phone-btn--active {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
}

.phone-btn--call {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.25);
  color: #fcd34d;
}

.phone-btn--call:hover {
  background: rgba(245, 158, 11, 0.2);
  border-color: rgba(245, 158, 11, 0.35);
  color: #fcd34d;
}

.phone-btn--end {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.25);
  color: #fca5a5;
}

.phone-btn--end:hover {
  background: rgba(239, 68, 68, 0.2);
  border-color: rgba(239, 68, 68, 0.35);
  color: #fca5a5;
}

/* Confirmation dialog */

.phone-confirm {
  padding: 14px 16px;
  border-radius: 12px;
  background: rgba(239, 68, 68, 0.06);
  border: 1px solid rgba(239, 68, 68, 0.15);
  margin-bottom: 10px;
  animation: content-up 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.phone-confirm-message {
  font-size: 13px;
  font-weight: 500;
  color: #e2e8f0;
  margin-bottom: 10px;
}

.phone-confirm-actions {
  display: flex;
  gap: 6px;
}

/* ── Message transitions ────────────────────────────────── */

.msg-enter-active {
  transition: opacity 0.3s ease, transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.msg-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

/* ── Recording card ────────────────────────────────────── */

.message--system {
  justify-content: center;
}

.recording-card {
  width: 100%;
  max-width: 400px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(245, 158, 11, 0.2);
  border-radius: 14px;
  padding: 16px;
  animation: content-up 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.recording-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.recording-icon {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  flex-shrink: 0;
}

.recording-title {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  flex: 1;
}

.recording-duration {
  font-size: 13px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: rgba(148, 163, 184, 0.5);
}

/* ── Waveform player ───────────────────────────────────── */

.waveform-player {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.waveform-play-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.waveform-play-btn:hover {
  background: rgba(245, 158, 11, 0.25);
  transform: scale(1.05);
}

.waveform-play-btn:active {
  transform: scale(0.95);
}

.waveform-bars {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 1px;
  height: 36px;
}

.waveform-bar {
  flex: 1;
  min-width: 2px;
  max-width: 5px;
  border-radius: 1px;
  background: rgba(245, 158, 11, 0.2);
  transition: background 0.1s;
}

.waveform-bar--played {
  background: #fbbf24;
}

.waveform-time {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  font-size: 11px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: rgba(148, 163, 184, 0.45);
  flex-shrink: 0;
  line-height: 1.4;
}

/* ── Transcript ────────────────────────────────────────── */

.transcript-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 8px 0 0;
  background: none;
  border: none;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  color: #c084fc;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: color 0.15s;
}

.transcript-toggle:hover {
  color: #d8b4fe;
}

.transcript-toggle svg:last-child {
  margin-left: auto;
  transition: transform 0.2s ease;
}

.transcript-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-top: 10px;
  animation: content-up 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.transcript-line {
  display: flex;
  gap: 10px;
  padding: 5px 0;
  font-size: 13px;
  line-height: 1.4;
}

.transcript-time {
  font-variant-numeric: tabular-nums;
  color: rgba(148, 163, 184, 0.35);
  flex-shrink: 0;
  min-width: 32px;
}

.transcript-speaker {
  font-weight: 600;
  color: #94a3b8;
  flex-shrink: 0;
  min-width: 50px;
}

.transcript-text {
  color: #cbd5e1;
}

/* ── Call / voicemail cards ──────────────────────────────── */

.call-card {
  background: rgba(99, 102, 241, 0.06);
  border: 1px solid rgba(99, 102, 241, 0.15);
  border-radius: 10px;
  padding: 12px 14px;
  margin-top: 4px;
}

.call-card--voicemail {
  background: rgba(168, 85, 247, 0.06);
  border-color: rgba(168, 85, 247, 0.15);
}

.call-card-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  color: #a5b4fc;
  margin-bottom: 8px;
}

.call-card--voicemail .call-card-header {
  color: #d8b4fe;
}

.call-card-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px 16px;
}

.call-card-field {
  display: flex;
  flex-direction: column;
  gap: 1px;
  font-size: 12px;
  color: #cbd5e1;
}

.call-card-label {
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgba(148, 163, 184, 0.5);
}

.call-card-recording {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 8px;
  padding: 4px 10px;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.1);
  color: #a5b4fc;
  font-size: 11px;
  font-weight: 500;
  text-decoration: none;
  transition: background 0.15s;
}

.call-card-recording:hover {
  background: rgba(99, 102, 241, 0.18);
}

/* ── Custom audio player for call recordings ───────────── */

.call-audio-player {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(99, 102, 241, 0.08);
}

.call-audio-player--voicemail {
  background: rgba(168, 85, 247, 0.08);
}

.call-audio-play-btn {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: none;
  background: rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s;
}

.call-audio-play-btn:hover {
  background: rgba(99, 102, 241, 0.35);
  transform: scale(1.05);
}

.call-audio-play-btn:active {
  transform: scale(0.95);
}

.call-audio-play-btn--voicemail {
  background: rgba(168, 85, 247, 0.2);
  color: #d8b4fe;
}

.call-audio-play-btn--voicemail:hover {
  background: rgba(168, 85, 247, 0.35);
}

.call-audio-track {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: rgba(99, 102, 241, 0.15);
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.call-audio-track--voicemail {
  background: rgba(168, 85, 247, 0.15);
}

.call-audio-progress {
  height: 100%;
  border-radius: 3px;
  background: #818cf8;
  transition: width 0.1s linear;
}

.call-audio-progress--voicemail {
  background: #c084fc;
}

.call-audio-time {
  font-size: 11px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: rgba(148, 163, 184, 0.5);
  flex-shrink: 0;
  white-space: nowrap;
}

.call-card-transcript {
  margin-top: 8px;
  padding: 10px 12px;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.2);
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
}

/* ── Web chat conversation ──────────────────────────────── */

.msg-body--webchat {
  padding: 0 !important;
}

.webchat-thread {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px 4px;
}

.webchat-line {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  max-width: 85%;
}

.webchat-line--bot {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.webchat-line--user {
  align-self: flex-start;
}

.webchat-avatar {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 2px;
}

.webchat-avatar--bot {
  background: rgba(99, 102, 241, 0.25);
  color: #818cf8;
}

.webchat-avatar--agent {
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
}

.webchat-avatar--customer {
  background: rgba(52, 211, 153, 0.2);
  color: #34d399;
}

.webchat-bubble {
  border-radius: 12px;
  padding: 6px 10px;
  font-size: 12.5px;
  line-height: 1.5;
}

.webchat-bubble--bot {
  background: rgba(255, 255, 255, 0.06);
  border-top-right-radius: 4px;
}

.webchat-bubble--user {
  background: rgba(52, 211, 153, 0.1);
  border-top-left-radius: 4px;
}

.webchat-meta {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 1px;
}

.webchat-speaker {
  font-size: 10.5px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.5);
}

.webchat-time {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.25);
}

.webchat-text {
  color: rgba(255, 255, 255, 0.82);
  white-space: pre-wrap;
  word-break: break-word;
}

/* ── Merge / system notices ─────────────────────────────── */

.msg-bubble--system {
  background: rgba(255, 255, 255, 0.02) !important;
  border: 1px dashed rgba(255, 255, 255, 0.08) !important;
}

.msg-body--merge {
  font-style: italic;
  color: rgba(148, 163, 184, 0.6);
  font-size: 12px;
}

.msg-body--note {
  border-left: 2px solid rgba(245, 158, 11, 0.4);
  padding-left: 10px;
}

/* Auto-linked URLs in plain text messages */
.msg-body :deep(.auto-link) {
  color: #818cf8;
  text-decoration: none;
  word-break: break-all;
}

.msg-body :deep(.auto-link:hover) {
  text-decoration: underline;
}
</style>
