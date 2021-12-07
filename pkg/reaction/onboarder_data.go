package reaction

import (
	"context"
	"fmt"
	"time"
)

var (
	onboarderKeyHangingMessages = "b4t:onboarder:hanging_messages:%d:%d" // channel:message_id

	onboarderKeyAnswers           = "b4t:onboarder:answers:%s"            // username
	onboarderKeyMembershipPending = "b4t:onboarder:membership-pending:%s" // username
)

func (r *Onboarder) addHangingMessage(ctx context.Context, messageID int, val string) error {
	key := fmt.Sprintf(onboarderKeyHangingMessages, r.groupChat.ID, messageID)
	err := r.rdb.Set(ctx, key, val, r.expiry).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) removeHangingMessage(ctx context.Context, messageID int) error {
	key := fmt.Sprintf(onboarderKeyHangingMessages, r.groupChat.ID, messageID)
	err := r.rdb.Del(ctx, onboarderKeyHangingMessages, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) isHangingMessage(ctx context.Context, messageID int) (exists bool, err error) {
	key := fmt.Sprintf(onboarderKeyHangingMessages, r.groupChat.ID, messageID)
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val != "", nil
}

func (r *Onboarder) userFromMessageID(ctx context.Context, messageID int) (val string, err error) {
	key := fmt.Sprintf(onboarderKeyHangingMessages, messageID)
	val, err = r.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *Onboarder) getAnswers(
	ctx context.Context,
	username string,
) (answers []string, err error) {
	key := fmt.Sprintf(onboarderKeyAnswers, username)
	answers, err = r.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (r *Onboarder) setAnswer(
	ctx context.Context,
	username string,
	answer string,
) error {
	key := fmt.Sprintf(onboarderKeyAnswers, username)
	if err := r.rdb.RPush(ctx, key, answer).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) getMembershipPending(
	ctx context.Context,
	username string,
) error {
	key := fmt.Sprintf(onboarderKeyMembershipPending, username)
	if err := r.rdb.Get(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) setMembershipPending(
	ctx context.Context,
	username string,
) error {
	key := fmt.Sprintf(onboarderKeyMembershipPending, username)
	if err := r.rdb.Set(ctx, key, "1", time.Duration(0)).Err(); err != nil {
		return err
	}
	return nil
}
