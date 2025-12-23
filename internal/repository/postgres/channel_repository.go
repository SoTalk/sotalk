package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/sotalk/internal/domain/channel"
	"gorm.io/gorm"
)

type channelRepository struct {
	db *gorm.DB
}

// NewChannelRepository creates a new channel repository
func NewChannelRepository(db *gorm.DB) channel.Repository {
	return &channelRepository{
		db: db,
	}
}

// Create creates a new channel
func (r *channelRepository) Create(ctx context.Context, ch *channel.Channel) error {
	dbChannel := toDBChannel(ch)
	result := r.db.WithContext(ctx).Create(dbChannel)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID finds a channel by ID
func (r *channelRepository) FindByID(ctx context.Context, id uuid.UUID) (*channel.Channel, error) {
	var dbChannel Channel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dbChannel)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, channel.ErrChannelNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainChannel(&dbChannel), nil
}

// FindByUsername finds a channel by username
func (r *channelRepository) FindByUsername(ctx context.Context, username string) (*channel.Channel, error) {
	var dbChannel Channel
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&dbChannel)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, channel.ErrChannelNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainChannel(&dbChannel), nil
}

// FindByConversationID finds a channel by conversation ID
func (r *channelRepository) FindByConversationID(ctx context.Context, conversationID uuid.UUID) (*channel.Channel, error) {
	var dbChannel Channel
	result := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).First(&dbChannel)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, channel.ErrChannelNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainChannel(&dbChannel), nil
}

// FindByOwnerID finds all channels owned by a user
func (r *channelRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*channel.Channel, error) {
	var dbChannels []Channel
	result := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbChannels)

	if result.Error != nil {
		return nil, result.Error
	}

	channels := make([]*channel.Channel, len(dbChannels))
	for i, dbChannel := range dbChannels {
		channels[i] = toDomainChannel(&dbChannel)
	}

	return channels, nil
}

// FindPublicChannels finds all public channels
func (r *channelRepository) FindPublicChannels(ctx context.Context, limit, offset int) ([]*channel.Channel, error) {
	var dbChannels []Channel
	result := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Order("subscriber_count DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbChannels)

	if result.Error != nil {
		return nil, result.Error
	}

	channels := make([]*channel.Channel, len(dbChannels))
	for i, dbChannel := range dbChannels {
		channels[i] = toDomainChannel(&dbChannel)
	}

	return channels, nil
}

// Update updates a channel
func (r *channelRepository) Update(ctx context.Context, ch *channel.Channel) error {
	dbChannel := toDBChannel(ch)
	result := r.db.WithContext(ctx).
		Model(&Channel{}).
		Where("id = ?", ch.ID).
		Updates(dbChannel)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return channel.ErrChannelNotFound
	}
	return nil
}

// Delete deletes a channel
func (r *channelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Channel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return channel.ErrChannelNotFound
	}
	return nil
}

