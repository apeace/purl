package app

import "encoding/json"

// ZendeskWebhookPayload is the envelope sent for all webhook event types.
type ZendeskWebhookPayload struct {
	AccountID           string          `json:"account_id"`
	Detail              json.RawMessage `json:"detail"`
	Event               json.RawMessage `json:"event"`
	ID                  string          `json:"id"`
	Subject             string          `json:"subject"`
	Time                string          `json:"time"`
	Type                string          `json:"type"`
	ZendeskEventVersion string          `json:"zendesk_event_version"`
}

// ZendeskTicketDetail holds detail fields for ticket events.
type ZendeskTicketDetail struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	RequesterID string `json:"requester_id"`
	AssigneeID  string `json:"assignee_id"`
	GroupID     string `json:"group_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ZendeskUserDetail holds detail fields for user events.
type ZendeskUserDetail struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organization_id"`
	DefaultGroupID string `json:"default_group_id"`
	ExternalID     string `json:"external_id"`
	Suspended      bool   `json:"suspended"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ZendeskOrgDetail holds detail fields for organization events.
type ZendeskOrgDetail struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ExternalID string `json:"external_id"`
	GroupID    string `json:"group_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ZendeskArticleDetail holds detail fields for article events.
type ZendeskArticleDetail struct {
	ID      string `json:"id"`
	BrandID string `json:"brand_id"`
	UserID  string `json:"user_id"`
}

// ZendeskCommunityPostDetail holds detail fields for community post events.
type ZendeskCommunityPostDetail struct {
	ID      string `json:"id"`
	BrandID string `json:"brand_id"`
	TopicID string `json:"topic_id"`
	UserID  string `json:"user_id"`
}

// ZendeskMessagingTicketDetail holds detail fields for messaging ticket events.
type ZendeskMessagingTicketDetail struct {
	ID string `json:"id"`
}

// ZendeskAgentAvailabilityDetail holds detail fields for agent availability events.
type ZendeskAgentAvailabilityDetail struct {
	AccountID string `json:"account_id"`
}

// ZendeskLiveMetricsDetail holds detail fields for live metrics events.
type ZendeskLiveMetricsDetail struct {
	AccountID string `json:"account_id"`
	Metric    string `json:"metric"`
}

// ZendeskOmnichannelConfigDetail holds detail fields for omnichannel config events.
type ZendeskOmnichannelConfigDetail struct{}

// ZendeskFieldChangeEvent represents a field change with current and previous values.
type ZendeskFieldChangeEvent struct {
	Current  any `json:"current"`
	Previous any `json:"previous"`
}

// ZendeskCommentCreatedEvent represents a new comment on a ticket.
type ZendeskCommentCreatedEvent struct {
	ID       string `json:"id"`
	Body     string `json:"body"`
	AuthorID string `json:"author_id"`
	Public   bool   `json:"public"`
	HTMLBody string `json:"html_body"`
}

// ZendeskMessagingMessageEvent represents a new message in a messaging conversation.
type ZendeskMessagingMessageEvent struct {
	Actor          ZendeskMessagingActor `json:"actor"`
	ConversationID string                `json:"conversation_id"`
	Message        ZendeskMessagingMsg   `json:"message"`
}

// ZendeskMessagingActor identifies the sender of a messaging event.
type ZendeskMessagingActor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ZendeskMessagingMsg holds the message body within a messaging event.
type ZendeskMessagingMsg struct {
	ID   string `json:"id"`
	Body string `json:"body"`
}

// ZendeskAgentStateEvent represents a change in agent availability state.
type ZendeskAgentStateEvent struct {
	PreviousState string `json:"previous_state"`
	NewState      string `json:"new_state"`
	Channel       string `json:"channel"`
	UpdatedAt     string `json:"updated_at"`
}

// ZendeskAgentWorkItemEvent represents a work item being added/updated/removed for an agent.
type ZendeskAgentWorkItemEvent struct {
	WorkItemID string `json:"work_item_id"`
	TicketID   string `json:"ticket_id"`
	Channel    string `json:"channel"`
}

// ZendeskLiveMetricsValueEvent represents a live metrics value update.
type ZendeskLiveMetricsValueEvent struct {
	Value int `json:"value"`
}

// ZendeskOmnichannelRoutingEvent represents an omnichannel routing feature change.
type ZendeskOmnichannelRoutingEvent struct {
	Feature string `json:"feature"`
	Enabled bool   `json:"enabled"`
}

