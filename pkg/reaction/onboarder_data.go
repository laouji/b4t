package reaction

import (
	"context"
	"fmt"
)

var (
	onboarderKeyHangingMessages = "b4t:onboarder:hanging_messages:%d"
	//onboarderKeyConversations   = "b4t:onboarder:conversations"
	onboarderKeyAnswers = "b4t:onboarder:answers:%s"
)

func (r *Onboarder) addHangingMessage(ctx context.Context, messageID int, val string) error {
	key := fmt.Sprintf(onboarderKeyHangingMessages, messageID)
	err := r.rdb.Set(ctx, key, val, r.expiry).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) removeHangingMessage(ctx context.Context, messageID int) error {
	key := fmt.Sprintf(onboarderKeyHangingMessages, messageID)
	err := r.rdb.Del(ctx, onboarderKeyHangingMessages, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) isHangingMessage(ctx context.Context, messageID int) (exists bool, err error) {
	key := fmt.Sprintf(onboarderKeyHangingMessages, messageID)
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

//func (r *Onboarder) hasConversation(
//	ctx context.Context,
//	conversationKey string,
//) (isMember bool, err error) {
//	isMember, err = r.rdb.SIsMember(ctx, onboarderKeyConversations, conversationKey).Result()
//	if err != nil {
//		return false, err
//	}
//	log.Printf("CHECKED KEY %s: %t", conversationKey, isMember)
//	return isMember, nil
//}
//
//func (r *Onboarder) addConversation(
//	ctx context.Context,
//	conversationKey string,
//) error {
//	if _, err := r.rdb.SAdd(
//		ctx,
//		onboarderKeyConversations,
//		conversationKey,
//	).Result(); err != nil {
//		return err
//	}
//	log.Printf("SET KEY %s", conversationKey)
//	return nil
//}

func (r *Onboarder) getAnswers(
	ctx context.Context,
	conversationKey string,
) (answers []string, err error) {
	key := fmt.Sprintf(onboarderKeyAnswers, conversationKey)
	answers, err = r.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (r *Onboarder) setAnswer(
	ctx context.Context,
	conversationKey string,
	answer string,
) error {
	key := fmt.Sprintf(onboarderKeyAnswers, conversationKey)
	if err := r.rdb.RPush(ctx, key, answer).Err(); err != nil {
		return err
	}
	return nil
}