// Subscribe subscribes a user to a channel
func (r *channelRepository) Subscribe(ctx context.Context, subscriber *channel.Subscriber) error {
	dbSubscriber := toDBChannelSubscriber(subscriber)
	result := r.db.WithContext(ctx).Create(dbSubscriber)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Unsubscribe unsubscribes a user from a channel
func (r *channelRepository) Unsubscribe(ctx context.Context, channelID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Delete(&ChannelSubscriber{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return channel.ErrSubscriberNotFound
	}
	return nil
}

// FindSubscriber finds a specific subscriber
func (r *channelRepository) FindSubscriber(ctx context.Context, channelID, userID uuid.UUID) (*channel.Subscriber, error) {
	var dbSubscriber ChannelSubscriber
	result := r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		First(&dbSubscriber)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, channel.ErrSubscriberNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainChannelSubscriber(&dbSubscriber), nil
}

// IsSubscribed checks if a user is subscribed to a channel
func (r *channelRepository) IsSubscribed(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&ChannelSubscriber{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// UpdateSubscriberNotifications updates subscriber notification settings
func (r *channelRepository) UpdateSubscriberNotifications(ctx context.Context, channelID, userID uuid.UUID, enabled bool) error {
	result := r.db.WithContext(ctx).
		Model(&ChannelSubscriber{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Update("notifications_enabled", enabled)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return channel.ErrSubscriberNotFound
	}
	return nil
}

// CountSubscribers counts the number of subscribers in a channel
func (r *channelRepository) CountSubscribers(ctx context.Context, channelID uuid.UUID) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&ChannelSubscriber{}).
		Where("channel_id = ?", channelID).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

// FindUserSubscriptions finds all channels a user is subscribed to
func (r *channelRepository) FindUserSubscriptions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*channel.Channel, error) {
	var dbChannels []Channel

	result := r.db.WithContext(ctx).
		Joins("JOIN channel_subscribers ON channels.id = channel_subscribers.channel_id").
		Where("channel_subscribers.user_id = ?", userID).
		Order("channels.updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbChannels)

	if result.Error != nil {
		return nil, result.Error
	}

	channels := make([]*channel.Channel, len(dbChannels))
	for i, dbChannel := range dbChannels {
		channels[i] = toDomainChannel(&dbChannel)
	}

	return channels, nil
}

// AddAdmin adds an admin to a channel
func (r *channelRepository) AddAdmin(ctx context.Context, admin *channel.Admin) error {
	dbAdmin := toDBChannelAdmin(admin)
	result := r.db.WithContext(ctx).Create(dbAdmin)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// RemoveAdmin removes an admin from a channel
func (r *channelRepository) RemoveAdmin(ctx context.Context, channelID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Delete(&ChannelAdmin{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return channel.ErrAdminNotFound
	}
	return nil
}

// FindAdmin finds a specific admin
func (r *channelRepository) FindAdmin(ctx context.Context, channelID, userID uuid.UUID) (*channel.Admin, error) {
	var dbAdmin ChannelAdmin
	result := r.db.WithContext(ctx).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		First(&dbAdmin)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, channel.ErrAdminNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return toDomainChannelAdmin(&dbAdmin), nil
}

// FindAdmins finds all admins of a channel
func (r *channelRepository) FindAdmins(ctx context.Context, channelID uuid.UUID) ([]*channel.Admin, error) {
	var dbAdmins []ChannelAdmin
	result := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("added_at ASC").
		Find(&dbAdmins)

	if result.Error != nil {
		return nil, result.Error
	}

	admins := make([]*channel.Admin, len(dbAdmins))
	for i, dbAdmin := range dbAdmins {
		admins[i] = toDomainChannelAdmin(&dbAdmin)
	}

	return admins, nil
}

// IsAdmin checks if a user is an admin of a channel
func (r *channelRepository) IsAdmin(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&ChannelAdmin{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// UpdateAdminPermissions updates an admin's permissions
func (r *channelRepository) UpdateAdminPermissions(ctx context.Context, channelID, userID uuid.UUID, permissions *channel.AdminPermissions) error {
	permissionsJSON, _ := json.Marshal(permissions)

	result := r.db.WithContext(ctx).
		Model(&ChannelAdmin{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Update("permissions", string(permissionsJSON))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return channel.ErrAdminNotFound
	}
	return nil
}

// Mapper functions

func toDBChannel(ch *channel.Channel) *Channel {
	var settingsJSON string
	if ch.Settings != nil {
		data, _ := json.Marshal(ch.Settings)
		settingsJSON = string(data)
	}

	return &Channel{
		ID:              ch.ID,
		ConversationID:  ch.ConversationID,
		Name:            ch.Name,
		Username:        ch.Username,
		Description:     ch.Description,
		Avatar:          ch.Avatar,
		OwnerID:         ch.OwnerID,
		IsPublic:        ch.IsPublic,
		SubscriberCount: ch.SubscriberCount,
		Settings:        settingsJSON,
		CreatedAt:       ch.CreatedAt,
		UpdatedAt:       ch.UpdatedAt,
	}
}

func toDomainChannel(dbChannel *Channel) *channel.Channel {
	var settings channel.Settings
	if dbChannel.Settings != "" {
		json.Unmarshal([]byte(dbChannel.Settings), &settings)
	}

	return &channel.Channel{
		ID:              dbChannel.ID,
		ConversationID:  dbChannel.ConversationID,
		Name:            dbChannel.Name,
		Username:        dbChannel.Username,
		Description:     dbChannel.Description,
		Avatar:          dbChannel.Avatar,
		OwnerID:         dbChannel.OwnerID,
		IsPublic:        dbChannel.IsPublic,
		SubscriberCount: dbChannel.SubscriberCount,
		Settings:        &settings,
		CreatedAt:       dbChannel.CreatedAt,
		UpdatedAt:       dbChannel.UpdatedAt,
	}
}

func toDBChannelSubscriber(subscriber *channel.Subscriber) *ChannelSubscriber {
	return &ChannelSubscriber{
		ChannelID:            subscriber.ChannelID,
		UserID:               subscriber.UserID,
		NotificationsEnabled: subscriber.NotificationsEnabled,
		SubscribedAt:         subscriber.SubscribedAt,
	}
}

func toDomainChannelSubscriber(dbSubscriber *ChannelSubscriber) *channel.Subscriber {
	return &channel.Subscriber{
		ChannelID:            dbSubscriber.ChannelID,
		UserID:               dbSubscriber.UserID,
		NotificationsEnabled: dbSubscriber.NotificationsEnabled,
		SubscribedAt:         dbSubscriber.SubscribedAt,
	}
}

func toDBChannelAdmin(admin *channel.Admin) *ChannelAdmin {
	var permissionsJSON string
	if admin.Permissions != nil {
		data, _ := json.Marshal(admin.Permissions)
		permissionsJSON = string(data)
	}

	return &ChannelAdmin{
		ChannelID:   admin.ChannelID,
		UserID:      admin.UserID,
		Permissions: permissionsJSON,
		AddedAt:     admin.AddedAt,
	}
}

func toDomainChannelAdmin(dbAdmin *ChannelAdmin) *channel.Admin {
	var permissions channel.AdminPermissions
	if dbAdmin.Permissions != "" {
		json.Unmarshal([]byte(dbAdmin.Permissions), &permissions)
	}

	return &channel.Admin{
		ChannelID:   dbAdmin.ChannelID,
		UserID:      dbAdmin.UserID,
		Permissions: &permissions,
		AddedAt:     dbAdmin.AddedAt,
	}
}