// Event type string constants.
const (
	// Ticket events
	EventTypeTicketCreated                = "zen:event-type:ticket.created"
	EventTypeTicketMerged                 = "zen:event-type:ticket.merged"
	EventTypeTicketSpam                   = "zen:event-type:ticket.spam"
	EventTypeTicketSoftDeleted            = "zen:event-type:ticket.soft_deleted"
	EventTypeTicketPermanentlyDeleted     = "zen:event-type:ticket.permanently_deleted"
	EventTypeTicketUndeleted              = "zen:event-type:ticket.undeleted"
	EventTypeTicketAssignmentChanged      = "zen:event-type:ticket.assignment_changed"
	EventTypeTicketGroupAssignmentChanged = "zen:event-type:ticket.group_assignment_changed"
	EventTypeTicketRequesterChanged       = "zen:event-type:ticket.requester_changed"
	EventTypeTicketSubmitterChanged       = "zen:event-type:ticket.submitter_changed"
	EventTypeTicketStatusChanged          = "zen:event-type:ticket.status_changed"
	EventTypeTicketCustomStatusChanged    = "zen:event-type:ticket.custom_status_changed"
	EventTypeTicketPriorityChanged        = "zen:event-type:ticket.priority_changed"
	EventTypeTicketSubjectChanged         = "zen:event-type:ticket.subject_changed"
	EventTypeTicketDescriptionChanged     = "zen:event-type:ticket.description_changed"
	EventTypeTicketTypeChanged            = "zen:event-type:ticket.type_changed"
	EventTypeTicketBrandChanged           = "zen:event-type:ticket.brand_changed"
	EventTypeTicketOrganizationChanged    = "zen:event-type:ticket.organization_changed"
	EventTypeTicketExternalIDChanged      = "zen:event-type:ticket.external_id_changed"
	EventTypeTicketTaskDueAtChanged       = "zen:event-type:ticket.task_due_at_changed"
	EventTypeTicketCustomFieldChanged     = "zen:event-type:ticket.custom_field_changed"
	EventTypeTicketFormChanged            = "zen:event-type:ticket.form_changed"
	EventTypeTicketTagsChanged            = "zen:event-type:ticket.tags_changed"
	EventTypeTicketCommentCreated         = "zen:event-type:ticket.comment_created"
	EventTypeTicketCommentMadePrivate     = "zen:event-type:ticket.comment_made_private"
	EventTypeTicketCommentRedacted        = "zen:event-type:ticket.comment_redacted"
	EventTypeTicketAttachmentLinked       = "zen:event-type:ticket.attachment_linked"
	EventTypeTicketAttachmentRedacted     = "zen:event-type:ticket.attachment_redacted"
	EventTypeTicketSLAPolicyChanged       = "zen:event-type:ticket.sla_policy_changed"
	EventTypeTicketScheduleChanged        = "zen:event-type:ticket.schedule_changed"
	EventTypeTicketOLAPolicyChanged       = "zen:event-type:ticket.ola_policy_changed"
	EventTypeTicketProblemLinkChanged     = "zen:event-type:ticket.problem_link_changed"
	EventTypeTicketNextSLABreachChanged   = "zen:event-type:ticket.next_sla_breach_changed"
	EventTypeTicketEmailCCsChanged        = "zen:event-type:ticket.email_ccs_changed"
	EventTypeTicketFollowersChanged       = "zen:event-type:ticket.followers_changed"
	EventTypeTicketCSATRequested          = "zen:event-type:ticket.csat_requested"
	EventTypeTicketCSATReceived           = "zen:event-type:ticket.csat_received"

	// User events
	EventTypeUserCreated                    = "zen:event-type:user.created"
	EventTypeUserDeleted                    = "zen:event-type:user.deleted"
	EventTypeUserMerged                     = "zen:event-type:user.merged"
	EventTypeUserAliasChanged               = "zen:event-type:user.alias_changed"
	EventTypeUserActiveChanged              = "zen:event-type:user.active_changed"
	EventTypeUserCustomFieldChanged         = "zen:event-type:user.custom_field_changed"
	EventTypeUserCustomRoleChanged          = "zen:event-type:user.custom_role_changed"
	EventTypeUserDefaultGroupChanged        = "zen:event-type:user.default_group_changed"
	EventTypeUserDetailsChanged             = "zen:event-type:user.details_changed"
	EventTypeUserExternalIDChanged          = "zen:event-type:user.external_id_changed"
	EventTypeUserGroupMembershipCreated     = "zen:event-type:user.group_membership_created"
	EventTypeUserGroupMembershipDeleted     = "zen:event-type:user.group_membership_deleted"
	EventTypeUserIdentityChanged            = "zen:event-type:user.identity_changed"
	EventTypeUserIdentityCreated            = "zen:event-type:user.identity_created"
	EventTypeUserIdentityDeleted            = "zen:event-type:user.identity_deleted"
	EventTypeUserLastLoginChanged           = "zen:event-type:user.last_login_changed"
	EventTypeUserNameChanged                = "zen:event-type:user.name_changed"
	EventTypeUserNotesChanged               = "zen:event-type:user.notes_changed"
	EventTypeUserOnlyPrivateCommentsChanged = "zen:event-type:user.only_private_comments_changed"
	EventTypeUserOrgMembershipCreated       = "zen:event-type:user.organization_membership_created"
	EventTypeUserOrgMembershipDeleted       = "zen:event-type:user.organization_membership_deleted"
	EventTypeUserPasswordChanged            = "zen:event-type:user.password_changed"
	EventTypeUserPhotoChanged               = "zen:event-type:user.photo_changed"
	EventTypeUserRoleChanged                = "zen:event-type:user.role_changed"
	EventTypeUserSuspendedChanged           = "zen:event-type:user.suspended_changed"
	EventTypeUserTagsChanged                = "zen:event-type:user.tags_changed"
	EventTypeUserTimeZoneChanged            = "zen:event-type:user.time_zone_changed"

	// Organization events
	EventTypeOrgCreated            = "zen:event-type:organization.created"
	EventTypeOrgDeleted            = "zen:event-type:organization.deleted"
	EventTypeOrgCustomFieldChanged = "zen:event-type:organization.custom_field_changed"
	EventTypeOrgExternalIDChanged  = "zen:event-type:organization.external_id_changed"
	EventTypeOrgNameChanged        = "zen:event-type:organization.name_changed"
	EventTypeOrgTagsChanged        = "zen:event-type:organization.tags_changed"

	// Agent availability events
	EventTypeAgentStateChanged        = "zen:event-type:agent.state_changed"
	EventTypeAgentWorkItemAdded       = "zen:event-type:agent.work_item_added"
	EventTypeAgentWorkItemUpdated     = "zen:event-type:agent.work_item_updated"
	EventTypeAgentWorkItemRemoved     = "zen:event-type:agent.work_item_removed"
	EventTypeAgentMaxCapacityChanged  = "zen:event-type:agent.max_capacity_changed"
	EventTypeAgentUnifiedStateChanged = "zen:event-type:agent.unified_state_changed"
	EventTypeAgentChannelCreated      = "zen:event-type:agent.channel_created"
	EventTypeAgentChannelDeleted      = "zen:event-type:agent.channel_deleted"
	EventTypeAgentGroupsUpdated       = "zen:event-type:agent.groups_updated"

	// Messaging events
	EventTypeMessagingTicketMessageAdded = "zen:event-type:messaging_ticket.message_added"

	// Article events
	EventTypeArticleAuthorChanged       = "zen:event-type:article.author_changed"
	EventTypeArticlePublished           = "zen:event-type:article.published"
	EventTypeArticleSubscriptionCreated = "zen:event-type:article.subscription_created"
	EventTypeArticleUnpublished         = "zen:event-type:article.unpublished"
	EventTypeArticleVoteCreated         = "zen:event-type:article.vote_created"
	EventTypeArticleVoteChanged         = "zen:event-type:article.vote_changed"
	EventTypeArticleVoteRemoved         = "zen:event-type:article.vote_removed"
	EventTypeArticleCommentCreated      = "zen:event-type:article.comment_created"
	EventTypeArticleCommentChanged      = "zen:event-type:article.comment_changed"
	EventTypeArticleCommentPublished    = "zen:event-type:article.comment_published"
	EventTypeArticleCommentUnpublished  = "zen:event-type:article.comment_unpublished"

	// Community post events
	EventTypeCommunityPostCreated             = "zen:event-type:community_post.created"
	EventTypeCommunityPostChanged             = "zen:event-type:community_post.changed"
	EventTypeCommunityPostPublished           = "zen:event-type:community_post.published"
	EventTypeCommunityPostUnpublished         = "zen:event-type:community_post.unpublished"
	EventTypeCommunityPostSubscriptionCreated = "zen:event-type:community_post.subscription_created"
	EventTypeCommunityPostVoteCreated         = "zen:event-type:community_post.vote_created"
	EventTypeCommunityPostVoteChanged         = "zen:event-type:community_post.vote_changed"
	EventTypeCommunityPostVoteRemoved         = "zen:event-type:community_post.vote_removed"
	EventTypeCommunityPostCommentCreated      = "zen:event-type:community_post.comment_created"
	EventTypeCommunityPostCommentChanged      = "zen:event-type:community_post.comment_changed"
	EventTypeCommunityPostCommentPublished    = "zen:event-type:community_post.comment_published"
	EventTypeCommunityPostCommentUnpublished  = "zen:event-type:community_post.comment_unpublished"
	EventTypeCommunityPostCommentVoteCreated  = "zen:event-type:community_post.comment_vote_created"
	EventTypeCommunityPostCommentVoteChanged  = "zen:event-type:community_post.comment_vote_changed"

	// Live messaging metrics events
	EventTypeLiveMetricsActiveAssigned                    = "zen:event-type:messaging_live_metrics.active_assigned_conversations"
	EventTypeLiveMetricsActiveAssignedByGroup             = "zen:event-type:messaging_live_metrics.active_assigned_conversations_by_group"
	EventTypeLiveMetricsActiveAssignedByViaType           = "zen:event-type:messaging_live_metrics.active_assigned_conversations_by_via_type"
	EventTypeLiveMetricsActiveUnassigned                  = "zen:event-type:messaging_live_metrics.active_unassigned_conversations"
	EventTypeLiveMetricsActiveUnassignedByGroup           = "zen:event-type:messaging_live_metrics.active_unassigned_conversations_by_group"
	EventTypeLiveMetricsActiveUnassignedByViaType         = "zen:event-type:messaging_live_metrics.active_unassigned_conversations_by_via_type"
	EventTypeLiveMetricsConversationsInQueue              = "zen:event-type:messaging_live_metrics.conversations_in_queue"
	EventTypeLiveMetricsConversationsInQueueByGroup       = "zen:event-type:messaging_live_metrics.conversations_in_queue_by_group"
	EventTypeLiveMetricsConversationsInQueueByViaType     = "zen:event-type:messaging_live_metrics.conversations_in_queue_by_via_type"
	EventTypeLiveMetricsTotalActive                       = "zen:event-type:messaging_live_metrics.total_active_conversations"
	EventTypeLiveMetricsTotalActiveByGroup                = "zen:event-type:messaging_live_metrics.total_active_conversations_by_group"
	EventTypeLiveMetricsTotalActiveByViaType              = "zen:event-type:messaging_live_metrics.total_active_conversations_by_via_type"
	EventTypeLiveMetricsAvgTimeInQueue                    = "zen:event-type:messaging_live_metrics.avg_time_in_queue"
	EventTypeLiveMetricsAvgTimeInQueueByGroup             = "zen:event-type:messaging_live_metrics.avg_time_in_queue_by_group"
	EventTypeLiveMetricsAvgTimeInQueueByViaType           = "zen:event-type:messaging_live_metrics.avg_time_in_queue_by_via_type"
	EventTypeLiveMetricsLongestTimeInQueue                = "zen:event-type:messaging_live_metrics.longest_time_in_queue"
	EventTypeLiveMetricsLongestTimeInQueueByGroup         = "zen:event-type:messaging_live_metrics.longest_time_in_queue_by_group"
	EventTypeLiveMetricsLongestTimeInQueueByViaType       = "zen:event-type:messaging_live_metrics.longest_time_in_queue_by_via_type"
	EventTypeLiveMetricsAvgRequesterWaitTime              = "zen:event-type:messaging_live_metrics.avg_requester_wait_time"
	EventTypeLiveMetricsAvgRequesterWaitTimeByGroup       = "zen:event-type:messaging_live_metrics.avg_requester_wait_time_by_group"
	EventTypeLiveMetricsAvgRequesterWaitTimeByViaType     = "zen:event-type:messaging_live_metrics.avg_requester_wait_time_by_via_type"
	EventTypeLiveMetricsLongestRequesterWaitTime          = "zen:event-type:messaging_live_metrics.longest_requester_wait_time"
	EventTypeLiveMetricsLongestRequesterWaitTimeByGroup   = "zen:event-type:messaging_live_metrics.longest_requester_wait_time_by_group"
	EventTypeLiveMetricsLongestRequesterWaitTimeByViaType = "zen:event-type:messaging_live_metrics.longest_requester_wait_time_by_via_type"
	EventTypeLiveMetricsAvgHandleTime                     = "zen:event-type:messaging_live_metrics.avg_handle_time"
	EventTypeLiveMetricsAvgHandleTimeByGroup              = "zen:event-type:messaging_live_metrics.avg_handle_time_by_group"
	EventTypeLiveMetricsAvgHandleTimeByViaType            = "zen:event-type:messaging_live_metrics.avg_handle_time_by_via_type"
	EventTypeLiveMetricsAvgConcurrency                    = "zen:event-type:messaging_live_metrics.avg_concurrency"
	EventTypeLiveMetricsAvgConcurrencyByGroup             = "zen:event-type:messaging_live_metrics.avg_concurrency_by_group"
	EventTypeLiveMetricsAvgConcurrencyByViaType           = "zen:event-type:messaging_live_metrics.avg_concurrency_by_via_type"

	// Omnichannel events
	EventTypeOmnichannelRoutingFeatureChanged = "zen:event-type:omnichannel_config.omnichannel_routing_feature_changed"
)
